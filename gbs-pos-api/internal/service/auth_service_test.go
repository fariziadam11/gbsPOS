package service

import (
	"os"
	"testing"

	"gbs-pos-api/internal/database"
	"gbs-pos-api/internal/model"
	"gbs-pos-api/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func setupAuthTest(t *testing.T) (*AuthService, *gorm.DB) {
	db, err := database.NewTestDB()
	require.NoError(t, err)

	// Seed test user
	hash, _ := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	db.Create(&model.User{Username: "admin", PasswordHash: string(hash), Name: "Admin", Role: "ADMIN"})

	os.Setenv("JWT_SECRET", "test-secret-key-minimum-32-characters")
	os.Setenv("JWT_EXPIRY_HOURS", "24")

	userRepo := repository.NewUserRepository(db)
	return NewAuthService(userRepo), db
}

func TestAuthService_Login_Success(t *testing.T) {
	svc, _ := setupAuthTest(t)

	result, err := svc.Login("admin", "admin123")
	require.NoError(t, err)
	assert.NotNil(t, result.User)
	assert.Equal(t, "admin", result.User.Username)
	assert.Equal(t, "ADMIN", result.User.Role)
	assert.NotEmpty(t, result.Token)
}

func TestAuthService_Login_InvalidPassword(t *testing.T) {
	svc, _ := setupAuthTest(t)

	_, err := svc.Login("admin", "wrongpassword")
	assert.Error(t, err)
	assert.Equal(t, "INVALID_CREDENTIALS", err.Error())
}

func TestAuthService_Login_UserNotFound(t *testing.T) {
	svc, _ := setupAuthTest(t)

	_, err := svc.Login("nonexistent", "admin123")
	assert.Error(t, err)
	assert.Equal(t, "INVALID_CREDENTIALS", err.Error())
}
