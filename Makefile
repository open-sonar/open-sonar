.PHONY: build test test-internal test-package test-e2e test-docker run clean dev tidy

# Default compiler flags
GO_FLAGS=-trimpath -ldflags "-s -w"

# Binary output
BINARY_NAME=opensonar

build:
	@echo "Building Open Sonar..."
	@go build $(GO_FLAGS) -o $(BINARY_NAME) ./cmd/server

# Individual test targets
test-internal:
	@echo "Running internal tests..."
	@go test ./internal/... -v

test-package:
	@echo "Running package tests..."
	@chmod +x ./scripts/package_test.sh
	@./scripts/package_test.sh

test-e2e:
	@echo "Running end-to-end tests..."
	@chmod +x ./scripts/e2e_test.sh
	@./scripts/e2e_test.sh

test-docker:
	@echo "Running Docker tests..."
	@chmod +x ./scripts/docker_test.sh
	@./scripts/docker_test.sh

# Master test target that runs all tests
test: test-internal test-package test-e2e test-docker
	@echo "All tests completed successfully"

run: build
	@echo "Starting Open Sonar server..."
	@./$(BINARY_NAME)

clean:
	@echo "Cleaning up..."
	@rm -f $(BINARY_NAME)
	@go clean

dev: 
	@echo "Starting development server..."
	@go run ./cmd/server/main.go

tidy:
	@echo "Tidying Go modules..."
	@go mod tidy
