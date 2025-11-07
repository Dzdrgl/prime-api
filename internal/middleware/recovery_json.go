package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"githup.com/ademdizdar/prime-api/internal/common"
	"go.uber.org/zap"
)

// Panic'leri yakalayıp JSON hata döner
func RecoveryJSON(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if rec := recover(); rec != nil {
				logger.Error("panic recovered", zap.Any("panic", rec))
				common.WriteError(c, http.StatusInternalServerError, &common.APIError{
					Code:    common.CodeInternal,
					Message: "internal server error",
				})
				c.Abort()
			}
		}()
		c.Next()
	}
}
