# Twelve-Factor App Compliance

This document describes how this application implements the [Twelve-Factor App](https://12factor.net/) methodology.

## Implementation Status

✅ **Fully Compliant**: 12/12 factors implemented

### Recent Improvements

- ✅ **Factor III (Config)**: Now fully environment-based with optional config files
- ✅ **Factor IX (Disposability)**: Graceful shutdown implemented for all modes
- ✅ **Factor XII (Admin Processes)**: Migration CLI commands added

---

## Factor III: Config - Store config in the environment

### Implementation

**Status**: ✅ **Fully Compliant**

The application can now run with **ONLY environment variables**, without any config files.

### Environment-Only Deployment

```bash
# Set configuration via environment variables
export APP_SERVER_PORT=8080
export APP_DATABASE_HOST=db.production.com
export APP_DATABASE_PORT=5432
export APP_DATABASE_USER=prod_user
export APP_DATABASE_PASSWORD=secret_password
export APP_DATABASE_DBNAME=go_clean_arch_prod
export APP_LOGGER_LEVEL=info
export APP_LOGGER_FORMAT=json

# Run application (no --config flag needed)
./app api
```

### Docker/Kubernetes Deployment

```dockerfile
# Dockerfile does NOT bundle config files
FROM alpine:latest
COPY --from=builder /app/bin/app .
# No COPY configs/ - environment-agnostic image
```

```yaml
# Kubernetes Deployment
apiVersion: apps/v1
kind: Deployment
spec:
  template:
    spec:
      containers:
      - name: api
        image: go-clean-arch:v1.0.0
        env:
        - name: APP_DATABASE_HOST
          value: "postgres.default.svc.cluster.local"
        - name: APP_DATABASE_PASSWORD
          valueFrom:
            secretKeyRef:
              name: db-credentials
              key: password
```

### Configuration Precedence

1. **Environment Variables** (highest priority)
2. **Config File** (if specified via --config flag)
3. **Default Config File** (configs/config.toml if exists)
4. **Built-in Defaults** (lowest priority)

### Config Files are Optional

Config files are provided for **local development convenience only**:

```bash
# Local development with config file
./app api --config=configs/config.toml

# Or use default if it exists
./app api

# Production with environment variables only
APP_DATABASE_HOST=prod-db ./app api
```

### Benefits

✅ **Environment-Agnostic Images**: Same Docker image for dev/staging/prod
✅ **No Secrets in Images**: Config files not bundled into containers
✅ **Easy Configuration Management**: Use Kubernetes ConfigMaps/Secrets
✅ **Twelve-Factor Compliant**: Strict separation of config from code

### Migration Guide

**Before** (required config file):
```bash
docker run go-clean-arch:latest --mode=api --config=/app/configs/config.yaml
```

**After** (environment variables):
```bash
docker run \
  -e APP_SERVER_PORT=8080 \
  -e APP_DATABASE_HOST=postgres \
  -e APP_DATABASE_PASSWORD=secret \
  go-clean-arch:latest api
```

---

## Factor XII: Admin Processes - Run admin/management tasks as one-off processes

### Implementation

**Status**: ✅ **Fully Compliant**

Database migrations and admin tasks are now run as one-off processes using the same codebase and environment.

### Migration Commands

```bash
# Apply all pending migrations
./app migrate up

# Rollback last migration
./app migrate down

# Check migration status
./app migrate status

# Create new migration file
./app migrate create "add_orders_table"
```

### Example Usage

```bash
# 1. Check current migration status
$ ./app migrate status

Migration Status:
================
✅ 001_create_users_table.sql (applied)
✅ 002_create_orders_table.sql (applied)
⏳ 003_add_email_index.sql (pending)

Total: 3 migrations (2 applied, 1 pending)

# 2. Apply pending migrations
$ ./app migrate up
[INFO] Running database migrations...
[INFO] Applying migration file=003_add_email_index.sql
[INFO] Migrations applied successfully count=1

# 3. Create new migration
$ ./app migrate create "add_payment_status"
[INFO] Migration file created file=migrations/004_add_payment_status.sql
✅ Created: migrations/004_add_payment_status.sql

Edit this file to add your SQL migration, then run:
  app migrate up
```

### Production Usage

```bash
# In production, migrations run with same environment config
export APP_DATABASE_HOST=prod-db.example.com
export APP_DATABASE_USER=admin
export APP_DATABASE_PASSWORD=secure_password

# Run migrations as one-off process
./app migrate up

# Then deploy application
./app api
```

### Kubernetes Job Example

```yaml
apiVersion: batch/v1
kind: Job
metadata:
  name: db-migration
spec:
  template:
    spec:
      containers:
      - name: migrate
        image: go-clean-arch:v1.0.0
        command: ["./app", "migrate", "up"]
        env:
        - name: APP_DATABASE_HOST
          value: "postgres.default.svc.cluster.local"
        - name: APP_DATABASE_PASSWORD
          valueFrom:
            secretKeyRef:
              name: db-credentials
              key: password
      restartPolicy: Never
```

### Auto-Migrate Disabled

Auto-migrate is now **disabled by default** for production safety:

```toml
# configs/config.example.toml
[database]
auto_migrate = false  # Use explicit migrations instead
```

**Why?**
- ✅ Explicit control over schema changes
- ✅ Migration history tracking
- ✅ Ability to rollback changes
- ✅ Better for team collaboration
- ✅ Follows Twelve-Factor App guidelines

### Migration File Structure

```sql
-- migrations/003_add_email_index.sql
-- Migration: add_email_index
-- Created: 2025-11-20

CREATE INDEX idx_users_email ON users(email);
```

### Benefits

✅ **One-Off Processes**: Migrations run as separate processes
✅ **Same Environment**: Uses same config/env vars as main app
✅ **Same Codebase**: No separate migration tools needed
✅ **Audit Trail**: Migrations tracked in database
✅ **Rollback Support**: Can undo migrations if needed

---

## Summary of Twelve-Factor Compliance

| Factor | Status | Implementation |
|--------|--------|----------------|
| I. Codebase | ✅ | Single Git repo, multiple deploys |
| II. Dependencies | ✅ | Go modules with explicit versions |
| III. Config | ✅ | **Environment-only support added** |
| IV. Backing Services | ✅ | Attachable resources via config |
| V. Build, Release, Run | ✅ | Multi-stage Docker builds |
| VI. Processes | ✅ | Stateless processes |
| VII. Port Binding | ✅ | Self-contained HTTP server |
| VIII. Concurrency | ✅ | Process types (api, worker, cron) |
| IX. Disposability | ✅ | **Graceful shutdown implemented** |
| X. Dev/Prod Parity | ✅ | Same images, same services |
| XI. Logs | ✅ | Structured logs to stdout |
| XII. Admin Processes | ✅ | **Migration commands added** |

**Total Score**: 12/12 ✅

---

## Best Practices

### Development

```bash
# Use config files for convenience
cp configs/config.example.toml configs/config.toml
# Edit configs/config.toml with your settings
./app api
```

### Production

```bash
# Use environment variables only
export APP_DATABASE_HOST=prod-db
export APP_DATABASE_PASSWORD=$(cat /run/secrets/db_password)
export APP_STRIPE_API_KEY=$(cat /run/secrets/stripe_key)

# Run migrations
./app migrate up

# Start application
./app api
```

### Docker

```bash
# Build image (no config files bundled)
docker build -t go-clean-arch:v1.0.0 .

# Run with environment variables
docker run \
  -e APP_SERVER_PORT=8080 \
  -e APP_DATABASE_HOST=postgres \
  -e APP_DATABASE_PASSWORD=secret \
  -p 8080:8080 \
  go-clean-arch:v1.0.0 api
```

### Docker Compose

See `deployments/docker-compose.yaml` for a complete example of environment-based configuration.

---

## References

- [Twelve-Factor App](https://12factor.net/)
- [Factor III - Config](https://12factor.net/config)
- [Factor IX - Disposability](https://12factor.net/disposability)
- [Factor XII - Admin Processes](https://12factor.net/admin-processes)
- [TWELVE_FACTOR_ASSESSMENT.md](./TWELVE_FACTOR_ASSESSMENT.md) - Detailed assessment
- [GRACEFUL_SHUTDOWN.md](./GRACEFUL_SHUTDOWN.md) - Graceful shutdown documentation

---

**Last Updated**: 2025-11-20
**Compliance Level**: 100% (12/12 factors)
