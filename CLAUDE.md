# CLAUDE.md

**For LLM Context** - This file provides quick reference for AI assistants working with this codebase.

For human developers, see:
- User documentation: [README.md](README.md)
- Coding standards: [CODING_STANDARDS.md](CODING_STANDARDS.md)

> **Sync Reminder**: This repo contains [docs/GO_CLEAN_ARCHITECTURE_REFERENCE.md](docs/GO_CLEAN_ARCHITECTURE_REFERENCE.md), a standalone architecture guide for AI assistants. When modifying architectural patterns, coding standards, or best practices in this repo, **also update that reference file** to keep them in sync.

---

## Quick Commands

### Development
```bash
# Run different modes
go run cmd/app/main.go api              # Start API server on :8080
go run cmd/app/main.go worker           # Start message queue worker
go run cmd/app/main.go cron             # Start cron scheduler
go run cmd/app/main.go micro            # Start microservice mode on :8081

# Makefile shortcuts
make run-api
make run-worker
make run-cron
```

### Testing
```bash
make test                # Run all tests with race detector
make test-coverage       # Generate HTML coverage report
go test ./internal/...   # Test specific package
```

### Code Quality
```bash
make lint                # Run golangci-lint (auto-installs if missing)
make fmt                 # Format code with go fmt and goimports
make vet                 # Run go vet
```

### Building
```bash
make build               # Build for current platform → bin/go-clean-arch
make build-all           # Cross-compile for multiple platforms
```

### Dependency Injection (Wire)
```bash
# After modifying internal/app/wire.go, regenerate Wire code:
wire gen ./internal/app
```

### Database Migrations (Atlas)
```bash
# Apply all pending migrations
make migrate-up
# or
./app migrate up

# Show migration status
make migrate-status

# Create a new migration
make migrate-create NAME="add_products_table"

# Rollback last migration
make migrate-down

# Validate migration files
make migrate-validate

# Install Atlas CLI (if not installed)
make atlas-install
```

---

## Critical Context ⚠️

### Current Status
✅ **Security**: Password hashing uses bcrypt (`internal/domain/user.go:119-136`)
✅ **Architecture**: All runtime modes include complete dependency sets
✅ **Lifecycle**: Graceful shutdown implemented for all modes (API, Worker, Cron)
✅ **Domain**: Entity reconstruction with `ReconstructUser()` properly implemented
✅ **Database**: All repositories migrated to sqlc + sqlx+Squirrel hybrid pattern
✅ **Testing**: High test coverage achieved
  - Domain layer: **93.8%** coverage ✅
  - Use case layer: **86.9%** coverage (near 90% target)
  - Tests use mocks for clean unit testing

### Known Issues
1. **TESTING**: Handler layer test coverage is low (29.5%)
   - `user_handler_test.go` exists with comprehensive tests
   - Missing tests for `order_handler.go` and `health_handler.go`
   - Current coverage: 29.5%, Target: 70%+

2. **ORDER CHECKOUT**: One failing test in order_interactor_test.go
   - `TestCheckout_OrderNotPending` - status transition validation needs review
   - Non-blocking issue, does not affect production code

### Tech Stack
- **Go**: 1.22+ (tested with Go 1.23.x)
- **HTTP Framework**: Gin
- **Database Access**:
  - **sqlc** (primary - type-safe SQL queries for static queries)
  - **sqlx + Squirrel** (auxiliary - dynamic query building)
  - Supports: PostgreSQL, MySQL, SQLite
- **Database Migrations**: Atlas (versioned migrations with validation)
- **Dependency Injection**: Wire (compile-time)
- **Configuration**: Viper (YAML + env vars)
- **Logging**: slog (structured logging with trace IDs)
- **Microservices**: go-micro v5
- **Message Queue**: RabbitMQ (optional)
- **Cron**: robfig/cron

---

## Architecture Quick Reference

### Clean Architecture Layers (Dependency Flow: Outer → Inner)

