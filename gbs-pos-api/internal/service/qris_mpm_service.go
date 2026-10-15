package service

import (
	"context"
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

type QrisMPMService struct {
	cfg       *config.Config
	repo      *repository.QrisTransactionRepository
	direct    *QrisDirectService
	btnClient *BtnSnapClient
}

func NewQrisMPMService(cfg *config.Config, repo *repository.QrisTransactionRepository, direct *QrisDirectService) *QrisMPMService {
	return &QrisMPMService{cfg: cfg, repo: repo, direct: direct, btnClient: NewBtnSnapClient(cfg)}
}

func (s *QrisMPMService) Refund(ctx context.Context, transactionID, reason string) (*dto.BTNMPMRefundResponse, error) {
	if !s.cfg.BtnSnapEnabled {
		return nil, fmt.Errorf("BTN SNAP is disabled")
	}
	if len(reason) == 0 || len(reason) > 256 {
		return nil, fmt.Errorf("refund reason must contain 1 to 256 bytes")
	}
	tx, err := s.repo.FindByID(ctx, transactionID)
	if err != nil {
		return nil, err
	}
	if tx.Provider != model.QrisProviderBTNSnap || (tx.FlowType != "" && tx.FlowType != model.QrisFlowMPM) {
		return nil, fmt.Errorf("BTN MPM transaction not found")
	}
	if s.direct == nil {
		return nil, fmt.Errorf("BTN MPM status verifier is not configured")
	}
	status, err := s.direct.GetTransactionStatus(ctx, tx.ID)
	if err != nil {
		return nil, err
	}
	if status.Status == model.QrisTransactionStatusRefunded {
		return s.refundResponse(ctx, tx)
	}
	if status.Status != model.QrisTransactionStatusPaid {
		return nil, fmt.Errorf("only a verified paid MPM transaction can be refunded")
	}
	previous, err := s.repo.FindRefundByTransactionID(ctx, tx.ID)
	if err != nil {
		return nil, err
	}
	if previous != nil {
		if previous.Status == model.QrisRefunded {
			return mpmRefundResponse(tx.ID, previous), nil
		}
		return nil, &BtnSnapOutcomeUnknownError{Cause: fmt.Errorf("an MPM refund attempt already exists; query the original payment before retrying")}
	}
	// Reload Query-updated reference and paid time before recording the refund attempt.
	tx, err = s.repo.FindByID(ctx, transactionID)
	if err != nil {
		return nil, err
	}
	if tx.BankReferenceNo == "" || tx.BankPaidAt == nil {
		return nil, fmt.Errorf("verified MPM transaction is missing its bank reference or paid time")
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
		return nil, fmt.Errorf("save MPM refund attempt before contacting BTN: %w", err)
	}
	wib := time.FixedZone("WIB", 7*60*60)
	generateTime := tx.GenerateTime
	if generateTime == "" {
		generateTime = tx.CreatedAt.In(wib).Format(time.RFC3339)
	}
	response, err := s.btnClient.RefundMPM(ctx, dto.BtnSnapMPMRefundRequest{
		MerchantID: tx.MerchantID, OriginalPartnerReferenceNo: tx.PartnerReferenceNo,
		OriginalReferenceNo: tx.BankReferenceNo, PartnerRefundNo: partnerRefundNo,
		RefundAmount: dto.BtnSnapAmount{Value: formatBTNAmount(int64(math.Round(tx.TotalAmount * 100))), Currency: "IDR"},
		Reason:       reason,
		AdditionalInfo: map[string]any{
			"generateTime": generateTime, "transactionDate": tx.BankPaidAt.In(wib).Format(time.RFC3339), "terminalId": tx.TerminalID,
		},
	})
	if err != nil {
		refund.Status = model.QrisRefundUnknown
		_ = s.repo.UpdateRefund(ctx, refund)
		return nil, &BtnSnapOutcomeUnknownError{Cause: err}
	}
	if response.OriginalPartnerReferenceNo != tx.PartnerReferenceNo || response.OriginalReferenceNo != tx.BankReferenceNo ||
		response.PartnerRefundNo != partnerRefundNo || response.RefundNo == "" || len(response.RefundNo) > 16 ||
		response.RefundAmount.Currency != "IDR" || !validBTNAmount(response.RefundAmount.Value) {
		refund.Status = model.QrisRefundUnknown
		_ = s.repo.UpdateRefund(ctx, refund)
		return nil, &BtnSnapOutcomeUnknownError{Cause: fmt.Errorf("BTN MPM Refund response did not match request")}
	}
	refundedCents, parseErr := strconv.ParseInt(strings.ReplaceAll(response.RefundAmount.Value, ".", ""), 10, 64)
	refundTime, timeErr := time.Parse(time.RFC3339, response.RefundTime)
	if parseErr != nil || refundedCents != int64(math.Round(tx.TotalAmount*100)) || timeErr != nil {
		refund.Status = model.QrisRefundUnknown
		_ = s.repo.UpdateRefund(ctx, refund)
		return nil, &BtnSnapOutcomeUnknownError{Cause: fmt.Errorf("BTN MPM Refund amount or time did not match")}
	}
	refund.Status, refund.RefundNo, refund.RefundedAt = model.QrisRefunded, response.RefundNo, &refundTime
	if err := s.repo.UpdateRefund(ctx, refund); err != nil {
		return nil, &BtnSnapOutcomeUnknownError{Cause: fmt.Errorf("save confirmed MPM refund: %w", err)}
	}
	if _, err := s.repo.ApplyBTNStatus(ctx, tx.ID, model.QrisTransactionStatusRefunded, "04", "Refunded", tx.BankReferenceNo, "", nil); err != nil {
		return nil, &BtnSnapOutcomeUnknownError{Cause: fmt.Errorf("save MPM refund status: %w", err)}
	}
	return mpmRefundResponse(tx.ID, refund), nil
}

func (s *QrisMPMService) Report(ctx context.Context, req dto.BTNMPMReportRequest) (*dto.BtnSnapMPMReportResponse, error) {
	if !s.cfg.BtnSnapEnabled {
		return nil, fmt.Errorf("BTN SNAP is disabled")
	}
	if !validBTNPartnerReference(req.PartnerReferenceNo) {
		return nil, fmt.Errorf("partnerReferenceNo must be at most 16 alphanumeric characters")
	}
	from, err := time.Parse(time.RFC3339, req.FromDateTime)
	if err != nil {
		return nil, fmt.Errorf("fromDateTime must be RFC3339")
	}
	to, err := time.Parse(time.RFC3339, req.ToDateTime)
	if err != nil || to.Before(from) || to.After(time.Now().Add(2*time.Minute)) {
		return nil, fmt.Errorf("toDateTime must be valid, after fromDateTime, and not in the future")
	}
	if req.PageSize < 1 || req.PageSize > 99 || req.PageNumber < 1 || req.PageNumber > 99999 {
		return nil, fmt.Errorf("pageSize must be 1-99 and pageNumber must be 1-99999")
	}
	transactionType := req.TransactionType
	if transactionType == "" {
		transactionType = "QRMPM"
	}
	if transactionType != "QRMPM" && transactionType != "QRCPM" {
		return nil, fmt.Errorf("transactionType must be QRMPM or QRCPM")
	}
	wib := time.FixedZone("WIB", 7*60*60)
	return s.btnClient.ReportMPM(ctx, dto.BtnSnapMPMReportRequest{
		PartnerReferenceNo: req.PartnerReferenceNo,
		FromDateTime:       from.In(wib).Format("2006-01-02T15:04:05-07:00"),
		ToDateTime:         to.In(wib).Format("2006-01-02T15:04:05-07:00"),
		PageSize:           strconv.Itoa(req.PageSize), PageNumber: strconv.Itoa(req.PageNumber),
		AdditionalInfo: map[string]any{"transactionType": transactionType, "merchantId": s.cfg.BtnSnapMerchantID},
	})
}

func (s *QrisMPMService) refundResponse(ctx context.Context, tx *model.QrisTransaction) (*dto.BTNMPMRefundResponse, error) {
	refund, err := s.repo.FindRefundByTransactionID(ctx, tx.ID)
	if err != nil {
		return nil, err
	}
	if refund == nil {
		return nil, fmt.Errorf("bank reports this MPM payment refunded but no local refund record exists")
	}
	if refund.Status != model.QrisRefunded {
		refund.Status = model.QrisRefunded
		if err := s.repo.UpdateRefund(ctx, refund); err != nil {
			return nil, err
		}
	}
	return mpmRefundResponse(tx.ID, refund), nil
}

func mpmRefundResponse(transactionID string, refund *model.QrisRefund) *dto.BTNMPMRefundResponse {
	return &dto.BTNMPMRefundResponse{TransactionID: transactionID, PartnerRefundNo: refund.PartnerRefundNo, RefundNo: refund.RefundNo, Status: refund.Status, Amount: refund.Amount}
}

func validBTNPartnerReference(value string) bool {
	if len(value) > 16 {
		return false
	}
	for _, r := range value {
		if !(r >= '0' && r <= '9' || r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z') {
			return false
		}
	}
	return true
}
