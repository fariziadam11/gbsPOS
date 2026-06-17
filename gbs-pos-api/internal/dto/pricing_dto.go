package dto

type PricingItemInput struct {
	ProductID uint `json:"productId" binding:"required"`
	Qty       int  `json:"qty" binding:"required"`
}

type PricingCalculationRequest struct {
	Items       []PricingItemInput `json:"items" binding:"required"`
	Subtotal    float64            `json:"subtotal"`
	VoucherCode *string            `json:"voucherCode"`
}

type AppliedDiscountResponse struct {
	ID     uint    `json:"id"`
	Scope  string  `json:"scope"`
	Code   *string `json:"code,omitempty"`
	Name   string  `json:"name"`
	Type   string  `json:"type"`
	Value  float64 `json:"value"`
	Amount float64 `json:"amount"`
}

type PricingItemResult struct {
	ProductID        uint                     `json:"productId"`
	Qty              int                      `json:"qty"`
	OriginalPrice    float64                  `json:"originalPrice"`
	OriginalSubtotal float64                  `json:"originalSubtotal"`
	Discount         *AppliedDiscountResponse `json:"discount,omitempty"`
	DiscountAmount   float64                  `json:"discountAmount"`
	FinalPrice       float64                  `json:"finalPrice"`
	FinalSubtotal    float64                  `json:"finalSubtotal"`
}

type PricingCalculationResponse struct {
	Items                     []PricingItemResult      `json:"items"`
	Subtotal                  float64                  `json:"subtotal"`
	ProductDiscountTotal      float64                  `json:"productDiscountTotal"`
	AfterProductDiscountTotal float64                  `json:"afterProductDiscountTotal"`
	TransactionDiscount       *AppliedDiscountResponse `json:"transactionDiscount,omitempty"`
	VoucherDiscount           *AppliedDiscountResponse `json:"voucherDiscount,omitempty"`
	TotalDiscountAmount       float64                  `json:"totalDiscountAmount"`
	FinalTotal                float64                  `json:"finalTotal"`
}
