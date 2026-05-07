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
	// env var MIGRATE_ON_STARTUP (default: "true"). The migrations directory
	// can be overridden with MIGRATIONS_DIR (default: "migrations").
	if os.Getenv("MIGRATE_ON_STARTUP") != "false" {
		migrationsDir := os.Getenv("MIGRATIONS_DIR")
		if migrationsDir == "" {
			migrationsDir = "migrations"
		}
		absDir, _ := filepath.Abs(migrationsDir)
		log.Printf("Running SQL migrations from %s", absDir)
		if err := migrations.ApplyMigrations(sqlDB, migrationsDir); err != nil {
			return nil, fmt.Errorf("apply migrations: %w", err)
		}
		// Apply a single idempotent seed file if present.
		seedFile := filepath.Join(migrationsDir, "064_default_seeds.sql")
		if _, err := os.Stat(seedFile); err == nil {
			b, err := os.ReadFile(seedFile)
			if err != nil {
				return nil, fmt.Errorf("read seed file: %w", err)
			}
			if len(b) > 0 {
				if _, err := sqlDB.Exec(string(b)); err != nil {
					return nil, fmt.Errorf("exec seed SQL: %w", err)
				}
				log.Printf("Applied seed SQL from %s", seedFile)
			}
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
