package dto

type UnsettledSummary struct {
	Count          int                             `json:"count"`
	Total          float64                         `json:"total"`
	PaymentSummary map[string]PaymentMethodSummary `json:"paymentSummary"`
}

type PaymentMethodSummary struct {
	Count int     `json:"count"`
	Total float64 `json:"total"`
}