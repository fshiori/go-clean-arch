# Go Clean Architecture Reference

**For AI Assistants** - This document provides comprehensive architectural guidelines for Go projects implementing Clean Architecture. Use this as a reference when generating or reviewing code.

**Version**: 1.0
**Last Updated**: 2025-01-23

---

## Table of Contents

1. [Architecture Overview](#architecture-overview)
2. [Directory Structure](#directory-structure)
3. [Layer Rules and Examples](#layer-rules-and-examples)
4. [Naming Conventions](#naming-conventions)
5. [Error Handling](#error-handling)
6. [Testing Standards](#testing-standards)
7. [Database Patterns](#database-patterns)
8. [Dependency Injection](#dependency-injection)
9. [Configuration Management](#configuration-management)
10. [Application Lifecycle](#application-lifecycle)
11. [Twelve-Factor App Essentials](#twelve-factor-app-essentials)
12. [Adding New Features](#adding-new-features)
13. [Common Pitfalls](#common-pitfalls)
14. [Quick Reference Tables](#quick-reference-tables)

---

## Architecture Overview

### Core Principle: Dependency Inversion

**Outer layers depend on inner layers. Inner layers know nothing about outer layers.**

```
┌─────────────────────────────────────────┐
│          Delivery Layer                 │  ← HTTP, Worker, Cron, Microservice
│      (Receives external requests)       │     Depends on: Use Case
└──────────────┬──────────────────────────┘
               │
               ▼
┌─────────────────────────────────────────┐
│          Use Case Layer                 │  ← Application business logic
│   (Orchestrates domain logic)           │     Depends on: Domain
│   (Defines port interfaces)             │     Defines: Repository/Gateway interfaces
└──────────────┬──────────────────────────┘
               │
               ▼
┌─────────────────────────────────────────┐
│          Domain Layer                   │  ← Pure business entities & rules
│    (Entities, Business Rules)           │     Depends on: Nothing (innermost)
└─────────────────────────────────────────┘

               ▲
               │ implements port interfaces
┌─────────────────────────────────────────┐
│          Adapter Layer                  │  ← Repository & Gateway implementations
│   (Connects to external systems)        │     Implements: Use Case port interfaces
│                                         │     Depends on: Domain (for conversion)
└─────────────────────────────────────────┘
```

### Dependency Summary

| Layer | Depends On | Depended By | Responsibility |
|-------|-----------|-------------|----------------|
| **Domain** | Nothing | Use Case, Adapter | Pure business rules, entities |
| **Use Case** | Domain | Delivery | Application logic, orchestration |
| **Adapter** | Domain, Use Case (interfaces) | None (injected via DI) | Infrastructure implementations |
| **Delivery** | Use Case | None (entry point) | HTTP/gRPC/MQ handling |

---

## Directory Structure

```
.
├── cmd/                          # Application entry points
│   └── app/
│       └── main.go               # Main entry point (Cobra CLI)
├── internal/                     # Private application code
│   ├── domain/                   # Domain entities and business rules
│   │   ├── user.go               # Entity with business logic
│   │   ├── user_test.go
│   │   └── errors.go             # Domain error definitions
│   ├── usecase/                  # Application business logic
│   │   ├── port/                 # Port interfaces (repositories, gateways)
│   │   │   ├── user_repository.go
│   │   │   └── payment_gateway.go
│   │   ├── user_interactor.go    # Use case implementation
│   │   ├── user_interactor_test.go
│   │   └── interfaces.go         # Use case interfaces
│   ├── adapter/                  # Adapter implementations
│   │   ├── repository/           # Database access
│   │   │   ├── user_repository_sqlc.go
│   │   │   ├── user_model.go     # DB model with tags
│   │   │   ├── queries/          # SQL files for sqlc
│   │   │   │   └── user.sql
│   │   │   └── sqlcgen/          # Generated sqlc code
│   │   └── gateway/              # External service clients
│   │       └── stripe_gateway.go
│   ├── delivery/                 # Delivery mechanisms
│   │   ├── http/
│   │   │   ├── handler/
│   │   │   │   ├── user_handler.go
│   │   │   │   └── user_dto.go   # DTOs with JSON tags
│   │   │   └── router.go
│   │   ├── consumer/             # Message queue consumers
│   │   └── job/                  # Cron jobs
│   └── app/                      # Application wiring
│       ├── wire.go               # Wire DI definitions
│       └── wire_gen.go           # Generated (don't edit)
├── pkg/                          # Public shared libraries
│   ├── config/
│   ├── logger/
│   └── middleware/
├── migrations/                   # Database migrations
├── configs/                      # Configuration files (optional)
└── api/                          # API definitions (OpenAPI, proto)
```

---

## Layer Rules and Examples

### 1. Domain Layer (`internal/domain/`)

**Rules**:
- ✅ Pure business logic and entities
- ✅ Rich domain models with behavior (not anemic)
- ✅ Validate in constructors and methods
- ✅ Use factory functions (`NewXxx`) for creation
- ✅ Use reconstruction functions (`ReconstructXxx`) for DB retrieval
- ❌ NO imports from other application layers
- ❌ NO infrastructure tags (`json:`, `db:`, `gorm:`)
- ❌ NO framework dependencies

**Example**:
```go
// internal/domain/user.go
package domain

import (
    "time"
    "golang.org/x/crypto/bcrypt"
)

type User struct {
    ID        int64
    Email     string
    password  string    // Private - protected by methods
    CreatedAt time.Time
    UpdatedAt time.Time
}

// NewUser creates a new user with validation
func NewUser(email, password string) (*User, error) {
    if email == "" {
        return nil, ErrInvalidEmail
    }
    if len(password) < 8 {
        return nil, ErrInvalidPassword
    }

    hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    if err != nil {
        return nil, err
    }

    now := time.Now()
    return &User{
        Email:     email,
        password:  string(hashed),
        CreatedAt: now,
        UpdatedAt: now,
    }, nil
}

// ReconstructUser rebuilds entity from storage (no validation)
func ReconstructUser(id int64, email, passwordHash string, createdAt, updatedAt time.Time) *User {
    return &User{
        ID:        id,
        Email:     email,
        password:  passwordHash,
        CreatedAt: createdAt,
        UpdatedAt: updatedAt,
    }
}

// Password returns the hashed password (read-only)
func (u *User) Password() string {
    return u.password
}

// ChangePassword enforces business rules for password changes
func (u *User) ChangePassword(oldPassword, newPassword string) error {
    if !u.IsPasswordCorrect(oldPassword) {
        return ErrInvalidPassword
    }
    if len(newPassword) < 8 {
        return ErrInvalidPassword
    }

    hashed, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
    if err != nil {
        return err
    }

    u.password = string(hashed)
    u.UpdatedAt = time.Now()
    return nil
}

// IsPasswordCorrect checks password against stored hash
func (u *User) IsPasswordCorrect(password string) bool {
    err := bcrypt.CompareHashAndPassword([]byte(u.password), []byte(password))
    return err == nil
}
```

**Domain Errors**:
```go
// internal/domain/errors.go
package domain

import "errors"

// Error codes for structured error handling
const (
    ErrCodeUserNotFound       = "USER_NOT_FOUND"
    ErrCodeEmailAlreadyExists = "EMAIL_ALREADY_EXISTS"
    ErrCodeInvalidPassword    = "INVALID_PASSWORD"
    ErrCodeInvalidEmail       = "INVALID_EMAIL"
)

// Sentinel errors
var (
    ErrUserNotFound       = errors.New("user not found")
    ErrEmailAlreadyExists = errors.New("email already exists")
    ErrInvalidPassword    = errors.New("invalid password")
    ErrInvalidEmail       = errors.New("invalid email")
)
```

---

### 2. Use Case Layer (`internal/usecase/`)

**Rules**:
- ✅ Define port interfaces in `port/` subdirectory
- ✅ Depend only on domain entities and port interfaces
- ✅ Accept `context.Context` as first parameter
- ✅ One interactor per entity/aggregate
- ✅ Orchestrate domain logic (don't duplicate it)
- ❌ NO imports from delivery or adapter packages
- ❌ NO framework-specific code

**Port Interface**:
```go
// internal/usecase/port/user_repository.go
package port

import (
    "context"
    "{{MODULE}}/internal/domain"
)

type UserRepository interface {
    Save(ctx context.Context, user *domain.User) error
    FindByID(ctx context.Context, id int64) (*domain.User, error)
    FindByEmail(ctx context.Context, email string) (*domain.User, error)
    Update(ctx context.Context, user *domain.User) error
    Delete(ctx context.Context, id int64) error
    List(ctx context.Context, offset, limit int) ([]*domain.User, int64, error)
}
```

**Use Case Interface**:
```go
// internal/usecase/interfaces.go
package usecase

import (
    "context"
    "{{MODULE}}/internal/domain"
)

type UserUsecase interface {
    CreateUser(ctx context.Context, email, password string) (*domain.User, error)
    GetUserByID(ctx context.Context, id int64) (*domain.User, error)
    UpdateUserPassword(ctx context.Context, id int64, oldPassword, newPassword string) error
    DeleteUser(ctx context.Context, id int64) error
    ListUsers(ctx context.Context, page, pageSize int) ([]*domain.User, int64, error)
}
```

**Interactor Implementation**:
```go
// internal/usecase/user_interactor.go
package usecase

import (
    "context"
    "{{MODULE}}/internal/domain"
    "{{MODULE}}/internal/usecase/port"
    "github.com/samber/oops"
)

type UserInteractor struct {
    userRepo port.UserRepository
}

func NewUserInteractor(userRepo port.UserRepository) *UserInteractor {
    return &UserInteractor{userRepo: userRepo}
}

func (i *UserInteractor) CreateUser(ctx context.Context, email, password string) (*domain.User, error) {
    // Check if email exists
    existing, err := i.userRepo.FindByEmail(ctx, email)
    if err == nil && existing != nil {
        return nil, oops.
            Code(domain.ErrCodeEmailAlreadyExists).
            Wrapf(domain.ErrEmailAlreadyExists, "email %s already registered", email)
    }

    // Create domain entity (validation happens in domain)
    user, err := domain.NewUser(email, password)
    if err != nil {
        return nil, oops.
            Code(domain.ErrCodeInvalidPassword).
            Wrapf(err, "failed to create user")
    }

    // Persist via repository
    if err := i.userRepo.Save(ctx, user); err != nil {
        return nil, oops.
            Code("DATABASE_ERROR").
            Wrapf(err, "failed to save user")
    }

    return user, nil
}

func (i *UserInteractor) GetUserByID(ctx context.Context, id int64) (*domain.User, error) {
    user, err := i.userRepo.FindByID(ctx, id)
    if err != nil {
        return nil, oops.
            Code(domain.ErrCodeUserNotFound).
            With("user_id", id).
            Wrapf(err, "user not found")
    }
    return user, nil
}
```

---

### 3. Adapter Layer (`internal/adapter/`)

**Rules**:
- ✅ Implement port interfaces
- ✅ Transform between domain entities and external formats
- ✅ Use tags for DB/serialization in model structs
- ✅ Handle infrastructure concerns (DB, external APIs)
- ❌ NO business logic (belongs in domain)
- ❌ NO exposure of infrastructure details to use cases

**DB Model** (separate from domain entity):
```go
// internal/adapter/repository/user_model.go
package repository

import (
    "time"
    "{{MODULE}}/internal/domain"
)

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

func UserFromDomain(user *domain.User) *UserDBModel {
    return &UserDBModel{
        ID:           user.ID,
        Email:        user.Email,
        PasswordHash: user.Password(),
        CreatedAt:    user.CreatedAt,
        UpdatedAt:    user.UpdatedAt,
    }
}
```

**Repository Implementation** (sqlc + sqlx hybrid):
```go
// internal/adapter/repository/user_repository_sqlc.go
package repository

import (
    "context"
    "database/sql"
    "errors"

    "github.com/Masterminds/squirrel"
    "github.com/jmoiron/sqlx"
    "github.com/samber/oops"
    "{{MODULE}}/internal/adapter/repository/sqlcgen"
    "{{MODULE}}/internal/domain"
    "{{MODULE}}/internal/usecase/port"
)

// Placeholder format: MySQL uses ?, PostgreSQL uses $1
var sq = squirrel.StatementBuilder.PlaceholderFormat(squirrel.Question)

type userRepositorySQLC struct {
    db      *sqlx.DB
    queries *sqlcgen.Queries
}

func NewUserRepository(db *sqlx.DB) port.UserRepository {
    return &userRepositorySQLC{
        db:      db,
        queries: sqlcgen.New(db.DB),
    }
}

// Use sqlc for static queries
func (r *userRepositorySQLC) FindByID(ctx context.Context, id int64) (*domain.User, error) {
    user, err := r.queries.GetUserByID(ctx, id)
    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return nil, oops.
                Code(domain.ErrCodeUserNotFound).
                Wrapf(domain.ErrUserNotFound, "user %d not found", id)
        }
        return nil, oops.Wrapf(err, "database query failed")
    }
    return sqlcUserToDomain(&user), nil
}

func (r *userRepositorySQLC) Save(ctx context.Context, user *domain.User) error {
    result, err := r.queries.CreateUser(ctx, sqlcgen.CreateUserParams{
        Email:        user.Email,
        PasswordHash: user.Password(),
        CreatedAt:    user.CreatedAt,
        UpdatedAt:    user.UpdatedAt,
    })
    if err != nil {
        return oops.Wrapf(err, "failed to insert user")
    }

    id, _ := result.LastInsertId()
    user.ID = id
    return nil
}

// Use sqlx + Squirrel for dynamic queries
func (r *userRepositorySQLC) List(ctx context.Context, offset, limit int) ([]*domain.User, int64, error) {
    // Count total
    var total int64
    countQuery, countArgs, _ := sq.Select("COUNT(*)").From("users").ToSql()
    if err := r.db.GetContext(ctx, &total, countQuery, countArgs...); err != nil {
        return nil, 0, oops.Wrapf(err, "failed to count users")
    }

    // Get paginated results
    query, args, _ := sq.
        Select("id", "email", "password_hash", "created_at", "updated_at").
        From("users").
        OrderBy("created_at DESC").
        Limit(uint64(limit)).
        Offset(uint64(offset)).
        ToSql()

    var models []UserDBModel
    if err := r.db.SelectContext(ctx, &models, query, args...); err != nil {
        return nil, 0, oops.Wrapf(err, "failed to list users")
    }

    users := make([]*domain.User, len(models))
    for i, m := range models {
        users[i] = m.ToDomain()
    }

    return users, total, nil
}

func sqlcUserToDomain(u *sqlcgen.User) *domain.User {
    return domain.ReconstructUser(u.ID, u.Email, u.PasswordHash, u.CreatedAt, u.UpdatedAt)
}
```

---

### 4. Delivery Layer (`internal/delivery/`)

**Rules**:
- ✅ Keep handlers thin (validation and transformation only)
- ✅ Use DTOs with serialization tags
- ✅ Transform between DTOs and domain entities
- ✅ Handle HTTP/gRPC-specific concerns
- ❌ NO business logic (belongs in use case)
- ❌ NO direct repository access

**DTOs**:
```go
// internal/delivery/http/handler/user_dto.go
package handler

import (
    "time"
    "{{MODULE}}/internal/domain"
)

type CreateUserRequest struct {
    Email    string `json:"email" binding:"required,email"`
    Password string `json:"password" binding:"required,min=8"`
}

type UpdatePasswordRequest struct {
    OldPassword string `json:"old_password" binding:"required"`
    NewPassword string `json:"new_password" binding:"required,min=8"`
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
```

**Handler**:
```go
// internal/delivery/http/handler/user_handler.go
package handler

import (
    "net/http"
    "strconv"

    "github.com/gin-gonic/gin"
    "{{MODULE}}/internal/usecase"
)

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
        _ = c.Error(err) // Let error middleware handle it
        return
    }

    c.JSON(http.StatusCreated, ToUserResponse(user))
}

func (h *UserHandler) GetByID(c *gin.Context) {
    id, err := strconv.ParseInt(c.Param("id"), 10, 64)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
        return
    }

    user, err := h.userUC.GetUserByID(c.Request.Context(), id)
    if err != nil {
        _ = c.Error(err)
        return
    }

    c.JSON(http.StatusOK, ToUserResponse(user))
}
```

---

## Naming Conventions

### File Names

| Component | Pattern | Example |
|-----------|---------|---------|
| Domain Entity | `{entity}.go` | `user.go`, `order.go` |
| Domain Errors | `errors.go` | Single file for all domain errors |
| Repository Interface | `{entity}_repository.go` | `user_repository.go` |
| Gateway Interface | `{service}_gateway.go` | `payment_gateway.go` |
| Use Case Interface | `interfaces.go` | Single file for all use case interfaces |
| Use Case Impl | `{entity}_interactor.go` | `user_interactor.go` |
| Repository Impl | `{entity}_repository_sqlc.go` | `user_repository_sqlc.go` |
| DB Model | `{entity}_model.go` | `user_model.go` |
| HTTP Handler | `{entity}_handler.go` | `user_handler.go` |
| HTTP DTO | `{entity}_dto.go` | `user_dto.go` |
| Test File | `{filename}_test.go` | `user_test.go` |

### Code Names

```go
// Packages: lowercase, no underscores
package domain
package repository
package handler

// Interfaces: descriptive, often -er suffix for single method
type UserRepository interface {}
type PaymentGateway interface {}
type Reader interface { Read() }

// Structs: PascalCase
type UserInteractor struct {}
type CreateUserRequest struct {}

// Private fields: camelCase
type User struct {
    password string  // private
    Email    string  // public
}

// Factory functions: NewXxx
func NewUser(email, password string) (*User, error)
func NewUserInteractor(repo port.UserRepository) *UserInteractor

// Reconstruction: ReconstructXxx
func ReconstructUser(id int64, email, passwordHash string, ...) *User

// Errors: ErrXxx
var ErrUserNotFound = errors.New("user not found")

// Error codes: SCREAMING_SNAKE_CASE
const ErrCodeUserNotFound = "USER_NOT_FOUND"
```

---

## Error Handling

### Using oops (Recommended)

```go
import "github.com/samber/oops"

// In repository
func (r *repo) FindByID(ctx context.Context, id int64) (*domain.User, error) {
    user, err := r.queries.GetUserByID(ctx, id)
    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return nil, oops.
                Code(domain.ErrCodeUserNotFound).
                In("repository").
                With("user_id", id).
                Wrapf(domain.ErrUserNotFound, "user %d not found", id)
        }
        return nil, oops.
            Code("DATABASE_ERROR").
            In("repository").
            Hint("Check database connectivity").
            Wrapf(err, "failed to query user")
    }
    return sqlcUserToDomain(&user), nil
}
```

### Error Middleware

```go
// pkg/middleware/error_handler.go
func ErrorHandler() gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Next()

        if len(c.Errors) == 0 {
            return
        }

        err := c.Errors.Last().Err
        ooErr := oops.AsOops(err)

        statusCode := mapErrorToStatus(ooErr.Code())
        c.JSON(statusCode, gin.H{
            "error": err.Error(),
            "code":  ooErr.Code(),
        })
    }
}

func mapErrorToStatus(code string) int {
    switch code {
    case domain.ErrCodeUserNotFound:
        return http.StatusNotFound
    case domain.ErrCodeEmailAlreadyExists:
        return http.StatusConflict
    case domain.ErrCodeInvalidPassword, domain.ErrCodeInvalidEmail:
        return http.StatusBadRequest
    default:
        return http.StatusInternalServerError
    }
}
```

---

## Testing Standards

### Coverage Targets

| Layer | Target | Strategy |
|-------|--------|----------|
| Domain | 90%+ | Unit tests, no mocks needed |
| Use Case | 90%+ | Mock repositories via interfaces |
| Repository | 80%+ | Integration tests with test DB |
| Handler | 70%+ | Mock use cases |

### Table-Driven Tests

```go
func TestNewUser(t *testing.T) {
    tests := []struct {
        name     string
        email    string
        password string
        wantErr  error
    }{
        {
            name:     "valid user",
            email:    "test@example.com",
            password: "SecurePass123",
            wantErr:  nil,
        },
        {
            name:     "empty email",
            email:    "",
            password: "SecurePass123",
            wantErr:  domain.ErrInvalidEmail,
        },
        {
            name:     "short password",
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

### Mocking with testify

```go
// Use mockery to generate mocks:
// mockery --dir=internal/usecase/port --all --output=internal/usecase/port/mocks

func TestUserInteractor_CreateUser(t *testing.T) {
    mockRepo := new(mocks.UserRepository)
    interactor := usecase.NewUserInteractor(mockRepo)

    mockRepo.On("FindByEmail", mock.Anything, "test@example.com").
        Return(nil, domain.ErrUserNotFound)
    mockRepo.On("Save", mock.Anything, mock.AnythingOfType("*domain.User")).
        Return(nil)

    user, err := interactor.CreateUser(context.Background(), "test@example.com", "password123")

    assert.NoError(t, err)
    assert.NotNil(t, user)
    mockRepo.AssertExpectations(t)
}
```

---

## Database Patterns

### sqlc + sqlx Hybrid Approach

**Use sqlc for**: Static queries (CRUD, simple JOINs)
**Use sqlx + Squirrel for**: Dynamic queries (search, filters, pagination)

**sqlc Query File**:
```sql
-- internal/adapter/repository/queries/user.sql

-- name: GetUserByID :one
SELECT * FROM users WHERE id = ? LIMIT 1;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = ? LIMIT 1;

-- name: CreateUser :execresult
INSERT INTO users (email, password_hash, created_at, updated_at)
VALUES (?, ?, ?, ?);

-- name: UpdateUser :exec
UPDATE users SET email = ?, password_hash = ?, updated_at = ? WHERE id = ?;

-- name: DeleteUser :exec
DELETE FROM users WHERE id = ?;
```

**Generate**: `sqlc generate` or `make sqlc-gen`

### Transaction Handling

```go
func (r *OrderRepository) CreateOrderWithItems(ctx context.Context, order *domain.Order) (err error) {
    tx, err := r.db.BeginTxx(ctx, nil)
    if err != nil {
        return fmt.Errorf("failed to begin transaction: %w", err)
    }

    defer func() {
        if p := recover(); p != nil {
            _ = tx.Rollback()
            panic(p)
        } else if err != nil {
            _ = tx.Rollback()
        }
    }()

    // Insert order
    // Insert order items
    // ...

    if err = tx.Commit(); err != nil {
        return fmt.Errorf("failed to commit: %w", err)
    }
    return nil
}
```

### Unit of Work (Cross-Repository Transactions)

```go
// internal/usecase/port/unit_of_work.go
type UnitOfWork interface {
    DoWithRepositories(ctx context.Context, fn func(ctx context.Context, repos Repositories) error) error
}

type Repositories struct {
    UserRepo  UserRepository
    OrderRepo OrderRepository
}

// Usage in use case
func (i *OrderInteractor) CreateOrderWithRewards(ctx context.Context, req CreateOrderRequest) error {
    return i.uow.DoWithRepositories(ctx, func(ctx context.Context, repos port.Repositories) error {
        // 1. Create order (OrderRepository)
        // 2. Update user points (UserRepository)
        // Both in same transaction
        return nil
    })
}
```

---

## Dependency Injection

### Wire Setup

```go
// internal/app/wire.go
//go:build wireinject

package app

import (
    "github.com/google/wire"
    "{{MODULE}}/internal/adapter/repository"
    "{{MODULE}}/internal/delivery/http/handler"
    "{{MODULE}}/internal/usecase"
    "{{MODULE}}/internal/usecase/port"
)

var RepositorySet = wire.NewSet(
    repository.NewUserRepository,
    wire.Bind(new(port.UserRepository), new(*repository.userRepositorySQLC)),
)

var UseCaseSet = wire.NewSet(
    usecase.NewUserInteractor,
    wire.Bind(new(usecase.UserUsecase), new(*usecase.UserInteractor)),
)

var HandlerSet = wire.NewSet(
    handler.NewUserHandler,
)

func InitializeAPI(db *sqlx.DB) (*gin.Engine, error) {
    wire.Build(
        RepositorySet,
        UseCaseSet,
        HandlerSet,
        NewRouter,
    )
    return nil, nil
}
```

**Generate**: `wire gen ./internal/app`

---

## Configuration Management

### Environment Variables (Twelve-Factor)

```bash
# All config via environment variables
export APP_SERVER_PORT=8080
export APP_DATABASE_HOST=localhost
export APP_DATABASE_USER=postgres
export APP_DATABASE_PASSWORD=secret
export APP_LOGGER_LEVEL=info

# Run without config file
./app api
```

### Naming Convention

- Prefix: `APP_`
- Nested keys: `database.host` → `APP_DATABASE_HOST`
- All uppercase

### Config Structure

```go
// pkg/config/config.go
type Config struct {
    Server   ServerConfig   `mapstructure:"server"`
    Database DatabaseConfig `mapstructure:"database"`
    Logger   LoggerConfig   `mapstructure:"logger"`
}

type ServerConfig struct {
    Port int    `mapstructure:"port"`
    Host string `mapstructure:"host"`
}
```

---

## Application Lifecycle

### Graceful Shutdown (Required)

```go
func (s *APIServer) Start() error {
    server := &http.Server{
        Addr:    fmt.Sprintf(":%d", s.port),
        Handler: s.router,
    }

    serverErrors := make(chan error, 1)
    go func() {
        serverErrors <- server.ListenAndServe()
    }()

    shutdown := make(chan os.Signal, 1)
    signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

    select {
    case err := <-serverErrors:
        return fmt.Errorf("server error: %w", err)

    case <-shutdown:
        ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
        defer cancel()

        if err := server.Shutdown(ctx); err != nil {
            server.Close()
            return fmt.Errorf("graceful shutdown failed: %w", err)
        }

        s.db.Close()
        return nil
    }
}
```

---

## Twelve-Factor App Essentials

| Factor | Implementation |
|--------|----------------|
| **III. Config** | Environment variables with `APP_` prefix |
| **VI. Processes** | Stateless processes, no local state |
| **IX. Disposability** | Graceful shutdown with signal handling |
| **XI. Logs** | Structured logging to stdout |
| **XII. Admin** | Migration commands as one-off processes |

---

## Adding New Features

### 7-Step Process

1. **Domain**: Create entity in `internal/domain/{entity}.go`
2. **Port**: Define repository interface in `internal/usecase/port/{entity}_repository.go`
3. **Repository**: Implement in `internal/adapter/repository/{entity}_repository_sqlc.go`
4. **Use Case**: Create interactor in `internal/usecase/{entity}_interactor.go`
5. **DTO**: Add request/response in `internal/delivery/http/handler/{entity}_dto.go`
6. **Handler**: Implement in `internal/delivery/http/handler/{entity}_handler.go`
7. **Wire**: Update `internal/app/wire.go` and run `wire gen ./internal/app`

---

## Common Pitfalls

### Architecture Violations

```go
// ❌ DON'T: Import adapter in usecase
import "{{MODULE}}/internal/adapter/repository"

// ❌ DON'T: Import any layer in domain
import "{{MODULE}}/internal/usecase"

// ❌ DON'T: Add tags to domain entities
type User struct {
    ID int64 `json:"id" db:"id"`  // WRONG
}

// ❌ DON'T: Business logic in handlers
func (h *Handler) Create(c *gin.Context) {
    if len(req.Password) < 8 {  // WRONG - belongs in domain
        // ...
    }
}

// ❌ DON'T: Access repository in handlers
func (h *Handler) Create(c *gin.Context) {
    h.userRepo.Save(ctx, user)  // WRONG - go through use case
}
```

### Wire DI

```go
// ❌ DON'T: Edit wire_gen.go manually

// ✅ DO: Remember to run after changes
wire gen ./internal/app

// ✅ DO: Add interface bindings
wire.Bind(new(port.UserRepository), new(*repository.userRepositorySQLC))
```

### Context Propagation

```go
// ✅ DO: Accept context as first parameter
func (i *Interactor) CreateUser(ctx context.Context, ...) error

// ✅ DO: Propagate context to repositories
user, err := i.userRepo.FindByID(ctx, id)

// ✅ DO: Use context-aware logging
logger.InfoContext(ctx, "User created", "user_id", user.ID)
```

---

## Quick Reference Tables

### HTTP Status Code Mapping

| Domain Error | HTTP Status |
|--------------|-------------|
| `ErrUserNotFound` | 404 Not Found |
| `ErrEmailAlreadyExists` | 409 Conflict |
| `ErrInvalidPassword` | 400 Bad Request |
| `ErrUnauthorized` | 401 Unauthorized |
| Default | 500 Internal Server Error |

### Layer Responsibilities

| Layer | Creates | Transforms | Calls |
|-------|---------|------------|-------|
| Domain | Entities, Errors | - | - |
| Use Case | - | - | Domain, Repository (via interface) |
| Adapter | DB Models | Domain ↔ DB Model | Database, External APIs |
| Delivery | DTOs | Domain ↔ DTO | Use Case |

### Data Flow

```
HTTP Request → DTO → Use Case → Domain Entity → Repository → DB Model → Database
                                     ↓
HTTP Response ← DTO ← Use Case ← Domain Entity ← Repository ← DB Model ← Database
```

---

## Changelog

- **1.0** (2025-01-23): Initial version consolidated from go-clean-arch project
