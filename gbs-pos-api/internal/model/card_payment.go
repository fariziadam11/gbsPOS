package model

import (
	"time"

	"github.com/google/uuid"
)

const (
	CardPaymentWaiting    = "WAITING_FOR_CARD"
	CardPaymentProcessing = "PROCESSING"
	CardPaymentSuccess    = "SUCCESS"
	CardPaymentFailed     = "FAILED"
	CardPaymentCancelled  = "CANCELLED"
	CardPaymentExpired    = "EXPIRED"
)

type CardPayment struct {
	ID            uuid.UUID `gorm:"type:uuid;primaryKey" json:"paymentId"`
	OrderID       string    `gorm:"size:50;not null;index" json:"orderId"`
	Amount        float64   `gorm:"type:decimal(15,2);not null" json:"amount"`
	Status        string    `gorm:"size:30;not null;index" json:"status"`
	DeviceID      string    `gorm:"size:100;index" json:"deviceId"`
	TerminalID    string    `gorm:"size:100;index" json:"terminalId"`
	TransactionID string    `gorm:"size:100" json:"transactionId,omitempty"`
	CardBrand     string    `gorm:"size:20" json:"cardBrand,omitempty"`
	MaskedCard    string    `gorm:"size:30" json:"maskedCard,omitempty"`
	AuthCode      string    `gorm:"size:30" json:"authCode,omitempty"`
	EntryMode     string    `gorm:"size:30" json:"entryMode,omitempty"`
	AcqMID        string    `gorm:"size:50" json:"acqMid,omitempty"`
	AcqTID        string    `gorm:"size:50" json:"acqTid,omitempty"`
	PosMessageID  string    `gorm:"size:100" json:"posMessageId,omitempty"`
	FailureReason string    `gorm:"type:text" json:"failureReason,omitempty"`
	ExpiresAt     time.Time `gorm:"not null;index" json:"expiresAt"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}
