# Go Clean Architecture

A production-ready Go application implementing Clean Architecture principles with support for multiple runtime modes (API, Worker, Cron).

## Architecture Overview

This project follows Clean Architecture and Standard Go Project Layout principles:

```
┌─────────────────────────────────────────┐
│          Delivery Layer                 │
│   (HTTP, Workers, Cron Jobs)            │
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
│       └── job/            # Cron jobs
│           └── daily_report_job.go
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

Copy the example config and modify as needed:

```bash
cp configs/config.yaml configs/config.local.yaml
# Edit configs/config.local.yaml with your settings
```

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
docker run -p 8080:8080 go-clean-arch:latest --mode=api

# Run Worker mode
docker run go-clean-arch:latest --mode=worker

# Run Cron mode
docker run go-clean-arch:latest --mode=cron
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
