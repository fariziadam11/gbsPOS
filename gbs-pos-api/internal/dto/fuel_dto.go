package dto

import "time"

// Fuel price
type FuelPriceResponse struct {
	Code          string  `json:"code"`
	Name          string  `json:"name"`
	PricePerLiter float64 `json:"pricePerLiter"`
	UpdatedAt     int64   `json:"updatedAt"`
}

type UpdateFuelPriceRequest struct {
	PricePerLiter float64 `json:"pricePerLiter" binding:"required,min=0"`
}

// Pump
type PumpResponse struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	IsActive bool   `json:"isActive"`
}

type CreatePumpRequest struct {
	ID   string `json:"id" binding:"required"`
	Name string `json:"name" binding:"required"`
}

type UpdatePumpRequest struct {
	Name     string `json:"name"`
	IsActive *bool  `json:"isActive"`
}

// Nozzle
type NozzleResponse struct {
	ID       string `json:"id"`
	PumpID   string `json:"pumpId"`
	Name     string `json:"name"`
	FuelCode string `json:"fuelCode"`
	IsActive bool   `json:"isActive"`
}

type CreateNozzleRequest struct {
	ID       string `json:"id" binding:"required"`
	PumpID   string `json:"pumpId" binding:"required"`
	Name     string `json:"name" binding:"required"`
	FuelCode string `json:"fuelCode" binding:"required"`
}

type UpdateNozzleRequest struct {
	Name     string `json:"name"`
	FuelCode string `json:"fuelCode"`
	IsActive *bool  `json:"isActive"`
}

// Fuel sale
type FuelSaleRequest struct {
	ID            string    `json:"id" binding:"required"`
	PumpID        string    `json:"pumpId" binding:"required"`
	NozzleID      string    `json:"nozzleId" binding:"required"`
	FuelCode      string    `json:"fuelCode" binding:"required"`
	PricePerLiter float64   `json:"pricePerLiter" binding:"required,min=0"`
	Liters        float64   `json:"liters" binding:"required,min=0"`
	TotalAmount   float64   `json:"totalAmount" binding:"required,min=0"`
	PaymentMethod string    `json:"paymentMethod" binding:"required"`
	TransactionID string    `json:"transactionId"`
	PosMessageID  string    `json:"posMessageId"`
	Timestamp     int64     `json:"timestamp"`
}

type FuelSaleResponse struct {
	ID            string    `json:"id"`
	PumpID        string    `json:"pumpId"`
	NozzleID      string    `json:"nozzleId"`
	FuelCode      string    `json:"fuelCode"`
	PricePerLiter float64   `json:"pricePerLiter"`
	Liters        float64   `json:"liters"`
	TotalAmount   float64   `json:"totalAmount"`
	PaymentMethod string    `json:"paymentMethod"`
	TransactionID string    `json:"transactionId,omitempty"`
	PosMessageID  string    `json:"posMessageId,omitempty"`
	Timestamp     time.Time `json:"timestamp"`
}

type FuelReportItem struct {
	FuelCode    string  `json:"fuelCode"`
	Liters      float64 `json:"liters"`
	TotalAmount float64 `json:"totalAmount"`
}

type PumpReportItem struct {
	PumpID      string  `json:"pumpId"`
	Liters      float64 `json:"liters"`
	TotalAmount float64 `json:"totalAmount"`
}

type FuelSalesReportResponse struct {
	Summary    []FuelReportItem `json:"summary"`
	PumpTotals []PumpReportItem `json:"pumpTotals"`
}
