package service

import (
	"errors"
	"gbs-pos-api/internal/dto"
	"gbs-pos-api/internal/model"
	"gbs-pos-api/internal/repository"
	"strings"

	"gorm.io/gorm"
)

var (
	ErrPricingEmptyCart       = errors.New("PRICING_EMPTY_CART")
	ErrPricingInvalidQuantity = errors.New("PRICING_INVALID_QUANTITY")
)

type PricingService struct {
	productRepo     *repository.ProductRepository
	discountService *DiscountService
}

func NewPricingService(
	productRepo *repository.ProductRepository,
	discountService *DiscountService,
) *PricingService {
	return &PricingService{
		productRepo:     productRepo,
		discountService: discountService,
	}
}

func (s *PricingService) Calculate(
	req dto.PricingCalculationRequest,
) (*dto.PricingCalculationResponse, error) {
	if len(req.Items) == 0 {
		return nil, ErrPricingEmptyCart
	}

	productIDs := make([]uint, 0, len(req.Items))
	for _, item := range req.Items {
		if item.Qty <= 0 {
			return nil, ErrPricingInvalidQuantity
		}
		productIDs = append(productIDs, item.ProductID)
	}

	productDiscounts, err := s.discountService.GetActiveDiscountsByProductIDs(productIDs)
	if err != nil {
		return nil, err
	}

	results := make([]dto.PricingItemResult, 0, len(req.Items))
	var subtotal float64
	var productDiscountTotal float64
	for _, item := range req.Items {
		product, err := s.productRepo.FindByID(item.ProductID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, ErrDiscountProductNotFound
			}
			return nil, err
		}

		result := s.calculateItem(product, item.Qty, productDiscounts[item.ProductID])
		subtotal += result.OriginalSubtotal
		productDiscountTotal += result.DiscountAmount
		results = append(results, result)
	}

	runningTotal := subtotal - productDiscountTotal
	transactionDiscount, transactionAmount, err := s.discountService.SelectBestTransactionDiscount(
		runningTotal,
	)
	if err != nil {
		return nil, err
	}
	runningTotal = clampMoney(runningTotal - transactionAmount)

	var voucherDiscount *model.Discount
	var voucherAmount float64
	if req.VoucherCode != nil && strings.TrimSpace(*req.VoucherCode) != "" {
		voucherDiscount, voucherAmount, err = s.discountService.SelectBestVoucherDiscount(
			*req.VoucherCode,
			runningTotal,
		)
		if err != nil {
			return nil, err
		}
		runningTotal = clampMoney(runningTotal - voucherAmount)
	}

	totalDiscountAmount := productDiscountTotal + transactionAmount + voucherAmount
	return &dto.PricingCalculationResponse{
		Items:                     results,
		Subtotal:                  subtotal,
		ProductDiscountTotal:      productDiscountTotal,
		AfterProductDiscountTotal: subtotal - productDiscountTotal,
		TransactionDiscount:       toAppliedDiscountResponse(transactionDiscount, transactionAmount),
		VoucherDiscount:           toAppliedDiscountResponse(voucherDiscount, voucherAmount),
		TotalDiscountAmount:       totalDiscountAmount,
		FinalTotal:                runningTotal,
	}, nil
}

func (s *PricingService) calculateItem(
	product *model.Product,
	qty int,
	discount *model.Discount,
) dto.PricingItemResult {
	originalSubtotal := product.Price * float64(qty)
	discountAmount := 0.0
	if discount != nil {
		discountAmount = s.discountService.CalculateDiscountAmount(product.Price, discount) * float64(qty)
	}
	finalSubtotal := clampMoney(originalSubtotal - discountAmount)
	finalPrice := 0.0
	if qty > 0 {
		finalPrice = finalSubtotal / float64(qty)
	}

	return dto.PricingItemResult{
		ProductID:        product.ID,
		Qty:              qty,
		OriginalPrice:    product.Price,
		OriginalSubtotal: originalSubtotal,
		Discount:         toAppliedDiscountResponse(discount, discountAmount),
		DiscountAmount:   discountAmount,
		FinalPrice:       finalPrice,
		FinalSubtotal:    finalSubtotal,
	}
}

func toAppliedDiscountResponse(
	discount *model.Discount,
	amount float64,
) *dto.AppliedDiscountResponse {
	if discount == nil {
		return nil
	}

	return &dto.AppliedDiscountResponse{
		ID:     discount.ID,
		Scope:  discount.Scope,
		Code:   discount.VoucherCode,
		Name:   discount.Name,
		Type:   discount.Type,
		Value:  discount.Value,
		Amount: amount,
	}
}

func clampMoney(value float64) float64 {
	if value < 0 {
		return 0
	}
	return value
}
