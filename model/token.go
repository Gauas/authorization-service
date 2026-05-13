package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Token struct {
	ID           uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID       uuid.UUID  `gorm:"type:uuid;not null;index"`
	DeviceID     string     `gorm:"not null"`
	Permission   string     `gorm:"not null"`
	RefreshToken string     `gorm:"uniqueIndex;not null"`
	ExpiresAt    time.Time  `gorm:"not null"`
	RevokedAt    gorm.DeletedAt `gorm:"column:revoked_at"`
	CreatedAt    time.Time
}

func (Token) TableName() string { return "refresh_tokens" }
