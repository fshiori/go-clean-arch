# Go Clean Architecture

A production-ready Go application implementing Clean Architecture principles with support for multiple runtime modes (API, Worker, Cron, Microservice).

> **For AI assistants**: See [CLAUDE.md](CLAUDE.md)
> **For contributors**: See [CODING_STANDARDS.md](CODING_STANDARDS.md)

## ✨ Highlights

- ✅ **Twelve-Factor App Compliant** (12/12 factors)
- ✅ **Clean Architecture** with clear separation of concerns
- ✅ **Multiple Runtime Modes**: API, Worker, Cron, Microservice from single codebase
- ✅ **Graceful Shutdown**: Proper signal handling for zero-downtime deployments
- ✅ **Environment-First Config**: Runs with only environment variables (no config files required)
- ✅ **Database Migrations**: Built-in migration commands (Factor XII compliant)
- ✅ **Production Ready**: Docker, Kubernetes compatible with best practices

## Quick Start

### Environment-Only Deployment (Twelve-Factor)

```bash
# Set configuration via environment variables
export APP_SERVER_PORT=8080
export APP_DATABASE_HOST=localhost
export APP_DATABASE_USER=postgres
export APP_DATABASE_PASSWORD=secret

# Run migrations
./app migrate up

# Start API server (no config file needed!)
./app api
```

### Traditional Deployment (with config file)

```bash
# Copy example config
cp configs/config.example.toml configs/config.toml

# Edit configuration
vim configs/config.toml

# Run migrations
./app migrate up

# Start server
./app api --config=configs/config.toml
```

## Architecture Overview

This project follows Clean Architecture and Standard Go Project Layout principles:

```
┌─────────────────────────────────────────┐
│          Delivery Layer                 │  ← HTTP, Workers, Cron, Microservice
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
               │ implements
┌─────────────────────────────────────────┐
│          Adapter Layer                  │  ← Repository & Gateway implementations
│   (Repositories, Gateways)              │     Implements: Use Case port interfaces
│   (Connects to external systems)        │     Depends on: Domain (for entity conversion)
└─────────────────────────────────────────┘
```

**Note**: Adapters implement the interfaces (ports) defined by Use Cases. They are injected at runtime via dependency injection, which is why they appear separately from the main dependency flow.

### Key Principles

- **Dependency Inversion**: Inner layers don't depend on outer layers
- **Domain-Driven Design**: Core business logic in the domain layer
- **Interface Segregation**: Interfaces defined in usecase/port
- **Single Responsibility**: Each layer has a clear purpose
- **Testability**: Easy to mock and test each layer

For detailed architecture documentation, see [CODING_STANDARDS.md](CODING_STANDARDS.md).

## Running the Application

### Prerequisites

- Go 1.22 or higher (tested with Go 1.23.x)
- Database: PostgreSQL, MySQL, or SQLite
- RabbitMQ (for worker mode, optional)

### Installation

```bash
# Clone the repository
git clone <repository-url>
cd go-clean-arch

# Install dependencies
go mod download

# Build
make build
```

### Configuration

**Option 1: Environment Variables (Recommended for Production)**

```bash
# Set all configuration via environment variables
export APP_SERVER_PORT=8080
export APP_DATABASE_HOST=localhost
export APP_DATABASE_PORT=5432
export APP_DATABASE_USER=postgres
export APP_DATABASE_PASSWORD=secret
export APP_DATABASE_DBNAME=go_clean_arch
export APP_LOGGER_LEVEL=info
export APP_LOGGER_FORMAT=json

# No config file needed!
./app api
```

**Option 2: Config File (For Local Development)**

```bash
cp configs/config.example.toml configs/config.toml
# Edit configs/config.toml with your settings
./app api --config=configs/config.toml
```

**Why Environment Variables?**

