# Makefile for database migrations using Goose

# Database configuration
DB_DRIVER ?= mysql
DB_STRING ?= "user:password@tcp(localhost:3306)/database?parseTime=true"
MIGRATIONS_DIR ?= ./migrations

# Set environment variables for Goose
export GOOSE_DRIVER=$(DB_DRIVER)
export GOOSE_DBSTRING=$(DB_STRING)

# Install goose binary
install-goose:
	@echo "Installing Goose..."
	go install github.com/pressly/goose/v3/cmd/goose@latest

# Create a new SQL migration
create-migration:
	@if [ -z "$(NAME)" ]; then \
		echo "Usage: make create-migration NAME=migration_name"; \
		exit 1; \
	fi
	@echo "Creating new migration: $(NAME)"
	goose -dir $(MIGRATIONS_DIR) create $(NAME) sql

# Create a new Go migration
create-go-migration:
	@if [ -z "$(NAME)" ]; then \
		echo "Usage: make create-go-migration NAME=migration_name"; \
		exit 1; \
	fi
	@echo "Creating new Go migration: $(NAME)"
	goose -dir $(MIGRATIONS_DIR) create $(NAME) go

# Apply all pending migrations
migrate-up:
	@echo "Applying migrations..."
	goose -dir $(MIGRATIONS_DIR) up

# Rollback the last migration
migrate-down:
	@echo "Rolling back last migration..."
	goose -dir $(MIGRATIONS_DIR) down

# Rollback to a specific version
migrate-down-to:
	@if [ -z "$(VERSION)" ]; then \
		echo "Usage: make migrate-down-to VERSION=version_number"; \
		exit 1; \
	fi
	@echo "Rolling back to version $(VERSION)..."
	goose -dir $(MIGRATIONS_DIR) down-to $(VERSION)

# Show migration status
migrate-status:
	@echo "Migration status:"
	goose -dir $(MIGRATIONS_DIR) status

# Validate migrations
migrate-validate:
	@echo "Validating migrations..."
	goose -dir $(MIGRATIONS_DIR) validate

# Reset database (rollback all migrations)
migrate-reset:
	@echo "Resetting database (rolling back all migrations)..."
	goose -dir $(MIGRATIONS_DIR) reset

# Fix migration sequence (useful when migrations are out of order)
migrate-fix:
	@echo "Fixing migration sequence..."
	goose -dir $(MIGRATIONS_DIR) fix

# Show version
migrate-version:
	@echo "Current migration version:"
	goose -dir $(MIGRATIONS_DIR) version

# Create initial migration from existing schema
create-initial-migration:
	@echo "Creating initial migration..."
	@echo "-- +goose Up" > $(MIGRATIONS_DIR)/$(shell date +%Y%m%d%H%M%S)_initial_schema.sql
	@echo "-- Add your existing schema here" >> $(MIGRATIONS_DIR)/$(shell date +%Y%m%d%H%M%S)_initial_schema.sql
	@echo "" >> $(MIGRATIONS_DIR)/$(shell date +%Y%m%d%H%M%S)_initial_schema.sql
	@echo "-- +goose Down" >> $(MIGRATIONS_DIR)/$(shell date +%Y%m%d%H%M%S)_initial_schema.sql
	@echo "-- Add rollback statements here" >> $(MIGRATIONS_DIR)/$(shell date +%Y%m%d%H%M%S)_initial_schema.sql
	@echo "Initial migration template created. Please edit the file to add your schema."

# Help
help:
	@echo "Available commands:"
	@echo "  install-goose          - Install Goose binary"
	@echo "  create-migration       - Create new SQL migration (requires NAME=migration_name)"
	@echo "  create-go-migration    - Create new Go migration (requires NAME=migration_name)"
	@echo "  migrate-up             - Apply all pending migrations"
	@echo "  migrate-down           - Rollback the last migration"
	@echo "  migrate-down-to        - Rollback to specific version (requires VERSION=version)"
	@echo "  migrate-status         - Show migration status"
	@echo "  migrate-validate       - Validate migrations"
	@echo "  migrate-reset          - Reset database (rollback all migrations)"
	@echo "  migrate-fix            - Fix migration sequence"
	@echo "  migrate-version        - Show current migration version"
	@echo "  create-initial-migration - Create initial migration template"
	@echo ""
	@echo "Environment variables:"
	@echo "  DB_DRIVER              - Database driver (default: mysql)"
	@echo "  DB_STRING              - Database connection string"
	@echo "  MIGRATIONS_DIR         - Migrations directory (default: ./migrations)"

.PHONY: install-goose create-migration create-go-migration migrate-up migrate-down migrate-down-to migrate-status migrate-validate migrate-reset migrate-fix migrate-version create-initial-migration help 