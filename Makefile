.PHONY: build run test clean install deps linux-build lint lint-fix fmt-check

# Build the application
build:
	rm -f fb2epub
	go build -trimpath -o fb2epub
	@echo "Build complete. If you encounter LC_UUID errors on macOS, use 'make run' instead."

# Run the application (recommended for local development)
run:
	go run main.go

# Run tests
test:
	go test -v ./...

# Clean build artifacts
clean:
	rm -f fb2epub
	rm -f fb2epub-linux
	rm -f fb2epub-macos
	rm -rf /tmp/fb2epub

# Install dependencies
deps:
	go mod download
	go mod tidy

# Build for Linux (for VPS deployment)
linux-build:
	GOOS=linux GOARCH=amd64 go build -o fb2epub-linux

# Build for macOS
macos-build:
	GOOS=darwin GOARCH=amd64 go build -o fb2epub-macos

# Format code
fmt:
	go fmt ./...
	@echo "Code formatted successfully."

# Lint code
lint:
	@echo "Running golangci-lint..."
	@GOPATH_BIN=$$(go env GOPATH)/bin; \
	if ! command -v golangci-lint >/dev/null 2>&1 && [ ! -f "$$GOPATH_BIN/golangci-lint" ]; then \
		echo "golangci-lint not found. Installing..."; \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$GOPATH_BIN latest; \
	fi; \
	if [ -f "$$GOPATH_BIN/golangci-lint" ]; then \
		$$GOPATH_BIN/golangci-lint run; \
	elif command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "Error: golangci-lint not found. Please install it manually."; \
		exit 1; \
	fi

# Lint and auto-fix issues
lint-fix:
	@echo "Running golangci-lint with auto-fix..."
	@GOPATH_BIN=$$(go env GOPATH)/bin; \
	if ! command -v golangci-lint >/dev/null 2>&1 && [ ! -f "$$GOPATH_BIN/golangci-lint" ]; then \
		echo "golangci-lint not found. Installing..."; \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$GOPATH_BIN latest; \
	fi; \
	if [ -f "$$GOPATH_BIN/golangci-lint" ]; then \
		$$GOPATH_BIN/golangci-lint run --fix; \
	elif command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run --fix; \
	else \
		echo "Error: golangci-lint not found. Please install it manually."; \
		exit 1; \
	fi

# Check if code is formatted
fmt-check:
	@echo "Checking code formatting..."
	@if [ $$(gofmt -l . | wc -l) -ne 0 ]; then \
		echo "Code is not formatted. Run 'make fmt' to fix."; \
		gofmt -d .; \
		exit 1; \
	fi
	@echo "Code is properly formatted."

# Run with development settings
dev:
	ENVIRONMENT=development PORT=8080 go run main.go

