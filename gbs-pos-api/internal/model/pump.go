package model

type Pump struct {
	ID       string `gorm:"primaryKey" json:"id"`
	Name     string `json:"name"`
	IsActive bool   `json:"isActive"`
}
