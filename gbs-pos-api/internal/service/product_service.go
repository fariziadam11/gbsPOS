package service

import (
	"gbs-pos-api/internal/model"
	"gbs-pos-api/internal/repository"
)

type ProductService struct {
	repo *repository.ProductRepository
}

func NewProductService(repo *repository.ProductRepository) *ProductService {
	return &ProductService{repo: repo}
}

func (s *ProductService) List(storeType, category string, lastSync int64) ([]model.Product, error) {
	return s.repo.FindAll(storeType, category, lastSync)
}

func (s *ProductService) Get(id uint) (*model.Product, error) {
	return s.repo.FindByID(id)
}

func (s *ProductService) Create(product *model.Product) error {
	return s.repo.Create(product)
}

func (s *ProductService) Update(id uint, updates *model.Product) (*model.Product, error) {
	product, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if updates.Name != "" {
		product.Name = updates.Name
	}
	if updates.Price >= 0 {
		product.Price = updates.Price
	}
	if updates.Category != "" {
		product.Category = updates.Category
	}
	product.ImageURL = updates.ImageURL
	if updates.StoreType != "" {
		product.StoreType = updates.StoreType
	}
	if err := s.repo.Update(product); err != nil {
		return nil, err
	}
	return product, nil
}

func (s *ProductService) Delete(id uint) error {
	return s.repo.Delete(id)
}
