package db

import (
	"context"
	"fmt"
	"time"

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

// Ping verifies the underlying database connection is reachable.
func Ping(ctx context.Context, gormDB *gorm.DB) error {
	sqlDB, err := gormDB.DB()
	if err != nil {
		return fmt.Errorf("get sql db: %w", err)
	}

	pingCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	if err := sqlDB.PingContext(pingCtx); err != nil {
		return fmt.Errorf("ping db: %w", err)
	}
	return nil
}
