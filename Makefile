.PHONY: run build migrate-up migrate-down test clean help

# Variables
BINARY_NAME=jobapply-api
MAIN_PATH=./cmd/api/main.go

# Default target
help:
	@echo "Available commands:"
	@echo "  make run          - Run the application"
	@echo "  make build        - Build the application"
	@echo "  make migrate-up   - Run database migrations (up)"
	@echo "  make migrate-down - Rollback database migrations (down)"
	@echo "  make test         - Run tests"
	@echo "  make clean        - Remove build artifacts"
	@echo "  make help         - Show this help message"

# Run the application
run:
	@echo "Starting server..."
	go run $(MAIN_PATH)

# Build the application
build:
	@echo "Building application..."
	go build -o bin/$(BINARY_NAME) $(MAIN_PATH)
	@echo "Build complete: bin/$(BINARY_NAME)"

# Run database migrations (manual execution of SQL)
migrate-up:
	@echo "Running migrations..."
	@echo "Note: Migrations are automatically run when the server starts"
	@echo "To manually run migrations, execute the SQL file:"
	@echo "psql $$DATABASE_URL -f internal/database/migrations/001_initial_schema.up.sql"

# Rollback database migrations
migrate-down:
	@echo "Rolling back migrations..."
	@echo "To manually rollback migrations, execute the SQL file:"
	@echo "psql $$DATABASE_URL -f internal/database/migrations/001_initial_schema.down.sql"

# Run tests
test:
	@echo "Running tests..."
	go test -v ./...

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -rf bin/
	rm -rf uploads/*
	@echo "Clean complete"

# Development dependencies
deps:
	@echo "Installing dependencies..."
	go mod download
	go mod tidy
	@echo "Dependencies installed"
