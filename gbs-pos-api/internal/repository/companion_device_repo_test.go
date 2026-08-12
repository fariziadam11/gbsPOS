package repository

import (
	"context"
	"testing"
	"time"

	"gbs-pos-api/internal/model"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestCompanionDeviceRepository_Upsert(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.CompanionDevice{}))

	repo := NewCompanionDeviceRepository(db)
	require.NoError(t, repo.Upsert(context.Background(), &model.CompanionDevice{
		DeviceID: "HP-001", DeviceName: "Old", LastSeenAt: time.Now(),
	}))
	require.NoError(t, repo.Upsert(context.Background(), &model.CompanionDevice{
		DeviceID: "HP-001", DeviceName: "New", SDKVersion: "36", LastSeenAt: time.Now(),
	}))

	var device model.CompanionDevice
	require.NoError(t, db.First(&device, "device_id = ?", "HP-001").Error)
	require.Equal(t, "New", device.DeviceName)
	require.Equal(t, "36", device.SDKVersion)
}
