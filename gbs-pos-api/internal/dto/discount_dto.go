package dto

import "time"

type CreateDiscountRequest struct {
	ProductID uint    `json:"productId" binding:"required"`
	Name      string  `json:"name" binding:"required"`
	Type      string  `json:"type" binding:"required"`
	Value     float64 `json:"value" binding:"required"`
	StartDate string  `json:"startDate" binding:"required"`
	EndDate   string  `json:"endDate" binding:"required"`
}

type UpdateDiscountRequest struct {
	ProductID *uint    `json:"productId"`
	Name      string   `json:"name"`
	Type      string   `json:"type"`
	Value     *float64 `json:"value"`
	StartDate string   `json:"startDate"`
	EndDate   string   `json:"endDate"`
}

type DiscountResponse struct {
	ID              uint      `json:"id"`
	ProductID       uint      `json:"productId"`
	Name            string    `json:"name"`
	Type            string    `json:"type"`
	Value           float64   `json:"value"`
	StartDate       time.Time `json:"startDate"`
	EndDate         time.Time `json:"endDate"`
	Status          string    `json:"status"`
	EffectiveStatus string    `json:"effectiveStatus"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
}
