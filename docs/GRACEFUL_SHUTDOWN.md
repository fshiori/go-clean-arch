# Graceful Shutdown Implementation

This document describes the graceful shutdown implementation for the go-clean-arch application, following the official Gin framework recommendations.

## Overview

All three application modes (API, Worker, Cron) now support graceful shutdown, ensuring:
- In-flight requests/jobs complete before shutdown
- Database connections are properly closed
- No data loss during deployment or restart
- Clean exit with proper signal handling

## Implementation Details

### References

This implementation follows the official Gin documentation:
- **Gin Official Docs**: https://gin-gonic.com/docs/examples/graceful-restart-or-stop/
- **Gin Examples**: https://github.com/gin-gonic/examples/tree/master/graceful-shutdown

### Key Features

1. **Signal Handling**: Uses `signal.NotifyContext` (Go 1.16+) to listen for:
   - `SIGTERM` - Standard termination signal (Docker, Kubernetes)
   - `SIGINT` - Interrupt signal (Ctrl+C)
   - `os.Interrupt` - Platform-independent interrupt

2. **Graceful Shutdown Timeout**: 30 seconds for all modes
   - API: Time for in-flight HTTP requests to complete
   - Worker: Time for current message processing to complete
   - Cron: Time for running cron jobs to complete

3. **Resource Cleanup**:
   - HTTP server stops accepting new connections
   - Database connections are closed
   - Background jobs complete or timeout

## API Server (internal/app/api.go)

### Implementation Pattern

```go
// Create HTTP server with Gin router
srv := &http.Server{
    Addr:    addr,
    Handler: router,
}

// Start server in goroutine (non-blocking)
go func() {
    srv.ListenAndServe()
}()

// Listen for shutdown signals
ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
defer stop()

// Wait for signal
<-ctx.Done()

// Gracefully shutdown with timeout
shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()
srv.Shutdown(shutdownCtx)
```

### Behavior

**Normal Operation:**
```
$ ./app api
[INFO] Starting API server...
[INFO] API server listening address=:8080
```

**Graceful Shutdown (Ctrl+C or SIGTERM):**
```
^C
[INFO] Shutdown signal received, starting graceful shutdown...
[INFO] Finishing 3 in-flight requests...
[INFO] API server stopped gracefully
```

## Worker (internal/app/worker.go)

### Implementation Pattern

```go
// Context for shutdown signals
ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
defer stop()

// Start consumer with context
go func() {
    orderConsumer.Start(ctx)
}()

// Wait for signal
<-ctx.Done()

// Consumer stops when context is cancelled
// Wait for current message to finish processing
```

### Behavior

**Normal Operation:**
```
$ ./app worker
[INFO] Starting worker...
[INFO] Worker is ready to consume messages
[INFO] Order consumer is running...
```

**Graceful Shutdown:**
```
^C
[INFO] Shutdown signal received, stopping worker...
[INFO] Consumer shutdown signal received
[INFO] Worker stopped gracefully
```

### RabbitMQ Integration Notes

When connecting to a real message queue, the consumer should:

1. **Disable auto-ack**: Set `auto-ack` to `false` for message safety
2. **Use select with context**: Check for context cancellation in message loop
3. **Complete current message**: Finish processing before shutdown
4. **Acknowledge or Nack**: Properly acknowledge messages before exit

Example:
```go
for {
    select {
    case msg := <-msgs:
        // Process message
        if err := c.ConsumeMessage(msg.Body); err != nil {
            msg.Nack(false, true) // Requeue on error
        } else {
            msg.Ack(false) // Acknowledge success
        }
    case <-ctx.Done():
        // Stop consuming, current message is finished
        return nil
    }
}
```

## Cron Scheduler (internal/app/cron.go)

### Implementation Pattern

```go
// Start scheduler
scheduler.Start()

// Listen for shutdown signals
ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
defer stop()

// Wait for signal
<-ctx.Done()

// Stop scheduler and wait for running jobs
jobsCtx := scheduler.Stop() // Returns context

// Wait for jobs with timeout
select {
case <-jobsCtx.Done():
    // All jobs completed
case <-time.After(30 * time.Second):
    // Timeout, force shutdown
}
```

### Behavior

**Normal Operation:**
```
$ ./app cron
[INFO] Starting cron scheduler...
[INFO] Cron scheduler started successfully jobs_count=1
[INFO] Cron scheduler is ready
```

**Graceful Shutdown:**
```
^C
[INFO] Shutdown signal received, stopping cron scheduler...
[INFO] Stopping cron scheduler...
[INFO] Cron scheduler stopped, waiting for jobs to complete...
[INFO] All cron jobs completed successfully
[INFO] Cron scheduler stopped gracefully
```

### Cron Library Behavior

The `robfig/cron/v3` library provides:
- `cron.Stop()`: Returns a context that completes when all jobs finish
- Jobs already running will complete
- New jobs will not start after Stop() is called
- Safe for concurrent use

## Testing

### Manual Testing

Test graceful shutdown with the provided script:

```bash
# Test API server
./scripts/test-graceful-shutdown.sh api

# Test Worker
./scripts/test-graceful-shutdown.sh worker

# Test Cron
./scripts/test-graceful-shutdown.sh cron
```

### Expected Output

