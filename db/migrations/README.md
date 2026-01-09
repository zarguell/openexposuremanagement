# Database Migrations

This directory contains SQL migrations for the Open Exposure Management database.

## Migration Tool

We use [golang-migrate](https://github.com/golang-migrate/migrate) for database migrations.

## Installation

```bash
# Install golang-migrate
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Or using brew (macOS)
brew install golang-migrate

# Or using curl (Linux)
curl -L https://github.com/golang-migrate/migrate/releases/download/v4.16.2/migrate.linux-amd64.tar.gz | tar xvz
mv migrate /usr/local/bin/migrate
```

## Usage

```bash
# Apply all pending migrations
make migrate-up

# Rollback the last migration
make migrate-down

# Create a new migration
make migrate-create name=create_users_table

# Check migration status
migrate -path db/migrations -database "$DATABASE_URL" version
```

## File Naming Convention

Migrations use the format: `{version}_{name}.up.sql` and `{version}_{name}.down.sql`

Example: `000001_create_tenants_table.up.sql`

## Guidelines

- Each migration should be atomic and reversible
- Use transactions in up migrations
- Provide corresponding down migrations for rollback
- Index creation should happen after table creation
- Foreign keys should have explicit names for easier management
