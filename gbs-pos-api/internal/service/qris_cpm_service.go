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

	"github.com/google/uuid"
)

type QrisCPMService struct {
	cfg       *config.Config
	repo      *repository.QrisTransactionRepository
	orderSvc  *OrderService
	btnClient *BtnSnapClient
}

func NewQrisCPMService(cfg *config.Config, repo *repository.QrisTransactionRepository, orderSvc *OrderService) *QrisCPMService {
	return &QrisCPMService{cfg: cfg, repo: repo, orderSvc: orderSvc, btnClient: NewBtnSnapClient(cfg)}
}

func (s *QrisCPMService) StartPayment(ctx context.Context, req dto.StartBTNCPMPaymentRequest) (*dto.BTNCPMPaymentResponse, error) {
	if !s.cfg.BtnSnapCPMEnabled {
		return nil, fmt.Errorf("BTN CPM is disabled")
	}
	if req.OrderID == "" || len(req.OrderID) > 32 || req.CheckoutSnapshot.ID != req.OrderID {
		return nil, fmt.Errorf("orderId and matching checkoutSnapshot are required")
	}
	if strings.TrimSpace(req.QRContent) == "" || len(req.QRContent) > 512 {
		return nil, fmt.Errorf("qrContent must contain 1 to 512 bytes")
	}
	if strings.TrimSpace(req.ScannerDeviceID) == "" || len(req.ScannerDeviceID) > 20 {
		return nil, fmt.Errorf("scannerDeviceId must contain 1 to 20 characters")
	}
	if s.orderSvc == nil {
		return nil, fmt.Errorf("BTN checkout validator is not configured")
	}
	if req.CheckoutSnapshot.Total >= 10_000_000_000 {
		return nil, fmt.Errorf("checkout total exceeds the BTN CPM amount limit")
	}
	if err := s.orderSvc.ValidateBTNCheckoutSnapshot(req.CheckoutSnapshot); err != nil {
		return nil, err
	}
	amountCents, err := rupiahCents(req.CheckoutSnapshot.Total)
	if err != nil || amountCents <= 0 {
		return nil, fmt.Errorf("invalid checkout amount")
	}

	previous, err := s.repo.FindLatestByOrderIDAndFlowType(ctx, req.OrderID, model.QrisFlowCPM)
	if err != nil {
		return nil, fmt.Errorf("find previous CPM attempt: %w", err)
	}
	if previous != nil {
		switch previous.Status {
		case model.QrisTransactionStatusPending, model.QrisTransactionStatusUnknown, model.QrisTransactionStatusPaid, model.QrisTransactionStatusRefunded:
			return s.GetStatus(ctx, previous.ID)
		}
	}

	wib := time.FixedZone("WIB", 7*60*60)
	now := time.Now().In(wib)
	partnerReferenceNo, err := newBTNExternalID()
	if err != nil {
		return nil, fmt.Errorf("create BTN partner reference: %w", err)
	}
	transactionDateTime := now.Format(time.RFC3339)
	checkoutSnapshot, err := json.Marshal(req.CheckoutSnapshot)
	if err != nil {
		return nil, fmt.Errorf("encode checkout snapshot: %w", err)
	}
	expiresAt := now.Add(15 * time.Minute)
	tx := &model.QrisTransaction{
		ID: "BTN-CPM-" + uuid.NewString(), OrderID: req.OrderID,
		PartnerReferenceNo: partnerReferenceNo, Amount: float64(amountCents) / 100,
		TotalAmount: float64(amountCents) / 100, MerchantName: s.cfg.BtnSnapMerchantName,
		MerchantID: s.cfg.BtnSnapMerchantID, MerchantCity: strings.ToUpper(s.cfg.BtnSnapMerchantLocation),
		Currency: "IDR", Provider: model.QrisProviderBTNSnap, FlowType: model.QrisFlowCPM,
		TerminalID: s.cfg.BtnSnapTerminalID, MerchantPAN: s.cfg.BtnSnapMerchantPAN,
		ScannerDeviceID: req.ScannerDeviceID, TransactionDateTime: transactionDateTime,
		Status: model.QrisTransactionStatusPending, CheckoutSnapshot: checkoutSnapshot,
		ExpiresAt: expiresAt,
	}
	if err := s.repo.Create(ctx, tx); err != nil {
		return nil, fmt.Errorf("save CPM attempt before contacting BTN: %w", err)
	}

	decoded, err := s.btnClient.DecodeCPM(ctx, dto.BtnSnapCPMDecodeRequest{
		PartnerReferenceNo: partnerReferenceNo, QRContent: req.QRContent,
		MerchantID: s.cfg.BtnSnapMerchantID, ScanTime: transactionDateTime,
		AdditionalInfo: map[string]any{},
	})
	if err != nil {
		_, _ = s.repo.ApplyBTNStatus(ctx, tx.ID, model.QrisTransactionStatusFailed, "", "CPM Decode rejected", "", "", nil)
		return nil, fmt.Errorf("BTN CPM Decode failed: %w", err)
	}
	if decoded.PartnerReferenceNo != partnerReferenceNo {
		_, _ = s.repo.ApplyBTNStatus(ctx, tx.ID, model.QrisTransactionStatusFailed, "", "CPM Decode reference mismatch", "", "", nil)
		return nil, fmt.Errorf("BTN CPM Decode reference does not match request")
	}
	expiry, err := time.Parse(time.RFC3339, decoded.ExpiryDateTime)
	if err != nil || !expiry.After(now) {
		_, _ = s.repo.ApplyBTNStatus(ctx, tx.ID, model.QrisTransactionStatusFailed, "", "CPM QR expired or has invalid expiry", "", "", nil)
		return nil, fmt.Errorf("BTN CPM QR has expired or returned an invalid expiry")
	}
	if err := s.repo.UpdateCPMExpiry(ctx, tx.ID, expiry); err != nil {
		return nil, fmt.Errorf("save CPM QR expiry: %w", err)
	}
	tx.ExpiresAt = expiry

	payment, err := s.btnClient.PayCPM(ctx, dto.BtnSnapCPMPaymentRequest{
		PartnerReferenceNo: partnerReferenceNo, QRContent: req.QRContent,
		Amount:     dto.BtnSnapAmount{Value: formatBTNAmount(amountCents), Currency: "IDR"},
		MerchantID: s.cfg.BtnSnapMerchantID, MerchantName: s.cfg.BtnSnapMerchantName,
		MerchantLocation: strings.ToUpper(s.cfg.BtnSnapMerchantLocation), TerminalID: s.cfg.BtnSnapTerminalID,
		Title: "POS Payment", ExpiryTime: expiry.In(wib).Format("2006-01-02T15:04:05-07:00"),
		ScannerInfo:    dto.BtnSnapScannerInfo{DeviceID: req.ScannerDeviceID},
		AdditionalInfo: map[string]any{"transactionDateTime": transactionDateTime, "merchantPAN": s.cfg.BtnSnapMerchantPAN},
	})
	if err != nil {
		status := model.QrisTransactionStatusUnknown
		if isCPMRejected(err) {
			status = model.QrisTransactionStatusFailed
		}
		_, updateErr := s.repo.ApplyBTNStatus(ctx, tx.ID, status, "", "CPM Payment result requires reconciliation", "", "", nil)
		if updateErr != nil {
			err = fmt.Errorf("%w (failed to save CPM outcome: %v)", err, updateErr)
		}
		if status == model.QrisTransactionStatusUnknown {
			result, queryErr := s.GetStatus(ctx, tx.ID)
			if queryErr != nil {
				return nil, &BtnSnapOutcomeUnknownError{Cause: fmt.Errorf("payment: %v; reconciliation: %w", err, queryErr)}
			}
			return result, nil
		}
		return nil, err
	}
	if payment.PartnerReferenceNo != partnerReferenceNo {
		_, _ = s.repo.ApplyBTNStatus(ctx, tx.ID, model.QrisTransactionStatusUnknown, "", "CPM Payment reference mismatch", payment.ReferenceNo, "", nil)
		return nil, &BtnSnapOutcomeUnknownError{Cause: fmt.Errorf("BTN CPM Payment response reference does not match request")}
	}
	if payment.ReferenceNo != "" {
		if err := s.repo.UpdateCPMPaymentReference(ctx, tx.ID, payment.ReferenceNo); err != nil {
			return nil, &BtnSnapOutcomeUnknownError{Cause: fmt.Errorf("save BTN CPM reference: %w", err)}
		}
		tx.BankReferenceNo = payment.ReferenceNo
	}
	return s.GetStatus(ctx, tx.ID)
}

