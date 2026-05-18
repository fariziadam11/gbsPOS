package service

import (
	"os"
	"testing"
	"time"

	"gbs-cms-api/internal/database"
	"gbs-cms-api/internal/model"
	"gbs-cms-api/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func setupCMSTest(t *testing.T) (*CMSService, *gorm.DB) {
	db, err := database.NewTestDB()
	require.NoError(t, err)

	adRepo := repository.NewAdRepository(db)
	playLogRepo := repository.NewAdPlayLogRepository(db)
	return NewCMSService(adRepo, playLogRepo, "./test-uploads"), db
}

func TestCMSService_CreateAd(t *testing.T) {
	svc, _ := setupCMSTest(t)

	startDate := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2025, 1, 31, 0, 0, 0, 0, time.UTC)

	ad, err := svc.CreateAd("Indomie Promo", "indomie.mp4", "video/mp4", 5242880, []string{"RETAIL"}, 0, &startDate, &endDate, nil, nil, 1)
	require.NoError(t, err)
	assert.Equal(t, "Indomie Promo", ad.Name)
	assert.Equal(t, []string{"RETAIL"}, ad.StoreTypes)
	assert.True(t, ad.IsActive)
}

func TestCMSService_CreateAd_InvalidSchedule(t *testing.T) {
	svc, _ := setupCMSTest(t)

	startDate := time.Date(2025, 2, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)

	_, err := svc.CreateAd("Invalid", "file.mp4", "video/mp4", 1000, []string{"RETAIL"}, 0, &startDate, &endDate, nil, nil, 1)
	assert.Error(t, err)
	assert.Equal(t, "INVALID_SCHEDULE", err.Error())
}

func TestCMSService_UpdateAd(t *testing.T) {
	svc, db := setupCMSTest(t)

	ad := &model.Ad{Name: "Old Name", Filename: "old.mp4", StoragePath: "/uploads/old.mp4", FileSize: 1000, MimeType: "video/mp4", StoreTypes: []string{"RETAIL"}, PlaylistOrder: 0, IsActive: true, CreatedBy: 1}
	db.Create(ad)

	updated, err := svc.UpdateAd(ad.ID, &model.Ad{Name: "New Name", StoreTypes: []string{"RETAIL", "FNB"}, PlaylistOrder: 1, IsActive: false})
	require.NoError(t, err)
	assert.Equal(t, "New Name", updated.Name)
	assert.Equal(t, []string{"RETAIL", "FNB"}, updated.StoreTypes)
	assert.Equal(t, 1, updated.PlaylistOrder)
	assert.False(t, updated.IsActive)
}

func TestCMSService_ToggleAd(t *testing.T) {
	svc, db := setupCMSTest(t)

	ad := &model.Ad{Name: "Ad", Filename: "ad.mp4", StoragePath: "/uploads/ad.mp4", FileSize: 1000, MimeType: "video/mp4", StoreTypes: []string{"RETAIL"}, IsActive: true, CreatedBy: 1}
	db.Create(ad)

	toggled, err := svc.ToggleAd(ad.ID)
	require.NoError(t, err)
	assert.False(t, toggled.IsActive)
}

func TestCMSService_GetActivePlaylist(t *testing.T) {
	svc, db := setupCMSTest(t)

	now := time.Now()
	startDate := now.AddDate(0, 0, -1)
	endDate := now.AddDate(0, 0, 1)

	ad := &model.Ad{Name: "Active Ad", Filename: "active.mp4", StoragePath: "/uploads/active.mp4", FileSize: 1000, MimeType: "video/mp4", StoreTypes: []string{"RETAIL"}, IsActive: true, StartDate: &startDate, EndDate: &endDate, CreatedBy: 1}
	db.Create(ad)

	// Inactive ad
	inactive := &model.Ad{Name: "Inactive", Filename: "inactive.mp4", StoragePath: "/uploads/inactive.mp4", FileSize: 1000, MimeType: "video/mp4", StoreTypes: []string{"RETAIL"}, IsActive: false, CreatedBy: 1}
	db.Select("is_active").Create(inactive)

	playlist, err := svc.GetActivePlaylist("RETAIL")
	require.NoError(t, err)
	assert.Len(t, playlist, 1)
	assert.Equal(t, "Active Ad", playlist[0].Name)
}

func TestCMSService_ValidateUploadFile(t *testing.T) {
	svc, _ := setupCMSTest(t)

	assert.NoError(t, svc.ValidateUploadFile("video.mp4", 10*1024*1024))
	assert.Error(t, svc.ValidateUploadFile("video.mp4", 60*1024*1024))
	assert.Error(t, svc.ValidateUploadFile("document.pdf", 10*1024*1024))
}

func TestCMSService_LogPlay(t *testing.T) {
	svc, db := setupCMSTest(t)

	ad := &model.Ad{Name: "Ad", Filename: "ad.mp4", StoragePath: "/uploads/ad.mp4", FileSize: 1000, MimeType: "video/mp4", StoreTypes: []string{"RETAIL"}, IsActive: true, CreatedBy: 1}
	db.Create(ad)

	err := svc.LogPlay(ad.ID, "POS-001", "RETAIL")
	require.NoError(t, err)
}

func TestParseTimePointer(t *testing.T) {
	result := ParseTimePointer("09:00:00")
	require.NotNil(t, result)
	assert.Equal(t, 9, result.Hour())
	assert.Equal(t, 0, result.Minute())

	result2 := ParseTimePointer("21:00")
	require.NotNil(t, result2)
	assert.Equal(t, 21, result2.Hour())

	result3 := ParseTimePointer("")
	assert.Nil(t, result3)
}

func TestParseDatePointer(t *testing.T) {
	result := ParseDatePointer("2025-01-15")
	require.NotNil(t, result)
	assert.Equal(t, 2025, result.Year())
	assert.Equal(t, time.January, result.Month())

	result2 := ParseDatePointer("")
	assert.Nil(t, result2)
}

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}
