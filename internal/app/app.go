package app

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type App struct {
	logger *zap.Logger
	db     *gorm.DB
	cfg    *Config
	redis  *redis.Client
}

func New(logger *zap.Logger, db *gorm.DB, cfg *Config, rc *redis.Client) (*App, error) {
	if logger == nil || db == nil || cfg == nil {
		return nil, fmt.Errorf("logger, db ve cfg zorunludur")
	}
	return &App{logger: logger, db: db, cfg: cfg, redis: rc}, nil
}

func (a *App) Logger() *zap.Logger  { return a.logger }
func (a *App) DB() *gorm.DB         { return a.db }
func (a *App) Config() *Config      { return a.cfg }
func (a *App) Redis() *redis.Client { return a.redis }

func (a *App) Close(ctx context.Context) error {
	var firstErr error

	// GORM -> *sql.DB
	if a.db != nil {
		if sqlDB, err := a.db.DB(); err == nil && sqlDB != nil {
			done := make(chan struct{})
			go func(db *sql.DB) {
				_ = db.Close()
				close(done)
			}(sqlDB)

			select {
			case <-done:
			case <-ctx.Done():
				if firstErr == nil {
					firstErr = ctx.Err()
				}
			}
		} else if firstErr == nil && err != nil {
			firstErr = err
		}
	}

	if a.redis != nil {
		_ = a.redis.Close()
	}

	// Logger flush
	if a.logger != nil {
		_ = a.logger.Sync()
	}
	return firstErr
}