```
┌─────────────────────────────────────────┐
│  Delivery (internal/delivery/)          │  ← HTTP, Worker, Cron, Microservice
│  - Converts DTOs ↔ Domain entities      │     Depends on: Use Case
│  - Request validation, error mapping    │
│  Examples:                              │
│    internal/delivery/http/handler/      │
│    internal/delivery/consumer/          │
│    internal/delivery/job/               │
│    internal/delivery/micro/handler/     │
└──────────────┬──────────────────────────┘
               │
               ▼
┌─────────────────────────────────────────┐
│  Use Cases (internal/usecase/)          │  ← Application business logic
│  - Orchestrates domain logic            │     Depends on: Domain
│  - Defines ports (interfaces)           │     Defines: Repository/Gateway interfaces
│  Examples:                              │
│    internal/usecase/user_interactor.go  │
│    internal/usecase/port/               │
└──────────────┬──────────────────────────┘
               │
               ▼
┌─────────────────────────────────────────┐
│  Domain (internal/domain/)              │  ← Pure business entities & rules
│  - NO external dependencies             │     Depends on: Nothing (innermost)
│  - Rich domain models, not anemic       │
│  Examples:                              │
│    internal/domain/user.go              │
│    internal/domain/order.go             │
└─────────────────────────────────────────┘

               ▲
               │ implements port interfaces
┌─────────────────────────────────────────┐
│  Adapters (internal/adapter/)           │  ← Repository & Gateway implementations
│  - Implements Use Case port interfaces  │     Implements: port.UserRepository, etc.
│  - Converts Domain ↔ DB models          │     Depends on: Domain (for conversion)
│  - External API integrations            │
│  Examples:                              │
│    internal/adapter/repository/         │
│    internal/adapter/gateway/            │
└─────────────────────────────────────────┘
```

**Note**: Adapters implement the interfaces (ports) defined in `usecase/port/`. They are injected via Wire DI, allowing the Use Case layer to remain independent of specific implementations.

### Key Architectural Decisions

1. **Interfaces in `internal/usecase/port/`** (not domain)
   - Use cases define the interfaces they need
   - Keeps domain pure and focused on business rules

2. **Single binary, multiple runtime modes**
   - Mode selected via Cobra subcommands: `api|worker|cron|micro|migrate`
   - Shared business logic across all modes
   - Cobra CLI framework handles commands

3. **Pure domain entities** (no infrastructure tags)
   - Domain entities have no JSON/DB tags
   - DTOs handle JSON serialization (in delivery layer)
   - DB models handle sqlx mappings (in adapter layer)

