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
