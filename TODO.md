# TODO List - Architecture Improvements

## 🔴 Critical (Must Fix Immediately)

### Security & Core Functionality

- [ ] **[CRITICAL] Implement proper password hashing** (`internal/domain/user.go:67-74`)
  - Replace `"hashed_" + password` with bcrypt or argon2
  - Update `ChangePassword` method to use proper hashing
  - Add password verification method
  - Reference: `golang.org/x/crypto/bcrypt`

- [ ] **[HIGH] Fix domain entity password field reconstruction** (`internal/adapter/repository/user_repository_gorm.go:102-115`)
  - Add `ReconstructUser()` factory method in domain package
  - Allow repositories to properly hydrate User entities with password hash
  - Maintain encapsulation while enabling proper persistence

- [ ] **[HIGH] Fix Wire dependency injection for API mode** (`internal/app/wire.go:66-75`)
  - Add `GatewaySet` to `InitializeAPIRouter`
  - Ensure PaymentGateway is available in API mode
  - Test order checkout functionality via REST API

- [ ] **[MEDIUM] Standardize configuration file format**
  - Choose one format (TOML recommended)
  - Update all references in code, docker-compose, and documentation
  - Current issue: `cmd/app/cmd/root.go:49` uses `.toml`, docker-compose uses `.yaml`

---

## 🟠 High Priority

### Testing & Quality

- [ ] **[HIGH] Add unit tests**
  - [ ] Domain layer tests (User, Order business logic)
  - [ ] Use case layer tests with mocked repositories
  - [ ] Repository integration tests
  - [ ] Handler tests with mocked use cases
  - Target: Minimum 70% code coverage

### Error Handling

- [ ] **[MEDIUM] Create custom domain error types**
  - Define error constants in `internal/domain/errors.go`
  - Replace `errors.New()` with typed errors
  - Examples: `ErrNotFound`, `ErrEmailAlreadyExists`, `ErrInvalidPassword`
  - Update handlers to map domain errors to proper HTTP status codes

### Missing Components

- [ ] **[MEDIUM] Implement OrderHandler** (`internal/delivery/http/`)
  - Create `order_handler.go` with CreateOrder, GetOrder, ListOrders methods
  - Define Order DTOs in `dto/order_dto.go`
  - Add routes in `router.go:48-53` (currently commented out)
  - Inject OrderUsecase via Wire

### Context & Lifecycle

- [ ] **[MEDIUM] Pass context.Context to all use case methods**
  - Update UserUsecase interface signatures
  - Update OrderUsecase interface signatures
  - Update all implementations
  - Enable timeout, cancellation, and request-scoped values

- [ ] **[MEDIUM] Implement graceful shutdown**
  - API mode: Handle SIGINT/SIGTERM, wait for in-flight requests
  - Worker mode: Finish processing current messages
  - Cron mode: Stop accepting new jobs
  - Close database connections cleanly
  - Reference: `internal/app/api.go:30-46`

### Health & Monitoring

- [ ] **[MEDIUM] Enhance health check endpoint** (`internal/delivery/http/router.go:28-32`)
  - Add database connectivity check
  - Add message queue status (when implemented)
  - Add external API gateway status
  - Return detailed health status (healthy/degraded/unhealthy)

---

## 🟡 Medium Priority

### Database & Persistence

- [ ] **[MEDIUM] Add database migration management**
  - Integrate golang-migrate or goose
  - Create initial migration files
  - Update Makefile migrate commands (`Makefile:76-82`)
  - Add migration version tracking

- [ ] **[LOW] Make database connection pool configurable**
  - Move hardcoded values to config (`internal/adapter/repository/db.go:68-70`)
  - Add `database.pool.max_idle_conns`, `max_open_conns`, `conn_max_lifetime` to config
  - Use config values in SetMaxIdleConns, SetMaxOpenConns, SetConnMaxLifetime

### External Integrations

- [ ] **[MEDIUM] Implement message queue integration** (`internal/delivery/consumer/order_consumer.go:118`)
  - Complete RabbitMQ/Kafka consumer implementation
  - Add connection management
  - Add error handling and retry logic
  - Add dead letter queue

- [ ] **[LOW] Implement report generation** (`internal/delivery/job/daily_report_job.go:58`)
  - Add actual report generation logic
  - Integrate with email service or file storage
  - Add report templates

### Configuration & Security

- [ ] **[MEDIUM] Improve sensitive data handling in config**
  - Remove credentials from example config files
  - Document environment variable requirements
  - Add config validation for required secrets
  - Consider using secret management (e.g., Vault, AWS Secrets Manager)

- [ ] **[LOW] Add config reload support**
  - Implement Viper watch functionality
  - Add signal handler for config reload
  - Update logger levels dynamically

---

## 🟢 Nice to Have

### API & Documentation

- [ ] **[LOW] Add validation error handler middleware**
  - Create centralized handler for Gin validation errors
  - Return consistent error response format
  - Include field-level error details

- [ ] **[LOW] Complete OpenAPI/Swagger documentation**
  - Add swagger annotations to handlers
  - Generate OpenAPI spec
  - Serve Swagger UI at `/swagger/*`

- [ ] **[LOW] Add API versioning strategy**
  - Document versioning approach (currently using `/v1`)
  - Plan for v2 endpoint structure
  - Add version deprecation handling

### Monitoring & Observability

- [ ] **[LOW] Add Prometheus metrics**
  - HTTP request duration histogram
  - Error rate counters by endpoint
  - Database query duration
  - External API call metrics

- [ ] **[LOW] Add distributed tracing**
  - Integrate OpenTelemetry
  - Add trace context propagation
  - Connect to Jaeger or Zipkin

### Security & Reliability

- [ ] **[LOW] Add rate limiting middleware**
  - Implement per-IP rate limiting
  - Add per-user rate limiting
  - Make limits configurable

- [ ] **[LOW] Implement CORS configuration**
  - Make CORS configurable via config file
  - Support multiple allowed origins
  - Configure allowed methods and headers

### Testing

- [ ] **[LOW] Add integration tests**
  - Test full request/response cycle
  - Use test database
  - Test worker and cron modes

- [ ] **[LOW] Add load/performance tests**
  - Use k6 or vegeta
  - Establish performance baselines
  - Add to CI/CD pipeline

---

## 📋 Technical Debt & Refactoring

- [ ] Review and remove all TODO comments in code
- [ ] Add godoc comments to all exported types and functions
- [ ] Consider adding linter configuration (golangci-lint)
- [ ] Review and optimize database indexes
- [ ] Add database query logging in development mode
- [ ] Consider adding request/response logging middleware
- [ ] Review and document deployment strategy
- [ ] Add database backup and restore procedures documentation

---

## 📊 Current Status

**Overall Architecture Grade:** B+ (Good, with specific areas needing attention)

**Blockers for Production:**
1. Password hashing security vulnerability
2. No test coverage
3. Missing graceful shutdown
4. No proper error handling

**Strengths:**
- ✅ Clean architecture with proper layer separation
- ✅ Wire dependency injection
- ✅ Structured logging with trace IDs
- ✅ Multiple runtime modes (API, Worker, Cron)
- ✅ Docker support
- ✅ Rich domain entities with business logic

---

## 📝 Notes

- This project demonstrates excellent architectural foundation
- Main gaps are in testing, security, and operational readiness
- With critical fixes applied, this will be production-ready
- Consider this a living document - update as you make progress

**Last Updated:** 2025-11-07
