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

.PHONY: all build clean run test lint fmt vet docker-build docker-push setup help

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

help: ## Show this help message
	@echo "Usage: make [target]"
	@echo
	@echo "Targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'