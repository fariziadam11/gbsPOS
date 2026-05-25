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
	id string,
	storeType string,
	total float64,
	timestamp int64,
) {

	settlement := model.Settlement{
		ID:          id,
		Timestamp:   timestamp,
		BatchCount:  1,
		TotalAmount: total,
		CardTotal:   0,
		QRISTotal:   0,
		CashTotal:   total,
		Status:      "SUCCESS",
		StoreType:   storeType,
		TerminalID:  "POS-001",
	}

	err := db.Create(&settlement).Error

	require.NoError(t, err)
}

func TestSettlementRepository_FindByID(t *testing.T) {

	db := setupSettlementRepoTestDB(t)

	repo := NewSettlementRepository(db)

	seedSettlement(
		t,
		db,
		"SETTLEMENT-001",
		"RETAIL",
		100000,
		time.Now().UnixMilli(),
	)

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

	seedSettlement(
		t,
		db,
		"SETTLEMENT-001",
		"RETAIL",
		100000,
		now,
	)

	seedSettlement(
		t,
		db,
		"SETTLEMENT-002",
		"FNB",
		200000,
		now+1,
	)

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

	seedSettlement(
		t,
		db,
		"SETTLEMENT-001",
		"RETAIL",
		100000,
		now,
	)

	seedSettlement(
		t,
		db,
		"SETTLEMENT-002",
		"RETAIL",
		200000,
		now+1,
	)

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