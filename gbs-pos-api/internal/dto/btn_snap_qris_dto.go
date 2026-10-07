package dto

type BtnSnapTokenRequest struct {
	GrantType      string         `json:"grantType"`
	AdditionalInfo map[string]any `json:"additionalInfo"`
}

type BtnSnapTokenResponse struct {
	ResponseCode    string `json:"responseCode"`
	ResponseMessage string `json:"responseMessage"`
	AccessToken     string `json:"accessToken"`
	TokenType       string `json:"tokenType"`
	ExpiresIn       string `json:"expiresIn"`
}

type BtnSnapAmount struct {
	Value    string `json:"value"`
	Currency string `json:"currency"`
}

type BtnSnapGenerateRequest struct {
	PartnerReferenceNo string         `json:"partnerReferenceNo"`
	Amount             BtnSnapAmount  `json:"amount"`
	MerchantID         string         `json:"merchantId"`
	TerminalID         string         `json:"terminalId"`
	ValidityPeriod     string         `json:"validityPeriod,omitempty"`
	AdditionalInfo     map[string]any `json:"additionalInfo"`
}

type BtnSnapGenerateResponse struct {
	ResponseCode       string `json:"responseCode"`
	ResponseMessage    string `json:"responseMessage"`
	ReferenceNo        string `json:"referenceNo"`
	PartnerReferenceNo string `json:"partnerReferenceNo"`
	QRContent          string `json:"qrContent"`
	MerchantName       string `json:"merchantName"`
	TerminalID         string `json:"terminalId"`
}

type BtnSnapQueryRequest struct {
	OriginalPartnerReferenceNo string         `json:"originalPartnerReferenceNo"`
	OriginalReferenceNo        string         `json:"originalReferenceNo,omitempty"`
	ServiceCode                string         `json:"serviceCode"`
	MerchantID                 string         `json:"merchantId"`
	AdditionalInfo             map[string]any `json:"additionalInfo"`
}

type BtnSnapQueryResponse struct {
	ResponseCode               string        `json:"responseCode"`
	ResponseMessage            string        `json:"responseMessage"`
	OriginalReferenceNo        string        `json:"originalReferenceNo"`
	OriginalPartnerReferenceNo string        `json:"originalPartnerReferenceNo"`
	ServiceCode                string        `json:"serviceCode"`
	LatestTransactionStatus    string        `json:"latestTransactionStatus"`
	TransactionStatusDesc      string        `json:"transactionStatusDesc"`
	PaidTime                   string        `json:"paidTime"`
	TerminalID                 string        `json:"terminalId"`
	Amount                     BtnSnapAmount `json:"amount"`
	AdditionalInfo             struct {
		MerchantID string `json:"merchantId"`
	} `json:"additionalInfo"`
}

type QrisCheckoutItem struct {
	ProductID    int     `json:"productId"`
	ProductName  string  `json:"productName"`
	ProductPrice float64 `json:"productPrice"`
	Qty          int     `json:"qty"`
	Subtotal     float64 `json:"subtotal"`
	VariantID    *int    `json:"variantId,omitempty"`
	VariantName  string  `json:"variantName,omitempty"`
	SKU          string  `json:"sku,omitempty"`
}

type QrisCheckoutPpobItem struct {
	PpobProductID string `json:"ppobProductId"`
	ProductName   string `json:"productName"`
	Provider      string `json:"provider"`
	Category      string `json:"category"`
	CustomerID    string `json:"customerId"`
	CustomerName  string `json:"customerName,omitempty"`
	Amount        int64  `json:"amount"`
	AdminFee      int64  `json:"adminFee"`
	Total         int64  `json:"total"`
}

type QrisCheckoutSnapshot struct {
	ID             string                 `json:"id"`
	Items          []QrisCheckoutItem     `json:"items"`
	PpobItems      []QrisCheckoutPpobItem `json:"ppobItems,omitempty"`
	Subtotal       float64                `json:"subtotal"`
	Tax            float64                `json:"tax"`
	Total          float64                `json:"total"`
	DiscountType   string                 `json:"discountType,omitempty"`
	DiscountValue  *float64               `json:"discountValue,omitempty"`
	DiscountAmount *float64               `json:"discountAmount,omitempty"`
}

type QrisCheckoutSnapshotRequest struct {
	ID             string                 `json:"id" binding:"required"`
	Items          []QrisCheckoutItem     `json:"items"`
	PpobItems      []QrisCheckoutPpobItem `json:"ppobItems,omitempty"`
	Subtotal       float64                `json:"subtotal"`
	Tax            float64                `json:"tax"`
	Total          float64                `json:"total"`
	DiscountType   string                 `json:"discountType,omitempty"`
	DiscountValue  *float64               `json:"discountValue,omitempty"`
	DiscountAmount *float64               `json:"discountAmount,omitempty"`
}
