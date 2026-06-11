package repository

import (
	"gbs-pos-api/internal/model"
	"gorm.io/gorm"
)

type ProductVariantRepository struct {
	db *gorm.DB
}

func NewProductVariantRepository(db *gorm.DB) *ProductVariantRepository {
	return &ProductVariantRepository{db: db}
}

func (r *ProductVariantRepository) FindByProductID(productID int) ([]model.ProductVariant, error) {
	var variants []model.ProductVariant
	if err := r.db.Where("product_id = ?", productID).Order("sort_order ASC").Find(&variants).Error; err != nil {
		return nil, err
	}
	return variants, nil
}

func (r *ProductVariantRepository) FindByID(id int) (*model.ProductVariant, error) {
	var variant model.ProductVariant
	if err := r.db.First(&variant, id).Error; err != nil {
		return nil, err
	}
	return &variant, nil
}

func (r *ProductVariantRepository) Create(variant *model.ProductVariant) error {
	return r.db.Create(variant).Error
}

func (r *ProductVariantRepository) Update(variant *model.ProductVariant) error {
	return r.db.Save(variant).Error
}

func (r *ProductVariantRepository) Delete(id int) error {
	return r.db.Delete(&model.ProductVariant{}, id).Error
}

func (r *ProductVariantRepository) HasVariants(productID int) bool {
	var count int64
	r.db.Model(&model.ProductVariant{}).Where("product_id = ? AND is_active = ?", productID, true).Count(&count)
	return count > 0
}
