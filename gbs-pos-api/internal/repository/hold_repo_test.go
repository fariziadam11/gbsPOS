package repository

import (
	"testing"
	"time"

	"gbs-pos-api/internal/database"
	"gbs-pos-api/internal/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

func setupHoldDB(t *testing.T) *gorm.DB {
	db, err := database.NewTestDB()
	require.NoError(t, err)

	return db
}

func seedHold(t *testing.T, db *gorm.DB, id string, status model.HoldStatus, terminal string) {
	err := db.Create(&model.HoldSession{
		ID:         id,
		StoreType:  "RETAIL",
		TerminalID: terminal,
		Payload:    datatypes.JSON([]byte(`{"items":[{"name":"Chitato","qty":1}]}`)),
		Total:      10000,
		Status:     status,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}).Error

	require.NoError(t, err)
}

func TestHoldRepository_CreateAndFindByID(t *testing.T) {
	db := setupHoldDB(t)
	repo := NewHoldRepository(db)

	session := &model.HoldSession{
		ID:         "HOLD-001",
		StoreType:  "RETAIL",
		TerminalID: "POS-001",
		Payload:    datatypes.JSON([]byte(`{"items":[]}`)),
		Total:      10000,
		Status:     model.HoldStatusActive,
	}

	err := repo.Create(session)
	require.NoError(t, err)

	result, err := repo.FindByID("HOLD-001")
	require.NoError(t, err)

	assert.Equal(t, "HOLD-001", result.ID)
	assert.Equal(t, "POS-001", result.TerminalID)
	assert.Equal(t, model.HoldStatusActive, result.Status)
}

func TestHoldRepository_FindActiveByTerminal(t *testing.T) {
	db := setupHoldDB(t)
	repo := NewHoldRepository(db)

	seedHold(t, db, "H1", model.HoldStatusActive, "POS-001")
	seedHold(t, db, "H2", model.HoldStatusResumed, "POS-001")
	seedHold(t, db, "H3", model.HoldStatusActive, "POS-002")

	results, err := repo.FindActiveByTerminal("POS-001")

	require.NoError(t, err)

	assert.Len(t, results, 1)
	assert.Equal(t, "H1", results[0].ID)
	assert.Equal(t, model.HoldStatusActive, results[0].Status)
}

func TestHoldRepository_UpdateStatus(t *testing.T) {
	db := setupHoldDB(t)
	repo := NewHoldRepository(db)

	seedHold(t, db, "H1", model.HoldStatusActive, "POS-001")

	err := repo.UpdateStatus("H1", model.HoldStatusResumed)
	require.NoError(t, err)

	var result model.HoldSession
	err = db.First(&result, "id = ?", "H1").Error
	require.NoError(t, err)

	assert.Equal(t, model.HoldStatusResumed, result.Status)
}

func TestHoldRepository_Delete(t *testing.T) {
	db := setupHoldDB(t)
	repo := NewHoldRepository(db)

	seedHold(t, db, "H1", model.HoldStatusActive, "POS-001")

	err := repo.Delete("H1")
	require.NoError(t, err)

	var result model.HoldSession
	err = db.First(&result, "id = ?", "H1").Error
	require.Error(t, err)
}

func TestHoldRepository_FindByID_NotFound(t *testing.T) {
	db := setupHoldDB(t)
	repo := NewHoldRepository(db)

	_, err := repo.FindByID("NOT-FOUND")
	require.Error(t, err)
}

func TestHoldRepository_UpdateStatus_NotFound(t *testing.T) {
	db := setupHoldDB(t)
	repo := NewHoldRepository(db)

	err := repo.UpdateStatus("NOT-FOUND", model.HoldStatusResumed)
	require.Error(t, err)
}

func TestHoldRepository_Delete_NotFound(t *testing.T) {
	db := setupHoldDB(t)
	repo := NewHoldRepository(db)

	err := repo.Delete("NOT-FOUND")
	require.Error(t, err)
}
