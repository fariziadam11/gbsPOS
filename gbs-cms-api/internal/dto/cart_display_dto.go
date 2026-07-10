package dto

// DeviceInfo represents device information sent from the Android POS app.
type DeviceInfo struct {
	DeviceModel        string `json:"deviceModel"`
	DeviceManufacturer string `json:"deviceManufacturer"`
	DeviceBrand        string `json:"deviceBrand"`
	AndroidVersion     string `json:"androidVersion"`
	SDKInt             int    `json:"sdkInt"`
	AppVersion         string `json:"appVersion"`
	AppVersionCode     int64  `json:"appVersionCode"`
}

// TerminalListItem is the response type for listing terminals.
type TerminalListItem struct {
	TerminalID         string  `json:"terminalId"`
	Payload            string  `json:"payload"`
	DeviceModel        *string `json:"deviceModel,omitempty"`
	DeviceManufacturer *string `json:"deviceManufacturer,omitempty"`
	DeviceBrand        *string `json:"deviceBrand,omitempty"`
	AndroidVersion     *string `json:"androidVersion,omitempty"`
	SDKInt             *int    `json:"sdkInt,omitempty"`
	AppVersion         *string `json:"appVersion,omitempty"`
	AppVersionCode     *int64  `json:"appVersionCode,omitempty"`
	UpdatedAt          string  `json:"updatedAt"`
}
