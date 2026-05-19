package service

import (
	"errors"
	"fmt"
	"gbs-pos-api/internal/model"
	"gbs-pos-api/internal/repository"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthService struct {
	userRepo       *repository.UserRepository
	jwtSecret      []byte
	jwtExpiryHours int
}

func NewAuthService(
	userRepo *repository.UserRepository,
	jwtSecret string,
	jwtExpiryHours int,
) *AuthService {
	if jwtExpiryHours == 0 {
		jwtExpiryHours = 24
	}
	return &AuthService{
		userRepo:       userRepo,
		jwtSecret:      []byte(jwtSecret),
		jwtExpiryHours: jwtExpiryHours,
	}
}

type LoginResult struct {
	User  *model.User
	Token string
}

func (s *AuthService) Login(username, password string) (*LoginResult, error) {
	user, err := s.userRepo.FindByUsername(username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("INVALID_CREDENTIALS")
		}
		return nil, fmt.Errorf("DB_ERROR: %w", err)
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, fmt.Errorf("INVALID_CREDENTIALS")
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":      user.ID,
		"username": user.Username,
		"role":     user.Role,
		"iat":      time.Now().Unix(),
		"exp":      time.Now().Add(time.Duration(s.jwtExpiryHours) * time.Hour).Unix(),
	})
	tokenString, err := token.SignedString(s.jwtSecret)
	if err != nil {
		return nil, err
	}
	return &LoginResult{User: user, Token: tokenString}, nil
}
