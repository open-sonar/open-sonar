#!/bin/bash

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[0;33m'
NC='\033[0m' # No Color

echo -e "${YELLOW}Testing Open Sonar Package${NC}"

# Check if Ollama is available
echo -e "${YELLOW}Checking Ollama availability...${NC}"
if curl -s http://localhost:11434/api/version > /dev/null; then
    echo -e "${GREEN}Ollama server detected${NC}"
    
    # Check if the default model is available
    OLLAMA_MODEL=${OLLAMA_MODEL:-"deepseek-r1:1.5b"}
    if ollama list 2>/dev/null | grep -q "$OLLAMA_MODEL"; then
        echo -e "${GREEN}Model '${OLLAMA_MODEL}' available${NC}"
    else
        echo -e "${YELLOW}Model '${OLLAMA_MODEL}' not found, tests may fail${NC}"
    fi
else
    echo -e "${YELLOW}Ollama server not detected, some tests may be skipped${NC}"
fi

# Run the package tests
echo -e "${YELLOW}Running package tests...${NC}"

if SKIP_OLLAMA_TESTS=${SKIP_OLLAMA_TESTS:-false} go test ./sonar/... -v; then
    echo -e "${GREEN}All package tests passed!${NC}"
    exit 0
else
    echo -e "${RED}Test failures detected${NC}"
    exit 1
fi
