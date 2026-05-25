package service

import (
	"errors"
	"fmt"
	"gbs-pos-api/internal/dto"
	"gbs-pos-api/internal/model"
	"gbs-pos-api/internal/repository"
	"time"

	"gorm.io/gorm"
)

type OrderService struct {
	repo *repository.OrderRepository
}

func NewOrderService(repo *repository.OrderRepository) *OrderService {
	return &OrderService{repo: repo}
}

func (s *OrderService) List(
	storeType string,
	startDate, endDate int64,
	isVoided, isSettled *bool,
	paymentMethod, terminalID string,
) ([]model.Order, error) {
	return s.repo.FindAll(
		storeType,
		startDate,
		endDate,
		isVoided,
		isSettled,
		paymentMethod,
		terminalID,
	)
}

func (s *OrderService) Get(id string) (*model.Order, error) {
	return s.repo.FindByIDWithItems(id)
}

func (s *OrderService) Create(order *model.Order) (*model.Order, bool, error) {
	existing, err := s.repo.FindByID(order.ID)
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, false, err
		}
	} else if existing != nil {
		fullOrder, err := s.repo.FindByIDWithItems(order.ID)
		if err != nil {
			return nil, false, err
		}
		return fullOrder, true, nil
	}
	if err := s.repo.Create(order); err != nil {
		return nil, false, err
	}
	return order, false, nil
}

func (s *OrderService) Void(id, reason, voidedBy string) (*model.Order, error) {
	order, err := s.repo.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("ORDER_NOT_FOUND")
	}
	if order.IsVoided {
		return nil, fmt.Errorf("ORDER_ALREADY_VOIDED")
	}
	if order.IsSettled {
		return nil, fmt.Errorf("ORDER_ALREADY_SETTLED")
	}
	now := time.Now()
	order.IsVoided = true
	order.VoidReason = reason
	order.VoidedBy = voidedBy
	order.VoidedAt = &now
	if err := s.repo.UpdateVoid(order); err != nil {
		return nil, err
	}
	// Reload with items for consistent response format
	return s.repo.FindByIDWithItems(id)
}

func (s *OrderService) BulkCreate(orders []model.Order) (*dto.BulkSyncResult, error) {
	result := &dto.BulkSyncResult{
		Orders: make([]model.Order, 0, len(orders)),
	}
	for i := range orders {
		createdOrder, idempotent, err := s.Create(&orders[i])
		if err != nil {
			result.Failed++
			continue
		}
		if idempotent {
			result.Existing++
			result.Idempotent = true
		} else {
			result.Created++
		}
		result.Orders = append(result.Orders, *createdOrder)
	}
	return result, nil
}
