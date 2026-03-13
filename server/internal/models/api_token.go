package models

import (
	"time"

	"gorm.io/gorm"
)

type APIToken struct {
	gorm.Model
	UserID    uint      `json:"user_id" gorm:"index;not null"`
	Token     string    `json:"token" gorm:"uniqueIndex;not null"`
	Name      string    `json:"name" gorm:"size:100;not null"`
	ExpiresAt time.Time `json:"expires_at" gorm:"not null"`
	IsRevoked bool      `json:"is_revoked" gorm:"default:false"`
}
