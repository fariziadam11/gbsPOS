package service

import (
	"context"
	"fmt"
	"time"

	"gbs-pos-api/internal/dto"
	"gbs-pos-api/internal/model"
	"gbs-pos-api/internal/repository"

	"github.com/google/uuid"
	qrislib "gbs-common/pkg/qris"
)

// QrisDirectService handles QRIS static to dynamic conversion
type QrisDirectService struct {
	qrisTxRepo *repository.QrisTransactionRepository
	orderRepo  *repository.OrderRepository
}

// NewQrisDirectService creates a new QRIS direct service
func NewQrisDirectService(
	qrisTxRepo *repository.QrisTransactionRepository,
	orderRepo *repository.OrderRepository,
) *QrisDirectService {
	return &QrisDirectService{
		qrisTxRepo: qrisTxRepo,
		orderRepo:  orderRepo,
	}
}

// ParseQRIS parses a QRIS string and returns structured data
func (s *QrisDirectService) ParseQRIS(ctx context.Context, req dto.ParseQRISRequest) (*dto.ParseQRISResponse, error) {
	parsed, err := qrislib.Parse(req.QrisString)
	if err != nil {
		return nil, fmt.Errorf("failed to parse QRIS: %v", err)
	}

	// Detect provider from merchant accounts
	provider := "Unknown"
	if len(parsed.MerchantAccounts) > 0 {
		provider = parsed.MerchantAccounts[0].ProviderID
		if provider == "" {
			provider = "Unknown"
		}
	}

	return &dto.ParseQRISResponse{
		IsStatic:      parsed.IsStatic,
		IsDynamic:     parsed.IsDynamic,
		MerchantName:  parsed.MerchantName,
		MerchantCity:  parsed.MerchantCity,
		Provider:      provider,
		MCC:           parsed.MCC,
		Currency:      parsed.Currency,
		CurrentAmount: parsed.Amount,
	}, nil
}

