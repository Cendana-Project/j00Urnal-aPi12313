.PHONY: run test build clean seed migrate seed-flush db-reset

# Default Go command
GO=go

# Application name
APP_NAME=soccernearu

DB_URL ?= postgres://postgres:password@localhost:5432/soccernearu?sslmode=disable

# Run the application
run:
	$(GO) run main.go server

# Build the application
build:
	$(GO) build -o $(APP_NAME) main.go

# Clean build artifacts
clean:
	rm -f $(APP_NAME)
