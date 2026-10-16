package service

import (
	"crypto"
	"crypto/hmac"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/sha512"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"gbs-pos-api/internal/config"
	"gbs-pos-api/internal/dto"
	"gbs-pos-api/internal/model"
	"gbs-pos-api/internal/repository"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

const (
	btnTestClientKey    = "oauth-test-client"
	btnTestPartnerID    = "partner-test-id"
	btnTestClientSecret = "client-secret-test"
	btnTestMerchantID   = "merchant-test-001"
	btnTestTerminalID   = "terminal-test-01"
)

func TestBtnSnapClientSignsAndSendsGenerateAndQuery(t *testing.T) {
	privateKey, keyPath := writeTestRSAPrivateKey(t)
	var externalIDs []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("read request body: %v", err)
			http.Error(w, "bad body", http.StatusBadRequest)
			return
		}
		switch r.URL.Path {
		case btnTokenPath:
			if r.Header.Get("X-CLIENT-KEY") != btnTestClientKey || r.Header.Get("Origin") != "pos.example.test" {
				t.Errorf("unexpected token headers: %#v", r.Header)
			}
			sig, err := base64.StdEncoding.DecodeString(r.Header.Get("X-SIGNATURE"))
			if err != nil {
				t.Errorf("decode token signature: %v", err)
				http.Error(w, "bad signature", http.StatusBadRequest)
				return
			}
			hash := sha256.Sum256([]byte(btnTestClientKey + "|" + r.Header.Get("X-TIMESTAMP")))
			if err := rsa.VerifyPKCS1v15(&privateKey.PublicKey, crypto.SHA256, hash[:], sig); err != nil {
				t.Errorf("invalid token RSA signature: %v", err)
			}
			_, _ = w.Write([]byte(`{"responseCode":"2007300","responseMessage":"Successful","accessToken":"test-token","tokenType":"Bearer","expiresIn":"900"}`))
		case btnGeneratePath, btnQueryPath:
			if r.Header.Get("Authorization") != "Bearer test-token" ||
				r.Header.Get("X-PARTNER-ID") != btnTestPartnerID ||
				r.Header.Get("CHANNEL-ID") != "00123" ||
				r.Header.Get("Origin") != "pos.example.test" {
				t.Errorf("unexpected API headers: %#v", r.Header)
			}
			requestID := r.Header.Get("X-EXTERNAL-ID")
			if len(requestID) != 16 {
				t.Errorf("external ID length = %d, want 16", len(requestID))
			}
			if _, err := hex.DecodeString(requestID); err != nil {
				t.Errorf("external ID is not alphanumeric hex: %v", err)
			}
			externalIDs = append(externalIDs, requestID)
			hash := sha256.Sum256(body)
			stringToSign := fmt.Sprintf("POST:%s:test-token:%s:%s", r.URL.Path, hex.EncodeToString(hash[:]), r.Header.Get("X-TIMESTAMP"))
			mac := hmac.New(sha512.New, []byte(btnTestClientSecret))
			_, _ = mac.Write([]byte(stringToSign))
			wantSignature := base64.StdEncoding.EncodeToString(mac.Sum(nil))
			if r.Header.Get("X-SIGNATURE") != wantSignature {
				t.Errorf("HMAC signature mismatch for %s", r.URL.Path)
			}
			if r.URL.Path == btnGeneratePath {
				var request dto.BtnSnapGenerateRequest
				if err := json.Unmarshal(body, &request); err != nil {
					t.Errorf("decode Generate request: %v", err)
					http.Error(w, "bad body", http.StatusBadRequest)
					return
				}
				if request.ValidityPeriod == "" || request.Amount.Value != "12500.00" || request.Amount.Currency != "IDR" || request.AdditionalInfo["type_qris"] != "D" {
					t.Errorf("unexpected Generate payload: %+v", request)
				}
				_, _ = w.Write([]byte(`{"responseCode":"2004700","responseMessage":"Success","referenceNo":"bank-reference-1","partnerReferenceNo":"partner-reference-1","qrContent":"000201010212","merchantName":"Test Merchant","terminalId":"terminal-test-01"}`))
				return
			}
			_, _ = w.Write([]byte(`{"responseCode":"2005100","responseMessage":"Success","originalReferenceNo":"bank-reference-1","originalPartnerReferenceNo":"partner-reference-1","serviceCode":"47","latestTransactionStatus":"03","transactionStatusDesc":"Pending","amount":{"value":"12500.00","currency":"IDR"},"additionalInfo":{"merchantId":"merchant-test-001"}}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	cfg := btnTestConfig(server.URL, keyPath)
	client := NewBtnSnapClient(cfg)
	generated, err := client.GenerateQR(t.Context(), dto.BtnSnapGenerateRequest{
		PartnerReferenceNo: "partner-reference-1",
		Amount:             dto.BtnSnapAmount{Value: "12500.00", Currency: "IDR"},
		MerchantID:         btnTestMerchantID,
		TerminalID:         btnTestTerminalID,
		ValidityPeriod:     time.Now().Add(15 * time.Minute).Format(time.RFC3339),
		AdditionalInfo:     map[string]any{"type_qris": "D"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if generated.QRContent != "000201010212" {
		t.Fatalf("QR content = %q", generated.QRContent)
	}
	query, err := client.QueryPayment(t.Context(), dto.BtnSnapQueryRequest{
		OriginalPartnerReferenceNo: "partner-reference-1",
		ServiceCode:                "47",
		MerchantID:                 btnTestMerchantID,
		AdditionalInfo:             map[string]any{},
	})
	if err != nil {
		t.Fatal(err)
	}
	if query.LatestTransactionStatus != "03" {
		t.Fatalf("bank status = %q, want 03", query.LatestTransactionStatus)
	}
	if len(externalIDs) != 2 || externalIDs[0] == externalIDs[1] {
		t.Fatalf("X-EXTERNAL-ID values were not unique: %v", externalIDs)
	}
}

func TestBTNGenerateRequiresValidityPeriod(t *testing.T) {
	client := NewBtnSnapClient(&config.Config{})
	_, err := client.generateQRWithToken(t.Context(), dto.BtnSnapGenerateRequest{}, "test-token")
	if err == nil || !strings.Contains(err.Error(), "validityPeriod is required") {
		t.Fatalf("Generate QR error = %v, want required validityPeriod error", err)
	}
}

func TestQrisDirectServiceBTNGenerateQueryAndConfirmGuards(t *testing.T) {
	_, keyPath := writeTestRSAPrivateKey(t)
	var partnerReference string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case btnTokenPath:
			_, _ = w.Write([]byte(`{"responseCode":"2007300","accessToken":"test-token","expiresIn":"900"}`))
		case btnGeneratePath:
			var request dto.BtnSnapGenerateRequest
			if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
				t.Errorf("decode Generate request: %v", err)
				http.Error(w, "bad body", http.StatusBadRequest)
				return
			}
			if _, err := time.Parse(time.RFC3339, request.ValidityPeriod); err != nil {
				t.Errorf("Generate validityPeriod = %q, want RFC3339 timestamp: %v", request.ValidityPeriod, err)
			}
			partnerReference = request.PartnerReferenceNo
			_, _ = fmt.Fprintf(w, `{"responseCode":"2004700","referenceNo":"bank-reference-2","partnerReferenceNo":%q,"qrContent":"btn-dynamic-qr","merchantName":"Test Merchant","terminalId":"%s"}`, partnerReference, btnTestTerminalID)
		case btnQueryPath:
			_, _ = fmt.Fprintf(w, `{"responseCode":"2005100","originalReferenceNo":"bank-reference-2","originalPartnerReferenceNo":%q,"serviceCode":"47","latestTransactionStatus":"00","transactionStatusDesc":"Success","paidTime":"2026-10-05T12:30:45+07:00","terminalId":"%s","amount":{"value":"12500.00","currency":"IDR"},"additionalInfo":{"merchantId":"%s"}}`, partnerReference, btnTestTerminalID, btnTestMerchantID)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&model.QrisTransaction{}, &model.Product{}); err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&model.Product{ID: 1, Name: "Product", Price: 12500, StockQuantity: 10}).Error; err != nil {
		t.Fatal(err)
	}
	qrisRepo := repository.NewQrisTransactionRepository(db)
	cfg := btnTestConfig(server.URL, keyPath)
	cfg.BtnSnapEnabled = true
	cfg.QrisDirectExpiresMinutes = 15
	service := NewQrisDirectService(cfg, qrisRepo, nil)
	productService := NewProductService(repository.NewProductRepository(db), nil)
	orderService := NewOrderService(nil, productService, nil, nil)
	orderService.ConfigureBTNQRISValidation(qrisRepo, true, 0)
	service.SetOrderService(orderService)
	snapshot := &dto.QrisCheckoutSnapshotRequest{
		ID: "ORDER-0001",
		Items: []dto.QrisCheckoutItem{{
			ProductID: 1, ProductName: "Product", ProductPrice: 12500,
			Qty: 1, Subtotal: 12500,
		}},
		Subtotal: 12500, Tax: 0, Total: 12500,
	}
	generated, err := service.ConvertQRIS(t.Context(), dto.ConvertQRISRequest{
		OrderID: "ORDER-0001", Amount: 12500, CheckoutSnapshot: snapshot,
	})
	if err != nil {
		t.Fatal(err)
	}
	if generated.Provider != model.QrisProviderBTNSnap || generated.DynamicQris != "btn-dynamic-qr" {
		t.Fatalf("unexpected BTN Generate response: %+v", generated)
	}
	status, err := service.GetTransactionStatus(t.Context(), generated.ID)
	if err != nil {
		t.Fatal(err)
	}
	if status.Status != model.QrisTransactionStatusPaid || status.PaidAt == nil || !status.PaidAt.Equal(time.Date(2026, 10, 5, 5, 30, 45, 0, time.UTC)) {
		t.Fatalf("unexpected Query result: %+v", status)
	}
	for _, confirm := range []func() error{
		func() error { return service.ConfirmPayment(t.Context(), generated.ID) },
		func() error { return service.ConfirmPending(t.Context(), generated.ID) },
		func() error { return service.CancelPayment(t.Context(), generated.ID, "", "cashier") },
	} {
		if err := confirm(); err == nil {
			t.Fatal("manual BTN status update should be rejected")
		}
	}
	stored, err := qrisRepo.FindByID(t.Context(), generated.ID)
	if err != nil {
		t.Fatal(err)
	}
	if stored.Status != model.QrisTransactionStatusPaid || stored.StaticQrisString != "" {
		t.Fatalf("unexpected saved BTN transaction: %+v", stored)
	}
}

func TestBTNGenerateSeparatesRejectionFromUnknownOutcome(t *testing.T) {
	for _, tt := range []struct {
		name         string
		wantStatus   string
		wantGenerate int
		httpStatus   int
		responseCode string
		responseText string
		definitive   bool
	}{
		{name: "definitive rejection", wantStatus: model.QrisTransactionStatusFailed, wantGenerate: 2, httpStatus: http.StatusUnauthorized, responseCode: "4014701", responseText: "Access Token Invalid", definitive: true},
		{name: "duplicate reference", wantStatus: model.QrisTransactionStatusUnknown, wantGenerate: 1, httpStatus: http.StatusConflict, responseCode: "4094701", responseText: "Duplicate partnerReferenceNo"},
		{name: "server outcome unknown", wantStatus: model.QrisTransactionStatusUnknown, wantGenerate: 1, httpStatus: http.StatusInternalServerError, responseCode: "5004700", responseText: "General Error"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			_, keyPath := writeTestRSAPrivateKey(t)
			generateCalls := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.URL.Path {
				case btnTokenPath:
					_, _ = w.Write([]byte(`{"responseCode":"2007300","accessToken":"test-token","expiresIn":"900"}`))
				case btnGeneratePath:
					generateCalls++
					var request dto.BtnSnapGenerateRequest
					if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
						t.Errorf("decode Generate request: %v", err)
						http.Error(w, "bad body", http.StatusBadRequest)
						return
					}
					if !tt.definitive || generateCalls == 1 {
						w.WriteHeader(tt.httpStatus)
						_, _ = fmt.Fprintf(w, `{"responseCode":%q,"responseMessage":%q}`, tt.responseCode, tt.responseText)
						return
					}
					_, _ = fmt.Fprintf(w, `{"responseCode":"2004700","referenceNo":"bank-reference-retry","partnerReferenceNo":%q,"qrContent":"retry-qr","merchantName":"Test Merchant","terminalId":"%s"}`, request.PartnerReferenceNo, btnTestTerminalID)
				default:
					http.NotFound(w, r)
				}
			}))
			defer server.Close()

			db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
			if err != nil {
				t.Fatal(err)
			}
			if err := db.AutoMigrate(&model.QrisTransaction{}, &model.Product{}); err != nil {
				t.Fatal(err)
			}
			if err := db.Create(&model.Product{ID: 1, Name: "Product", Price: 12500, StockQuantity: 10}).Error; err != nil {
				t.Fatal(err)
			}
			repo := repository.NewQrisTransactionRepository(db)
			cfg := btnTestConfig(server.URL, keyPath)
			cfg.BtnSnapEnabled = true
			service := NewQrisDirectService(cfg, repo, nil)
			productService := NewProductService(repository.NewProductRepository(db), nil)
			orderService := NewOrderService(nil, productService, nil, nil)
			orderService.ConfigureBTNQRISValidation(repo, true, 0)
			service.SetOrderService(orderService)
			snapshot := &dto.QrisCheckoutSnapshotRequest{
				ID: "ORDER-GENERATE-ERROR",
				Items: []dto.QrisCheckoutItem{{
					ProductID: 1, ProductName: "Product", ProductPrice: 12500,
					Qty: 1, Subtotal: 12500,
				}},
				Subtotal: 12500, Tax: 0, Total: 12500,
			}
			request := dto.ConvertQRISRequest{OrderID: snapshot.ID, Amount: 12500, CheckoutSnapshot: snapshot}

			_, err = service.ConvertQRIS(t.Context(), request)
			if err == nil {
				t.Fatal("Generate error was accepted")
			}
			if tt.definitive {
				var rejected *BtnSnapRejectedError
				if !errors.As(err, &rejected) || rejected.ResponseCode != "4014701" {
					t.Fatalf("expected a definitive BTN rejection, got %v", err)
				}
			} else {
				var unknown *BtnSnapOutcomeUnknownError
				if !errors.As(err, &unknown) {
					t.Fatalf("expected an unknown BTN outcome, got %v", err)
				}
			}

			transactions, err := repo.FindByOrderID(t.Context(), snapshot.ID)
			if err != nil {
				t.Fatal(err)
			}
			if len(transactions) != 1 || transactions[0].Status != tt.wantStatus {
				t.Fatalf("saved attempt = %+v, want status %s", transactions, tt.wantStatus)
			}
			if tt.definitive {
				status, err := service.GetLatestBTNStatusForOrder(t.Context(), snapshot.ID)
				if err != nil || status.Status != model.QrisTransactionStatusFailed {
					t.Fatalf("persisted Generate rejection status = %+v, error=%v", status, err)
				}
				generated, err := service.ConvertQRIS(t.Context(), request)
				if err != nil || generated.DynamicQris != "retry-qr" {
					t.Fatalf("definitive rejection should allow a fresh attempt: response=%+v error=%v", generated, err)
				}
			} else if _, err := service.ConvertQRIS(t.Context(), request); err == nil {
				t.Fatal("unknown outcome must not start a fresh Generate request")
			}
			if generateCalls != tt.wantGenerate {
				t.Fatalf("Generate calls = %d, want %d", generateCalls, tt.wantGenerate)
			}
		})
	}
}

func writeTestRSAPrivateKey(t *testing.T) (*rsa.PrivateKey, string) {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "btn-private-key.pem")
	if err := os.WriteFile(path, pem.EncodeToMemory(&pem.Block{
		Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key),
	}), 0o600); err != nil {
		t.Fatal(err)
	}
	return key, path
}

func btnTestConfig(baseURL, privateKeyPath string) *config.Config {
	return &config.Config{
		BtnSnapBaseURL:        baseURL,
		BtnSnapClientKey:      btnTestClientKey,
		BtnSnapPartnerID:      btnTestPartnerID,
		BtnSnapClientSecret:   btnTestClientSecret,
		BtnSnapPrivateKeyPath: privateKeyPath,
		BtnSnapChannelID:      "00123",
		BtnSnapOrigin:         "pos.example.test",
		BtnSnapMerchantID:     btnTestMerchantID,
		BtnSnapTerminalID:     btnTestTerminalID,
	}
}

func TestBTNQueryDoesNotTreatNotFoundAsPaid(t *testing.T) {
	_, keyPath := writeTestRSAPrivateKey(t)
	queryHTTPStatus := http.StatusOK
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == btnTokenPath {
			_, _ = w.Write([]byte(`{"responseCode":"2007300","accessToken":"test-token","expiresIn":"900"}`))
			return
		}
		w.WriteHeader(queryHTTPStatus)
		_, _ = w.Write([]byte(`{"responseCode":"4045101","responseMessage":"Transaction Not Found"}`))
	}))
	defer server.Close()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&model.QrisTransaction{}); err != nil {
		t.Fatal(err)
	}
	repo := repository.NewQrisTransactionRepository(db)
	tx := &model.QrisTransaction{
		ID: "BTN-test", OrderID: "ORDER-0002", PartnerReferenceNo: "partner-reference-2",
		Amount: 5000, TotalAmount: 5000, Provider: model.QrisProviderBTNSnap,
		MerchantID: btnTestMerchantID, Currency: "IDR", TerminalID: btnTestTerminalID,
		Status: model.QrisTransactionStatusPending, ExpiresAt: time.Now().Add(time.Minute),
	}
	if err := repo.Create(t.Context(), tx); err != nil {
		t.Fatal(err)
	}
	cfg := btnTestConfig(server.URL, keyPath)
	cfg.BtnSnapEnabled = true
	service := NewQrisDirectService(cfg, repo, nil)
	for _, httpStatus := range []int{http.StatusOK, http.StatusNotFound} {
		t.Run(fmt.Sprint(httpStatus), func(t *testing.T) {
			queryHTTPStatus = httpStatus
			status, err := service.GetTransactionStatus(t.Context(), tx.ID)
			if err != nil {
				t.Fatal(err)
			}
			if status.Status != model.QrisTransactionStatusPending {
				t.Fatalf("4045101 mapped to %q, want PENDING", status.Status)
			}
		})
	}
}

func TestBTNQueryRejectsUnexpectedHTTPResponses(t *testing.T) {
	for _, tt := range []struct {
		name         string
		httpStatus   int
		responseCode string
	}{
		{name: "other not found", httpStatus: http.StatusNotFound, responseCode: "4045102"},
		{name: "success body with HTTP error", httpStatus: http.StatusNotFound, responseCode: "2005100"},
		{name: "not found body with server error", httpStatus: http.StatusInternalServerError, responseCode: "4045101"},
		{name: "invalid token", httpStatus: http.StatusUnauthorized, responseCode: "4015101"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.httpStatus)
				_, _ = fmt.Fprintf(w, `{"responseCode":%q}`, tt.responseCode)
			}))
			defer server.Close()
			client := NewBtnSnapClient(&config.Config{BtnSnapBaseURL: server.URL})
			client.token = "test-token"
			client.tokenUntil = time.Now().Add(time.Hour)
			if _, err := client.QueryPayment(t.Context(), dto.BtnSnapQueryRequest{}); err == nil {
				t.Fatal("unexpected HTTP error was accepted as a valid query result")
			}
		})
	}
}

func TestOrderServiceRequiresMatchingPaidBTNTransaction(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&model.QrisTransaction{}); err != nil {
		t.Fatal(err)
	}
	repo := repository.NewQrisTransactionRepository(db)
	snapshot := dto.QrisCheckoutSnapshotRequest{
		ID: "ORDER-0003",
		Items: []dto.QrisCheckoutItem{{
			ProductID: 7, ProductName: "Product", ProductPrice: 10000,
			Qty: 1, Subtotal: 10000,
		}},
		Subtotal: 10000, Total: 10000,
	}
	snapshotJSON, err := json.Marshal(snapshot)
	if err != nil {
		t.Fatal(err)
	}
	tx := &model.QrisTransaction{
		ID: "BTN-order-3", OrderID: snapshot.ID, PartnerReferenceNo: "partner-reference-3",
		Amount: 10000, TotalAmount: 10000, Provider: model.QrisProviderBTNSnap,
		MerchantID: btnTestMerchantID, Currency: "IDR", TerminalID: btnTestTerminalID,
		Status: model.QrisTransactionStatusPending, CheckoutSnapshot: snapshotJSON,
		ExpiresAt: time.Now().Add(time.Minute),
	}
	if err := repo.Create(t.Context(), tx); err != nil {
		t.Fatal(err)
	}
	service := &OrderService{qrisTxRepo: repo, btnQrisEnabled: true}
	order := &model.Order{
		ID: snapshot.ID, PaymentMethod: "QRIS", TransactionID: tx.ID,
		Subtotal: 10000, Total: 10000,
		Items: []model.OrderItem{{ProductID: 7, ProductName: "Product", ProductPrice: 10000, Qty: 1, Subtotal: 10000}},
	}
	if err := service.validateBTNQRISPayment(order); err == nil {
		t.Fatal("pending BTN payment must not create an order")
	}
	if err := db.Model(&model.QrisTransaction{}).Where("id = ?", tx.ID).Update("status", model.QrisTransactionStatusPaid).Error; err != nil {
		t.Fatal(err)
	}
	if err := service.validateBTNQRISPayment(order); err != nil {
		t.Fatalf("matching paid BTN payment rejected: %v", err)
	}
	if order.TerminalID != btnTestTerminalID || order.BankName != "BTN" {
		t.Fatalf("order terminal/provider not filled from bank transaction: %+v", order)
	}
	order.Items[0].Subtotal = 9000
	if err := service.validateBTNQRISPayment(order); err == nil {
		t.Fatal("order with changed checkout items must be rejected")
	}
}
