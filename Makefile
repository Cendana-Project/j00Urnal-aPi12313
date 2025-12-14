.PHONY: run test build clean migrate-up migrate-down migrate-status migrate-create help

# Default Go command
GO=go

# Application name
APP_NAME=journal-api

# Help command
help:
	@echo "Available commands:"
	@echo "  make run              - Run the application server"
	@echo "  make build            - Build the application binary"
	@echo "  make test             - Run tests"
	@echo "  make test-coverage    - Run tests with coverage"
	@echo "  make migrate-up       - Run all pending migrations"
	@echo "  make migrate-down     - Rollback last migration"
	@echo "  make migrate-status   - Show migration status"
	@echo "  make migrate-create   - Create a new migration (use NAME=migration_name)"
	@echo "  make clean            - Remove build artifacts"
	@echo "  make lint             - Run linter"
	@echo "  make fmt              - Format code"

# Run the application
run:
	@echo "Starting $(APP_NAME) server..."
	$(GO) run main.go server

# Build the application
build:
	@echo "Building $(APP_NAME)..."
	$(GO) build -o $(APP_NAME) main.go
	@echo "Build complete: ./$(APP_NAME)"

# Run tests
test:
	@echo "Running tests..."
	$(GO) test -v ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	$(GO) test -cover -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run migrations up
migrate-up:
	@echo "Running migrations..."
	$(GO) run main.go migrate up

# Rollback last migration
migrate-down:
	@echo "Rolling back last migration..."
	$(GO) run main.go migrate down

# Show migration status
migrate-status:
	@echo "Migration status:"
	$(GO) run main.go migrate status

# Create new migration
migrate-create:
ifndef NAME
	@echo "Error: NAME is required. Usage: make migrate-create NAME=migration_name"
	@exit 1
endif
	@echo "Creating migration: $(NAME)"
	$(GO) run main.go migrate create $(NAME)

# Format code
fmt:
	@echo "Formatting code..."
	$(GO) fmt ./...

# Run linter
lint:
	@echo "Running linter..."
	@which golangci-lint > /dev/null || (echo "golangci-lint not installed. Install: https://golangci-lint.run/usage/install/" && exit 1)
	golangci-lint run ./...

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -f $(APP_NAME)
	rm -f coverage.out coverage.html
	@echo "Clean complete"

# Development setup helper
dev-setup:
	@echo "Setting up development environment..."
	@if [ ! -f .env ]; then \
		cp env.example .env; \
		echo ".env file created from env.example"; \
		echo "Please update .env with your configuration"; \
	else \
		echo ".env file already exists"; \
	fi
	@echo "Development setup complete"
