package service

import (
	"testing"

	"gbs-pos-api/internal/database"
	"gbs-pos-api/internal/dto"
	"gbs-pos-api/internal/model"
	"gbs-pos-api/internal/repository"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

func setupHoldService(t *testing.T) (*HoldService, *gorm.DB) {
	db, err := database.NewTestDB()
	require.NoError(t, err)

	err = db.AutoMigrate(&model.HoldSession{})
	require.NoError(t, err)

	repo := repository.NewHoldRepository(db)
	svc := NewHoldService(repo)

	return svc, db
}

func seedHold(t *testing.T, db *gorm.DB, id string, status model.HoldStatus, terminal string) {
	err := db.Create(&model.HoldSession{
		ID:         id,
		StoreType:  "RETAIL",
		TerminalID: terminal,
		Payload:    datatypes.JSON([]byte(`{"items":[{"name":"Chitato","qty":1}]}`)),
		Total:      10000,
		Status:     status,
	}).Error

	require.NoError(t, err)
}

func TestHoldService_Create(t *testing.T) {
	svc, _ := setupHoldService(t)

	req := dto.CreateHoldRequest{
		StoreType:  "RETAIL",
		TerminalID: "POS-001",
		Payload:    []byte(`{"items":[{"name":"Chitato","qty":2}]}`),
		Total:      20000,
	}

	session, err := svc.Create(req)

	require.NoError(t, err)

	assert.Equal(t, "RETAIL", session.StoreType)
	assert.Equal(t, "POS-001", session.TerminalID)
	assert.JSONEq(t, string(req.Payload), string(session.Payload))
	assert.Equal(t, model.HoldStatusActive, session.Status)
	assert.NotEmpty(t, session.ID)
}

func TestHoldService_Get(t *testing.T) {
	svc, db := setupHoldService(t)

	seedHold(t, db, "H1", model.HoldStatusActive, "POS-001")

	result, err := svc.Get("H1")

	require.NoError(t, err)

	assert.Equal(t, "H1", result.ID)
	assert.Equal(t, model.HoldStatusActive, result.Status)
}

func TestHoldService_List(t *testing.T) {
	svc, db := setupHoldService(t)

	seedHold(t, db, "H1", model.HoldStatusActive, "POS-001")
	seedHold(t, db, "H2", model.HoldStatusResumed, "POS-001")
	seedHold(t, db, "H3", model.HoldStatusActive, "POS-002")

	list, err := svc.List("POS-001")

	require.NoError(t, err)

	assert.Len(t, list, 1)
	assert.Equal(t, "H1", list[0].ID)
}

func TestHoldService_Resume_ActiveHold(t *testing.T) {
	svc, db := setupHoldService(t)

	seedHold(t, db, "H1", model.HoldStatusActive, "POS-001")

	session, err := svc.Resume("H1")

	require.NoError(t, err)

	assert.Equal(t, model.HoldStatusResumed, session.Status)

	var stored model.HoldSession
	require.NoError(t, db.First(&stored, "id = ?", "H1").Error)
	assert.Equal(t, model.HoldStatusResumed, stored.Status)
}

func TestHoldService_Resume_RejectResumedHold(t *testing.T) {
	svc, db := setupHoldService(t)

	seedHold(t, db, "H1", model.HoldStatusResumed, "POS-001")

	_, err := svc.Resume("H1")

	require.Error(t, err)
	assert.ErrorIs(t, err, ErrHoldCannotResumeNonActive)
}

func TestHoldService_Delete_ActiveHold(t *testing.T) {
	svc, db := setupHoldService(t)

	seedHold(t, db, "H1", model.HoldStatusActive, "POS-001")

	err := svc.Delete("H1")

	require.NoError(t, err)

	var result model.HoldSession
	err = db.First(&result, "id = ?", "H1").Error

	assert.Error(t, err)
}

func TestHoldService_Delete_RejectResumedHold(t *testing.T) {
	svc, db := setupHoldService(t)

	seedHold(t, db, "H1", model.HoldStatusResumed, "POS-001")

	err := svc.Delete("H1")

	require.Error(t, err)
	assert.ErrorIs(t, err, ErrHoldCannotDeleteNonActive)

	var stored model.HoldSession
	require.NoError(t, db.First(&stored, "id = ?", "H1").Error)
	assert.Equal(t, model.HoldStatusResumed, stored.Status)
}
