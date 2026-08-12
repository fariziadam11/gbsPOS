package websocket

import "encoding/json"

type Message struct {
	Type          string  `json:"type"`
	PaymentID     string  `json:"payment_id,omitempty"`
	OrderID       string  `json:"order_id,omitempty"`
	Amount        float64 `json:"amount,omitempty"`
	Currency      string  `json:"currency,omitempty"`
	ExpiresAt     string  `json:"expires_at,omitempty"`
	TerminalID    string  `json:"terminal_id,omitempty"`
	Status        string  `json:"status,omitempty"`
	TransactionID string  `json:"transaction_id,omitempty"`
	CardBrand     string  `json:"card_brand,omitempty"`
	MaskedCard    string  `json:"masked_card,omitempty"`
	AuthCode      string  `json:"auth_code,omitempty"`
	EntryMode     string  `json:"entry_mode,omitempty"`
	AcqMID        string  `json:"acq_mid,omitempty"`
	AcqTID        string  `json:"acq_tid,omitempty"`
	PosMessageID  string  `json:"pos_message_id,omitempty"`
	FailureReason string  `json:"failure_reason,omitempty"`
	UpdatedAt     string  `json:"updated_at,omitempty"`
	Message       string  `json:"message,omitempty"`
	DeviceID      string  `json:"device_id,omitempty"`
}

func DecodeMessage(data []byte) (Message, error) {
	var message Message
	if err := json.Unmarshal(data, &message); err != nil {
		return Message{}, err
	}
	return message, nil
}
