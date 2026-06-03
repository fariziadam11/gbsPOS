package service

import (
	"encoding/csv"
	"errors"
	"fmt"
	"gbs-pos-api/internal/dto"
	"gbs-pos-api/internal/model"
	"gbs-pos-api/internal/repository"
	"io"
	"strconv"
	"strings"

	"gorm.io/gorm"
)

type ProductService struct {
	repo         *repository.ProductRepository
	movementRepo *repository.StockMovementRepository
}

func NewProductService(repo *repository.ProductRepository, movementRepo *repository.StockMovementRepository) *ProductService {
	return &ProductService{repo: repo, movementRepo: movementRepo}
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
	if updates.LowStockThreshold > 0 {
		product.LowStockThreshold = updates.LowStockThreshold
	}
	if err := s.repo.Update(product); err != nil {
		return nil, err
	}
	return product, nil
}

func (s *ProductService) Delete(id uint) error {
	return s.repo.Delete(id)
}

func (s *ProductService) AdjustStock(productID uint, adjustmentType string, quantity int, reason string, user string) error {
	product, err := s.repo.FindByID(productID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("PRODUCT_NOT_FOUND")
		}
		return err
	}

	switch adjustmentType {
	case "IN":
		product.StockQuantity += quantity
	case "OUT":
		if product.StockQuantity < quantity {
			return fmt.Errorf("INSUFFICIENT_STOCK")
		}
		product.StockQuantity -= quantity
	case "ADJUSTMENT":
		product.StockQuantity = quantity
	default:
		return fmt.Errorf("INVALID_ADJUSTMENT_TYPE")
	}

	if err := s.repo.Update(product); err != nil {
		return err
	}

	movement := &model.StockMovement{
		ProductID:   int(productID),
		Type:        adjustmentType,
		Quantity:    quantity,
		Reason:      reason,
		CreatedBy:   user,
	}
	return s.movementRepo.Create(movement)
}

func (s *ProductService) GetStockHistory(productID int) ([]model.StockMovement, error) {
	return s.movementRepo.FindByProductID(productID)
}

func (s *ProductService) GetLowStockProducts(threshold int) ([]model.Product, error) {
	products, err := s.repo.FindAll("", "", 0)
	if err != nil {
		return nil, err
	}
	var lowStock []model.Product
	for _, p := range products {
		if p.StockQuantity <= p.LowStockThreshold {
			lowStock = append(lowStock, p)
		}
	}
	return lowStock, nil
}

func (s *ProductService) DeductStock(tx *gorm.DB, productID int, quantity int, orderID string) error {
	var product model.Product
	if err := tx.First(&product, productID).Error; err != nil {
		return err
	}
	if product.StockQuantity < quantity {
		return fmt.Errorf("INSUFFICIENT_STOCK: product %s has %d stock, requested %d", product.Name, product.StockQuantity, quantity)
	}
	product.StockQuantity -= quantity
	if err := tx.Save(&product).Error; err != nil {
		return err
	}
	movement := &model.StockMovement{
		ProductID:   productID,
		Type:        "OUT",
		Quantity:    quantity,
		Reason:      "Order created",
		ReferenceID: orderID,
	}
	return tx.Create(movement).Error
}

func (s *ProductService) RestoreStock(tx *gorm.DB, productID int, quantity int, orderID string) error {
	var product model.Product
	if err := tx.First(&product, productID).Error; err != nil {
		return err
	}
	product.StockQuantity += quantity
	if err := tx.Save(&product).Error; err != nil {
		return err
	}
	movement := &model.StockMovement{
		ProductID:   productID,
		Type:        "IN",
		Quantity:    quantity,
		Reason:      "Order voided",
		ReferenceID: orderID,
	}
	return tx.Create(movement).Error
}

func (s *ProductService) ImportCSV(file interface{ Read([]byte) (int, error) }, filename string, storeType string) (*dto.ImportResult, error) {
	reader := csv.NewReader(file.(io.Reader))

	headers, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV header: %w", err)
	}

	colMap := make(map[string]int)
	for i, h := range headers {
		colMap[strings.TrimSpace(strings.ToLower(h))] = i
	}

	result := &dto.ImportResult{}
	rowNum := 1

	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			result.Failed++
			result.Errors = append(result.Errors, fmt.Sprintf("row %d: %v", rowNum, err))
			rowNum++
			continue
		}

		product, validationErr := parseCSVRow(row, colMap, storeType)
		if validationErr != "" {
			result.Failed++
			result.Errors = append(result.Errors, fmt.Sprintf("row %d: %s", rowNum, validationErr))
			rowNum++
			continue
		}

		if err := s.repo.Upsert(product); err != nil {
			result.Failed++
			result.Errors = append(result.Errors, fmt.Sprintf("row %d: failed to save: %v", rowNum, err))
		} else {
			result.Success++
		}

		rowNum++
	}

	return result, nil
}

func (s *ProductService) ExportCSV(w io.Writer, storeType string) error {
	products, err := s.repo.FindAll(storeType, "", 0)
	if err != nil {
		return err
	}

	writer := csv.NewWriter(w)
	writer.Write([]string{"id", "name", "price", "category", "image_url", "store_type", "stock_quantity", "low_stock_threshold"})

	for _, p := range products {
		writer.Write([]string{
			strconv.Itoa(int(p.ID)),
			p.Name,
			fmt.Sprintf("%.2f", p.Price),
			p.Category,
			p.ImageURL,
			p.StoreType,
			strconv.Itoa(p.StockQuantity),
			strconv.Itoa(p.LowStockThreshold),
		})
	}

	writer.Flush()
	return writer.Error()
}

func parseCSVRow(row []string, colMap map[string]int, defaultStoreType string) (*model.Product, string) {
	getCol := func(name string) string {
		if idx, ok := colMap[name]; ok && idx < len(row) {
			return strings.TrimSpace(row[idx])
		}
		return ""
	}

	name := getCol("name")
	if name == "" {
		return nil, "name is required"
	}

	priceStr := getCol("price")
	price, err := strconv.ParseFloat(priceStr, 64)
	if err != nil || price < 0 {
		return nil, "invalid price"
	}

	category := getCol("category")
	if category == "" {
		category = "Uncategorized"
	}

	imageURL := getCol("image_url")
	st := getCol("store_type")
	if st == "" {
		st = defaultStoreType
	}
	if st == "" {
		st = "RETAIL"
	}

	stockQty := 0
	if sq := getCol("stock_quantity"); sq != "" {
		if v, err := strconv.Atoi(sq); err == nil {
			stockQty = v
		}
	}

	lowThreshold := 10
	if lt := getCol("low_stock_threshold"); lt != "" {
		if v, err := strconv.Atoi(lt); err == nil {
			lowThreshold = v
		}
	}

	return &model.Product{
		Name:              name,
		Price:             price,
		Category:          category,
		ImageURL:          imageURL,
		StoreType:         st,
		StockQuantity:     stockQty,
		LowStockThreshold: lowThreshold,
	}, ""
}
