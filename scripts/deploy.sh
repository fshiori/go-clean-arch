#!/bin/bash

# Deployment script for go-clean-arch
# This is a simple example; adjust for your specific deployment needs

set -e

APP_NAME="go-clean-arch"
VERSION=${1:-latest}

echo "========================================="
echo "Deploying $APP_NAME - Version: $VERSION"
echo "========================================="

# Step 1: Run tests
echo "Step 1: Running tests..."
go test ./...

# Step 2: Build the application
echo "Step 2: Building application..."
make build

# Step 3: Build Docker image
echo "Step 3: Building Docker image..."
docker build -t $APP_NAME:$VERSION .
docker tag $APP_NAME:$VERSION $APP_NAME:latest

# Step 4: Push to registry (if needed)
# Uncomment and modify for your registry
# echo "Step 4: Pushing to Docker registry..."
# docker tag $APP_NAME:$VERSION your-registry/$APP_NAME:$VERSION
# docker push your-registry/$APP_NAME:$VERSION

# Step 5: Deploy to environment
# This is where you'd add your deployment commands
# Examples:
# - kubectl apply -f k8s/
# - docker-compose up -d
# - scp binary to server

echo "========================================="
echo "Deployment completed successfully!"
echo "========================================="
