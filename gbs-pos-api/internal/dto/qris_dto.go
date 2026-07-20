package dto

import "time"

// SumoPod API Request - Create Payment
type SumoPodCreatePaymentRequest struct {
	OrderID              string  `json:"order_id"`
	Amount               float64 `json:"amount"`
	Currency             string  `json:"currency"`
	ExpiresInHours       int     `json:"expires_in_hours,omitempty"`
	SuccessReturnURL     string  `json:"success_return_url,omitempty"`
	CancelReturnURL      string  `json:"cancel_return_url,omitempty"`
	PaymentMethodTypeCode string  `json:"payment_method_type_code,omitempty"`
}

// SumoPod API Response - Create Payment
type SumoPodCreatePaymentResponse struct {
	PaymentID     string    `json:"payment_id"`
	OrderID       string    `json:"order_id"`
	Amount        float64   `json:"amount"`
	Fee           float64   `json:"fee"`
	NetAmount     float64   `json:"net_amount"`
	PaymentLinkURL string   `json:"payment_link_url"`
	Status        string    `json:"status"`
	ExpiresAt     time.Time `json:"expires_at"`
}

// SumoPod Webhook Payload
type SumoPodWebhookPayload struct {
	EventType string                       `json:"event_type"`
	Data      SumoPodWebhookPaymentData     `json:"data"`
}

type SumoPodWebhookPaymentData struct {
	PaymentID     string    `json:"payment_id"`
	OrderID       string    `json:"order_id"`
	Amount        float64   `json:"amount"`
	Fee           float64   `json:"fee"`
	NetAmount     float64   `json:"net_amount"`
	Status        string    `json:"status"`
	PaymentMethod string    `json:"payment_method"`
	CompletedAt   time.Time `json:"completed_at"`
}

// QRIS Status enum
const (
	QrisStatusPending   = "pending"
	QrisStatusCompleted = "completed"
	QrisStatusFailed    = "failed"
	QrisStatusExpired   = "expired"
)

// Webhook Event Types
const (
	WebhookEventPaymentCompleted = "payment.completed"
	WebhookEventPaymentFailed   = "payment.failed"
	WebhookEventPaymentExpired  = "payment.expired"
	WebhookEventPaymentTest     = "payment.test"
)

// QRIS Payment DTOs for internal use
type CreateQrisPaymentRequest struct {
	OrderID string  `json:"orderId" binding:"required"`
	Amount  float64 `json:"amount" binding:"required,gt=0"`
}

type QrisPaymentStatusResponse struct {
	OrderID       string     `json:"orderId"`
	PaymentID     string     `json:"paymentId"`
	Status        string     `json:"status"`
	PaymentLinkURL string    `json:"paymentLinkUrl,omitempty"`
	ExpiresAt     *time.Time `json:"expiresAt,omitempty"`
	Fee           float64    `json:"fee"`
	NetAmount     float64    `json:"netAmount"`
	CompletedAt   *time.Time `json:"completedAt,omitempty"`
}

type QrisInitResponse struct {
	OrderID       string    `json:"orderId"`
	PaymentID     string    `json:"paymentId"`
	PaymentLinkURL string   `json:"paymentLinkUrl"`
	Amount        float64   `json:"amount"`
	Fee           float64   `json:"fee"`
	ExpiresAt     time.Time `json:"expiresAt"`
}
