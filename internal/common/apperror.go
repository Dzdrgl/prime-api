package common

import (
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type ErrorCode string

const (
	CodeBadRequest   ErrorCode = "BAD_REQUEST"
	CodeValidation   ErrorCode = "VALIDATION_FAILED"
	CodeNotFound     ErrorCode = "NOT_FOUND"
	CodeConflict     ErrorCode = "CONFLICT"
	CodeInternal     ErrorCode = "INTERNAL"
	CodeUnauthorized ErrorCode = "UNAUTHORIZED"
	CodeForbidden    ErrorCode = "FORBIDDEN"
)

type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Tag     string `json:"tag,omitempty"` // min,max,email vb.
}

type APIError struct {
	Code      ErrorCode    `json:"code"`
	Message   string       `json:"message"`
	Fields    []FieldError `json:"fields,omitempty"`
	RequestID string       `json:"request_id,omitempty"`
}

type Envelope[T any] struct {
	Data  *T        `json:"data,omitempty"`
	Meta  any       `json:"meta,omitempty"`
	Error *APIError `json:"error,omitempty"`
}

// --- Helpers ---

func ValidationFrom(err error) *APIError {
	ve, ok := err.(validator.ValidationErrors)
	if !ok {
		return &APIError{Code: CodeBadRequest, Message: "invalid payload"}
	}
	fields := make([]FieldError, 0, len(ve))
	for _, fe := range ve {
		fields = append(fields, FieldError{
			Field:   fe.Field(),
			Message: fe.Error(), // istersen özelleştir
			Tag:     fe.Tag(),
		})
	}
	return &APIError{
		Code:    CodeValidation,
		Message: "validation failed",
		Fields:  fields,
	}
}

func WriteError(c *gin.Context, status int, apiErr *APIError) {
	if rid := c.GetString("request_id"); rid != "" {
		apiErr.RequestID = rid
	}
	c.JSON(status, Envelope[any]{Error: apiErr})
}

func WriteOK[T any](c *gin.Context, status int, data *T, meta any) {
	c.JSON(status, Envelope[T]{Data: data, Meta: meta})
}
