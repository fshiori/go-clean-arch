# Database Migrations

This directory contains SQL migration files for database schema management using **Atlas**.

## What is Atlas?

[Atlas](https://atlasgo.io/) is a modern database schema management tool that provides:
- **Versioned migrations** with automatic tracking
- **Schema validation** and safety checks
- **Multiple database support**: PostgreSQL, MySQL, SQLite, and more
- **Declarative schema definitions**
- **Team collaboration** features with better conflict resolution

## Migration Files

- `001_create_users_table.sql` - Creates the users table
- `002_create_orders_table.sql` - Creates the orders table with foreign key to users

## Prerequisites

### Install Atlas CLI

**macOS:**
```bash
brew install ariga/tap/atlas
```

**Linux:**
```bash
curl -sSf https://atlasgo.sh | sh
```

**Windows:**
Download from https://release.ariga.io/atlas/atlas-windows-amd64-latest.exe

**Or use Make:**
```bash
make atlas-install
```

**Or use Docker:**
```bash
docker pull arigaio/atlas
```

### Verify Installation

```bash
atlas version
# or
make atlas-version
```

## Configuration

Atlas uses the `atlas.hcl` configuration file at the project root. The configuration supports multiple environments:
- `local` - For local development
- `dev` - For development environment
- `production` - For production environment

### Environment Variables

Set your database connection URL:

```bash
# PostgreSQL
export DATABASE_URL="postgres://user:password@localhost:5432/dbname?sslmode=disable"

# MySQL
export DATABASE_URL="mysql://user:password@tcp(localhost:3306)/dbname"

# SQLite
export DATABASE_URL="sqlite://file.db"
```

Or use individual environment variables (Twelve-Factor compliant):
```bash
export APP_DATABASE_HOST=localhost
export APP_DATABASE_USER=postgres
export APP_DATABASE_PASSWORD=secret
export APP_DATABASE_DBNAME=go_clean_arch
```

### Optional: Set Atlas Environment

```bash
export ATLAS_ENV=local  # default
# or
export ATLAS_ENV=production
```

## How to Use Migrations

### Option 1: Using the Built-in CLI (Recommended)

The application has built-in migration commands that use Atlas:

```bash
# Apply all pending migrations
./app migrate up
# or
go run ./cmd/app/main.go migrate up

# Show migration status
./app migrate status

# Create a new migration
./app migrate create "add_products_table"

# Rollback last migration
./app migrate down

# Rollback 2 migrations
./app migrate down 2

# Validate migration files
./app migrate validate
```

### Option 2: Using Make Commands

```bash
# Apply all pending migrations
make migrate-up

# Show migration status
make migrate-status

# Create a new migration
make migrate-create NAME="add_products_table"

# Rollback last migration
make migrate-down

# Validate migration files
make migrate-validate
```

### Option 3: Using Atlas CLI Directly

```bash
# Apply migrations
atlas migrate apply --env local --url "$DATABASE_URL"

# Show status
atlas migrate status --env local --url "$DATABASE_URL"

# Create new migration
atlas migrate new --env local add_products_table

# Validate migrations
atlas migrate validate --env local --url "$DATABASE_URL"

# Lint migrations
atlas migrate lint --env local
```

### Option 4: Using Docker

```bash
docker run --rm \
  -v $(pwd)/migrations:/migrations \
  -e DATABASE_URL="postgres://user:pass@host:5432/db" \
  arigaio/atlas migrate apply --env local --url "$DATABASE_URL"
```

## Migration File Naming

Atlas uses timestamp-based naming for better conflict resolution:
- Format: `YYYYMMDDHHMMSS_description.sql`
- Example: `20240115120000_add_products_table.sql`

This is better than sequential numbering (001, 002, etc.) because:
- **No conflicts** when multiple developers create migrations simultaneously
- **Better ordering** based on creation time
- **Team collaboration** friendly

## Migration Best Practices

### 1. Always Create Migrations for Schema Changes

Never modify the database schema directly. Always create a migration file:

```bash
make migrate-create NAME="add_column_to_users"
```

### 2. Test Migrations Before Applying

```bash
# Validate migration files
make migrate-validate

# Check status
make migrate-status
```

### 3. Use Transactions

Atlas automatically wraps migrations in transactions (when supported by the database).

### 4. Keep Migrations Idempotent

Write migrations that can be safely re-run:

```sql
-- Good: Check if exists
CREATE TABLE IF NOT EXISTS products (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(255) NOT NULL
);

-- Good: Check before adding column
ALTER TABLE users ADD COLUMN IF NOT EXISTS phone VARCHAR(20);

-- Avoid: Migrations that fail on re-run
CREATE TABLE products (...);  -- Will fail if table exists
```

### 5. Never Modify Applied Migrations

Once a migration is applied to production, never modify it. Create a new migration instead.

### 6. Use Descriptive Names

```bash
# Good
make migrate-create NAME="add_email_index_to_users"
make migrate-create NAME="create_orders_table"

# Bad
make migrate-create NAME="fix"
make migrate-create NAME="update"
```

### 7. Add Comments

```sql
-- Migration: Add email index to users table
-- Purpose: Improve query performance for email lookups
-- Date: 2024-01-15

CREATE INDEX idx_users_email ON users(email);
```

## Database Schema

### Users Table

| Column        | Type         | Description                    |
|---------------|--------------|--------------------------------|
| id            | BIGINT       | Primary key, auto-increment    |
| email         | VARCHAR(255) | Unique email address           |
| password_hash | VARCHAR(255) | Hashed password                |
| created_at    | TIMESTAMP    | Record creation time           |
| updated_at    | TIMESTAMP    | Record last update time        |

### Orders Table

| Column       | Type          | Description                           |
|--------------|---------------|---------------------------------------|
| id           | BIGINT        | Primary key, auto-increment           |
| user_id      | BIGINT        | Foreign key to users table            |
| total_amount | DECIMAL(10,2) | Order total amount                    |
| status       | VARCHAR(50)   | Order status (pending, paid, etc.)    |
| items        | JSON          | Order items stored as JSON            |
| created_at   | TIMESTAMP     | Record creation time                  |
| updated_at   | TIMESTAMP     | Record last update time               |

## Troubleshooting

### Atlas not found

```bash
# Install Atlas
make atlas-install

# Or manually install
brew install ariga/tap/atlas  # macOS
curl -sSf https://atlasgo.sh | sh  # Linux
```

### Database connection failed

```bash
# Check DATABASE_URL is set
echo $DATABASE_URL

# Or check individual config values
echo $APP_DATABASE_HOST
echo $APP_DATABASE_USER

# Test connection
atlas migrate status --env local --url "$DATABASE_URL"
```

### Migration validation failed

```bash
# Validate all migrations
make migrate-validate

# Check migration file syntax
atlas migrate validate --env local --url "$DATABASE_URL"

# Lint migrations for common issues
atlas migrate lint --env local
```

### Migration conflicts

If you have migration conflicts (e.g., two developers created migrations with same number):
1. Atlas uses timestamps, so this should rarely happen
2. If it does, rename one migration file to have a later timestamp
3. Coordinate with your team to avoid simultaneous migrations

## Twelve-Factor App Compliance

This migration setup follows Twelve-Factor App methodology:

- **Factor III (Config)**: Database URL from environment variables
- **Factor XII (Admin Processes)**: Migrations run as one-off admin processes
- **Factor VIII (Concurrency)**: Migrations are separate from app runtime
- **Factor V (Build, Release, Run)**: Migrations are part of release process

## CI/CD Integration

### GitHub Actions Example

```yaml
- name: Install Atlas
  run: |
    curl -sSf https://atlasgo.sh | sh

- name: Run migrations
  env:
    DATABASE_URL: ${{ secrets.DATABASE_URL }}
  run: |
    atlas migrate apply --env production --url "$DATABASE_URL"
```

### GitLab CI Example

```yaml
migrate:
  script:
    - curl -sSf https://atlasgo.sh | sh
    - atlas migrate apply --env production --url "$DATABASE_URL"
  only:
    - main
```

## Additional Resources

- [Atlas Documentation](https://atlasgo.io/docs)
- [Atlas CLI Reference](https://atlasgo.io/cli-reference)
- [Migration Best Practices](https://atlasgo.io/guides/migration-best-practices)
- [Twelve-Factor App](https://12factor.net/)

## Notes

- All tables use InnoDB engine for transaction support (MySQL)
- Character set is utf8mb4 for full Unicode support (including emojis)
- Timestamps are managed by database triggers or application logic
- Foreign key constraints ensure referential integrity
- Atlas tracks applied migrations in the `atlas_schema_revisions` table
