package repository

import (
	"errors"
	"testing"

	"gbs-cms-api/internal/model"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func setupUserTestDB(t *testing.T) *gorm.DB {

	db, err := gorm.Open(
		sqlite.Open(":memory:"),
		&gorm.Config{},
	)

	if err != nil {
		t.Fatalf("failed connect db: %v", err)
	}

	err = db.AutoMigrate(&model.User{})

	if err != nil {
		t.Fatalf("failed migrate: %v", err)
	}

	return db
}

func TestFindByUsername(t *testing.T) {

	db := setupUserTestDB(t)

	repo := NewUserRepository(db)

	user := model.User{
		Username:     "rido",
		PasswordHash: "hashed-password",
		Name:         "Rido",
		Role:         "admin",
	}

	err := db.Create(&user).Error

	if err != nil {
		t.Fatalf("failed seed user: %v", err)
	}

	result, err := repo.FindByUsername("rido")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Username != "rido" {
		t.Errorf(
			"expected username rido, got %s",
			result.Username,
		)
	}
}

func TestFindByUsername_NotFound(t *testing.T) {

	db := setupUserTestDB(t)

	repo := NewUserRepository(db)

	result, err := repo.FindByUsername("unknown")

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf(
			"expected ErrRecordNotFound, got %v",
			err,
		)
	}

	if result != nil {
		t.Fatalf("expected nil result")
	}
}