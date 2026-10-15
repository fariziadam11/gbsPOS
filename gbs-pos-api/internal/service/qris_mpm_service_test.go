package service

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"gbs-pos-api/internal/dto"
	"gbs-pos-api/internal/model"
	"gbs-pos-api/internal/repository"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func TestQrisMPMRefundVerifiesAndDoesNotDuplicate(t *testing.T) {
	_, keyPath := writeTestRSAPrivateKey(t)
	queryCalls, refundCalls := 0, 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case btnTokenPath:
			_, _ = w.Write([]byte(`{"responseCode":"2007300","accessToken":"test-token","expiresIn":"900"}`))
		case btnQueryPath:
			queryCalls++
			bankStatus := "00"
			if refundCalls > 0 {
				bankStatus = "04"
			}
			_, _ = fmt.Fprintf(w, `{"responseCode":"2005100","originalReferenceNo":"mpm-bank-ref","originalPartnerReferenceNo":"mpm-partner-ref","serviceCode":"47","latestTransactionStatus":%q,"transactionStatusDesc":"status","paidTime":"2026-10-13T12:00:00+07:00","terminalId":"TERM001","amount":{"value":"12500.00","currency":"IDR"},"additionalInfo":{"merchantId":"MERCHANT01","merchantPAN":"merchant-pan"}}`, bankStatus)
		case btnMPMRefundPath:
			refundCalls++
			var req dto.BtnSnapMPMRefundRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				t.Errorf("decode MPM Refund request: %v", err)
			}
			if req.OriginalPartnerReferenceNo != "mpm-partner-ref" || req.OriginalReferenceNo != "mpm-bank-ref" ||
				req.RefundAmount.Value != "12500.00" || req.AdditionalInfo["generateTime"] != "2026-10-13T11:50:00+07:00" ||
				req.AdditionalInfo["transactionDate"] != "2026-10-13T12:00:00+07:00" {
				t.Errorf("unexpected MPM Refund request: %+v", req)
			}
			_, _ = fmt.Fprintf(w, `{"responseCode":"2007800","originalPartnerReferenceNo":"mpm-partner-ref","originalReferenceNo":"mpm-bank-ref","refundNo":"refund-bank-01","partnerRefundNo":%q,"refundAmount":{"value":"12500.00","currency":"IDR"},"refundTime":"2026-10-13T12:02:00+07:00"}`, req.PartnerRefundNo)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&model.QrisTransaction{}, &model.QrisRefund{}); err != nil {
		t.Fatal(err)
	}
	repo := repository.NewQrisTransactionRepository(db)
	tx := &model.QrisTransaction{
		ID: "BTN-MPM-1", OrderID: "ORDER-MPM-1", PartnerReferenceNo: "mpm-partner-ref",
		BankReferenceNo: "mpm-bank-ref", Amount: 12500, TotalAmount: 12500,
		Provider: model.QrisProviderBTNSnap, FlowType: model.QrisFlowMPM, MerchantID: "MERCHANT01",
		MerchantPAN: "merchant-pan", Currency: "IDR", TerminalID: "TERM001",
		GenerateTime: "2026-10-13T11:50:00+07:00", Status: model.QrisTransactionStatusPending,
		ExpiresAt: time.Now().Add(time.Hour),
	}
	if err := repo.Create(t.Context(), tx); err != nil {
		t.Fatal(err)
	}
	cfg := btnTestConfig(server.URL, keyPath)
	cfg.BtnSnapEnabled = true
	direct := NewQrisDirectService(cfg, repo, nil)
	service := NewQrisMPMService(cfg, repo, direct)
	refunded, err := service.Refund(t.Context(), tx.ID, "Customer complaint")
	if err != nil {
		t.Fatal(err)
	}
	if refunded.Status != model.QrisRefunded || refunded.Amount != 12500 || refundCalls != 1 {
		t.Fatalf("unexpected MPM refund result: %+v, refund calls %d", refunded, refundCalls)
	}
	retry, err := service.Refund(t.Context(), tx.ID, "Customer complaint")
	if err != nil || retry.PartnerRefundNo != refunded.PartnerRefundNo || refundCalls != 1 {
		t.Fatalf("refund retry was not idempotent: result=%+v error=%v calls=%d", retry, err, refundCalls)
	}
	if queryCalls < 2 {
		t.Fatalf("refund did not query the bank before/after reconciliation: queries %d", queryCalls)
	}
}

func TestQrisMPMReportUsesServerMerchantAndPagination(t *testing.T) {
	_, keyPath := writeTestRSAPrivateKey(t)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == btnTokenPath {
			_, _ = w.Write([]byte(`{"responseCode":"2007300","accessToken":"test-token","expiresIn":"900"}`))
			return
		}
		if r.URL.Path != btnMPMReportPath {
			http.NotFound(w, r)
			return
		}
		var req dto.BtnSnapMPMReportRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("decode MPM Report request: %v", err)
		}
		if req.PageSize != "10" || req.PageNumber != "2" || req.AdditionalInfo["transactionType"] != "QRCPM" || req.AdditionalInfo["merchantId"] != "MERCHANT01" {
			t.Errorf("unexpected Report request: %+v", req)
		}
		_, _ = w.Write([]byte(`{"responseCode":"2009900","responseMessage":"Successful","detailData":{"dateTime":"2026-10-13T11:00:00+07:00","paidTime":"2026-10-13T11:05:00+07:00","amount":{"value":"12500.00","currency":"IDR"},"status":"Transaction Success","type":"PAYMENT","additionalInfo":{"transactionMethod":"MPM","channelResponseCode":"00"}},"additionalInfo":{"transaction":"MPM","merchantId":"MERCHANT01","totalData":"21"}}`))
	}))
	defer server.Close()
	cfg := btnTestConfig(server.URL, keyPath)
	cfg.BtnSnapEnabled = true
	cfg.BtnSnapMerchantID = "MERCHANT01"
	service := NewQrisMPMService(cfg, nil, nil)
	result, err := service.Report(t.Context(), dto.BTNMPMReportRequest{
		FromDateTime: "2026-10-13T10:00:00+07:00", ToDateTime: "2026-10-13T12:00:00+07:00",
		PageSize: 10, PageNumber: 2, TransactionType: "QRCPM",
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.AdditionalInfo.TotalData != "21" || result.DetailData.Amount.Value != "12500.00" {
		t.Fatalf("unexpected MPM Report response: %+v", result)
	}
}
