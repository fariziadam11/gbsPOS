package model

import "time"

type FuelPrice struct {
	Code          string  `gorm:"primaryKey" json:"code"`
	Name          string  `json:"name"`
	PricePerLiter float64 `json:"pricePerLiter"`
	UpdatedAt     time.Time
}
