package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"githup.com/ademdizdar/prime-api/internal/common"
)

// NoRoute ve NoMethod handler’ları için yardımcılar
func HandleNoRoute() gin.HandlerFunc {
	return func(c *gin.Context) {
		common.WriteError(c, http.StatusNotFound, &common.APIError{
			Code:    common.CodeNotFound,
			Message: "route not found",
		})
	}
}

func HandleNoMethod() gin.HandlerFunc {
	return func(c *gin.Context) {
		common.WriteError(c, http.StatusMethodNotAllowed, &common.APIError{
			Code:    common.CodeBadRequest, // istersen CodeNotFound yerine CodeBadRequest/başka bir kod
			Message: "method not allowed",
		})
	}
}
