package service

import (
	"gbs-cms-api/internal/dto"
	"gbs-cms-api/internal/repository"
)

type CartDisplayService struct {
	repo *repository.CartDisplayRepository
}

func NewCartDisplayService(repo *repository.CartDisplayRepository) *CartDisplayService {
	return &CartDisplayService{repo: repo}
}

// SaveCartDisplay persists the latest cart display payload for a terminal.
func (s *CartDisplayService) SaveCartDisplay(terminalID string, payload string, deviceInfo *dto.DeviceInfo) error {
	return s.repo.Upsert(terminalID, payload, deviceInfo)
}

// GetCartDisplay returns the stored payload for a terminal.
func (s *CartDisplayService) GetCartDisplay(terminalID string) (string, error) {
	display, err := s.repo.GetByTerminalID(terminalID)
	if err != nil {
		return "", err
	}

	return display.Payload, nil
}

// GetAllTerminals returns all cart displays for the terminal list page.
func (s *CartDisplayService) GetAllTerminals() ([]dto.TerminalListItem, error) {
	displays, err := s.repo.GetAll()
	if err != nil {
		return nil, err
	}

	result := make([]dto.TerminalListItem, len(displays))
	for i, d := range displays {
		result[i] = dto.TerminalListItem{
			TerminalID:         d.TerminalID,
			Payload:           d.Payload,
			DeviceModel:       d.DeviceModel,
			DeviceManufacturer: d.DeviceManufacturer,
			DeviceBrand:       d.DeviceBrand,
			AndroidVersion:    d.AndroidVersion,
			SDKInt:            d.SDKInt,
			AppVersion:        d.AppVersion,
			AppVersionCode:    d.AppVersionCode,
			UpdatedAt:         d.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
	}

	return result, nil
}
