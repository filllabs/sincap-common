# Database Migration Guide with Goose

This guide explains how to use Goose for database migrations, replacing GORM's AutoMigrate functionality.

## Why Goose?

After migrating from GORM to sqlx, we chose **Goose** over golang-migrate because:

- ✅ **Supports both SQL and Go migrations** - Perfect for complex data transformations
- ✅ **Timestamp-based versioning** - Prevents conflicts when multiple developers create migrations
- ✅ **Easier migration from GORM** - Can handle both schema and data changes
- ✅ **Simpler setup and configuration**
- ✅ **Active community and good documentation**

## Installation

### Install Goose CLI

```bash
# Install globally
go install github.com/pressly/goose/v3/cmd/goose@latest

# Or use our Makefile
make install-goose
```

### Verify Installation

```bash
goose version
```

## Configuration

### Environment Variables

Set these environment variables for your database:

```bash
export GOOSE_DRIVER=mysql
export GOOSE_DBSTRING="user:password@tcp(localhost:3306)/database?parseTime=true"
```

### Configuration File

We've included a `goose.yaml` configuration file. You can customize it for your environment:

```yaml
driver: mysql
dbstring: "user:password@tcp(localhost:3306)/database?parseTime=true"
dir: "./migrations"
table: "goose_db_version"
verbose: false
allow_missing: false
```

## Migration Workflow

### 1. Creating Migrations

#### SQL Migrations (Most Common)

```bash
# Using Makefile
make create-migration NAME=create_users_table

# Using Goose directly
goose -dir ./migrations create create_users_table sql
```

This creates a file like `20231201120000_create_users_table.sql`:

```sql
-- +goose Up
CREATE TABLE users (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

-- +goose Down
DROP TABLE users;
```

#### Go Migrations (For Complex Data Operations)

```bash
# Using Makefile
make create-go-migration NAME=migrate_user_data

# Using Goose directly
goose -dir ./migrations create migrate_user_data go
```

This creates a Go file for complex data transformations:

```go
package migrations

import (
    "context"
    "database/sql"
    "github.com/pressly/goose/v3"
)

func init() {
    goose.AddMigrationContext(upMigrateUserData, downMigrateUserData)
}

func upMigrateUserData(ctx context.Context, tx *sql.Tx) error {
    // Complex data migration logic here
    _, err := tx.ExecContext(ctx, `
        UPDATE users 
        SET status = 'active' 
        WHERE created_at > DATE_SUB(NOW(), INTERVAL 30 DAY)
    `)
    return err
}

func downMigrateUserData(ctx context.Context, tx *sql.Tx) error {
    // Rollback logic
    _, err := tx.ExecContext(ctx, `
        UPDATE users 
        SET status = NULL 
        WHERE status = 'active'
    `)
    return err
}
```

### 2. Applying Migrations

```bash
# Apply all pending migrations
make migrate-up

# Or using Goose directly
goose -dir ./migrations up
```

### 3. Rolling Back Migrations

```bash
# Rollback last migration
make migrate-down

# Rollback to specific version
make migrate-down-to VERSION=20231201120000

# Reset all migrations (careful!)
make migrate-reset
```

### 4. Checking Migration Status

```bash
# Show current status
make migrate-status

# Show current version
make migrate-version

# Validate migrations
make migrate-validate
```

## Migration Best Practices

### 1. **Always Write Rollback Scripts**

Every migration should have a corresponding `-- +goose Down` section:

```sql
-- +goose Up
ALTER TABLE users ADD COLUMN phone VARCHAR(20);

-- +goose Down
ALTER TABLE users DROP COLUMN phone;
```

### 2. **Use Transactions for Data Safety**

Goose automatically wraps migrations in transactions, but be aware of DDL limitations:

```sql
-- +goose Up
-- +goose StatementBegin
CREATE TABLE temp_users AS SELECT * FROM users WHERE active = 1;
DROP TABLE users;
RENAME TABLE temp_users TO users;
-- +goose StatementEnd

-- +goose Down
-- Rollback logic here
```

### 3. **Test Migrations Thoroughly**

Always test migrations on a copy of production data:

```bash
# Create test database
mysql -e "CREATE DATABASE test_db;"

# Test migration
GOOSE_DBSTRING="user:pass@tcp(localhost:3306)/test_db?parseTime=true" make migrate-up

# Test rollback
GOOSE_DBSTRING="user:pass@tcp(localhost:3306)/test_db?parseTime=true" make migrate-down
```

