# Migrations Directory

This directory contains database migration files for the sincap-common project.

## File Naming Convention

Goose uses timestamp-based naming:
- `YYYYMMDDHHMMSS_migration_name.sql` for SQL migrations
- `YYYYMMDDHHMMSS_migration_name.go` for Go migrations

## Example Files

### SQL Migration (`20231201120000_example_users_table.sql`)

This is the most common type of migration for schema changes:

```sql
-- +goose Up
CREATE TABLE users (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- +goose Down
DROP TABLE users;
```

### Go Migration (for complex data operations)

For complex data transformations, create a Go migration:

```bash
# Create a Go migration
make create-go-migration NAME=migrate_user_data
```

This creates a file like `20231201130000_migrate_user_data.go`:

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
        newValue := processData(oldField)

        _, err := tx.ExecContext(ctx, 
            "UPDATE users SET new_field = ? WHERE id = ?", 
            newValue, id)
        if err != nil {
            return err
        }
    }
    return rows.Err()
}

func downMigrateUserData(ctx context.Context, tx *sql.Tx) error {
    // Rollback logic
    _, err := tx.ExecContext(ctx, "UPDATE users SET new_field = NULL")
    return err
}
```

## Usage

### Creating Migrations

```bash
# SQL migration
make create-migration NAME=add_user_phone

# Go migration  
make create-go-migration NAME=migrate_user_data
```

### Applying Migrations

```bash
# Apply all pending migrations
make migrate-up

# Check status
make migrate-status
```

### Rolling Back

```bash
# Rollback last migration
make migrate-down

# Rollback to specific version
make migrate-down-to VERSION=20231201120000
```

## Best Practices

1. **Always include rollback logic** in the `-- +goose Down` section
2. **Test migrations** on a copy of production data first
3. **Use Go migrations** for complex data transformations
4. **Keep migrations small** and focused on a single change
5. **Never modify** existing migration files after they've been applied

## Migration States

- **Pending**: Migration file exists but hasn't been applied
- **Applied**: Migration has been successfully applied to the database
- **Failed**: Migration failed during application (needs manual intervention)

Check the status with:
```bash
make migrate-status
``` 