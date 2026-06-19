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

func (s *DashboardService) GetSummary(storeType string, startDate, endDate *time.Time) (*dto.DashboardSummary, error) {
	now := time.Now().UTC()
	start := now.Truncate(24 * time.Hour)
	end := start

	if startDate != nil {
		start = startDate.UTC().Truncate(24 * time.Hour)
	}
	if endDate != nil {
		end = endDate.UTC().Truncate(24 * time.Hour)
	}

	if end.Before(start) {
		end = start
	}

	return s.repo.GetSummary(storeType, start, end)
}

func (s *DashboardService) GetRevenueTrend(storeType string, startDate, endDate *time.Time, days int) ([]dto.RevenuePoint, error) {
	now := time.Now().UTC().Truncate(24 * time.Hour)

	start := now.AddDate(0, 0, -days+1)
	end := now

	if startDate != nil {
		start = startDate.UTC().Truncate(24 * time.Hour)
	}
	if endDate != nil {
		end = endDate.UTC().Truncate(24 * time.Hour)
	}

	if end.Before(start) {
		end = start
	}

	return s.repo.GetRevenueTrend(storeType, start, end)
}

func (s *DashboardService) GetTopProducts(storeType string, limit int) ([]dto.TopProduct, error) {
	if limit <= 0 {
		limit = 10
	}
	return s.repo.GetTopProducts(storeType, limit)
}
