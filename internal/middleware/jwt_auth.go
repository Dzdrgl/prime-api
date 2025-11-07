package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"githup.com/ademdizdar/prime-api/internal/common"
)

func JWTAuth(maker *common.JWTMaker, bl interface{ Has(string) bool }) gin.HandlerFunc {
	return func(c *gin.Context) {
		h := c.GetHeader("Authorization")
		if !strings.HasPrefix(strings.ToLower(h), "bearer ") {
			common.WriteError(c, http.StatusUnauthorized, &common.APIError{Code: common.CodeBadRequest, Message: "missing bearer token"})
			c.Abort()
			return
		}
		raw := strings.TrimSpace(h[len("Bearer "):])
		parsed, err := jwt.ParseWithClaims(raw, &common.Claims{}, func(t *jwt.Token) (any, error) { return maker.Secret, nil }, jwt.WithAudience(maker.Audience), jwt.WithIssuer(maker.Issuer))
		if err != nil { /* ... aynı hata ... */
			common.WriteError(c, http.StatusUnauthorized, &common.APIError{Code: common.CodeBadRequest, Message: "invalid token"})
			c.Abort()
			return
		}

		claims := parsed.Claims.(*common.Claims)
		// blacklist kontrolü
		if bl != nil && claims.ID != "" && bl.Has(claims.ID) {
			common.WriteError(c, http.StatusUnauthorized, &common.APIError{Code: common.CodeBadRequest, Message: "token revoked"})
			c.Abort()
			return
		}
		c.Set("user_id", claims.UserID)
		c.Set("user_email", claims.Email)
		c.Set("access_jti", claims.ID)
		c.Next()
	}
}
