package service

import (
	"errors"
	"testing"
	"time"

	"gbs-pos-api/internal/database"
	"gbs-pos-api/internal/dto"
	"gbs-pos-api/internal/model"
	"gbs-pos-api/internal/repository"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func setupDiscountTest(t *testing.T) (*DiscountService, *ProductService, *gorm.DB) {
	db, err := database.NewTestDB()
	require.NoError(t, err)

	product := &model.Product{
		Name:              "Chitato",
		Price:             10000,
		Category:          "Snacks",
		StoreType:         "RETAIL",
		StockQuantity:     100,
		LowStockThreshold: 10,
	}
	require.NoError(t, db.Create(product).Error)

	productRepo := repository.NewProductRepository(db)
	movementRepo := repository.NewStockMovementRepository(db)
	discountRepo := repository.NewDiscountRepository(db)

	productSvc := NewProductService(productRepo, movementRepo)
	discountSvc := NewDiscountService(discountRepo, productRepo)
	productSvc.SetDiscountService(discountSvc)

	return discountSvc, productSvc, db
}

func TestDiscountService_GetEffectiveStatus(t *testing.T) {
	discountSvc, _, _ := setupDiscountTest(t)
	now := time.Date(2026, 7, 1, 12, 0, 0, 0, time.UTC)
	discountSvc.now = func() time.Time { return now }

	tests := []struct {
		name     string
		discount model.Discount
		expected string
	}{
		{
			name: "pending",
			discount: model.Discount{
				StartDate: now.Add(time.Hour),
				EndDate:   now.Add(2 * time.Hour),
				Status:    DiscountStatusPending,
			},
			expected: DiscountStatusPending,
		},
		{
			name: "active",
			discount: model.Discount{
				StartDate: now.Add(-time.Hour),
				EndDate:   now.Add(time.Hour),
				Status:    DiscountStatusPending,
			},
			expected: DiscountStatusActive,
		},
		{
			name: "expired",
			discount: model.Discount{
				StartDate: now.Add(-2 * time.Hour),
				EndDate:   now.Add(-time.Hour),
				Status:    DiscountStatusActive,
			},
			expected: DiscountStatusExpired,
		},
		{
			name: "stopped terminal",
			discount: model.Discount{
				StartDate: now.Add(-time.Hour),
				EndDate:   now.Add(time.Hour),
				Status:    DiscountStatusStopped,
			},
			expected: DiscountStatusStopped,
		},
		{
			name: "cancelled terminal",
			discount: model.Discount{
				StartDate: now.Add(time.Hour),
				EndDate:   now.Add(2 * time.Hour),
				Status:    DiscountStatusCancelled,
			},
			expected: DiscountStatusCancelled,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, discountSvc.GetEffectiveStatus(&tt.discount))
		})
	}
}

func TestDiscountService_CreateValidation(t *testing.T) {
	discountSvc, _, _ := setupDiscountTest(t)

	tests := []struct {
		name string
		req  dto.CreateDiscountRequest
	}{
		{
			name: "percentage too high",
			req: dto.CreateDiscountRequest{
				ProductID: 1,
				Name:      "Bad Percentage",
				Type:      DiscountTypePercentage,
				Value:     101,
				StartDate: "2026-07-01",
				EndDate:   "2026-07-31",
			},
		},
		{
			name: "fixed zero",
			req: dto.CreateDiscountRequest{
				ProductID: 1,
				Name:      "Bad Fixed",
				Type:      DiscountTypeFixed,
				Value:     0,
				StartDate: "2026-07-01",
				EndDate:   "2026-07-31",
			},
		},
		{
			name: "invalid dates",
			req: dto.CreateDiscountRequest{
				ProductID: 1,
				Name:      "Bad Dates",
				Type:      DiscountTypeFixed,
				Value:     5000,
				StartDate: "2026-08-01",
				EndDate:   "2026-07-31",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := discountSvc.Create(tt.req)
			var validationErr *DiscountValidationError
			require.Error(t, err)
			assert.True(t, errors.As(err, &validationErr))
		})
	}
}

