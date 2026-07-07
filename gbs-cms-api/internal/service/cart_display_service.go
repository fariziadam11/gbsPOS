package service

import (
	"gbs-cms-api/internal/repository"
)

type CartDisplayService struct {
	repo *repository.CartDisplayRepository
}

func NewCartDisplayService(repo *repository.CartDisplayRepository) *CartDisplayService {
	return &CartDisplayService{repo: repo}
}

// SaveCartDisplay persists the latest cart display payload for a terminal.
func (s *CartDisplayService) SaveCartDisplay(terminalID string, payload string) error {
	return s.repo.Upsert(terminalID, payload)
}

// GetCartDisplay returns the stored payload for a terminal.
func (s *CartDisplayService) GetCartDisplay(terminalID string) (string, error) {
	display, err := s.repo.GetByTerminalID(terminalID)
	if err != nil {
		return "", err
	}

	return display.Payload, nil
}
