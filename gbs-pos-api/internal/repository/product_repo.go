package repository

import (
	"errors"
	"gbs-pos-api/internal/model"
	"time"

	"gorm.io/gorm"
)

type ProductRepository struct {
	db *gorm.DB
}

func NewProductRepository(db *gorm.DB) *ProductRepository {
	return &ProductRepository{db: db}
}

func (r *ProductRepository) FindAll(
	storeType, category string,
	lastSync int64,
) ([]model.Product, error) {
	var products []model.Product
	query := r.db.Order("id ASC")
	if storeType != "" {
		query = query.Where("store_type = ?", storeType)
	}
	if category != "" {
		query = query.Where("category = ?", category)
	}
	if lastSync > 0 {
		syncTime := time.UnixMilli(lastSync)
		query = query.Where("updated_at > ?", syncTime)
	}
	if err := query.Find(&products).Error; err != nil {
		return nil, err
	}
	return products, nil
}

func (r *ProductRepository) FindByID(id uint) (*model.Product, error) {
	var product model.Product
	if err := r.db.First(&product, id).Error; err != nil {
		return nil, err
	}
	return &product, nil
}

func (r *ProductRepository) Create(product *model.Product) error {
	return r.db.Create(product).Error
}

func (r *ProductRepository) Update(product *model.Product) error {
	return r.db.Save(product).Error
}

func (r *ProductRepository) Delete(id uint) error {
	return r.db.Delete(&model.Product{}, id).Error
}

func (r *ProductRepository) FindAllByStoreType(storeType string) ([]model.Product, error) {
	return r.FindAll(storeType, "", 0)
}

func (r *ProductRepository) GetLowStock(storeType string) ([]model.Product, error) {
	var products []model.Product
	query := r.db.Where("stock_quantity <= low_stock_threshold AND low_stock_threshold > 0")
	if storeType != "" {
		query = query.Where("store_type = ?", storeType)
	}
	if err := query.Find(&products).Error; err != nil {
		return nil, err
	}
	return products, nil
}

func (r *ProductRepository) Upsert(product *model.Product) error {
	var existing model.Product
	err := r.db.Where("name = ? AND store_type = ?", product.Name, product.StoreType).First(&existing).Error
	if err == nil {
		existing.Price = product.Price
		existing.Category = product.Category
		existing.ImageURL = product.ImageURL
		existing.StockQuantity = product.StockQuantity
		existing.LowStockThreshold = product.LowStockThreshold
		return r.db.Save(&existing).Error
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return r.db.Create(product).Error
	}
	return err
}
