# Twelve-Factor App Methodology Assessment

**Date**: 2025-11-20
**Repository**: go-clean-arch
**Branch**: develop

## Executive Summary

This assessment evaluates the `go-clean-arch` repository against the [Twelve-Factor App](https://12factor.net/) methodology. The application demonstrates **strong alignment** with most factors, scoring **9.5 out of 12** factors as fully or substantially compliant.

### Compliance Overview

| Factor | Status | Score | Priority |
|--------|--------|-------|----------|
| I. Codebase | ✅ Compliant | 1.0 | - |
| II. Dependencies | ✅ Compliant | 1.0 | - |
| III. Config | ⚠️ Partial | 0.7 | Medium |
| IV. Backing Services | ✅ Compliant | 1.0 | - |
| V. Build, Release, Run | ✅ Compliant | 1.0 | - |
| VI. Processes | ✅ Compliant | 1.0 | - |
| VII. Port Binding | ✅ Compliant | 1.0 | - |
| VIII. Concurrency | ✅ Compliant | 1.0 | - |
| IX. Disposability | ❌ Non-Compliant | 0.3 | **High** |
| X. Dev/Prod Parity | ✅ Compliant | 1.0 | - |
| XI. Logs | ✅ Compliant | 1.0 | - |
| XII. Admin Processes | ⚠️ Partial | 0.5 | Low |
| **Total** | | **9.5/12** | |

**Key Findings**:
- ✅ **Strong Foundation**: Excellent architecture with clean separation of concerns
- ❌ **Critical Gap**: Missing graceful shutdown/signal handling (Factor IX)
- ⚠️ **Improvement Needed**: Config relies on files; should be fully environment-based (Factor III)
- ⚠️ **Minor Gap**: No standardized way to run one-off admin tasks (Factor XII)

---

## Detailed Factor Analysis

### I. Codebase ✅ COMPLIANT

> **Twelve-Factor Principle**: One codebase tracked in revision control, many deploys

**Status**: ✅ **Fully Compliant**

**Evidence**:
- Single Git repository at `/home/user/go-clean-arch`
- Clean branch structure with `main` and `develop` branches
- Supports multiple deployment targets (api, worker, cron) from single codebase
- Multi-platform build support (Linux, macOS, Windows)

**Observations**:
```bash
git branch -a
# Shows clean branch structure
# Multiple deployment modes from same source
```

**Files**:
- `cmd/app/cmd/api.go` - API mode
- `cmd/app/cmd/worker.go` - Worker mode
- `cmd/app/cmd/cron.go` - Cron mode

**Score**: 1.0/1.0

**Recommendation**: ✅ No action needed

---

### II. Dependencies ✅ COMPLIANT

> **Twelve-Factor Principle**: Explicitly declare and isolate dependencies

**Status**: ✅ **Fully Compliant**

**Evidence**:
- ✅ `go.mod` explicitly declares all dependencies with versions
- ✅ `go.sum` provides cryptographic checksums for integrity
- ✅ No reliance on system-wide packages
- ✅ Docker multi-stage build ensures isolated build environment
- ✅ Vendor-agnostic dependency management via Go modules

**Key Dependencies**:
```go
// go.mod
module go-clean-arch

go 1.24

require (
    github.com/gin-gonic/gin v1.9.1
    github.com/jmoiron/sqlx v1.3.5
    github.com/spf13/cobra v1.8.0
    github.com/spf13/viper v1.18.2
    github.com/google/wire v0.6.0
    // ... all explicitly versioned
)
```

**Build Isolation**:
```dockerfile
# Dockerfile lines 10-14
COPY go.mod go.sum ./
RUN go mod download
# Dependencies downloaded in isolated container
```

**Score**: 1.0/1.0

**Recommendation**: ✅ No action needed

---

### III. Config ⚠️ PARTIALLY COMPLIANT

> **Twelve-Factor Principle**: Store config in the environment

**Status**: ⚠️ **Partially Compliant** (70%)

**Current Implementation**:

✅ **What's Good**:
- Environment variable override support via `APP_*` prefix
- Viper automatically reads env vars: `pkg/config/config.go:89-94`
- `.env.example` provides template for environment-based config
- Docker Compose uses environment variables: `deployments/docker-compose.yaml:41-47`

❌ **What's Missing**:
- **Config files still required**: Application fails without `configs/config.toml`
- **File-based config is primary**: Env vars only override file values
- **Not fully environment-based**: Cannot run with only environment variables

**Evidence**:
```go
// pkg/config/config.go:100-102
if err := v.ReadInConfig(); err != nil {
    return nil, fmt.Errorf("failed to read config file: %w", err)
}
// Application REQUIRES config file to exist
```

**Dockerfile**:
```dockerfile
# Dockerfile line 38
COPY --from=builder /app/configs ./configs
# Config files bundled into image - antipattern
```

**Current Behavior**:
```bash
# This WORKS (12-factor compliant)
APP_DATABASE_PASSWORD=secret ./app api

# This FAILS (not 12-factor compliant)
export APP_SERVER_PORT=8080
export APP_DATABASE_HOST=db.example.com
./app api
# Error: failed to read config file: configs/config.toml not found
```

**Score**: 0.7/1.0

**Recommendation**: 🔧 **IMPROVE** (Priority: Medium)

**Required Changes**:

1. **Make config file optional**:
```go
// pkg/config/config.go
func Load(path string) (*Config, error) {
    v := viper.New()

    // Set defaults first
    setDefaults(v)

    // Try to read config file (non-fatal)
    if path != "" {
        v.SetConfigFile(path)
        if err := v.ReadInConfig(); err != nil {
            // Log warning but continue
            log.Printf("Config file not found, using env vars and defaults: %v", err)
        }
    }

    // Environment variables can provide all config
    v.SetEnvPrefix("APP")
    v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
    v.AutomaticEnv()

    // ... rest of function
}
```

2. **Update Dockerfile to not bundle configs**:
```dockerfile
# Remove this line:
# COPY --from=builder /app/configs ./configs

# App should work without config files
```

3. **Document environment-first approach**:
```markdown
# README.md
## Configuration

The application can run with ONLY environment variables:

export APP_SERVER_PORT=8080
export APP_DATABASE_HOST=postgres
export APP_DATABASE_USER=myuser
export APP_DATABASE_PASSWORD=secret
./app api

Config files are optional for local development convenience.
```

**Files to Modify**:
- `pkg/config/config.go` - Make file optional
- `Dockerfile` - Remove config file copy
- `cmd/app/cmd/root.go` - Handle missing config file gracefully
- `README.md` - Document env-var-only usage

---

### IV. Backing Services ✅ COMPLIANT

> **Twelve-Factor Principle**: Treat backing services as attached resources

**Status**: ✅ **Fully Compliant**

**Evidence**:
- ✅ Database connection configured via URL/connection string
- ✅ RabbitMQ configured via connection URL
- ✅ Stripe gateway configured via API keys
- ✅ All backing services swappable via configuration
- ✅ No distinction between local and third-party services

**Configuration Examples**:
```toml
# configs/config.example.toml

[database]
host = "localhost"  # Could be "prod-db.aws.com"
port = 5432
# Swappable without code changes

[rabbitmq]
url = "amqp://guest:guest@localhost:5672/"
# Could be "amqp://user:pass@cloudamqp.com:5672/"

[stripe]
api_key = "your-stripe-api-key"
# Third-party service treated identically
```

**Implementation**:
```go
// internal/adapter/repository/db.go
// Database is abstracted via interface
type UserRepository interface {
    Create(ctx context.Context, user *domain.User) error
    // ... can swap PostgreSQL, MySQL, etc.
}

// internal/adapter/gateway/stripe_gateway.go
// External services abstracted
type PaymentGateway interface {
    Charge(amount int, currency string) error
    // ... can swap Stripe, PayPal, etc.
}
```

**Docker Compose**:
```yaml
# docker-compose.yaml
environment:
  DB_HOST: postgres  # Local dev
  # In production: DB_HOST: rds.amazonaws.com
```

**Score**: 1.0/1.0

**Recommendation**: ✅ No action needed

---

### V. Build, Release, Run ✅ COMPLIANT

> **Twelve-Factor Principle**: Strictly separate build and run stages

**Status**: ✅ **Fully Compliant**

**Evidence**:

**1. Build Stage** (Dockerfile lines 1-20):
```dockerfile
FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/app ./cmd/app
# ✅ Compile-time: source → binary
```

**2. Release Stage** (Implicit):
```bash
# Image tagging creates immutable releases
docker build -t go-clean-arch:v1.2.3 .
docker push registry.example.com/go-clean-arch:v1.2.3
# ✅ Release = Build + Config
```

**3. Run Stage** (Dockerfile lines 22-52):
```dockerfile
FROM alpine:latest
COPY --from=builder /app/bin/app .
ENTRYPOINT ["./app"]
CMD ["--mode=api"]
# ✅ Runtime: execute pre-built binary
```

**Stage Separation**:
- ✅ **Build**: `make build` produces platform-specific binaries
- ✅ **Release**: Docker images tagged with versions
- ✅ **Run**: Container execution with environment-specific config

**Makefile**:
```makefile
# Makefile
build:
    go build -o bin/app ./cmd/app

build-all:
    GOOS=linux GOARCH=amd64 go build -o bin/app-linux-amd64
    GOOS=darwin GOARCH=amd64 go build -o bin/app-darwin-amd64
    # ✅ Separate build artifacts
```

**Score**: 1.0/1.0

**Recommendation**: ✅ No action needed

---

### VI. Processes ✅ COMPLIANT

> **Twelve-Factor Principle**: Execute the app as one or more stateless processes

**Status**: ✅ **Fully Compliant**

**Evidence**:
- ✅ No local filesystem state (beyond temp files)
- ✅ No in-process session storage
- ✅ Persistent state stored in database (PostgreSQL)
- ✅ Shared state via backing services (RabbitMQ)
- ✅ Process restarts are safe

**Architecture**:
```go
// internal/usecase/user_interactor.go
type UserInteractor struct {
    repo port.UserRepository  // ✅ State in database
    // No in-memory state stored
}

func (i *UserInteractor) CreateUser(ctx context.Context, ...) error {
    // ✅ All state persisted to database
    return i.repo.Create(ctx, user)
}
```

**No Session Affinity**:
```yaml
# docker-compose.yaml
api:
  restart: unless-stopped
  # ✅ Process can restart without data loss
  # ✅ Multiple instances can run simultaneously
```

**Stateless HTTP Handlers**:
```go
// internal/delivery/http/handler/user_handler.go
type UserHandler struct {
    usecase usecase.UserUsecase
    // ✅ No request state stored in handler
}
```

**Score**: 1.0/1.0

**Recommendation**: ✅ No action needed

---

### VII. Port Binding ✅ COMPLIANT

> **Twelve-Factor Principle**: Export services via port binding

**Status**: ✅ **Fully Compliant**

**Evidence**:
- ✅ Self-contained HTTP server (Gin framework)
- ✅ No reliance on external web server (Apache/Nginx)
- ✅ Port configured via environment
- ✅ Service accessible via HTTP protocol

**Implementation**:
```go
// internal/app/api.go:40
func (s *APIServer) Start() error {
    router := InitializeAPIRouter(s.db, s.cfg)
    addr := fmt.Sprintf(":%d", s.port)

    // ✅ Self-contained HTTP server
    if err := router.Run(addr); err != nil {
        return fmt.Errorf("failed to start server: %w", err)
    }
    return nil
}
```

**Configuration**:
```toml
[server]
port = 8080  # ✅ Configurable port binding
host = "0.0.0.0"
```

**Docker**:
```dockerfile
EXPOSE 8080  # ✅ Service exported on port
```

```yaml
# docker-compose.yaml
ports:
  - "8080:8080"  # ✅ Port binding to host
```

**Score**: 1.0/1.0

**Recommendation**: ✅ No action needed

---

### VIII. Concurrency ✅ COMPLIANT

> **Twelve-Factor Principle**: Scale out via the process model

**Status**: ✅ **Fully Compliant**

**Evidence**:
- ✅ Multiple process types: `api`, `worker`, `cron`
- ✅ Each process type handles different workloads
- ✅ Horizontal scaling via process replication
- ✅ No shared state between processes

**Process Types**:
```bash
# Run multiple API processes for HTTP scaling
./app api --config=config.toml  # Process 1
./app api --config=config.toml  # Process 2
# ✅ Load balancer distributes requests

# Run multiple workers for background job scaling
./app worker --config=config.toml  # Worker 1
./app worker --config=config.toml  # Worker 2
# ✅ RabbitMQ distributes messages

# Run cron for scheduled tasks
./app cron --config=config.toml
# ✅ Single process for scheduled work
```

**Docker Compose Scaling**:
```yaml
services:
  api:
    # ✅ Can scale: docker-compose up --scale api=3

  worker:
    # ✅ Can scale: docker-compose up --scale worker=5

  cron:
    # ✅ Single instance for scheduled tasks
```

**Process Architecture**:
```
┌─────────────────────────────────────┐
│         Load Balancer               │
└────────┬────────┬────────┬──────────┘
         │        │        │
    ┌────▼───┐┌──▼────┐┌──▼────┐
    │ API    ││ API   ││ API   │  ✅ Horizontal scaling
    │ Proc 1 ││ Proc 2││ Proc 3│
    └────┬───┘└───┬───┘└───┬───┘
         │        │        │
    ┌────▼────────▼────────▼────┐
    │      PostgreSQL            │
    └────────────────────────────┘

┌─────────────────────────────────────┐
│         RabbitMQ Queue              │
└────┬────────┬────────┬──────────────┘
     │        │        │
┌────▼───┐┌──▼────┐┌──▼────┐
│ Worker ││Worker ││Worker │  ✅ Concurrent processing
│ Proc 1 ││Proc 2 ││Proc 3 │
└────────┘└───────┘└───────┘
```

**Score**: 1.0/1.0

**Recommendation**: ✅ No action needed

---

### IX. Disposability ❌ NON-COMPLIANT

> **Twelve-Factor Principle**: Maximize robustness with fast startup and graceful shutdown

**Status**: ❌ **Non-Compliant** (30%)

**What's Good** (30%):
- ✅ Fast startup time (Go compiled binary)
- ✅ Minimal initialization overhead

**What's Missing** (70%):
- ❌ **No graceful shutdown handling**
- ❌ **No SIGTERM/SIGINT signal handling**
- ❌ **Active connections abruptly terminated**
- ❌ **In-flight requests may fail**
- ❌ **Database connections not properly closed**
- ❌ **RabbitMQ consumers not gracefully stopped**

**Current Implementation**:
```go
// internal/app/api.go:40-43
func (s *APIServer) Start() error {
    router := InitializeAPIRouter(s.db, s.cfg)
    addr := fmt.Sprintf(":%d", s.port)

    // ❌ Blocking call with no graceful shutdown
    if err := router.Run(addr); err != nil {
        return fmt.Errorf("failed to start server: %w", err)
    }
    return nil
}
// When process receives SIGTERM, it immediately exits
// No cleanup, no graceful connection draining
```

**Evidence of Non-Compliance**:
```bash
# No signal handling found
grep -r "signal.Notify\|syscall.SIGTERM\|graceful" --include="*.go"
# No results - no graceful shutdown implemented
```

**Impact**:
- 🔴 **High Risk**: API requests fail mid-processing during deployment
- 🔴 **Data Loss Risk**: Worker messages may be lost during shutdown
- 🔴 **Connection Leaks**: Database connections not released properly
- 🔴 **Poor User Experience**: 502/504 errors during rolling deploys

**Score**: 0.3/1.0

**Recommendation**: 🔧 **CRITICAL - IMPLEMENT IMMEDIATELY** (Priority: High)

**Required Implementation**:

#### 1. API Server Graceful Shutdown

**File**: `internal/app/api.go`

```go
package app

import (
    "context"
    "fmt"
    "go-clean-arch/pkg/config"
    "go-clean-arch/pkg/logger"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"

    "github.com/jmoiron/sqlx"
)

type APIServer struct {
    db     *sqlx.DB
    cfg    *config.Config
    port   int
    server *http.Server
}

func NewAPIServer(db *sqlx.DB, cfg *config.Config, port int) *APIServer {
    return &APIServer{
        db:   db,
        cfg:  cfg,
        port: port,
    }
}

func (s *APIServer) Start() error {
    logger.Info("Starting API server...")

    router := InitializeAPIRouter(s.db, s.cfg)

    s.server = &http.Server{
        Addr:    fmt.Sprintf(":%d", s.port),
        Handler: router,
    }

    // Channel to listen for errors from server
    serverErrors := make(chan error, 1)

    // Start server in goroutine
    go func() {
        logger.Info("API server listening", "address", s.server.Addr)
        serverErrors <- s.server.ListenAndServe()
    }()

    // Channel to listen for interrupt signals
    shutdown := make(chan os.Signal, 1)
    signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

    // Block until error or shutdown signal
    select {
    case err := <-serverErrors:
        return fmt.Errorf("server error: %w", err)

    case sig := <-shutdown:
        logger.Info("Shutdown signal received", "signal", sig)

        // Give outstanding requests 30 seconds to complete
        ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
        defer cancel()

        // Gracefully shutdown server
        if err := s.server.Shutdown(ctx); err != nil {
            logger.Error("Graceful shutdown failed", "error", err)

            // Force close if graceful shutdown fails
            if err := s.server.Close(); err != nil {
                return fmt.Errorf("failed to close server: %w", err)
            }
            return fmt.Errorf("failed to gracefully shutdown: %w", err)
        }

        // Close database connection
        if err := s.db.Close(); err != nil {
            logger.Error("Failed to close database", "error", err)
        }

        logger.Info("Server stopped gracefully")
        return nil
    }
}
```

#### 2. Worker Graceful Shutdown

**File**: `internal/app/worker.go`

```go
func (w *Worker) Start() error {
    logger.Info("Starting worker...")

    consumer := InitializeOrderConsumer(w.db, w.cfg)

    // Channel for shutdown signal
    shutdown := make(chan os.Signal, 1)
    signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

    // Channel for consumer errors
    consumerErrors := make(chan error, 1)

    // Start consuming in goroutine
    go func() {
        consumerErrors <- consumer.Start()
    }()

    // Block until error or shutdown
    select {
    case err := <-consumerErrors:
        return fmt.Errorf("consumer error: %w", err)

    case sig := <-shutdown:
        logger.Info("Shutdown signal received", "signal", sig)

        // Stop accepting new messages
        ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
        defer cancel()

        // Wait for in-flight messages to complete
        if err := consumer.Shutdown(ctx); err != nil {
            logger.Error("Failed to gracefully shutdown consumer", "error", err)
        }

        // Close database
        if err := w.db.Close(); err != nil {
            logger.Error("Failed to close database", "error", err)
        }

        logger.Info("Worker stopped gracefully")
        return nil
    }
}
```

#### 3. Cron Graceful Shutdown

**File**: `internal/app/cron.go`

```go
func (c *CronScheduler) Start() error {
    logger.Info("Starting cron scheduler...")

    scheduler := InitializeCron(c.db, c.cfg)
    scheduler.Start()

    logger.Info("Cron scheduler started")

    // Wait for shutdown signal
    shutdown := make(chan os.Signal, 1)
    signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

    sig := <-shutdown
    logger.Info("Shutdown signal received", "signal", sig)

    // Stop scheduler and wait for running jobs
    ctx := scheduler.Stop()

    // Wait for jobs with timeout
    select {
    case <-ctx.Done():
        logger.Info("All jobs completed")
    case <-time.After(30 * time.Second):
        logger.Warn("Timeout waiting for jobs to complete")
    }

    // Close database
    if err := c.db.Close(); err != nil {
        logger.Error("Failed to close database", "error", err)
    }

    logger.Info("Cron scheduler stopped gracefully")
    return nil
}
```

#### 4. Testing Graceful Shutdown

**File**: `scripts/test-graceful-shutdown.sh`

```bash
#!/bin/bash

# Start server
./app api &
PID=$!

# Wait for startup
sleep 2

# Send SIGTERM
echo "Sending SIGTERM to process $PID"
kill -TERM $PID

# Wait for graceful shutdown
wait $PID
EXIT_CODE=$?

if [ $EXIT_CODE -eq 0 ]; then
    echo "✅ Graceful shutdown successful"
else
    echo "❌ Graceful shutdown failed with exit code $EXIT_CODE"
fi
```

**Files to Modify**:
- `internal/app/api.go` - Add graceful shutdown for API server
- `internal/app/worker.go` - Add graceful shutdown for worker
- `internal/app/cron.go` - Add graceful shutdown for cron
- `internal/delivery/consumer/order_consumer.go` - Add Shutdown() method
- `scripts/test-graceful-shutdown.sh` - Add test script

**Expected Behavior After Fix**:
```
# On receiving SIGTERM:
[INFO] Shutdown signal received signal=SIGTERM
[INFO] Finishing 3 in-flight requests
[INFO] Closing database connections
[INFO] Server stopped gracefully
```

---

### X. Dev/Prod Parity ✅ COMPLIANT

> **Twelve-Factor Principle**: Keep development, staging, and production as similar as possible

**Status**: ✅ **Fully Compliant**

**Evidence**:

**1. Time Gap** (Minimize delay between dev and deploy):
- ✅ Docker ensures same runtime environment
- ✅ CI/CD can deploy immediately after merge
- ✅ No manual build steps required

**2. Personnel Gap** (Developers deploy their own code):
- ✅ Docker Compose allows devs to run full stack locally
- ✅ Same commands work in dev and prod
- ✅ Infrastructure as code (docker-compose.yaml)

**3. Tools Gap** (Use same backing services):
```yaml
# docker-compose.yaml
services:
  postgres:
    image: postgres:15-alpine  # ✅ Same version as production

  rabbitmq:
    image: rabbitmq:3-management-alpine  # ✅ Same version

  api:
    build: .  # ✅ Same Docker image as production
```

**Same Environment**:
```dockerfile
# Dockerfile
FROM alpine:latest
# ✅ Production uses same base image
# ✅ Same binary runs in dev and prod
```

**Configuration Parity**:
```bash
# Development
docker-compose up

# Production (same image, different config)
docker run -e APP_DATABASE_HOST=prod-db go-clean-arch:v1.2.3
```

**Backing Service Parity**:
- ✅ Dev uses PostgreSQL 15 → Prod uses PostgreSQL 15 (not SQLite → PostgreSQL)
- ✅ Dev uses RabbitMQ 3 → Prod uses RabbitMQ 3
- ✅ No service substitution between environments

**Score**: 1.0/1.0

**Recommendation**: ✅ No action needed

---

### XI. Logs ✅ COMPLIANT

> **Twelve-Factor Principle**: Treat logs as event streams

**Status**: ✅ **Fully Compliant**

**Evidence**:
- ✅ Structured logging to stdout/stderr
- ✅ Application doesn't manage log files
- ✅ JSON format for machine parsing
- ✅ No log rotation in application
- ✅ Execution environment handles log aggregation

**Implementation**:
```go
// pkg/logger/logger.go:68-73
var handler slog.Handler
if cfg.Format == "json" {
    handler = slog.NewJSONHandler(os.Stdout, opts)  // ✅ Stdout
} else {
    handler = slog.NewTextHandler(os.Stdout, opts)  // ✅ Stdout
}
```

**Structured Logging**:
```go
// Example usage
logger.Info("Server started", "port", 8080, "mode", "production")
logger.ErrorContext(ctx, "Request failed", "error", err, "trace_id", traceID)

// Output (JSON format):
{
  "time": "2025-11-20T10:30:00Z",
  "level": "INFO",
  "msg": "Server started",
  "port": 8080,
  "mode": "production"
}
```

**Context-Aware Logging**:
```go
// pkg/logger/logger.go:86-92
func FromContext(ctx context.Context) *slog.Logger {
    if traceID, ok := ctx.Value(traceIDKey).(string); ok {
        return defaultLogger.With("trace_id", traceID)
    }
    return defaultLogger
}
// ✅ Automatic trace ID correlation
```

**HTTP Request Logging**:
```go
// pkg/middleware/logger.go
// Logs each request to stdout with structured data
logger.Info("Request completed",
    "method", c.Request.Method,
    "path", c.Request.URL.Path,
    "status", c.Writer.Status(),
    "duration_ms", duration.Milliseconds(),
    "trace_id", traceID,
)
```

**Docker Integration**:
```bash
# Logs accessible via Docker
docker logs go-clean-arch-api

# Production: Logs forwarded to aggregation service
# (CloudWatch, Datadog, Splunk, etc.)
```

**Score**: 1.0/1.0

**Recommendation**: ✅ No action needed

---

### XII. Admin Processes ⚠️ PARTIALLY COMPLIANT

> **Twelve-Factor Principle**: Run admin/management tasks as one-off processes

**Status**: ⚠️ **Partially Compliant** (50%)

**What's Good** (50%):
- ✅ Database migrations exist: `migrations/` directory
- ✅ SQL-based migrations (001_create_users_table.sql, etc.)
- ✅ Clean separation of migration files

**What's Missing** (50%):
- ❌ No standardized command to run migrations
- ❌ No `migrate` or `admin` CLI subcommand
- ❌ Manual migration execution required
- ❌ No one-off task runner framework
- ⚠️ Auto-migrate in config (not ideal for production)

**Current State**:
```toml
# configs/config.example.toml
[database]
auto_migrate = true  # ⚠️ Not suitable for production
```

**Migration Files**:
```bash
$ ls migrations/
001_create_users_table.sql
002_create_orders_table.sql
README.md
# ✅ Migrations exist but no runner
```

**Missing Functionality**:
```bash
# This should exist but doesn't:
$ ./app migrate up
$ ./app migrate down
$ ./app migrate status

# One-off tasks should be possible:
$ ./app run-task send-welcome-emails
$ ./app run-task cleanup-old-data
```

**Score**: 0.5/1.0

**Recommendation**: 🔧 **IMPROVE** (Priority: Low)

**Recommended Implementation**:

#### Option 1: Add Migrate Command (Recommended)

**File**: `cmd/app/cmd/migrate.go`

```go
package cmd

import (
    "database/sql"
    "fmt"
    "go-clean-arch/internal/app"

    "github.com/golang-migrate/migrate/v4"
    "github.com/golang-migrate/migrate/v4/database/postgres"
    _ "github.com/golang-migrate/migrate/v4/source/file"
    "github.com/spf13/cobra"
)

var migrateCmd = &cobra.Command{
    Use:   "migrate",
    Short: "Run database migrations",
    Long:  `Run database migrations as one-off admin process`,
}

var migrateUpCmd = &cobra.Command{
    Use:   "up",
    Short: "Apply all pending migrations",
    RunE:  runMigrateUp,
}

var migrateDownCmd = &cobra.Command{
    Use:   "down",
    Short: "Rollback last migration",
    RunE:  runMigrateDown,
}

var migrateStatusCmd = &cobra.Command{
    Use:   "status",
    Short: "Show migration status",
    RunE:  runMigrateStatus,
}

func init() {
    rootCmd.AddCommand(migrateCmd)
    migrateCmd.AddCommand(migrateUpCmd)
    migrateCmd.AddCommand(migrateDownCmd)
    migrateCmd.AddCommand(migrateStatusCmd)
}

func runMigrateUp(cmd *cobra.Command, args []string) error {
    m, err := getMigrator()
    if err != nil {
        return err
    }

    if err := m.Up(); err != nil && err != migrate.ErrNoChange {
        return fmt.Errorf("migration failed: %w", err)
    }

    fmt.Println("✅ Migrations applied successfully")
    return nil
}

func runMigrateDown(cmd *cobra.Command, args []string) error {
    m, err := getMigrator()
    if err != nil {
        return err
    }

    if err := m.Steps(-1); err != nil {
        return fmt.Errorf("rollback failed: %w", err)
    }

    fmt.Println("✅ Rollback completed")
    return nil
}

func runMigrateStatus(cmd *cobra.Command, args []string) error {
    m, err := getMigrator()
    if err != nil {
        return err
    }

    version, dirty, err := m.Version()
    if err != nil {
        return err
    }

    fmt.Printf("Current version: %d\n", version)
    fmt.Printf("Dirty: %v\n", dirty)
    return nil
}

func getMigrator() (*migrate.Migrate, error) {
    cfg := getConfig()

    // Connect to database
    dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
        cfg.Database.User,
        cfg.Database.Password,
        cfg.Database.Host,
        cfg.Database.Port,
        cfg.Database.DBName,
        cfg.Database.SSLMode,
    )

    db, err := sql.Open("postgres", dsn)
    if err != nil {
        return nil, err
    }

    driver, err := postgres.WithInstance(db, &postgres.Config{})
    if err != nil {
        return nil, err
    }

    m, err := migrate.NewWithDatabaseInstance(
        "file://migrations",
        "postgres",
        driver,
    )
    if err != nil {
        return nil, err
    }

    return m, nil
}
```

**Usage**:
```bash
# Apply migrations
./app migrate up

# Rollback last migration
./app migrate down

# Check migration status
./app migrate status
```

#### Option 2: Generic Task Runner

**File**: `cmd/app/cmd/task.go`

```go
package cmd

import (
    "fmt"
    "go-clean-arch/internal/tasks"

    "github.com/spf13/cobra"
)

var taskCmd = &cobra.Command{
    Use:   "task [task-name]",
    Short: "Run one-off admin tasks",
    Long:  `Run administrative tasks as one-off processes`,
    Args:  cobra.ExactArgs(1),
    RunE:  runTask,
}

func init() {
    rootCmd.AddCommand(taskCmd)
}

func runTask(cmd *cobra.Command, args []string) error {
    taskName := args[0]

    db, err := getDB()
    if err != nil {
        return err
    }

    cfg := getConfig()

    // Registry of available tasks
    registry := tasks.NewRegistry(db, cfg)

    task, exists := registry.Get(taskName)
    if !exists {
        return fmt.Errorf("unknown task: %s\n\nAvailable tasks:\n%s",
            taskName, registry.List())
    }

    fmt.Printf("Running task: %s\n", taskName)
    if err := task.Execute(); err != nil {
        return fmt.Errorf("task failed: %w", err)
    }

    fmt.Println("✅ Task completed successfully")
    return nil
}
```

**Usage**:
```bash
# Run one-off tasks
./app task send-welcome-emails
./app task cleanup-old-data
./app task recalculate-metrics
```

**Files to Create**:
- `cmd/app/cmd/migrate.go` - Migration commands
- `cmd/app/cmd/task.go` - Generic task runner
- `internal/tasks/registry.go` - Task registry
- `internal/tasks/examples.go` - Example tasks

**Update go.mod**:
```go
require (
    github.com/golang-migrate/migrate/v4 v4.17.0
)
```

**Remove Auto-Migrate**:
```go
// internal/app/database.go
// Remove auto-migration code
// Force explicit migration via CLI
```

---

## Priority Action Items

### 🔴 Critical (Must Fix)

1. **Factor IX - Disposability**
   - Implement graceful shutdown with signal handling
   - Add cleanup for database connections
   - Ensure in-flight requests complete before shutdown
   - **Impact**: High - Prevents data loss and improves reliability
   - **Effort**: Medium - 2-3 hours of development
   - **Files**: `internal/app/api.go`, `internal/app/worker.go`, `internal/app/cron.go`

### 🟡 Medium Priority (Should Fix)

2. **Factor III - Config**
   - Make config files optional
   - Support environment-only configuration
   - Remove config files from Docker image
   - **Impact**: Medium - Better containerization practices
   - **Effort**: Low - 1-2 hours
   - **Files**: `pkg/config/config.go`, `Dockerfile`, `cmd/app/cmd/root.go`

### 🟢 Low Priority (Nice to Have)

3. **Factor XII - Admin Processes**
   - Add `migrate` CLI command
   - Implement task runner framework
   - Remove auto_migrate from production config
   - **Impact**: Low - Better operational practices
   - **Effort**: Medium - 3-4 hours
   - **Files**: `cmd/app/cmd/migrate.go`, `cmd/app/cmd/task.go`

---

## Implementation Roadmap

### Phase 1: Critical Fixes (Week 1)
- [ ] Implement graceful shutdown for API server
- [ ] Implement graceful shutdown for Worker
- [ ] Implement graceful shutdown for Cron
- [ ] Add signal handling tests
- [ ] Update documentation

### Phase 2: Config Improvements (Week 2)
- [ ] Make config file optional in `pkg/config/config.go`
- [ ] Update Dockerfile to not bundle configs
- [ ] Add environment variable documentation
- [ ] Test environment-only deployment

### Phase 3: Admin Processes (Week 3)
- [ ] Add `migrate` subcommand
- [ ] Create task runner framework
- [ ] Remove auto_migrate from production
- [ ] Add migration documentation

---

## Testing Checklist

### Disposability Testing
- [ ] Server gracefully shuts down on SIGTERM
- [ ] Server gracefully shuts down on SIGINT (Ctrl+C)
- [ ] In-flight requests complete before shutdown
- [ ] Database connections properly closed
- [ ] Worker completes current message before shutdown
- [ ] Shutdown completes within 30 seconds

### Config Testing
- [ ] Application runs with only environment variables (no config file)
- [ ] Environment variables override config file values
- [ ] Application fails gracefully with invalid config
- [ ] Docker container runs without bundled config files

### Admin Process Testing
- [ ] Migrations apply successfully
- [ ] Migration status shows correct version
- [ ] Migrations can be rolled back
- [ ] One-off tasks execute in same environment as main app

---

## Conclusion

The `go-clean-arch` repository demonstrates **strong alignment** with the Twelve-Factor App methodology, scoring **9.5 out of 12** factors. The architecture is well-designed with clean separation of concerns and modern practices.

### Strengths
- ✅ Excellent dependency management
- ✅ Stateless process architecture
- ✅ Strong dev/prod parity
- ✅ Proper structured logging
- ✅ Multiple process types for scaling

### Critical Gap
- ❌ **Missing graceful shutdown** - This is the only critical issue that should be addressed immediately

### Recommendations
1. **Implement graceful shutdown (High Priority)** - Essential for production reliability
2. **Improve config management (Medium Priority)** - Better aligns with twelve-factor principles
3. **Add admin process framework (Low Priority)** - Operational convenience

With the recommended changes, this application would achieve **11.7/12** compliance, representing an **excellent** implementation of twelve-factor principles.

---

**Assessment Date**: 2025-11-20
**Assessor**: Claude (AI Assistant)
**Methodology**: [Twelve-Factor App](https://12factor.net/) v1.0