4. **Wire dependency injection** (compile-time)
   - `internal/app/wire.go` defines provider sets
   - `wire_gen.go` is auto-generated (DON'T edit manually)
   - Run `wire gen ./internal/app` after changes

5. **Twelve-Factor App compliant**
   - Config via environment variables (Factor III)
   - Config files are OPTIONAL (for local dev only)
   - Graceful shutdown required (Factor IX)
   - Logs to stdout (Factor XI)

---

## Adding New Features

### Standard 7-Step Process

**Example**: Adding a "Product" entity

**Step 1: Domain Layer**
```go
// internal/domain/product.go
type Product struct {
    ID        int64
    Name      string
    Price     int64
    CreatedAt time.Time
    UpdatedAt time.Time
}

func NewProduct(name string, price int64) (*Product, error) {
    // Validation and business rules
}
```

**Step 2: Define Repository Interface**
```go
// internal/usecase/port/product_repository.go
type ProductRepository interface {
    Save(ctx context.Context, product *domain.Product) error
    FindByID(ctx context.Context, id int64) (*domain.Product, error)
}
```

**Step 3: Implement Repository**

For static queries (recommended - use sqlc):
```sql
-- internal/adapter/repository/queries/product.sql
-- name: GetProduct :one
SELECT * FROM products WHERE id = ? LIMIT 1;

-- name: CreateProduct :exec
INSERT INTO products (name, price, created_at, updated_at)
VALUES (?, ?, ?, ?);
```

After adding queries, regenerate sqlc code:
```bash
make sqlc-gen
# or
sqlc generate
```

For dynamic queries (use sqlx + Squirrel):
```go
// internal/adapter/repository/product_repository_sqlc.go
import (
    "github.com/Masterminds/squirrel"
    "github.com/jmoiron/sqlx"
    "go-clean-arch/internal/adapter/repository/sqlcgen"
)

var mysql = squirrel.StatementBuilder.PlaceholderFormat(squirrel.Question)

type productRepositorySQLC struct {
    db      *sqlx.DB
    queries *sqlcgen.Queries
}

func NewProductRepository(db *sqlx.DB) port.ProductRepository {
    return &productRepositorySQLC{
        db:      db,
        queries: sqlcgen.New(db.DB),
    }
}

// Use sqlc for static queries
func (r *productRepositorySQLC) FindByID(ctx context.Context, id int64) (*domain.Product, error) {
    product, err := r.queries.GetProduct(ctx, id)
    // ... convert to domain
}

// Use Squirrel for dynamic queries (e.g., search with filters)
func (r *productRepositorySQLC) Search(ctx context.Context, filters ProductFilters) ([]*domain.Product, error) {
    query := mysql.Select("*").From("products")
    if filters.MinPrice > 0 {
        query = query.Where(squirrel.GtOrEq{"price": filters.MinPrice})
    }
    // ... build and execute dynamic query
}
```

**Step 4: Create Use Case**
```go
// internal/usecase/interfaces.go - Add interface
type ProductUsecase interface {
    CreateProduct(ctx context.Context, name string, price int64) (*domain.Product, error)
}
```

```go
// internal/usecase/product_interactor.go
type ProductInteractor struct {
    productRepo port.ProductRepository
}

func NewProductInteractor(productRepo port.ProductRepository) *ProductInteractor { ... }
func (i *ProductInteractor) CreateProduct(ctx context.Context, name string, price int64) (*domain.Product, error) { ... }
```

**Step 5: Add HTTP Handler**
```go
// internal/delivery/http/handler/product_dto.go
type CreateProductRequest struct {
    Name  string `json:"name" binding:"required"`
    Price int64  `json:"price" binding:"required,min=0"`
}

type ProductResponse struct {
    ID        int64     `json:"id"`
    Name      string    `json:"name"`
    Price     int64     `json:"price"`
    CreatedAt time.Time `json:"created_at"`
}
```

```go
// internal/delivery/http/handler/product_handler.go
type ProductHandler struct {
    productUC usecase.ProductUsecase
}

func NewProductHandler(productUC usecase.ProductUsecase) *ProductHandler { ... }
func (h *ProductHandler) Create(c *gin.Context) { ... }
```

**Step 6: Register Routes**
```go
// internal/delivery/http/router.go
v1 := router.Group("/api/v1")
{
    v1.POST("/products", productHandler.Create)
    v1.GET("/products/:id", productHandler.GetByID)
}
```

**Step 7: Wire Dependencies**
```go
// internal/app/wire.go

// Add to RepositorySet:
var RepositorySet = wire.NewSet(
    repository.NewUserRepository,
    repository.NewProductRepository,  // ← Add this
    wire.Bind(new(port.UserRepository), new(*repository.userRepositorySQLC)),
    wire.Bind(new(port.ProductRepository), new(*repository.productRepositorySQLC)),  // ← Add this
)

// Add to UseCaseSet:
var UseCaseSet = wire.NewSet(
    usecase.NewUserInteractor,
    usecase.NewProductInteractor,  // ← Add this
    wire.Bind(new(usecase.UserUsecase), new(*usecase.UserInteractor)),
    wire.Bind(new(usecase.ProductUsecase), new(*usecase.ProductInteractor)),  // ← Add this
)

// Add to HandlerSet:
var HandlerSet = wire.NewSet(
    handler.NewUserHandler,
    handler.NewProductHandler,  // ← Add this
)
```

Then run:
```bash
wire gen ./internal/app
```

---

## File Naming Conventions

| Component | File Path | Naming Pattern |
|-----------|-----------|----------------|
| Domain Entity | `internal/domain/{entity}.go` | `user.go`, `order.go` |
| Domain Errors | `internal/domain/errors.go` | Single file for all errors |
| Repository Interface | `internal/usecase/port/{entity}_repository.go` | `user_repository.go` |
| Gateway Interface | `internal/usecase/port/{service}_gateway.go` | `payment_gateway.go` |
| Use Case Interface | `internal/usecase/interfaces.go` | Single file for all interfaces |
| Use Case Implementation | `internal/usecase/{entity}_interactor.go` | `user_interactor.go` |
| Repository Implementation | `internal/adapter/repository/{entity}_repository_sqlc.go` | `user_repository_sqlc.go` |
| DB Model (for sqlx) | `internal/adapter/repository/{entity}_model.go` | `user_model.go`, `order_model.go` |
| sqlc Queries | `internal/adapter/repository/queries/{entity}.sql` | `user.sql`, `order.sql` |
| sqlc Generated Code | `internal/adapter/repository/sqlcgen/*.go` | Auto-generated by sqlc |
| Gateway Implementation | `internal/adapter/gateway/{service}_gateway.go` | `stripe_gateway.go` |
| HTTP Handler | `internal/delivery/http/handler/{entity}_handler.go` | `user_handler.go` |
| HTTP DTO | `internal/delivery/http/handler/{entity}_dto.go` | `user_dto.go` |
| Worker Consumer | `internal/delivery/consumer/{entity}_consumer.go` | `order_consumer.go` |
| Cron Job | `internal/delivery/job/{job_name}_job.go` | `daily_report_job.go` |
| Microservice Handler | `internal/delivery/micro/handler/{entity}_service.go` | `user_service.go` |
| Test File | `{filename}_test.go` | `user_test.go` |

---

## Common Pitfalls

### Architecture Violations
- ❌ **DON'T** import `adapter` or `delivery` in `usecase` layer
- ❌ **DON'T** import any other layer in `domain` layer
- ❌ **DON'T** add JSON/DB tags to domain entities
- ❌ **DON'T** put business logic in handlers (keep them thin)
- ❌ **DON'T** access repositories directly in handlers

### Wire DI
- ❌ **DON'T** edit `wire_gen.go` manually (it's auto-generated)
- ⚠️ **REMEMBER** to run `wire gen ./internal/app` after changing `wire.go`
- ⚠️ **REMEMBER** to add interface bindings: `wire.Bind(new(Interface), new(*Implementation))`