func (s *QrisCPMService) GetStatus(ctx context.Context, transactionID string) (*dto.BTNCPMPaymentResponse, error) {
	tx, err := s.repo.FindByID(ctx, transactionID)
	if err != nil {
		return nil, err
	}
	if tx.Provider != model.QrisProviderBTNSnap || tx.FlowType != model.QrisFlowCPM {
		return nil, fmt.Errorf("BTN CPM transaction not found")
	}

	query, err := s.btnClient.QueryCPM(ctx, dto.BtnSnapCPMQueryRequest{
		OriginalPartnerReferenceNo: tx.PartnerReferenceNo, OriginalReferenceNo: tx.BankReferenceNo,
		MerchantID: tx.MerchantID, TerminalID: tx.TerminalID, MerchantPAN: tx.MerchantPAN,
		AdditionalInfo: map[string]any{
			"terminalId": tx.TerminalID, "merchantPAN": tx.MerchantPAN,
			"transactionDateTime": tx.TransactionDateTime,
			"scannerInfo":         map[string]string{"deviceId": tx.ScannerDeviceID},
		},
	})
	if err != nil {
		if tx.Status != model.QrisTransactionStatusPaid && tx.Status != model.QrisTransactionStatusRefunded {
			_, _ = s.repo.ApplyBTNStatus(ctx, tx.ID, model.QrisTransactionStatusUnknown, "", "CPM Query result requires reconciliation", "", "", nil)
		}
		return nil, &BtnSnapOutcomeUnknownError{Cause: err}
	}
	if query.OriginalPartnerReferenceNo != tx.PartnerReferenceNo ||
		(query.OriginalReferenceNo != "" && tx.BankReferenceNo != "" && query.OriginalReferenceNo != tx.BankReferenceNo) {
		_, _ = s.repo.ApplyBTNStatus(ctx, tx.ID, model.QrisTransactionStatusUnknown, "", "CPM Query reference mismatch", "", "", nil)
		return nil, &BtnSnapOutcomeUnknownError{Cause: fmt.Errorf("BTN CPM Query reference does not match transaction")}
	}

	status := model.QrisTransactionStatusUnknown
	var paidAt *time.Time
	switch query.LatestTransactionStatus {
	case "00":
		if query.OriginalReferenceNo == "" || query.AdditionalInfo.MerchantPAN != tx.MerchantPAN ||
			query.AdditionalInfo.TerminalID != tx.TerminalID ||
			(query.AdditionalInfo.ScannerInfo.DeviceID != "" && query.AdditionalInfo.ScannerInfo.DeviceID != tx.ScannerDeviceID) ||
			(query.AdditionalInfo.DeviceID != "" && query.AdditionalInfo.DeviceID != tx.ScannerDeviceID) ||
			query.AdditionalInfo.Amount.Currency != "IDR" || !validBTNAmount(query.AdditionalInfo.Amount.Value) {
			_, _ = s.repo.ApplyBTNStatus(ctx, tx.ID, model.QrisTransactionStatusUnknown, "", "CPM Query success response mismatch", "", "", nil)
			return nil, &BtnSnapOutcomeUnknownError{Cause: fmt.Errorf("BTN CPM success response is missing or mismatches payment data")}
		}
		paidCents, parseErr := strconv.ParseInt(strings.ReplaceAll(query.AdditionalInfo.Amount.Value, ".", ""), 10, 64)
		if parseErr != nil || paidCents != int64(math.Round(tx.TotalAmount*100)) {
			_, _ = s.repo.ApplyBTNStatus(ctx, tx.ID, model.QrisTransactionStatusUnknown, "", "CPM Query amount mismatch", "", "", nil)
			return nil, &BtnSnapOutcomeUnknownError{Cause: fmt.Errorf("BTN CPM payment amount does not match checkout")}
		}
		parsed, parseErr := time.Parse(time.RFC3339, query.PaidTime)
		if parseErr != nil || query.PaidTime == "" {
			_, _ = s.repo.ApplyBTNStatus(ctx, tx.ID, model.QrisTransactionStatusUnknown, "", "CPM Query paid time invalid", "", "", nil)
			return nil, &BtnSnapOutcomeUnknownError{Cause: fmt.Errorf("BTN CPM success response is missing a valid paidTime")}
		}
		status, paidAt = model.QrisTransactionStatusPaid, &parsed
	case "01", "02", "03":
		status = model.QrisTransactionStatusPending
	case "04":
		status = model.QrisTransactionStatusRefunded
	case "05":
		status = model.QrisTransactionStatusCancelled
	case "06":
		status = model.QrisTransactionStatusFailed
	case "07":
		// CPM not-found is not proof of expiry or a safe signal to initiate again.
	}
	updated, err := s.repo.ApplyBTNStatus(ctx, tx.ID, status, query.LatestTransactionStatus,
		query.TransactionStatusDesc, query.OriginalReferenceNo, "", paidAt)
	if err != nil {
		return nil, err
	}
	if updated && status == model.QrisTransactionStatusRefunded {
		if refund, findErr := s.repo.FindRefundByTransactionID(ctx, tx.ID); findErr == nil && refund != nil {
			refund.Status = model.QrisRefunded
			_ = s.repo.UpdateRefund(ctx, refund)
		}
	}
	if !updated {
		fresh, findErr := s.repo.FindByID(ctx, tx.ID)
		if findErr == nil {
			tx = fresh
		}
	} else {
		tx.Status, tx.BankStatus, tx.BankStatusDescription = status, query.LatestTransactionStatus, query.TransactionStatusDesc
		if query.OriginalReferenceNo != "" {
			tx.BankReferenceNo = query.OriginalReferenceNo
		}
		if paidAt != nil {
			tx.BankPaidAt = paidAt
		}
	}
	return cpmResponse(tx), nil
}

