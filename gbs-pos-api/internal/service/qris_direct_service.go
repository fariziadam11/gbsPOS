package service

import (
	"context"
	"fmt"
	"time"

	"gbs-pos-api/internal/config"
	"gbs-pos-api/internal/dto"
	"gbs-pos-api/internal/model"
	"gbs-pos-api/internal/repository"

	"github.com/google/uuid"
	qrislib "gbs-common/pkg/qris"
)

// QrisDirectService handles QRIS static to dynamic conversion
type QrisDirectService struct {
	cfg        *config.Config
	qrisTxRepo *repository.QrisTransactionRepository
	orderRepo  *repository.OrderRepository
}

// NewQrisDirectService creates a new QRIS direct service
func NewQrisDirectService(
	cfg *config.Config,
	qrisTxRepo *repository.QrisTransactionRepository,
	orderRepo *repository.OrderRepository,
) *QrisDirectService {
	// Set CRC validation based on config
	qrislib.SkipCRCValidation = cfg.QrisDirectSkipCRCValidate

	return &QrisDirectService{
		cfg:        cfg,
		qrisTxRepo: qrisTxRepo,
		orderRepo:  orderRepo,
	}
}

// ConvertQRIS converts the configured static QRIS to dynamic and creates a transaction record
func (s *QrisDirectService) ConvertQRIS(ctx context.Context, req dto.ConvertQRISRequest) (*dto.ConvertQRISResponse, error) {
	// Get static QRIS from config
	staticQris := s.cfg.QrisDirectStaticQRIS
	if staticQris == "" {
		return nil, fmt.Errorf("QRIS_DIRECT_STATIC_QRIS is not configured")
	}

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
	dynamicQris, err := qrislib.ConvertWithFee(staticQris, qrislib.ConvertOptions{
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

	// Generate unique transaction ID
	txID := fmt.Sprintf("QRIS-%s", uuid.New().String()[:12])

	// Set expiration time from config (default 15 minutes)
	expiresMinutes := s.cfg.QrisDirectExpiresMinutes
	if expiresMinutes <= 0 {
		expiresMinutes = 15
	}
	expiresAt := time.Now().Add(time.Duration(expiresMinutes) * time.Minute)

	// Create transaction record
	tx := &model.QrisTransaction{
		ID:                txID,
		OrderID:           req.OrderID,
		StaticQrisString:  staticQris,
		DynamicQrisString: dynamicQris,
		Amount:            req.Amount,
		FeeType:           req.FeeType,
		FeeValue:          req.FeeValue,
		FeeAmount:         feeAmount,
		TotalAmount:       totalAmount,
		MerchantName:      s.cfg.QrisDirectMerchantName,
		MerchantCity:      s.cfg.QrisDirectMerchantCity,
		Provider:          s.cfg.QrisDirectProvider,
		Status:            model.QrisTransactionStatusPending,
		ExpiresAt:         expiresAt,
	}

	if err := s.qrisTxRepo.Create(ctx, tx); err != nil {
		return nil, fmt.Errorf("failed to create transaction: %v", err)
	}

	return &dto.ConvertQRISResponse{
		ID:            txID,
		OrderID:       req.OrderID,
		DynamicQris:   dynamicQris,
		Amount:        req.Amount,
		FeeType:       req.FeeType,
		FeeValue:      req.FeeValue,
		FeeAmount:     feeAmount,
		TotalAmount:   totalAmount,
		MerchantName:  s.cfg.QrisDirectMerchantName,
		MerchantCity:  s.cfg.QrisDirectMerchantCity,
		Provider:      s.cfg.QrisDirectProvider,
		QRCodeBase64:  dynamicQris, // Frontend will generate QR from this string
		ExpiresAt:     expiresAt,
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

	// Allow cancellation for PENDING or AWAITING_CONFIRMATION status
	if tx.Status != model.QrisTransactionStatusPending && tx.Status != model.QrisTransactionStatusAwaitingConfirmation {
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

// ConfirmPending marks a transaction as awaiting auto-confirmation
// It updates status to AWAITING_CONFIRMATION and spawns a goroutine
// that will auto-confirm the payment after the configured delay
func (s *QrisDirectService) ConfirmPending(ctx context.Context, transactionID string) error {
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

	// Update status to AWAITING_CONFIRMATION
	if err := s.qrisTxRepo.UpdateStatus(ctx, transactionID, model.QrisTransactionStatusAwaitingConfirmation); err != nil {
		return fmt.Errorf("failed to update status: %v", err)
	}

	// Get delay from config (default 3 seconds)
	delaySeconds := s.cfg.QrisDirectAutoConfirmDelaySeconds
	if delaySeconds <= 0 {
		delaySeconds = 3
	}

	// Spawn goroutine to auto-confirm after delay
	go func() {
		time.Sleep(time.Duration(delaySeconds) * time.Second)

		// Re-fetch the transaction to verify it's still awaiting confirmation
		// and not cancelled or expired
		bgCtx := context.Background()
		txCheck, err := s.qrisTxRepo.FindByID(bgCtx, transactionID)
		if err != nil {
			return
		}

		if txCheck.Status != model.QrisTransactionStatusAwaitingConfirmation {
			// Status changed, don't confirm
			return
		}

		if time.Now().After(txCheck.ExpiresAt) {
			// Expired, don't confirm
			return
		}

		// Confirm the payment
		if err := s.qrisTxRepo.MarkAsPaid(bgCtx, transactionID); err != nil {
			// Log error but don't return - goroutine
			return
		}
	}()

	return nil
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
