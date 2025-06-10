.PHONY: all build clean test broker router coder protocol install-deps

# Build output directory
BIN_DIR := bin

# Default target
all: install-deps build

# Install dependencies
install-deps:
	@echo "Installing dependencies..."
	cd protocol/go && go mod tidy
	cd broker && go mod tidy
	cd router && go mod tidy
	cd bodies/coder && go mod tidy

# Build all components
build: broker router coder

# Build broker
broker:
	@echo "Building fem-broker..."
	@mkdir -p $(BIN_DIR)
	cd broker && go build -o ../$(BIN_DIR)/fem-broker ./cmd/fem-broker

# Build router
router:
	@echo "Building fem-router..."
	@mkdir -p $(BIN_DIR)
	cd router && go build -o ../$(BIN_DIR)/fem-router ./cmd/fem-router

# Build coder
coder:
	@echo "Building fem-coder..."
	@mkdir -p $(BIN_DIR)
	cd bodies/coder && go build -o ../../$(BIN_DIR)/fem-coder ./cmd/fem-coder

# Build protocol package
protocol:
	@echo "Building protocol package..."
	cd protocol/go && go build ./...

# Run tests
test:
	@echo "Running tests..."
	cd protocol/go && go test ./...
	cd broker && go test ./...
	cd router && go test ./...
	cd bodies/coder && go test ./...

# Run broker
run-broker: broker
	./$(BIN_DIR)/fem-broker

# Run router
run-router: router
	./$(BIN_DIR)/fem-router

# Run coder
run-coder: coder
	./$(BIN_DIR)/fem-coder

# Docker builds
docker-build:
	docker build -t fem-broker broker/
	docker build -t fem-router router/
	docker build -t fem-coder bodies/coder/

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -rf $(BIN_DIR)
	find . -name "*.test" -delete
	find . -name "*.out" -delete

# Format code
fmt:
	@echo "Formatting code..."
	cd protocol/go && go fmt ./...
	cd broker && go fmt ./...
	cd router && go fmt ./...
	cd bodies/coder && go fmt ./...

# Lint code
lint:
	@echo "Linting code..."
	cd protocol/go && go vet ./...
	cd broker && go vet ./...
	cd router && go vet ./...
	cd bodies/coder && go vet ./...

# Generate self-signed certificates for testing
gen-certs:
	@echo "Generating self-signed certificates..."
	@mkdir -p certs
	openssl req -x509 -newkey rsa:4096 -keyout certs/server.key -out certs/server.crt -days 365 -nodes -subj "/CN=localhost"

# Quick start - build and run broker
quickstart: build gen-certs
	@echo "Starting FEM broker on :4433..."
	./$(BIN_DIR)/fem-broker --listen :4433