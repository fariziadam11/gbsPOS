package repository

import (
	"testing"
	"time"

	"gbs-pos-api/internal/model"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

func setupHoldDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	err = db.AutoMigrate(&model.PosHoldSession{})
	require.NoError(t, err)

	return db
}

func seedHold(db *gorm.DB, id string, status string, terminal string) {
	err := db.Create(&model.PosHoldSession{
		ID:         id,
		StoreType:  "RETAIL",
		TerminalID: terminal,
		Payload:    datatypes.JSON([]byte(`{"items":[{"name":"Chitato","qty":1}]}`)),
		Total:      10000,
		Status:     status,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}).Error

	require.NoError(nil, err)
}

func TestPosHoldRepository_CreateAndFindByID(t *testing.T) {
	db := setupHoldDB(t)
	repo := NewPosHoldRepository(db)

	session := &model.PosHoldSession{
		ID:         "HOLD-001",
		StoreType:  "RETAIL",
		TerminalID: "POS-001",
		Payload:    datatypes.JSON([]byte(`{"items":[]}`)),
		Total:      10000,
		Status:     "ACTIVE",
	}

	err := repo.Create(session)
	require.NoError(t, err)

	result, err := repo.FindByID("HOLD-001")
	require.NoError(t, err)

	assert.Equal(t, "HOLD-001", result.ID)
	assert.Equal(t, "POS-001", result.TerminalID)
	assert.Equal(t, "ACTIVE", result.Status)
}

func TestPosHoldRepository_FindActiveByTerminal(t *testing.T) {
	db := setupHoldDB(t)
	repo := NewPosHoldRepository(db)

	seedHold(db, "H1", "ACTIVE", "POS-001")
	seedHold(db, "H2", "RESUMED", "POS-001")
	seedHold(db, "H3", "ACTIVE", "POS-002")

	results, err := repo.FindActiveByTerminal("POS-001")

	require.NoError(t, err)

	assert.Len(t, results, 1)
	assert.Equal(t, "H1", results[0].ID)
	assert.Equal(t, "ACTIVE", results[0].Status)
}

func TestPosHoldRepository_UpdateStatus(t *testing.T) {
	db := setupHoldDB(t)
	repo := NewPosHoldRepository(db)

	seedHold(db, "H1", "ACTIVE", "POS-001")

	err := repo.UpdateStatus("H1", "RESUMED")
	require.NoError(t, err)

	var result model.PosHoldSession
	err = db.First(&result, "id = ?", "H1").Error
	require.NoError(t, err)

	assert.Equal(t, "RESUMED", result.Status)
}

func TestPosHoldRepository_Delete(t *testing.T) {
	db := setupHoldDB(t)
	repo := NewPosHoldRepository(db)

	seedHold(db, "H1", "ACTIVE", "POS-001")

	err := repo.Delete("H1")
	require.NoError(t, err)

	var result model.PosHoldSession
	err = db.First(&result, "id = ?", "H1").Error
	require.Error(t, err)
}

func TestPosHoldRepository_FindByID_NotFound(t *testing.T) {
	db := setupHoldDB(t)
	repo := NewPosHoldRepository(db)

	_, err := repo.FindByID("NOT-FOUND")
	require.Error(t, err)
}