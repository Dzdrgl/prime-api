package app

import (
	"fmt"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type Option func(*App) error

func WithDB(db *gorm.DB) Option {
	return func(a *App) error {
		if db == nil {
			return fmt.Errorf("nil db")
		}
		a.db = db
		return nil
	}
}

func WithLogger(l *zap.Logger) Option {
	return func(a *App) error {
		if l == nil {
			return fmt.Errorf("nil logger")
		}
		a.logger = l
		return nil
	}
}

func NewApp(opts ...Option) (*App, error) {
	a := &App{}
	for _, opt := range opts {
		if err := opt(a); err != nil {
			return nil, err
		}
	}
	if a.logger == nil || a.db == nil {
		return nil, fmt.Errorf("logger ve db zorunludur")
	}
	return a, nil
}
