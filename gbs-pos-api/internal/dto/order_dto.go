package dto

import (
	"gbs-pos-api/internal/model"
)

type BulkSyncOrderRequest struct {
	TerminalID string        `json:"terminalId"`
	Orders     []model.Order `json:"orders" binding:"required"`
}

type SettleOrderRequest struct {
	SettlementID string `json:"settlementId" binding:"required"`
	Timestamp    int64  `json:"timestamp" binding:"required"`
	StoreType    string `json:"storeType"`
	TerminalID   string `json:"terminalId"`
}

type CreateOrderRequest struct {
	ID    string `json:"id" binding:"required"`
	Items []struct {
		ProductID    int     `json:"productId" binding:"required"`
		ProductName  string  `json:"productName" binding:"required"`
		ProductPrice float64 `json:"productPrice" binding:"required"`
		Qty          int     `json:"qty" binding:"required"`
		Subtotal     float64 `json:"subtotal" binding:"required"`
	} `json:"items" binding:"required"`
	Subtotal      float64  `json:"subtotal" binding:"required"`
	Tax           float64  `json:"tax" binding:"required"`
	Total         float64  `json:"total" binding:"required"`
	PaymentMethod string   `json:"paymentMethod" binding:"required"`
	CashReceived  *float64 `json:"cashReceived"`
	ChangeAmount  *float64 `json:"changeAmount"`
	Timestamp     int64    `json:"timestamp" binding:"required"`
	StoreType     string   `json:"storeType"`
	TerminalID    string   `json:"terminalId"`
	TransactionID string   `json:"transactionId"`
	ApprovalCode  string   `json:"approvalCode"`
	EntryMode     string   `json:"entryMode"`
	MaskedAccount string   `json:"maskedAccount"`
	AcqMid        string   `json:"acqMid"`
	AcqTid        string   `json:"acqTid"`
	PosMessageID  string   `json:"posMessageId"`
	BankName      string   `json:"bankName"`
}

type VoidOrderRequest struct {
	Reason string `json:"reason"`
}

type PaymentSummary struct {
	Count int
	Total float64
}

type PaymentMethodQueryResult struct {
	PaymentMethod string
	Count         int
	Total         float64
}

type BulkSyncResult struct {
	Orders     []model.Order `json:"orders"`
	Created    int           `json:"created"`
	Existing   int           `json:"existing"`
	Failed     int           `json:"failed"`
	Idempotent bool          `json:"idempotent"`
}