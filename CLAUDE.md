# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

### Development
```bash
# Run different modes
go run cmd/app/main.go --mode=api       # Start API server on :8080
go run cmd/app/main.go --mode=worker    # Start message queue worker
go run cmd/app/main.go --mode=cron      # Start cron scheduler

# Or use Makefile shortcuts
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
make build-all           # Cross-compile for Linux, macOS (amd64/arm64), Windows
```

### Dependency Injection (Wire)
```bash
# After modifying internal/app/wire.go, regenerate Wire code:
wire gen ./internal/app
```

## Architecture

### Clean Architecture Layers (Dependency Flow: Outer → Inner)

```
┌─────────────────────────────────────────┐
│  Delivery (cmd/, internal/delivery/)    │  ← HTTP handlers, Workers, Cron jobs
│  - Converts DTOs ↔ Domain entities      │
│  - Request validation, error mapping    │
└──────────────┬──────────────────────────┘
               │
┌──────────────▼──────────────────────────┐
│  Use Cases (internal/usecase/)          │  ← Application business logic
│  - Orchestrates domain logic            │
│  - Defines ports (interfaces)           │
└──────────────┬──────────────────────────┘
               │
┌──────────────▼──────────────────────────┐
│  Adapters (internal/adapter/)           │  ← Repository & Gateway implementations
│  - Converts Domain ↔ DB models          │
│  - External API integrations            │
└──────────────┬──────────────────────────┘
               │
┌──────────────▼──────────────────────────┐
│  Domain (internal/domain/)              │  ← Pure business entities & rules
│  - NO external dependencies             │
│  - Rich domain models, not anemic       │
└─────────────────────────────────────────┘
```

### Key Architectural Decisions

**1. Interface Location: `internal/usecase/port/`**
- Use cases define the interfaces they need (ports)
- Repositories and gateways implement these interfaces
- This is pragmatic for most applications and keeps domain pure

**2. Single Binary, Multiple Runtime Modes**
- One codebase compiles to one binary
- Mode selected via CLI flag: `--mode=api|worker|cron`
- Shared business logic across all modes
- Cobra CLI framework handles commands

**3. Data Transformation Strategy**
- **Domain entities** remain pure (no JSON/DB tags)
- **HTTP layer**: DTOs ↔ Domain (in handlers)
- **Repository layer**: Domain ↔ DB models (in repository)
- Each layer owns its transformation logic

