# Global variables
APP_NAME := alya
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME := $(shell date -u '+%Y-%m-%d %H:%M:%S')
COMMIT_HASH := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
LDFLAGS := -ldflags "-X main.version=$(VERSION) -X main.buildTime=$(BUILD_TIME) -X main.commitHash=$(COMMIT_HASH)"

# Go related variables
GOBASE := $(shell pwd)
GOPATH := $(shell go env GOPATH)
GOBIN := $(GOBASE)/bin
GOFILES := $(shell find . -type f -name "*.go" -not -path "./vendor/*")

# Docker related variables
DOCKER_REGISTRY ?= localhost:5000
DOCKER_IMAGE := $(DOCKER_REGISTRY)/$(APP_NAME)
DOCKER_TAG ?= latest

# Database related variables
DB_HOST ?= localhost
DB_PORT ?= 5433
DB_USER ?= postgres
DB_PASSWORD ?= postgres
DB_NAME ?= alya
DB_SSLMODE ?= disable
MIGRATIONS_DIR := migrations

# Migration tool
MIGRATE := $(GOBIN)/migrate
MIGRATE_VERSION := v4.15.2

.PHONY: all build clean run test lint fmt vet docker-build docker-push setup help
.PHONY: db-setup db-create-migration db-migrate db-rollback db-drop db-reset db-version db-force

all: setup lint test build ## Default: setup, lint, test, and build

build: ## Build the application
	@echo "Building $(APP_NAME)..."
	@go build $(LDFLAGS) -o $(GOBIN)/$(APP_NAME) ./main.go

clean: ## Clean the project
	@echo "Cleaning..."
	@rm -rf $(GOBIN)
	@go clean -cache -testcache -modcache

run: ## Run the application
	@echo "Running $(APP_NAME)..."
	@go run $(LDFLAGS) ./main.go

test: ## Run tests
	@echo "Running tests..."
	@go test -v ./...

lint: fmt vet ## Run all linters

fmt: ## Run gofmt
	@echo "Running gofmt..."
	@gofmt -s -w $(GOFILES)

vet: ## Run go vet
	@echo "Running go vet..."
	@go vet ./...

docker-build: ## Build Docker image
	@echo "Building Docker image..."
	@docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .

docker-push: ## Push Docker image
	@echo "Pushing Docker image..."
	@docker push $(DOCKER_IMAGE):$(DOCKER_TAG)

docker-run: ## Run Docker container
	@echo "Running Docker container..."
	@docker run -p 8080:8080 $(DOCKER_IMAGE):$(DOCKER_TAG)

docker-compose-up: ## Start services with Docker Compose
	@echo "Starting services with Docker Compose..."
	@docker-compose up -d

docker-compose-down: ## Stop services with Docker Compose
	@echo "Stopping services with Docker Compose..."
	@docker-compose down

setup: ## Set up the development environment
	@echo "Setting up development environment..."
	@mkdir -p $(GOBIN)
	@go mod tidy

# Database migration commands
# Add these rules to your existing Makefile

# Database migration commands
db-setup: ## Install migrate CLI tool for database migrations
	@echo "Installing migrate CLI tool..."
	@if [ ! -f $(MIGRATE) ]; then \
		echo "Downloading migration tool..."; \
		GOBIN=$(GOBIN) go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@$(MIGRATE_VERSION); \
	else \
		echo "Migration tool already installed."; \
	fi
	@mkdir -p $(MIGRATIONS_DIR)

db-create-migration: db-setup ## Create a new migration (usage: make db-create-migration name=create_users_table)
	@if [ -z "$(name)" ]; then \
		echo "Error: Migration name is required. Usage: make db-create-migration name=create_users_table"; \
		exit 1; \
	fi
	@echo "Creating migration $(name)..."
	@$(MIGRATE) create -ext sql -dir $(MIGRATIONS_DIR) -seq $(name)
	@echo "Migration files created in $(MIGRATIONS_DIR) directory."

db-migrate: db-setup ## Run all pending migrations
	@echo "Running migrations..."
	@$(MIGRATE) -path $(MIGRATIONS_DIR) -database postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSLMODE) up
	@echo "Migrations applied successfully."

db-migrate-container: ## Run migrations directly in the database container
	@echo "Running migrations from inside the container..."
	@docker cp $(MIGRATIONS_DIR) alya-db:/tmp/
	@docker exec alya-db sh -c "apt-get update && apt-get install -y wget && \
		wget -q https://github.com/golang-migrate/migrate/releases/download/v4.15.2/migrate.linux-amd64.tar.gz && \
		tar -xzf migrate.linux-amd64.tar.gz && \
		./migrate -path=/tmp/migrations -database postgres://postgres:postgres@localhost:5432/alya?sslmode=disable up"
	@echo "Migrations applied successfully."

db-migrate-psql: ## Run migrations using psql directly in the container
	@echo "Running migrations using psql directly..."
	@docker cp $(MIGRATIONS_DIR) alya-db:/tmp/
	@docker exec alya-db psql -U postgres -d alya -f /tmp/migrations/000001_create_videos_table.up.sql
	@echo "Migrations applied successfully."

db-migrate-docker: ## Run migrations inside Docker container
	@echo "Running migrations inside Docker container..."
	@docker run --rm --network backend_default \
		-v $(GOBASE)/$(MIGRATIONS_DIR):/migrations \
		migrate/migrate:v4 \
		-path=/migrations \
		-database postgres://$(DB_USER):$(DB_PASSWORD)@alya-db:5432/$(DB_NAME)?sslmode=$(DB_SSLMODE) \
		up -verbose
	@echo "Migrations applied successfully."

db-rollback: db-setup ## Rollback the last migration
	@echo "Rolling back the last migration..."
	@$(MIGRATE) -path $(MIGRATIONS_DIR) -database postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSLMODE) down 1
	@echo "Rollback completed successfully."

db-version: db-setup ## Check current migration version
	@echo "Current migration version:"
	@$(MIGRATE) -path $(MIGRATIONS_DIR) -database postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSLMODE) version

db-drop: db-setup ## Drop all tables in the database
	@echo "Dropping all tables from the database..."
	@$(MIGRATE) -path $(MIGRATIONS_DIR) -database postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSLMODE) drop -f
	@echo "All tables dropped."

db-reset: db-drop db-migrate ## Drop all tables and run all migrations

db-force: db-setup ## Force migration version (usage: make db-force version=000001)
	@if [ -z "$(version)" ]; then \
		echo "Error: Version number is required. Usage: make db-force version=000001"; \
		exit 1; \
	fi
	@echo "Forcing migration version to $(version)..."
	@$(MIGRATE) -path $(MIGRATIONS_DIR) -database postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSLMODE) force $(version)
	@echo "Migration version set to $(version)."

help: ## Show this help message
	@echo "Usage: make [target]"
	@echo
	@echo "Targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'