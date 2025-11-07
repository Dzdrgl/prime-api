package apidoc

import (
	common "githup.com/ademdizdar/prime-api/internal/common"
	"githup.com/ademdizdar/prime-api/internal/modules/user/dto"
)

// @name UserEnvelope
type UserEnvelope struct {
	Data  dto.UserResponse `json:"data"`
	Meta  interface{}      `json:"meta,omitempty"`
	Error *common.APIError `json:"error,omitempty"`
}

// @name UserListEnvelope
type UserListEnvelope struct {
	Data  dto.UserListResponse `json:"data"`
	Meta  PageMeta             `json:"meta"`
	Error *common.APIError     `json:"error,omitempty"`
}
