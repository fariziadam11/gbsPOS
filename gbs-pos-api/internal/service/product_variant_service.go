package service

import (
	"fmt"
	"gbs-pos-api/internal/dto"
	"gbs-pos-api/internal/model"
	"gbs-pos-api/internal/repository"
	"time"

	"gorm.io/gorm"
)

type ProductVariantService struct {
	repo *repository.ProductVariantRepository
}

func NewProductVariantService(repo *repository.ProductVariantRepository) *ProductVariantService {
	return &ProductVariantService{repo: repo}
}

func (s *ProductVariantService) ListByProduct(productID int) ([]dto.VariantResponse, error) {
	variants, err := s.repo.FindByProductID(productID)
	if err != nil {
		return nil, err
	}
	result := make([]dto.VariantResponse, len(variants))
	for i, v := range variants {
		result[i] = toVariantResponse(v)
	}
	return result, nil
}

func (s *ProductVariantService) Get(id int) (*dto.VariantResponse, error) {
	v, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}
	resp := toVariantResponse(*v)
	return &resp, nil
}

func (s *ProductVariantService) Create(productID int, req dto.CreateVariantRequest) (*dto.VariantResponse, error) {
	if req.Attributes == nil {
		req.Attributes = make(map[string]interface{})
	}
	variant := &model.ProductVariant{
		ProductID:         productID,
		SKU:               req.SKU,
		Name:              req.Name,
		Attributes:        req.Attributes,
		Price:             req.Price,
		StockQuantity:     req.StockQuantity,
		LowStockThreshold: req.LowStockThreshold,
		SortOrder:         req.SortOrder,
		IsActive:          true,
	}
	if req.IsActive != nil {
		variant.IsActive = *req.IsActive
	}
	if err := s.repo.Create(variant); err != nil {
		return nil, err
	}
	return s.Get(int(variant.ID))
}

func (s *ProductVariantService) Update(id int, req dto.UpdateVariantRequest) (*dto.VariantResponse, error) {
	v, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if req.SKU != nil {
		v.SKU = *req.SKU
	}
	if req.Name != nil {
		v.Name = *req.Name
	}
	if req.Attributes != nil {
		v.Attributes = req.Attributes
	}
	if req.Price != nil {
		v.Price = req.Price
	}
	if req.StockQuantity != nil {
		v.StockQuantity = *req.StockQuantity
	}
	if req.LowStockThreshold != nil {
		v.LowStockThreshold = req.LowStockThreshold
	}
	if req.IsActive != nil {
		v.IsActive = *req.IsActive
	}
	if req.SortOrder != nil {
		v.SortOrder = *req.SortOrder
	}
	if err := s.repo.Update(v); err != nil {
		return nil, err
	}
	return s.Get(int(v.ID))
}

func (s *ProductVariantService) Delete(id int) error {
	return s.repo.Delete(id)
}

func (s *ProductVariantService) DeductVariantStock(tx *gorm.DB, variantID int, quantity int, orderID string) error {
	var variant model.ProductVariant
	if err := tx.First(&variant, variantID).Error; err != nil {
		return err
	}
	if variant.StockQuantity < quantity {
		return fmt.Errorf("insufficient stock for variant: %s (have %d, need %d)", variant.Name, variant.StockQuantity, quantity)
	}
	variant.StockQuantity -= quantity
	return tx.Save(&variant).Error
}

func (s *ProductVariantService) RestoreVariantStock(tx *gorm.DB, variantID int, quantity int, orderID string) error {
	var variant model.ProductVariant
	if err := tx.First(&variant, variantID).Error; err != nil {
		return err
	}
	variant.StockQuantity += quantity
	return tx.Save(&variant).Error
}

func toVariantResponse(v model.ProductVariant) dto.VariantResponse {
	return dto.VariantResponse{
		ID:                v.ID,
		ProductID:         v.ProductID,
		SKU:               v.SKU,
		Name:              v.Name,
		Attributes:        v.Attributes,
		Price:             v.Price,
		StockQuantity:     v.StockQuantity,
		LowStockThreshold: v.LowStockThreshold,
		IsActive:          v.IsActive,
		SortOrder:         v.SortOrder,
		CreatedAt:         v.CreatedAt.Format(time.RFC3339),
		UpdatedAt:         v.UpdatedAt.Format(time.RFC3339),
	}
}
