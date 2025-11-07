package router

import (
	"github.com/gin-gonic/gin"
	"githup.com/ademdizdar/prime-api/internal/app"
	"githup.com/ademdizdar/prime-api/internal/common"
	"githup.com/ademdizdar/prime-api/internal/middleware"
	"githup.com/ademdizdar/prime-api/internal/modules/auth"
	"githup.com/ademdizdar/prime-api/internal/modules/user"
	"go.uber.org/zap"
)

// JWT maker dışarıdan geliyor (main'den)
func RegisterV1Routes(r *gin.Engine, a *app.App, maker *common.JWTMaker, bl auth.Blacklist) {
	v1 := r.Group("/api/v1")

	// AUTH (public uçlar)
	if err := auth.Routes(v1, a, maker, a.Config().JWTAccessTTL, a.Config().RefreshTTL, []byte(a.Config().RefreshPepper), bl); err != nil {
		a.Logger().Fatal("auth routes", zap.Error(err))
	}

	// PROTECTED
	protected := v1.Group("")
	protected.Use(middleware.JWTAuth(maker, bl))
	if err := user.Routes(protected, a); err != nil {
		a.Logger().Fatal("user routes", zap.Error(err))
	}
}
