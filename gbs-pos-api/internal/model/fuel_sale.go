package model

import "time"

// FuelSale status lifecycle
const (
	FuelSaleStatusPaid        = "PAID"        // sale is paid, not yet scanned by the dispenser
	FuelSaleStatusAuthorized  = "AUTHORIZED"  // scanned & authorized once by the dispenser
)

type FuelSale struct {
	ID                 string     `gorm:"primaryKey" json:"id"`
	PumpID             string     `json:"pumpId"`
	NozzleID           string     `json:"nozzleId"`
	FuelCode           string     `json:"fuelCode"`
	PricePerLiter      float64    `json:"pricePerLiter"`
	Liters             float64    `json:"liters"`
	TotalAmount        float64    `json:"totalAmount"`
	PaymentMethod      string     `json:"paymentMethod"`
	TransactionID      string     `json:"transactionId,omitempty"`
	PosMessageID       string     `json:"posMessageId,omitempty"`
	Timestamp          time.Time  `json:"timestamp"`
	CreatedAt          time.Time
	ReceiptToken       string     `gorm:"size:64;uniqueIndex" json:"receiptToken,omitempty"`
	AuthorizationToken string     `gorm:"size:64;uniqueIndex" json:"authorizationToken,omitempty"`
	Status             string     `gorm:"size:16;default:PAID" json:"status"`
	AuthorizedAt       *time.Time `json:"authorizedAt,omitempty"`
}

