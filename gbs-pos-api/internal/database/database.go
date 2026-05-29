package database

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func Connect(databaseURL string, logLevel string) (*gorm.DB, error) {
	var level logger.LogLevel
	switch logLevel {
	case "silent":
		level = logger.Silent
	case "error":
		level = logger.Error
	case "warn":
		level = logger.Warn
	default:
		level = logger.Info
	}

	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold: time.Second,
			LogLevel:      level,
			Colorful:      true,
		},
	)

	db, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{
		Logger: newLogger,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetConnMaxLifetime(time.Hour)

	return db, nil
}

// ResolveMigrationsPath finds the migrations directory (absolute path for file:// source).
func ResolveMigrationsPath(path string) (string, error) {
	candidates := []string{path}
	if path == "" {
		candidates = []string{"migrations", "../migrations", "/app/migrations"}
	}
	for _, p := range candidates {
		if p == "" {
			continue
		}
		abs, err := filepath.Abs(p)
		if err != nil {
			continue
		}
		info, err := os.Stat(abs)
		if err == nil && info.IsDir() {
			return filepath.ToSlash(abs), nil
		}
	}
	return "", fmt.Errorf("migrations directory not found (MIGRATIONS_PATH=%q)", path)
}

func RunMigrations(databaseURL string, migrationsPath string) error {
	dir, err := ResolveMigrationsPath(migrationsPath)
	if err != nil {
		return err
	}
	m, err := migrate.New(
		"file://"+dir,
		databaseURL,
	)
	if err != nil {
		return fmt.Errorf("failed to create migration instance: %w", err)
	}
	defer m.Close()
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations: %w", err)
	}
	return nil
}
