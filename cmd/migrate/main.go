package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/filllabs/sincap-common/config"
	"github.com/filllabs/sincap-common/db"
	"github.com/filllabs/sincap-common/db/migration"
)

func main() {
	var (
		configFile    = flag.String("config", "config.json", "Path to JSON configuration file")
		migrationsDir = flag.String("migrations", "./migrations", "Path to migrations directory")
		dbName        = flag.String("db", "default", "Database name from config")
		command       = flag.String("command", "up", "Migration command: up, down, status, create")
		migrationName = flag.String("name", "", "Migration name (for create command)")
		migrationType = flag.String("type", "sql", "Migration type: sql or go (for create command)")
		help          = flag.Bool("help", false, "Show help")
	)
	flag.Parse()

	if *help {
		showHelp()
		return
	}

	// Load configuration
	var cfg config.Config
	if err := loadConfig(*configFile, &cfg); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Find the specified database configuration
	var selectedDB *db.Config
	for _, dbConf := range cfg.DB {
		if dbConf.Name == *dbName {
			selectedDB = &dbConf
			break
		}
	}

	if selectedDB == nil {
		log.Fatalf("Database '%s' not found in configuration", *dbName)
	}

	// Execute command
	switch *command {
	case "up":
		if err := migration.RunMigrationForDB(*migrationsDir, *selectedDB); err != nil {
			log.Fatalf("Migration failed: %v", err)
		}
		fmt.Printf("✓ Migrations completed successfully for database '%s'\n", *dbName)

	case "down":
		if err := migration.RollbackMigration(*migrationsDir, *selectedDB); err != nil {
			log.Fatalf("Rollback failed: %v", err)
		}
		fmt.Printf("✓ Rollback completed for database '%s'\n", *dbName)

	case "status":
		if err := migration.GetMigrationStatus(*migrationsDir, *selectedDB); err != nil {
			log.Fatalf("Status check failed: %v", err)
		}

	case "create":
		if *migrationName == "" {
			log.Fatal("Migration name is required for create command")
		}
		if err := migration.CreateMigration(*migrationsDir, *migrationName, *migrationType); err != nil {
			log.Fatalf("Failed to create migration: %v", err)
		}
		fmt.Printf("✓ Created %s migration: %s\n", *migrationType, *migrationName)

	default:
		log.Fatalf("Unknown command: %s", *command)
	}
}

func loadConfig(path string, cfg *config.Config) error {
	// Check if file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("config file not found: %s", path)
	}

	// Read file
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse JSON
	if err := json.Unmarshal(data, cfg); err != nil {
		return fmt.Errorf("failed to parse config file: %w", err)
	}

	return nil
}

func showHelp() {
	fmt.Println("Sincap-Common Migration Tool")
	fmt.Println("Runs database migrations using your JSON configuration")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  migrate [options]")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  -config string     Path to JSON configuration file (default: config.json)")
	fmt.Println("  -migrations string Path to migrations directory (default: ./migrations)")
	fmt.Println("  -db string         Database name from config (default: default)")
	fmt.Println("  -command string    Migration command: up, down, status, create (default: up)")
	fmt.Println("  -name string       Migration name (required for create command)")
	fmt.Println("  -type string       Migration type: sql or go (default: sql)")
	fmt.Println("  -help              Show this help message")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  up      Run all pending migrations")
	fmt.Println("  down    Rollback the last migration")
	fmt.Println("  status  Show migration status")
	fmt.Println("  create  Create a new migration file")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  # Run migrations using default config")
	fmt.Println("  migrate")
	fmt.Println()
	fmt.Println("  # Run migrations with custom config")
	fmt.Println("  migrate -config production.json")
	fmt.Println()
	fmt.Println("  # Create a new SQL migration")
	fmt.Println("  migrate -command create -name create_users_table")
	fmt.Println()
	fmt.Println("  # Create a new Go migration")
	fmt.Println("  migrate -command create -name seed_data -type go")
	fmt.Println()
	fmt.Println("  # Check migration status for specific database")
	fmt.Println("  migrate -command status -db analytics")
	fmt.Println()
	fmt.Println("  # Rollback last migration")
	fmt.Println("  migrate -command down")
	fmt.Println()
	fmt.Println("Your JSON config should include a 'db' array like:")
	fmt.Println(`  {
    "db": [
      {
        "name": "default",
        "dialog": "mysql",
        "args": ["user:password@tcp(localhost:3306)/database?parseTime=true"]
      }
    ]
  }`)
}
