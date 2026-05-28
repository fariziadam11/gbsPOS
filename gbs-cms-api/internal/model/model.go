package model

import "time"

type User struct {
	ID           uint      `gorm:"primaryKey"                   json:"id"`
	Username     string    `gorm:"size:50;uniqueIndex;not null" json:"username"`
	PasswordHash string    `gorm:"size:255;not null"            json:"-"`
	Name         string    `gorm:"size:100"                     json:"name"`
	Role         string    `gorm:"size:20;not null"             json:"role"`
	CreatedAt    time.Time `                                    json:"createdAt"`
	UpdatedAt    time.Time `                                    json:"updatedAt"`
}

type Ad struct {
	ID              uint       `gorm:"primaryKey"                         json:"id"`
	Name            string     `gorm:"size:200;not null"                  json:"name"`
	Filename        string     `gorm:"size:255;not null"                  json:"filename"`
	StoragePath     string     `gorm:"size:500;not null"                  json:"storagePath"`
	FileSize        int64      `gorm:"not null"                           json:"fileSize"`
	MimeType        string     `gorm:"size:50;not null"                   json:"mimeType"`
	DurationSeconds *int       `                                          json:"durationSeconds"`
	StoreTypes      []string   `gorm:"type:text;serializer:json;not null" json:"storeTypes"`
	PlaylistOrder   int        `gorm:"not null;default:0"                 json:"playlistOrder"`
	IsActive        bool       `gorm:"not null;"                           json:"isActive"`
	StartDate       *time.Time `gorm:"type:date"                          json:"startDate"`
	EndDate         *time.Time `gorm:"type:date"                          json:"endDate"`
	StartTime       *time.Time `gorm:"type:time"                          json:"startTime"`
	EndTime         *time.Time `gorm:"type:time"                          json:"endTime"`
	CreatedBy       uint       `gorm:"not null"                           json:"createdBy"`
	CreatedAt       time.Time  `                                          json:"createdAt"`
	UpdatedAt       time.Time  `                                          json:"updatedAt"`
}

type AdPlayLog struct {
	ID         uint      `gorm:"primaryKey"              json:"id"`
	AdID       uint      `gorm:"not null;index"          json:"adId"`
	TerminalID string    `gorm:"size:32"                 json:"terminalId"`
	StoreType  string    `gorm:"size:20;not null"        json:"storeType"`
	PlayedAt   time.Time `gorm:"not null;autoCreateTime" json:"playedAt"`
}
