package service

import (
	"context"
	"testing"
	"time"

	"gbs-pos-api/internal/model"
	"gbs-pos-api/internal/repository"
	ws "gbs-pos-api/internal/websocket"

	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestCardPaymentService_CompanionSuccessFinalizesOrder(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.Order{}, &model.OrderItem{}, &model.CardPayment{}))

	orderID := "WEB-ORDER-001"
	require.NoError(t, db.Create(&model.Order{ID: orderID, Total: 25000, PaymentMethod: "CARD", Timestamp: time.Now().UnixMilli()}).Error)
	paymentID := uuid.New()
	require.NoError(t, db.Create(&model.CardPayment{
		ID: paymentID, OrderID: orderID, Amount: 25000, Status: model.CardPaymentWaiting,
		DeviceID: "HP-001", TerminalID: "POS-001", ExpiresAt: time.Now().Add(time.Minute),
	}).Error)

	hub := ws.NewHub()
	service := NewCardPaymentService(
		repository.NewCardPaymentRepository(db),
		NewOrderService(repository.NewOrderRepository(db), nil, nil, nil),
		repository.NewCompanionDeviceRepository(db),
		hub,
	)
	err = service.UpdateFromCompanion(context.Background(), &ws.Client{Type: ws.ClientCompanion, ID: "HP-001"}, ws.Message{
		PaymentID: paymentID.String(), Status: model.CardPaymentSuccess,
		TransactionID: "TX-001", AuthCode: "AUTH-001", EntryMode: "CONTACTLESS",
		MaskedCard: "****1234", AcqMID: "MID-001", AcqTID: "TID-001", PosMessageID: "POS-MSG-001",
	})
	require.NoError(t, err)

	var payment model.CardPayment
	require.NoError(t, db.First(&payment, "id = ?", paymentID).Error)
	require.Equal(t, model.CardPaymentSuccess, payment.Status)

	var order model.Order
	require.NoError(t, db.First(&order, "id = ?", orderID).Error)
	require.Equal(t, "TX-001", order.TransactionID)
	require.Equal(t, "AUTH-001", order.ApprovalCode)
	require.Equal(t, "CONTACTLESS", order.EntryMode)
	require.Equal(t, "****1234", order.MaskedAccount)
	require.Equal(t, "MID-001", order.AcqMid)
	require.Equal(t, "TID-001", order.AcqTid)
	require.Equal(t, "POS-MSG-001", order.PosMessageID)
}

func setupCardPaymentTest(t *testing.T) (*gorm.DB, *CardPaymentService, *ws.Hub) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.Order{}, &model.OrderItem{}, &model.CardPayment{}))

	hub := ws.NewHub()
	svc := NewCardPaymentService(
		repository.NewCardPaymentRepository(db),
		NewOrderService(repository.NewOrderRepository(db), nil, nil, nil),
		repository.NewCompanionDeviceRepository(db),
		hub,
	)
	return db, svc, hub
}

func TestCardPaymentService_ExpireProcessing(t *testing.T) {
	db, svc, _ := setupCardPaymentTest(t)

	orderID := "WEB-ORDER-PROC"
	require.NoError(t, db.Create(&model.Order{ID: orderID, Total: 50000, PaymentMethod: "CARD", Timestamp: time.Now().UnixMilli()}).Error)
	paymentID := uuid.New()
	require.NoError(t, db.Create(&model.CardPayment{
		ID: paymentID, OrderID: orderID, Amount: 50000, Status: model.CardPaymentProcessing,
		DeviceID: "HP-001", TerminalID: "POS-001", ExpiresAt: time.Now().Add(-time.Minute),
	}).Error)

	require.NoError(t, svc.Expire(context.Background()))

	var payment model.CardPayment
	require.NoError(t, db.First(&payment, "id = ?", paymentID).Error)
	require.Equal(t, model.CardPaymentExpired, payment.Status)

	// PROCESSING expiry must NOT void the order: the card may already have been
	// charged, so the order is left open for reconciliation.
	var order model.Order
	require.NoError(t, db.First(&order, "id = ?", orderID).Error)
	require.False(t, order.IsVoided)
}

func TestCardPaymentService_ExpireWaitingVoidsOrder(t *testing.T) {
	db, svc, _ := setupCardPaymentTest(t)

	orderID := "WEB-ORDER-WAIT-EXPIRE"
	require.NoError(t, db.Create(&model.Order{ID: orderID, Total: 50000, PaymentMethod: "CARD", Timestamp: time.Now().UnixMilli()}).Error)
	paymentID := uuid.New()
	require.NoError(t, db.Create(&model.CardPayment{
		ID: paymentID, OrderID: orderID, Amount: 50000, Status: model.CardPaymentWaiting,
		DeviceID: "HP-001", TerminalID: "POS-001", ExpiresAt: time.Now().Add(-time.Minute),
	}).Error)

	require.NoError(t, svc.Expire(context.Background()))

	var order model.Order
	require.NoError(t, db.First(&order, "id = ?", orderID).Error)
	require.True(t, order.IsVoided)
}

