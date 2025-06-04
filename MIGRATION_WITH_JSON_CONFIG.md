# Database Migrations with JSON Configuration

This guide shows how to use Goose database migrations with your existing JSON configuration format. No local `goose.yaml` file is needed - everything is configured through your project's JSON config file.

## Overview

The migration system integrates seamlessly with your existing configuration format:

```json
{
  "db": [
    {
      "name": "default",
      "dialog": "mysql", 
      "args": ["user:password@tcp(localhost:3306)/database?parseTime=true"],
      "logMode": true
    },
    {
      "name": "analytics",
      "dialog": "mysql",
      "args": ["user:password@tcp(localhost:3306)/analytics?parseTime=true"]
    }
  ]
}
```

## Important: Query API Updates

✅ **Query API is now GORM-free!** The Query API has been completely updated to remove GORM dependencies:

- **No more GORM tags required** - The Query API no longer depends on `gorm:"many2many:..."` or `gorm:"polymorphic:..."` tags
- **Explicit join configuration** - Relationships are now configured explicitly using a `JoinRegistry`
- **Better performance** - Uses proper SQL JOINs instead of subqueries when configured
- **More control** - You can specify exactly how tables should be joined

### Simple Queries (No Changes)
```go
// Simple queries work exactly the same
query := &qapi.Query{
    Filter: []qapi.Filter{
        {Name: "name", Value: "John", Operation: qapi.EQ},
    },
}
result, err := queryapi.GenerateSQL(query, User{})
```

### Relationship Queries (New System)
```go
// Set up explicit joins for relationships
joinRegistry := queryapi.NewJoinRegistry()
joinRegistry.Register("Profile", queryapi.JoinConfig{
    Type:       queryapi.OneToOne,
    Table:      "profiles", 
    LocalKey:   "id",
    ForeignKey: "user_id",
})

// Use with QueryOptions
options := &queryapi.QueryOptions{JoinRegistry: joinRegistry}
result, err := queryapi.GenerateSQLWithOptions(query, User{}, options)
```

📖 **See `QUERY_API_JOINS.md` for complete documentation** on the new join system.

## Quick Start

### 1. Install the Migration CLI

In your project that uses sincap-common:

```bash
go install github.com/filllabs/sincap-common/cmd/migrate@latest
```

Or build it locally:

```bash
cd /path/to/sincap-common
go build -o migrate cmd/migrate/main.go
```

### 2. Create Migrations Directory

```bash
mkdir migrations
```

### 3. Create Your First Migration

```bash
# Create a SQL migration
./migrate -command create -name create_users_table

# Create a Go migration (for complex data transformations)
./migrate -command create -name seed_initial_data -type go
```

### 4. Run Migrations

```bash
# Run migrations using default config.json and default database
./migrate

# Run migrations with custom config file
./migrate -config production.json

# Run migrations for specific database
./migrate -db analytics
```

## CLI Commands

### Basic Usage

```bash
migrate [options]
```

### Options

- `-config string`: Path to JSON configuration file (default: `config.json`)
- `-migrations string`: Path to migrations directory (default: `./migrations`)
- `-db string`: Database name from config (default: `default`)
- `-command string`: Migration command: `up`, `down`, `status`, `create` (default: `up`)
- `-name string`: Migration name (required for create command)
- `-type string`: Migration type: `sql` or `go` (default: `sql`)
- `-help`: Show help message

### Commands

#### `up` - Run Migrations
```bash
# Run all pending migrations
./migrate -command up

# Run migrations for specific database
./migrate -command up -db analytics
```

#### `down` - Rollback Migration
```bash
# Rollback the last migration
./migrate -command down

# Rollback for specific database
./migrate -command down -db analytics
```

#### `status` - Check Migration Status
```bash
# Check migration status
./migrate -command status

# Check status for specific database
./migrate -command status -db analytics
```

#### `create` - Create New Migration
```bash
# Create SQL migration
./migrate -command create -name add_user_email_index

# Create Go migration
./migrate -command create -name migrate_user_data -type go
```

## Migration Files

### SQL Migrations

SQL migrations are stored in `migrations/` directory with timestamp prefixes:

```sql
-- migrations/20231201120000_create_users_table.sql
-- +goose Up
CREATE TABLE users (
    id INT AUTO_INCREMENT PRIMARY KEY,
    email VARCHAR(255) NOT NULL UNIQUE,
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- +goose Down
DROP TABLE users;
```

### Go Migrations

Go migrations allow complex data transformations:

