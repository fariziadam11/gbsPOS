package service

import (
	"gbs-pos-api/internal/repository"

	"gorm.io/datatypes"
)

type CartDisplayService struct {
	repo *repository.CartDisplayRepository
}

func NewCartDisplayService(repo *repository.CartDisplayRepository) *CartDisplayService {
	return &CartDisplayService{repo: repo}
}

// Save persists the latest cart display payload for a terminal.
func (s *CartDisplayService) Save(terminalID string, payload datatypes.JSON) error {
	return s.repo.Upsert(terminalID, payload)
}

// Get returns the stored payload for a terminal, or an error if none exists.
func (s *CartDisplayService) Get(terminalID string) (datatypes.JSON, error) {
	display, err := s.repo.GetByTerminalID(terminalID)
	if err != nil {
		return nil, err
	}

	return display.Payload, nil
}

// Delete removes the stored payload for a terminal.
func (s *CartDisplayService) Delete(terminalID string) error {
	return s.repo.Delete(terminalID)
}
