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

type seedSettlementRequest struct {
	ID         string
	StoreType  string
	Total      float64
	Timestamp  int64
	TerminalID string
	Status     string
}

func setupSettlementRepoTestDB(t *testing.T) *gorm.DB {

	db, err := gorm.Open(
		sqlite.Open(":memory:"),
		&gorm.Config{},
	)

	require.NoError(t, err)

	err = db.AutoMigrate(&model.Settlement{})

	require.NoError(t, err)

	return db
}

func seedSettlement(
	t *testing.T,
	db *gorm.DB,
	req seedSettlementRequest,
) {

	settlement := model.Settlement{
		ID:          req.ID,
		Timestamp:   req.Timestamp,
		BatchCount:  1,
		TotalAmount: req.Total,
		CardTotal:   0,
		QRISTotal:   0,
		CashTotal:   req.Total,
		Status:      req.Status,
		StoreType:   req.StoreType,
		TerminalID:  req.TerminalID,
	}

	err := db.Create(&settlement).Error

	require.NoError(t, err)
}

func TestSettlementRepository_FindByID(t *testing.T) {

	db := setupSettlementRepoTestDB(t)

	repo := NewSettlementRepository(db)

	seedSettlement(t, db, seedSettlementRequest{
		ID:         "SETTLEMENT-001",
		StoreType:  "RETAIL",
		Total:      100000,
		Timestamp:  time.Now().UnixMilli(),
		TerminalID: "POS-001",
		Status:     "SUCCESS",
	})

	result, err := repo.FindByID("SETTLEMENT-001")

	require.NoError(t, err)

	assert.Equal(
		t,
		"SETTLEMENT-001",
		result.ID,
	)

	assert.Equal(
		t,
		100000.0,
		result.TotalAmount,
	)
}

func TestSettlementRepository_FindAll_FilterStoreType(t *testing.T) {

	db := setupSettlementRepoTestDB(t)

	repo := NewSettlementRepository(db)

	now := time.Now().UnixMilli()

	seedSettlement(t, db, seedSettlementRequest{
		ID:         "SETTLEMENT-001",
		StoreType:  "RETAIL",
		Total:      100000,
		Timestamp:  now,
		TerminalID: "POS-001",
		Status:     "SUCCESS",
	})

	seedSettlement(t, db, seedSettlementRequest{
		ID:         "SETTLEMENT-002",
		StoreType:  "FNB",
		Total:      200000,
		Timestamp:  now + 1,
		TerminalID: "POS-001",
		Status:     "SUCCESS",
	})

	result, err := repo.FindAll(
		0,
		"RETAIL",
	)

	require.NoError(t, err)

	assert.Len(t, result, 1)

	assert.Equal(
		t,
		"SETTLEMENT-001",
		result[0].ID,
	)
}

func TestSettlementRepository_FindAll_Limit(t *testing.T) {

	db := setupSettlementRepoTestDB(t)

	repo := NewSettlementRepository(db)

	now := time.Now().UnixMilli()

	seedSettlement(t, db, seedSettlementRequest{
		ID:         "SETTLEMENT-001",
		StoreType:  "RETAIL",
		Total:      100000,
		Timestamp:  now,
		TerminalID: "POS-001",
		Status:     "SUCCESS",
	})

	seedSettlement(t, db, seedSettlementRequest{
		ID:         "SETTLEMENT-002",
		StoreType:  "RETAIL",
		Total:      200000,
		Timestamp:  now + 1,
		TerminalID: "POS-001",
		Status:     "SUCCESS",
	})

	result, err := repo.FindAll(
		1,
		"",
	)

	require.NoError(t, err)

	assert.Len(t, result, 1)

	assert.Equal(
		t,
		"SETTLEMENT-002",
		result[0].ID,
	)
}