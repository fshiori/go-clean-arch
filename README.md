# Go Clean Architecture

A production-ready Go application implementing Clean Architecture principles with support for multiple runtime modes (API, Worker, Cron).

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
│          Delivery Layer                 │
│   (HTTP, Workers, Cron, Microservice)   │
└──────────────┬──────────────────────────┘
               │
┌──────────────▼──────────────────────────┐
│       Interface/Adapter Layer           │
│   (Repositories, Gateways)              │
└──────────────┬──────────────────────────┘
               │
┌──────────────▼──────────────────────────┐
│          Use Case Layer                 │
│    (Application Business Logic)         │
└──────────────┬──────────────────────────┘
               │
┌──────────────▼──────────────────────────┐
│          Domain Layer                   │
│    (Entities, Business Rules)           │
└─────────────────────────────────────────┘
```

### Key Principles

- **Dependency Inversion**: Inner layers don't depend on outer layers
- **Domain-Driven Design**: Core business logic in the domain layer
- **Interface Segregation**: Interfaces defined in usecase/port
- **Single Responsibility**: Each layer has a clear purpose
- **Testability**: Easy to mock and test each layer

## Project Structure

```
.
├── cmd/                    # Application entry points
│   └── app/
│       └── main.go         # Main application with mode switching
├── internal/               # Private application code
│   ├── app/                # Application bootstrappers
│   │   ├── api.go          # API server setup
│   │   ├── worker.go       # Worker setup
│   │   └── cron.go         # Cron scheduler setup
│   ├── domain/             # Domain entities and business rules
│   │   ├── user.go
│   │   └── order.go
│   ├── usecase/            # Application business logic
│   │   ├── port/           # Port/Interface definitions
│   │   │   ├── user_repository.go
│   │   │   ├── order_repository.go
│   │   │   └── payment_gateway.go
│   │   ├── user_interactor.go
│   │   └── order_interactor.go
│   ├── adapter/            # Adapter/Implementation layer
│   │   ├── repository/     # Data access implementations
│   │   │   ├── user_repository_gorm.go
│   │   │   ├── order_repository_gorm.go
│   │   │   └── db.go
│   │   └── gateway/        # External API clients
│   │       └── stripe_gateway.go
│   └── delivery/           # Delivery mechanisms
│       ├── http/           # HTTP handlers
│       │   ├── handler/
│       │   │   ├── user_handler.go
│       │   │   └── user_dto.go
│       │   └── router.go
│       ├── consumer/       # Message queue consumers
│       │   └── order_consumer.go
│       ├── job/            # Cron jobs
│       │   └── daily_report_job.go
│       └── micro/          # Microservice RPC handlers
│           └── handler/
│               └── user_service_simple.go
├── pkg/                    # Public shared libraries
│   ├── config/
│   │   └── config.go
│   └── logger/
│       └── logger.go
├── configs/                # Configuration files
│   └── config.yaml
├── scripts/                # Build and deployment scripts
├── docs/                   # Documentation
├── api/                    # API definitions (OpenAPI/Swagger)
├── go.mod
├── go.sum
├── Makefile
├── Dockerfile
└── README.md
```

## Running the Application

### Prerequisites

- Go 1.24 or higher
- MySQL (PostgreSQL support coming soon)
- RabbitMQ (for worker mode, optional)

### Installation

```bash
# Clone the repository
git clone <repository-url>
cd go-clean-arch

# Install dependencies
go mod download
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

See [docs/TWELVE_FACTOR_COMPLIANCE.md](docs/TWELVE_FACTOR_COMPLIANCE.md) for details.

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
go run cmd/app/main.go --mode=api
```

The API server will start on `http://localhost:8080`

#### Worker Mode

```bash
go run cmd/app/main.go --mode=worker
```

The worker will start consuming messages from the message queue.

#### Cron Scheduler Mode

```bash
go run cmd/app/main.go --mode=cron
```

The cron scheduler will start and run scheduled jobs.

#### Microservice Mode

```bash
# Start with default settings (service name: go.micro.service.user, address: :8081)
./app micro

# Start with custom settings
./app micro --service-name user.service --version v1.0.0 --address :8082

# With config file
./app micro --config configs/config.toml
```

The microservice will start a JSON-RPC style server for inter-service communication.

**Implementation Note**: The current implementation uses a simple HTTP-based RPC approach with handler signatures compatible with go-micro framework. This demonstrates the microservice delivery layer pattern while remaining framework-agnostic. For production deployments, you can easily integrate with [go-micro](https://go-micro.dev) or other RPC frameworks.

**Available RPC Endpoints:**

- `POST /rpc/UserServiceSimple.CreateUser` - Create a new user
- `POST /rpc/UserServiceSimple.GetUser` - Get user by ID
- `POST /rpc/UserServiceSimple.ListUsers` - List users (with pagination)
- `POST /rpc/UserServiceSimple.UpdatePassword` - Update user password
- `POST /rpc/UserServiceSimple.DeleteUser` - Delete user
- `GET /info` - Service information

**Example RPC call:**

```bash
curl -X POST http://localhost:8081/rpc/UserServiceSimple.CreateUser \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"secret123"}'
```

**Integrating with go-micro**:

To use with the actual go-micro framework, add the dependency:

```bash
go get go-micro.dev/v5@latest
```

The handlers in `internal/delivery/micro/handler/` follow go-micro's signature pattern `func(ctx, req, rsp) error` and can be registered directly with go-micro's server.

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
docker run -p 8080:8080 go-clean-arch:latest api

# Run Worker mode
docker run go-clean-arch:latest worker

# Run Cron mode
docker run go-clean-arch:latest cron

# Run Microservice mode
docker run -p 8081:8081 go-clean-arch:latest micro --address :8081
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

## Development

### Adding a New Feature

1. **Define Domain Entity** in `internal/domain/`
2. **Define Repository Interface** in `internal/usecase/port/`
3. **Implement Repository** in `internal/interface/repository/`
4. **Create Use Case** in `internal/usecase/`
5. **Add HTTP Handler** in `internal/delivery/http/handler/`
6. **Register Routes** in `internal/delivery/http/router.go`
7. **Wire Dependencies** in `internal/app/api.go`

### Testing

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests for a specific package
go test ./internal/usecase/...
```

### Code Quality

```bash
# Format code
go fmt ./...

# Run linter
golangci-lint run

# Run tests
go test ./...
```

## Design Decisions

### Why Interfaces in usecase/port?

Following the discussion in the Gemini conversation, interfaces are placed in the `usecase/port` directory because:
- Use cases define what capabilities they need
- This is more pragmatic for most applications
- It keeps the domain layer clean and focused on business rules
- It's easier to understand and maintain

### Why Single Binary, Multiple Modes?

- **Code Reuse**: All modes share the same business logic
- **Simplified Dependencies**: Single `go.mod` file
- **Atomic Changes**: Changes to shared logic deploy together
- **Easier Development**: One repository to clone and work with

### Data Transformation Layers

1. **Domain Entity** ↔ **DB Model**: Transformed in Repository layer
2. **Domain Entity** ↔ **DTO**: Transformed in HTTP Handler layer
3. **Domain remains pure**: No database or JSON tags

## License

MIT

## Contributing

Please read CONTRIBUTING.md for details on our code of conduct and the process for submitting pull requests.