func TestDiscountService_CreateRejectsOverlappingPendingOrActiveDiscount(t *testing.T) {
	discountSvc, _, _ := setupDiscountTest(t)
	now := time.Date(2026, 6, 15, 12, 0, 0, 0, time.UTC)
	discountSvc.now = func() time.Time { return now }

	_, err := discountSvc.Create(dto.CreateDiscountRequest{
		ProductID: 1,
		Name:      "July Promo",
		Type:      DiscountTypePercentage,
		Value:     10,
		StartDate: "2026-07-01",
		EndDate:   "2026-07-31",
	})
	require.NoError(t, err)

	_, err = discountSvc.Create(dto.CreateDiscountRequest{
		ProductID: 1,
		Name:      "Overlap Promo",
		Type:      DiscountTypeFixed,
		Value:     5000,
		StartDate: "2026-07-15",
		EndDate:   "2026-08-15",
	})

	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrDiscountOverlap))
}

func TestDiscountService_StopAndCancel(t *testing.T) {
	discountSvc, _, _ := setupDiscountTest(t)
	now := time.Date(2026, 7, 15, 12, 0, 0, 0, time.UTC)
	discountSvc.now = func() time.Time { return now }

	active, err := discountSvc.Create(dto.CreateDiscountRequest{
		ProductID: 1,
		Name:      "Active Promo",
		Type:      DiscountTypePercentage,
		Value:     10,
		StartDate: "2026-07-01",
		EndDate:   "2026-07-31",
	})
	require.NoError(t, err)

	stopped, err := discountSvc.Stop(active.ID)
	require.NoError(t, err)
	assert.Equal(t, DiscountStatusStopped, stopped.EffectiveStatus)

	pending, err := discountSvc.Create(dto.CreateDiscountRequest{
		ProductID: 1,
		Name:      "August Promo",
		Type:      DiscountTypeFixed,
		Value:     2000,
		StartDate: "2026-08-01",
		EndDate:   "2026-08-31",
	})
	require.NoError(t, err)

	cancelled, err := discountSvc.Cancel(pending.ID)
	require.NoError(t, err)
	assert.Equal(t, DiscountStatusCancelled, cancelled.EffectiveStatus)
}

func TestProductService_ListWithDiscountsAppliesActiveDiscountOnly(t *testing.T) {
	discountSvc, productSvc, _ := setupDiscountTest(t)
	now := time.Date(2026, 7, 15, 12, 0, 0, 0, time.UTC)
	discountSvc.now = func() time.Time { return now }

	_, err := discountSvc.Create(dto.CreateDiscountRequest{
		ProductID: 1,
		Name:      "Weekend Promo",
		Type:      DiscountTypePercentage,
		Value:     10,
		StartDate: "2026-07-01",
		EndDate:   "2026-07-31",
	})
	require.NoError(t, err)

	products, err := productSvc.ListWithDiscounts("", "", 0)
	require.NoError(t, err)
	require.Len(t, products, 1)
	require.NotNil(t, products[0].Discount)
	assert.Equal(t, 9000.0, products[0].FinalPrice)
	assert.Equal(t, DiscountStatusActive, products[0].Discount.Status)
}

func TestProductService_ListWithDiscountsClampsFixedDiscountAtZero(t *testing.T) {
	discountSvc, productSvc, _ := setupDiscountTest(t)
	now := time.Date(2026, 7, 15, 12, 0, 0, 0, time.UTC)
	discountSvc.now = func() time.Time { return now }

	_, err := discountSvc.Create(dto.CreateDiscountRequest{
		ProductID: 1,
		Name:      "Big Promo",
		Type:      DiscountTypeFixed,
		Value:     50000,
		StartDate: "2026-07-01",
		EndDate:   "2026-07-31",
	})
	require.NoError(t, err)

	products, err := productSvc.ListWithDiscounts("", "", 0)
	require.NoError(t, err)
	require.Len(t, products, 1)
	assert.Equal(t, 0.0, products[0].FinalPrice)
}
