package dto

import "time"

// Tekil user gösterimi (public)
type UserResponse struct {
	ID          string     `json:"id"`
	Email       string     `json:"email"`
	FullName    string     `json:"full_name"`
	AvatarURL   *string    `json:"avatar_url,omitempty"`
	IsActive    bool       `json:"is_active"`
	LastLoginAt *time.Time `json:"last_login_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// Liste dönüşü
type UserListResponse struct {
	Items []UserResponse `json:"items"`
}