This application follows [Twelve-Factor App](https://12factor.net/) methodology:
- ✅ Same Docker image for dev/staging/prod
- ✅ No secrets in config files
- ✅ Easy Kubernetes/Docker deployment
- ✅ Environment-specific configuration

See [CODING_STANDARDS.md#twelve-factor-app-compliance](CODING_STANDARDS.md#twelve-factor-app-compliance) for details.

### Database Migrations

Run migrations before starting the application:

```bash
# Check migration status
./app migrate status

# Apply all pending migrations
./app migrate up

# Rollback last migration
./app migrate down

# Create new migration
./app migrate create "add_users_index"
```

**Note**: Auto-migrate is disabled by default. Use explicit migration commands for production safety.

### Running Different Modes

#### API Server Mode

```bash
# Start API server
./app api

# With custom port
./app api --port 9090
```

The API server will start on `http://localhost:8080`

#### Worker Mode

```bash
# Start message queue worker
./app worker
```

The worker will start consuming messages from the message queue.

**Note**: Worker mode is currently a skeleton implementation. RabbitMQ integration is commented out (marked as TODO). See `internal/delivery/consumer/` for implementation details.

#### Cron Scheduler Mode

```bash
# Start cron scheduler
./app cron
```

The cron scheduler will start and run scheduled jobs.

#### Microservice Mode

```bash
# Start with default settings
./app micro

# Start with custom settings
./app micro --service-name user.service --version v1.0.0 --address :8082
```

The microservice uses **go-micro v5** framework for RPC communication. It provides service discovery, load balancing, and pluggable transports out of the box.

**Key Features:**
- ✅ Full go-micro v5 integration
- ✅ Service registry and discovery
- ✅ Pluggable transports (HTTP, gRPC, etc.)
- ✅ Handler-based RPC with automatic serialization
- ✅ Clean architecture maintained

**Using HTTP transport:**

```bash
curl -X POST http://localhost:8081/UserServiceSimple/CreateUser \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"secret123"}'
```

For detailed microservice integration, see [CODING_STANDARDS.md#microservices-delivery-layer](CODING_STANDARDS.md#microservices-delivery-layer).

### Building

```bash
# Build for current platform
make build

# Build for all platforms
make build-all

# Run tests
make test

# Run linter
make lint
```

### Docker

```bash
# Build Docker image
docker build -t go-clean-arch:latest .

# Run API mode
docker run -p 8080:8080 \
  -e APP_DATABASE_HOST=host.docker.internal \
  -e APP_DATABASE_PASSWORD=secret \
  go-clean-arch:latest api

# Run Worker mode
docker run \
  -e APP_DATABASE_HOST=host.docker.internal \
  -e APP_RABBITMQ_HOST=host.docker.internal \
  go-clean-arch:latest worker

# Run Cron mode
docker run \
  -e APP_DATABASE_HOST=host.docker.internal \
  go-clean-arch:latest cron

# Run Microservice mode
docker run -p 8081:8081 \
  -e APP_DATABASE_HOST=host.docker.internal \
  go-clean-arch:latest micro --address :8081
```

## API Endpoints

### Users

- `POST /api/v1/users` - Create a new user
- `GET /api/v1/users` - List users (with pagination)
- `GET /api/v1/users/:id` - Get user by ID
- `PUT /api/v1/users/:id/password` - Update user password
- `DELETE /api/v1/users/:id` - Delete user

### Health Check

- `GET /health` - Health check endpoint

### Example Requests

```bash
# Create user
curl -X POST http://localhost:8080/api/v1/users \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"secret123"}'

# Get user
curl http://localhost:8080/api/v1/users/1

# List users
curl http://localhost:8080/api/v1/users?page=1&page_size=10
```

## Development

### Adding a New Feature

For detailed step-by-step guide, see [CLAUDE.md#adding-new-features](CLAUDE.md#adding-new-features) or [CODING_STANDARDS.md#clean-architecture-layers](CODING_STANDARDS.md#clean-architecture-layers).

Quick overview:
1. Define domain entity in `internal/domain/`
2. Define repository interface in `internal/usecase/port/`
3. Implement repository in `internal/adapter/repository/`
4. Create use case in `internal/usecase/`
5. Add HTTP handler in `internal/delivery/http/handler/`
6. Register routes in `internal/delivery/http/router.go`
7. Wire dependencies in `internal/app/wire.go`

### Testing

```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage

# Run tests for a specific package
go test ./internal/usecase/...
```

### Code Quality

```bash
# Format code
make fmt

# Run linter
make lint

# Run static analysis
make vet
```

## Documentation

- [CLAUDE.md](CLAUDE.md) - Quick reference for AI assistants
- [CODING_STANDARDS.md](CODING_STANDARDS.md) - Comprehensive coding standards
- [TWELVE_FACTOR_COMPLIANCE.md](TWELVE_FACTOR_COMPLIANCE.md) - Twelve-Factor compliance details

## Design Philosophy

This project follows:
- **Clean Architecture** by Robert C. Martin
- **Standard Go Project Layout**
- **Twelve-Factor App** methodology
- **Domain-Driven Design** principles

### Why Single Binary, Multiple Modes?

- **Code Reuse**: All modes share the same business logic
- **Simplified Dependencies**: Single `go.mod` file
- **Atomic Changes**: Changes to shared logic deploy together
- **Easier Development**: One repository to clone and work with

### Data Transformation Layers

1. **Domain Entity** ↔ **DB Model**: Transformed in Repository layer
2. **Domain Entity** ↔ **DTO**: Transformed in HTTP Handler layer
3. **Domain remains pure**: No database or JSON tags

For detailed design decisions, see [CODING_STANDARDS.md](CODING_STANDARDS.md).

## Technology Stack

- **Language**: Go 1.22+
- **HTTP Framework**: Gin
- **Database Access**:
  - **sqlc** (primary - type-safe SQL queries for static queries)
  - **sqlx + Squirrel** (auxiliary - dynamic query building)
  - Supports: PostgreSQL, MySQL, SQLite
- **Database Migrations**: Atlas (versioned migrations with validation)
- **Dependency Injection**: Wire (compile-time)
- **Configuration**: Viper (environment-first, config files optional)
- **Logging**: slog (structured logging)
- **Microservices**: go-micro v5
- **Message Queue**: RabbitMQ
- **Cron**: robfig/cron

> **Framework Agnostic Design (2025):**
> While this project uses **Gin** for its maturity and ecosystem, the Clean Architecture design allows us to swap the delivery mechanism easily.
>
> With the release of **Go 1.22+**, the standard library's `net/http` router has become powerful enough for many services. For new microservices requiring minimal dependencies, teams are encouraged to evaluate standard **`http.ServeMux`** or **Chi** as lightweight alternatives, provided they adhere to the same Handler/DTO patterns defined here.

## Contributing

Please read [CODING_STANDARDS.md](CODING_STANDARDS.md) for details on our coding standards and the process for submitting pull requests.

### Quick Contribution Checklist

- [ ] Follow Clean Architecture principles
- [ ] Write tests (90%+ coverage for domain/use case layers)
- [ ] Use structured logging with context
- [ ] Support environment-only deployment
- [ ] Implement graceful shutdown
- [ ] Format code with `make fmt`
- [ ] Pass linter checks with `make lint`

## License

MIT

## Support

For questions or issues:
- Open an issue on GitHub
- Check [CLAUDE.md](CLAUDE.md) for common patterns
- Review [CODING_STANDARDS.md](CODING_STANDARDS.md) for detailed guidelines
