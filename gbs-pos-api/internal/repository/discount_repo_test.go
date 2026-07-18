package repository

import (
	"testing"
	"time"

	"gbs-pos-api/internal/database"
	"gbs-pos-api/internal/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func setupDiscountRepoTest(t *testing.T) (*DiscountRepository, *gorm.DB) {
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

	return NewDiscountRepository(db), db
}

func TestDiscountRepository_FindByProductID_ProductDiscountRegression(t *testing.T) {
	repo, db := setupDiscountRepoTest(t)
	now := time.Date(2026, 7, 15, 12, 0, 0, 0, time.UTC)
	productID := uint(1)

	require.NoError(t, db.Create(&model.Discount{
		ProductID: &productID,
		Scope:     model.DiscountScopeProduct,
		Name:      "Product Promo",
		Type:      "PERCENTAGE",
		Value:     10,
		StartDate: now.Add(-time.Hour),
		EndDate:   now.Add(time.Hour),
		Status:    "ACTIVE",
	}).Error)
	require.NoError(t, db.Create(&model.Discount{
		Scope:     model.DiscountScopeTransaction,
		Name:      "Transaction Promo",
		Type:      "PERCENTAGE",
		Value:     5,
		StartDate: now.Add(-time.Hour),
		EndDate:   now.Add(time.Hour),
		Status:    "ACTIVE",
	}).Error)

	discounts, err := repo.FindByProductID(productID)

	require.NoError(t, err)
	require.Len(t, discounts, 1)
	assert.Equal(t, model.DiscountScopeProduct, discounts[0].Scope)
	assert.Equal(t, "Product Promo", discounts[0].Name)
}

func TestDiscountRepository_FindActiveByScope(t *testing.T) {
	repo, db := setupDiscountRepoTest(t)
	now := time.Date(2026, 7, 15, 12, 0, 0, 0, time.UTC)

	require.NoError(t, db.Create(&model.Discount{
		Scope:     model.DiscountScopeTransaction,
		Name:      "Active Transaction Promo",
		Type:      "PERCENTAGE",
		Value:     10,
		StartDate: now.Add(-time.Hour),
		EndDate:   now.Add(time.Hour),
		Status:    "ACTIVE",
	}).Error)
	require.NoError(t, db.Create(&model.Discount{
		Scope:     model.DiscountScopeTransaction,
		Name:      "Expired Transaction Promo",
		Type:      "PERCENTAGE",
		Value:     20,
		StartDate: now.Add(-2 * time.Hour),
		EndDate:   now.Add(-time.Hour),
		Status:    "ACTIVE",
	}).Error)

	discounts, err := repo.FindActiveByScope(model.DiscountScopeTransaction, now)

	require.NoError(t, err)
	require.Len(t, discounts, 1)
	assert.Equal(t, "Active Transaction Promo", discounts[0].Name)
}

func TestDiscountRepository_FindVoucherDiscountsByCodeReturnsAllMatchingCode(t *testing.T) {
	repo, db := setupDiscountRepoTest(t)
	now := time.Date(2026, 7, 15, 12, 0, 0, 0, time.UTC)
	code := "WELCOME50"
	otherCode := "OTHER"

	require.NoError(t, db.Create(&model.Discount{
		Scope:       model.DiscountScopeVoucher,
		VoucherCode: &code,
		Name:        "Welcome Percentage",
		Type:        "PERCENTAGE",
		Value:       10,
		StartDate:   now.Add(-time.Hour),
		EndDate:     now.Add(time.Hour),
		Status:      "ACTIVE",
	}).Error)
	require.NoError(t, db.Create(&model.Discount{
		Scope:       model.DiscountScopeVoucher,
		VoucherCode: &code,
		Name:        "Welcome Fixed",
		Type:        "FIXED",
		Value:       15000,
		StartDate:   now.Add(-time.Hour),
		EndDate:     now.Add(time.Hour),
		Status:      "ACTIVE",
	}).Error)
	require.NoError(t, db.Create(&model.Discount{
		Scope:       model.DiscountScopeVoucher,
		VoucherCode: &otherCode,
		Name:        "Other Voucher",
		Type:        "FIXED",
		Value:       5000,
		StartDate:   now.Add(-time.Hour),
		EndDate:     now.Add(time.Hour),
		Status:      "ACTIVE",
	}).Error)

	discounts, err := repo.FindVoucherDiscountsByCode(code)

	require.NoError(t, err)
	require.Len(t, discounts, 2)
	for i := range discounts {
		assert.Equal(t, code, *discounts[i].VoucherCode)
	}
}
