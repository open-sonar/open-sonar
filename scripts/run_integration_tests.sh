#!/bin/bash
# Run integration tests with Ollama

set -e

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[0;33m'
NC='\033[0m' # No Color

echo -e "${YELLOW}Checking if Ollama is available...${NC}"

# Check if Ollama is available
if ! curl -s "http://localhost:11434/api/version" > /dev/null; then
  echo -e "${RED}Ollama server not detected at http://localhost:11434${NC}"
  echo -e "${YELLOW}Please ensure Ollama is running with the deepseek-r1:1.5b model.${NC}"
  echo -e "${YELLOW}You can pull it with: ollama pull deepseek-r1:1.5b${NC}"
  echo -e "${YELLOW}Skipping integration tests.${NC}"
  exit 0
fi

echo -e "${GREEN}Ollama server detected!${NC}"

# Export variables to ensure we don't use TEST_MODE
unset TEST_MODE
export OLLAMA_MODEL=deepseek-r1:1.5b

# Run the integration tests
echo -e "${YELLOW}Running Ollama integration tests...${NC}"
go test -v ./internal/llm -run TestOllamaIntegration

echo -e "${GREEN}Integration tests completed!${NC}"
