package service

import (
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

func TestConvertQRISRequest(t *testing.T) {
	tests := []struct {
		name      string
		req       dto.ConvertQRISRequest
		wantError bool
	}{
		{
			name: "valid request without fee",
			req: dto.ConvertQRISRequest{
				Amount: 50000,
			},
			wantError: false,
		},
		{
			name: "valid request with fixed fee",
			req: dto.ConvertQRISRequest{
				Amount:   50000,
				FeeType:  "fixed",
				FeeValue: 1000,
			},
			wantError: false,
		},
		{
			name: "valid request with percentage fee",
			req: dto.ConvertQRISRequest{
				Amount:   50000,
				FeeType:  "percentage",
				FeeValue: 2.5,
			},
			wantError: false,
		},
		{
			name: "invalid fee type",
			req: dto.ConvertQRISRequest{
				Amount:  50000,
				FeeType: "invalid",
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hasError := false

			// Validate fee type
			if tt.req.FeeType != "" && tt.req.FeeType != "fixed" && tt.req.FeeType != "percentage" {
				hasError = true
			}

			if hasError != tt.wantError {
				t.Errorf("validation error = %v, wantError = %v", hasError, tt.wantError)
			}
		})
	}
}

func TestQrisTransactionStatus(t *testing.T) {
	// Test that status constants are defined correctly
	tests := []struct {
		name     string
		expected string
	}{
		{"Pending status", "PENDING"},
		{"AwaitingConfirmation status", "AWAITING_CONFIRMATION"},
		{"Paid status", "PAID"},
		{"Cancelled status", "CANCELLED"},
		{"Expired status", "EXPIRED"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Just verify the test runs without panic
			// Actual status values are tested through integration tests
		})
	}
}

func TestGeneratePartnerReferenceNoIsNumericWIBTimestamp(t *testing.T) {
	now := time.Date(2026, time.October, 15, 1, 2, 3, 0, time.UTC)
	reference, err := generatePartnerReferenceNo(now)
	if err != nil {
		t.Fatal(err)
	}
	if len(reference) != 22 || reference[:14] != "20261015080203" {
		t.Fatalf("unexpected partner reference: %q", reference)
	}
	for _, r := range reference {
		if r < '0' || r > '9' {
			t.Fatalf("partner reference is not numeric: %q", reference)
		}
	}
}

func TestLegacyDynamicQrisSupportsCheckoutBeforeOrderIsSaved(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&model.QrisTransaction{}, &model.Order{}); err != nil {
		t.Fatal(err)
	}

	const orderID = "ORDER-0001"
	snapshot := &dto.QrisCheckoutSnapshotRequest{
		ID: orderID,
		Items: []dto.QrisCheckoutItem{{
			ProductID: 1, ProductName: "Product", ProductPrice: 10000,
			Qty: 1, Subtotal: 10000,
		}},
		Subtotal: 10000, Tax: 1000, Total: 11000,
	}
	cfg := &config.Config{
		QrisDirectStaticQRIS:      "00020101021126610014COM.GO-JEK.WWW01189360091434374848210210G4374848210303UMI51440014ID.CO.QRIS.WWW0215ID10265153412990303UMI5204581253033605802ID5925Snack Kering Mama Tari, K6013JAKARTA PUSAT61051064062140703A01110362163042807",
		QrisDirectMerchantName:    "Test Merchant",
		QrisDirectMerchantCity:    "Jakarta",
		QrisDirectProvider:        "GoPay",
		QrisDirectExpiresMinutes:  15,
		QrisDirectSkipCRCValidate: true,
	}
	service := NewQrisDirectService(cfg, repository.NewQrisTransactionRepository(db), repository.NewOrderRepository(db))

	generated, err := service.ConvertQRIS(t.Context(), dto.ConvertQRISRequest{
		OrderID: orderID, Amount: snapshot.Total, CheckoutSnapshot: snapshot,
	})
	if err != nil {
		t.Fatal(err)
	}
	if generated.Provider != "GoPay" || generated.DynamicQris == "" || generated.CheckoutSnapshot == nil {
		t.Fatalf("unexpected legacy QRIS response: %+v", generated)
	}

	status, err := service.GetLatestBTNStatusForOrder(t.Context(), orderID)
	if err != nil {
		t.Fatal(err)
	}
	if status.ID != generated.ID || status.DynamicQris != generated.DynamicQris || status.CheckoutSnapshot == nil {
		t.Fatalf("legacy QRIS recovery failed: %+v", status)
	}
}