func TestCardPaymentService_LateSuccessAfterExpiredIsRejected(t *testing.T) {
	db, svc, _ := setupCardPaymentTest(t)

	orderID := "WEB-ORDER-LATE"
	require.NoError(t, db.Create(&model.Order{ID: orderID, Total: 50000, PaymentMethod: "CARD", Timestamp: time.Now().UnixMilli()}).Error)
	paymentID := uuid.New()
	require.NoError(t, db.Create(&model.CardPayment{
		ID: paymentID, OrderID: orderID, Amount: 50000, Status: model.CardPaymentExpired,
		DeviceID: "HP-001", TerminalID: "POS-001", ExpiresAt: time.Now().Add(-time.Minute),
	}).Error)

	// A late SUCCESS after EXPIRED must not finalize or change state — but the
	// service should NOT error out silently; it logs & notifies. No error is returned.
	err := svc.UpdateFromCompanion(context.Background(), &ws.Client{Type: ws.ClientCompanion, ID: "HP-001"}, ws.Message{
		PaymentID: paymentID.String(), Status: model.CardPaymentSuccess,
		TransactionID: "TX-LATE-001",
	})
	require.NoError(t, err)

	var payment model.CardPayment
	require.NoError(t, db.First(&payment, "id = ?", paymentID).Error)
	require.Equal(t, model.CardPaymentExpired, payment.Status)
}

func TestCardPaymentService_DuplicateProcessingIsIdempotent(t *testing.T) {
	db, svc, _ := setupCardPaymentTest(t)

	orderID := "WEB-ORDER-DUP"
	require.NoError(t, db.Create(&model.Order{ID: orderID, Total: 50000, PaymentMethod: "CARD", Timestamp: time.Now().UnixMilli()}).Error)
	paymentID := uuid.New()
	oldExpiry := time.Now().Add(3 * time.Minute)
	require.NoError(t, db.Create(&model.CardPayment{
		ID: paymentID, OrderID: orderID, Amount: 50000, Status: model.CardPaymentProcessing,
		DeviceID: "HP-001", TerminalID: "POS-001", ExpiresAt: oldExpiry,
	}).Error)

	// Duplicate PROCESSING on an already-PROCESSING payment is idempotent: no error,
	// and the expiry window is NOT extended by a stale client.
	err := svc.UpdateFromCompanion(context.Background(), &ws.Client{Type: ws.ClientCompanion, ID: "HP-001"}, ws.Message{
		PaymentID: paymentID.String(), Status: model.CardPaymentProcessing,
	})
	require.NoError(t, err)

	var payment model.CardPayment
	require.NoError(t, db.First(&payment, "id = ?", paymentID).Error)
	require.Equal(t, model.CardPaymentProcessing, payment.Status)
	require.Equal(t, oldExpiry.Unix(), payment.ExpiresAt.Unix())
}

func TestCardPaymentService_ProcessingRefreshesExpiry(t *testing.T) {
	db, svc, _ := setupCardPaymentTest(t)

	orderID := "WEB-ORDER-REFRESH"
	require.NoError(t, db.Create(&model.Order{ID: orderID, Total: 50000, PaymentMethod: "CARD", Timestamp: time.Now().UnixMilli()}).Error)
	paymentID := uuid.New()
	oldExpiry := time.Now().Add(10 * time.Second)
	require.NoError(t, db.Create(&model.CardPayment{
		ID: paymentID, OrderID: orderID, Amount: 50000, Status: model.CardPaymentWaiting,
		DeviceID: "HP-001", TerminalID: "POS-001", ExpiresAt: oldExpiry,
	}).Error)

	err := svc.UpdateFromCompanion(context.Background(), &ws.Client{Type: ws.ClientCompanion, ID: "HP-001"}, ws.Message{
		PaymentID: paymentID.String(), Status: model.CardPaymentProcessing,
	})
	require.NoError(t, err)

	var payment model.CardPayment
	require.NoError(t, db.First(&payment, "id = ?", paymentID).Error)
	require.Equal(t, model.CardPaymentProcessing, payment.Status)
	require.True(t, payment.ExpiresAt.After(oldExpiry), "expiry should be refreshed past the old deadline")
}

func TestCardPaymentService_CancelRejectedWhenAlreadySuccess(t *testing.T) {
	db, svc, _ := setupCardPaymentTest(t)

	orderID := "WEB-ORDER-CANCELRACE"
	require.NoError(t, db.Create(&model.Order{ID: orderID, Total: 50000, PaymentMethod: "CARD", Timestamp: time.Now().UnixMilli()}).Error)
	paymentID := uuid.New()
	require.NoError(t, db.Create(&model.CardPayment{
		ID: paymentID, OrderID: orderID, Amount: 50000, Status: model.CardPaymentSuccess,
		DeviceID: "HP-001", TerminalID: "POS-001", ExpiresAt: time.Now().Add(time.Minute),
	}).Error)

	_, err := svc.Cancel(context.Background(), paymentID)
	require.Error(t, err)

	// Order must NOT be voided when cancel loses the race to SUCCESS.
	var order model.Order
	require.NoError(t, db.First(&order, "id = ?", orderID).Error)
	require.False(t, order.IsVoided)
}

func TestCardPaymentService_CancelWaitingSucceeds(t *testing.T) {
	db, svc, _ := setupCardPaymentTest(t)

	orderID := "WEB-ORDER-CANCELOK"
	require.NoError(t, db.Create(&model.Order{ID: orderID, Total: 50000, PaymentMethod: "CARD", Timestamp: time.Now().UnixMilli()}).Error)
	paymentID := uuid.New()
	require.NoError(t, db.Create(&model.CardPayment{
		ID: paymentID, OrderID: orderID, Amount: 50000, Status: model.CardPaymentWaiting,
		DeviceID: "HP-001", TerminalID: "POS-001", ExpiresAt: time.Now().Add(time.Minute),
	}).Error)

	payment, err := svc.Cancel(context.Background(), paymentID)
	require.NoError(t, err)
	require.Equal(t, model.CardPaymentCancelled, payment.Status)

	var order model.Order
	require.NoError(t, db.First(&order, "id = ?", orderID).Error)
	require.True(t, order.IsVoided)
}
