// +++ DEĞİŞEN (etiketler eklendi) +++
package user

import (
	"time"

	"github.com/google/uuid"
)

func (User) TableName() string { return "app_users" }

type User struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey"`
	Email        string    `gorm:"uniqueIndex;not null"`
	FullName     string    `gorm:"not null;size:100"`
	PasswordHash string    `gorm:"not null"`
	AvatarURL    *string   `gorm:"null"`
	IsActive     bool      `gorm:"not null;default:true"`
	LastLoginAt  *time.Time
	CreatedAt    time.Time  `gorm:"not null"`
	UpdatedAt    time.Time  `gorm:"not null"`
	DeletedAt    *time.Time `gorm:"index"`
}

func NewUser(email, fullName, passwordHash string) *User {
	now := time.Now().UTC()
	return &User{
		ID:           uuid.New(),
		Email:        email,
		FullName:     fullName,
		PasswordHash: passwordHash,
		IsActive:     true,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}
