package repository

import (
	"testing"

	"gbs-pos-api/internal/database"
	"gbs-pos-api/internal/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

func setupCartDisplayRepoTest(t *testing.T) *gorm.DB {
	db, err := database.NewTestDB()
	require.NoError(t, err)

	return db
}

func seedCartDisplay(t *testing.T, db *gorm.DB, terminalID string, payload datatypes.JSON) {
	err := db.Create(&model.CartDisplay{
		TerminalID: terminalID,
		Payload:    payload,
	}).Error

	require.NoError(t, err)
}

func TestCartDisplayRepository_Upsert_Insert(t *testing.T) {
	db := setupCartDisplayRepoTest(t)
	repo := NewCartDisplayRepository(db)

	payload := datatypes.JSON([]byte(`{"Initial":{},"DaftarBelanja":[],"Summary":{},"TeksSelesai":"Transaksi"}`))

	err := repo.Upsert("POS-001", payload, nil)
	require.NoError(t, err)

	var stored model.CartDisplay
	err = db.Where("terminal_id = ?", "POS-001").First(&stored).Error
	require.NoError(t, err)

	assert.Equal(t, "POS-001", stored.TerminalID)
	assert.JSONEq(t, string(payload), string(stored.Payload))
}

func TestCartDisplayRepository_Upsert_Update(t *testing.T) {
	db := setupCartDisplayRepoTest(t)
	repo := NewCartDisplayRepository(db)

	initial := datatypes.JSON([]byte(`{"TeksSelesai":"Initial"}`))
	updated := datatypes.JSON([]byte(`{"TeksSelesai":"Updated"}`))

	seedCartDisplay(t, db, "POS-001", initial)

	err := repo.Upsert("POS-001", updated, nil)
	require.NoError(t, err)

	var stored model.CartDisplay
	err = db.Where("terminal_id = ?", "POS-001").First(&stored).Error
	require.NoError(t, err)

	assert.JSONEq(t, string(updated), string(stored.Payload))
}

func TestCartDisplayRepository_GetByTerminalID_Exists(t *testing.T) {
	db := setupCartDisplayRepoTest(t)
	repo := NewCartDisplayRepository(db)

	payload := datatypes.JSON([]byte(`{"Initial":{"NamaKasir":"DOMAR"}}`))
	seedCartDisplay(t, db, "POS-002", payload)

	result, err := repo.GetByTerminalID("POS-002")

	require.NoError(t, err)
	assert.Equal(t, "POS-002", result.TerminalID)
	assert.JSONEq(t, string(payload), string(result.Payload))
}

func TestCartDisplayRepository_GetByTerminalID_NotFound(t *testing.T) {
	db := setupCartDisplayRepoTest(t)
	repo := NewCartDisplayRepository(db)

	_, err := repo.GetByTerminalID("UNKNOWN")

	require.Error(t, err)
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

func TestCartDisplayRepository_Delete_Exists(t *testing.T) {
	db := setupCartDisplayRepoTest(t)
	repo := NewCartDisplayRepository(db)

	seedCartDisplay(t, db, "POS-003", datatypes.JSON([]byte(`{}`)))

	err := repo.Delete("POS-003")
	require.NoError(t, err)

	var result model.CartDisplay
	err = db.Where("terminal_id = ?", "POS-003").First(&result).Error
	require.Error(t, err)
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

func TestCartDisplayRepository_Delete_NotFound(t *testing.T) {
	db := setupCartDisplayRepoTest(t)
	repo := NewCartDisplayRepository(db)

	err := repo.Delete("UNKNOWN")

	require.Error(t, err)
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
}
