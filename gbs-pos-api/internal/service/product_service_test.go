package service

import (
	"testing"

	"gbs-pos-api/internal/database"
	"gbs-pos-api/internal/model"
	"gbs-pos-api/internal/repository"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func setupProductTest(t *testing.T) (*ProductService, *gorm.DB) {
	db, err := database.NewTestDB()
	require.NoError(t, err)

	db.Create(
		&model.Product{Name: "Chitato", Price: 11500, Category: "Snacks", StoreType: "RETAIL", StockQuantity: 100, LowStockThreshold: 10},
	)
	db.Create(&model.Product{Name: "Nasi Goreng", Price: 25000, Category: "Food", StoreType: "FNB", StockQuantity: 50, LowStockThreshold: 5})

	repo := repository.NewProductRepository(db)
	movementRepo := repository.NewStockMovementRepository(db)
	return NewProductService(repo, movementRepo), db
}

func TestProductService_List(t *testing.T) {
	svc, _ := setupProductTest(t)

	products, err := svc.List("", "", 0)
	require.NoError(t, err)
	assert.Len(t, products, 2)
}

func TestProductService_List_FilterStoreType(t *testing.T) {
	svc, _ := setupProductTest(t)

	products, err := svc.List("RETAIL", "", 0)
	require.NoError(t, err)
	assert.Len(t, products, 1)
	assert.Equal(t, "Chitato", products[0].Name)
}

func TestProductService_Get(t *testing.T) {
	svc, _ := setupProductTest(t)

	product, err := svc.Get(1)
	require.NoError(t, err)
	assert.Equal(t, "Chitato", product.Name)
}

func TestProductService_Get_NotFound(t *testing.T) {
	svc, _ := setupProductTest(t)

	_, err := svc.Get(999)
	assert.Error(t, err)
}

func TestProductService_Create(t *testing.T) {
	svc, _ := setupProductTest(t)

	product := &model.Product{
		Name:      "Teh Botol",
		Price:     5000,
		Category:  "Beverages",
		StoreType: "RETAIL",
	}
	err := svc.Create(product)
	require.NoError(t, err)
	assert.NotZero(t, product.ID)
}

func TestProductService_Update(t *testing.T) {
	svc, _ := setupProductTest(t)

	updated, err := svc.Update(
		1,
		&model.Product{Name: "Chitato BBQ", Price: 12000, Category: "Snacks", StoreType: "RETAIL"},
	)
	require.NoError(t, err)
	assert.Equal(t, "Chitato BBQ", updated.Name)
	assert.Equal(t, 12000.0, updated.Price)
}

func TestProductService_Delete(t *testing.T) {
	svc, _ := setupProductTest(t)

	err := svc.Delete(1)
	require.NoError(t, err)

	_, err = svc.Get(1)
	assert.Error(t, err)
}
