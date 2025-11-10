// @title           Prime API
// @version         1.0
// @description     Prime API için OpenAPI dokümantasyonu.
// @BasePath        /api/v1

// JWT Bearer şeması:
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// Örnek: "Authorization: Bearer <token>"
package main

import (
	"context"

	"github.com/gin-gonic/gin"
	"githup.com/ademdizdar/prime-api/internal/app"
	"githup.com/ademdizdar/prime-api/internal/common"
	"githup.com/ademdizdar/prime-api/internal/middleware"
	"githup.com/ademdizdar/prime-api/internal/modules/auth"
	"githup.com/ademdizdar/prime-api/internal/router"
	"go.uber.org/zap"

	_ "githup.com/ademdizdar/prime-api/internal/apidocs"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func main() {
	// 1) Config
	cfg, err := app.LoadConfig()
	if err != nil {
		panic(err) // logger henüz yok
	}

	// 2) Logger
	logger, err := app.NewLogger(cfg.LogLevel)
	if err != nil {
		panic(err)
	}
	defer logger.Sync()
	// 2.1) JWT Maker
	jwtMaker := &common.JWTMaker{
		Secret:   []byte(cfg.JWTSecret),
		Issuer:   cfg.JWTIssuer,
		Audience: cfg.JWTAudience,
	}

	// 3) DB + migrate
	gdb, err := app.InitGorm(cfg.DatabaseURL)
	if err != nil {
		logger.Fatal("db init", zap.Error(err))
	}

	rc, err := app.InitRedis(cfg.RedisAddr, cfg.RedisPass, cfg.RedisDB)
	if err != nil {
		logger.Fatal("redis init", zap.Error(err))
	}
	bl := auth.NewRedisBlacklist(rc, "bl:jti:")

	// 3.1) Object Storage (Minio/S3)
	store, err := app.InitMinio(cfg)
	if err != nil {
		logger.Fatal("minio init", zap.Error(err))
	}
	// 4) App container
	container, err := app.New(logger, gdb, cfg, rc, store)
	if err != nil {
		logger.Fatal("app init", zap.Error(err))
	}
	defer container.Close(context.Background())

	// 5) Gin + routes
	r := gin.New()
	r.Use(middleware.RequestID())
	r.Use(middleware.Logging(logger))
	r.Use(middleware.RecoveryJSON(logger))

	// Global 404/405
	r.NoRoute(middleware.HandleNoRoute())
	r.NoMethod(middleware.HandleNoMethod())

	// Swagger UI (sadece dev/stage ortamında aç)
	if cfg.AppEnv != "production" {
		r.GET("/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

	router.RegisterV1Routes(r, container, jwtMaker, bl)

	// 6) Server run
	srv := app.NewHTTPServer(r, cfg.AppPort)
	wait := cfg.ShutdownWait
	app.RunHTTPServer(logger, srv, wait)
}
