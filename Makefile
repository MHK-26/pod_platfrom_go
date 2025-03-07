# Makefile
.PHONY: run build test clean migrate-up migrate-down help swag deps fmt lint sync-rss

# Service names
SERVICES := auth-service content-service analytics-service recommendation-service

# Database configuration
DB_USER := postgres
DB_PASSWORD := postgres
DB_NAME := podcast_platform
DB_HOST := localhost
DB_PORT := 5432
DB_URL := postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable

# Run all services
run:
	@echo "Running all services..."
	@for service in $(SERVICES); do \
		go run ./cmd/$$service/main.go & \
	done

# Run a specific service
run-%:
	@echo "Running $*..."
	go run ./cmd/$*/main.go

# Build all services
build:
	@echo "Building all services..."
	@for service in $(SERVICES); do \
		go build -o bin/$$service ./cmd/$$service/main.go; \
	done

# Build a specific service
build-%:
	@echo "Building $*..."
	go build -o bin/$* ./cmd/$*/main.go

# Run tests
test:
	go test -v ./...

# Clean build artifacts
clean:
	rm -rf bin/
	rm -rf vendor/

# Apply database migrations
migrate-up:
	migrate -path ./scripts/migrations -database "$(DB_URL)" -verbose up

# Rollback database migrations
migrate-down:
	migrate -path ./scripts/migrations -database "$(DB_URL)" -verbose down

# Generate API documentation
swag:
	swag init -g cmd/auth-service/main.go -o api/swagger

# Install dependencies
deps:
	go mod download

# Format code
fmt:
	go fmt ./...

# Lint code
lint:
	golangci-lint run ./...

# Sync RSS feeds for all podcasts
sync-rss:
	go run ./cmd/content-service/main.go -sync-rss

# Help
help:
	@echo "Available targets:"
	@echo "  run                - Run all services"
	@echo "  run-SERVICE        - Run a specific service (e.g., run-auth-service)"
	@echo "  build              - Build all services"
	@echo "  build-SERVICE      - Build a specific service (e.g., build-auth-service)"
	@echo "  test               - Run tests"
	@echo "  clean              - Clean build artifacts"
	@echo "  migrate-up         - Apply database migrations"
	@echo "  migrate-down       - Rollback database migrations"
	@echo "  swag               - Generate API documentation"
	@echo "  deps               - Install dependencies"
	@echo "  fmt                - Format code"
	@echo "  lint               - Lint code"
	@echo "  sync-rss           - Manually trigger RSS feed synchronization"