package dto

// GET /api/v1/users  → listeleme sorguları (page/limit common’da)
type UserListQuery struct {
	Email  string `form:"email"`  // opsiyonel filtre
	Active *bool  `form:"active"` // ?active=true/false
}

// GET /api/v1/users/:id  → param zaten path’de

// POST /api/v1/users  → create
type CreateUserRequest struct {
	Email         string  `json:"email" binding:"required,email,max=254" validate:"required,email,max=254"`
	FullName      string  `json:"full_name" binding:"required,min=2,max=100" validate:"required,min=2,max=100"`
	Password      string  `json:"password" binding:"required,min=8,max=72" validate:"required,min=8,max=72"`
	AvatarDataURL *string `json:"avatar_data_url,omitempty"`
}

// PATCH /api/v1/users/:id  → partial update (örnek)
type UpdateUserRequest struct {
	FullName  *string `json:"full_name" binding:"omitempty,min=2,max=100"`
	AvatarURL *string `json:"avatar_url" binding:"omitempty,url"`
	Active    *bool   `json:"active" binding:"omitempty"`
}
