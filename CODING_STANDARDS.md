# Go Clean Architecture Coding Standards

**Version**: 1.2
**Last Updated**: 2025-11-26

> **For AI assistants**: See [CLAUDE.md](CLAUDE.md) for a quick reference guide.
>
> **For project overview**: See [README.md](README.md) for installation and getting started.
>
> This document contains comprehensive coding standards for human contributors.

## Table of Contents

1. [Introduction](#introduction)
2. [Twelve-Factor App Compliance](#twelve-factor-app-compliance)
3. [Project Structure](#project-structure)
4. [Naming Conventions](#naming-conventions)
5. [Clean Architecture Layers](#clean-architecture-layers)
6. [Code Style Guidelines](#code-style-guidelines)
7. [Error Handling](#error-handling)
8. [Testing Standards](#testing-standards)
9. [Dependency Injection](#dependency-injection)
10. [Logging Standards](#logging-standards)
11. [API Design](#api-design)
12. [Protobuf and gRPC Standards](#protobuf-and-grpc-standards)
13. [Configuration Management](#configuration-management)
14. [Database Standards](#database-standards)
15. [Application Lifecycle](#application-lifecycle)
16. [Generated Code Policy](#generated-code-policy)
17. [Git Workflow](#git-workflow)

---

## Introduction

This document defines coding standards for Go projects implementing Clean Architecture principles. Following these standards ensures consistency, maintainability, and scalability across projects.

### Core Principles

1. **Dependency Inversion**: Inner layers don't depend on outer layers
2. **Domain-Driven Design**: Core business logic in the domain layer
3. **Interface Segregation**: Interfaces defined where they're needed
4. **Single Responsibility**: Each component has one clear purpose
5. **Testability**: Every layer can be tested in isolation

---

## Twelve-Factor App Compliance

This project follows the **[Twelve-Factor App](https://12factor.net/)** methodology for building modern, cloud-native applications.

### Why Twelve-Factor?

The Twelve-Factor methodology provides best practices for:
- ✅ **Portability**: Deploy anywhere (cloud, on-premise, containers)
- ✅ **Scalability**: Scale horizontally without code changes
- ✅ **Maintainability**: Clear separation of concerns
- ✅ **Reliability**: Graceful startup and shutdown

### Key Compliance Areas

| Factor | Implementation | Priority |
|--------|----------------|----------|
| **I. Codebase** | Single Git repo, multiple deployments | ✅ Compliant |
| **II. Dependencies** | Go modules (`go.mod`/`go.sum`) | ✅ Compliant |
| **III. Config** | Environment variables (`APP_*` prefix) | ✅ Compliant |
| **IV. Backing Services** | Database, RabbitMQ as attached resources | ✅ Compliant |
| **V. Build/Release/Run** | Docker multi-stage builds | ✅ Compliant |
| **VI. Processes** | Stateless processes | ✅ Compliant |
| **VII. Port Binding** | Self-contained HTTP server | ✅ Compliant |
| **VIII. Concurrency** | Multiple process types (api, worker, cron) | ✅ Compliant |
| **IX. Disposability** | Graceful shutdown with signal handling | ✅ **REQUIRED** |
| **X. Dev/Prod Parity** | Docker Compose for local dev | ✅ Compliant |
| **XI. Logs** | Structured logging to stdout | ✅ Compliant |
| **XII. Admin Processes** | `migrate` command for one-off tasks | ✅ Compliant |

**Overall Compliance**: **12/12 factors** ✅

For detailed assessment, see [TWELVE_FACTOR_COMPLIANCE.md](TWELVE_FACTOR_COMPLIANCE.md).

### Critical Requirements

**1. Factor III - Config** (Environment Variables)
```bash
# ✅ DO: Support environment-only deployment
export APP_DATABASE_HOST=prod-db.example.com
export APP_DATABASE_PASSWORD=$SECRET
./app api

# ❌ DON'T: Require config files in production
```

**2. Factor IX - Disposability** (Graceful Shutdown)
```go
// ✅ DO: Implement signal handling
signal.Notify(shutdown, syscall.SIGTERM, syscall.SIGINT)

// ❌ DON'T: Block indefinitely
router.Run(":8080")  // No cleanup on shutdown
```

**3. Factor XI - Logs** (Treat logs as event streams)
```go
// ✅ DO: Log to stdout with structured format
logger.InfoContext(ctx, "User created", "user_id", id)

// ❌ DON'T: Write to log files
f, _ := os.OpenFile("app.log", os.O_APPEND, 0644)
```

### Development Workflow

```bash
# Local development (Factor X - Dev/Prod Parity)
docker-compose up  # Same services as production

# Environment-based config (Factor III)
export APP_DATABASE_HOST=localhost
export APP_DATABASE_PASSWORD=devpass

# Multiple process types (Factor VIII)
./app api      # HTTP server
./app worker   # Background jobs
./app cron     # Scheduled tasks

# Graceful shutdown (Factor IX)
kill -TERM $PID  # Waits for in-flight requests
```

### Production Deployment

```bash
# Build stage (Factor V)
docker build -t myapp:v1.2.3 .

# Run stage with environment config (Factor III + VII)
docker run \
    -e APP_SERVER_PORT=8080 \
    -e APP_DATABASE_HOST=prod-db.aws.com \
    -e APP_DATABASE_PASSWORD=$DB_SECRET \
    -p 8080:8080 \
    myapp:v1.2.3 api

# Horizontal scaling (Factor VIII)
docker-compose up --scale api=3
```

### Checklist for New Features

When adding new features, ensure:
- [ ] Configuration via environment variables (Factor III)
- [ ] No local state in processes (Factor VI)
- [ ] Graceful shutdown support (Factor IX)
- [ ] Logs to stdout/stderr (Factor XI)
- [ ] Same code runs in dev and prod (Factor X)

---

## Project Structure

### Standard Directory Layout

```
.
├── cmd/                      # Application entry points
│   └── app/
│       ├── main.go          # Main entry point
│       └── cmd/             # Cobra commands (if using CLI)
│           ├── root.go
│           ├── api.go
│           ├── worker.go
│           └── cron.go
├── internal/                # Private application code
│   ├── domain/              # Domain entities and business rules
│   │   ├── user.go
│   │   ├── user_test.go
│   │   └── errors.go
│   ├── usecase/             # Application business logic
│   │   ├── port/            # Port interfaces
│   │   │   ├── user_repository.go
│   │   │   └── payment_gateway.go
│   │   ├── user_interactor.go
│   │   ├── user_interactor_test.go
│   │   └── interfaces.go
│   ├── adapter/             # Adapter implementations
│   │   ├── repository/      # Data access layer
│   │   │   ├── db.go
│   │   │   ├── user_model.go
│   │   │   ├── user_repository_sqlx.go
│   │   │   └── user_repository_test.go
│   │   └── gateway/         # External service clients
│   │       └── stripe_gateway.go
│   ├── delivery/            # Delivery mechanisms
│   │   ├── http/
│   │   │   ├── handler/
│   │   │   │   ├── user_handler.go
│   │   │   │   ├── user_handler_test.go
│   │   │   │   └── user_dto.go
│   │   │   └── router.go
│   │   ├── consumer/        # Message queue consumers
│   │   └── job/             # Scheduled jobs
│   └── app/                 # Application wiring
│       ├── api.go
│       ├── worker.go
│       ├── cron.go
│       ├── wire.go          # Wire dependency definitions
│       └── wire_gen.go      # Generated by Wire (don't edit)
├── pkg/                     # Public shared libraries
│   ├── config/
│   ├── logger/
│   └── middleware/
├── configs/                 # Configuration files
├── deployments/             # IaaS, PaaS, container orchestration deployment configs
│   ├── docker-compose.yaml  # Docker Compose for local development
│   └── kubernetes/          # Kubernetes manifests (if applicable)
├── migrations/              # Database migrations
├── scripts/                 # Build and deployment scripts
├── docs/                    # Documentation
├── api/                     # API definitions (OpenAPI/Swagger)
├── .editorconfig           # Editor configuration
├── .golangci.yml           # Linter configuration
├── go.mod
├── go.sum
├── Makefile
├── Dockerfile
└── README.md
```

### File Naming Rules

- **Go files**: `snake_case.go` (e.g., `user_repository.go`)
- **Test files**: `*_test.go` (e.g., `user_test.go`)
- **DTO files**: `*_dto.go` (e.g., `user_dto.go`)
- **Model files**: `*_model.go` (e.g., `user_model.go`)
- **Mock files**: `*_mock.go` (in `mocks/` subdirectory)
- **Interface files**: Descriptive names (e.g., `user_repository.go` for repository interface)

---

## Naming Conventions

### Package Names

```go
// ✅ Good: Short, lowercase, no underscores
package domain
package repository
package handler

// ❌ Bad: Mixed case, underscores
package domainLayer
package user_repository
```

### Variables and Functions

```go
// ✅ Good: camelCase for private, PascalCase for public
var userCount int
func getUserByID(id int64) (*User, error)

type User struct {
    ID        int64  // Public field
    email     string // Private field
}

func NewUser(email string) (*User, error) // Factory function
func (u *User) ChangePassword(newPassword string) error // Method

// ❌ Bad: Inconsistent casing
var UserCount int // Should be userCount (private)
func GetUser_ById(id int64) (*User, error) // Underscore in name
```

### Interface Names

```go
// ✅ Good: Descriptive, often ends with -er for single method
type UserRepository interface {}
type PaymentGateway interface {}
type Logger interface {}

// ✅ Also good: For single method interfaces
type Reader interface { Read() }
type Writer interface { Write() }

// ❌ Bad: Interface prefix
type IUserRepository interface {}
type UserRepositoryInterface interface {}
```

### Constants and Errors

```go
// ✅ Good: Constants in PascalCase or SCREAMING_SNAKE_CASE
const (
    MaxRetries = 3
    DefaultTimeout = 30 * time.Second
)

const (
    STATUS_PENDING   = "pending"
    STATUS_COMPLETED = "completed"
)

// ✅ Good: Errors with Err prefix
var (
    ErrUserNotFound      = errors.New("user not found")
    ErrInvalidPassword   = errors.New("invalid password")
    ErrEmailAlreadyExists = errors.New("email already exists")
)

// ❌ Bad: Inconsistent naming
var UserNotFoundError = errors.New("user not found")
const maxRetries = 3 // Should be capitalized
```

---

## Clean Architecture Layers

### 1. Domain Layer (`internal/domain/`)

**Rules**:
- ✅ **DO**: Pure business logic and entities
- ✅ **DO**: Rich domain models with behavior
- ✅ **DO**: Validate in constructors and methods
- ❌ **DON'T**: Import any other application layer
- ❌ **DON'T**: Use infrastructure tags (`json:`, `db:`, `gorm:`, etc.)
- ❌ **DON'T**: Depend on frameworks or external libraries

**Example**:
```go
// ✅ Good: Go-idiomatic domain entity with proper encapsulation
// Sensitive fields (password) are private, while most fields are public for simplicity.
// Validation and invariants are enforced through constructors and methods.
type User struct {
    ID        int64
    Email     string
    password  string // Private field - only accessible via methods
    CreatedAt time.Time
    UpdatedAt time.Time
}

// NewUser is a factory function that ensures the entity is created in a valid state.
// This is the Go-idiomatic way to handle validation without needing getters/setters.
func NewUser(email, password string) (*User, error) {
    if err := validateEmail(email); err != nil {
        return nil, err
    }
    if err := validatePassword(password); err != nil {
        return nil, err
    }

    hashedPassword, err := hashPassword(password)
    if err != nil {
        return nil, err
    }

    now := time.Now()
    return &User{
        Email:     email,
        password:  hashedPassword, // Store hashed password
        CreatedAt: now,
        UpdatedAt: now,
    }, nil
}

// Password returns the hashed password (read-only access)
func (u *User) Password() string {
    return u.password
}

// ChangePassword is a method that maintains business invariants.
// It ensures password changes follow business rules.
func (u *User) ChangePassword(oldPassword, newPassword string) error {
    if !u.IsPasswordCorrect(oldPassword) {
        return ErrInvalidPassword
    }

    if err := validatePassword(newPassword); err != nil {
        return err
    }

    hashedPassword, err := hashPassword(newPassword)
    if err != nil {
        return err
    }

    u.password = hashedPassword
    u.UpdatedAt = time.Now()
    return nil
}

// IsPasswordCorrect checks if the provided password matches the stored hash
func (u *User) IsPasswordCorrect(password string) bool {
    return comparePassword(u.password, password)
}

// ReconstructUser is used by the repository layer to rebuild entities from storage.
// This separates creation logic (NewUser) from reconstruction logic.
func ReconstructUser(id int64, email, passwordHash string, createdAt, updatedAt time.Time) *User {
    return &User{
        ID:        id,
        Email:     email,
        password:  passwordHash, // Already hashed
        CreatedAt: createdAt,
        UpdatedAt: updatedAt,
    }
}

// ❌ Bad: Anemic domain model with infrastructure tags
type User struct {
    ID       int64  `json:"id" db:"id"`
    Email    string `json:"email" db:"email"`
    Password string `json:"-" db:"password_hash"`
}
// This violates Clean Architecture because domain entities
// should not know about infrastructure concerns (JSON, database).
```

### 2. Use Case Layer (`internal/usecase/`)

**Rules**:
- ✅ **DO**: Define port interfaces in `port/` subdirectory
- ✅ **DO**: Depend only on domain entities and port interfaces
- ✅ **DO**: Accept `context.Context` as first parameter
- ✅ **DO**: One interactor per entity/aggregate
- ❌ **DON'T**: Import delivery or adapter implementations
- ❌ **DON'T**: Contain framework-specific code

**Example**:
```go
// ✅ Good: Port interface defined by use case
// internal/usecase/port/user_repository.go
type UserRepository interface {
    Save(ctx context.Context, user *domain.User) error
    FindByID(ctx context.Context, id int64) (*domain.User, error)
    FindByEmail(ctx context.Context, email string) (*domain.User, error)
}

// ✅ Good: Interactor depends on interfaces
// internal/usecase/user_interactor.go
type UserInteractor struct {
    userRepo port.UserRepository
}

func NewUserInteractor(userRepo port.UserRepository) *UserInteractor {
    return &UserInteractor{
        userRepo: userRepo,
    }
}

func (i *UserInteractor) CreateUser(ctx context.Context, email, password string) (*domain.User, error) {
    // Check if user exists
    existingUser, err := i.userRepo.FindByEmail(ctx, email)
    if err == nil && existingUser != nil {
        return nil, domain.ErrEmailAlreadyExists
    }

    // Create domain entity
    user, err := domain.NewUser(email, password)
    if err != nil {
        return nil, err
    }

    // Persist
    if err := i.userRepo.Save(ctx, user); err != nil {
        return nil, err
    }

    return user, nil
}
```

### 3. Adapter Layer (`internal/adapter/`)

**Rules**:
- ✅ **DO**: Implement port interfaces
- ✅ **DO**: Transform between domain entities and external formats
- ✅ **DO**: Handle infrastructure concerns (DB, external APIs)
- ✅ **DO**: Use tags for DB/serialization in model structs
- ❌ **DON'T**: Expose infrastructure details to use cases

**Example**:
```go
// ✅ Good: DB model separate from domain entity
// internal/adapter/repository/user_model.go
type UserDBModel struct {
    ID           int64     `db:"id"`
    Email        string    `db:"email"`
    PasswordHash string    `db:"password_hash"`
    CreatedAt    time.Time `db:"created_at"`
    UpdatedAt    time.Time `db:"updated_at"`
}

func (m *UserDBModel) ToDomain() *domain.User {
    return domain.ReconstructUser(m.ID, m.Email, m.PasswordHash, m.CreatedAt, m.UpdatedAt)
}

func FromDomain(user *domain.User) *UserDBModel {
    return &UserDBModel{
        ID:           user.ID,
        Email:        user.Email,
        PasswordHash: user.Password(),
        CreatedAt:    user.CreatedAt,
        UpdatedAt:    user.UpdatedAt,
    }
}

// ✅ Good: Repository implements port interface using sqlx + Squirrel
// internal/adapter/repository/user_repository_sqlx.go
import (
    "context"
    "database/sql"
    "errors"
    "fmt"

    "github.com/Masterminds/squirrel"
    "github.com/jmoiron/sqlx"
    "go-clean-arch/internal/domain"
)

// PostgreSQL uses $1, $2 placeholders (use squirrel.Question for MySQL)
var psql = squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

type UserRepositorySQLX struct {
    db *sqlx.DB
}

func NewUserRepository(db *sqlx.DB) *UserRepositorySQLX {
    return &UserRepositorySQLX{db: db}
}

func (r *UserRepositorySQLX) Save(ctx context.Context, user *domain.User) error {
    model := FromDomain(user)

    // Build SQL with Squirrel
    query, args, err := psql.Insert("users").
        Columns("email", "password_hash", "created_at", "updated_at").
        Values(model.Email, model.PasswordHash, model.CreatedAt, model.UpdatedAt).
        Suffix("RETURNING id").  // PostgreSQL-specific
        ToSql()

    if err != nil {
        return fmt.Errorf("failed to build query: %w", err)
    }

    // Execute with sqlx
    err = r.db.QueryRowxContext(ctx, query, args...).Scan(&model.ID)
    if err != nil {
        return fmt.Errorf("failed to insert user: %w", err)
    }

    return nil
}

func (r *UserRepositorySQLX) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
    query, args, err := psql.Select("*").
        From("users").
        Where(squirrel.Eq{"email": email}).
        ToSql()

    if err != nil {
        return nil, fmt.Errorf("failed to build query: %w", err)
    }

    var model UserDBModel
    err = r.db.GetContext(ctx, &model, query, args...)
    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return nil, nil  // or return domain.ErrUserNotFound
        }
        return nil, fmt.Errorf("database error: %w", err)
    }

    return model.ToDomain(), nil
}
```

### 4. Delivery Layer (`internal/delivery/`)

**Rules**:
- ✅ **DO**: Keep handlers thin (validation and transformation only)
- ✅ **DO**: Use DTOs with serialization tags
- ✅ **DO**: Transform between DTOs and domain entities
- ✅ **DO**: Handle HTTP/framework-specific concerns
- ❌ **DON'T**: Put business logic in handlers
- ❌ **DON'T**: Access repositories directly

> **Framework Agnostic Design (2025):**
> While this project uses **Gin** for its maturity and ecosystem, the Clean Architecture design allows us to swap the delivery mechanism easily.
>
> With the release of **Go 1.22+**, the standard library's `net/http` router has become powerful enough for many services. For new microservices requiring minimal dependencies, teams are encouraged to evaluate standard **`http.ServeMux`** or **Chi** as lightweight alternatives, provided they adhere to the same Handler/DTO patterns defined here.

**Example**:
```go
// ✅ Good: DTO with validation tags
// internal/delivery/http/handler/user_dto.go
type CreateUserRequest struct {
    Email    string `json:"email" binding:"required,email"`
    Password string `json:"password" binding:"required,min=8"`
}

type UserResponse struct {
    ID        int64     `json:"id"`
    Email     string    `json:"email"`
    CreatedAt time.Time `json:"created_at"`
}

func ToUserResponse(user *domain.User) *UserResponse {
    return &UserResponse{
        ID:        user.ID,
        Email:     user.Email,
        CreatedAt: user.CreatedAt,
    }
}

// ✅ Good: Thin handler
// internal/delivery/http/handler/user_handler.go
type UserHandler struct {
    userUC usecase.UserUsecase
}

func NewUserHandler(userUC usecase.UserUsecase) *UserHandler {
    return &UserHandler{userUC: userUC}
}

func (h *UserHandler) Create(c *gin.Context) {
    var req CreateUserRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    user, err := h.userUC.CreateUser(c.Request.Context(), req.Email, req.Password)
    if err != nil {
        // Error handling middleware will handle this
        _ = c.Error(err)
        return
    }

    c.JSON(http.StatusCreated, ToUserResponse(user))
}
```

### 5. Microservices Delivery Layer (`internal/delivery/micro/`)

This project supports microservices using **go-micro v5** framework for RPC-based communication.

**Rules**:
- ✅ **DO**: Keep handlers thin (validation and transformation only)
- ✅ **DO**: Convert between protobuf messages and domain entities
- ✅ **DO**: Use same use case layer as HTTP delivery
- ✅ **DO**: Handle errors and convert to appropriate RPC errors
- ❌ **DON'T**: Put business logic in service handlers
- ❌ **DON'T**: Access repositories directly
- ❌ **DON'T**: Duplicate use case logic

**Example**:
```go
// ✅ Good: Thin microservice handler
// internal/delivery/micro/handler/user_service.go
package handler

import (
    "context"
    "go-clean-arch/internal/domain"
    "go-clean-arch/internal/usecase"
    pb "go-clean-arch/proto/user"
    "github.com/samber/oops"
)

type UserService struct {
    userUsecase usecase.UserUsecase
}

func NewUserService(userUsecase usecase.UserUsecase) *UserService {
    return &UserService{
        userUsecase: userUsecase,
    }
}

// CreateUser creates a new user
func (s *UserService) CreateUser(ctx context.Context, req *pb.CreateUserRequest, rsp *pb.CreateUserResponse) error {
    // Call usecase (same as HTTP handler)
    user, err := s.userUsecase.CreateUser(ctx, req.Email, req.Password)
    if err != nil {
        return convertError(err)
    }

    // Convert domain entity to proto response
    rsp.Id = user.ID
    rsp.Email = user.Email
    rsp.CreatedAt = user.CreatedAt.Format("2006-01-02T15:04:05Z07:00")
    rsp.UpdatedAt = user.UpdatedAt.Format("2006-01-02T15:04:05Z07:00")

    return nil
}

// convertError converts domain errors to appropriate RPC errors
func convertError(err error) error {
    ooErr, ok := err.(oops.OopsError)
    if !ok {
        return err
    }

    // Map domain errors to RPC error codes
    switch ooErr.Code() {
    case domain.ErrCodeUserNotFound:
        return oops.
            Code("NOT_FOUND").
            With("message", "user not found").
            Errorf("user not found")
    case domain.ErrCodeEmailAlreadyExists:
        return oops.
            Code("ALREADY_EXISTS").
            With("message", "email already exists").
            Errorf("email already exists")
    default:
        return err
    }
}
```

**Key Differences from HTTP Delivery**:
- Uses protobuf messages instead of JSON DTOs
- Returns errors directly (no HTTP status codes)
- go-micro handles service discovery and load balancing

**Registering Microservice**:
```go
// internal/app/micro.go
package app

import (
    "go-clean-arch/internal/delivery/micro/handler"
    pb "go-clean-arch/proto/user"
    "go-micro.dev/v5"
)

func NewMicroService(db *sqlx.DB, cfg *config.Config) (micro.Service, error) {
    // Create microservice
    service := micro.NewService(
        micro.Name("go.micro.srv.user"),
        micro.Version("latest"),
    )

    service.Init()

    // Initialize dependencies
    userService := InitializeUserService(db, cfg)

    // Register handler
    if err := pb.RegisterUserServiceHandler(service.Server(), userService); err != nil {
        return nil, err
    }

    return service, nil
}
```

---

## Code Style Guidelines

### General Go Style

Follow the official [Effective Go](https://golang.org/doc/effective_go) and [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments).

### Formatting

```bash
# Always format code before committing
go fmt ./...
goimports -w .
```

### Line Length

- Prefer lines under 120 characters
- Break long lines logically

### Comments

```go
// ✅ Good: Package comments
// Package domain contains core business entities and rules.
package domain

// ✅ Good: Exported function comments
// NewUser creates a new user with the given email and password.
// It validates the input and returns an error if validation fails.
func NewUser(email, password string) (*User, error) {
    // Implementation
}

// ✅ Good: Complex logic explanation
// We need to check both active and pending orders because
// a user might have initiated checkout but not completed payment
orders, err := r.repo.FindByUserIDAndStatus(ctx, userID, StatusActive, StatusPending)

// ❌ Bad: Obvious comments
// Get user by ID
user := getUserByID(id)

// ❌ Bad: No comment on exported function
func ProcessPayment(amount int64) error {
    // ...
}
```

### Error Handling

```go
// ✅ Good: Check errors immediately
user, err := repo.FindByID(ctx, id)
if err != nil {
    return nil, fmt.Errorf("failed to find user: %w", err)
}

// ✅ Good: Named return values for clarity (when appropriate)
func CreateUser(email string) (user *User, err error) {
    defer func() {
        if err != nil {
            log.Error("CreateUser failed", "email", email, "error", err)
        }
    }()
    // Implementation
}

// ❌ Bad: Ignoring errors
user, _ := repo.FindByID(ctx, id)

// ❌ Bad: Multiple error checks in one line
if user, err := repo.FindByID(ctx, id); err != nil { return nil, err } else { return user, nil }
```

---

## Generics (Go 1.18+)

### When to Use Generics

Generics are useful for reducing boilerplate code in specific scenarios. Use them judiciously:

**✅ Good Use Cases:**
- Generic data structures (pagination, result wrappers)
- Utility functions operating on multiple types
- DTO conversion helpers
- Repository pagination helpers

**❌ Avoid:**
- Overusing generics where simple interfaces suffice
- Making code unnecessarily complex
- Using generics in domain entities (keep them simple)

### Generic Pagination Response

```go
// ✅ Good: Generic pagination for DTOs
// pkg/common/response.go or internal/delivery/http/common/
type PageResponse[T any] struct {
    Data       []T   `json:"data"`
    Total      int64 `json:"total"`
    Page       int   `json:"page"`
    PageSize   int   `json:"page_size"`
    TotalPages int   `json:"total_pages"`
}

func NewPageResponse[T any](data []T, total int64, page, pageSize int) *PageResponse[T] {
    totalPages := (int(total) + pageSize - 1) / pageSize
    return &PageResponse[T]{
        Data:       data,
        Total:      total,
        Page:       page,
        PageSize:   pageSize,
        TotalPages: totalPages,
    }
}

// Usage in handler
func (h *UserHandler) List(c *gin.Context) {
    users, total, err := h.userUC.ListUsers(ctx, page, pageSize)
    // ...

    // Convert domain entities to DTOs
    userDTOs := make([]UserResponse, len(users))
    for i, u := range users {
        userDTOs[i] = *ToUserResponse(u)
    }

    response := NewPageResponse(userDTOs, total, page, pageSize)
    c.JSON(http.StatusOK, response)
}
```

### Generic DTO Conversion

```go
// ✅ Good: Generic slice transformation
func MapSlice[T any, U any](items []T, fn func(T) U) []U {
    result := make([]U, len(items))
    for i, item := range items {
        result[i] = fn(item)
    }
    return result
}

// Usage
func (h *UserHandler) List(c *gin.Context) {
    users, total, err := h.userUC.ListUsers(ctx, page, pageSize)
    // ...

    // Concise DTO conversion
    userDTOs := MapSlice(users, func(u *domain.User) UserResponse {
        return *ToUserResponse(u)
    })

    response := NewPageResponse(userDTOs, total, page, pageSize)
    c.JSON(http.StatusOK, response)
}
```

### Generic Repository Helpers

```go
// ✅ Good: Generic pagination query helper
type PaginationParams struct {
    Page     int
    PageSize int
    Offset   int
}

func NewPaginationParams(page, pageSize int) PaginationParams {
    if page < 1 {
        page = 1
    }
    if pageSize < 1 || pageSize > 100 {
        pageSize = 20
    }
    return PaginationParams{
        Page:     page,
        PageSize: pageSize,
        Offset:   (page - 1) * pageSize,
    }
}

// Can be used across all repositories
func (r *UserRepositorySQLX) List(ctx context.Context, params PaginationParams) ([]*domain.User, int64, error) {
    // Count total
    var total int64
    countQuery, countArgs, _ := psql.Select("COUNT(*)").From("users").ToSql()
    r.db.GetContext(ctx, &total, countQuery, countArgs...)

    // Get paginated results
    query, args, _ := psql.Select("*").
        From("users").
        OrderBy("created_at DESC").
        Limit(uint64(params.PageSize)).
        Offset(uint64(params.Offset)).
        ToSql()

    var models []UserDBModel
    if err := r.db.SelectContext(ctx, &models, query, args...); err != nil {
        return nil, 0, err
    }

    users := make([]*domain.User, len(models))
    for i, m := range models {
        users[i] = m.ToDomain()
    }

    return users, total, nil
}
```

### Generic Result Wrapper (Optional)

```go
// ✅ Good: Generic result type for better error handling
type Result[T any] struct {
    value T
    err   error
}

func Ok[T any](value T) Result[T] {
    return Result[T]{value: value}
}

func Err[T any](err error) Result[T] {
    return Result[T]{err: err}
}

func (r Result[T]) Unwrap() (T, error) {
    return r.value, r.err
}

func (r Result[T]) IsOk() bool {
    return r.err == nil
}

func (r Result[T]) IsErr() bool {
    return r.err != nil
}

// Usage (more Rust-like, use sparingly in Go)
func (i *UserInteractor) CreateUser(ctx context.Context, email, password string) Result[*domain.User] {
    user, err := domain.NewUser(email, password)
    if err != nil {
        return Err[*domain.User](err)
    }

    if err := i.userRepo.Save(ctx, user); err != nil {
        return Err[*domain.User](err)
    }

    return Ok(user)
}
```

### Best Practices

**DO:**
- ✅ Use generics for common data structures (pagination, lists)
- ✅ Use generics for utility functions (mapping, filtering)
- ✅ Keep generic functions simple and focused
- ✅ Provide type parameters with clear names: `[T any]`, `[K comparable, V any]`

**DON'T:**
- ❌ Use generics in domain entities (keep business logic simple)
- ❌ Over-engineer with unnecessary type parameters
- ❌ Replace simple interfaces with generics
- ❌ Use generics when a simple `any` or interface works better

---

## Error Handling

### Domain Error Types

Define custom error types in `internal/domain/errors.go`:

```go
package domain

import "errors"

var (
    // User errors
    ErrUserNotFound       = errors.New("user not found")
    ErrEmailAlreadyExists = errors.New("email already exists")
    ErrInvalidPassword    = errors.New("invalid password")
    ErrInvalidEmail       = errors.New("invalid email format")

    // Order errors
    ErrOrderNotFound      = errors.New("order not found")
    ErrInvalidOrderStatus = errors.New("invalid order status transition")
    ErrOrderAlreadyPaid   = errors.New("order already paid")
)

// Sentinel errors for different categories
var (
    ErrNotFound      = errors.New("resource not found")
    ErrUnauthorized  = errors.New("unauthorized")
    ErrForbidden     = errors.New("forbidden")
    ErrBadRequest    = errors.New("bad request")
)
```

### Error Wrapping and Context

#### Option 1: Standard Library (Recommended for Simple Projects)

Use Go's built-in error wrapping with `fmt.Errorf` and `%w`:

```go
import (
    "context"
    "database/sql"
    "errors"
    "fmt"
)

// ✅ Good: Standard library error wrapping
func (r *UserRepositorySQLX) FindByID(ctx context.Context, id int64) (*domain.User, error) {
    query, args, err := psql.Select("*").
        From("users").
        Where(squirrel.Eq{"id": id}).
        ToSql()

    if err != nil {
        return nil, fmt.Errorf("failed to build query: %w", err)
    }

    var model UserDBModel
    err = r.db.GetContext(ctx, &model, query, args...)
    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return nil, fmt.Errorf("user %d: %w", id, domain.ErrUserNotFound)
        }
        return nil, fmt.Errorf("failed to query user %d: %w", id, err)
    }

    return model.ToDomain(), nil
}
```

#### Option 2: Rich Error Context with oops (Optional)

For complex projects requiring structured error context, you can use the [oops](https://github.com/samber/oops) library:

```go
import "github.com/samber/oops"

// ✅ Good: Rich error context with oops
func (r *UserRepositorySQLX) FindByID(ctx context.Context, id int64) (*domain.User, error) {
    query, args, err := psql.Select("*").
        From("users").
        Where(squirrel.Eq{"id": id}).
        ToSql()

    if err != nil {
        return nil, oops.
            Code("QUERY_BUILD_ERROR").
            In("repository").
            With("user_id", id).
            Wrapf(err, "failed to build query")
    }

    var model UserDBModel
    err = r.db.GetContext(ctx, &model, query, args...)
    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return nil, oops.
                Code("USER_NOT_FOUND").
                In("repository").
                With("user_id", id).
                Hint("Check if user ID exists in database").
                Wrapf(domain.ErrUserNotFound, "user with ID %d not found", id)
        }
        return nil, oops.
            Code("DATABASE_ERROR").
            In("repository").
            With("user_id", id).
            Hint("Check database connectivity").
            Wrapf(err, "failed to query user from database")
    }

    return model.ToDomain(), nil
}
```

**When to use oops:**
- ✅ Complex microservices with distributed tracing
- ✅ Need structured logging with error context
- ✅ Multiple teams require consistent error codes

**When to use standard library:**
- ✅ Simple projects or getting started
- ✅ Want to minimize dependencies
- ✅ Basic error wrapping is sufficient

### Error Mapping in Middleware

Map domain errors to HTTP status codes:

```go
// pkg/middleware/error_handler.go
func ErrorHandler() gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Next()

        if len(c.Errors) == 0 {
            return
        }

        err := c.Errors.Last().Err

        var statusCode int
        var message string

        switch {
        case errors.Is(err, domain.ErrUserNotFound),
             errors.Is(err, domain.ErrOrderNotFound):
            statusCode = http.StatusNotFound
            message = err.Error()
        case errors.Is(err, domain.ErrEmailAlreadyExists):
            statusCode = http.StatusConflict
            message = err.Error()
        case errors.Is(err, domain.ErrInvalidPassword),
             errors.Is(err, domain.ErrInvalidEmail):
            statusCode = http.StatusBadRequest
            message = err.Error()
        default:
            statusCode = http.StatusInternalServerError
            message = "Internal server error"
        }

        c.JSON(statusCode, gin.H{
            "error": message,
            "code":  oops.AsOops(err).Code(),
        })
    }
}
```

---

## Testing Standards

### Test File Organization

```go
// internal/domain/user_test.go
package domain_test // Use separate test package for black-box testing

import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/yourorg/yourproject/internal/domain"
)
```

### Table-Driven Tests

```go
// ✅ Good: Table-driven tests
func TestNewUser(t *testing.T) {
    tests := []struct {
        name        string
        email       string
        password    string
        wantErr     error
    }{
        {
            name:     "valid user",
            email:    "test@example.com",
            password: "SecurePass123",
            wantErr:  nil,
        },
        {
            name:     "invalid email",
            email:    "invalid-email",
            password: "SecurePass123",
            wantErr:  domain.ErrInvalidEmail,
        },
        {
            name:     "weak password",
            email:    "test@example.com",
            password: "123",
            wantErr:  domain.ErrInvalidPassword,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            user, err := domain.NewUser(tt.email, tt.password)

            if tt.wantErr != nil {
                assert.ErrorIs(t, err, tt.wantErr)
                assert.Nil(t, user)
            } else {
                assert.NoError(t, err)
                assert.NotNil(t, user)
                assert.Equal(t, tt.email, user.Email)
            }
        })
    }
}
```

### Mock Generation

Use [mockery](https://github.com/vektra/mockery) for generating mocks:

```bash
# Install mockery
go install github.com/vektra/mockery/v2@latest

# Generate mocks for all interfaces in a directory
mockery --dir=internal/usecase/port --all --output=internal/usecase/port/mocks
```

### Test Coverage Goals

- **Domain layer**: 90%+ coverage (focus on core business logic and methods with behavior; simple constructors and data holders can be exempted)
- **Use case layer**: 90%+ coverage
- **Repository layer**: 80%+ coverage (integration tests)
- **Handler layer**: 70%+ coverage

**Note**: While we strive for high coverage, 100% coverage is often impractical and provides diminishing returns. Focus testing efforts on:
- Business logic and validation rules
- State transitions and invariants
- Error handling paths
- Complex algorithms

Simple getters, setters, and trivial constructors don't require exhaustive testing if they contain no logic.

```bash
# Check coverage
go test -cover ./...

# Generate HTML coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

---

## Dependency Injection

### Using Wire

This project uses [Google Wire](https://github.com/google/wire) for compile-time dependency injection.

**Wire Setup**:

```go
// internal/app/wire.go
//go:build wireinject
// +build wireinject

package app

import (
    "github.com/google/wire"
    "github.com/yourorg/yourproject/internal/adapter/repository"
    "github.com/yourorg/yourproject/internal/usecase"
    "github.com/yourorg/yourproject/internal/delivery/http/handler"
)

// Provider Sets
var RepositorySet = wire.NewSet(
    repository.NewUserRepository,
    wire.Bind(new(port.UserRepository), new(*repository.UserRepositorySQLX)),
)

var UseCaseSet = wire.NewSet(
    usecase.NewUserInteractor,
    wire.Bind(new(usecase.UserUsecase), new(*usecase.UserInteractor)),
)

var HandlerSet = wire.NewSet(
    handler.NewUserHandler,
)

// Injector function
func InitializeAPIRouter(db *sqlx.DB, logger *slog.Logger) (*gin.Engine, error) {
    wire.Build(
        RepositorySet,
        UseCaseSet,
        HandlerSet,
        NewRouter,
    )
    return nil, nil
}
```

**Generate Wire code**:

```bash
# Install Wire
go install github.com/google/wire/cmd/wire@latest

# Generate wire code
wire gen ./internal/app
```

---

## Logging Standards

### Structured Logging with slog

Use Go's built-in `log/slog` for structured logging.

```go
import "log/slog"

// ✅ Good: Context-aware logging
func (i *UserInteractor) CreateUser(ctx context.Context, email string) (*domain.User, error) {
    logger.InfoContext(ctx, "Creating new user",
        "email", email,
    )

    user, err := domain.NewUser(email, password)
    if err != nil {
        logger.ErrorContext(ctx, "Failed to create user entity",
            "email", email,
            "error", err,
        )
        return nil, err
    }

    logger.InfoContext(ctx, "User created successfully",
        "user_id", user.ID,
        "email", email,
    )

    return user, nil
}
```

### Log Levels

- **DEBUG**: Detailed information for debugging
- **INFO**: General informational messages
- **WARN**: Warning messages for potentially harmful situations
- **ERROR**: Error messages for error events

### Trace ID Propagation

Use middleware to inject trace IDs:

```go
// pkg/middleware/trace.go
func TraceID() gin.HandlerFunc {
    return func(c *gin.Context) {
        traceID := c.GetHeader("X-Trace-ID")
        if traceID == "" {
            traceID = uuid.New().String()
        }

        ctx := context.WithValue(c.Request.Context(), "trace_id", traceID)
        c.Request = c.Request.WithContext(ctx)
        c.Header("X-Trace-ID", traceID)

        c.Next()
    }
}
```

---

## API Design

### RESTful Principles

```go
// ✅ Good: RESTful routes
POST   /api/v1/users              // Create user
GET    /api/v1/users              // List users
GET    /api/v1/users/:id          // Get user
PUT    /api/v1/users/:id          // Update user (full)
PATCH  /api/v1/users/:id          // Update user (partial)
DELETE /api/v1/users/:id          // Delete user

// Resource-specific actions
POST   /api/v1/users/:id/password // Change password
POST   /api/v1/orders/:id/checkout // Checkout order

// ❌ Bad: Non-RESTful routes
GET    /api/v1/getUser/:id
POST   /api/v1/createUser
POST   /api/v1/user/changePassword
```

### Request/Response Format

```go
// ✅ Good: Consistent response format
{
    "id": 1,
    "email": "user@example.com",
    "created_at": "2025-11-15T10:00:00Z"
}

// ✅ Good: Error response format
{
    "error": "User not found",
    "code": "USER_NOT_FOUND",
    "trace_id": "abc-123-def"
}

// ✅ Good: List response with pagination
{
    "data": [...],
    "pagination": {
        "page": 1,
        "page_size": 20,
        "total": 100,
        "total_pages": 5
    }
}
```

### HTTP Status Codes

- `200 OK`: Successful GET, PUT, PATCH
- `201 Created`: Successful POST
- `204 No Content`: Successful DELETE
- `400 Bad Request`: Invalid input
- `401 Unauthorized`: Authentication required
- `403 Forbidden`: Permission denied
- `404 Not Found`: Resource not found
- `409 Conflict`: Duplicate resource
- `500 Internal Server Error`: Server error
- `503 Service Unavailable`: Service degraded

---

## Protobuf and gRPC Standards

### Protobuf File Organization

Following [golang-standards/project-layout](https://github.com/golang-standards/project-layout), protobuf files **must** be placed in the **`api/proto/`** directory:

```
.
├── api/                         # API definitions
│   ├── proto/                   # Protocol Buffer definitions
│   │   └── user/
│   │       ├── user.proto       # User service definition
│   │       └── v1/              # Versioned APIs
│   │           └── user.proto
│   └── openapi/                 # OpenAPI/Swagger specs
│       └── openapi.yaml
```

This structure keeps API contracts organized and follows Go community standards.

### Protobuf Style Guide

Follow [Google's Protocol Buffers Style Guide](https://protobuf.dev/programming-guides/style/):

```protobuf
// ✅ Good: Well-structured proto file
// api/proto/user/user.proto
syntax = "proto3";

package user;

option go_package = "go-clean-arch/api/proto/user;user";

// UserService handles user management operations
service UserService {
    // CreateUser creates a new user account
    rpc CreateUser(CreateUserRequest) returns (CreateUserResponse);

    // GetUser retrieves a user by ID
    rpc GetUser(GetUserRequest) returns (GetUserResponse);

    // ListUsers lists users with pagination
    rpc ListUsers(ListUsersRequest) returns (ListUsersResponse);
}

// CreateUserRequest contains fields for user creation
message CreateUserRequest {
    string email = 1;
    string password = 2;
}

// CreateUserResponse returns the created user
message CreateUserResponse {
    int64 id = 1;
    string email = 2;
    string created_at = 3;
    string updated_at = 4;
}

// UserInfo represents basic user information
message UserInfo {
    int64 id = 1;
    string email = 2;
    string created_at = 3;
    string updated_at = 4;
}
```

### Naming Conventions

**Services**:
- Use `PascalCase` for service names: `UserService`, `OrderService`
- Service names should be nouns: `AuthService`, not `Authenticator`

**Methods (RPCs)**:
- Use `PascalCase` for method names: `CreateUser`, `GetOrder`
- Start with verb: `Create`, `Get`, `Update`, `Delete`, `List`

**Messages**:
- Use `PascalCase` for message names: `CreateUserRequest`, `UserResponse`
- Request messages: `{Method}{Resource}Request`
- Response messages: `{Method}{Resource}Response`
- Nested resources: `User.Address`, `Order.Item`

**Fields**:
- Use `snake_case` for field names: `user_id`, `created_at`
- Field numbers start at 1
- Reserve field numbers 19000-19999 for future use

### Code Generation

```bash
# Install protoc compiler
# macOS
brew install protobuf

# Linux
apt-get install -y protobuf-compiler

# Install Go plugins
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Generate Go code from proto files
protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    api/proto/**/*.proto

# Or use Makefile
make proto-gen
```

**Makefile Target**:
```makefile
.PHONY: proto-gen
proto-gen: ## Generate Go code from proto files
	@echo "Generating protobuf code..."
	protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		api/proto/**/*.proto
```

### Versioning API

Use directories for API versioning:

```
api/
├── proto/
│   └── user/
│       ├── v1/
│       │   └── user.proto       # Version 1
│       └── v2/
│           └── user.proto       # Version 2 (breaking changes)
```

```protobuf
// api/proto/user/v1/user.proto
syntax = "proto3";

package user.v1;

option go_package = "go-clean-arch/api/proto/user/v1;userv1";

// Version 1 of UserService
service UserService {
    rpc CreateUser(CreateUserRequest) returns (CreateUserResponse);
}
```

### Error Handling

Use standard gRPC error codes:

```go
import (
    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"
)

// ✅ Good: Proper gRPC error handling
func (s *UserService) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.GetUserResponse, error) {
    user, err := s.userUsecase.GetUserByID(ctx, req.Id)
    if err != nil {
        // Map domain errors to gRPC codes
        if errors.Is(err, domain.ErrUserNotFound) {
            return nil, status.Error(codes.NotFound, "user not found")
        }
        return nil, status.Error(codes.Internal, "internal error")
    }

    return &pb.GetUserResponse{
        Id:    user.ID,
        Email: user.Email,
    }, nil
}
```

**Standard gRPC Codes**:
- `codes.OK` - Success (0)
- `codes.NotFound` - Resource not found (5)
- `codes.AlreadyExists` - Resource already exists (6)
- `codes.InvalidArgument` - Invalid request (3)
- `codes.Unauthenticated` - Authentication required (16)
- `codes.PermissionDenied` - Permission denied (7)
- `codes.Internal` - Internal server error (13)

### Best Practices

**DO**:
- ✅ Use protobuf3 syntax
- ✅ Add comments to all services, methods, and messages
- ✅ Version your APIs (v1, v2)
- ✅ Use descriptive field names
- ✅ Keep messages focused and small
- ✅ Generate code as part of build process
- ✅ Commit generated `*.pb.go` files so the project is buildable after cloning
- ⚠️ **WARNING**: Ensure you use the same `protoc` version as the team (defined in Makefile/Dockerfile) to avoid noise in diffs

**DON'T**:
- ❌ Reuse field numbers (after removing fields)
- ❌ Change field types (breaks backward compatibility)
- ❌ Use reserved keywords as field names
- ❌ Put business logic in proto files

### Integration with Clean Architecture

```
┌─────────────────────────────────────────┐
│  Protobuf Definitions (api/proto/)     │  ← API contracts
└──────────────┬──────────────────────────┘
               │
┌──────────────▼──────────────────────────┐
│  Delivery Layer (internal/delivery/)    │  ← Converts Proto ↔ Domain
│  - micro/handler/user_service.go        │
└──────────────┬──────────────────────────┘
               │
┌──────────────▼──────────────────────────┐
│  Use Case Layer (internal/usecase/)     │  ← Business logic (proto-agnostic)
└──────────────┬──────────────────────────┘
               │
┌──────────────▼──────────────────────────┐
│  Domain Layer (internal/domain/)        │  ← Pure domain entities
└─────────────────────────────────────────┘
```

**Key Point**: Protobuf messages are **infrastructure concerns** and should **not** leak into use case or domain layers.

---

## Configuration Management

### Using Viper for Configuration

This project uses [Viper](https://github.com/spf13/viper) for configuration management, following the [Twelve-Factor App](https://12factor.net/) methodology (Factor III - Config).

**Key Principles**:
- ✅ Config files are **OPTIONAL** (Twelve-Factor compliant)
- ✅ Environment variables override all config values
- ✅ Support environment-only deployment
- ✅ No secrets in config files

### Environment Variables

All configuration can be provided via environment variables:

```bash
# Environment variables use APP_ prefix
export APP_SERVER_PORT=8080
export APP_SERVER_HOST=0.0.0.0
export APP_DATABASE_HOST=postgres.example.com
export APP_DATABASE_USER=myuser
export APP_DATABASE_PASSWORD=secret
export APP_LOGGER_LEVEL=info
export APP_LOGGER_FORMAT=json

# Run application with environment variables only
./app api
```

**Naming Convention**:
- Prefix: `APP_`
- Nested keys use underscore: `database.host` → `APP_DATABASE_HOST`
- All uppercase

### Configuration Files (Optional)

Config files are provided for **local development convenience only**:

```toml
# configs/config.example.toml
[server]
port = 8080
host = "0.0.0.0"
mode = "debug"  # debug, release, test

[database]
driver = "postgres"
host = "localhost"
port = 5432
user = "postgres"
password = "postgres"
dbname = "cleanarch"
sslmode = "disable"
auto_migrate = false  # Use explicit migrations in production

[logger]
level = "info"   # debug, info, warn, error
format = "json"  # json, text
```

### Configuration Structure

```go
// pkg/config/config.go
type Config struct {
    Server   ServerConfig   `mapstructure:"server"`
    Database DatabaseConfig `mapstructure:"database"`
    RabbitMQ RabbitMQConfig `mapstructure:"rabbitmq"`
    Logger   LoggerConfig   `mapstructure:"logger"`
}

type ServerConfig struct {
    Port int    `mapstructure:"port"`
    Host string `mapstructure:"host"`
    Mode string `mapstructure:"mode"`
}
```

### Loading Configuration

```go
// Config file is optional - application can run with env vars only
cfg, err := config.Load("configs/config.toml")  // File path is optional
if err != nil {
    return err
}

// Or for containerized deployments
cfg, err := config.Load("")  // Empty path = env vars + defaults only
```

### Best Practices

**DO**:
- ✅ Support environment-only deployment
- ✅ Provide `config.example.toml` for local development
- ✅ Use defaults for non-critical settings
- ✅ Validate configuration on startup
- ✅ Document all environment variables in README

**DON'T**:
- ❌ Commit secrets in config files
- ❌ Require config files in production
- ❌ Hard-code configuration values
- ❌ Bundle config files in Docker images
- ❌ Use different config structures per environment

### Docker Configuration

```dockerfile
# Dockerfile - No config files bundled
FROM alpine:latest
COPY --from=builder /app/bin/app .
# ❌ DON'T: COPY configs/ ./configs/
# ✅ DO: Use environment variables
ENTRYPOINT ["./app"]
```

```yaml
# docker-compose.yaml
services:
  api:
    environment:
      APP_SERVER_PORT: 8080
      APP_DATABASE_HOST: postgres
      APP_DATABASE_USER: ${DB_USER}
      APP_DATABASE_PASSWORD: ${DB_PASSWORD}
```

### Environment-Specific Configuration

```bash
# Development (local)
export APP_SERVER_MODE=debug
export APP_LOGGER_LEVEL=debug

# Staging
export APP_SERVER_MODE=release
export APP_LOGGER_LEVEL=info
export APP_DATABASE_HOST=staging-db.example.com

# Production
export APP_SERVER_MODE=release
export APP_LOGGER_LEVEL=warn
export APP_DATABASE_HOST=prod-db.example.com
export APP_DATABASE_PASSWORD=$(vault read -field=password secret/db)
```

---

## Database Standards

### Database Access Layer Standards

This project adopts a **"sqlc primary, sqlx+Squirrel auxiliary"** approach for database access.

**Primary Tool (sqlc):**
- ✅ **USE FOR**: All static queries (INSERT, UPDATE specific fields, SELECT by ID, fixed JOINs)
- ✅ **BENEFITS**: Strict type safety, compile-time query validation, best performance
- ✅ **IDEAL FOR**: CRUD operations, simple queries, queries that rarely change

**Secondary Tool (sqlx + Squirrel):**
- ✅ **USE FOR**: Dynamic search queries (listing APIs with multiple optional filters)
- ✅ **BENEFITS**: Flexible query building, handles dynamic WHERE clauses elegantly
- ✅ **IDEAL FOR**: Complex search, filtering, pagination with dynamic conditions

**Decision Matrix:**

| Scenario | Recommended Tool | Reason |
|----------|-----------------|--------|
| Get user by ID | **sqlc** | Static query, type-safe, fast |
| Create/Update user | **sqlc** | Fixed fields, compile-time validation |
| Search products with filters | **sqlx + Squirrel** | Dynamic WHERE clauses |
| List orders with pagination | **sqlx + Squirrel** | Flexible sorting and filtering |
| Complex multi-table JOIN | **sqlc** (if static) or **sqlx + Squirrel** (if dynamic) | Depends on query complexity |

**Example Comparison:**

```go
// ✅ sqlc: For static queries
// query.sql
-- name: GetUser :one
SELECT * FROM users WHERE id = ? LIMIT 1;

-- name: CreateUser :exec
INSERT INTO users (email, password_hash, created_at, updated_at)
VALUES (?, ?, ?, ?);

// Generated code (type-safe, no runtime reflection)
func (q *Queries) GetUser(ctx context.Context, id int64) (User, error) {
    row := q.db.QueryRowContext(ctx, getUser, id)
    var i User
    err := row.Scan(&i.ID, &i.Email, &i.PasswordHash, &i.CreatedAt, &i.UpdatedAt)
    return i, err
}

// Repository implementation
func (r *Repo) GetUser(ctx context.Context, id int64) (*domain.User, error) {
    u, err := r.queries.GetUser(ctx, id)
    if err != nil {
        return nil, err
    }
    return domain.ReconstructUser(u.ID, u.Email, u.PasswordHash, u.CreatedAt, u.UpdatedAt), nil
}
```

```go
// ✅ sqlx + Squirrel: For dynamic queries
func (r *Repo) SearchProducts(ctx context.Context, filters ProductFilters) ([]*domain.Product, error) {
    query := psql.Select("*").From("products")

    // Build dynamic WHERE clauses
    if filters.Category != "" {
        query = query.Where(squirrel.Eq{"category": filters.Category})
    }
    if filters.MinPrice > 0 {
        query = query.Where(squirrel.GtOrEq{"price": filters.MinPrice})
    }
    if filters.MaxPrice > 0 {
        query = query.Where(squirrel.LtOrEq{"price": filters.MaxPrice})
    }
    if filters.InStock {
        query = query.Where(squirrel.Gt{"stock": 0})
    }

    // Dynamic sorting
    if filters.SortBy != "" {
        query = query.OrderBy(filters.SortBy + " " + filters.SortOrder)
    }

    sql, args, _ := query.ToSql()
    var models []ProductModel
    err := r.db.SelectContext(ctx, &models, sql, args...)
    // ...
}
```

**Migration Path:**

If your project currently uses plain sqlx (without Squirrel):
1. Identify static queries → migrate to **sqlc**
2. Identify dynamic queries → migrate to **sqlx + Squirrel**
3. Update CODING_STANDARDS.md examples to reflect the chosen approach

**Note:** This project's example code currently uses plain sqlx. Teams should evaluate whether to migrate to sqlc for static queries based on project requirements.

---

### Migration Files

**IMPORTANT**: Database migrations are **mandatory** when using sqlx (unlike GORM which has AutoMigrate).
This project uses **[Atlas](https://atlasgo.io/)** as the default migration tool for its modern features:
- Versioned migrations with automatic tracking
- Schema validation and safety checks
- Multiple database support (PostgreSQL, MySQL, SQLite)
- Team collaboration with timestamp-based naming

**Alternative tools**: [golang-migrate](https://github.com/golang-migrate/migrate), [goose](https://github.com/pressly/goose)

**Installation:**
```bash
# macOS
brew install ariga/tap/atlas

# Linux
curl -sSf https://atlasgo.sh | sh

# Or use Make
make atlas-install
```

**Quick Commands:**
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

**Migration File Example:**
```sql
-- migrations/001_create_users_table.sql
CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_users_email ON users(email);
```

**Note**: Atlas uses timestamp-based naming for new migrations (e.g., `20240115120000_add_products_table.sql`) to avoid conflicts in team environments. See [migrations/README.md](migrations/README.md) for detailed usage.

### Repository Patterns

#### Transaction Handling with sqlx

Unlike GORM's closure-based transactions, sqlx requires manual transaction management:

```go
// ✅ Good: Use sqlx transactions with strict error handling
// IMPORTANT: Use named return value to ensure defer captures errors correctly
func (r *OrderRepository) CreateOrderWithItems(ctx context.Context, order *domain.Order) (err error) {
    // 1. Begin transaction
    tx, err := r.db.BeginTxx(ctx, nil)
    if err != nil {
        return fmt.Errorf("failed to begin transaction: %w", err)
    }

    // Ensure rollback on panic or error
    // This pattern prevents issues with variable shadowing
    defer func() {
        if p := recover(); p != nil {
            _ = tx.Rollback()
            panic(p) // Re-throw panic after rollback
        } else if err != nil {
            _ = tx.Rollback()
        }
    }()

    // 2. Prepare SQL using Squirrel
    orderModel := ToOrderModel(order)
    q1, args1, err := psql.Insert("orders").
        Columns("user_id", "total_amount", "status").
        Values(orderModel.UserID, orderModel.TotalAmount, orderModel.Status).
        Suffix("RETURNING id").
        ToSql()
    if err != nil {
        return err
    }

    // 3. Execute within transaction (use tx, not r.db)
    if err := tx.QueryRowxContext(ctx, q1, args1...).Scan(&orderModel.ID); err != nil {
        return fmt.Errorf("failed to create order: %w", err)
    }

    // 4. Handle related data
    for _, item := range order.GetItems() {
        q2, args2, err := psql.Insert("order_items").
            Columns("order_id", "product_id", "quantity").
            Values(orderModel.ID, item.ProductID, item.Quantity).
            ToSql()
        if err != nil {
            return err
        }

        if _, err := tx.ExecContext(ctx, q2, args2...); err != nil {
            return fmt.Errorf("failed to create order item: %w", err)
        }
    }

    // 5. Commit transaction
    if err = tx.Commit(); err != nil {
        return fmt.Errorf("failed to commit transaction: %w", err)
    }

    return nil
}
```

#### Query Building with Squirrel

**Important**: Configure the correct placeholder format for your database:

```go
import "github.com/Masterminds/squirrel"

// PostgreSQL: uses $1, $2, $3
var psql = squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

// MySQL: uses ? placeholders (default)
var mysql = squirrel.StatementBuilder.PlaceholderFormat(squirrel.Question)
```

**Common patterns**:

```go
// SELECT with WHERE
query, args, err := psql.Select("*").
    From("users").
    Where(squirrel.Eq{"id": userID}).
    ToSql()

// SELECT with multiple conditions
query, args, err := psql.Select("*").
    From("orders").
    Where(squirrel.And{
        squirrel.Eq{"user_id": userID},
        squirrel.Gt{"created_at": startDate},
    }).
    ToSql()

// UPDATE
query, args, err := psql.Update("users").
    Set("email", newEmail).
    Set("updated_at", time.Now()).
    Where(squirrel.Eq{"id": userID}).
    ToSql()

// DELETE
query, args, err := psql.Delete("users").
    Where(squirrel.Eq{"id": userID}).
    ToSql()
```

---

## Application Lifecycle

### Graceful Shutdown (Critical Requirement)

All application modes (API, Worker, Cron) **MUST** implement graceful shutdown to ensure:
- In-flight requests complete before shutdown
- Database connections are properly closed
- Message queues are gracefully drained
- No data loss occurs during deployments

**This is a Twelve-Factor App requirement (Factor IX - Disposability).**

### Signal Handling

Applications must handle the following signals:
- `SIGTERM` - Graceful shutdown (from orchestrator)
- `SIGINT` - Graceful shutdown (Ctrl+C)

```go
// ✅ Good: Proper signal handling
import (
    "context"
    "os"
    "os/signal"
    "syscall"
    "time"
)

func (s *APIServer) Start() error {
    // Create HTTP server
    server := &http.Server{
        Addr:    fmt.Sprintf(":%d", s.port),
        Handler: router,
    }

    // Channel for server errors
    serverErrors := make(chan error, 1)

    // Start server in goroutine
    go func() {
        logger.Info("API server listening", "address", server.Addr)
        serverErrors <- server.ListenAndServe()
    }()

    // Channel for shutdown signals
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
        if err := server.Shutdown(ctx); err != nil {
            logger.Error("Graceful shutdown failed", "error", err)
            // Force close if graceful fails
            server.Close()
            return fmt.Errorf("failed to gracefully shutdown: %w", err)
        }

        // Close database connections
        if err := s.db.Close(); err != nil {
            logger.Error("Failed to close database", "error", err)
        }

        logger.Info("Server stopped gracefully")
        return nil
    }
}
```

### API Server Shutdown

**Requirements**:
- ✅ Stop accepting new requests
- ✅ Complete in-flight HTTP requests
- ✅ Close database connections
- ✅ Timeout after 30 seconds maximum

**Example**: See `internal/app/api.go:40-80`

### Worker Shutdown

**Requirements**:
- ✅ Stop consuming new messages
- ✅ Complete processing of current messages
- ✅ Acknowledge/nack messages properly
- ✅ Close message queue connections
- ✅ Close database connections

```go
// ✅ Good: Worker graceful shutdown
func (w *Worker) Start() error {
    consumer := InitializeOrderConsumer(w.db, w.cfg)

    // Shutdown signal channel
    shutdown := make(chan os.Signal, 1)
    signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

    // Consumer error channel
    consumerErrors := make(chan error, 1)

    go func() {
        consumerErrors <- consumer.Start()
    }()

    select {
    case err := <-consumerErrors:
        return fmt.Errorf("consumer error: %w", err)

    case sig := <-shutdown:
        logger.Info("Shutdown signal received", "signal", sig)

        ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
        defer cancel()

        // Gracefully shutdown consumer
        if err := consumer.Shutdown(ctx); err != nil {
            logger.Error("Failed to shutdown consumer", "error", err)
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

### Cron Job Shutdown

**Requirements**:
- ✅ Stop accepting new job triggers
- ✅ Wait for running jobs to complete
- ✅ Timeout after 30 seconds
- ✅ Close database connections

```go
// ✅ Good: Cron graceful shutdown
func (c *CronScheduler) Start() error {
    scheduler := cron.New()
    // ... add jobs ...
    scheduler.Start()

    // Wait for shutdown signal
    shutdown := make(chan os.Signal, 1)
    signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

    <-shutdown
    logger.Info("Shutdown signal received")

    // Stop scheduler and get context
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

### Testing Graceful Shutdown

Create a test script to verify graceful shutdown:

```bash
#!/bin/bash
# scripts/test-graceful-shutdown.sh

echo "Starting API server..."
./app api &
PID=$!

sleep 2

echo "Sending SIGTERM to process $PID"
kill -TERM $PID

wait $PID
EXIT_CODE=$?

if [ $EXIT_CODE -eq 0 ]; then
    echo "✅ Graceful shutdown successful"
else
    echo "❌ Graceful shutdown failed with exit code $EXIT_CODE"
    exit 1
fi
```

### Expected Shutdown Behavior

```
[INFO] API server listening address=:8080
[INFO] Received 3 requests
[INFO] Shutdown signal received signal=SIGTERM
[INFO] Stopping server, waiting for 3 in-flight requests
[INFO] All requests completed
[INFO] Closing database connections
[INFO] Server stopped gracefully
```

### Common Mistakes

```go
// ❌ Bad: No graceful shutdown
func (s *APIServer) Start() error {
    router := gin.Default()
    // ...
    return router.Run(":8080")  // Blocks forever, no cleanup
}

// ❌ Bad: Force close immediately
func (s *APIServer) Start() error {
    // ...
    <-shutdown
    s.server.Close()  // Abruptly closes all connections
    return nil
}

// ❌ Bad: No timeout
func (s *APIServer) Start() error {
    // ...
    <-shutdown
    s.server.Shutdown(context.Background())  // May wait forever
    return nil
}
```

---

## Generated Code Policy

This project uses code generation tools for dependency injection, protobuf, and mocks. Generated code **MUST be committed** to ensure the repository is buildable immediately after cloning, following Go ecosystem best practices.

### Why Commit Generated Code?

**✅ Benefits**:
- **Immediate Buildability**: `git clone` → `go build` works without installing code generation tools
- **CI/CD Simplicity**: No need to install `wire`, `protoc`, and plugins in CI pipeline
- **Lower Developer Onboarding**: New developers don't need to set up complex toolchains
- **Library Compatibility**: If used as a Go library, `go get` will fetch working code
- **Code Review**: Changes to generated code help reviewers catch breaking changes

**⚠️ Trade-offs**:
- Larger repository size (usually negligible for Go projects)
- Generated code in pull requests (can be hidden with `.gitattributes`)

### Files to Commit

| Generated File | Tool | Commit? | Reason |
|----------------|------|---------|--------|
| `wire_gen.go` | Wire | ✅ **MUST** | Required for compilation; stable output |
| `*.pb.go` | protoc | ✅ **MUST** | Required for compilation; ensures version consistency |
| `mock_*.go` | mockery | ✅ **SHOULD** | Simplifies testing; reviewable interface changes |

### Keeping Generated Files Up-to-Date

**Developer Workflow**:
```bash
# After modifying wire.go, regenerate
wire gen ./internal/app

# After modifying .proto files, regenerate
make proto-gen

# After modifying interfaces, regenerate mocks
make mock-gen

# Or regenerate everything
make gen
```

**Best Practices**:
- ✅ **DO**: Regenerate before committing
- ✅ **DO**: Include generated files in the same commit as source changes
- ✅ **DO**: Use consistent tool versions (defined in Makefile/Dockerfile)
- ✅ **DO**: Run `make gen` in CI to verify files are up-to-date
- ❌ **DON'T**: Edit generated files manually (changes will be overwritten)
- ❌ **DON'T**: Commit source changes without regenerating

### CI Verification

Add a CI step to ensure generated files are up-to-date:

```yaml
# .github/workflows/ci.yml
name: CI

on: [push, pull_request]

jobs:
  verify-generated:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.24'

      - name: Install generation tools
        run: |
          go install github.com/google/wire/cmd/wire@latest
          go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
          go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

      - name: Regenerate all code
        run: make gen

      - name: Check for uncommitted changes
        run: |
          if [ -n "$(git status --porcelain)" ]; then
            echo "❌ Generated files are out of date. Run 'make gen' and commit the changes."
            git diff
            exit 1
          fi
          echo "✅ All generated files are up to date"
```

### Makefile Integration

Ensure your `Makefile` has a `gen` target that regenerates all code:

```makefile
.PHONY: gen
gen: wire-gen proto-gen mock-gen ## Regenerate all generated code

.PHONY: wire-gen
wire-gen: ## Regenerate Wire dependency injection code
	@echo "Generating Wire code..."
	@command -v wire >/dev/null 2>&1 || go install github.com/google/wire/cmd/wire@latest
	wire gen ./internal/app

.PHONY: proto-gen
proto-gen: ## Generate Go code from proto files
	@echo "Generating protobuf code..."
	@command -v protoc-gen-go >/dev/null 2>&1 || go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	@command -v protoc-gen-go-grpc >/dev/null 2>&1 || go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		api/proto/**/*.proto

.PHONY: mock-gen
mock-gen: ## Generate mocks for testing
	@echo "Generating mocks..."
	@command -v mockery >/dev/null 2>&1 || go install github.com/vektra/mockery/v2@latest
	mockery --dir=internal/usecase/port --all --output=internal/usecase/port/mocks
```

### When NOT to Commit Generated Code

Only skip committing generated files if:
- ✅ Your team uses **mandatory** Dev Containers with consistent tooling
- ✅ Your CI pipeline is **extremely fast** at code generation
- ✅ Generated files are **very large** (hundreds of MB) and impact Git performance
- ✅ Your project is **not used as a library** by other Go projects

For most projects, **committing generated code is the right choice**.

---

## Git Workflow

### Branch Naming

```bash
# Feature branches
feature/user-authentication
feature/order-checkout

# Bug fixes
fix/password-validation
fix/order-status-update

# Hotfixes
hotfix/security-vulnerability

# Claude Code branches (auto-generated)
claude/feature-name-sessionID
```

### Commit Messages

Follow [Conventional Commits](https://www.conventionalcommits.org/):

```bash
# Format
<type>(<scope>): <subject>

# Types
feat:     # New feature
fix:      # Bug fix
docs:     # Documentation only
style:    # Formatting, missing semicolons, etc.
refactor: # Code refactoring
test:     # Adding tests
chore:    # Maintenance tasks

# Examples
feat(user): add user registration endpoint
fix(order): fix order status transition bug
docs(readme): update installation instructions
refactor(repository): simplify database query logic
test(domain): add user entity validation tests
```

### Pull Request Guidelines

1. **Title**: Clear and descriptive
2. **Description**: Explain what and why
3. **Testing**: Include test results
4. **Breaking Changes**: Clearly marked
5. **Screenshots**: For UI changes

---

## Checklist for New Features

- [ ] Domain entity created with business logic
- [ ] Repository interface defined in `usecase/port/`
- [ ] Repository implementation with tests
- [ ] Use case interactor with tests
- [ ] HTTP handler with DTOs
- [ ] Routes registered
- [ ] Wire dependencies configured
- [ ] Error handling implemented
- [ ] Logging added
- [ ] Documentation updated
- [ ] API tests written
- [ ] Code formatted and linted

---

## Tools and IDE Setup

### Required Tools

```bash
# Go toolchain
go install golang.org/x/tools/cmd/goimports@latest

# Wire for DI
go install github.com/google/wire/cmd/wire@latest

# Mockery for mocks
go install github.com/vektra/mockery/v2@latest

# Linter
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

### VS Code Settings

```json
{
  "go.formatTool": "goimports",
  "go.lintTool": "golangci-lint",
  "go.lintOnSave": "package",
  "editor.formatOnSave": true,
  "editor.codeActionsOnSave": {
    "source.organizeImports": true
  }
}
```

---

## References

- [Clean Architecture by Robert C. Martin](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html)
- [Effective Go](https://golang.org/doc/effective_go)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- [Standard Go Project Layout](https://github.com/golang-standards/project-layout)
- [Google Wire](https://github.com/google/wire)
- [OOPS Error Handling](https://github.com/samber/oops)

---

## License

This coding standards document is part of the go-clean-arch project and is licensed under MIT.
