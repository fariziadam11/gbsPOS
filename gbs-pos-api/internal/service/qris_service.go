package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"gbs-pos-api/internal/config"
	"gbs-pos-api/internal/dto"
	"gbs-pos-api/internal/model"
	"gbs-pos-api/internal/repository"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// QrisService handles QRIS payment operations via SumoPod
type QrisService struct {
	cfg       *config.Config
	db        *gorm.DB
	orderRepo *repository.OrderRepository
	client    *http.Client
}

// NewQrisService creates a new QRIS service
func NewQrisService(cfg *config.Config, db *gorm.DB, orderRepo *repository.OrderRepository) *QrisService {
	return &QrisService{
		cfg:       cfg,
		db:        db,
		orderRepo: orderRepo,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// InitPayment creates a QRIS payment link via SumoPod and updates the order
func (s *QrisService) InitPayment(ctx context.Context, req dto.CreateQrisPaymentRequest) (*dto.QrisInitResponse, error) {
	// Check if order exists and is QRIS payment
	order, err := s.orderRepo.FindByID(req.OrderID)
	if err != nil {
		return nil, fmt.Errorf("ORDER_NOT_FOUND: order %s not found", req.OrderID)
	}

	if order.PaymentMethod != "QRIS" {
		return nil, fmt.Errorf("INVALID_PAYMENT_METHOD: order is not QRIS payment")
	}

	// Check if payment already initialized
	if order.QrisPaymentID != "" && order.QrisStatus == dto.QrisStatusPending {
		if order.QrisExpiresAt == nil {
			return nil, fmt.Errorf("QRIS_STATE_ERROR: payment initialized but expires_at is nil")
		}
		return &dto.QrisInitResponse{
			OrderID:        order.ID,
			PaymentID:      order.QrisPaymentID,
			PaymentLinkURL: order.QrisLinkURL,
			Amount:         order.Total,
			Fee:            order.QrisFee,
			ExpiresAt:      *order.QrisExpiresAt,
		}, nil
	}

	// Create payment via SumoPod API
	sumopodReq := dto.SumoPodCreatePaymentRequest{
		OrderID:              order.ID,
		Amount:              req.Amount,
		Currency:            "IDR",
		ExpiresInHours:       s.cfg.SumoPodExpiresHours,
		SuccessReturnURL:     s.cfg.SumoPodSuccessURL,
		CancelReturnURL:      s.cfg.SumoPodCancelURL,
		PaymentMethodTypeCode: "QRIS",
	}

	sumopodResp, err := s.callSumoPodAPI(ctx, "/payments", sumopodReq)
	if err != nil {
		return nil, fmt.Errorf("SUMOPOD_API_ERROR: %v", err)
	}

	// Update order with QRIS payment info
	expiresAt := sumopodResp.ExpiresAt
	order.QrisPaymentID = sumopodResp.PaymentID
	order.QrisStatus = sumopodResp.Status
	order.QrisLinkURL = sumopodResp.PaymentLinkURL
	order.QrisExpiresAt = &expiresAt
	order.QrisFee = sumopodResp.Fee
	order.QrisNetAmount = sumopodResp.NetAmount

	if err := s.db.Save(order).Error; err != nil {
		return nil, fmt.Errorf("DATABASE_ERROR: failed to update order: %v", err)
	}

	return &dto.QrisInitResponse{
		OrderID:        order.ID,
		PaymentID:      sumopodResp.PaymentID,
		PaymentLinkURL: sumopodResp.PaymentLinkURL,
		Amount:         sumopodResp.Amount,
		Fee:            sumopodResp.Fee,
		ExpiresAt:      expiresAt,
	}, nil
}

// GetPaymentStatus retrieves QRIS payment status for an order
func (s *QrisService) GetPaymentStatus(ctx context.Context, orderID string) (*dto.QrisPaymentStatusResponse, error) {
	order, err := s.orderRepo.FindByID(orderID)
	if err != nil {
		return nil, fmt.Errorf("ORDER_NOT_FOUND: order %s not found", orderID)
	}

	if order.PaymentMethod != "QRIS" {
		return nil, fmt.Errorf("INVALID_PAYMENT_METHOD: order is not QRIS payment")
	}

	return &dto.QrisPaymentStatusResponse{
		OrderID:        order.ID,
		PaymentID:      order.QrisPaymentID,
		Status:         order.QrisStatus,
		PaymentLinkURL: order.QrisLinkURL,
		ExpiresAt:      order.QrisExpiresAt,
		Fee:            order.QrisFee,
		NetAmount:      order.QrisNetAmount,
		CompletedAt:    order.QrisCompletedAt,
	}, nil
}

// HandleWebhook processes SumoPod webhook notifications
func (s *QrisService) HandleWebhook(ctx context.Context, payload dto.SumoPodWebhookPayload) error {
	// Skip test events
	if payload.EventType == dto.WebhookEventPaymentTest {
		return nil
	}

	// Find order by payment ID
	order, err := s.findOrderByPaymentID(payload.Data.PaymentID)
	if err != nil {
		return err
	}

	// Update order based on event type
	switch payload.EventType {
	case dto.WebhookEventPaymentCompleted:
		return s.handlePaymentCompleted(ctx, order, payload.Data)
	case dto.WebhookEventPaymentFailed:
		return s.handlePaymentFailed(ctx, order)
	case dto.WebhookEventPaymentExpired:
		return s.handlePaymentExpired(ctx, order)
	default:
		return fmt.Errorf("UNKNOWN_EVENT_TYPE: %s", payload.EventType)
	}
}

func (s *QrisService) handlePaymentCompleted(ctx context.Context, order *model.Order, data dto.SumoPodWebhookPaymentData) error {
	order.QrisStatus = dto.QrisStatusCompleted
	order.QrisCompletedAt = &data.CompletedAt
	order.QrisFee = data.Fee
	order.QrisNetAmount = data.NetAmount

	if err := s.db.Save(order).Error; err != nil {
		return fmt.Errorf("DATABASE_ERROR: failed to update order: %v", err)
	}

	return nil
}

func (s *QrisService) handlePaymentFailed(ctx context.Context, order *model.Order) error {
	order.QrisStatus = dto.QrisStatusFailed

	if err := s.db.Save(order).Error; err != nil {
		return fmt.Errorf("DATABASE_ERROR: failed to update order: %v", err)
	}

	return nil
}

func (s *QrisService) handlePaymentExpired(ctx context.Context, order *model.Order) error {
	order.QrisStatus = dto.QrisStatusExpired

	if err := s.db.Save(order).Error; err != nil {
		return fmt.Errorf("DATABASE_ERROR: failed to update order: %v", err)
	}

	return nil
}

func (s *QrisService) findOrderByPaymentID(paymentID string) (*model.Order, error) {
	var order model.Order
	if err := s.db.Where("qris_payment_id = ?", paymentID).First(&order).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("ORDER_NOT_FOUND: no order found with payment_id %s", paymentID)
		}
		return nil, fmt.Errorf("DATABASE_ERROR: %v", err)
	}
	return &order, nil
}

func (s *QrisService) callSumoPodAPI(ctx context.Context, path string, body interface{}) (*dto.SumoPodCreatePaymentResponse, error) {
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("JSON_ERROR: failed to marshal request: %v", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.cfg.SumoPodAPIURL+path, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("REQUEST_ERROR: failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Api-Key", s.cfg.SumoPodAPIKey)

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("NETWORK_ERROR: failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API_ERROR: status=%d, body=%s", resp.StatusCode, string(respBody))
	}

	var result dto.SumoPodCreatePaymentResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("PARSE_ERROR: failed to decode response: %v", err)
	}

	return &result, nil
}

// GenerateOrderID generates a unique order ID for QRIS payment
func GenerateOrderID() string {
	return fmt.Sprintf("QRIS-%s", uuid.New().String()[:12])
}
