package app

import (
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	AppEnv       string        `mapstructure:"APP_ENV"`
	AppPort      string        `mapstructure:"APP_PORT"`
	LogLevel     string        `mapstructure:"LOG_LEVEL"`
	DatabaseURL  string        `mapstructure:"DATABASE_URL"`
	ShutdownWait time.Duration `mapstructure:"SHUTDOWN_WAIT"`
	// JWT
	JWTSecret     string        `mapstructure:"JWT_SECRET"`
	JWTIssuer     string        `mapstructure:"JWT_ISSUER"`
	JWTAudience   string        `mapstructure:"JWT_AUDIENCE"`
	JWTAccessTTL  time.Duration `mapstructure:"JWT_ACCESS_TTL"` // 15m
	RefreshTTL    time.Duration `mapstructure:"REFRESH_TTL"`    // 720h (30d) öneri
	RefreshPepper string        `mapstructure:"REFRESH_PEPPER"` // server-side secret for hashing
	// Rate Limiting
	LoginRate   int           `mapstructure:"LOGIN_RATE"`   // 5
	RefreshRate int           `mapstructure:"REFRESH_RATE"` // 5
	RateWindow  time.Duration `mapstructure:"RATE_WINDOW"`  // 1m

	// Redis
	RedisAddr string `mapstructure:"REDIS_ADDR"` // "localhost:6379"
	RedisPass string `mapstructure:"REDIS_PASS"`
	RedisDB   int    `mapstructure:"REDIS_DB"`

	// S3
	S3Endpoint      string `mapstructure:"S3_ENDPOINT"` // http://minio:9000
	S3Region        string `mapstructure:"S3_REGION"`   // us-east-1
	S3AccessKey     string `mapstructure:"S3_ACCESS_KEY"`
	S3SecretKey     string `mapstructure:"S3_SECRET_KEY"`
	S3Bucket        string `mapstructure:"S3_BUCKET"`
	S3UsePathStyle  bool   `mapstructure:"S3_USE_PATH_STYLE"`  // true
	S3PublicBaseURL string `mapstructure:"S3_PUBLIC_BASE_URL"` // http://localhost:9000/mybucket
}

func LoadConfig() (*Config, error) {
	// Viper
	v := viper.New()
	v.SetConfigFile(".env")
	v.AddConfigPath(".")
	v.AddConfigPath("./cmd/api")
	v.SetConfigType("env")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Varsayılanlar
	v.SetDefault("APP_ENV", "development")
	v.SetDefault("APP_PORT", "8080")
	v.SetDefault("LOG_LEVEL", "info")
	v.SetDefault("SHUTDOWN_WAIT", "5s")

	// JWT
	v.SetDefault("JWT_ACCESS_TTL", "15m")
	v.SetDefault("REFRESH_TTL", "720h")
	v.SetDefault("LOGIN_RATE", 5)
	v.SetDefault("REFRESH_RATE", 5)
	v.SetDefault("RATE_WINDOW", "1m")

	// Redis
	v.SetDefault("REDIS_ADDR", "localhost:6379")
	v.SetDefault("REDIS_DB", 0)

	// S3
	// === S3 / MinIO ===
	v.SetDefault("S3_ENDPOINT", "http://localhost:9000")
	v.SetDefault("S3_REGION", "us-east-1")
	v.SetDefault("S3_ACCESS_KEY", "admin")
	v.SetDefault("S3_SECRET_KEY", "Minio.adem54")
	v.SetDefault("S3_BUCKET", "primeapi")
	v.SetDefault("S3_USE_PATH_STYLE", true)
	v.SetDefault("S3_PUBLIC_BASE_URL", "http://localhost:9000/primeapi")

	_ = v.ReadInConfig() // dosya yoksa sorun değil

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
