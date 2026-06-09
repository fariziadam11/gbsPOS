package dto

import "time"

type ProductDiscountResponse struct {
	ID     uint    `json:"id"`
	Name   string  `json:"name"`
	Type   string  `json:"type"`
	Value  float64 `json:"value"`
	Status string  `json:"status"`
}

type ProductResponse struct {
	ID                uint                     `json:"id"`
	Name              string                   `json:"name"`
	Price             float64                  `json:"price"`
	Category          string                   `json:"category"`
	ImageURL          string                   `json:"imageUrl"`
	StoreType         string                   `json:"storeType"`
	StockQuantity     int                      `json:"stockQuantity"`
	LowStockThreshold int                      `json:"lowStockThreshold"`
	Discount          *ProductDiscountResponse `json:"discount"`
	FinalPrice        float64                  `json:"finalPrice"`
	CreatedAt         time.Time                `json:"createdAt"`
	UpdatedAt         time.Time                `json:"updatedAt"`
}
