package postgres

import (
	"time"

	"github.com/ductongnguyen/vivy-chat/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func New(cfg *config.PostgresConfig) (*gorm.DB, error) {
	dsn := cfg.URI

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifeTime) * time.Second)

	return db, nil
}
