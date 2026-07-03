package model

type Nozzle struct {
	ID       string `gorm:"primaryKey" json:"id"`
	PumpID   string `json:"pumpId"`
	Name     string `json:"name"`
	FuelCode string `json:"fuelCode"`
	IsActive bool   `json:"isActive"`
}