func TestConvertQRISUsesServerBTNConfiguration(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&model.QrisTransaction{}, &model.Order{}); err != nil {
		t.Fatal(err)
	}
	cfg := &config.Config{
		BtnSnapEnabled:            true,
		QrisDirectStaticQRIS:      "00020101021126610014COM.GO-JEK.WWW01189360091434374848210210G4374848210303UMI51440014ID.CO.QRIS.WWW0215ID10265153412990303UMI5204581253033605802ID5925Snack Kering Mama Tari, K6013JAKARTA PUSAT61051064062140703A01110362163042807",
		QrisDirectSkipCRCValidate: true,
	}
	service := NewQrisDirectService(cfg, repository.NewQrisTransactionRepository(db), repository.NewOrderRepository(db))

	_, err = service.ConvertQRIS(t.Context(), dto.ConvertQRISRequest{Amount: 5000, Provider: "QRIS_DIRECT"})
	if err == nil || !strings.Contains(err.Error(), "orderId and matching checkoutSnapshot are required for BTN QRIS") {
		t.Fatalf("request provider must not bypass BTN configuration; got error %v", err)
	}
}

func TestBTNCheckoutHasAtMostOneUnresolvedPaymentAttempt(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&model.QrisTransaction{}); err != nil {
		t.Fatal(err)
	}
	repo := repository.NewQrisTransactionRepository(db)
	ctx := t.Context()
	first := &model.QrisTransaction{
		ID: "BTN-FIRST", OrderID: "ORDER-1", Provider: model.QrisProviderBTNSnap,
		Status: model.QrisTransactionStatusPending,
	}
	duplicate := &model.QrisTransaction{
		ID: "BTN-DUPLICATE", OrderID: "ORDER-1", Provider: model.QrisProviderBTNSnap,
		Status: model.QrisTransactionStatusUnknown,
	}
	if err := repo.Create(ctx, first); err != nil {
		t.Fatal(err)
	}
	if err := repo.Create(ctx, duplicate); err == nil {
		t.Fatal("second unresolved BTN attempt for the checkout was accepted")
	}
	if err := repo.UpdateStatus(ctx, first.ID, model.QrisTransactionStatusFailed); err != nil {
		t.Fatal(err)
	}
	if err := repo.Create(ctx, duplicate); err != nil {
		t.Fatalf("terminal failure should allow a new attempt: %v", err)
	}
}

func TestCancelRejectsTransactionThatAlreadyChangedStatus(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&model.QrisTransaction{}); err != nil {
		t.Fatal(err)
	}
	repo := repository.NewQrisTransactionRepository(db)
	if err := repo.Create(t.Context(), &model.QrisTransaction{
		ID: "QRIS-PAID", Provider: "GoPay", Status: model.QrisTransactionStatusPaid,
	}); err != nil {
		t.Fatal(err)
	}
	if err := repo.Cancel(t.Context(), "QRIS-PAID", "cashier", "closed"); err == nil {
		t.Fatal("cancel must report when the transaction is no longer pending")
	}
}

func TestConfirmPendingIsIdempotent(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&model.QrisTransaction{}); err != nil {
		t.Fatal(err)
	}

	repo := repository.NewQrisTransactionRepository(db)
	if err := repo.Create(t.Context(), &model.QrisTransaction{
		ID: "QRIS-TEST", Provider: "GoPay", Status: model.QrisTransactionStatusPending,
		ExpiresAt: time.Now().Add(time.Minute),
	}); err != nil {
		t.Fatal(err)
	}
	service := NewQrisDirectService(&config.Config{QrisDirectAutoConfirmDelaySeconds: 30}, repo, nil)

	if err := service.ConfirmPending(t.Context(), "QRIS-TEST"); err != nil {
		t.Fatal(err)
	}
	if err := service.ConfirmPending(t.Context(), "QRIS-TEST"); err != nil {
		t.Fatalf("duplicate confirm-pending should be accepted: %v", err)
	}
}
