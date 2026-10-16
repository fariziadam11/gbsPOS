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
	ValidityPeriod     string         `json:"validityPeriod"`
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

type BtnSnapScannerInfo struct {
	DeviceID string `json:"deviceId,omitempty"`
}

type BtnSnapCPMDecodeRequest struct {
	PartnerReferenceNo string `json:"partnerReferenceNo"`
	QRContent          string `json:"qrContent"`
	MerchantID         string `json:"merchantId"`
	ScanTime           string `json:"scanTime"`
	AdditionalInfo     any    `json:"additionalInfo,omitempty"`
}

type BtnSnapCPMDecodeResponse struct {
	ResponseCode       string `json:"responseCode"`
	ResponseMessage    string `json:"responseMessage"`
	PartnerReferenceNo string `json:"partnerReferenceNo"`
	ExpiryDateTime     string `json:"expiryDateTime"`
	IssuerName         string `json:"issuerName"`
	CustomerName       string `json:"customerName"`
	CustomerPAN        string `json:"customerPAN"`
}

type BtnSnapCPMPaymentRequest struct {
	PartnerReferenceNo string             `json:"partnerReferenceNo"`
	QRContent          string             `json:"qrContent"`
	Amount             BtnSnapAmount      `json:"amount"`
	MerchantID         string             `json:"merchantId"`
	MerchantName       string             `json:"merchantName"`
	MerchantLocation   string             `json:"merchantLocation"`
	TerminalID         string             `json:"terminalId"`
	Title              string             `json:"title,omitempty"`
	ExpiryTime         string             `json:"expiryTime"`
	ScannerInfo        BtnSnapScannerInfo `json:"scannerInfo,omitempty"`
	AdditionalInfo     map[string]any     `json:"additionalInfo"`
}

type BtnSnapCPMPaymentResponse struct {
	ResponseCode       string `json:"responseCode"`
	ResponseMessage    string `json:"responseMessage"`
	ReferenceNo        string `json:"referenceNo"`
	PartnerReferenceNo string `json:"partnerReferenceNo"`
	TransactionDate    string `json:"transactionDate"`
	TransactionStatus  string `json:"transactionStatus"`
	AdditionalInfo     struct {
		TransactionStatus string `json:"transactionStatus"`
	} `json:"additionalInfo"`
}

type BtnSnapCPMQueryRequest struct {
	OriginalPartnerReferenceNo string         `json:"originalPartnerReferenceNo"`
	OriginalReferenceNo        string         `json:"originalReferenceNo,omitempty"`
	MerchantID                 string         `json:"merchantId"`
	TerminalID                 string         `json:"terminalId"`
	MerchantPAN                string         `json:"merchantPAN"`
	AdditionalInfo             map[string]any `json:"additionalInfo"`
}

type BtnSnapCPMQueryResponse struct {
	ResponseCode               string `json:"responseCode"`
	ResponseMessage            string `json:"responseMessage"`
	OriginalReferenceNo        string `json:"originalReferenceNo"`
	OriginalPartnerReferenceNo string `json:"originalPartnerReferenceNo"`
	LatestTransactionStatus    string `json:"latestTransactionStatus"`
	TransactionStatusDesc      string `json:"transactionStatusDesc"`
	PaidTime                   string `json:"paidTime"`
	AdditionalInfo             struct {
		MerchantPAN string             `json:"merchantPAN"`
		Amount      BtnSnapAmount      `json:"amount"`
		TerminalID  string             `json:"terminalId"`
		ScannerInfo BtnSnapScannerInfo `json:"scannerInfo"`
		DeviceID    string             `json:"deviceId"`
	} `json:"additionalInfo"`
}

type BtnSnapCPMCancelRequest struct {
	OriginalPartnerReferenceNo string         `json:"originalPartnerReferenceNo"`
	OriginalReferenceNo        string         `json:"originalReferenceNo,omitempty"`
	MerchantID                 string         `json:"merchantId"`
	Amount                     BtnSnapAmount  `json:"amount"`
	Reason                     string         `json:"reason,omitempty"`
	AdditionalInfo             map[string]any `json:"additionalInfo"`
}