### Configuration
- ❌ **DON'T** require config files in production
- ✅ **DO** support environment-only deployment (Twelve-Factor)
- ✅ **DO** use `APP_` prefix for environment variables
- ✅ **DO** provide `config.example.toml` for local dev

### Security
- ❌ **DON'T** commit secrets in config files
- ❌ **DON'T** log sensitive data (passwords, tokens)

### Context Propagation
- ✅ **DO** accept `context.Context` as first parameter in use cases
- ✅ **DO** propagate context to repositories
- ✅ **DO** use `logger.InfoContext(ctx, ...)` for trace ID propagation

### Error Handling
- ✅ **DO** define domain errors in `internal/domain/errors.go`
- ✅ **DO** wrap errors with context: `fmt.Errorf("operation failed: %w", err)`
- ✅ **DO** use `errors.Is()` for error checking

---

## Testing Guidelines

### Coverage Targets
- **Domain layer**: 90%+ coverage (focus on business logic)
- **Use case layer**: 90%+ coverage (mock repositories)
- **Repository layer**: 80%+ coverage (integration tests)
- **Handler layer**: 70%+ coverage (mock use cases)

### Testing Strategy
- **Domain**: Test business logic in isolation (no dependencies)
- **Use Case**: Mock repositories (interfaces in `usecase/port/`)
- **Repository**: Integration tests with test database
- **Handler**: Mock use cases, test request/response mapping

