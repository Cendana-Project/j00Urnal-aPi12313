.PHONY: run test build clean migrate-up migrate-down migrate-status migrate-create seed seed-flush help
.PHONY: docker-build docker-run docker-test docker-clean deploy-check gen-secrets
.PHONY: render-deploy render-logs render-shell dev-setup

# Default Go command
GO=go

# Application name
APP_NAME=journal-api

# Docker configuration
DOCKER_IMAGE=$(APP_NAME)
DOCKER_TAG=latest

# Help command
help:
	@echo "=========================================="
	@echo "Journal API - Available Commands"
	@echo "=========================================="
	@echo ""
	@echo "Development:"
	@echo "  make run              - Run the application server"
	@echo "  make build            - Build the application binary"
	@echo "  make test             - Run tests"
	@echo "  make test-coverage    - Run tests with coverage"
	@echo "  make dev-setup        - Setup development environment"
	@echo "  make fmt              - Format code"
	@echo "  make lint             - Run linter"
	@echo ""
	@echo "Database Migrations:"
	@echo "  make migrate-up       - Run all pending migrations"
	@echo "  make migrate-down     - Rollback last migration"
	@echo "  make migrate-status   - Show migration status"
	@echo "  make migrate-test     - Test migration idempotency and safety"
	@echo "  make migrate-create   - Create new migration (use NAME=migration_name)"
	@echo "  make seed             - Seed the database with initial data"
	@echo "  make seed-flush       - Flush seeded data from the database"
	@echo ""
	@echo "Docker:"
	@echo "  make docker-build     - Build Docker image"
	@echo "  make docker-run       - Run Docker container locally"
	@echo "  make docker-test      - Test Docker container"
	@echo "  make docker-clean     - Remove Docker images and containers"
	@echo ""
	@echo "Deployment:"
	@echo "  make deploy-check     - Validate deployment configuration"
	@echo "  make gen-secrets      - Generate secure secrets for deployment"
	@echo "  make render-deploy    - Manual trigger Render deployment"
	@echo "  make render-logs      - View Render logs (requires render-cli)"
	@echo ""
	@echo "Utility:"
	@echo "  make clean            - Remove build artifacts"
	@echo "=========================================="

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

# Run migrations up and seed
migrate-up:
	@echo "Running migrations..."
	$(GO) run main.go migrate up
	@echo "Seeding database..."
	$(GO) run main.go seed

# Seed the database
seed:
	@echo "Seeding database..."
	$(GO) run main.go seed

# Flush seeded data
seed-flush:
	@echo "Flushing seeded data..."
	$(GO) run main.go seed-flush

# Rollback last migration
migrate-down:
	@echo "Rolling back last migration..."
	$(GO) run main.go migrate down

# Show migration status
migrate-status:
	@echo "Migration status:"
	$(GO) run main.go migrate --action=status

# Test migrations (idempotency and safety)
migrate-test:
	@echo "Testing migrations..."
	@bash scripts/test-migrations.sh

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

# ========================================
# Docker Commands
# ========================================

# Build Docker image
docker-build:
	@echo "Building Docker image: $(DOCKER_IMAGE):$(DOCKER_TAG)"
	docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .
	@echo "Docker image built successfully"

# Run Docker container locally (requires .env file)
docker-run: docker-build
	@echo "Running Docker container..."
	@if [ ! -f .env ]; then \
		echo "Error: .env file not found. Run 'make dev-setup' first."; \
		exit 1; \
	fi
	docker run --rm \
		--env-file .env \
		-p 8080:8080 \
		--name $(APP_NAME) \
		$(DOCKER_IMAGE):$(DOCKER_TAG)

# Test Docker container build and health
docker-test: docker-build
	@echo "Testing Docker container..."
	@docker run --rm -d \
		-e ENV=development \
		-e LOG_LEVEL=info \
		-e DATABASE_DSN=postgres://test:test@localhost:5432/test \
		-e TOKEN_PASSWORD_SALT=test-salt-min-16-chars \
		-e TOKEN_ACCESS_TOKEN_SECRET=test-access-secret-16 \
		-e TOKEN_REFRESH_TOKEN_SECRET=test-refresh-secret-16 \
		-e REDIS_IS_CACHE_DISABLE=true \
		-p 8080:8080 \
		--name $(APP_NAME)-test \
		$(DOCKER_IMAGE):$(DOCKER_TAG) || true
	@sleep 5
	@echo "Checking container health..."
	@docker ps | grep $(APP_NAME)-test || (echo "Container not running" && exit 1)
	@echo "Cleaning up test container..."
	@docker stop $(APP_NAME)-test || true
	@docker rm $(APP_NAME)-test || true
	@echo "Docker test complete"

# Clean Docker resources
docker-clean:
	@echo "Cleaning Docker resources..."
	@docker stop $(APP_NAME) 2>/dev/null || true
	@docker rm $(APP_NAME) 2>/dev/null || true
	@docker rmi $(DOCKER_IMAGE):$(DOCKER_TAG) 2>/dev/null || true
	@echo "Docker cleanup complete"

# ========================================
# Deployment Commands
# ========================================

# Generate secure secrets for deployment
gen-secrets:
	@echo "=========================================="
	@echo "Generating Secure Secrets"
	@echo "=========================================="
	@echo ""
	@echo "Copy these values to your Render dashboard:"
	@echo ""
	@echo "TOKEN_PASSWORD_SALT="
	@openssl rand -base64 32
	@echo ""
	@echo "TOKEN_ACCESS_TOKEN_SECRET="
	@openssl rand -base64 32
	@echo ""
	@echo "TOKEN_REFRESH_TOKEN_SECRET="
	@openssl rand -base64 32
	@echo ""
	@echo "=========================================="
	@echo "⚠️  IMPORTANT: Save these securely!"
	@echo "=========================================="

# Validate deployment configuration
deploy-check:
	@echo "Running comprehensive deployment validation..."
	@bash scripts/validate-deployment.sh

# Manual Render deployment (requires render-cli)
render-deploy:
	@echo "Triggering Render deployment..."
	@which render > /dev/null || (echo "Error: render CLI not installed. Visit: https://render.com/docs/cli" && exit 1)
	@render deploy
	@echo "Deployment triggered. Check Render dashboard for status."

# View Render logs (requires render-cli)
render-logs:
	@which render > /dev/null || (echo "Error: render CLI not installed. Visit: https://render.com/docs/cli" && exit 1)
	@render logs

# ========================================
# Pre-deployment Tests
# ========================================

# Run all pre-deployment checks
pre-deploy: clean fmt lint test docker-test deploy-check
	@echo ""
	@echo "=========================================="
	@echo "✅ All pre-deployment checks passed!"
	@echo "=========================================="
	@echo ""
	@echo "You're ready to deploy! 🚀"
	@echo ""