**4. Wire Dependency Injection**
- `internal/app/wire.go` defines provider sets
- `wire_gen.go` is auto-generated (don't edit manually)
- Different injector functions for each mode:
  - `InitializeAPIRouter()` - API mode
  - `InitializeWorker()` - Worker mode
  - `InitializeCronScheduler()` - Cron mode
- NOTE: API mode currently missing `GatewaySet` (see TODO.md)

**5. Structured Logging with Trace ID**
- Uses Go's `log/slog` for structured logging
- Trace IDs propagate through requests via context
- Middleware auto-injects trace IDs from `X-Trace-ID` header or generates UUID
- Use context-aware logging: `logger.InfoContext(ctx, "msg", "key", value)`

## Adding New Features

### Example: Adding a New Entity (e.g., "Product")

**Step 1: Domain Layer**
```go
// internal/domain/product.go
type Product struct {
    id          int64
    name        string
    price       int64
    createdAt   time.Time
}
// Add business logic methods (validation, state transitions)
```

**Step 2: Define Repository Interface**
```go
// internal/usecase/port/product_repository.go
type ProductRepository interface {
    Create(product *domain.Product) error
    FindByID(id int64) (*domain.Product, error)
}
```

**Step 3: Implement Repository**
```go
// internal/adapter/repository/product_model.go
type ProductModel struct {
    ID        int64     `gorm:"primaryKey"`
    Name      string    `gorm:"not null"`
    Price     int64     `gorm:"not null"`
    CreatedAt time.Time `gorm:"autoCreateTime"`
}
// Add conversion methods: ToModel(), ToDomain()
```

```go
// internal/adapter/repository/product_repository_gorm.go
type ProductRepositoryGorm struct {
    db *gorm.DB
}
// Implement ProductRepository interface
```

**Step 4: Create Use Case**
```go
// internal/usecase/interfaces.go - Add interface
type ProductUsecase interface {
    CreateProduct(name string, price int64) (*domain.Product, error)
}
```

```go
// internal/usecase/product_interactor.go
type ProductInteractor struct {
    productRepo port.ProductRepository
}
// Implement ProductUsecase methods
```

**Step 5: Add HTTP Handler**
```go
// internal/delivery/http/handler/product_dto.go
type CreateProductRequest struct {
    Name  string `json:"name" binding:"required"`
    Price int64  `json:"price" binding:"required,min=0"`
}
```

```go
// internal/delivery/http/handler/product_handler.go
type ProductHandler struct {
    productUC usecase.ProductUsecase
}
// Implement handlers (Create, Get, List, etc.)
```

**Step 6: Register Routes**
```go
// internal/delivery/http/router.go
v1.POST("/products", productHandler.Create)
v1.GET("/products/:id", productHandler.GetByID)
```

**Step 7: Wire Dependencies**
```go
// internal/app/wire.go
// Add to RepositorySet:
repository.NewProductRepository,

// Add to UseCaseSet:
usecase.NewProductInteractor,
wire.Bind(new(usecase.ProductUsecase), new(*usecase.ProductInteractor)),

// Add to HandlerSet:
handler.NewProductHandler,
```

Then run `wire gen ./internal/app` to regenerate Wire code.

## Configuration

- **Config file**: `configs/config.yaml` (NOTE: There's inconsistency between `.yaml` and `.toml` - currently using YAML)
- **Viper** loads config with environment variable override support
- **Environment variables**: Override config with format `APP_DATABASE_HOST=localhost`
- **Config structure**: See `pkg/config/config.go`

### Database Configuration
- Supports PostgreSQL, MySQL, SQLite
- Auto-migration enabled by default in config
- Connection pooling hardcoded in `internal/adapter/repository/db.go:68-70` (should be configurable)

## Critical Issues (See TODO.md for full list)

### Security
⚠️ **CRITICAL**: Password hashing is insecure (`internal/domain/user.go:67-74`)
- Currently uses `"hashed_" + password` prefix
- Must replace with bcrypt/argon2 before production

### Missing in API Mode
- `GatewaySet` not included in `InitializeAPIRouter()` (`internal/app/wire.go:66-75`)
- PaymentGateway unavailable in API mode
- Worker and Cron modes correctly include it

### Domain Entity Reconstruction
- Repositories cannot properly hydrate User entities with password hash
- Need `ReconstructUser()` factory method to maintain encapsulation

## Testing Strategy

When writing tests:
- **Domain layer**: Test business logic in isolation
- **Use case layer**: Mock repositories (interfaces in `usecase/port/`)
- **Repository layer**: Integration tests with test database
- **Handler layer**: Mock use cases, test request/response mapping

Current status: No tests exist yet (see TODO.md)

## Common Patterns

### Error Handling
Current: Uses `errors.New()` and basic error wrapping
Recommended: Define domain error types (e.g., `ErrUserNotFound`, `ErrEmailAlreadyExists`)

### Context Propagation
- Use cases should accept `context.Context` as first parameter
- Enables timeout, cancellation, and trace ID propagation
- Currently missing from interface signatures

### Logging
```go
// In handlers and use cases
logger.InfoContext(ctx, "Creating user", "email", email)
logger.ErrorContext(ctx, "Failed to save", "error", err)
```

### DTOs vs Domain
- DTOs have JSON binding tags (Gin validation)
- Domain entities are tag-free
- Always convert at layer boundaries:
  ```go
  domainUser := dto.ToDomain()      // In handler
  response := dto.FromDomain(user)  // In handler
  ```

## Known Gaps

1. No graceful shutdown implementation
2. No custom error types (all errors are generic)
3. OrderHandler not implemented (routes commented out)
4. Message queue consumer skeleton only
5. No database migrations tooling (golang-migrate/goose)
6. No OpenAPI/Swagger documentation
7. Health check endpoint doesn't verify dependencies

## Runtime Modes

### API Mode (`--mode=api`)
- Starts Gin HTTP server on configured port
- Serves REST API endpoints
- Health check: `GET /health`

### Worker Mode (`--mode=worker`)
- Consumes messages from RabbitMQ
- Currently skeleton implementation
- Processes order events

### Cron Mode (`--mode=cron`)
- Runs scheduled jobs using robfig/cron
- Daily report job configured
- Currently generates placeholder reports
