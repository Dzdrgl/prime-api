package apidoc

import (
	common "githup.com/ademdizdar/prime-api/internal/common"
	"githup.com/ademdizdar/prime-api/internal/modules/auth/dto"
)

// --- Ortak meta ---

// @name PageMeta
type PageMeta struct {
	Page  int `json:"page"`
	Limit int `json:"limit"`
	// gerekirse: TotalItems, TotalPages vb. ekleyebilirsin
}

// --- Auth ---

// @name LoginEnvelope
type LoginEnvelope struct {
	Data  dto.LoginResponse `json:"data"`
	Meta  interface{}       `json:"meta,omitempty"`
	Error *common.APIError  `json:"error,omitempty"`
}

// @name RefreshEnvelope
type RefreshEnvelope struct {
	Data  dto.RefreshResponse `json:"data"`
	Meta  interface{}         `json:"meta,omitempty"`
	Error *common.APIError    `json:"error,omitempty"`
}
