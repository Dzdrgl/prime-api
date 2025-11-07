package auth

import (
	"time"

	"github.com/gin-gonic/gin"
	"githup.com/ademdizdar/prime-api/internal/app"
	"githup.com/ademdizdar/prime-api/internal/common"
	"githup.com/ademdizdar/prime-api/internal/middleware"
)

func Routes(v1 *gin.RouterGroup, a *app.App, maker *common.JWTMaker,
	accessTTL, refreshTTL time.Duration, pepper []byte, bl Blacklist) error {

	h, err := NewHandler(a, maker, accessTTL, refreshTTL, pepper, bl)
	if err != nil {
		return err
	}

	g := v1.Group("/auth")
	{
		// Mobil cihaz bazlı rate-limit (header: X-Device-ID)
		g.POST("/login",
			middleware.RateLimitRedis(
				a.Redis(),
				"rl:login:",
				a.Config().LoginRate,
				a.Config().RateWindow,
				middleware.KeyByDeviceID("X-Device-ID"),
			),
			h.Login,
		)

		g.POST("/refresh",
			middleware.RateLimitRedis(
				a.Redis(),
				"rl:refresh:",
				a.Config().RefreshRate,
				a.Config().RateWindow,
				middleware.KeyByDeviceID("X-Device-ID"),
			),
			h.Refresh,
		)

		g.POST("/logout", h.Logout)
	}

	return nil
}
