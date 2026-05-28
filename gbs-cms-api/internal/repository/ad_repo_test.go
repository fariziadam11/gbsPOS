package repository

import (
	"encoding/json"
	"testing"
	"time"

	"gbs-cms-api/internal/model"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func setupAdTestDB(t *testing.T) *gorm.DB {

	db, err := gorm.Open(
		sqlite.Open(":memory:"),
		&gorm.Config{},
	)

	if err != nil {
		t.Fatalf("failed connect db: %v", err)
	}

	err = db.AutoMigrate(&model.Ad{})

	if err != nil {
		t.Fatalf("failed migrate: %v", err)
	}

	return db
}

func TestFindActiveByStoreType(t *testing.T) {

	db := setupAdTestDB(t)

	repo := NewAdRepository(db)

	now := time.Now()

	yesterday := now.AddDate(0, 0, -1)
	tomorrow := now.AddDate(0, 0, 1)

	ads := []model.Ad{
		{
			Name:          "Valid Ad",
			Filename:      "valid.mp4",
			StoragePath:   "/ads/valid.mp4",
			FileSize:      1000,
			MimeType:      "video/mp4",
			StoreTypes:    []string{"indomaret"},
			IsActive:      true,
			StartDate:     &yesterday,
			EndDate:       &tomorrow,
			PlaylistOrder: 1,
			CreatedBy:     1,
		},
		{
			Name:          "Inactive Ad",
			Filename:      "inactive.mp4",
			StoragePath:   "/ads/inactive.mp4",
			FileSize:      1000,
			MimeType:      "video/mp4",
			StoreTypes:    []string{"indomaret"},
			IsActive:      false,
			StartDate:     &yesterday,
			EndDate:       &tomorrow,
			PlaylistOrder: 2,
			CreatedBy:     1,
		},
		{
			Name:          "Expired Ad",
			Filename:      "expired.mp4",
			StoragePath:   "/ads/expired.mp4",
			FileSize:      1000,
			MimeType:      "video/mp4",
			StoreTypes:    []string{"indomaret"},
			IsActive:      true,
			StartDate:     &yesterday,
			EndDate:       &yesterday,
			PlaylistOrder: 3,
			CreatedBy:     1,
		},
		{
			Name:          "Different Store",
			Filename:      "other.mp4",
			StoragePath:   "/ads/other.mp4",
			FileSize:      1000,
			MimeType:      "video/mp4",
			StoreTypes:    []string{"alfamart"},
			IsActive:      true,
			StartDate:     &yesterday,
			EndDate:       &tomorrow,
			PlaylistOrder: 4,
			CreatedBy:     1,
		},
	}

	for _, ad := range ads {
		err := db.Select("*").Create(&ad).Error

		if err != nil {
			t.Fatalf("failed seed ad: %v", err)
		}
	}

	// var dbAds []model.Ad

	// db.Find(&dbAds)

	// b2, _ := json.MarshalIndent(dbAds, "", "  ")

	// t.Fatalf("DB DATA:\n%s", string(b2))

	result, err := repo.FindActiveByStoreType("indomaret")

	b, _ := json.MarshalIndent(result, "", "  ")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != 1 {
		t.Fatalf("expected 1 ad, got %d, res: %s", len(result), b)
	}

	if result[0].Name != "Valid Ad" {
		t.Errorf(
			"expected Valid Ad, got %s",
			result[0].Name,
		)
	}
}