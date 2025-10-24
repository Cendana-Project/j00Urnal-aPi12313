.PHONY: run test build clean seed migrate seed-flush db-reset

# Default Go command
GO=go

# Application name
APP_NAME=medikaone

DB_URL ?= postgres://medikaone:dayak1352@localhost:5432/medikaone?sslmode=disable

# Run the application
run:
	$(GO) run main.go server

# Build the application
build:
	$(GO) build -o $(APP_NAME) main.go

# Clean build artifacts
clean:
	rm -f $(APP_NAME)

# Seed the database
seed:
	$(GO) run main.go seed

# Run database migration
migrate-up:
	$(GO) run main.go migrate --action up

# Create new migration
# make migrate-create name=nama_migrasi
migrate-create:
	$(GO) run main.go migrate --action create --name $(name)

# Reset database (down and up)
migrate-reset:
	$(GO) run main.go migrate --action reset

# Show migration status
migrate-status:
	$(GO) run main.go migrate --action status

# Database reset (PostgreSQL)
db-reset:
	@echo "⚠️  Resetting database schema..."
	@echo "Using DB_URL: $(DB_URL)"
	@psql "$(DB_URL)" -f scripts/reset-db.sql
	@echo "✅ Database schema reset complete"

# Seed flush
seed-flush:
	$(GO) run main.go seed-flush