type BtnSnapCPMCancelResponse struct {
	ResponseCode               string `json:"responseCode"`
	ResponseMessage            string `json:"responseMessage"`
	OriginalReferenceNo        string `json:"originalReferenceNo"`
	OriginalPartnerReferenceNo string `json:"originalPartnerReferenceNo"`
}

type BtnSnapCPMRefundRequest struct {
	OriginalPartnerReferenceNo string         `json:"originalPartnerReferenceNo"`
	OriginalReferenceNo        string         `json:"originalReferenceNo,omitempty"`
	MerchantID                 string         `json:"merchantId"`
	PartnerRefundNo            string         `json:"partnerRefundNo"`
	RefundAmount               BtnSnapAmount  `json:"refundAmount"`
	Reason                     string         `json:"reason,omitempty"`
	AdditionalInfo             map[string]any `json:"additionalInfo"`
}

type BtnSnapCPMRefundResponse struct {
	ResponseCode               string        `json:"responseCode"`
	ResponseMessage            string        `json:"responseMessage"`
	OriginalReferenceNo        string        `json:"originalReferenceNo"`
	OriginalPartnerReferenceNo string        `json:"originalPartnerReferenceNo"`
	RefundNo                   string        `json:"refundNo"`
	PartnerRefundNo            string        `json:"partnerRefundNo"`
	RefundAmount               BtnSnapAmount `json:"refundAmount"`
	RefundTime                 string        `json:"refundTime"`
}

type BtnSnapMPMRefundRequest struct {
	MerchantID                 string         `json:"merchantId"`
	OriginalPartnerReferenceNo string         `json:"originalPartnerReferenceNo"`
	OriginalReferenceNo        string         `json:"originalReferenceNo,omitempty"`
	PartnerRefundNo            string         `json:"partnerRefundNo"`
	RefundAmount               BtnSnapAmount  `json:"refundAmount"`
	Reason                     string         `json:"reason"`
	AdditionalInfo             map[string]any `json:"additionalInfo"`
}

type BtnSnapMPMRefundResponse struct {
	ResponseCode               string        `json:"responseCode"`
	ResponseMessage            string        `json:"responseMessage"`
	OriginalPartnerReferenceNo string        `json:"originalPartnerReferenceNo"`
	OriginalReferenceNo        string        `json:"originalReferenceNo"`
	RefundNo                   string        `json:"refundNo"`
	PartnerRefundNo            string        `json:"partnerRefundNo"`
	RefundAmount               BtnSnapAmount `json:"refundAmount"`
	RefundTime                 string        `json:"refundTime"`
}

type BtnSnapMPMReportRequest struct {
	PartnerReferenceNo string         `json:"partnerReferenceNo,omitempty"`
	FromDateTime       string         `json:"fromDateTime"`
	ToDateTime         string         `json:"toDateTime"`
	PageSize           string         `json:"pageSize"`
	PageNumber         string         `json:"pageNumber"`
	AdditionalInfo     map[string]any `json:"additionalInfo"`
}

type BtnSnapMPMReportResponse struct {
	ResponseCode       string `json:"responseCode"`
	ResponseMessage    string `json:"responseMessage"`
	ReferenceNo        string `json:"referenceNo"`
	PartnerReferenceNo string `json:"partnerReferenceNo"`
	DetailData         struct {
		DateTime       string        `json:"dateTime"`
		PaidTime       string        `json:"paidTime"`
		Amount         BtnSnapAmount `json:"amount"`
		Status         string        `json:"status"`
		Type           string        `json:"type"`
		AdditionalInfo struct {
			TransactionType      string `json:"transactionType"`
			TransactionMethod    string `json:"transactionMethod"`
			ChannelResponseCode  string `json:"channelResponseCode"`
			ChannelApprovalCode  string `json:"channelApprovalCode"`
			STAN                 string `json:"stan"`
			RetrievalReferenceNo string `json:"retrievalReferenceNo"`
			IssuerName           string `json:"issuerName"`
		} `json:"additionalInfo"`
	} `json:"detailData"`
	AdditionalInfo struct {
		Transaction      string `json:"transaction"`
		MerchantID       string `json:"merchantId"`
		MerchantName     string `json:"merchantName"`
		MerchantCategory string `json:"merchantCategory"`
		TotalData        string `json:"totalData"`
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
