package service

import (
	"fmt"
	"gbs-cms-api/internal/dto"
	"gbs-cms-api/internal/model"
	"gbs-cms-api/internal/repository"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	repo *repository.UserRepository
}

func NewUserService(repo *repository.UserRepository) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) List() ([]dto.UserResponse, error) {
	users, err := s.repo.FindAll()
	if err != nil {
		return nil, err
	}
	result := make([]dto.UserResponse, len(users))
	for i, u := range users {
		result[i] = dto.UserResponse{
			ID:        u.ID,
			Username:  u.Username,
			Name:      u.Name,
			Role:      u.Role,
			Gender:    u.Gender,
			CreatedAt: u.CreatedAt.Format(time.RFC3339),
			UpdatedAt: u.UpdatedAt.Format(time.RFC3339),
		}
	}
	return result, nil
}

func (s *UserService) Get(id uint) (*dto.UserResponse, error) {
	user, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}
	return mapUserToResponse(user), nil
}

func (s *UserService) FindByUsername(username string) (*dto.UserResponse, error) {
	user, err := s.repo.FindByUsername(username)
	if err != nil {
		return nil, err
	}
	return mapUserToResponse(user), nil
}

func mapUserToResponse(user *model.User) *dto.UserResponse {
	return &dto.UserResponse{
		ID:        user.ID,
		Username:  user.Username,
		Name:      user.Name,
		Role:      user.Role,
		Gender:    user.Gender,
		CreatedAt: user.CreatedAt.Format(time.RFC3339),
		UpdatedAt: user.UpdatedAt.Format(time.RFC3339),
	}
}

func (s *UserService) Create(req dto.CreateUserRequest) (*dto.UserResponse, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}
	user := &model.User{
		Username:     req.Username,
		PasswordHash: string(hashed),
		Name:         req.Name,
		Role:         req.Role,
		Gender:       req.Gender,
	}
	if err := s.repo.Create(user); err != nil {
		return nil, err
	}
	return s.Get(user.ID)
}

func (s *UserService) Update(id uint, req dto.UpdateUserRequest) (*dto.UserResponse, error) {
	user, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if req.Name != "" {
		user.Name = req.Name
	}
	if req.Role != "" {
		user.Role = req.Role
	}
	if req.Gender != "" {
		user.Gender = req.Gender
	}
	if req.Password != "" {
		hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			return nil, fmt.Errorf("failed to hash password: %w", err)
		}
		user.PasswordHash = string(hashed)
	}
	if err := s.repo.Update(user); err != nil {
		return nil, err
	}
	return s.Get(user.ID)
}

func (s *UserService) Delete(id uint) error {
	return s.repo.Delete(id)
}
