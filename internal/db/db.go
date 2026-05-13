package db

import (
	"inventory/internal/model"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Connect opens a GORM DB connection and runs auto-migration for core models.
func Connect(dsn string) (*gorm.DB, error) {
	gormDB, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// Auto-migrate core models
	if err := gormDB.AutoMigrate(&model.Product{}); err != nil {
		return nil, err
	}

	return gormDB, nil
}
