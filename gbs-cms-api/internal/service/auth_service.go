package service

import (
	"fmt"
	"os"
	"strconv"
	"time"
	"gbs-cms-api/internal/model"
	"gbs-cms-api/internal/repository"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	userRepo *repository.UserRepository
}

func NewAuthService(userRepo *repository.UserRepository) *AuthService {
	return &AuthService{userRepo: userRepo}
}

type LoginResult struct {
	User  *model.User
	Token string
}

func (s *AuthService) Login(username, password string) (*LoginResult, error) {
	user, err := s.userRepo.FindByUsername(username)
	if err != nil {
		return nil, fmt.Errorf("INVALID_CREDENTIALS")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, fmt.Errorf("INVALID_CREDENTIALS")
	}
	expiryHours, _ := strconv.Atoi(os.Getenv("JWT_EXPIRY_HOURS"))
	if expiryHours == 0 {
		expiryHours = 24
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":      user.ID,
		"username": user.Username,
		"role":     user.Role,
		"iat":      time.Now().Unix(),
		"exp":      time.Now().Add(time.Duration(expiryHours) * time.Hour).Unix(),
	})
	tokenString, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		return nil, err
	}
	return &LoginResult{User: user, Token: tokenString}, nil
}
