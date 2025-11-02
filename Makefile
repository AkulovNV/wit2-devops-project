.PHONY: help build test run docker-build docker-run clean lint fmt

VERSION ?= 1.0.0
BINARY := app
DOCKER_IMAGE := go-simple-api
DOCKER_REGISTRY := ghcr.io/devops-mentor
APP_DIR := go-simple-api

help:
	@echo "Go Simple API - Makefile targets:"
	@echo ""
	@echo "Development:"
	@echo "  make run          - Run application locally"
	@echo "  make test         - Run tests"
	@echo "  make test-cover   - Run tests with coverage"
	@echo "  make build        - Build binary"
	@echo "  make clean        - Clean build artifacts"
	@echo ""
	@echo "Code Quality:"
	@echo "  make lint         - Run go vet"
	@echo "  make fmt          - Format code"
	@echo "  make fmt-check    - Check formatting"
	@echo ""
	@echo "Docker:"
	@echo "  make docker-build - Build Docker image"
	@echo "  make docker-run   - Run Docker container"
	@echo "  make docker-push  - Push Docker image to registry"
	@echo ""
	@echo "Kubernetes:"
	@echo "  make k8s-dev      - Deploy to dev (kustomize)"
	@echo "  make k8s-prod     - Deploy to prod (kustomize)"
	@echo "  make k8s-clean    - Clean K8s deployments"
	@echo ""

# Development targets

run:
	@echo "Starting application..."
	cd $(APP_DIR) && PORT=8080 LOG_LEVEL=info go run main.go

build:
	@echo "Building binary..."
	cd $(APP_DIR) && CGO_ENABLED=0 GOOS=linux go build \
		-a -installsuffix cgo \
		-ldflags="-w -s -X main.Version=$(VERSION)" \
		-o $(BINARY) main.go
	@echo "Binary: $(APP_DIR)/$(BINARY)"

test:
	@echo "Running tests..."
	cd $(APP_DIR) && go test -v

test-cover:
	@echo "Running tests with coverage..."
	cd $(APP_DIR) && go test -v -coverprofile=coverage.out
	cd $(APP_DIR) && go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: $(APP_DIR)/coverage.html"

clean:
	@echo "Cleaning..."
	cd $(APP_DIR) && rm -f $(BINARY) coverage.out coverage.html
	cd $(APP_DIR) && go clean

# Code quality targets

lint:
	@echo "Running go vet..."
	cd $(APP_DIR) && go vet ./...

fmt:
	@echo "Formatting code..."
	cd $(APP_DIR) && gofmt -s -w .

fmt-check:
	@echo "Checking formatting..."
	@cd $(APP_DIR) && if gofmt -s -l . | grep -q .; then \
		echo "✗ Code needs formatting"; \
		gofmt -s -d .; \
		exit 1; \
	else \
		echo "✓ Code is formatted"; \
	fi

# Docker targets

docker-build:
	@echo "Building Docker image: $(DOCKER_IMAGE):$(VERSION)"
	cd $(APP_DIR) && docker build -t $(DOCKER_IMAGE):$(VERSION) -t $(DOCKER_IMAGE):latest .
	@echo "Image built successfully"
	@docker images | grep $(DOCKER_IMAGE)

docker-run:
	@echo "Running Docker container..."
	docker run -p 8080:8080 \
		-e PORT=8080 \
		-e LOG_LEVEL=info \
		$(DOCKER_IMAGE):latest

docker-push:
	@echo "Pushing Docker image to registry..."
	docker tag $(DOCKER_IMAGE):$(VERSION) $(DOCKER_REGISTRY)/$(DOCKER_IMAGE):$(VERSION)
	docker tag $(DOCKER_IMAGE):latest $(DOCKER_REGISTRY)/$(DOCKER_IMAGE):latest
	docker push $(DOCKER_REGISTRY)/$(DOCKER_IMAGE):$(VERSION)
	docker push $(DOCKER_REGISTRY)/$(DOCKER_IMAGE):latest


# Check dependencies

deps:
	@echo "Checking dependencies..."
	@command -v go >/dev/null 2>&1 || { echo "Go is not installed"; exit 1; }
	@command -v docker >/dev/null 2>&1 || { echo "Docker is not installed"; exit 1; }
	@command -v kubectl >/dev/null 2>&1 || { echo "kubectl is not installed"; exit 1; }
	@echo "All dependencies are installed"

# Defaults

.DEFAULT_GOAL := help