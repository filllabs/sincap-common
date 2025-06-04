package migration

import (
	"database/sql"
	"fmt"

	"github.com/filllabs/sincap-common/db"
	"github.com/filllabs/sincap-common/logging"
	_ "github.com/go-sql-driver/mysql"
	"github.com/pressly/goose/v3"
	"go.uber.org/zap"
)

// MigrationConfig holds migration-specific configuration
type MigrationConfig struct {
	Dir   string `json:"dir,omitempty"`   // Migration files directory (defaults to "./migrations")
	Table string `json:"table,omitempty"` // Migration table name (defaults to "goose_db_version")
}

// buildConnectionString builds a connection string from db.Config
func buildConnectionString(config db.Config) (string, error) {
	if len(config.Args) == 0 {
		return "", fmt.Errorf("no connection string provided for database '%s'", config.Name)
	}
	return config.Args[0], nil
}

// RunMigrations runs database migrations for all configured databases
// migrationsDir: path to directory containing migration files
// configs: database configurations from JSON
func RunMigrations(migrationsDir string, configs []db.Config) error {
	for _, config := range configs {
		logging.Logger.Info("Running migrations for database", zap.String("name", config.Name))

		if err := RunMigrationForDB(migrationsDir, config); err != nil {
			return fmt.Errorf("migration failed for database %s: %w", config.Name, err)
		}
	}

	return nil
}

// RunMigrationForDB runs migrations for a specific database configuration
func RunMigrationForDB(migrationsDir string, config db.Config) error {
	// Build connection string
	dsn, err := buildConnectionString(config)
	if err != nil {
		return fmt.Errorf("failed to build connection string: %w", err)
	}

	// Open database connection
	sqlDB, err := sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer sqlDB.Close()

	// Test connection
	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	// Run migrations
	if err := goose.Up(sqlDB, migrationsDir); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	logging.Logger.Info("Migrations completed successfully", zap.String("database", config.Name))
	return nil
}

// CreateMigration creates a new migration file
func CreateMigration(migrationsDir, name, migrationType string) error {
	// Ensure migrations directory exists
	if err := goose.Fix(migrationsDir); err != nil {
		return fmt.Errorf("failed to create migrations directory: %w", err)
	}

	var err error
	switch migrationType {
	case "go":
		err = goose.Create(nil, migrationsDir, name, "go")
	case "sql":
		err = goose.Create(nil, migrationsDir, name, "sql")
	default:
		return fmt.Errorf("invalid migration type: %s (must be 'sql' or 'go')", migrationType)
	}

	if err != nil {
		return fmt.Errorf("failed to create migration: %w", err)
	}

	logging.Logger.Info("Migration created successfully",
		zap.String("name", name),
		zap.String("type", migrationType),
		zap.String("dir", migrationsDir))

	return nil
}

// GetMigrationStatus returns the current migration status for a database
func GetMigrationStatus(migrationsDir string, config db.Config) error {
	dsn, err := buildConnectionString(config)
	if err != nil {
		return fmt.Errorf("failed to build connection string: %w", err)
	}

	sqlDB, err := sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer sqlDB.Close()

	return goose.Status(sqlDB, migrationsDir)
}

// RollbackMigration rolls back the last migration
func RollbackMigration(migrationsDir string, config db.Config) error {
	dsn, err := buildConnectionString(config)
	if err != nil {
		return fmt.Errorf("failed to build connection string: %w", err)
	}

	sqlDB, err := sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer sqlDB.Close()

	return goose.Down(sqlDB, migrationsDir)
}

// ResetMigrations resets all migrations (USE WITH CAUTION!)
func ResetMigrations(migrationsDir string, config db.Config) error {
	dsn, err := buildConnectionString(config)
	if err != nil {
		return fmt.Errorf("failed to build connection string: %w", err)
	}

	sqlDB, err := sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer sqlDB.Close()

	return goose.Reset(sqlDB, migrationsDir)
}
