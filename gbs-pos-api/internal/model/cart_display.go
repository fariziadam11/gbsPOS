package model

import (
	"time"

	"gorm.io/datatypes"
)

// CartDisplay stores the latest JSON payload for a POS terminal customer display.
// The backend treats Payload as an opaque JSONB document.
type CartDisplay struct {
	ID         uint           `gorm:"primaryKey"`
	TerminalID string         `gorm:"uniqueIndex;size:64;not null"`
	Payload    datatypes.JSON `gorm:"type:jsonb;not null"`
	UpdatedAt  time.Time
}
