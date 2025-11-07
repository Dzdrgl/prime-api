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

	JWTSecret     string        `mapstructure:"JWT_SECRET"`
	JWTIssuer     string        `mapstructure:"JWT_ISSUER"`
	JWTAudience   string        `mapstructure:"JWT_AUDIENCE"`
	JWTAccessTTL  time.Duration `mapstructure:"JWT_ACCESS_TTL"` // 15m
	RefreshTTL    time.Duration `mapstructure:"REFRESH_TTL"`    // 720h (30d) öneri
	RefreshPepper string        `mapstructure:"REFRESH_PEPPER"` // server-side secret for hashing
	LoginRate     int           `mapstructure:"LOGIN_RATE"`     // 5
	RefreshRate   int           `mapstructure:"REFRESH_RATE"`   // 5
	RateWindow    time.Duration `mapstructure:"RATE_WINDOW"`    // 1m

	RedisAddr string `mapstructure:"REDIS_ADDR"` // "localhost:6379"
	RedisPass string `mapstructure:"REDIS_PASS"`
	RedisDB   int    `mapstructure:"REDIS_DB"`
}

func LoadConfig() (*Config, error) {
	v := viper.New()
	v.SetConfigName(".env")
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

	_ = v.ReadInConfig() // dosya yoksa sorun değil

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
