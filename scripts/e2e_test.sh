#!/bin/bash
# End-to-End test for Open Sonar
# Tests the complete flow from compilation to API response

set -e  # Exit on error

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[0;33m'
NC='\033[0m' # No Color

echo -e "${YELLOW}Starting Open Sonar E2E Test${NC}"

# Setup test environment variables
export PORT=8765
export TEST_MODE=true
export LOG_LEVEL=INFO

# Working directory
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
PROJ_ROOT="$(dirname "$SCRIPT_DIR")"
cd "$PROJ_ROOT"

# Build the server
echo -e "${YELLOW}Compiling server...${NC}"
go build -o opensonar_test ./cmd/server

# Run server in background
echo -e "${YELLOW}Starting server on port $PORT...${NC}"
./opensonar_test &
SERVER_PID=$!

# Give the server time to start
echo -e "${YELLOW}Waiting for server to start...${NC}"
sleep 2

# Function to clean up server process on exit
function cleanup {
  echo -e "${YELLOW}Stopping server...${NC}"
  kill $SERVER_PID 2>/dev/null || true
  rm -f opensonar_test
}

# Register the cleanup function to be called on script exit
trap cleanup EXIT

# Check if server is running
if ! curl -s "http://localhost:$PORT/test" > /dev/null; then
  echo -e "${RED}Server failed to start! Exiting.${NC}"
  exit 1
fi

echo -e "${GREEN}Server is running.${NC}"

# Test 1: Basic test endpoint
echo -e "\n${YELLOW}Test 1: Basic server test:${NC}"
curl -s "http://localhost:$PORT/test" | jq

# Test 2: Direct LLM query (no search)
echo -e "\n${YELLOW}Test 2: Direct LLM query:${NC}"
curl -s -X POST "http://localhost:$PORT/chat" \
  -H "Content-Type: application/json" \
  -d '{
    "query": "What is the capital of France?",
    "needSearch": false,
    "provider": "mock"
  }' | jq

# Test 3: Web search enhanced query
echo -e "\n${YELLOW}Test 3: Web search enhanced query:${NC}"
curl -s -X POST "http://localhost:$PORT/chat" \
  -H "Content-Type: application/json" \
  -d '{
    "query": "Latest news about artificial intelligence",
    "needSearch": true,
    "pages": 1,
    "retries": 1,
    "provider": "mock"
  }' | jq

# Test 4: Chat completions endpoint (OpenAI API compatible)
echo -e "\n${YELLOW}Test 4: Chat completions API:${NC}"
curl -s -X POST "http://localhost:$PORT/chat/completions" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer test-token" \
  -d '{
    "model": "mock",
    "messages": [
      {"role": "system", "content": "You are a helpful assistant."},
      {"role": "user", "content": "What are the benefits of open source software?"}
    ],
    "temperature": 0.7,
    "top_p": 0.9,
    "max_tokens": 500
  }' | jq

echo -e "\n${GREEN}All tests completed successfully!${NC}"
echo "Note: These tests used the mock LLM provider. For real responses, configure your .env with proper API keys."
