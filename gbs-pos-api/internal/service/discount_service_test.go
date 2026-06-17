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

func TestDiscountService_ProductDiscountStillWorks(t *testing.T) {
	discountSvc, _, _ := setupDiscountTest(t)
	now := time.Date(2026, 7, 15, 12, 0, 0, 0, time.UTC)
	discountSvc.now = func() time.Time { return now }

	created, err := discountSvc.Create(dto.CreateDiscountRequest{
		ProductID: 1,
		Name:      "Product Promo",
		Type:      DiscountTypePercentage,
		Value:     10,
		StartDate: "2026-07-01",
		EndDate:   "2026-07-31",
	})

	require.NoError(t, err)
	require.NotNil(t, created.ProductID)
	assert.Equal(t, uint(1), *created.ProductID)
	assert.Equal(t, model.DiscountScopeProduct, created.Scope)

	activeDiscounts, err := discountSvc.GetActiveDiscountsByProductIDs([]uint{1})
	require.NoError(t, err)
	require.Contains(t, activeDiscounts, uint(1))
	assert.Equal(t, 9000.0, discountSvc.ApplyDiscount(10000, activeDiscounts[1]))
}

func TestDiscountService_ProductDiscountRejectsDuplicateActiveDiscount(t *testing.T) {
	discountSvc, _, _ := setupDiscountTest(t)
	now := time.Date(2026, 7, 15, 12, 0, 0, 0, time.UTC)
	discountSvc.now = func() time.Time { return now }

	_, err := discountSvc.Create(dto.CreateDiscountRequest{
		ProductID: 1,
		Name:      "Product Promo",
		Type:      DiscountTypePercentage,
		Value:     10,
		StartDate: "2026-07-01",
		EndDate:   "2026-07-31",
	})
	require.NoError(t, err)

	_, err = discountSvc.Create(dto.CreateDiscountRequest{
		ProductID: 1,
		Name:      "Duplicate Promo",
		Type:      DiscountTypeFixed,
		Value:     1000,
		StartDate: "2026-07-01",
		EndDate:   "2026-07-31",
	})

	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrDiscountOverlap))
}

func TestDiscountService_ApplyTransactionPercentageDiscount(t *testing.T) {
	discountSvc, _, _ := setupDiscountTest(t)
	now := time.Date(2026, 7, 15, 12, 0, 0, 0, time.UTC)
	discountSvc.now = func() time.Time { return now }

	_, err := discountSvc.Create(dto.CreateDiscountRequest{
		Scope:     model.DiscountScopeTransaction,
		Name:      "Weekend Promo",
		Type:      DiscountTypePercentage,
		Value:     10,
		StartDate: "2026-07-01",
		EndDate:   "2026-07-31",
	})
	require.NoError(t, err)

	finalTotal, discount, err := discountSvc.ApplyTransactionDiscount(100000)

	require.NoError(t, err)
	require.NotNil(t, discount)
	assert.Equal(t, 90000.0, finalTotal)
	assert.Equal(t, model.DiscountScopeTransaction, discount.Scope)
}

func TestDiscountService_TransactionAllowsMultipleActiveAndSelectsBiggestDiscount(t *testing.T) {
	discountSvc, _, _ := setupDiscountTest(t)
	now := time.Date(2026, 7, 15, 12, 0, 0, 0, time.UTC)
	discountSvc.now = func() time.Time { return now }

	_, err := discountSvc.Create(dto.CreateDiscountRequest{
		Scope:     model.DiscountScopeTransaction,
		Name:      "Ten Percent",
		Type:      DiscountTypePercentage,
		Value:     10,
		StartDate: "2026-07-01",
		EndDate:   "2026-07-31",
	})
	require.NoError(t, err)
	_, err = discountSvc.Create(dto.CreateDiscountRequest{
		Scope:     model.DiscountScopeTransaction,
		Name:      "Twenty Percent",
		Type:      DiscountTypePercentage,
		Value:     20,
		StartDate: "2026-07-01",
		EndDate:   "2026-07-31",
	})
	require.NoError(t, err)

	finalTotal, discount, err := discountSvc.ApplyTransactionDiscount(100000)

	require.NoError(t, err)
	require.NotNil(t, discount)
	assert.Equal(t, "Twenty Percent", discount.Name)
	assert.Equal(t, 80000.0, finalTotal)
}

func TestDiscountService_ApplyTransactionFixedDiscount(t *testing.T) {
	discountSvc, _, _ := setupDiscountTest(t)
	now := time.Date(2026, 7, 15, 12, 0, 0, 0, time.UTC)
	discountSvc.now = func() time.Time { return now }

	_, err := discountSvc.Create(dto.CreateDiscountRequest{
		Scope:     model.DiscountScopeTransaction,
		Name:      "Fixed Transaction Promo",
		Type:      DiscountTypeFixed,
		Value:     15000,
		StartDate: "2026-07-01",
		EndDate:   "2026-07-31",
	})
	require.NoError(t, err)

	finalTotal, discount, err := discountSvc.ApplyTransactionDiscount(100000)

	require.NoError(t, err)
	require.NotNil(t, discount)
	assert.Equal(t, 85000.0, finalTotal)
}

