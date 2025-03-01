#!/bin/bash
# Run all tests in the project

# Set color variables
GREEN="\033[0;32m"
RED="\033[0;31m"
YELLOW="\033[0;33m"
NC="\033[0m" # No color

echo -e "${YELLOW}Running tests for all packages...${NC}"

# Run tests for all packages with verbose output
go test ./internal/... -v

# Check the exit status
if [ $? -eq 0 ]; then
    echo -e "${GREEN}All tests passed!${NC}"
    exit 0
else
    echo -e "${RED}Some tests failed!${NC}"
    exit 1
fi
