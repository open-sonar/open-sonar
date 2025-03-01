.PHONY: build test run clean

# Default compiler flags
GO_FLAGS=-trimpath -ldflags "-s -w"

# Binary output
BINARY_NAME=opensonar

build:
	@echo "Building Open Sonar..."
	@go build $(GO_FLAGS) -o $(BINARY_NAME) ./cmd/server

test:
	@echo "Running tests..."
	@go test ./internal/... -v

run: build
	@echo "Starting Open Sonar server..."
	@./$(BINARY_NAME)

clean:
	@echo "Cleaning up..."
	@rm -f $(BINARY_NAME)
	@go clean
