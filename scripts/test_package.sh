#!/bin/bash
# Test script for the sonar package

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[0;33m'
NC='\033[0m' # No Color

echo -e "${YELLOW}Testing Open Sonar Package${NC}"

# Check if Ollama is running
echo -e "${YELLOW}Checking Ollama availability...${NC}"
if ! curl -s "http://localhost:11434/api/version" > /dev/null; then
  echo -e "${RED}Ollama server not detected${NC}"
  echo -e "${YELLOW}Setting SKIP_OLLAMA_TESTS=true${NC}"
  export SKIP_OLLAMA_TESTS=true
else
  echo -e "${GREEN}Ollama server detected${NC}"
  export SKIP_OLLAMA_TESTS=false
  
  # Check if deepseek model is available
  if ! curl -s "http://localhost:11434/api/tags" | grep -q "deepseek-r1:1.5b"; then
    echo -e "${YELLOW}Model 'deepseek-r1:1.5b' not found in Ollama${NC}"
    echo -e "${YELLOW}Consider pulling it with: ollama pull deepseek-r1:1.5b${NC}"
    echo -e "${YELLOW}Setting SKIP_OLLAMA_TESTS=true${NC}"
    export SKIP_OLLAMA_TESTS=true
  else
    echo -e "${GREEN}Model 'deepseek-r1:1.5b' available${NC}"
  fi
fi

# Run the tests with verbose output
echo -e "${YELLOW}Running package tests...${NC}"
go test -v ./sonar

if [ $? -eq 0 ]; then
  echo -e "${GREEN}All tests passed!${NC}"
  exit 0
else
  echo -e "${RED}Test failures detected${NC}"
  exit 1
fi
