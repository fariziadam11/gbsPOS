package database

import (
	"gbs-cms-api/internal/model"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func NewTestDB() (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	if err := db.AutoMigrate(
		&model.User{},
		&model.Ad{},
		&model.AdPlayLog{},
	); err != nil {
		return nil, err
	}
	return db, nil
}
