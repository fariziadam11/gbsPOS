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
	OrderID        string    `json:"orderId"`
	PaymentID     string    `json:"paymentId"`
	PaymentLinkURL string   `json:"paymentLinkUrl"`
	Amount        float64   `json:"amount"`
	Fee           float64   `json:"fee"`
	ExpiresAt     time.Time `json:"expiresAt"`
}

// QRIS Direct (Static to Dynamic) DTOs

// ConvertQRISRequest represents a request to convert static to dynamic QRIS
// The static QRIS string is taken from config (QRIS_DIRECT_STATIC_QRIS)
type ConvertQRISRequest struct {
	Amount   float64 `json:"amount" binding:"required,gt=0"`
	FeeType  string  `json:"feeType,omitempty"`   // "fixed" or "percentage"
	FeeValue float64 `json:"feeValue,omitempty"`  // Fee amount or percentage
	OrderID  string  `json:"orderId,omitempty"`   // Optional order association
}

// ConvertQRISResponse represents the converted QRIS data
type ConvertQRISResponse struct {
	ID           string    `json:"id"`
	OrderID      string    `json:"orderId,omitempty"`
	OriginalQris string    `json:"originalQris,omitempty"`
	DynamicQris  string    `json:"dynamicQris"`
	Amount       float64   `json:"amount"`
	FeeType      string    `json:"feeType,omitempty"`
	FeeValue     float64   `json:"feeValue,omitempty"`
	FeeAmount    float64   `json:"feeAmount"`
	TotalAmount  float64   `json:"totalAmount"`
	MerchantName string    `json:"merchantName"`
	MerchantCity string    `json:"merchantCity"`
	Provider     string    `json:"provider"`
	QRCodeBase64 string    `json:"qrCodeBase64"`
	ExpiresAt    time.Time `json:"expiresAt"`
}

// GetQRISStatusResponse represents QRIS transaction status
type GetQRISStatusResponse struct {
	ID          string     `json:"id"`
	OrderID     string     `json:"orderId"`
	Amount      float64    `json:"amount"`
	FeeAmount   float64    `json:"feeAmount"`
	TotalAmount float64    `json:"totalAmount"`
	Provider    string     `json:"provider"`
	Status      string     `json:"status"`
	PaidAt      *time.Time `json:"paidAt,omitempty"`
	ExpiresAt   time.Time  `json:"expiresAt"`
	CreatedAt   time.Time  `json:"createdAt"`
}

// ConfirmQRISPaymentRequest represents a request to confirm payment
type ConfirmQRISPaymentRequest struct {
	TransactionID string `json:"transactionId" binding:"required"`
}

// CancelQRISPaymentRequest represents a request to cancel payment
type CancelQRISPaymentRequest struct {
	TransactionID string `json:"transactionId" binding:"required"`
	Reason        string `json:"reason,omitempty"`
}

// GenerateQRCodeRequest represents a request to generate QR code image
type GenerateQRCodeRequest struct {
	QrisString string `json:"qrisString" binding:"required"`
}

// GenerateQRCodeResponse represents QR code image response
type GenerateQRCodeResponse struct {
	Base64 string `json:"base64"`
	URL    string `json:"url"`
}
