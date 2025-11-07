package middleware

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"githup.com/ademdizdar/prime-api/internal/common"
	"golang.org/x/net/context"
)

// keyFunc ip veya device_id bazlı anahtar üretir
type KeyFunc func(c *gin.Context) string

// Sabit pencere limitleyici: window süresince max istek
func RateLimitRedis(rc *redis.Client, namespace string, max int, window time.Duration, keyFn KeyFunc) gin.HandlerFunc {
	if namespace == "" {
		namespace = "rl:"
	}
	return func(c *gin.Context) {
		ctx := context.Background()
		key := namespace + keyFn(c)
		pipe := rc.TxPipeline()
		incr := pipe.Incr(ctx, key)
		pipe.Expire(ctx, key, window)
		_, _ = pipe.Exec(ctx)

		if n, err := incr.Result(); err != nil || int(n) > max {
			common.WriteError(c, http.StatusTooManyRequests, &common.APIError{
				Code:    common.CodeBadRequest,
				Message: "too many requests",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

// Örnek key’ler
func KeyByIP() KeyFunc {
	return func(c *gin.Context) string {
		return "ip:" + c.ClientIP()
	}
}

func KeyByDeviceID(field string) KeyFunc {
	return func(c *gin.Context) string {
		d := c.GetHeader(field) // ör: "X-Device-ID"
		if d == "" {
			d = "unknown"
		}
		return "dev:" + d
	}
}
