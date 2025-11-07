# Database Migrations

This directory contains SQL migration files for MySQL database schema.

## Migration Files

- `001_create_users_table.sql` - Creates the users table
- `002_create_orders_table.sql` - Creates the orders table with foreign key to users

## How to Run Migrations

### Option 1: Manually using MySQL CLI

```bash
mysql -u root -p go_clean_arch < migrations/001_create_users_table.sql
mysql -u root -p go_clean_arch < migrations/002_create_orders_table.sql
```

### Option 2: Using golang-migrate CLI

Install golang-migrate:
```bash
# macOS
brew install golang-migrate

# Linux
curl -L https://github.com/golang-migrate/migrate/releases/download/v4.17.0/migrate.linux-amd64.tar.gz | tar xvz
sudo mv migrate /usr/local/bin/

# Windows
scoop install migrate
```

Run migrations:
```bash
# Set your database credentials
export DB_URL="mysql://user:password@tcp(localhost:3306)/go_clean_arch"

# Run migrations
migrate -path migrations -database "$DB_URL" up

# Rollback
migrate -path migrations -database "$DB_URL" down
```

### Option 3: Using mysql command from Docker

```bash
docker exec -i mysql_container mysql -uroot -ppassword go_clean_arch < migrations/001_create_users_table.sql
docker exec -i mysql_container mysql -uroot -ppassword go_clean_arch < migrations/002_create_orders_table.sql
```

## Schema

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

## Notes

- All tables use InnoDB engine for transaction support
- Character set is utf8mb4 for full Unicode support (including emojis)
- Timestamps are managed by MySQL triggers (auto-update on change)
- Foreign key constraint on orders.user_id ensures referential integrity
