# Makefile for go-cqrs project
# Compatible with both Windows and Linux

# Detect OS
ifeq ($(OS),Windows_NT)
	# Windows commands
	DETECTED_OS := Windows
	RM := del /Q /F
	RMDIR := rmdir /S /Q
	MKDIR := mkdir
	PATHSEP := \\
	BINARY_EXT := .exe
else
	# Linux/Unix commands
	DETECTED_OS := $(shell uname -s)
	RM := rm -f
	RMDIR := rm -rf
	MKDIR := mkdir -p
	PATHSEP := /
	BINARY_EXT :=
endif

# Project variables
PROJECT_NAME := go-cqrs
MODULE_NAME := github.com/victoragudo/go-cqrs
BUILD_DIR := build
BINARY_NAME := $(PROJECT_NAME)$(BINARY_EXT)
COVERAGE_FILE := coverage.out

# Go commands
GOCMD := go
GOBUILD := $(GOCMD) build
GOCLEAN := $(GOCMD) clean
GOTEST := $(GOCMD) test
GOGET := $(GOCMD) get
GOMOD := $(GOCMD) mod
GOFMT := $(GOCMD) fmt

# Build flags
BUILD_FLAGS := -v
TEST_FLAGS := -v -race
COVERAGE_FLAGS := -v -race -coverprofile=$(COVERAGE_FILE)

.PHONY: help fmt build test clean deps tidy check coverage install uninstall run dev lint vet

# Default target
all: deps fmt vet test build

# Display help information
help:
	@echo "Available targets for $(PROJECT_NAME) ($(DETECTED_OS)):"
	@echo ""
	@echo "  help        - Show this help message"
	@echo "  fmt         - Format Go source code"
	@echo "  build       - Build the project"
	@echo "  test        - Run tests"
	@echo "  test-cover  - Run tests with coverage"
	@echo "  clean       - Clean build artifacts"
	@echo "  deps        - Download dependencies"
	@echo "  tidy        - Clean up go.mod and go.sum"
	@echo "  check       - Run fmt, vet, and test"
	@echo "  coverage    - Generate and display test coverage"
	@echo "  lint        - Run golint (requires golint to be installed)"
	@echo "  vet         - Run go vet"
	@echo "  dev         - Development mode (fmt + test + build)"
	@echo ""
	@echo "OS detected: $(DETECTED_OS)"

# Format Go source code
fmt:
	@echo "Formatting Go source code..."
	$(GOFMT) ./...

# Build the project
build: fmt
	@echo "Building $(PROJECT_NAME) for $(DETECTED_OS)..."
ifeq ($(OS),Windows_NT)
	@if not exist $(BUILD_DIR) $(MKDIR) $(BUILD_DIR)
else
	@$(MKDIR) $(BUILD_DIR)
endif
	$(GOBUILD) $(BUILD_FLAGS) -o $(BUILD_DIR)$(PATHSEP)$(BINARY_NAME) $(MODULE_NAME)

# Run tests
test: fmt
	@echo "Running tests..."
	$(GOTEST) $(TEST_FLAGS) ./...

# Run tests with coverage
test-cover: fmt
	@echo "Running tests with coverage..."
	$(GOTEST) $(COVERAGE_FLAGS) ./...

# Generate and display coverage report
coverage: test-cover
	@echo "Generating coverage report..."
	$(GOCMD) tool cover -html=$(COVERAGE_FILE) -o coverage.html
	@echo "Coverage report generated: coverage.html"
ifeq ($(OS),Windows_NT)
	@echo "Opening coverage report..."
	@start coverage.html
else
	@echo "Coverage report saved as coverage.html"
	@echo "Open with: xdg-open coverage.html"
endif

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	$(GOCLEAN)
ifeq ($(OS),Windows_NT)
	@if exist $(BUILD_DIR) $(RMDIR) $(BUILD_DIR) 2>nul || echo Build directory already clean
	@if exist $(COVERAGE_FILE) $(RM) $(COVERAGE_FILE) 2>nul || echo Coverage file already clean
	@if exist coverage.html $(RM) coverage.html 2>nul || echo Coverage HTML already clean
else
	@$(RMDIR) $(BUILD_DIR) 2>/dev/null || echo "Build directory already clean"
	@$(RM) $(COVERAGE_FILE) 2>/dev/null || echo "Coverage file already clean"
	@$(RM) coverage.html 2>/dev/null || echo "Coverage HTML already clean"
endif

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	$(GOGET) -d ./...

# Clean up go.mod and go.sum
tidy:
	@echo "Tidying go.mod and go.sum..."
	$(GOMOD) tidy

# Run formatting, vetting, and testing
check: fmt vet test
	@echo "All checks passed!"

# Run go vet
vet:
	@echo "Running go vet..."
	$(GOCMD) vet ./...

# Run golint (requires golint to be installed)
lint:
	@echo "Running golint..."
	@which golint > /dev/null || (echo "golint not found. Install with: go install golang.org/x/lint/golint@latest" && exit 1)
	golint ./...

# Development mode - quick feedback loop
dev: fmt test build
	@echo "Development build complete!"

# Install the binary to GOPATH/bin
install: build
	@echo "Installing $(PROJECT_NAME)..."
	$(GOCMD) install $(MODULE_NAME)

# Uninstall the binary from GOPATH/bin
uninstall:
	@echo "Uninstalling $(PROJECT_NAME)..."
ifeq ($(OS),Windows_NT)
	@if exist $(GOPATH)\bin\$(BINARY_NAME) $(RM) $(GOPATH)\bin\$(BINARY_NAME)
else
	@$(RM) $(GOPATH)/bin/$(BINARY_NAME)
endif

# Show project information
info:
	@echo "Project: $(PROJECT_NAME)"
	@echo "Module: $(MODULE_NAME)"
	@echo "OS: $(DETECTED_OS)"
	@echo "Go version: $(shell $(GOCMD) version)"
	@echo "Build directory: $(BUILD_DIR)"
	@echo "Binary name: $(BINARY_NAME)"