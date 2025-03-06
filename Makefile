# Makefile
.PHONY: run build test clean docker-build docker-run migrateup migratedown

# Service names
SERVICES := auth-service content-service analytics-service recommendation-service payment-service

# Docker configuration
DOCKER_REPO := your-docker-repo
VERSION := latest

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

# Build Docker images for all services
docker-build:
	@echo "Building Docker images for all services..."
	@for service in $(SERVICES); do \
		docker build -t $(DOCKER_REPO)/$$service:$(VERSION) -f deployments/docker/$$service/Dockerfile .; \
	done

# Build Docker image for a specific service
docker-build-%:
	@echo "Building Docker image for $*..."
	docker build -t $(DOCKER_REPO)/$*:$(VERSION) -f deployments/docker/$*/Dockerfile .

# Run Docker containers for all services
docker-run:
	@echo "Running Docker containers for all services..."
	@for service in $(SERVICES); do \
		docker run -d -p 8080:8080 --name $$service $(DOCKER_REPO)/$$service:$(VERSION); \
	done

# Run Docker container for a specific service
docker-run-%:
	@echo "Running Docker container for $*..."
	docker run -d -p 8080:8080 --name $* $(DOCKER_REPO)/$*:$(VERSION)

# Apply database migrations
migrateup:
	migrate -path ./scripts/migrations -database "$(DB_URL)" -verbose up

# Rollback database migrations
migratedown:
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

# Help
help:
	@echo "Available targets:"
	@echo "  run                - Run all services"
	@echo "  run-SERVICE        - Run a specific service (e.g., run-auth-service)"
	@echo "  build              - Build all services"
	@echo "  build-SERVICE      - Build a specific service (e.g., build-auth-service)"
	@echo "  test               - Run tests"
	@echo "  clean              - Clean build artifacts"
	@echo "  docker-build       - Build Docker images for all services"
	@echo "  docker-build-SERVICE - Build Docker image for a specific service"
	@echo "  docker-run         - Run Docker containers for all services"
	@echo "  docker-run-SERVICE - Run Docker container for a specific service"
	@echo "  migrateup          - Apply database migrations"
	@echo "  migratedown        - Rollback database migrations"
	@echo "  swag               - Generate API documentation"
	@echo "  deps               - Install dependencies"
	@echo "  fmt                - Format code"
	@echo "  lint               - Lint code"