package app

import (
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func InitGorm(dsn string) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(25)
	sqlDB.SetConnMaxLifetime(1 * time.Hour)

	if err := sqlDB.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}