func TestDiscountService_ApplyTransactionFixedDiscountClampsAtZero(t *testing.T) {
	discountSvc, _, _ := setupDiscountTest(t)
	now := time.Date(2026, 7, 15, 12, 0, 0, 0, time.UTC)
	discountSvc.now = func() time.Time { return now }

	_, err := discountSvc.Create(dto.CreateDiscountRequest{
		Scope:     model.DiscountScopeTransaction,
		Name:      "Large Transaction Promo",
		Type:      DiscountTypeFixed,
		Value:     15000,
		StartDate: "2026-07-01",
		EndDate:   "2026-07-31",
	})
	require.NoError(t, err)

	finalTotal, discount, err := discountSvc.ApplyTransactionDiscount(10000)

	require.NoError(t, err)
	require.NotNil(t, discount)
	assert.Equal(t, 0.0, finalTotal)
}

func TestDiscountService_ValidateVoucherValid(t *testing.T) {
	discountSvc, _, _ := setupDiscountTest(t)
	now := time.Date(2026, 7, 15, 12, 0, 0, 0, time.UTC)
	discountSvc.now = func() time.Time { return now }
	code := " welcome50 "
	minTransaction := 50000.0

	_, err := discountSvc.Create(dto.CreateDiscountRequest{
		Scope:          model.DiscountScopeVoucher,
		VoucherCode:    &code,
		MinTransaction: &minTransaction,
		Name:           "Welcome Voucher",
		Type:           DiscountTypeFixed,
		Value:          5000,
		StartDate:      "2026-07-01",
		EndDate:        "2026-07-31",
	})
	require.NoError(t, err)

	discount, err := discountSvc.ValidateVoucher("welcome50", 60000)

	require.NoError(t, err)
	require.NotNil(t, discount)
	assert.Equal(t, model.DiscountScopeVoucher, discount.Scope)
	assert.Equal(t, "WELCOME50", *discount.VoucherCode)
}

func TestDiscountService_VoucherSelectsHighestDiscountForCode(t *testing.T) {
	discountSvc, _, _ := setupDiscountTest(t)
	now := time.Date(2026, 7, 15, 12, 0, 0, 0, time.UTC)
	discountSvc.now = func() time.Time { return now }
	code := "WELCOME"

	_, err := discountSvc.Create(dto.CreateDiscountRequest{
		Scope:       model.DiscountScopeVoucher,
		VoucherCode: &code,
		Name:        "Welcome Ten Percent",
		Type:        DiscountTypePercentage,
		Value:       10,
		StartDate:   "2026-07-01",
		EndDate:     "2026-07-31",
	})
	require.NoError(t, err)
	_, err = discountSvc.Create(dto.CreateDiscountRequest{
		Scope:       model.DiscountScopeVoucher,
		VoucherCode: &code,
		Name:        "Welcome Fixed",
		Type:        DiscountTypeFixed,
		Value:       15000,
		StartDate:   "2026-07-01",
		EndDate:     "2026-07-31",
	})
	require.NoError(t, err)

	discount, amount, err := discountSvc.SelectBestVoucherDiscount("welcome", 100000)

	require.NoError(t, err)
	require.NotNil(t, discount)
	assert.Equal(t, "Welcome Fixed", discount.Name)
	assert.Equal(t, 15000.0, amount)
}

func TestDiscountService_ValidateVoucherRejectsMinimumTransaction(t *testing.T) {
	discountSvc, _, _ := setupDiscountTest(t)
	now := time.Date(2026, 7, 15, 12, 0, 0, 0, time.UTC)
	discountSvc.now = func() time.Time { return now }
	code := "WELCOME50"
	minTransaction := 50000.0

	_, err := discountSvc.Create(dto.CreateDiscountRequest{
		Scope:          model.DiscountScopeVoucher,
		VoucherCode:    &code,
		MinTransaction: &minTransaction,
		Name:           "Welcome Voucher",
		Type:           DiscountTypeFixed,
		Value:          5000,
		StartDate:      "2026-07-01",
		EndDate:        "2026-07-31",
	})
	require.NoError(t, err)

	_, err = discountSvc.ValidateVoucher("WELCOME50", 40000)

	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrVoucherMinimumNotMet))
}

func TestDiscountService_ValidateVoucherRejectsExpired(t *testing.T) {
	discountSvc, _, _ := setupDiscountTest(t)
	now := time.Date(2026, 8, 1, 12, 0, 0, 0, time.UTC)
	discountSvc.now = func() time.Time { return now }
	code := "WELCOME50"

	_, err := discountSvc.Create(dto.CreateDiscountRequest{
		Scope:       model.DiscountScopeVoucher,
		VoucherCode: &code,
		Name:        "Welcome Voucher",
		Type:        DiscountTypeFixed,
		Value:       5000,
		StartDate:   "2026-07-01",
		EndDate:     "2026-07-31",
	})
	require.NoError(t, err)

	_, err = discountSvc.ValidateVoucher("WELCOME50", 60000)

	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrVoucherInvalid))
}
