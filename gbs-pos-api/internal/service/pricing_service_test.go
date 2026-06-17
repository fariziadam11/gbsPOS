package service

import (
	"testing"
	"time"

	"gbs-pos-api/internal/database"
	"gbs-pos-api/internal/dto"
	"gbs-pos-api/internal/model"
	"gbs-pos-api/internal/repository"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupPricingTest(t *testing.T) (*PricingService, *DiscountService) {
	t.Helper()

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
	discountRepo := repository.NewDiscountRepository(db)
	discountSvc := NewDiscountService(discountRepo, productRepo)
	pricingSvc := NewPricingService(productRepo, discountSvc)

	return pricingSvc, discountSvc
}

func TestPricingService_CalculateAppliesProductDiscount(t *testing.T) {
	pricingSvc, discountSvc := setupPricingTest(t)
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

	result, err := pricingSvc.Calculate(dto.PricingCalculationRequest{
		Items: []dto.PricingItemInput{{ProductID: 1, Qty: 1}},
	})

	require.NoError(t, err)
	require.Len(t, result.Items, 1)
	assert.Equal(t, 10000.0, result.Items[0].OriginalPrice)
	assert.Equal(t, 1000.0, result.Items[0].DiscountAmount)
	assert.Equal(t, 9000.0, result.Items[0].FinalPrice)
	assert.Equal(t, 9000.0, result.FinalTotal)
}

func TestPricingService_CalculateSelectsBiggestTransactionDiscount(t *testing.T) {
	pricingSvc, discountSvc := setupPricingTest(t)
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

	result, err := pricingSvc.Calculate(dto.PricingCalculationRequest{
		Items: []dto.PricingItemInput{{ProductID: 1, Qty: 10}},
	})

	require.NoError(t, err)
	require.NotNil(t, result.TransactionDiscount)
	assert.Equal(t, "Twenty Percent", result.TransactionDiscount.Name)
	assert.Equal(t, 20000.0, result.TransactionDiscount.Amount)
	assert.Equal(t, 80000.0, result.FinalTotal)
}

func TestPricingService_CalculateSelectsHighestVoucherDiscount(t *testing.T) {
	pricingSvc, discountSvc := setupPricingTest(t)
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

	result, err := pricingSvc.Calculate(dto.PricingCalculationRequest{
		Items:       []dto.PricingItemInput{{ProductID: 1, Qty: 10}},
		VoucherCode: &code,
	})

	require.NoError(t, err)
	require.NotNil(t, result.VoucherDiscount)
	assert.Equal(t, "Welcome Fixed", result.VoucherDiscount.Name)
	assert.Equal(t, 15000.0, result.VoucherDiscount.Amount)
	assert.Equal(t, 85000.0, result.FinalTotal)
}

func TestPricingService_CalculateChainsProductTransactionVoucherDiscounts(t *testing.T) {
	pricingSvc, discountSvc := setupPricingTest(t)
	now := time.Date(2026, 7, 15, 12, 0, 0, 0, time.UTC)
	discountSvc.now = func() time.Time { return now }
	code := "WELCOME"

	_, err := discountSvc.Create(dto.CreateDiscountRequest{
		ProductID: 1,
		Name:      "Product Ten Percent",
		Type:      DiscountTypePercentage,
		Value:     10,
		StartDate: "2026-07-01",
		EndDate:   "2026-07-31",
	})
	require.NoError(t, err)
	_, err = discountSvc.Create(dto.CreateDiscountRequest{
		Scope:     model.DiscountScopeTransaction,
		Name:      "Transaction Five Percent",
		Type:      DiscountTypePercentage,
		Value:     5,
		StartDate: "2026-07-01",
		EndDate:   "2026-07-31",
	})
	require.NoError(t, err)
	_, err = discountSvc.Create(dto.CreateDiscountRequest{
		Scope:       model.DiscountScopeVoucher,
		VoucherCode: &code,
		Name:        "Voucher Fixed",
		Type:        DiscountTypeFixed,
		Value:       1000,
		StartDate:   "2026-07-01",
		EndDate:     "2026-07-31",
	})
	require.NoError(t, err)

	result, err := pricingSvc.Calculate(dto.PricingCalculationRequest{
		Items:       []dto.PricingItemInput{{ProductID: 1, Qty: 1}},
		VoucherCode: &code,
	})

	require.NoError(t, err)
	require.Len(t, result.Items, 1)
	require.NotNil(t, result.TransactionDiscount)
	require.NotNil(t, result.VoucherDiscount)
	assert.Equal(t, 10000.0, result.Subtotal)
	assert.Equal(t, 1000.0, result.ProductDiscountTotal)
	assert.Equal(t, 9000.0, result.AfterProductDiscountTotal)
	assert.Equal(t, 450.0, result.TransactionDiscount.Amount)
	assert.Equal(t, 1000.0, result.VoucherDiscount.Amount)
	assert.Equal(t, 2450.0, result.TotalDiscountAmount)
	assert.Equal(t, 7550.0, result.FinalTotal)
}
