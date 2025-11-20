# Architecture Documentation

## Overview

This project implements Clean Architecture principles, organizing code into distinct layers with clear dependencies flowing inward toward the domain.

## Layer Responsibilities

### 1. Domain Layer (`internal/domain/`)

**Purpose**: Contains core business entities and business rules.

**Characteristics**:
- No dependencies on any other layer
- Pure Go structs with business logic
- No infrastructure concerns (no database, HTTP, etc.)
- No tags (no `json:`, `gorm:`, etc.)

**Examples**:
- `User` entity with password validation
- `Order` entity with status management
- Business rule methods like `ChangePassword()`, `Ship()`, etc.

### 2. Use Case Layer (`internal/usecase/`)

**Purpose**: Contains application-specific business logic.

**Characteristics**:
- Orchestrates domain entities
- Defines interfaces it needs (`port` subdirectory)
- Independent of delivery mechanisms and external services
- Contains the "user stories" or "use cases"

**Structure**:
- `port/`: Interface definitions for repositories and gateways
- Interactors: Business logic implementations

**Examples**:
- `UserInteractor`: Handles user creation, password updates, etc.
- `OrderInteractor`: Handles order processing, checkout, etc.

### 3. Adapter Layer (`internal/adapter/`)

**Purpose**: Implements the ports (interfaces) defined by the use case layer.

**Characteristics**:
- Adapts external services to our domain needs
- Handles data transformation
- Contains infrastructure code (database, external APIs)

**Structure**:
- `repository/`: Database implementations
  - DB models with ORM tags
  - Repository implementations
  - Model ↔ Entity transformations
- `gateway/`: External API clients
  - Third-party service integrations
  - Request/Response transformations

**Examples**:
- `UserRepositoryGORM`: Implements `UserRepository` using GORM
- `StripeGateway`: Implements `PaymentGateway` using Stripe API

### 4. Delivery Layer (`internal/delivery/`)

**Purpose**: Handles interaction with the outside world.

**Characteristics**:
- Framework-specific code
- Protocol handlers (HTTP, gRPC, CLI, etc.)
- Contains DTOs with serialization tags
- Transforms between external formats and domain

**Structure**:
- `http/`: HTTP handlers and routes
  - DTOs with `json:` tags
  - Request/response handling
- `consumer/`: Message queue consumers
- `job/`: Scheduled jobs (cron)

**Examples**:
- `UserHandler`: HTTP endpoints for user operations
- `OrderConsumer`: Processes order messages from queue
- `DailyReportJob`: Generates daily reports

### 5. App Layer (`internal/app/`)

**Purpose**: Application bootstrapping and dependency injection.

**Characteristics**:
- Wires all layers together
- Handles dependency injection
- Different files for different runtime modes

**Structure**:
- `api.go`: Sets up API server
- `worker.go`: Sets up message queue worker
- `cron.go`: Sets up cron scheduler

## Data Flow

### Request Flow (API Example)

```
1. HTTP Request
   ↓
2. Handler (Delivery Layer)
   - Parse request
   - Validate input
   - Convert DTO to domain types
   ↓
3. Use Case (Interactor)
   - Execute business logic
   - Call repository/gateway interfaces
   ↓
4. Repository (Interface Layer)
   - Convert domain entity to DB model
   - Execute database query
   - Convert DB model back to domain entity
   ↓
5. Back to Use Case
   - Process result
   ↓
6. Back to Handler
   - Convert domain entity to response DTO
   - Return HTTP response
```

### Data Transformations

#### 1. Repository Layer Transformation
```
Domain Entity ←→ DB Model

Example:
domain.User {
    ID: 1,
    Email: "user@example.com",
    password: "hashed_..."
}
↕
repository.UserModel {
    ID: 1,
    Email: "user@example.com",
    PasswordHash: "hashed_...",
}
```

#### 2. Handler Layer Transformation
```
Domain Entity ←→ DTO

Example:
domain.User {
    ID: 1,
    Email: "user@example.com",
    password: "hashed_..."
}
↕
handler.UserResponse {
    ID: 1 `json:"id"`,
    Email: "user@example.com" `json:"email"`,
    // No password field!
}
```

## Dependency Rules

### The Dependency Rule

**Source code dependencies must point only inward, toward higher-level policies.**

```
Delivery → Use Case → Domain
   ↓           ↓
Interface  →  Port
```

- **Domain** depends on nothing
- **Use Case** depends only on Domain and Port interfaces
- **Adapter** depends on Use Case (Ports) and Domain
- **Delivery** depends on Use Case and Domain

### Why This Matters

1. **Testability**: Mock outer layers when testing inner layers
2. **Flexibility**: Change frameworks without touching business logic
3. **Maintainability**: Clear boundaries prevent coupling
4. **Independence**: Domain logic works regardless of UI or database

## Multi-Mode Architecture

This application supports three runtime modes from a single codebase:

### 1. API Mode
- Starts HTTP server
- Handles REST API requests
- Serves web clients

### 2. Worker Mode
- Connects to message queue
- Processes background jobs
- Handles async tasks

### 3. Cron Mode
- Runs scheduled jobs
- Generates reports
- Performs maintenance tasks

### Benefits

- **Code Reuse**: All modes share domain and use case logic
- **Consistency**: Same business rules everywhere
- **Simplified Deployment**: Single binary, different flags
- **Easier Testing**: Test business logic once, use everywhere

## Best Practices

### 1. Domain Layer
- ✅ Use factory functions (`NewUser()`)
- ✅ Validate in constructors
- ✅ Keep fields private when needed
- ✅ Provide business methods
- ❌ No external dependencies
- ❌ No tags

### 2. Use Case Layer
- ✅ Define interfaces in `port/`
- ✅ Depend only on interfaces (ports)
- ✅ One interactor per entity
- ❌ Don't import delivery or adapter packages

### 3. Adapter Layer
- ✅ Transform between models and entities
- ✅ Handle infrastructure concerns
- ✅ Implement port interfaces
- ✅ Connect to databases, external APIs, etc.
- ❌ Don't expose infrastructure details

### 4. Delivery Layer
- ✅ Keep handlers thin
- ✅ Use DTOs with tags
- ✅ Validate input
- ❌ Don't put business logic here

## Testing Strategy

### Unit Tests
- Test each layer independently
- Mock dependencies using interfaces
- Focus on business logic

### Integration Tests
- Test database repositories
- Test external gateway integrations
- Use test databases/services

### End-to-End Tests
- Test complete request flows
- Test different runtime modes
- Use Docker Compose for dependencies

## Further Reading

- [Clean Architecture by Robert C. Martin](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html)
- [Standard Go Project Layout](https://github.com/golang-standards/project-layout)
- [Dependency Inversion Principle](https://en.wikipedia.org/wiki/Dependency_inversion_principle)