// ConvertQRIS converts a static QRIS to dynamic and creates a transaction record
func (s *QrisDirectService) ConvertQRIS(ctx context.Context, req dto.ConvertQRISRequest) (*dto.ConvertQRISResponse, error) {
	// Validate order if provided
	if req.OrderID != "" {
		order, err := s.orderRepo.FindByID(req.OrderID)
		if err != nil {
			return nil, fmt.Errorf("order not found: %v", err)
		}
		if order.PaymentMethod != "QRIS" {
			return nil, fmt.Errorf("order is not a QRIS payment")
		}
	}

	// Convert static to dynamic
	dynamicQris, err := qrislib.ConvertWithFee(req.QrisString, qrislib.ConvertOptions{
		Amount:   req.Amount,
		FeeType:  req.FeeType,
		FeeValue: req.FeeValue,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to convert QRIS: %v", err)
	}

	// Calculate fee
	feeAmount := qrislib.GetFee(req.Amount, req.FeeType, req.FeeValue)
	totalAmount := req.Amount + feeAmount

	// Parse the original QRIS for merchant info
	parsed, _ := qrislib.Parse(req.QrisString)

	// Generate unique transaction ID
	txID := fmt.Sprintf("QRIS-%s", uuid.New().String()[:12])

	// Create transaction record
	now := time.Now()
	expiresAt := now.Add(15 * time.Minute) // Default 15 minutes expiry

	tx := &model.QrisTransaction{
		ID:                txID,
		OrderID:           req.OrderID,
		StaticQrisString:  req.QrisString,
		DynamicQrisString: dynamicQris,
		Amount:            req.Amount,
		FeeType:           req.FeeType,
		FeeValue:          req.FeeValue,
		FeeAmount:         feeAmount,
		TotalAmount:       totalAmount,
		MerchantName:      parsed.MerchantName,
		MerchantCity:      parsed.MerchantCity,
		Provider:          detectProvider(req.QrisString),
		Status:            model.QrisTransactionStatusPending,
		ExpiresAt:         expiresAt,
	}

	if err := s.qrisTxRepo.Create(ctx, tx); err != nil {
		return nil, fmt.Errorf("failed to create transaction: %v", err)
	}

	// Generate QR code (placeholder - frontend should generate from dynamicQris string)
	qrCodeBase64 := dynamicQris // For now, return the string; frontend will handle QR generation

	return &dto.ConvertQRISResponse{
		ID:            txID,
		OrderID:       req.OrderID,
		OriginalQris:  req.QrisString,
		DynamicQris:   dynamicQris,
		Amount:        req.Amount,
		FeeType:      req.FeeType,
		FeeValue:     req.FeeValue,
		FeeAmount:    feeAmount,
		TotalAmount:  totalAmount,
		MerchantName: parsed.MerchantName,
		MerchantCity: parsed.MerchantCity,
		Provider:     detectProvider(req.QrisString),
		QRCodeBase64: qrCodeBase64,
		ExpiresAt:    expiresAt,
	}, nil
}

// GetTransactionStatus retrieves the status of a QRIS transaction
func (s *QrisDirectService) GetTransactionStatus(ctx context.Context, transactionID string) (*dto.GetQRISStatusResponse, error) {
	tx, err := s.qrisTxRepo.FindByID(ctx, transactionID)
	if err != nil {
		return nil, err
	}

	// Check if expired
	if tx.Status == model.QrisTransactionStatusPending && time.Now().After(tx.ExpiresAt) {
		// Mark as expired
		s.qrisTxRepo.MarkAsExpired(ctx, tx.ID)
		tx.Status = model.QrisTransactionStatusExpired
	}

	return &dto.GetQRISStatusResponse{
		ID:          tx.ID,
		OrderID:     tx.OrderID,
		Amount:      tx.Amount,
		FeeAmount:   tx.FeeAmount,
		TotalAmount: tx.TotalAmount,
		Provider:    tx.Provider,
		Status:      tx.Status,
		PaidAt:      tx.PaidAt,
		ExpiresAt:   tx.ExpiresAt,
		CreatedAt:   tx.CreatedAt,
	}, nil
}

// ConfirmPayment confirms that a QRIS payment has been completed
func (s *QrisDirectService) ConfirmPayment(ctx context.Context, transactionID string) error {
	tx, err := s.qrisTxRepo.FindByID(ctx, transactionID)
	if err != nil {
		return err
	}

	if tx.Status != model.QrisTransactionStatusPending {
		return fmt.Errorf("transaction is not pending, current status: %s", tx.Status)
	}

	if time.Now().After(tx.ExpiresAt) {
		return fmt.Errorf("transaction has expired")
	}

	if err := s.qrisTxRepo.MarkAsPaid(ctx, transactionID); err != nil {
		return fmt.Errorf("failed to confirm payment: %v", err)
	}

	return nil
}

// CancelPayment cancels a QRIS payment
func (s *QrisDirectService) CancelPayment(ctx context.Context, transactionID string, reason string, cancelledBy string) error {
	tx, err := s.qrisTxRepo.FindByID(ctx, transactionID)
	if err != nil {
		return err
	}

	if tx.Status != model.QrisTransactionStatusPending {
		return fmt.Errorf("transaction is not pending, current status: %s", tx.Status)
	}

	if err := s.qrisTxRepo.Cancel(ctx, transactionID, cancelledBy, reason); err != nil {
		return fmt.Errorf("failed to cancel payment: %v", err)
	}

	return nil
}

// GetPendingTransactions gets all pending QRIS transactions for a terminal
func (s *QrisDirectService) GetPendingTransactions(ctx context.Context, terminalID string) ([]model.QrisTransaction, error) {
	return s.qrisTxRepo.FindPendingByTerminalID(ctx, terminalID)
}

// ProcessExpiredTransactions marks expired transactions
func (s *QrisDirectService) ProcessExpiredTransactions(ctx context.Context) error {
	expired, err := s.qrisTxRepo.FindExpired(ctx)
	if err != nil {
		return err
	}

	for _, tx := range expired {
		if err := s.qrisTxRepo.MarkAsExpired(ctx, tx.ID); err != nil {
			// Log but continue
			continue
		}
	}

	return nil
}

// detectProvider detects payment provider from QRIS string
func detectProvider(qrisString string) string {
	parsed, err := qrislib.Parse(qrisString)
	if err != nil {
		return "Unknown"
	}

	if len(parsed.MerchantAccounts) > 0 {
		return parsed.MerchantAccounts[0].ProviderID
	}

	return "Unknown"
}
