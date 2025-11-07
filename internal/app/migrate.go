package app

import "gorm.io/gorm"

func AutoMigrate(gdb *gorm.DB, models ...interface{}) error {
	return gdb.AutoMigrate(models...)
}
