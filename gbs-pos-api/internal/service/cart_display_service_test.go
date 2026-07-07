package service

import (
	"testing"

	"gbs-pos-api/internal/database"
	"gbs-pos-api/internal/model"
	"gbs-pos-api/internal/repository"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

func setupCartDisplayServiceTest(t *testing.T) (*CartDisplayService, *gorm.DB) {
	db, err := database.NewTestDB()
	require.NoError(t, err)

	repo := repository.NewCartDisplayRepository(db)
	svc := NewCartDisplayService(repo)

	return svc, db
}

func seedCartDisplay(t *testing.T, db *gorm.DB, terminalID string, payload datatypes.JSON) {
	err := db.Create(&model.CartDisplay{
		TerminalID: terminalID,
		Payload:    payload,
	}).Error

	require.NoError(t, err)
}

func TestCartDisplayService_Save(t *testing.T) {
	svc, db := setupCartDisplayServiceTest(t)

	payload := datatypes.JSON([]byte(`{"TeksSelesai":"Transaksi"}`))

	err := svc.Save("POS-001", payload)
	require.NoError(t, err)

	var stored model.CartDisplay
	err = db.Where("terminal_id = ?", "POS-001").First(&stored).Error
	require.NoError(t, err)

	assert.JSONEq(t, string(payload), string(stored.Payload))
}

func TestCartDisplayService_Get(t *testing.T) {
	svc, db := setupCartDisplayServiceTest(t)

	payload := datatypes.JSON([]byte(`{"DaftarBelanja":[{"name":"Chitato"}]}`))
	seedCartDisplay(t, db, "POS-002", payload)

	result, err := svc.Get("POS-002")

	require.NoError(t, err)
	assert.JSONEq(t, string(payload), string(result))
}

func TestCartDisplayService_Get_NotFound(t *testing.T) {
	svc, _ := setupCartDisplayServiceTest(t)

	_, err := svc.Get("UNKNOWN")

	require.Error(t, err)
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

func TestCartDisplayService_Delete(t *testing.T) {
	svc, db := setupCartDisplayServiceTest(t)

	seedCartDisplay(t, db, "POS-003", datatypes.JSON([]byte(`{}`)))

	err := svc.Delete("POS-003")
	require.NoError(t, err)

	var result model.CartDisplay
	err = db.Where("terminal_id = ?", "POS-003").First(&result).Error
	require.Error(t, err)
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
}