### 4. **Handle Large Tables Carefully**

For large tables, consider:

- Using `ALGORITHM=INPLACE` for MySQL
- Adding indexes concurrently
- Batching data updates

```sql
-- +goose Up
-- For large tables, use ALGORITHM=INPLACE when possible
ALTER TABLE large_table 
ADD COLUMN new_column VARCHAR(255),
ALGORITHM=INPLACE, LOCK=NONE;

-- +goose Down
ALTER TABLE large_table 
DROP COLUMN new_column,
ALGORITHM=INPLACE, LOCK=NONE;
```

## Common Migration Patterns

### 1. **Adding a Column**

```sql
-- +goose Up
ALTER TABLE users ADD COLUMN phone VARCHAR(20);

-- +goose Down
ALTER TABLE users DROP COLUMN phone;
```

### 2. **Creating an Index**

```sql
-- +goose Up
CREATE INDEX idx_users_email ON users(email);

-- +goose Down
DROP INDEX idx_users_email ON users;
```

### 3. **Modifying Column Type**

```sql
-- +goose Up
ALTER TABLE users MODIFY COLUMN phone VARCHAR(30);

-- +goose Down
ALTER TABLE users MODIFY COLUMN phone VARCHAR(20);
```

### 4. **Data Migration with Go**

For complex data transformations, use Go migrations:

```go
func upMigrateUserData(ctx context.Context, tx *sql.Tx) error {
    rows, err := tx.QueryContext(ctx, "SELECT id, old_field FROM users")
    if err != nil {
        return err
    }
    defer rows.Close()

    for rows.Next() {
        var id int
        var oldField string
        if err := rows.Scan(&id, &oldField); err != nil {
            return err
        }

        // Transform data
        newValue := transformData(oldField)

        _, err := tx.ExecContext(ctx, 
            "UPDATE users SET new_field = ? WHERE id = ?", 
            newValue, id)
        if err != nil {
            return err
        }
    }
    return rows.Err()
}
```

## CI/CD Integration

### GitHub Actions Example

```yaml
name: Database Migration
on:
  push:
    branches: [main]

jobs:
  migrate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Setup Go
        uses: actions/setup-go@v3
        with:
          go-version: '1.18'
          
      - name: Install Goose
        run: go install github.com/pressly/goose/v3/cmd/goose@latest
        
      - name: Run Migrations
        env:
          GOOSE_DRIVER: mysql
          GOOSE_DBSTRING: ${{ secrets.DATABASE_URL }}
        run: goose -dir ./migrations up
```

### Docker Integration

```dockerfile
FROM golang:1.18-alpine

# Install Goose
RUN go install github.com/pressly/goose/v3/cmd/goose@latest

COPY migrations/ /migrations/
COPY migrate.sh /migrate.sh

CMD ["/migrate.sh"]
```

## Troubleshooting

### Common Issues

1. **Migration Out of Order**
   ```bash
   make migrate-fix
   ```

2. **Failed Migration**
   ```bash
   # Check status
   make migrate-status
   
   # Fix the migration file and retry
   make migrate-up
   ```

3. **Rollback Issues**
   ```bash
   # Check what went wrong
   make migrate-status
   
   # Manual intervention may be needed
   mysql -e "DELETE FROM goose_db_version WHERE version_id = 'problematic_version';"
   ```

### Debugging

Enable verbose output:

```bash
# Set verbose in goose.yaml or use environment variable
export GOOSE_VERBOSE=true
make migrate-up
```

## Migrating from GORM AutoMigrate

### Step 1: Create Initial Schema Migration

If you have existing tables from GORM, create an initial migration:

```bash
make create-initial-migration
```

Edit the generated file to match your current schema.

### Step 2: Mark as Applied

If the schema already exists in production:

```bash
# Apply only the migration record, not the SQL
goose -dir ./migrations up-by-one --no-versioning
```

### Step 3: Future Changes

From now on, all schema changes should be done through migrations:

```bash
# Instead of changing your struct and running AutoMigrate
# Create a migration
make create-migration NAME=add_user_status_column
```

## Summary

Goose provides a robust, version-controlled approach to database schema management that replaces GORM's AutoMigrate functionality. Key benefits:

- ✅ **Explicit control** over schema changes
- ✅ **Version tracking** and rollback capabilities  
- ✅ **Team collaboration** without conflicts
- ✅ **Production safety** with tested migrations
- ✅ **CI/CD integration** for automated deployments

For more information, see the [official Goose documentation](https://github.com/pressly/goose). 