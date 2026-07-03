package model

import "time"

type FuelSale struct {
	ID            string    `gorm:"primaryKey" json:"id"`
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
	CreatedAt     time.Time
}
