package db

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/linskybing/platform-go/internal/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func createEnums() {
	extensions := []string{
		`CREATE EXTENSION IF NOT EXISTS ltree;`,
		`CREATE EXTENSION IF NOT EXISTS btree_gist;`,
		`CREATE EXTENSION IF NOT EXISTS "uuid-ossp";`,
	}
	for _, ext := range extensions {
		if err := DB.Exec(ext).Error; err != nil {
			slog.Error("failed to create extension", "extension", ext, "error", err)
		}
	}

	enums := []string{
		`DO $$ BEGIN CREATE TYPE user_type AS ENUM ('origin', 'oauth2'); EXCEPTION WHEN duplicate_object THEN null; END $$;`,
		`DO $$ BEGIN CREATE TYPE user_status AS ENUM ('online', 'offline', 'delete'); EXCEPTION WHEN duplicate_object THEN null; END $$;`,
		`DO $$ BEGIN CREATE TYPE user_role AS ENUM ('admin', 'manager', 'user'); EXCEPTION WHEN duplicate_object THEN null; END $$;`,
		`DO $$ BEGIN CREATE TYPE resource_type AS ENUM ('Pod', 'Service', 'Deployment', 'ConfigMap', 'Ingress', 'Job'); EXCEPTION WHEN duplicate_object THEN null; END $$;`,
	}

	for _, enum := range enums {
		if err := DB.Exec(enum).Error; err != nil {
			slog.Error("failed to create enum", "enum", enum, "error", err)
		}
	}
}

func EnsureConstraints() {
	// Add EXCLUDE constraint for resource_plans
	constraintSQL := `
	ALTER TABLE resource_plans 
	ADD CONSTRAINT exclude_overlapping_resource_plans 
	EXCLUDE USING GIST (project_id WITH =, week_window WITH &&);
	`
	// Use a check to avoid failing if constraint already exists
	checkSQL := `SELECT 1 FROM pg_constraint WHERE conname = 'exclude_overlapping_resource_plans'`
	var exists int
	DB.Raw(checkSQL).Scan(&exists)
	if exists == 0 {
		if err := DB.Exec(constraintSQL).Error; err != nil {
			slog.Warn("failed to create exclusion constraint", "error", err)
		}
	}
}

func Init() {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		config.DbHost,
		config.DbPort,
		config.DbUser,
		config.DbPassword,
		config.DbName,
		config.DbSSLMode,
	)

	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		slog.Error("failed to connect to database",
			"host", config.DbHost,
			"port", config.DbPort,
			"dbname", config.DbName,
			"error", err)
		panic(fmt.Sprintf("failed to connect to DB: %v", err))
	}

	sqlDB, err := DB.DB()
	if err != nil {
		slog.Error("failed to get sql DB handle",
			"error", err)
		panic(fmt.Sprintf("failed to get sql DB handle: %v", err))
	}

	sqlDB.SetMaxOpenConns(config.DbMaxOpenConns)
	sqlDB.SetMaxIdleConns(config.DbMaxIdleConns)
	sqlDB.SetConnMaxLifetime(time.Duration(config.DbConnMaxLifetimeSeconds) * time.Second)
	sqlDB.SetConnMaxIdleTime(time.Duration(config.DbConnMaxIdleTimeSeconds) * time.Second)

	// Create enums
	createEnums()

	slog.Info("database connected and migrated")
}

func InitWithGormDB(gormDB *gorm.DB) {
	DB = gormDB
}