```go
// migrations/20231201120001_seed_initial_data.go
package migrations

import (
    "database/sql"
    "github.com/pressly/goose/v3"
)

func init() {
    goose.AddMigration(upSeedInitialData, downSeedInitialData)
}

func upSeedInitialData(tx *sql.Tx) error {
    // Complex data migration logic here
    _, err := tx.Exec("INSERT INTO users (email, name) VALUES (?, ?)", "admin@example.com", "Admin User")
    return err
}

func downSeedInitialData(tx *sql.Tx) error {
    _, err := tx.Exec("DELETE FROM users WHERE email = ?", "admin@example.com")
    return err
}
```

## Programmatic Usage

You can also use the migration functions directly in your Go code:

```go
package main

import (
    "github.com/filllabs/sincap-common/config"
    "github.com/filllabs/sincap-common/db/migration"
)

func main() {
    // Load your config
    cfg := config.LoadConfig("config.json")
    
    // Run migrations for all databases
    if err := migration.RunMigrations("./migrations", cfg.DB); err != nil {
        log.Fatal(err)
    }
    
    // Or run for specific database
    for _, dbConfig := range cfg.DB {
        if dbConfig.Name == "default" {
            if err := migration.RunMigrationForDB("./migrations", dbConfig); err != nil {
                log.Fatal(err)
            }
            break
        }
    }
}
```

## Integration with Makefile

Add these targets to your project's Makefile:

```makefile
# Migration commands
.PHONY: migrate-up migrate-down migrate-status migrate-create

migrate-up: ## Run database migrations
	./migrate -command up

migrate-down: ## Rollback last migration
	./migrate -command down

migrate-status: ## Check migration status
	./migrate -command status

migrate-create: ## Create new migration (usage: make migrate-create NAME=migration_name)
	./migrate -command create -name $(NAME)

migrate-create-go: ## Create new Go migration (usage: make migrate-create-go NAME=migration_name)
	./migrate -command create -name $(NAME) -type go

# Environment-specific migrations
migrate-prod: ## Run migrations on production
	./migrate -config production.json -command up

migrate-staging: ## Run migrations on staging
	./migrate -config staging.json -command up
```

## Configuration Examples

### Single Database
```json
{
  "db": [
    {
      "name": "default",
      "dialog": "mysql",
      "args": ["user:password@tcp(localhost:3306)/myapp?parseTime=true"]
    }
  ]
}
```

### Multiple Databases
```json
{
  "db": [
    {
      "name": "main",
      "dialog": "mysql",
      "args": ["user:password@tcp(localhost:3306)/myapp?parseTime=true"]
    },
    {
      "name": "analytics",
      "dialog": "mysql", 
      "args": ["user:password@tcp(localhost:3306)/analytics?parseTime=true"]
    },
    {
      "name": "logs",
      "dialog": "mysql",
      "args": ["user:password@tcp(localhost:3306)/logs?parseTime=true"]
    }
  ]
}
```

### Environment-Specific Configs

**development.json**
```json
{
  "db": [
    {
      "name": "default",
      "dialog": "mysql",
      "args": ["root:@tcp(localhost:3306)/myapp_dev?parseTime=true"]
    }
  ]
}
```

**production.json**
```json
{
  "db": [
    {
      "name": "default", 
      "dialog": "mysql",
      "args": ["prod_user:secure_password@tcp(prod-db:3306)/myapp?parseTime=true"]
    }
  ]
}
```

## Best Practices

### 1. Migration Naming
- Use descriptive names: `create_users_table`, `add_email_index`, `migrate_user_preferences`
- Use timestamps for ordering (automatically added by Goose)

### 2. Migration Content
- Always include both `Up` and `Down` migrations
- Test rollbacks in development
- Keep migrations small and focused
- Use transactions for data migrations

### 3. Environment Management
- Use separate config files for each environment
- Never run migrations directly on production without testing
- Always backup before running migrations on production

### 4. Team Collaboration
- Commit migration files to version control
- Use descriptive commit messages for migrations
- Coordinate with team when creating migrations that affect shared tables

## Troubleshooting

### Common Issues

1. **"Database not found in configuration"**
   - Check that the database name matches exactly in your config
   - Verify the config file path is correct

2. **"Connection refused"**
   - Verify database server is running
   - Check connection string format
   - Ensure database exists

3. **"Migration failed"**
   - Check migration syntax
   - Verify database permissions
   - Look at the specific error message

### Getting Help

```bash
# Show help
./migrate -help

# Check migration status to see what's applied
./migrate -command status
```

## Migration from GORM AutoMigrate

If you were previously using GORM's AutoMigrate, you'll need to:

1. **Create initial migration** from your current schema:
   ```bash
   ./migrate -command create -name initial_schema
   ```

2. **Export current schema** and add it to the migration file

3. **Mark as applied** (if database already has the schema):
   ```sql
   INSERT INTO goose_db_version (version_id, is_applied) VALUES (20231201120000, 1);
   ```

4. **Remove AutoMigrate calls** from your code

This migration system gives you full control over your database schema changes while maintaining compatibility with your existing JSON configuration format. 