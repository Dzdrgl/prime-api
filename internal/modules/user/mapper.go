package user

import "githup.com/ademdizdar/prime-api/internal/modules/user/dto"

// DOĞRU, SAF (PURE) UYGULAMA
func ToUserResponse(u *User) dto.UserResponse {
	if u == nil {
		return dto.UserResponse{}
	}
	return dto.UserResponse{
		ID:          u.ID.String(),
		Email:       u.Email,
		FullName:    u.FullName,
		AvatarURL:   u.AvatarURL,
		IsActive:    u.IsActive,
		LastLoginAt: u.LastLoginAt,
		CreatedAt:   u.CreatedAt,
		UpdatedAt:   u.UpdatedAt,
	}
}

// Liste için de sadece okuma:
func ToUserListResponse(list []*User) dto.UserListResponse {
	out := make([]dto.UserResponse, 0, len(list))
	for _, u := range list {
		out = append(out, ToUserResponse(u))
	}
	return dto.UserListResponse{Items: out}
}
