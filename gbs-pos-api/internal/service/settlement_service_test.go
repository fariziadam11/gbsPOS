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

func setupSettlementTest(t *testing.T) (*SettlementService, *repository.OrderRepository, *gorm.DB) {
	db, err := database.NewTestDB()
	require.NoError(t, err)

	orderRepo := repository.NewOrderRepository(db)
	settlementRepo := repository.NewSettlementRepository(db)
	return NewSettlementService(orderRepo, settlementRepo), orderRepo, db
}

func TestSettlementService_GetUnsettledSummary(t *testing.T) {
	svc, _, db := setupSettlementTest(t)

	db.Create(
		&model.Order{
			ID:            "O1",
			Total:         10000,
			PaymentMethod: "CASH",
			Timestamp:     time.Now().UnixMilli(),
			IsVoided:      false,
			IsSettled:     false,
			StoreType:     "RETAIL",
		},
	)
	db.Create(
		&model.Order{
			ID:            "O2",
			Total:         20000,
			PaymentMethod: "CARD",
			Timestamp:     time.Now().UnixMilli(),
			IsVoided:      false,
			IsSettled:     false,
			StoreType:     "RETAIL",
		},
	)
	db.Create(
		&model.Order{
			ID:            "O3",
			Total:         5000,
			PaymentMethod: "QRIS",
			Timestamp:     time.Now().UnixMilli(),
			IsVoided:      false,
			IsSettled:     false,
			StoreType:     "RETAIL",
		},
	)

	summary, err := svc.GetUnsettledSummary("RETAIL", "")
	require.NoError(t, err)
	assert.Equal(t, 3, summary.Count)
	assert.Equal(t, 35000.0, summary.Total)
	assert.Equal(t, 1, summary.PaymentSummary["CASH"].Count)
	assert.Equal(t, 1, summary.PaymentSummary["CARD"].Count)
	assert.Equal(t, 1, summary.PaymentSummary["QRIS"].Count)
}

func TestSettlementService_Settle(t *testing.T) {
	svc, orderRepo, db := setupSettlementTest(t)

	db.Create(
		&model.Order{
			ID:            "O1",
			Total:         10000,
			PaymentMethod: "CASH",
			Timestamp:     time.Now().UnixMilli(),
			IsVoided:      false,
			IsSettled:     false,
			StoreType:     "RETAIL",
		},
	)
	db.Create(
		&model.Order{
			ID:            "O2",
			Total:         20000,
			PaymentMethod: "CARD",
			Timestamp:     time.Now().UnixMilli(),
			IsVoided:      false,
			IsSettled:     false,
			StoreType:     "RETAIL",
		},
	)

	settlement, err := svc.Settle("SETTLE-001", time.Now().UnixMilli(), "RETAIL", "")
	require.NoError(t, err)
	assert.Equal(t, "SETTLE-001", settlement.ID)
	assert.Equal(t, 2, settlement.BatchCount)
	assert.Equal(t, 30000.0, settlement.TotalAmount)
	assert.Equal(t, 10000.0, settlement.CashTotal)
	assert.Equal(t, 20000.0, settlement.CardTotal)
	assert.Equal(t, 0.0, settlement.QRISTotal)
	assert.Equal(t, "SUCCESS", settlement.Status)

	// Verify orders are marked settled
	orders, _ := orderRepo.FindUnsettledOrders("RETAIL", "", false)
	assert.Len(t, orders, 0)
}

func TestSettlementService_Settle_NoOrders(t *testing.T) {
	svc, _, _ := setupSettlementTest(t)

	_, err := svc.Settle("SETTLE-001", time.Now().UnixMilli(), "RETAIL", "")
	assert.Error(t, err)
	assert.Equal(t, "NO_UNSETTLED_ORDERS", err.Error())
}

func TestSettlementService_Settle_DoesNotIncludeVoided(t *testing.T) {
	svc, _, db := setupSettlementTest(t)

	db.Create(
		&model.Order{
			ID:            "O1",
			Total:         10000,
			PaymentMethod: "CASH",
			Timestamp:     time.Now().UnixMilli(),
			IsVoided:      true,
			IsSettled:     false,
			StoreType:     "RETAIL",
		},
	)
	db.Create(
		&model.Order{
			ID:            "O2",
			Total:         20000,
			PaymentMethod: "CARD",
			Timestamp:     time.Now().UnixMilli(),
			IsVoided:      false,
			IsSettled:     false,
			StoreType:     "RETAIL",
		},
	)

	settlement, err := svc.Settle("SETTLE-001", time.Now().UnixMilli(), "RETAIL", "")
	require.NoError(t, err)
	assert.Equal(t, 1, settlement.BatchCount)
	assert.Equal(t, 20000.0, settlement.TotalAmount)
}
