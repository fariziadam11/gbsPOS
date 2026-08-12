package service

import (
	"context"
	"fmt"
	"math"
	"time"

	"gbs-pos-api/internal/model"
	"gbs-pos-api/internal/repository"
	ws "gbs-pos-api/internal/websocket"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

type CardPaymentService struct {
	repo    *repository.CardPaymentRepository
	orders  *OrderService
	devices *repository.CompanionDeviceRepository
	hub     *ws.Hub
}

func NewCardPaymentService(repo *repository.CardPaymentRepository, orders *OrderService, devices *repository.CompanionDeviceRepository, hub *ws.Hub) *CardPaymentService {
	return &CardPaymentService{repo: repo, orders: orders, devices: devices, hub: hub}
}

func (s *CardPaymentService) Create(ctx context.Context, orderID string, amount float64, terminalID, deviceID string) (*model.CardPayment, error) {
	if orderID == "" || amount <= 0 || terminalID == "" || deviceID == "" {
		return nil, fmt.Errorf("orderId, amount, terminalId, and deviceId are required")
	}
	order, err := s.orders.Get(orderID)
	if err != nil {
		return nil, fmt.Errorf("order not found")
	}
	if order.IsVoided || order.IsSettled {
		return nil, fmt.Errorf("order is not payable")
	}
	if math.Abs(order.Total-amount) > 0.01 {
		return nil, fmt.Errorf("payment amount does not match order total")
	}
	if order.TerminalID != "" && order.TerminalID != terminalID {
		return nil, fmt.Errorf("terminal does not match order")
	}
	registered, err := s.devices.Exists(ctx, deviceID)
	if err != nil {
		return nil, err
	}
	if !registered {
		return nil, fmt.Errorf("companion device is not registered")
	}
	payment := &model.CardPayment{
		ID:         uuid.New(),
		OrderID:    orderID,
		Amount:     amount,
		Status:     model.CardPaymentWaiting,
		TerminalID: terminalID,
		DeviceID:   deviceID,
		ExpiresAt:  time.Now().Add(5 * time.Minute),
	}
	if err := s.repo.Create(ctx, payment); err != nil {
		return nil, err
	}
	s.broadcastRequest(payment)
	s.broadcastStatus(payment)
	return payment, nil
}

func (s *CardPaymentService) Get(ctx context.Context, id uuid.UUID) (*model.CardPayment, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *CardPaymentService) Pending(ctx context.Context, deviceID string) ([]model.CardPayment, error) {
	if deviceID == "" {
		return nil, fmt.Errorf("deviceId is required")
	}
	return s.repo.FindPendingByDevice(ctx, deviceID)
}

func (s *CardPaymentService) Cancel(ctx context.Context, id uuid.UUID) (*model.CardPayment, error) {
	payment, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if payment.Status != model.CardPaymentWaiting {
		return nil, fmt.Errorf("payment cannot be cancelled in status %s", payment.Status)
	}
	payment.Status = model.CardPaymentCancelled
	if err := s.repo.Update(ctx, payment); err != nil {
		return nil, err
	}
	s.voidOrder(payment.OrderID, "card payment cancelled")
	s.broadcastStatus(payment)
	return payment, nil
}

func (s *CardPaymentService) UpdateFromCompanion(ctx context.Context, client *ws.Client, message ws.Message) error {
	if client.Type != ws.ClientCompanion || message.PaymentID == "" {
		return fmt.Errorf("invalid companion payment update")
	}
	id, err := uuid.Parse(message.PaymentID)
	if err != nil {
		return fmt.Errorf("invalid payment id")
	}
	payment, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if payment.DeviceID != client.ID {
		return fmt.Errorf("payment does not belong to companion")
	}
	if message.Status != model.CardPaymentProcessing && message.Status != model.CardPaymentSuccess && message.Status != model.CardPaymentFailed {
		return fmt.Errorf("invalid payment status")
	}
	if message.Status == model.CardPaymentSuccess && message.TransactionID == "" {
		return fmt.Errorf("successful payment requires transaction id")
	}
	if payment.Status == model.CardPaymentSuccess || payment.Status == model.CardPaymentCancelled || payment.Status == model.CardPaymentExpired {
		return nil
	}
	if message.Status == model.CardPaymentProcessing && payment.Status != model.CardPaymentWaiting {
		return fmt.Errorf("payment cannot start processing from status %s", payment.Status)
	}
	if (message.Status == model.CardPaymentSuccess || message.Status == model.CardPaymentFailed) && payment.Status != model.CardPaymentWaiting && payment.Status != model.CardPaymentProcessing {
		return fmt.Errorf("payment cannot complete from status %s", payment.Status)
	}
	payment.Status = message.Status
	payment.TransactionID = message.TransactionID
	payment.CardBrand = message.CardBrand
	payment.MaskedCard = message.MaskedCard
	payment.AuthCode = message.AuthCode
	payment.EntryMode = message.EntryMode
	payment.AcqMID = message.AcqMID
	payment.AcqTID = message.AcqTID
	payment.PosMessageID = message.PosMessageID
	payment.FailureReason = message.FailureReason
	if payment.Status == model.CardPaymentSuccess {
		if err := s.repo.FinalizeSuccess(ctx, payment); err != nil {
			return err
		}
	} else {
		if err := s.repo.Update(ctx, payment); err != nil {
			return err
		}
	}
	if payment.Status == model.CardPaymentFailed || payment.Status == model.CardPaymentCancelled {
		s.voidOrder(payment.OrderID, "card payment not completed")
	}
	s.broadcastStatus(payment)
	return nil
}

func (s *CardPaymentService) Expire(ctx context.Context) error {
	payments, err := s.repo.FindExpired(ctx, time.Now())
	if err != nil {
		return err
	}
	for _, payment := range payments {
		updated, err := s.repo.MarkExpired(ctx, payment.ID, time.Now())
		if err != nil {
			return err
		}
		if updated {
			payment.Status = model.CardPaymentExpired
			s.voidOrder(payment.OrderID, "card payment expired")
			s.broadcastStatus(&payment)
		}
	}
	return nil
}

func (s *CardPaymentService) voidOrder(orderID, reason string) {
	if _, err := s.orders.Void(orderID, reason, "system"); err != nil {
		log.Error().Err(err).Str("order_id", orderID).Str("reason", reason).Msg("failed to void card payment order")
	}
}

func (s *CardPaymentService) HandleMessage(client *ws.Client, message ws.Message) {
	if message.Type == "PING" {
		s.hub.Send(client.Type, client.ID, ws.Message{Type: "PONG"})
		return
	}
	if message.Type == "PAYMENT_STATUS_UPDATE" {
		_ = s.UpdateFromCompanion(context.Background(), client, message)
	}
}

func (s *CardPaymentService) broadcastRequest(payment *model.CardPayment) {
	s.hub.Send(ws.ClientCompanion, payment.DeviceID, ws.Message{
		Type:       "PAYMENT_REQUEST",
		PaymentID:  payment.ID.String(),
		OrderID:    payment.OrderID,
		Amount:     payment.Amount,
		Currency:   "IDR",
		ExpiresAt:  payment.ExpiresAt.UTC().Format(time.RFC3339),
		TerminalID: payment.TerminalID,
	})
}

func (s *CardPaymentService) broadcastStatus(payment *model.CardPayment) {
	s.hub.Send(ws.ClientPOS, payment.TerminalID, ws.Message{
		Type:          "PAYMENT_STATUS",
		PaymentID:     payment.ID.String(),
		OrderID:       payment.OrderID,
		Status:        payment.Status,
		Amount:        payment.Amount,
		TransactionID: payment.TransactionID,
		CardBrand:     payment.CardBrand,
		MaskedCard:    payment.MaskedCard,
		AuthCode:      payment.AuthCode,
		EntryMode:     payment.EntryMode,
		AcqMID:        payment.AcqMID,
		AcqTID:        payment.AcqTID,
		PosMessageID:  payment.PosMessageID,
		FailureReason: payment.FailureReason,
		UpdatedAt:     payment.UpdatedAt.UTC().Format(time.RFC3339),
	})
}
