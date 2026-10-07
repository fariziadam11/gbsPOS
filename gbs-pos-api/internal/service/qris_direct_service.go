package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"gbs-pos-api/internal/config"
	"gbs-pos-api/internal/dto"
	"gbs-pos-api/internal/model"
	"gbs-pos-api/internal/repository"

	qrislib "gbs-common/pkg/qris"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// QrisDirectService handles QRIS static to dynamic conversion
type QrisDirectService struct {
	cfg          *config.Config
	qrisTxRepo   *repository.QrisTransactionRepository
	orderRepo    *repository.OrderRepository
	btnClient    *BtnSnapClient
	orderService *OrderService
}

func (s *QrisDirectService) SetOrderService(orderService *OrderService) {
	s.orderService = orderService
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
		btnClient:  NewBtnSnapClient(cfg),
	}
}

// ConvertQRIS converts the configured static QRIS to dynamic and creates a transaction record
func (s *QrisDirectService) ConvertQRIS(ctx context.Context, req dto.ConvertQRISRequest) (*dto.ConvertQRISResponse, error) {
	if s.cfg.BtnSnapEnabled {
		return s.convertBTNSnap(ctx, req)
	}

	// Get static QRIS from config
	staticQris := s.cfg.QrisDirectStaticQRIS
	if staticQris == "" {
		return nil, fmt.Errorf("QRIS_DIRECT_STATIC_QRIS is not configured")
	}

	var checkoutSnapshot *dto.QrisCheckoutSnapshotRequest
	var checkoutSnapshotJSON []byte
	if req.CheckoutSnapshot != nil {
		if req.OrderID == "" || req.CheckoutSnapshot.ID != req.OrderID {
			return nil, fmt.Errorf("checkout snapshot must match orderId")
		}
		if err := validateQrisCheckoutSnapshot(*req.CheckoutSnapshot); err != nil {
			return nil, err
		}
		amountCents, err := rupiahCents(req.Amount)
		if err != nil {
			return nil, err
		}
		totalCents, err := rupiahCents(req.CheckoutSnapshot.Total)
		if err != nil {
			return nil, err
		}
		if amountCents != totalCents {
			return nil, fmt.Errorf("QRIS amount does not match checkout total")
		}
		checkoutSnapshot = req.CheckoutSnapshot
		encodedSnapshot, err := json.Marshal(checkoutSnapshot)
		if err != nil {
			return nil, fmt.Errorf("encode checkout snapshot: %w", err)
		}
		checkoutSnapshotJSON = encodedSnapshot
	}

	// Legacy QRIS can start before the order is stored, using its checkout snapshot.
	if req.OrderID != "" {
		order, err := s.orderRepo.FindByID(req.OrderID)
		if err != nil && !(checkoutSnapshot != nil && errors.Is(err, gorm.ErrRecordNotFound)) {
			return nil, fmt.Errorf("order not found: %v", err)
		}
		if order != nil && order.PaymentMethod != "QRIS" {
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
		CheckoutSnapshot:  checkoutSnapshotJSON,
		ExpiresAt:         expiresAt,
	}

	if err := s.qrisTxRepo.Create(ctx, tx); err != nil {
		return nil, fmt.Errorf("failed to create transaction: %v", err)
	}

	return &dto.ConvertQRISResponse{
		ID:               txID,
		OrderID:          req.OrderID,
		CheckoutSnapshot: checkoutSnapshot,
		DynamicQris:      dynamicQris,
		Amount:           req.Amount,
		FeeType:          req.FeeType,
		FeeValue:         req.FeeValue,
		FeeAmount:        feeAmount,
		TotalAmount:      totalAmount,
		MerchantName:     s.cfg.QrisDirectMerchantName,
		MerchantCity:     s.cfg.QrisDirectMerchantCity,
		Provider:         s.cfg.QrisDirectProvider,
		QRCodeBase64:     dynamicQris, // Frontend will generate QR from this string
		ExpiresAt:        expiresAt,
	}, nil
}

func (s *QrisDirectService) convertBTNSnap(ctx context.Context, req dto.ConvertQRISRequest) (*dto.ConvertQRISResponse, error) {
	if req.OrderID == "" || req.CheckoutSnapshot == nil || req.CheckoutSnapshot.ID != req.OrderID {
		return nil, fmt.Errorf("orderId and matching checkoutSnapshot are required for BTN QRIS")
	}
	if req.FeeType != "" || req.FeeValue != 0 {
		return nil, fmt.Errorf("BTN QRIS amount must match the checkout total; fees are not supported")
	}
	snapshot := req.CheckoutSnapshot
	if s.orderService == nil {
		return nil, fmt.Errorf("BTN checkout validation is not configured")
	}
	if err := s.orderService.ValidateBTNCheckoutSnapshot(*snapshot); err != nil {
		return nil, err
	}
	requestedCents, err := rupiahCents(req.Amount)
	if err != nil || requestedCents <= 0 {
		return nil, fmt.Errorf("invalid BTN QRIS amount")
	}
	snapshotCents, _ := rupiahCents(snapshot.Total)
	if requestedCents != snapshotCents {
		return nil, fmt.Errorf("QRIS amount does not match checkout total")
	}

	previous, err := s.qrisTxRepo.FindLatestByOrderIDAndProvider(ctx, req.OrderID, model.QrisProviderBTNSnap)
	if err != nil {
		return nil, fmt.Errorf("find previous BTN transaction: %w", err)
	}
	if previous != nil {
		if previous.Status == model.QrisTransactionStatusPaid || previous.Status == model.QrisTransactionStatusUnknown ||
			previous.Status == model.QrisTransactionStatusPending || previous.Status == model.QrisTransactionStatusAwaitingConfirmation ||
			(previous.Status == model.QrisTransactionStatusExpired && previous.BankStatus != "07") ||
			(previous.Status == model.QrisTransactionStatusCancelled && previous.BankStatus != "05") ||
			(previous.Status == model.QrisTransactionStatusFailed && previous.BankStatus != "06") {
			if previous.DynamicQrisString != "" {
				return s.btnConvertResponse(previous)
			}
			return nil, fmt.Errorf("previous BTN QR result is unknown; check its status before retrying")
		}
	}

	partnerReferenceNo := uuid.NewString()
	transactionID := "BTN-" + uuid.NewString()
	expiresMinutes := s.cfg.QrisDirectExpiresMinutes
	if expiresMinutes <= 0 {
		expiresMinutes = 15
	}
	wib := time.FixedZone("WIB", 7*60*60)
	now := time.Now().In(wib)
	expiresAt := now.Add(time.Duration(expiresMinutes) * time.Minute)
	checkoutJSON, err := json.Marshal(snapshot)
	if err != nil {
		return nil, fmt.Errorf("encode checkout snapshot: %w", err)
	}

	tx := &model.QrisTransaction{
		ID:                 transactionID,
		OrderID:            req.OrderID,
		PartnerReferenceNo: partnerReferenceNo,
		Amount:             float64(requestedCents) / 100,
		TotalAmount:        float64(requestedCents) / 100,
		MerchantName:       s.cfg.BtnSnapMerchantID,
		MerchantID:         s.cfg.BtnSnapMerchantID,
		Currency:           "IDR",
		Provider:           model.QrisProviderBTNSnap,
		TerminalID:         s.cfg.BtnSnapTerminalID,
		Status:             model.QrisTransactionStatusPending,
		CheckoutSnapshot:   checkoutJSON,
		ExpiresAt:          expiresAt,
	}
	if err := s.qrisTxRepo.Create(ctx, tx); err != nil {
		return nil, fmt.Errorf("save BTN transaction before Generate: %w", err)
	}

	validityPeriod := expiresAt.Format(time.RFC3339)
	generated, err := s.btnClient.GenerateQR(ctx, dto.BtnSnapGenerateRequest{
		PartnerReferenceNo: partnerReferenceNo,
		Amount:             dto.BtnSnapAmount{Value: fmt.Sprintf("%d.%02d", requestedCents/100, requestedCents%100), Currency: "IDR"},
		MerchantID:         s.cfg.BtnSnapMerchantID,
		TerminalID:         s.cfg.BtnSnapTerminalID,
		ValidityPeriod:     validityPeriod,
		AdditionalInfo:     map[string]any{"type_qris": "D"},
	})
	if err != nil {
		_, _ = s.qrisTxRepo.ApplyBTNStatus(ctx, tx.ID, model.QrisTransactionStatusUnknown, "", "Generate result unknown", "", "", nil)
		return nil, fmt.Errorf("BTN Generate outcome is unknown; reconcile using the same reference: %w", err)
	}
	if generated.PartnerReferenceNo != partnerReferenceNo || generated.QRContent == "" || generated.TerminalID != tx.TerminalID {
		_, _ = s.qrisTxRepo.ApplyBTNStatus(ctx, tx.ID, model.QrisTransactionStatusUnknown, "", "Generate response did not match request", generated.ReferenceNo, "", nil)
		return nil, fmt.Errorf("BTN Generate response did not match the saved payment attempt")
	}
	if err := s.qrisTxRepo.UpdateBTNGenerated(ctx, tx.ID, generated.ReferenceNo, generated.QRContent, generated.MerchantName, tx.TerminalID); err != nil {
		return nil, fmt.Errorf("save BTN QR response: %w", err)
	}
	tx.BankReferenceNo = generated.ReferenceNo
	tx.DynamicQrisString = generated.QRContent
	if generated.MerchantName != "" {
		tx.MerchantName = generated.MerchantName
	}
	return s.btnConvertResponse(tx)
}

func validateQrisCheckoutSnapshot(snapshot dto.QrisCheckoutSnapshotRequest) error {
	if snapshot.ID == "" || (len(snapshot.Items) == 0 && len(snapshot.PpobItems) == 0) {
		return fmt.Errorf("checkout snapshot requires an order ID and at least one item")
	}
	var retailSubtotal int64
	for _, item := range snapshot.Items {
		price, err := rupiahCents(item.ProductPrice)
		if err != nil || item.Qty <= 0 {
			return fmt.Errorf("invalid checkout item")
		}
		lineSubtotal, err := rupiahCents(item.Subtotal)
		if err != nil || lineSubtotal != price*int64(item.Qty) {
			return fmt.Errorf("checkout item subtotal does not match its quantity and price")
		}
		retailSubtotal += lineSubtotal
	}
	for _, item := range snapshot.PpobItems {
		if item.PpobProductID == "" || item.CustomerID == "" || item.Amount < 0 || item.AdminFee < 0 || item.Total != item.Amount+item.AdminFee {
			return fmt.Errorf("invalid PPOB checkout item")
		}
		retailSubtotal += item.Total
	}
	subtotal, err := rupiahCents(snapshot.Subtotal)
	if err != nil || subtotal != retailSubtotal {
		return fmt.Errorf("checkout subtotal does not match its items")
	}
	tax, err := rupiahCents(snapshot.Tax)
	if err != nil {
		return fmt.Errorf("invalid checkout tax")
	}
	total, err := rupiahCents(snapshot.Total)
	if err != nil || total <= 0 {
		return fmt.Errorf("invalid checkout total")
	}
	discount := int64(0)
	if snapshot.DiscountAmount != nil {
		discount, err = rupiahCents(*snapshot.DiscountAmount)
		if err != nil || discount < 0 {
			return fmt.Errorf("invalid checkout discount")
		}
	}
	if subtotal+tax-discount != total {
		return fmt.Errorf("checkout total does not match subtotal, tax, and discount")
	}
	return nil
}

func rupiahCents(amount float64) (int64, error) {
	if math.IsNaN(amount) || math.IsInf(amount, 0) || amount < 0 {
		return 0, fmt.Errorf("invalid amount")
	}
	cents := math.Round(amount * 100)
	if math.Abs(amount*100-cents) > 0.0001 {
		return 0, fmt.Errorf("amount must have no more than two decimal places")
	}
	return int64(cents), nil
}

func (s *QrisDirectService) btnConvertResponse(tx *model.QrisTransaction) (*dto.ConvertQRISResponse, error) {
	var snapshot *dto.QrisCheckoutSnapshotRequest
	if len(tx.CheckoutSnapshot) != 0 {
		var stored dto.QrisCheckoutSnapshotRequest
		if err := json.Unmarshal(tx.CheckoutSnapshot, &stored); err != nil {
			return nil, fmt.Errorf("decode stored checkout snapshot: %w", err)
		}
		snapshot = &stored
	}
	return &dto.ConvertQRISResponse{
		ID: tx.ID, OrderID: tx.OrderID, PartnerReferenceNo: tx.PartnerReferenceNo,
		BankReferenceNo: tx.BankReferenceNo, DynamicQris: tx.DynamicQrisString,
		Amount: tx.Amount, TotalAmount: tx.TotalAmount, FeeAmount: 0,
		MerchantName: tx.MerchantName, Provider: tx.Provider, Currency: tx.Currency,
		Status: tx.Status, QRCodeBase64: tx.DynamicQrisString, ExpiresAt: tx.ExpiresAt,
		CheckoutSnapshot: snapshot,
	}, nil
}

// GetTransactionStatus retrieves the status of a QRIS transaction
func (s *QrisDirectService) GetTransactionStatus(ctx context.Context, transactionID string) (*dto.GetQRISStatusResponse, error) {
	tx, err := s.qrisTxRepo.FindByID(ctx, transactionID)
	if err != nil {
		return nil, err
	}
	if tx.Provider == model.QrisProviderBTNSnap {
		return s.getBTNStatus(ctx, tx)
	}

	// Check if expired and atomically mark as expired
	// This prevents TOCTOU race where transaction could be confirmed between check and return
	if tx.Status == model.QrisTransactionStatusPending && time.Now().After(tx.ExpiresAt) {
		// Atomically mark as expired - safe even if already confirmed in another request
		if err := s.qrisTxRepo.MarkAsExpired(ctx, tx.ID); err == nil {
			tx.Status = model.QrisTransactionStatusExpired
		}
		// If error, continue with original status - it will be marked expired next time
	}
	var snapshot *dto.QrisCheckoutSnapshotRequest
	if len(tx.CheckoutSnapshot) != 0 {
		var stored dto.QrisCheckoutSnapshotRequest
		if err := json.Unmarshal(tx.CheckoutSnapshot, &stored); err != nil {
			return nil, fmt.Errorf("decode stored checkout snapshot: %w", err)
		}
		snapshot = &stored
	}

	return &dto.GetQRISStatusResponse{
		ID:               tx.ID,
		OrderID:          tx.OrderID,
		Amount:           tx.Amount,
		FeeAmount:        tx.FeeAmount,
		TotalAmount:      tx.TotalAmount,
		Provider:         tx.Provider,
		Status:           tx.Status,
		PaidAt:           tx.PaidAt,
		ExpiresAt:        tx.ExpiresAt,
		CreatedAt:        tx.CreatedAt,
		DynamicQris:      tx.DynamicQrisString,
		CheckoutSnapshot: snapshot,
	}, nil
}

func (s *QrisDirectService) GetLatestBTNStatusForOrder(ctx context.Context, orderID string) (*dto.GetQRISStatusResponse, error) {
	provider := s.cfg.QrisDirectProvider
	if s.cfg.BtnSnapEnabled {
		provider = model.QrisProviderBTNSnap
	}
	tx, err := s.qrisTxRepo.FindLatestByOrderIDAndProvider(ctx, orderID, provider)
	if err != nil {
		return nil, err
	}
	if tx == nil {
		return nil, fmt.Errorf("QRIS transaction not found")
	}
	return s.GetTransactionStatus(ctx, tx.ID)
}

func (s *QrisDirectService) getBTNStatus(ctx context.Context, tx *model.QrisTransaction) (*dto.GetQRISStatusResponse, error) {
	query, err := s.btnClient.QueryPayment(ctx, dto.BtnSnapQueryRequest{
		OriginalPartnerReferenceNo: tx.PartnerReferenceNo,
		OriginalReferenceNo:        tx.BankReferenceNo,
		ServiceCode:                "47",
		MerchantID:                 tx.MerchantID,
		AdditionalInfo:             map[string]any{},
	})
	if err != nil {
		return nil, err
	}
	if query.ResponseCode == "4045101" {
		return s.btnStatusResponse(tx)
	}
	if query.ResponseCode != "2005100" {
		return nil, fmt.Errorf("BTN Query Payment rejected: %s %s", query.ResponseCode, query.ResponseMessage)
	}
	if query.OriginalPartnerReferenceNo != tx.PartnerReferenceNo || (query.ServiceCode != "" && query.ServiceCode != "47") {
		return nil, fmt.Errorf("BTN Query response reference does not match transaction")
	}
	if query.AdditionalInfo.MerchantID != "" && query.AdditionalInfo.MerchantID != tx.MerchantID {
		return nil, fmt.Errorf("BTN Query merchant does not match transaction")
	}
	if query.TerminalID != "" && query.TerminalID != tx.TerminalID {
		return nil, fmt.Errorf("BTN Query terminal does not match transaction")
	}

	status := model.QrisTransactionStatusUnknown
	var paidAt *time.Time
	switch query.LatestTransactionStatus {
	case "00":
		if query.OriginalReferenceNo == "" || query.Amount.Currency != "IDR" ||
			query.AdditionalInfo.MerchantID != tx.MerchantID || query.TerminalID != tx.TerminalID {
			return nil, fmt.Errorf("BTN success response is missing or mismatches payment reference, merchant, terminal, or currency")
		}
		paidCents, err := strconv.ParseInt(strings.ReplaceAll(query.Amount.Value, ".", ""), 10, 64)
		if err != nil || !validBTNAmount(query.Amount.Value) || paidCents != int64(math.Round(tx.TotalAmount*100)) {
			return nil, fmt.Errorf("BTN payment amount does not match checkout")
		}
		parsed, err := time.Parse(time.RFC3339, query.PaidTime)
		if err != nil || query.PaidTime == "" {
			return nil, fmt.Errorf("BTN success response is missing a valid paidTime")
		}
		status = model.QrisTransactionStatusPaid
		paidAt = &parsed
	case "01", "02", "03":
		status = model.QrisTransactionStatusPending
	case "04":
		status = model.QrisTransactionStatusRefunded
	case "05":
		status = model.QrisTransactionStatusCancelled
	case "06":
		status = model.QrisTransactionStatusFailed
	case "07":
		if !time.Now().Before(tx.ExpiresAt) && query.Amount.Value == "0.00" && query.OriginalReferenceNo == "" {
			status = model.QrisTransactionStatusExpired
		}
	}
	updated, err := s.qrisTxRepo.ApplyBTNStatus(
		ctx, tx.ID, status, query.LatestTransactionStatus, query.TransactionStatusDesc,
		query.OriginalReferenceNo, "", paidAt,
	)
	if err != nil {
		return nil, err
	}
	if updated {
		tx.Status = status
		tx.BankStatus = query.LatestTransactionStatus
		tx.BankStatusDescription = query.TransactionStatusDesc
		if query.OriginalReferenceNo != "" {
			tx.BankReferenceNo = query.OriginalReferenceNo
		}
		if paidAt != nil {
			tx.PaidAt = paidAt
			tx.BankPaidAt = paidAt
		}
	}
	return s.btnStatusResponse(tx)
}

func validBTNAmount(value string) bool {
	if len(value) < 4 || !strings.Contains(value, ".") {
		return false
	}
	parts := strings.Split(value, ".")
	if len(parts) != 2 || len(parts[1]) != 2 || parts[0] == "" {
		return false
	}
	for _, part := range parts {
		for _, r := range part {
			if r < '0' || r > '9' {
				return false
			}
		}
	}
	return true
}

func (s *QrisDirectService) btnStatusResponse(tx *model.QrisTransaction) (*dto.GetQRISStatusResponse, error) {
	var snapshot *dto.QrisCheckoutSnapshotRequest
	if len(tx.CheckoutSnapshot) != 0 {
		var stored dto.QrisCheckoutSnapshotRequest
		if err := json.Unmarshal(tx.CheckoutSnapshot, &stored); err != nil {
			return nil, fmt.Errorf("decode stored checkout snapshot: %w", err)
		}
		snapshot = &stored
	}
	return &dto.GetQRISStatusResponse{
		ID: tx.ID, OrderID: tx.OrderID, PartnerReferenceNo: tx.PartnerReferenceNo,
		BankReferenceNo: tx.BankReferenceNo, BankPaymentReference: tx.BankPaymentReference,
		Amount: tx.Amount, FeeAmount: tx.FeeAmount, TotalAmount: tx.TotalAmount,
		Provider: tx.Provider, Currency: tx.Currency, Status: tx.Status,
		BankStatus: tx.BankStatus, BankStatusDescription: tx.BankStatusDescription,
		PaidAt: tx.BankPaidAt, ExpiresAt: tx.ExpiresAt, CreatedAt: tx.CreatedAt,
		DynamicQris: tx.DynamicQrisString, CheckoutSnapshot: snapshot,
	}, nil
}

// ConfirmPayment confirms that a QRIS payment has been completed
func (s *QrisDirectService) ConfirmPayment(ctx context.Context, transactionID string) error {
	tx, err := s.qrisTxRepo.FindByID(ctx, transactionID)
	if err != nil {
		return err
	}
	if tx.Provider == model.QrisProviderBTNSnap {
		return fmt.Errorf("BTN payment must be confirmed by Query Payment")
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
	if tx.Provider == model.QrisProviderBTNSnap {
		return fmt.Errorf("BTN QR cannot be canceled by POS; check its bank status")
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
	if tx.Provider == model.QrisProviderBTNSnap {
		return fmt.Errorf("BTN payment must be confirmed by Query Payment")
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

		// Use atomic update to prevent race condition
		// Only confirms if status is still AWAITING_CONFIRMATION
		bgCtx := context.Background()
		updated, err := s.qrisTxRepo.AtomicConfirmIfStatus(bgCtx, transactionID, model.QrisTransactionStatusAwaitingConfirmation)
		if err != nil {
			// Log error but don't return - goroutine
			return
		}
		if !updated {
			// Status changed (cancelled, expired, or already confirmed) - normal exit
			return
		}
		// Successfully confirmed - exit
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
