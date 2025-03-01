.PHONY: build test run clean deps clean-deps e2e-test langchain integration-test e2e-test-prod

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

deps:
	@echo "Updating dependencies..."
	@chmod +x scripts/update_deps.sh
	@./scripts/update_deps.sh

clean-deps:
	@echo "Cleaning and rebuilding dependencies..."
	@chmod +x scripts/clean_deps.sh
	@./scripts/clean_deps.sh

e2e-test:
	@echo "Running E2E tests with mock LLM..."
	@chmod +x scripts/e2e_test.sh
	@./scripts/e2e_test.sh

e2e-test-prod:
	@echo "Running E2E tests with real Ollama LLM..."
	@chmod +x scripts/e2e_test_prod.sh
	@./scripts/e2e_test_prod.sh

langchain:
	@echo "Updating langchain dependencies..."
	@chmod +x scripts/update_langchain.sh
	@./scripts/update_langchain.sh

integration-test:
	@echo "Running integration tests..."
	@chmod +x scripts/run_integration_tests.sh
	@./scripts/run_integration_tests.sh
