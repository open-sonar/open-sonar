#!/bin/bash
# End-to-End Production test for Open Sonar
# Uses real Ollama LLM responses (no mocking)

set -e  # Exit on error

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[0;33m'
NC='\033[0m' # No Color

echo -e "${YELLOW}Running End-to-End Tests${NC}"

# Build the server binary
go build -o opensonar_test_bin ./cmd/server

# Check if the build succeeded
if [ $? -ne 0 ]; then
    echo -e "${RED}Failed to build server binary${NC}"
    exit 1
fi

# Start the server in the background
echo -e "${YELLOW}Starting test server...${NC}"
TEST_MODE=true AUTH_TOKEN=test-token PORT=8099 ./opensonar_test_bin > test_server.log 2>&1 &
SERVER_PID=$!

# Ensure we kill the server when the script exits
trap "kill $SERVER_PID 2>/dev/null || true; rm -f opensonar_test_bin test_server.log; echo -e '${YELLOW}Test server stopped and cleaned up${NC}'" EXIT

# Wait for server to be ready
echo -e "${YELLOW}Waiting for server to be ready...${NC}"
attempt=0
max_attempts=30
while [ $attempt -lt $max_attempts ]; do
    if curl -s http://localhost:8099/test > /dev/null; then
        echo -e "${GREEN}Server is ready${NC}"
        break
    fi
    attempt=$((attempt+1))
    sleep 1
    echo -n "."
done

if [ $attempt -eq $max_attempts ]; then
    echo -e "\n${RED}Server failed to start in time${NC}"
    cat test_server.log
    exit 1
fi

# Run E2E tests
echo -e "${YELLOW}Running E2E API tests...${NC}"

# Test 1: Basic test endpoint
echo -e "${YELLOW}Test 1: Basic test endpoint${NC}"
response=$(curl -s http://localhost:8099/test)
if echo "$response" | grep -q "open-sonar server is running"; then
    echo -e "${GREEN}Test 1 passed${NC}"
else
    echo -e "${RED}Test 1 failed: $response${NC}"
    exit 1
fi

# Test 2: Check chat endpoint with simple query
echo -e "${YELLOW}Test 2: Chat endpoint${NC}"
response=$(curl -s -X POST http://localhost:8099/chat \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer test-token" \
  -d '{"query": "What is 2+2?", "needSearch": false, "provider": "mock"}')

if echo "$response" | grep -q "response"; then
    echo -e "${GREEN}Test 2 passed${NC}"
else
    echo -e "${RED}Test 2 failed: $response${NC}"
    exit 1
fi

# All tests passed
echo -e "${GREEN}All E2E tests passed!${NC}"
exit 0
