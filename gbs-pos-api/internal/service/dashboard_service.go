package service

import (
	"gbs-pos-api/internal/dto"
	"gbs-pos-api/internal/repository"
	"time"
)

type DashboardService struct {
	repo *repository.DashboardRepository
}

func NewDashboardService(repo *repository.DashboardRepository) *DashboardService {
	return &DashboardService{repo: repo}
}

func (s *DashboardService) GetSummary(storeType string) (*dto.DashboardSummary, error) {
	since := time.Now().Truncate(24 * time.Hour)
	return s.repo.GetSummary(storeType, since)
}

func (s *DashboardService) GetRevenueTrend(storeType string, days int) ([]dto.RevenuePoint, error) {
	if days <= 0 {
		days = 7
	}
	return s.repo.GetRevenueTrend(storeType, days)
}

func (s *DashboardService) GetTopProducts(storeType string, limit int) ([]dto.TopProduct, error) {
	if limit <= 0 {
		limit = 10
	}
	return s.repo.GetTopProducts(storeType, limit)
}
