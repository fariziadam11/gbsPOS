package dto

type SettingsResponse struct {
	Settings map[string]string `json:"settings"`
}

type UpdateSettingsRequest struct {
	Settings map[string]string `json:"settings" binding:"required"`
}
