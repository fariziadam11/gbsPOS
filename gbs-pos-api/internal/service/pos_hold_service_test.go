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

func setupPosHoldService(t *testing.T) (*PosHoldService, *gorm.DB) {
	db, err := database.NewTestDB()
	require.NoError(t, err)

	err = db.AutoMigrate(&model.PosHoldSession{})
	require.NoError(t, err)

	repo := repository.NewPosHoldRepository(db)
	svc := NewPosHoldService(repo)

	return svc, db
}

func seedHold(db *gorm.DB, id string, status string, terminal string) {
	err := db.Create(&model.PosHoldSession{
		ID:         id,
		StoreType:  "RETAIL",
		TerminalID: terminal,
		Payload:    datatypes.JSON([]byte(`{"items":[{"name":"Chitato","qty":1}]}`)),
		Total:      10000,
		Status:     status,
	}).Error

	if err != nil {
		panic(err)
	}
}

func TestPosHoldService_Hold(t *testing.T) {
	svc, _ := setupPosHoldService(t)

	payload := map[string]any{
		"items": []map[string]any{
			{"name": "Chitato", "qty": 2},
		},
	}

	session, err := svc.Hold("RETAIL", "POS-001", payload, 20000)

	require.NoError(t, err)

	assert.Equal(t, "RETAIL", session.StoreType)
	assert.Equal(t, "POS-001", session.TerminalID)
	assert.Equal(t, model.PosHoldStatusActive, session.Status)
	assert.NotEmpty(t, session.ID)
}

func TestPosHoldService_Get(t *testing.T) {
	svc, db := setupPosHoldService(t)

	seedHold(db, "H1", model.PosHoldStatusActive, "POS-001")

	result, err := svc.Get("H1")

	require.NoError(t, err)

	assert.Equal(t, "H1", result.ID)
	assert.Equal(t, model.PosHoldStatusActive, result.Status)
}

func TestPosHoldService_List(t *testing.T) {
	svc, db := setupPosHoldService(t)

	seedHold(db, "H1", model.PosHoldStatusActive, "POS-001")
	seedHold(db, "H2", model.PosHoldStatusResumed, "POS-001")
	seedHold(db, "H3", model.PosHoldStatusActive, "POS-002")

	list, err := svc.List("POS-001")

	require.NoError(t, err)

	assert.Len(t, list, 1)
	assert.Equal(t, "H1", list[0].ID)
}

func TestPosHoldService_Resume_Success(t *testing.T) {
	svc, db := setupPosHoldService(t)

	seedHold(db, "H1", model.PosHoldStatusActive, "POS-001")

	session, err := svc.Resume("H1")

	require.NoError(t, err)

	assert.Equal(t, model.PosHoldStatusResumed, session.Status)
}

func TestPosHoldService_Resume_NotActive(t *testing.T) {
	svc, db := setupPosHoldService(t)

	seedHold(db, "H1", model.PosHoldStatusResumed, "POS-001")

	_, err := svc.Resume("H1")

	require.Error(t, err)
	assert.Equal(t, "CANNOT_RESUME_NON_ACTIVE_HOLD", err.Error())
}

func TestPosHoldService_Delete(t *testing.T) {
	svc, db := setupPosHoldService(t)

	seedHold(db, "H1", model.PosHoldStatusActive, "POS-001")

	err := svc.Delete("H1")

	require.NoError(t, err)

	var result model.PosHoldSession
	err = db.First(&result, "id = ?", "H1").Error

	assert.Error(t, err)
}

func TestPosHoldService_Delete_NotActive(t *testing.T) {
	svc, db := setupPosHoldService(t)

	seedHold(db, "H1", model.PosHoldStatusResumed, "POS-001")

	err := svc.Delete("H1")

	require.Error(t, err)
	assert.Equal(t, "CANNOT_DELETE_NON_ACTIVE_HOLD", err.Error())
}
