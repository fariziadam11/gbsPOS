package model

import (
	"time"
)

// CartDisplay stores the latest JSON payload for a POS terminal customer display.
type CartDisplay struct {
	ID         uint      `gorm:"primaryKey"`
	TerminalID string    `gorm:"uniqueIndex;size:64;not null"`
	Payload    string    `gorm:"type:jsonb;not null"` // stores entire JSON
	UpdatedAt  time.Time `gorm:"autoUpdateTime"`
}
