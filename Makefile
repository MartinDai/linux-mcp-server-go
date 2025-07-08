# Linux MCP Server Go - Cross-Platform Build Makefile

BINARY_NAME=linux-mcp-server
VERSION=v0.0.1
BUILD_DIR=build
SOURCE_DIR=.

# Go build flags
LDFLAGS=-ldflags "-X main.Version=$(VERSION)"

# Default target
.PHONY: all
all: clean build

# Clean build directory
.PHONY: clean
clean:
	rm -rf $(BUILD_DIR)
	mkdir -p $(BUILD_DIR)

# Build for current platform
.PHONY: build
build:
	go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(SOURCE_DIR)

# Cross-platform builds
.PHONY: build-linux-amd64
build-linux-amd64:
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(SOURCE_DIR)

.PHONY: build-linux-arm64
build-linux-arm64:
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 $(SOURCE_DIR)

.PHONY: build-linux-386
build-linux-386:
	GOOS=linux GOARCH=386 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-386 $(SOURCE_DIR)

.PHONY: build-darwin-amd64
build-darwin-amd64:
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 $(SOURCE_DIR)

.PHONY: build-darwin-arm64
build-darwin-arm64:
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 $(SOURCE_DIR)

.PHONY: build-windows-amd64
build-windows-amd64:
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe $(SOURCE_DIR)

.PHONY: build-windows-386
build-windows-386:
	GOOS=windows GOARCH=386 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-386.exe $(SOURCE_DIR)

.PHONY: build-freebsd-amd64
build-freebsd-amd64:
	GOOS=freebsd GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-freebsd-amd64 $(SOURCE_DIR)

# Build all platforms
.PHONY: build-all
build-all: clean build-linux-amd64 build-linux-arm64 build-linux-386 build-darwin-amd64 build-darwin-arm64 build-windows-amd64 build-windows-386 build-freebsd-amd64

# Development tasks
.PHONY: deps
deps:
	go mod download
	go mod tidy

.PHONY: test
test:
	go test -v ./...

.PHONY: fmt
fmt:
	go fmt ./...

.PHONY: vet
vet:
	go vet ./...

.PHONY: lint
lint:
	golangci-lint run

# Run the server
.PHONY: run
run: build
	./$(BUILD_DIR)/$(BINARY_NAME)

# Run with HTTP mode
.PHONY: run-http
run-http: build
	./$(BUILD_DIR)/$(BINARY_NAME) -http :8080

# Install dependencies
.PHONY: install-deps
install-deps:
	go mod download

# Show help
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  all              - Clean and build for current platform"
	@echo "  clean            - Remove build directory"
	@echo "  build            - Build for current platform"
	@echo "  build-all        - Build for all supported platforms"
	@echo "  build-linux-*    - Build for Linux (amd64, arm64, 386)"
	@echo "  build-darwin-*   - Build for macOS (amd64, arm64)"
	@echo "  build-windows-*  - Build for Windows (amd64, 386)"
	@echo "  build-freebsd-*  - Build for FreeBSD (amd64)"
	@echo "  deps             - Download and tidy dependencies"
	@echo "  test             - Run tests"
	@echo "  fmt              - Format code"
	@echo "  vet              - Run go vet"
	@echo "  lint             - Run golangci-lint"
	@echo "  run              - Build and run server"
	@echo "  run-http         - Build and run server in HTTP mode"
	@echo "  help             - Show this help"