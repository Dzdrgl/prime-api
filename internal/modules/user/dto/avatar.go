package dto

type CreateUploadURLRequest struct {
	ContentType string `json:"content_type" binding:"required"` // image/jpeg|png|webp
	Ext         string `json:"ext,omitempty"`                   // .jpg/.png/.webp
}

type CreateUploadURLResponse struct {
	Key          string `json:"key"`            // users/<uuid>/avatar.ext
	UploadURL    string `json:"upload_url"`     // presigned PUT
	ExpiresInSec int64  `json:"expires_in_sec"` // ör. 600
}

type ConfirmAvatarRequest struct {
	Key string `json:"key" binding:"required"` // aynı key
}

type ConfirmAvatarResponse struct {
	AvatarURL *string `json:"avatar_url,omitempty"`
}