### Current Status
- **Domain layer**: 93.8% coverage ✅ (meets 90%+ target)
- **Use case layer**: 86.9% coverage ⚠️ (near 90% target, has 1 failing test)
- **Handler layer**: 29.5% coverage ❌ (far below 70%+ target)
- **Repository layer**: Not tested yet ⚠️ (target: 80%+)

---

## Configuration

### Environment Variables (Recommended)
```bash
# All config via environment variables (Twelve-Factor compliant)
export APP_SERVER_PORT=8080
export APP_DATABASE_HOST=localhost
export APP_DATABASE_USER=postgres
export APP_DATABASE_PASSWORD=secret
export APP_LOGGER_LEVEL=info

# Run without config file
./app api
```

### Config File (Optional, for local dev)
```bash
# Copy example config
cp configs/config.example.toml configs/config.toml

# Run with config file
./app api --config=configs/config.toml
```

### Naming Convention
- Prefix: `APP_`
- Nested keys use underscore: `database.host` → `APP_DATABASE_HOST`
- All uppercase

---

## Logging

### Structured Logging with Trace IDs
```go
// Use context-aware logging for trace ID propagation
logger.InfoContext(ctx, "Creating user", "email", email)
logger.ErrorContext(ctx, "Failed to save", "error", err)
```

### Log Levels
- **DEBUG**: Detailed debugging information
- **INFO**: General informational messages
- **WARN**: Warning messages
- **ERROR**: Error messages

---

## Runtime Modes

| Mode | Command | Port | Purpose |
|------|---------|------|---------|
| API | `./app api` | 8080 | HTTP REST API server |
| Worker | `./app worker` | - | Message queue consumer |
| Cron | `./app cron` | - | Scheduled jobs |
| Microservice | `./app micro` | 8081 | go-micro RPC service |
| Migrate | `./app migrate` | - | Database migration admin process |

### Mode-Specific Notes

**API Mode**:
- Serves REST API endpoints
- Health check: `GET /health`

**Worker Mode**:
- Consumes messages from RabbitMQ
- Currently skeleton implementation

**Cron Mode**:
- Runs scheduled jobs using robfig/cron
- Daily report job configured

**Microservice Mode**:
- Uses go-micro v5 for RPC communication
- Service discovery and load balancing
- HTTP transport available by default

**Migrate Mode** (Admin Process):
- Database migration management using Atlas
- Supports: up, down, status, create, validate
- Twelve-Factor App compliant (Factor XII)
- See [Database Migrations](#database-migrations-atlas) section

---

## Data Transformation Strategy

### Transformation at Layer Boundaries

```
HTTP Request (JSON)
    ↓ [Handler Layer]
DTO (with JSON tags)
    ↓ [Handler converts to Domain]
Domain Entity (pure, no tags)
    ↓ [Use Case Layer]
Domain Entity (unchanged)
    ↓ [Repository converts to Model]
DB Model (with db tags for sqlx)
    ↓ [Database]
```

### Key Points
- **Domain entities** remain pure (no JSON/DB tags)
- **DTOs** handle JSON serialization (in delivery layer)
- **DB models** handle sqlx mappings with `db` tags (in adapter layer)
- **Squirrel** builds SQL queries in a type-safe, fluent manner
- Each layer owns its transformation logic

---

## References

For detailed information, see:
- **Full coding standards**: [CODING_STANDARDS.md](CODING_STANDARDS.md)
- **User documentation**: [README.md](README.md)
- **Twelve-Factor compliance**: [CODING_STANDARDS.md#twelve-factor-app-compliance](CODING_STANDARDS.md#twelve-factor-app-compliance)
