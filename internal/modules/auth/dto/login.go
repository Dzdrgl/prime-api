package dto

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8,max=72"`
	DeviceID string `json:"device_id" binding:"required,min=3,max=128"`
}

type LoginResponse struct {
	AccessToken  string  `json:"access_token"`
	AccessExp    int64   `json:"access_exp"`
	RefreshToken string  `json:"refresh_token"`
	RefreshExp   int64   `json:"refresh_exp"`
	SessionID    string  `json:"session_id"`
	AvatarURL    *string `json:"avatar_url,omitempty"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
	DeviceID     string `json:"device_id" binding:"required"`
}

type RefreshResponse struct {
	AccessToken  string `json:"access_token"`
	AccessExp    int64  `json:"access_exp"`
	RefreshToken string `json:"refresh_token"`
	RefreshExp   int64  `json:"refresh_exp"`
	SessionID    string `json:"session_id"`
}

type LogoutRequest struct {
	SessionID string `json:"session_id" binding:"required"`
}
