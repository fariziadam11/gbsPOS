package model

import (
	"time"

	"gorm.io/datatypes"
)

type User struct {
	ID           uint      `gorm:"primaryKey"                   json:"id"`
	Username     string    `gorm:"size:50;uniqueIndex;not null" json:"username"`
	PasswordHash string    `gorm:"size:255;not null" json:"-"`
	Name         string    `gorm:"size:100" json:"name"`
	Role         string    `gorm:"size:20;not null" json:"role"`
	Gender       string    `gorm:"size:100" json:"gender"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

type Product struct {
	ID                uint      `gorm:"primaryKey"                                      json:"id"`
	Name              string    `gorm:"size:200;not null"                               json:"name"`
	Price             float64   `gorm:"type:decimal(12,2);not null"                     json:"price"`
	Category          string    `gorm:"size:100;not null"                               json:"category"`
	ImageURL          string    `gorm:"size:500"                                        json:"imageUrl"`
	StoreType         string    `gorm:"size:20;not null"                                json:"storeType"`
	StockQuantity     int       `gorm:"not null;default:0"                              json:"stockQuantity"`
	LowStockThreshold int       `gorm:"not null;default:10"                             json:"lowStockThreshold"`
	Barcode           string    `gorm:"size:100;index"                                  json:"barcode"`
	CreatedAt         time.Time `                                                        json:"createdAt"`
	UpdatedAt         time.Time `                                                        json:"updatedAt"`
}

const (
	DiscountScopeProduct     = "PRODUCT"
	DiscountScopeTransaction = "TRANSACTION"
	DiscountScopeVoucher     = "VOUCHER"
)

type Discount struct {
	ID             uint      `gorm:"primaryKey"                            json:"id"`
	ProductID      *uint     `gorm:"index"                                 json:"productId,omitempty"`
	Product        *Product  `gorm:"foreignKey:ProductID;constraint:OnDelete:CASCADE" json:"-"`
	Scope          string    `gorm:"size:20;not null;default:PRODUCT;index" json:"scope"`
	VoucherCode    *string   `gorm:"size:100;index"                         json:"voucherCode,omitempty"`
	MinTransaction float64   `gorm:"type:decimal(12,2);not null;default:0"  json:"minTransaction"`
	Name           string    `gorm:"size:200;not null"                      json:"name"`
	Type           string    `gorm:"size:20;not null;index"                 json:"type"` // PERCENTAGE, FIXED
	Value          float64   `gorm:"type:decimal(12,2);not null"            json:"value"`
	StartDate      time.Time `gorm:"not null;index"                         json:"startDate"`
	EndDate        time.Time `gorm:"not null;index"                         json:"endDate"`
	Status         string    `gorm:"size:20;not null;index"                 json:"status"` // PENDING, ACTIVE, EXPIRED, STOPPED, CANCELLED
	CreatedAt      time.Time `                                              json:"createdAt"`
	UpdatedAt      time.Time `                                              json:"updatedAt"`
}

type Order struct {
	ID                  string      `gorm:"primaryKey;size:32" json:"id"`
	Subtotal            float64     `gorm:"type:decimal(12,2);not null" json:"subtotal"`
	Tax                 float64     `gorm:"type:decimal(12,2);not null" json:"tax"`
	Total               float64     `gorm:"type:decimal(12,2);not null" json:"total"`
	PaymentMethod       string      `gorm:"size:20;not null" json:"paymentMethod"`
	CashReceived        *float64    `gorm:"type:decimal(12,2)" json:"cashReceived"`
	ChangeAmount        *float64    `gorm:"type:decimal(12,2)" json:"changeAmount"`
	Timestamp           int64       `gorm:"not null" json:"timestamp"`
	IsVoided            bool        `gorm:"not null;default:false" json:"isVoided"`
	IsSettled           bool        `gorm:"not null;default:false" json:"isSettled"`
	TransactionID       string      `gorm:"size:100" json:"transactionId"`
	ApprovalCode        string      `gorm:"size:50" json:"approvalCode"`
	EntryMode           string      `gorm:"size:20" json:"entryMode"`
	MaskedAccount       string      `gorm:"size:50" json:"maskedAccount"`
	AcqMid              string      `gorm:"size:50" json:"acqMid"`
	AcqTid              string      `gorm:"size:50" json:"acqTid"`
	PosMessageID        string      `gorm:"size:100" json:"posMessageId"`
	BankName            string      `gorm:"size:50" json:"bankName"`
	StoreType           string      `gorm:"size:20" json:"storeType"`
	TerminalID          string      `gorm:"size:32" json:"terminalId"`
	VoidReason          string      `gorm:"size:255" json:"voidReason"`
	VoidedBy            string      `gorm:"size:50" json:"voidedBy"`
	VoidedAt            *time.Time  `json:"voidedAt"`
	CustomerID          *int        `gorm:"index" json:"customerId"`
	CustomerPhone       string      `gorm:"size:50" json:"customerPhone"`
	CustomerName        string      `gorm:"size:255" json:"customerName"`
	LoyaltyPointsEarned int         `gorm:"not null;default:0" json:"loyaltyPointsEarned"`
	DiscountType        string      `gorm:"size:20" json:"discountType"`
	DiscountValue       *float64    `gorm:"type:decimal(12,2)" json:"discountValue"`
	DiscountAmount      *float64    `gorm:"type:decimal(12,2)" json:"discountAmount"`
	// QRIS Payment Gateway Fields (SumoPod)
	QrisPaymentID   string     `gorm:"size:100;index" json:"qrisPaymentId,omitempty"`
	QrisStatus      string     `gorm:"size:20;default:pending;index" json:"qrisStatus,omitempty"`
	QrisLinkURL     string     `gorm:"type:text" json:"qrisLinkUrl,omitempty"`
	QrisExpiresAt   *time.Time `json:"qrisExpiresAt,omitempty"`
	QrisFee         float64    `gorm:"type:decimal(12,2);default:0" json:"qrisFee,omitempty"`
	QrisNetAmount   float64    `gorm:"type:decimal(12,2);default:0" json:"qrisNetAmount,omitempty"`
	QrisCompletedAt *time.Time `json:"qrisCompletedAt,omitempty"`
	CreatedAt       time.Time  `json:"createdAt"`
	UpdatedAt       time.Time  `json:"updatedAt"`
	Items           []OrderItem `gorm:"foreignKey:OrderID;constraint:OnDelete:CASCADE" json:"items"`
}

type OrderItem struct {
	ID           uint    `gorm:"primaryKey"                  json:"-"`
	OrderID      string  `gorm:"size:32;not null;index"      json:"-"`
	ProductID    int     `gorm:"not null"                    json:"productId"`
	ProductName  string  `gorm:"size:200;not null"           json:"productName"`
	ProductPrice float64 `gorm:"type:decimal(12,2);not null" json:"productPrice"`
	Qty          int     `gorm:"not null;check:qty > 0"      json:"qty"`
	Subtotal     float64 `gorm:"type:decimal(12,2);not null" json:"subtotal"`
	VariantID    *int    `json:"variantId"`
	VariantName  string  `gorm:"size:255" json:"variantName"`
	SKU          string  `gorm:"size:100" json:"sku"`
}

type Settlement struct {
	ID          string    `gorm:"primaryKey;size:64"          json:"id"`
	Timestamp   int64     `gorm:"not null"                    json:"timestamp"`
	BatchCount  int       `gorm:"not null"                    json:"batchCount"`
	TotalAmount float64   `gorm:"type:decimal(12,2);not null" json:"totalAmount"`
	CardTotal   float64   `gorm:"type:decimal(12,2);not null" json:"cardTotal"`
	QRISTotal   float64   `gorm:"type:decimal(12,2);not null" json:"qrisTotal"`
	CashTotal   float64   `gorm:"type:decimal(12,2);not null" json:"cashTotal"`
	Status      string    `gorm:"size:20;not null"            json:"status"`
	StoreType   string    `gorm:"size:20"                     json:"storeType"`
	TerminalID  string    `gorm:"size:32"                     json:"terminalId"`
	CreatedAt   time.Time `                                   json:"createdAt"`
}

type Customer struct {
	ID            uint      `gorm:"primaryKey"          json:"id"`
	Name          string    `gorm:"size:255"            json:"name"`
	Phone         string    `gorm:"size:50;uniqueIndex" json:"phone"`
	Email         string    `gorm:"size:255"            json:"email"`
	Address       string    `gorm:"type:text"           json:"address"`
	LoyaltyPoints int       `gorm:"not null;default:0" json:"loyaltyPoints"`
	CreatedAt     time.Time `                           json:"createdAt"`
	UpdatedAt     time.Time `                           json:"updatedAt"`
}

type StockMovement struct {
	ID          uint      `gorm:"primaryKey"                   json:"id"`
	ProductID   int       `gorm:"not null;index"               json:"productId"`
	Type        string    `gorm:"size:20;not null"             json:"type"` // IN, OUT, ADJUSTMENT
	Quantity    int       `gorm:"not null"                     json:"quantity"`
	Reason      string    `gorm:"size:255"                     json:"reason"`
	ReferenceID string    `gorm:"size:100"                     json:"referenceId"`
	CreatedBy   string    `gorm:"size:100"                     json:"createdBy"`
	CreatedAt   time.Time `                                    json:"createdAt"`
}

type ProductVariant struct {
	ID                int                    `gorm:"primaryKey"                             json:"id"`
	ProductID         int                    `gorm:"not null;index"                         json:"productId"`
	SKU               string                 `gorm:"size:100"                                json:"sku"`
	Name              string                 `gorm:"size:255;not null"                       json:"name"`
	Attributes        map[string]interface{} `gorm:"serializer:json;not null"               json:"attributes"`
	Price             *float64               `gorm:"type:decimal(12,2)"                     json:"price"`
	StockQuantity     int                    `gorm:"not null;default:0"                      json:"stockQuantity"`
	LowStockThreshold *int                   `                                              json:"lowStockThreshold"`
	Barcode           string                 `gorm:"size:100;index"                          json:"barcode"`
	IsActive          bool                   `gorm:"not null;default:true"                   json:"isActive"`
	SortOrder         int                    `gorm:"not null;default:0"                      json:"sortOrder"`
	CreatedAt         time.Time              `                                              json:"createdAt"`
	UpdatedAt         time.Time              `                                              json:"updatedAt"`
}

type HoldStatus string

const (
	HoldStatusActive  HoldStatus = "ACTIVE"
	HoldStatusResumed HoldStatus = "RESUMED"
)

type HoldSession struct {
	ID         string         `gorm:"type:uuid;primaryKey" json:"id"`
	StoreType  string         `json:"storeType"`
	TerminalID string         `json:"terminalId"`
	Payload    datatypes.JSON `json:"payload"`
	Total      float64        `json:"total"`
	Status     HoldStatus     `json:"status"`
	CreatedAt  time.Time      `json:"createdAt"`
	UpdatedAt  time.Time      `json:"updatedAt"`
}

func (HoldSession) TableName() string {
	return "pos_hold_sessions"
}
