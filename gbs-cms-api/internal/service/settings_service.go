package service

import (
	"gbs-cms-api/internal/repository"
)

type SettingsService struct {
	repo *repository.SettingsRepository
}

func NewSettingsService(repo *repository.SettingsRepository) *SettingsService {
	return &SettingsService{repo: repo}
}

func DefaultSettings() map[string]string {
	return map[string]string{
		"tax_rate":       "0.10",
		"receipt_header": "GBS POS",
		"receipt_footer": "Terima kasih telah berbelanja",
		"store_name":     "GBS Store",
		"currency":       "IDR",
	}
}

func (s *SettingsService) GetAll() (map[string]string, error) {
	settings, err := s.repo.GetAll()
	if err != nil {
		return nil, err
	}
	result := make(map[string]string)
	for _, setting := range settings {
		result[setting.Key] = setting.Value
	}
	defaults := DefaultSettings()
	for key, val := range defaults {
		if _, exists := result[key]; !exists {
			result[key] = val
		}
	}
	return result, nil
}

func (s *SettingsService) Update(updates map[string]string) (map[string]string, error) {
	if err := s.repo.UpsertAll(updates); err != nil {
		return nil, err
	}
	return s.GetAll()
}

func (s *SettingsService) Get(key string) string {
	setting, err := s.repo.GetByKey(key)
	if err != nil {
		defaults := DefaultSettings()
		if val, exists := defaults[key]; exists {
			return val
		}
		return ""
	}
	return setting.Value
}
