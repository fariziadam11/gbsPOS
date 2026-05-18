package service

import (
	"testing"
	"time"

	"gbs-pos-api/internal/database"
	"gbs-pos-api/internal/model"
	"gbs-pos-api/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func setupOrderTest(t *testing.T) (*OrderService, *gorm.DB) {
	db, err := database.NewTestDB()
	require.NoError(t, err)

	repo := repository.NewOrderRepository(db)
	return NewOrderService(repo), db
}

func createTestOrder(db *gorm.DB, id string, paymentMethod string, isVoided, isSettled bool) {
	db.Create(&model.Order{
		ID:            id,
		Subtotal:      10000,
		Tax:           1000,
		Total:         11000,
		PaymentMethod: paymentMethod,
		Timestamp:     time.Now().UnixMilli(),
		IsVoided:      isVoided,
		IsSettled:     isSettled,
		StoreType:     "RETAIL",
		TerminalID:    "POS-001",
		Items: []model.OrderItem{
			{ProductID: 1, ProductName: "Chitato", ProductPrice: 10000, Qty: 1, Subtotal: 10000},
		},
	})
}

func TestOrderService_Create(t *testing.T) {
	svc, _ := setupOrderTest(t)

	order := &model.Order{
		ID:            "ORDER-001",
		Subtotal:      20000,
		Tax:           2000,
		Total:         22000,
		PaymentMethod: "CASH",
		Timestamp:     time.Now().UnixMilli(),
		StoreType:     "RETAIL",
		TerminalID:    "POS-001",
		Items: []model.OrderItem{
			{ProductID: 1, ProductName: "Chitato", ProductPrice: 10000, Qty: 2, Subtotal: 20000},
		},
	}

	created, idempotent, err := svc.Create(order)
	require.NoError(t, err)
	assert.False(t, idempotent)
	assert.Equal(t, "ORDER-001", created.ID)
	assert.Len(t, created.Items, 1)
}

func TestOrderService_Create_Idempotent(t *testing.T) {
	svc, db := setupOrderTest(t)
	createTestOrder(db, "ORDER-001", "CASH", false, false)

	order := &model.Order{
		ID:            "ORDER-001",
		Subtotal:      20000,
		Tax:           2000,
		Total:         22000,
		PaymentMethod: "CASH",
		Timestamp:     time.Now().UnixMilli(),
		StoreType:     "RETAIL",
		TerminalID:    "POS-001",
		Items: []model.OrderItem{
			{ProductID: 1, ProductName: "Chitato", ProductPrice: 10000, Qty: 2, Subtotal: 20000},
		},
	}

	result, idempotent, err := svc.Create(order)
	require.NoError(t, err)
	assert.True(t, idempotent)
	assert.Equal(t, "ORDER-001", result.ID)
}

func TestOrderService_Void(t *testing.T) {
	svc, db := setupOrderTest(t)
	createTestOrder(db, "ORDER-001", "CASH", false, false)

	order, err := svc.Void("ORDER-001", "Customer cancelled", "admin")
	require.NoError(t, err)
	assert.True(t, order.IsVoided)
	assert.Equal(t, "Customer cancelled", order.VoidReason)
	assert.Equal(t, "admin", order.VoidedBy)
	assert.NotNil(t, order.VoidedAt)
}

func TestOrderService_Void_AlreadyVoided(t *testing.T) {
	svc, db := setupOrderTest(t)
	createTestOrder(db, "ORDER-001", "CASH", true, false)

	_, err := svc.Void("ORDER-001", "reason", "admin")
	assert.Error(t, err)
	assert.Equal(t, "ORDER_ALREADY_VOIDED", err.Error())
}

func TestOrderService_Void_AlreadySettled(t *testing.T) {
	svc, db := setupOrderTest(t)
	createTestOrder(db, "ORDER-001", "CASH", false, true)

	_, err := svc.Void("ORDER-001", "reason", "admin")
	assert.Error(t, err)
	assert.Equal(t, "ORDER_ALREADY_SETTLED", err.Error())
}

func TestOrderService_Void_NotFound(t *testing.T) {
	svc, _ := setupOrderTest(t)

	_, err := svc.Void("NONEXISTENT", "reason", "admin")
	assert.Error(t, err)
	assert.Equal(t, "ORDER_NOT_FOUND", err.Error())
}

func TestOrderService_BulkCreate(t *testing.T) {
	svc, _ := setupOrderTest(t)

	orders := []model.Order{
		{ID: "ORDER-001", Subtotal: 10000, Tax: 1000, Total: 11000, PaymentMethod: "CASH", Timestamp: time.Now().UnixMilli(), StoreType: "RETAIL"},
		{ID: "ORDER-002", Subtotal: 20000, Tax: 2000, Total: 22000, PaymentMethod: "CARD", Timestamp: time.Now().UnixMilli(), StoreType: "RETAIL"},
	}

	result, err := svc.BulkCreate(orders)
	require.NoError(t, err)
	assert.Equal(t, 2, result.Created)
	assert.Equal(t, 0, result.Existing)
	assert.Equal(t, 0, result.Failed)
	assert.Len(t, result.Orders, 2)
}

func TestOrderService_BulkCreate_Idempotent(t *testing.T) {
	svc, db := setupOrderTest(t)
	createTestOrder(db, "ORDER-001", "CASH", false, false)

	orders := []model.Order{
		{ID: "ORDER-001", Subtotal: 10000, Tax: 1000, Total: 11000, PaymentMethod: "CASH", Timestamp: time.Now().UnixMilli(), StoreType: "RETAIL"},
		{ID: "ORDER-002", Subtotal: 20000, Tax: 2000, Total: 22000, PaymentMethod: "CARD", Timestamp: time.Now().UnixMilli(), StoreType: "RETAIL"},
	}

	result, err := svc.BulkCreate(orders)
	require.NoError(t, err)
	assert.Equal(t, 1, result.Created)
	assert.Equal(t, 1, result.Existing)
	assert.True(t, result.Idempotent)
}
