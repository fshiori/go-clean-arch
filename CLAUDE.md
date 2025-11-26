# CLAUDE.md

**For LLM Context** - This file provides quick reference for AI assistants working with this codebase.

For human developers, see:
- User documentation: [README.md](README.md)
- Coding standards: [CODING_STANDARDS.md](CODING_STANDARDS.md)

---

## Quick Commands

### Development
```bash
# Run different modes
go run cmd/app/main.go --mode=api       # Start API server on :8080
go run cmd/app/main.go --mode=worker    # Start message queue worker
go run cmd/app/main.go --mode=cron      # Start cron scheduler
./app micro                             # Start microservice mode on :8081

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

---

## Critical Context ⚠️

### Known Issues
1. **SECURITY**: Password hashing is insecure (`internal/domain/user.go:67-74`)
   - Currently uses `"hashed_" + password` prefix
   - Must replace with bcrypt/argon2 before production

2. **ARCHITECTURE**: GatewaySet missing in API mode
   - `InitializeAPIRouter()` doesn't include `GatewaySet` (`internal/app/wire.go:66-75`)
   - PaymentGateway unavailable in API mode
   - Worker and Cron modes correctly include it

3. **DOMAIN**: Entity reconstruction issue
   - Repositories cannot properly hydrate User entities with password hash
   - Need `ReconstructUser()` factory method to maintain encapsulation

4. **LIFECYCLE**: No graceful shutdown implemented yet
   - Must implement signal handling for SIGTERM/SIGINT
   - Critical for Twelve-Factor compliance (Factor IX)

5. **TESTING**: No tests exist yet
   - Need to implement tests for all layers

### Tech Stack
- **Go**: 1.24+
- **HTTP Framework**: Gin
- **ORM**: GORM (PostgreSQL, MySQL, SQLite)
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
│  - Converts DTOs ↔ Domain entities      │
│  - Request validation, error mapping    │
│  Examples:                              │
│    internal/delivery/http/handler/      │
│    internal/delivery/consumer/          │
│    internal/delivery/job/               │
│    internal/delivery/micro/handler/     │
└──────────────┬──────────────────────────┘
               │
┌──────────────▼──────────────────────────┐
│  Use Cases (internal/usecase/)          │  ← Application business logic
│  - Orchestrates domain logic            │
│  - Defines ports (interfaces)           │
│  Examples:                              │
│    internal/usecase/user_interactor.go  │
│    internal/usecase/port/               │
└──────────────┬──────────────────────────┘
               │
┌──────────────▼──────────────────────────┐
│  Adapters (internal/adapter/)           │  ← Repository & Gateway implementations
│  - Converts Domain ↔ DB models          │
│  - External API integrations            │
│  Examples:                              │
│    internal/adapter/repository/         │
│    internal/adapter/gateway/            │
└──────────────┬──────────────────────────┘
               │
┌──────────────▼──────────────────────────┐
│  Domain (internal/domain/)              │  ← Pure business entities & rules
│  - NO external dependencies             │
│  - Rich domain models, not anemic       │
│  Examples:                              │
│    internal/domain/user.go              │
│    internal/domain/order.go             │
└─────────────────────────────────────────┘
```

### Key Architectural Decisions

1. **Interfaces in `internal/usecase/port/`** (not domain)
   - Use cases define the interfaces they need
   - Keeps domain pure and focused on business rules

2. **Single binary, multiple runtime modes**
   - Mode selected via CLI flag: `--mode=api|worker|cron|micro`
   - Shared business logic across all modes
   - Cobra CLI framework handles commands

3. **Pure domain entities** (no infrastructure tags)
   - Domain entities have no JSON/GORM tags
   - DTOs handle JSON serialization
   - DB models handle GORM mappings

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
```go
// internal/adapter/repository/product_model.go
type ProductModel struct {
    ID        int64     `gorm:"primaryKey"`
    Name      string    `gorm:"not null"`
    Price     int64     `gorm:"not null"`
    CreatedAt time.Time `gorm:"autoCreateTime"`
}

func (m *ProductModel) ToDomain() *domain.Product { ... }
func ToProductModel(p *domain.Product) *ProductModel { ... }
```

```go
// internal/adapter/repository/product_repository_gorm.go
type ProductRepositoryGORM struct {
    db *gorm.DB
}

func NewProductRepository(db *gorm.DB) *ProductRepositoryGORM { ... }
func (r *ProductRepositoryGORM) Save(ctx context.Context, product *domain.Product) error { ... }
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
    wire.Bind(new(port.UserRepository), new(*repository.UserRepositoryGORM)),
    wire.Bind(new(port.ProductRepository), new(*repository.ProductRepositoryGORM)),  // ← Add this
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
| Repository Implementation | `internal/adapter/repository/{entity}_repository_gorm.go` | `user_repository_gorm.go` |
| DB Model | `internal/adapter/repository/{entity}_model.go` | `user_model.go` |
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
- ❌ **DON'T** add JSON/GORM tags to domain entities
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
- ⚠️ **CRITICAL**: Current password hashing is insecure
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
- ⚠️ **No tests exist yet** - need to implement tests for all layers

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

### Mode-Specific Notes

**API Mode**:
- Serves REST API endpoints
- Health check: `GET /health`
- ⚠️ Missing GatewaySet in wire config

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
DB Model (with GORM tags)
    ↓ [Database]
```

### Key Points
- **Domain entities** remain pure (no JSON/DB tags)
- **DTOs** handle JSON serialization (in delivery layer)
- **DB models** handle GORM mappings (in adapter layer)
- Each layer owns its transformation logic

---

## References

For detailed information, see:
- **Full coding standards**: [CODING_STANDARDS.md](CODING_STANDARDS.md)
- **User documentation**: [README.md](README.md)
- **Twelve-Factor compliance**: [CODING_STANDARDS.md#twelve-factor-app-compliance](CODING_STANDARDS.md#twelve-factor-app-compliance)
