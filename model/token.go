package model

import (
	"time"

	"github.com/google/uuid"
)

type Token struct {
	ID           uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID       uuid.UUID  `gorm:"type:uuid;not null;index"`
	DeviceID     string     `gorm:"not null"`
	Permission   string     `gorm:"not null"`
	RefreshToken string     `gorm:"uniqueIndex;not null"`
	IssuedAt     time.Time  `gorm:"not null"`
	ExpiresAt    time.Time  `gorm:"not null"`
	RevokedAt    *time.Time `gorm:"default:null"`
	CreatedAt    time.Time
}

func (Token) TableName() string { return "refresh_tokens" }
