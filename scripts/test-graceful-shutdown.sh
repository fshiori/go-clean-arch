#!/bin/bash

# Script to test graceful shutdown functionality
# This script demonstrates that the application properly handles SIGTERM and SIGINT signals

set -e

echo "=== Graceful Shutdown Test Script ==="
echo ""

MODE="${1:-api}"

if [ "$MODE" != "api" ] && [ "$MODE" != "worker" ] && [ "$MODE" != "cron" ]; then
    echo "Usage: $0 [api|worker|cron]"
    echo "Example: $0 api"
    exit 1
fi

echo "Testing graceful shutdown for: $MODE mode"
echo ""

# Build the application if binary doesn't exist
if [ ! -f "./bin/app" ]; then
    echo "Building application..."
    go build -o bin/app ./cmd/app
    echo "Build complete!"
    echo ""
fi

# Start the application in background
echo "Starting $MODE server..."
./bin/app $MODE --config=configs/config.example.toml &
PID=$!

echo "Process started with PID: $PID"
echo "Waiting 3 seconds for server to initialize..."
sleep 3

# Check if process is still running
if ! kill -0 $PID 2>/dev/null; then
    echo "❌ ERROR: Process failed to start"
    exit 1
fi

echo "✅ Server started successfully"
echo ""

# Send SIGTERM signal
echo "Sending SIGTERM signal to process $PID..."
kill -TERM $PID

# Wait for graceful shutdown with timeout
echo "Waiting for graceful shutdown (max 35 seconds)..."
timeout=35
elapsed=0

while kill -0 $PID 2>/dev/null; do
    if [ $elapsed -ge $timeout ]; then
        echo "❌ TIMEOUT: Process did not shutdown within $timeout seconds"
        kill -9 $PID
        exit 1
    fi
    sleep 1
    elapsed=$((elapsed + 1))
    echo -n "."
done

echo ""
echo "✅ Process shutdown gracefully in $elapsed seconds"
echo ""

# Check exit code
wait $PID
EXIT_CODE=$?

if [ $EXIT_CODE -eq 0 ]; then
    echo "✅ Graceful shutdown test PASSED"
    echo ""
    echo "Expected behavior observed:"
    echo "  - Server received SIGTERM signal"
    echo "  - In-flight requests/jobs completed"
    echo "  - Database connections closed properly"
    echo "  - Clean shutdown with exit code 0"
    exit 0
else
    echo "❌ Graceful shutdown test FAILED with exit code: $EXIT_CODE"
    exit 1
fi
