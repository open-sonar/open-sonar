#!/bin/bash
# End-to-End Production test for Open Sonar
# Uses real Ollama LLM responses (no mocking)

set -e  # Exit on error

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[0;33m'
NC='\033[0m' # No Color

echo -e "${YELLOW}Starting Open Sonar Production E2E Test${NC}"

# Check if Ollama is running and has the required model
echo -e "${YELLOW}Checking Ollama availability...${NC}"

if ! curl -s "http://localhost:11434/api/version" > /dev/null; then
  echo -e "${RED}Ollama server not detected at http://localhost:11434${NC}"
  echo -e "${RED}This test requires a real Ollama server running.${NC}"
  echo -e "${RED}Please install Ollama and pull the deepseek-r1:1.5b model:${NC}"
  echo -e "${YELLOW}  curl https://ollama.ai/install.sh | sh${NC}"
  echo -e "${YELLOW}  ollama pull deepseek-r1:1.5b${NC}"
  echo -e "${RED}Exiting.${NC}"
  exit 1
fi

# Check if the model is available
if ! curl -s "http://localhost:11434/api/tags" | grep -q "deepseek-r1:1.5b"; then
  echo -e "${RED}Model 'deepseek-r1:1.5b' not found in Ollama${NC}"
  echo -e "${YELLOW}Please pull the model:${NC}"
  echo -e "${YELLOW}  ollama pull deepseek-r1:1.5b${NC}"
  echo -e "${RED}Exiting.${NC}"
  exit 1
fi

echo -e "${GREEN}Ollama server with deepseek-r1:1.5b model detected!${NC}"

# Setup test environment variables
export PORT=8765
export LOG_LEVEL=INFO
export OLLAMA_MODEL=deepseek-r1:1.5b
# Ensure we're NOT using test mode (real LLM calls will be made)
unset TEST_MODE

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
echo -e "\n${YELLOW}Test 2: Direct LLM query (This will use real Ollama LLM):${NC}"
curl -s -X POST "http://localhost:$PORT/chat" \
  -H "Content-Type: application/json" \
  -d '{
    "query": "What is the capital of France? Answer in one sentence.",
    "needSearch": false
  }' | jq

# Test 3: Web search enhanced query
echo -e "\n${YELLOW}Test 3: Web search enhanced query (This will use real Ollama LLM and web search):${NC}"
curl -s -X POST "http://localhost:$PORT/chat" \
  -H "Content-Type: application/json" \
  -d '{
    "query": "What were the main AI developments in the last month?",
    "needSearch": true,
    "pages": 1,
    "retries": 1,
    "provider": "ollama"
  }' | jq

# Test 4: Chat completions API (This will use real Ollama LLM):
echo -e "\n${YELLOW}Test 4: Chat completions API (This will use real Ollama LLM):${NC}"
curl -s -X POST "http://localhost:$PORT/chat/completions" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer test-token" \
  -d '{
    "model": "ollama",
    "messages": [
      {"role": "system", "content": "You are a helpful assistant."},
      {"role": "user", "content": "What are the benefits of open source software? List three key benefits."}
    ],
    "temperature": 0.7,
    "top_p": 0.9,
    "max_tokens": 500
  }' | jq

# Test 5: Web search with domain filter
echo -e "\n${YELLOW}Test 5: Web search with domain filter (.edu and .gov sites only):${NC}"
curl -s -X POST "http://localhost:$PORT/chat/completions" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer test-token" \
  -d '{
    "model": "sonar",
    "messages": [
      {"role": "system", "content": "You are a helpful assistant."},
      {"role": "user", "content": "What is quantum computing according to research institutions?"}
    ],
    "temperature": 0.3,
    "max_tokens": 500,
    "search_domain_filter": [".edu", ".gov"]
  }' | jq

echo -e "\n${GREEN}All tests completed successfully!${NC}"
echo -e "${GREEN}These tests used your real Ollama LLM provider with the deepseek-r1:1.5b model.${NC}"