func (s *QrisCPMService) GetByOrderID(ctx context.Context, orderID string) (*dto.BTNCPMPaymentResponse, error) {
	tx, err := s.repo.FindLatestByOrderIDAndFlowType(ctx, orderID, model.QrisFlowCPM)
	if err != nil {
		return nil, err
	}
	if tx == nil {
		return nil, fmt.Errorf("BTN CPM transaction not found")
	}
	return s.GetStatus(ctx, tx.ID)
}

func (s *QrisCPMService) Cancel(ctx context.Context, transactionID, reason string) (*dto.BTNCPMPaymentResponse, error) {
	if len(reason) > 20 {
		return nil, fmt.Errorf("cancel reason must be at most 20 characters")
	}
	tx, err := s.repo.FindByID(ctx, transactionID)
	if err != nil {
		return nil, err
	}
	if tx.Provider != model.QrisProviderBTNSnap || tx.FlowType != model.QrisFlowCPM {
		return nil, fmt.Errorf("BTN CPM transaction not found")
	}
	current, err := s.GetStatus(ctx, transactionID)
	if err != nil {
		return nil, err
	}
	if current.Status != model.QrisTransactionStatusPending {
		return current, fmt.Errorf("BTN CPM transaction is not pending")
	}
	now := time.Now().In(time.FixedZone("WIB", 7*60*60))
	claimed, err := s.repo.MarkCPMCancelRequested(ctx, tx.ID, reason, now)
	if err != nil {
		return nil, err
	}
	if !claimed {
		return nil, &BtnSnapOutcomeUnknownError{Cause: fmt.Errorf("CPM cancel was already requested; query the original transaction")}
	}
	cancelResult, err := s.btnClient.CancelCPM(ctx, dto.BtnSnapCPMCancelRequest{
		OriginalPartnerReferenceNo: tx.PartnerReferenceNo, OriginalReferenceNo: tx.BankReferenceNo,
		MerchantID: tx.MerchantID,
		Amount:     dto.BtnSnapAmount{Value: formatBTNAmount(int64(math.Round(tx.TotalAmount * 100))), Currency: "IDR"},
		Reason:     reason,
		AdditionalInfo: map[string]any{
			"terminalId": tx.TerminalID, "scannerInfo": map[string]string{"deviceId": tx.ScannerDeviceID},
			"cancelTime": now.Format(time.RFC3339), "transactionDateTime": tx.TransactionDateTime,
		},
	})
	if err != nil {
		_, _ = s.repo.ApplyBTNStatus(ctx, tx.ID, model.QrisTransactionStatusUnknown, "", "CPM Cancel result requires reconciliation", "", "", nil)
		return nil, &BtnSnapOutcomeUnknownError{Cause: err}
	}
	if cancelResult.OriginalPartnerReferenceNo != tx.PartnerReferenceNo ||
		(cancelResult.OriginalReferenceNo != "" && tx.BankReferenceNo != "" && cancelResult.OriginalReferenceNo != tx.BankReferenceNo) {
		_, _ = s.repo.ApplyBTNStatus(ctx, tx.ID, model.QrisTransactionStatusUnknown, "", "CPM Cancel reference mismatch", "", "", nil)
		return nil, &BtnSnapOutcomeUnknownError{Cause: fmt.Errorf("BTN CPM Cancel response reference does not match transaction")}
	}
	result, err := s.GetStatus(ctx, transactionID)
	if err != nil {
		return nil, err
	}
	if result.Status != model.QrisTransactionStatusCancelled {
		return result, &BtnSnapOutcomeUnknownError{Cause: fmt.Errorf("BTN accepted CPM Cancel but Query has not confirmed cancellation")}
	}
	return result, nil
}

