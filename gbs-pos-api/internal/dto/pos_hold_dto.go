package dto

import (
	"encoding/json"
	"time"
)

type CreatePosHoldRequest struct {
	StoreType  string          `json:"storeType"  binding:"required"`
	TerminalID string          `json:"terminalId" binding:"required"`
	Payload    json.RawMessage `json:"payload"    binding:"required"`
	Total      float64         `json:"total"      binding:"required"`
}

type PosHoldResponse struct {
	ID         string          `json:"id"`
	StoreType  string          `json:"storeType"`
	TerminalID string          `json:"terminalId"`
	Payload    json.RawMessage `json:"payload"`
	Total      float64         `json:"total"`
	Status     string          `json:"status"`
	CreatedAt  time.Time       `json:"createdAt"`
	UpdatedAt  time.Time       `json:"updatedAt"`
}
