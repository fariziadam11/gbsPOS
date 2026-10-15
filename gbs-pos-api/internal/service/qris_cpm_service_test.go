package service

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"gbs-pos-api/internal/config"
	"gbs-pos-api/internal/dto"
	"gbs-pos-api/internal/model"
	"gbs-pos-api/internal/repository"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func TestQrisCPMStartVerifiesPaymentAndGuardsQRISScanOrder(t *testing.T) {
	_, keyPath := writeTestRSAPrivateKey(t)
	partnerReference := ""
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case btnTokenPath:
			_, _ = w.Write([]byte(`{"responseCode":"2007300","accessToken":"test-token","expiresIn":"900"}`))
		case btnCPMDecodePath:
			var req dto.BtnSnapCPMDecodeRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				t.Errorf("decode CPM Decode request: %v", err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			partnerReference = req.PartnerReferenceNo
			if len(partnerReference) != 16 || req.QRContent != "QR-PAYLOAD-case-sensitive" {
				t.Errorf("unexpected Decode request: %+v", req)
			}
			_, _ = fmt.Fprintf(w, `{"responseCode":"2000000","partnerReferenceNo":%q,"expiryDateTime":%q,"customerPAN":"must-not-be-stored"}`,
				partnerReference, time.Now().Add(10*time.Minute).Format(time.RFC3339))
		case btnCPMPaymentPath:
			var req dto.BtnSnapCPMPaymentRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				t.Errorf("decode CPM Payment request: %v", err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			if req.PartnerReferenceNo != partnerReference || req.Amount.Value != "12500.00" ||
				req.MerchantLocation != "JAKARTA PUSAT" || req.AdditionalInfo["merchantPAN"] != "1234567890123456789" ||
				req.ScannerInfo.DeviceID != "scanner-01" {
				t.Errorf("unexpected CPM Payment request: %+v", req)
			}
			_, _ = fmt.Fprintf(w, `{"responseCode":"2006000","referenceNo":"cpm-bank-ref","partnerReferenceNo":%q}`, partnerReference)
		case btnCPMQueryPath:
			var req dto.BtnSnapCPMQueryRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				t.Errorf("decode CPM Query request: %v", err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			if req.OriginalPartnerReferenceNo != partnerReference || req.MerchantPAN != "1234567890123456789" ||
				req.AdditionalInfo["transactionDateTime"] == nil {
				t.Errorf("unexpected CPM Query request: %+v", req)
			}
			_, _ = fmt.Fprintf(w, `{"responseCode":"2006100","originalReferenceNo":"cpm-bank-ref","originalPartnerReferenceNo":%q,"latestTransactionStatus":"00","transactionStatusDesc":"Success","paidTime":"2026-10-13T12:00:00+07:00","additionalInfo":{"merchantPAN":"1234567890123456789","terminalId":"TERM001","scannerInfo":{"deviceId":"scanner-01"},"amount":{"value":"12500.00","currency":"IDR"}}}`, partnerReference)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	db, repo, orderService := setupCPMTestDB(t)
	cfg := cpmTestConfig(server.URL, keyPath)
	service := NewQrisCPMService(cfg, repo, orderService)
	snapshot := cpmTestSnapshot("ORDER-CPM-1")
	result, err := service.StartPayment(t.Context(), dto.StartBTNCPMPaymentRequest{
		OrderID: snapshot.ID, QRContent: "QR-PAYLOAD-case-sensitive", ScannerDeviceID: "scanner-01", CheckoutSnapshot: snapshot,
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != model.QrisTransactionStatusPaid || result.BankReferenceNo != "cpm-bank-ref" {
		t.Fatalf("unexpected CPM result: %+v", result)
	}
	var order model.Order
	order.ID, order.PaymentMethod, order.TransactionID = snapshot.ID, "QRIS_SCAN", result.TransactionID
	order.Subtotal, order.Total = snapshot.Subtotal, snapshot.Total
	order.Items = []model.OrderItem{{ProductID: 1, ProductName: "Product", ProductPrice: 12500, Qty: 1, Subtotal: 12500}}
	if err := orderService.validateBTNQRISPayment(&order); err != nil {
		t.Fatalf("verified CPM order rejected: %v", err)
	}
	tx, err := repo.FindByID(t.Context(), result.TransactionID)
	if err != nil {
		t.Fatal(err)
	}
	if tx.FlowType != model.QrisFlowCPM || tx.DynamicQrisString != "" || tx.MerchantPAN != cfg.BtnSnapMerchantPAN {
		t.Fatalf("unexpected persisted CPM transaction: %+v", tx)
	}
	if err := db.First(&model.QrisRefund{}).Error; err == nil {
		t.Fatal("unexpected refund record")
	}
}

func TestQrisCPMUnknownDoesNotStartDuplicatePayment(t *testing.T) {
	_, keyPath := writeTestRSAPrivateKey(t)
	partnerReference := ""
	paymentCalls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case btnTokenPath:
			_, _ = w.Write([]byte(`{"responseCode":"2007300","accessToken":"test-token","expiresIn":"900"}`))
		case btnCPMDecodePath:
			var req dto.BtnSnapCPMDecodeRequest
			_ = json.NewDecoder(r.Body).Decode(&req)
			partnerReference = req.PartnerReferenceNo
			_, _ = fmt.Fprintf(w, `{"responseCode":"2000000","partnerReferenceNo":%q,"expiryDateTime":%q}`, partnerReference, time.Now().Add(time.Hour).Format(time.RFC3339))
		case btnCPMPaymentPath:
			paymentCalls++
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(`{"responseCode":"5006000","responseMessage":"General Error"}`))
		case btnCPMQueryPath:
			_, _ = fmt.Fprintf(w, `{"responseCode":"2006100","originalPartnerReferenceNo":%q,"latestTransactionStatus":"07","transactionStatusDesc":"Not found"}`, partnerReference)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	_, repo, orderService := setupCPMTestDB(t)
	service := NewQrisCPMService(cpmTestConfig(server.URL, keyPath), repo, orderService)
	snapshot := cpmTestSnapshot("ORDER-CPM-UNKNOWN")
	req := dto.StartBTNCPMPaymentRequest{OrderID: snapshot.ID, QRContent: "QR-PAYLOAD", ScannerDeviceID: "scanner-01", CheckoutSnapshot: snapshot}
	first, err := service.StartPayment(t.Context(), req)
	if err != nil {
		t.Fatal(err)
	}
	if first.Status != model.QrisTransactionStatusUnknown {
		t.Fatalf("not-found CPM result = %q, want UNKNOWN", first.Status)
	}
	second, err := service.StartPayment(t.Context(), req)
	if err != nil {
		t.Fatal(err)
	}
	if second.TransactionID != first.TransactionID || paymentCalls != 1 {
		t.Fatalf("unknown attempt was duplicated: first=%+v second=%+v paymentCalls=%d", first, second, paymentCalls)
	}
}

func TestQrisCPMRefundIsFullAndIdempotent(t *testing.T) {
	_, keyPath := writeTestRSAPrivateKey(t)
	partnerReference := ""
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case btnTokenPath:
			_, _ = w.Write([]byte(`{"responseCode":"2007300","accessToken":"test-token","expiresIn":"900"}`))
		case btnCPMDecodePath:
			var req dto.BtnSnapCPMDecodeRequest
			_ = json.NewDecoder(r.Body).Decode(&req)
			partnerReference = req.PartnerReferenceNo
			_, _ = fmt.Fprintf(w, `{"responseCode":"2000000","partnerReferenceNo":%q,"expiryDateTime":%q}`, partnerReference, time.Now().Add(time.Hour).Format(time.RFC3339))
		case btnCPMPaymentPath:
			_, _ = fmt.Fprintf(w, `{"responseCode":"2006000","referenceNo":"refund-source","partnerReferenceNo":%q}`, partnerReference)
		case btnCPMQueryPath:
			_, _ = fmt.Fprintf(w, `{"responseCode":"2006100","originalReferenceNo":"refund-source","originalPartnerReferenceNo":%q,"latestTransactionStatus":"00","paidTime":"2026-10-13T12:00:00+07:00","additionalInfo":{"merchantPAN":"1234567890123456789","terminalId":"TERM001","scannerInfo":{"deviceId":"scanner-01"},"amount":{"value":"12500.00","currency":"IDR"}}}`, partnerReference)
		case btnCPMRefundPath:
			var req dto.BtnSnapCPMRefundRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				t.Errorf("decode refund request: %v", err)
			}
			if req.RefundAmount.Value != "12500.00" || req.PartnerRefundNo == "" {
				t.Errorf("refund was not full or lacked reference: %+v", req)
			}
			_, _ = fmt.Fprintf(w, `{"responseCode":"2008000","originalReferenceNo":"refund-source","originalPartnerReferenceNo":%q,"refundNo":"refund-bank-1","partnerRefundNo":%q,"refundAmount":{"value":"12500.00","currency":"IDR"},"refundTime":"2026-10-13T12:01:00+07:00"}`, partnerReference, req.PartnerRefundNo)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	_, repo, orderService := setupCPMTestDB(t)
	service := NewQrisCPMService(cpmTestConfig(server.URL, keyPath), repo, orderService)
	snapshot := cpmTestSnapshot("ORDER-CPM-REFUND")
	started, err := service.StartPayment(t.Context(), dto.StartBTNCPMPaymentRequest{
		OrderID: snapshot.ID, QRContent: "QR-PAYLOAD", ScannerDeviceID: "scanner-01", CheckoutSnapshot: snapshot,
	})
	if err != nil {
		t.Fatal(err)
	}
	refunded, err := service.Refund(t.Context(), started.TransactionID, "Return")
	if err != nil {
		t.Fatal(err)
	}
	if refunded.Status != model.QrisRefunded || refunded.Amount != 12500 {
		t.Fatalf("unexpected refund result: %+v", refunded)
	}
	retry, err := service.Refund(t.Context(), started.TransactionID, "Return")
	if err != nil || retry.PartnerRefundNo != refunded.PartnerRefundNo {
		t.Fatalf("refund retry was not idempotent: result=%+v error=%v", retry, err)
	}
	stored, err := repo.FindByID(t.Context(), started.TransactionID)
	if err != nil || stored.Status != model.QrisTransactionStatusRefunded {
		t.Fatalf("refund status not saved: transaction=%+v error=%v", stored, err)
	}
}

func TestQrisCPMCancelRequiresBankConfirmationAndIsNotRepeated(t *testing.T) {
	_, keyPath := writeTestRSAPrivateKey(t)
	partnerReference := ""
	queryCalls, cancelCalls := 0, 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case btnTokenPath:
			_, _ = w.Write([]byte(`{"responseCode":"2007300","accessToken":"test-token","expiresIn":"900"}`))
		case btnCPMDecodePath:
			var req dto.BtnSnapCPMDecodeRequest
			_ = json.NewDecoder(r.Body).Decode(&req)
			partnerReference = req.PartnerReferenceNo
			_, _ = fmt.Fprintf(w, `{"responseCode":"2000000","partnerReferenceNo":%q,"expiryDateTime":%q}`, partnerReference, time.Now().Add(time.Hour).Format(time.RFC3339))
		case btnCPMPaymentPath:
			_, _ = fmt.Fprintf(w, `{"responseCode":"2006000","referenceNo":"cancel-source","partnerReferenceNo":%q}`, partnerReference)
		case btnCPMQueryPath:
			queryCalls++
			bankStatus := "03"
			if queryCalls >= 3 {
				bankStatus = "05"
			}
			_, _ = fmt.Fprintf(w, `{"responseCode":"2006100","originalReferenceNo":"cancel-source","originalPartnerReferenceNo":%q,"latestTransactionStatus":%q,"transactionStatusDesc":"status"}`, partnerReference, bankStatus)
		case btnCPMCancelPath:
			cancelCalls++
			_, _ = fmt.Fprintf(w, `{"responseCode":"2006200","originalReferenceNo":"cancel-source","originalPartnerReferenceNo":%q}`, partnerReference)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	_, repo, orderService := setupCPMTestDB(t)
	service := NewQrisCPMService(cpmTestConfig(server.URL, keyPath), repo, orderService)
	snapshot := cpmTestSnapshot("ORDER-CPM-CANCEL")
	started, err := service.StartPayment(t.Context(), dto.StartBTNCPMPaymentRequest{
		OrderID: snapshot.ID, QRContent: "QR-PAYLOAD", ScannerDeviceID: "scanner-01", CheckoutSnapshot: snapshot,
	})
	if err != nil {
		t.Fatal(err)
	}
	if started.Status != model.QrisTransactionStatusPending {
		t.Fatalf("expected pending before cancel, got %+v", started)
	}
	cancelled, err := service.Cancel(t.Context(), started.TransactionID, "Suspect")
	if err != nil || cancelled.Status != model.QrisTransactionStatusCancelled {
		t.Fatalf("cancel was not verified by Query: result=%+v error=%v", cancelled, err)
	}
	if _, err := service.Cancel(t.Context(), started.TransactionID, "Suspect"); err == nil {
		t.Fatal("terminal transaction accepted another cancel")
	}
	if cancelCalls != 1 {
		t.Fatalf("bank cancel calls = %d, want 1", cancelCalls)
	}
}

func TestQRISScanCannotFinalizeWhenBTNCPMIsDisabled(t *testing.T) {
	service := &OrderService{}
	if err := service.validateBTNQRISPayment(&model.Order{PaymentMethod: "QRIS_SCAN"}); err == nil {
		t.Fatal("QRIS_SCAN order bypassed disabled BTN CPM validation")
	}
}

func setupCPMTestDB(t *testing.T) (*gorm.DB, *repository.QrisTransactionRepository, *OrderService) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&model.QrisTransaction{}, &model.QrisRefund{}, &model.Product{}); err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&model.Product{ID: 1, Name: "Product", Price: 12500, StockQuantity: 10}).Error; err != nil {
		t.Fatal(err)
	}
	repo := repository.NewQrisTransactionRepository(db)
	productService := NewProductService(repository.NewProductRepository(db), nil)
	orderService := NewOrderService(nil, productService, nil, nil)
	orderService.ConfigureBTNQRISValidation(repo, true, 0)
	return db, repo, orderService
}

func cpmTestConfig(baseURL, privateKeyPath string) *config.Config {
	cfg := btnTestConfig(baseURL, privateKeyPath)
	cfg.BtnSnapEnabled = true
	cfg.BtnSnapCPMEnabled = true
	cfg.BtnSnapMerchantID = "MERCHANT01"
	cfg.BtnSnapTerminalID = "TERM001"
	cfg.BtnSnapMerchantName = "Test Merchant"
	cfg.BtnSnapMerchantLocation = "Jakarta Pusat"
	cfg.BtnSnapMerchantPAN = "1234567890123456789"
	cfg.QrisDirectExpiresMinutes = 15
	return cfg
}

func cpmTestSnapshot(id string) dto.QrisCheckoutSnapshotRequest {
	return dto.QrisCheckoutSnapshotRequest{
		ID:       id,
		Items:    []dto.QrisCheckoutItem{{ProductID: 1, ProductName: "Product", ProductPrice: 12500, Qty: 1, Subtotal: 12500}},
		Subtotal: 12500, Total: 12500,
	}
}