func (s *QrisCPMService) Refund(ctx context.Context, transactionID, reason string) (*dto.BTNCPMRefundResponse, error) {
	if len(reason) > 20 {
		return nil, fmt.Errorf("refund reason must be at most 20 characters")
	}
	tx, err := s.repo.FindByID(ctx, transactionID)
	if err != nil {
		return nil, err
	}
	if tx.Provider != model.QrisProviderBTNSnap || tx.FlowType != model.QrisFlowCPM {
		return nil, fmt.Errorf("BTN CPM transaction not found")
	}
	current, err := s.GetStatus(ctx, transactionID)
	if err != nil {
		return nil, err
	}
	if current.Status == model.QrisTransactionStatusRefunded {
		refund, _ := s.repo.FindRefundByTransactionID(ctx, tx.ID)
		if refund != nil {
			return &dto.BTNCPMRefundResponse{TransactionID: tx.ID, PartnerRefundNo: refund.PartnerRefundNo, RefundNo: refund.RefundNo, Status: refund.Status, Amount: refund.Amount}, nil
		}
	}
	if current.Status != model.QrisTransactionStatusPaid {
		return nil, fmt.Errorf("only a verified paid CPM transaction can be refunded")
	}
	previous, err := s.repo.FindRefundByTransactionID(ctx, tx.ID)
	if err != nil {
		return nil, err
	}
	if previous != nil {
		if previous.Status == model.QrisRefunded {
			return &dto.BTNCPMRefundResponse{TransactionID: tx.ID, PartnerRefundNo: previous.PartnerRefundNo, RefundNo: previous.RefundNo, Status: previous.Status, Amount: previous.Amount}, nil
		}
		return nil, &BtnSnapOutcomeUnknownError{Cause: fmt.Errorf("a CPM refund attempt already exists; reconcile it before retrying")}
	}
	partnerRefundNo, err := newBTNExternalID()
	if err != nil {
		return nil, err
	}
	refund := &model.QrisRefund{
		ID: "BTN-REFUND-" + uuid.NewString(), TransactionID: tx.ID,
		PartnerRefundNo: partnerRefundNo, Amount: tx.TotalAmount,
		Status: model.QrisRefundPending, Reason: reason,
	}
	if err := s.repo.CreateRefund(ctx, refund); err != nil {
		return nil, fmt.Errorf("save CPM refund attempt before contacting BTN: %w", err)
	}
	now := time.Now().In(time.FixedZone("WIB", 7*60*60))
	bankRefund, err := s.btnClient.RefundCPM(ctx, dto.BtnSnapCPMRefundRequest{
		OriginalPartnerReferenceNo: tx.PartnerReferenceNo, OriginalReferenceNo: tx.BankReferenceNo,
		MerchantID: tx.MerchantID, PartnerRefundNo: partnerRefundNo,
		RefundAmount: dto.BtnSnapAmount{Value: formatBTNAmount(int64(math.Round(tx.TotalAmount * 100))), Currency: "IDR"},
		Reason:       reason,
		AdditionalInfo: map[string]any{
			"terminalId": tx.TerminalID, "scannerInfo": map[string]string{"deviceId": tx.ScannerDeviceID},
			"refundDate": now.Format(time.RFC3339),
		},
	})
	if err != nil {
		refund.Status = model.QrisRefundUnknown
		_ = s.repo.UpdateRefund(ctx, refund)
		return nil, &BtnSnapOutcomeUnknownError{Cause: err}
	}
	if bankRefund.PartnerRefundNo != partnerRefundNo || bankRefund.OriginalPartnerReferenceNo != tx.PartnerReferenceNo ||
		(bankRefund.OriginalReferenceNo != "" && tx.BankReferenceNo != "" && bankRefund.OriginalReferenceNo != tx.BankReferenceNo) || bankRefund.RefundNo == "" ||
		len(bankRefund.RefundNo) > 16 || bankRefund.RefundAmount.Currency != "IDR" || !validBTNAmount(bankRefund.RefundAmount.Value) {
		refund.Status = model.QrisRefundUnknown
		_ = s.repo.UpdateRefund(ctx, refund)
		return nil, &BtnSnapOutcomeUnknownError{Cause: fmt.Errorf("BTN CPM Refund response did not match request")}
	}
	refundedCents, parseErr := strconv.ParseInt(strings.ReplaceAll(bankRefund.RefundAmount.Value, ".", ""), 10, 64)
	refundTime, timeErr := time.Parse(time.RFC3339, bankRefund.RefundTime)
	if parseErr != nil || refundedCents != int64(math.Round(tx.TotalAmount*100)) || timeErr != nil {
		refund.Status = model.QrisRefundUnknown
		_ = s.repo.UpdateRefund(ctx, refund)
		return nil, &BtnSnapOutcomeUnknownError{Cause: fmt.Errorf("BTN CPM Refund amount or time did not match")}
	}
	refund.Status, refund.RefundNo, refund.RefundedAt = model.QrisRefunded, bankRefund.RefundNo, &refundTime
	if err := s.repo.UpdateRefund(ctx, refund); err != nil {
		return nil, &BtnSnapOutcomeUnknownError{Cause: fmt.Errorf("save confirmed CPM refund: %w", err)}
	}
	if _, err := s.repo.ApplyBTNStatus(ctx, tx.ID, model.QrisTransactionStatusRefunded, "04", "Refunded", tx.BankReferenceNo, "", nil); err != nil {
		return nil, &BtnSnapOutcomeUnknownError{Cause: fmt.Errorf("save CPM refund status: %w", err)}
	}
	return &dto.BTNCPMRefundResponse{TransactionID: tx.ID, PartnerRefundNo: partnerRefundNo, RefundNo: bankRefund.RefundNo, Status: refund.Status, Amount: refund.Amount}, nil
}

func cpmResponse(tx *model.QrisTransaction) *dto.BTNCPMPaymentResponse {
	return &dto.BTNCPMPaymentResponse{
		TransactionID: tx.ID, OrderID: tx.OrderID, PartnerReferenceNo: tx.PartnerReferenceNo,
		BankReferenceNo: tx.BankReferenceNo, Status: tx.Status, BankStatus: tx.BankStatus,
		ExpiresAt: tx.ExpiresAt,
	}
}

func formatBTNAmount(cents int64) string {
	return fmt.Sprintf("%d.%02d", cents/100, cents%100)
}

func isCPMRejected(err error) bool {
	var rejected *BtnSnapCPMRejectedError
	return errors.As(err, &rejected)
}
