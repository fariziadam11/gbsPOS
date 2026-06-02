package dto

import "gbs-pos-api/internal/model"

type CreateCustomerRequest struct {
	Name    string `json:"name"`
	Phone   string `json:"phone" binding:"required"`
	Email   string `json:"email"`
	Address string `json:"address"`
}

type UpdateCustomerRequest struct {
	Name    string `json:"name"`
	Phone   string `json:"phone"`
	Email   string `json:"email"`
	Address string `json:"address"`
}

type AdjustStockRequest struct {
	Type     string `json:"type" binding:"required,oneof=IN OUT ADJUSTMENT"`
	Quantity int    `json:"quantity" binding:"required,min=1"`
	Reason   string `json:"reason"`
}

type CustomerResponse struct {
	Customer      model.Customer `json:"customer"`
	OrderHistory  []model.Order  `json:"orderHistory,omitempty"`
	TotalSpent    float64        `json:"totalSpent"`
	TotalOrders   int            `json:"totalOrders"`
}