```
=== Graceful Shutdown Test Script ===

Testing graceful shutdown for: api mode

Starting api server...
Process started with PID: 12345
Waiting 3 seconds for server to initialize...
✅ Server started successfully

Sending SIGTERM signal to process 12345...
Waiting for graceful shutdown (max 35 seconds)...
.....
✅ Process shutdown gracefully in 5 seconds

✅ Graceful shutdown test PASSED

Expected behavior observed:
  - Server received SIGTERM signal
  - In-flight requests/jobs completed
  - Database connections closed properly
  - Clean shutdown with exit code 0
```

### Docker Testing

Test with Docker:

```bash
# Start container
docker run --name test-graceful go-clean-arch:latest

# In another terminal, send SIGTERM
docker stop test-graceful

# Check logs
docker logs test-graceful
```

Expected logs should show graceful shutdown messages.

### Kubernetes Testing

In Kubernetes, `SIGTERM` is sent during pod termination:

```yaml
apiVersion: v1
kind: Pod
spec:
  terminationGracePeriodSeconds: 30  # Matches our shutdown timeout
  containers:
  - name: api
    image: go-clean-arch:latest
```

## Production Considerations

### Timeout Configuration

The 30-second timeout is appropriate for most use cases:
- **Web requests**: Typically complete in < 5 seconds
- **Background jobs**: Most complete in < 30 seconds
- **Cron jobs**: Long-running jobs should be designed for graceful interruption

For longer operations:
- Increase timeout in code
- Adjust Kubernetes `terminationGracePeriodSeconds`
- Design jobs to be idempotent and resumable

### Load Balancer Configuration

For zero-downtime deployments:

1. **Deregister from load balancer**: Remove instance before sending SIGTERM
2. **Wait for connection drain**: Allow existing connections to complete
3. **Send SIGTERM**: Application performs graceful shutdown
4. **Health checks fail**: Mark instance as unhealthy immediately on shutdown

Example with AWS ALB:
```yaml
deregistration_delay: 30  # Seconds to wait before deregistering
health_check_interval: 5   # Check health every 5 seconds
```

### Database Connection Pooling

Current implementation closes database on shutdown:
```go
if err := s.db.Close(); err != nil {
    logger.Error("Error closing database connection", "error", err)
}
```

This ensures:
- No connection leaks
- Proper cleanup of prepared statements
- Database server resources are freed

### Monitoring

Monitor graceful shutdown with:

**Metrics:**
- `shutdown_duration_seconds` - Time taken to shutdown
- `shutdown_signals_total` - Count of shutdown signals received
- `active_connections_on_shutdown` - Connections active when shutdown started

**Logs:**
- Search for "Shutdown signal received" in logs
- Monitor for "forced to shutdown" errors
- Track shutdown duration in production

**Alerts:**
- Alert if shutdown takes > 25 seconds (approaching timeout)
- Alert if forced shutdowns occur (indicates timeout issues)
- Alert if shutdown errors occur

## Troubleshooting

### Shutdown Takes Too Long

**Symptoms**: Process takes full 30 seconds to shutdown

**Possible Causes**:
1. Long-running HTTP requests
2. Slow database queries
3. Stuck background jobs

**Solutions**:
- Add request timeouts
- Implement query timeouts
- Design jobs for graceful interruption
- Increase shutdown timeout if legitimate

### Forced Shutdown

**Symptoms**: "Server forced to shutdown" error in logs

**Possible Causes**:
1. Active connections not closing
2. Database connections stuck
3. External API calls hanging

**Solutions**:
- Add context timeouts to all operations
- Use `context.WithTimeout` for external calls
- Implement connection keep-alive properly
- Check for goroutine leaks

### Database Connection Errors

**Symptoms**: "Error closing database connection" in logs

**Possible Causes**:
1. Connections already closed
2. Active transactions during shutdown
3. Database server unavailable

**Solutions**:
- Ensure transactions are committed/rolled back
- Add proper error handling for db.Close()
- Use connection pooling correctly
- Check database health during shutdown

## References

- [Gin Graceful Shutdown Documentation](https://gin-gonic.com/docs/examples/graceful-restart-or-stop/)
- [Gin Examples Repository](https://github.com/gin-gonic/examples/tree/master/graceful-shutdown)
- [Go HTTP Server Shutdown](https://pkg.go.dev/net/http#Server.Shutdown)
- [Twelve-Factor App - Disposability](https://12factor.net/disposability)
- [robfig/cron Documentation](https://pkg.go.dev/github.com/robfig/cron/v3)

## Changelog

### 2025-11-20 - Initial Implementation

- ✅ Implemented graceful shutdown for API server
- ✅ Implemented graceful shutdown for Worker
- ✅ Implemented graceful shutdown for Cron scheduler
- ✅ Added signal handling (SIGTERM, SIGINT)
- ✅ Added 30-second shutdown timeout
- ✅ Added database connection cleanup
- ✅ Created test script
- ✅ Created documentation

### Compliance

This implementation satisfies:
- **Twelve-Factor App - Factor IX (Disposability)**: ✅ Maximized robustness with fast startup and graceful shutdown
- **Production Best Practices**: ✅ Zero-downtime deployments
- **Kubernetes Compatibility**: ✅ Proper SIGTERM handling
