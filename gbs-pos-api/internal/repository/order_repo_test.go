package repository

import (
	"testing"
	"time"

	"gbs-pos-api/internal/model"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

type seedOrderRequest struct {
	ID            string
	PaymentMethod string
	IsVoided      bool
	IsSettled     bool
	StoreType     string
	TerminalID    string
	Total         float64
}

func setupOrderRepoTestDB(t *testing.T) *gorm.DB {

	db, err := gorm.Open(
		sqlite.Open(":memory:"),
		&gorm.Config{},
	)

	require.NoError(t, err)

	err = db.AutoMigrate(
		&model.Order{},
		&model.OrderItem{},
	)

	require.NoError(t, err)

	return db
}

func seedOrder(
	t *testing.T,
	db *gorm.DB,
	req seedOrderRequest,
) {

	order := model.Order{
		ID:            req.ID,
		Subtotal:      req.Total,
		Tax:           0,
		Total:         req.Total,
		PaymentMethod: req.PaymentMethod,
		Timestamp:     time.Now().UnixMilli(),
		IsVoided:      req.IsVoided,
		IsSettled:     req.IsSettled,
		StoreType:     req.StoreType,
		TerminalID:    req.TerminalID,
		Items: []model.OrderItem{
			{
				ProductID:    1,
				ProductName:  "Chitato",
				ProductPrice: req.Total,
				Qty:           1,
				Subtotal:      req.Total,
			},
		},
	}

	err := db.Create(&order).Error

	require.NoError(t, err)
}

func TestOrderRepository_FindByIDWithItems(t *testing.T) {

	db := setupOrderRepoTestDB(t)

	repo := NewOrderRepository(db)

	seedOrder(t, db, seedOrderRequest{
		ID:            "ORDER-001",
		PaymentMethod: "CASH",
		IsVoided:      false,
		IsSettled:     false,
		StoreType:     "RETAIL",
		TerminalID:    "POS-001",
		Total:         10000,
	})

	order, err := repo.FindByIDWithItems("ORDER-001")

	require.NoError(t, err)

	assert.Equal(t, "ORDER-001", order.ID)

	assert.Len(t, order.Items, 1)

	assert.Equal(
		t,
		"Chitato",
		order.Items[0].ProductName,
	)
}

func TestOrderRepository_FindAll_FilterStoreType(t *testing.T) {

	db := setupOrderRepoTestDB(t)

	repo := NewOrderRepository(db)

	seedOrder(t, db, seedOrderRequest{
		ID:            "ORDER-001",
		PaymentMethod: "CASH",
		IsVoided:      false,
		IsSettled:     false,
		StoreType:     "RETAIL",
		TerminalID:    "POS-001",
		Total:         10000,
	})

	seedOrder(t, db, seedOrderRequest{
		ID:            "ORDER-002",
		PaymentMethod: "CARD",
		IsVoided:      false,
		IsSettled:     false,
		StoreType:     "FNB",
		TerminalID:    "POS-002",
		Total:         20000,
	})

	orders, err := repo.FindAll(
		"RETAIL",
		0,
		0,
		nil,
		nil,
		"",
		"",
	)

	require.NoError(t, err)

	assert.Len(t, orders, 1)

	assert.Equal(
		t,
		"ORDER-001",
		orders[0].ID,
	)
}

func TestOrderRepository_FindUnsettledSummary(t *testing.T) {

	db := setupOrderRepoTestDB(t)

	repo := NewOrderRepository(db)

	seedOrder(t, db, seedOrderRequest{
		ID:            "ORDER-001",
		PaymentMethod: "CASH",
		IsVoided:      false,
		IsSettled:     false,
		StoreType:     "RETAIL",
		TerminalID:    "POS-001",
		Total:         10000,
	})

	seedOrder(t, db, seedOrderRequest{
		ID:            "ORDER-002",
		PaymentMethod: "CARD",
		IsVoided:      false,
		IsSettled:     false,
		StoreType:     "RETAIL",
		TerminalID:    "POS-001",
		Total:         20000,
	})

	seedOrder(t, db, seedOrderRequest{
		ID:            "ORDER-003",
		PaymentMethod: "CASH",
		IsVoided:      true,
		IsSettled:     false,
		StoreType:     "RETAIL",
		TerminalID:    "POS-001",
		Total:         30000,
	})

	count, total, summary, err := repo.FindUnsettledSummary(
		"RETAIL",
		"POS-001",
	)

	require.NoError(t, err)

	assert.Equal(t, 2, count)

	assert.Equal(t, 30000.0, total)

	assert.Equal(t, 1, summary["CASH"].Count)

	assert.Equal(t, 10000.0, summary["CASH"].Total)

	assert.Equal(t, 1, summary["CARD"].Count)

	assert.Equal(t, 20000.0, summary["CARD"].Total)

	assert.Equal(t, 0, summary["QRIS"].Count)
}

func TestOrderRepository_FindUnsettledOrders(t *testing.T) {

	db := setupOrderRepoTestDB(t)

	repo := NewOrderRepository(db)

	seedOrder(t, db, seedOrderRequest{
		ID:            "ORDER-001",
		PaymentMethod: "CASH",
		IsVoided:      false,
		IsSettled:     false,
		StoreType:     "RETAIL",
		TerminalID:    "POS-001",
		Total:         10000,
	})

	seedOrder(t, db, seedOrderRequest{
		ID:            "ORDER-002",
		PaymentMethod: "CASH",
		IsVoided:      true,
		IsSettled:     false,
		StoreType:     "RETAIL",
		TerminalID:    "POS-001",
		Total:         10000,
	})

	seedOrder(t, db, seedOrderRequest{
		ID:            "ORDER-003",
		PaymentMethod: "CASH",
		IsVoided:      false,
		IsSettled:     true,
		StoreType:     "RETAIL",
		TerminalID:    "POS-001",
		Total:         10000,
	})

	orders, err := repo.FindUnsettledOrders(
		"RETAIL",
		"POS-001",
		false,
	)

	require.NoError(t, err)

	assert.Len(t, orders, 1)

	assert.Equal(
		t,
		"ORDER-001",
		orders[0].ID,
	)
}

func TestOrderRepository_MarkSettled(t *testing.T) {

	db := setupOrderRepoTestDB(t)

	repo := NewOrderRepository(db)

	seedOrder(t, db, seedOrderRequest{
		ID:            "ORDER-001",
		PaymentMethod: "CASH",
		IsVoided:      false,
		IsSettled:     false,
		StoreType:     "RETAIL",
		TerminalID:    "POS-001",
		Total:         10000,
	})

	err := repo.MarkSettled([]string{"ORDER-001"})

	require.NoError(t, err)

	order, err := repo.FindByID("ORDER-001")

	require.NoError(t, err)

	assert.True(t, order.IsSettled)
}