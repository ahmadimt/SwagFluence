.PHONY: build test lint clean install run fmt

BINARY_NAME=SwagFluence
BUILD_DIR=bin
MAIN_PATH=./cmd/swagfluence

# Build the application
build:
	@echo "Building..."
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY_NAME) -ldflags="-s -w" $(MAIN_PATH)

# Run tests with coverage
test:
	@echo "Running tests..."
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Run linter
lint:
	@echo "Running linter..."
	golangci-lint run

# Format code
fmt:
	@echo "Formatting code..."
	gofmt -s -w .
	goimports -w .

# Run the application
run:
	go run $(MAIN_PATH) $(ARGS)

# Install dependencies
install:
	go mod download
	go mod tidy

# Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html

# Cross-compile for multiple platforms
build-all:
	@echo "Building for multiple platforms..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(MAIN_PATH)
	GOOS=darwin GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 $(MAIN_PATH)
	GOOS=darwin GOARCH=arm64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 $(MAIN_PATH)
	GOOS=windows GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe $(MAIN_PATH)

# Run with example
example: build
	./$(BUILD_DIR)/$(BINARY_NAME) https://petstore.swagger.io/v2/swagger.json

help:
	@echo "Available targets:"
	@echo "  build       - Build the application"
	@echo "  test        - Run tests with coverage"
	@echo "  lint        - Run linter"
	@echo "  fmt         - Format code"
	@echo "  run         - Run the application (use ARGS='url')"
	@echo "  install     - Install dependencies"
	@echo "  clean       - Clean build artifacts"
	@echo "  build-all   - Cross-compile for multiple platforms"
	@echo "  example     - Run with petstore example"