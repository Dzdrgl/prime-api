package auth

import (
	"time"

	"github.com/google/uuid"
)

type UserSession struct {
	ID              uuid.UUID `gorm:"type:uuid;primaryKey"`
	UserID          uuid.UUID `gorm:"type:uuid;index"`
	DeviceID        string    `gorm:"size:128;index"`
	RefreshHash     string    `gorm:"size:128;not null"`
	PrevRefreshHash *string   `gorm:"size:128"` // reuse detection
	IssuedAt        time.Time `gorm:"not null"`
	ExpiresAt       time.Time `gorm:"not null;index"`
	LastUsedAt      time.Time `gorm:"not null"`
	IP              string    `gorm:"size:64"`
	UserAgent       string    `gorm:"size:256"`
	IsRevoked       bool      `gorm:"not null;default:false;index"`
	CreatedAt       time.Time `gorm:"not null"`
	UpdatedAt       time.Time `gorm:"not null"`
}

type AuditEvent struct {
	ID        uuid.UUID  `gorm:"type:uuid;primaryKey"`
	UserID    *uuid.UUID `gorm:"type:uuid;index"`
	SessionID *uuid.UUID `gorm:"type:uuid;index"`
	Event     string     `gorm:"size:64;index"` // login, refresh, logout, revoke, reuse_detected
	IP        string     `gorm:"size:64"`
	UserAgent string     `gorm:"size:256"`
	CreatedAt time.Time  `gorm:"not null"`
}
