package service

import (
	"errors"
	"fmt"
	"gbs-pos-api/internal/dto"
	"gbs-pos-api/internal/model"
	"gbs-pos-api/internal/repository"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var ErrOrderAlreadyExists = errors.New("ORDER_ALREADY_EXISTS")

type OrderService struct {
	repo              *repository.OrderRepository
	productService    *ProductService
	customerService   *CustomerService
	variantService    *ProductVariantService
}

func NewOrderService(
	repo *repository.OrderRepository,
	productService *ProductService,
	customerService *CustomerService,
	variantService *ProductVariantService,
) *OrderService {
	return &OrderService{
		repo:            repo,
		productService:  productService,
		customerService: customerService,
		variantService:  variantService,
	}
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

	// Resolve customer by phone if no CustomerID provided
	if order.CustomerID == nil && order.CustomerPhone != "" {
		customer, err := s.customerService.GetByPhone(order.CustomerPhone)
		if err != nil {
			// Customer not found — create new one
			newCustomer := &model.Customer{
				Name:  order.CustomerName,
				Phone: order.CustomerPhone,
			}
			if newCustomer.Name == "" {
				newCustomer.Name = "Pelanggan " + order.CustomerPhone
			}
			if err := s.customerService.Create(newCustomer); err == nil {
				cid := int(newCustomer.ID)
				order.CustomerID = &cid
			}
		} else {
			cid := int(customer.ID)
			order.CustomerID = &cid
		}
	}

	// Calculate loyalty points: 1% of total (rounded down)
	loyaltyPoints := int(order.Total / 100)
	if loyaltyPoints < 1 {
		loyaltyPoints = 0
	}
	order.LoyaltyPointsEarned = loyaltyPoints

	// Calculate discount
	if order.DiscountType != "" && order.DiscountValue != nil && *order.DiscountValue > 0 {
		preDiscountTotal := order.Total
		var discountAmount float64
		if order.DiscountType == "PERCENTAGE" {
			discountAmount = preDiscountTotal * (*order.DiscountValue) / 100
		} else {
			discountAmount = *order.DiscountValue
		}
		order.DiscountAmount = &discountAmount
		order.Total = preDiscountTotal - discountAmount
		if order.Total < 0 {
			order.Total = 0
		}
		// Recalculate loyalty points on discounted total
		loyaltyPoints = int(order.Total / 100)
		if loyaltyPoints < 1 {
			loyaltyPoints = 0
		}
		order.LoyaltyPointsEarned = loyaltyPoints
	}

	// Deduct stock in transaction
	if err := s.repo.Transaction(func(tx *gorm.DB) error {
		txRepo := s.repo.WithTx(tx)

		// Lock the order row to prevent race condition on idempotency check
		// This ensures only one request can create an order with a given ID
		lockErr := tx.Raw("SELECT id FROM orders WHERE id = ? FOR UPDATE", order.ID).Error
		if lockErr != nil && !errors.Is(lockErr, gorm.ErrRecordNotFound) {
			return lockErr
		}

		// Check if order already exists after acquiring lock
		existingCheck, _ := txRepo.FindByID(order.ID)
		if existingCheck != nil {
			fullOrder, err := txRepo.FindByIDWithItems(order.ID)
			if err != nil {
				return err
			}
			return ErrOrderAlreadyExists
		}

		if err := txRepo.Create(order); err != nil {
			return err
		}
		for _, item := range order.Items {
			if item.VariantID != nil {
				if err := s.variantService.DeductVariantStock(tx, *item.VariantID, item.Qty, order.ID); err != nil {
					return err
				}
			}
			if err := s.productService.DeductStock(tx, item.ProductID, item.Qty, order.ID); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		if errors.Is(err, ErrOrderAlreadyExists) {
			// Return idempotent response
			existingOrder, _ := s.repo.FindByIDWithItems(order.ID)
			if existingOrder != nil {
				return existingOrder, true, nil
			}
		}
		return nil, false, err
	}

	// Add loyalty points to customer (outside transaction to avoid lock contention)
	if order.CustomerID != nil && loyaltyPoints > 0 {
		if err := s.customerService.AddLoyaltyPoints(uint(*order.CustomerID), loyaltyPoints); err != nil {
			// Log error but don't fail the order creation
			// In production, use structured logging: log.Printf("failed to add loyalty points: %v", err)
		}
	}

	return order, false, nil
}

func (s *OrderService) Void(id, reason, voidedBy string) (*model.Order, error) {
	order, err := s.repo.FindByIDWithItems(id)
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

	// Restore stock in transaction
	if err := s.repo.Transaction(func(tx *gorm.DB) error {
		txRepo := s.repo.WithTx(tx)
		if err := txRepo.UpdateVoid(order); err != nil {
			return err
		}
		for _, item := range order.Items {
			if item.VariantID != nil {
				if err := s.variantService.RestoreVariantStock(tx, *item.VariantID, item.Qty, order.ID); err != nil {
					return err
				}
			}
			if err := s.productService.RestoreStock(tx, item.ProductID, item.Qty, order.ID); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}

	// Deduct loyalty points earned from this order
	if order.CustomerID != nil && order.LoyaltyPointsEarned > 0 {
		if err := s.customerService.AddLoyaltyPoints(uint(*order.CustomerID), -order.LoyaltyPointsEarned); err != nil {
			// Log error but don't fail the void operation
			// In production, use structured logging
		}
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
