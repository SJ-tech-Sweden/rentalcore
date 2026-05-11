// Package repository provides PostgreSQL database connection for RentalCore
package repository

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"go-barcode-webapp/internal/config"
	"go-barcode-webapp/internal/migrations"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

// Database wraps the GORM database connection
type Database struct {
	*gorm.DB
}

// startupMigrationsLockKey is a fixed, repository-scoped advisory lock key used
// only for startup migration/seed execution to prevent concurrent runners.
const startupMigrationsLockKey int64 = 73043001

// NewDatabase erstellt eine neue PostgreSQL-Datenbankverbindung
func NewDatabase(cfg *config.DatabaseConfig) (*Database, error) {
	dsn := cfg.DSN()

	var logLevel logger.LogLevel
	if cfg.EnableQueryLogging {
		logLevel = logger.Info
	} else {
		logLevel = logger.Warn
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger:                 logger.Default.LogMode(logLevel),
		PrepareStmt:            cfg.PrepareStmt,
		SkipDefaultTransaction: false,
		CreateBatchSize:        1000,
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
		DisableForeignKeyConstraintWhenMigrating: cfg.DisableForeignKeyConstraintWhenMigrating,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get sql.DB: %w", err)
	}

	// PostgreSQL unterstützt viele parallele Verbindungen
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(time.Hour)
	sqlDB.SetConnMaxIdleTime(30 * time.Minute)

	log.Printf("PostgreSQL database connected: %s:%d/%s", cfg.Host, cfg.Port, cfg.Name)

	// Optionally run SQL migrations and seeds on startup. Controlled by
	// env var MIGRATE_ON_STARTUP (default: "false"). The migrations directory
	// can be overridden with MIGRATIONS_DIR (default: "migrations").
	if os.Getenv("MIGRATE_ON_STARTUP") == "true" {
		migrationsDir := os.Getenv("MIGRATIONS_DIR")
		if migrationsDir == "" {
			migrationsDir = "migrations"
		}
		if _, err := sqlDB.Exec("SELECT pg_advisory_lock($1)", startupMigrationsLockKey); err != nil {
			return nil, fmt.Errorf("acquire startup migration lock: %w", err)
		}
		defer func() {
			if _, err := sqlDB.Exec("SELECT pg_advisory_unlock($1)", startupMigrationsLockKey); err != nil {
				log.Printf("WARNING: failed to release startup migration lock: %v", err)
			}
		}()

		absDir, _ := filepath.Abs(migrationsDir)
		log.Printf("Running SQL migrations from %s", absDir)
		if err := migrations.ApplyMigrations(sqlDB, migrationsDir); err != nil {
			return nil, fmt.Errorf("apply migrations: %w", err)
		}
		if err := migrations.ApplySeeds(sqlDB, filepath.Join(migrationsDir, "seeds")); err != nil {
			return nil, fmt.Errorf("apply seeds: %w", err)
		}
		log.Println("Migrations and startup seeds applied")
	}
	return &Database{db}, nil
}

// Close schließt die Datenbankverbindung
func (db *Database) Close() error {
	sqlDB, err := db.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// Ping testet die Datenbankverbindung
func (db *Database) Ping() error {
	sqlDB, err := db.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Ping()
}
