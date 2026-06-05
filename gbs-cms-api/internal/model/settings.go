package model

import "time"

type Setting struct {
	ID        uint      `gorm:"primaryKey"                   json:"id"`
	Key       string    `gorm:"size:100;uniqueIndex;not null" json:"key"`
	Value     string    `gorm:"type:text;not null"            json:"value"`
	CreatedAt time.Time `                                     json:"createdAt"`
	UpdatedAt time.Time `                                     json:"updatedAt"`
}
