#!/bin/bash
# Development server for Open Sonar
# Builds and runs the server with hot reloading

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[0;33m'
NC='\033[0m' # No Color

echo -e "${YELLOW}Setting up Open Sonar development server${NC}"

# Check if .env file exists, create if not
if [ ! -f .env ]; then
  echo -e "${YELLOW}Creating default .env file...${NC}"
  cp example.env .env
  echo -e "${GREEN}Created .env file. Please edit it with your settings if needed.${NC}"
fi

# Check if Ollama is running
if ! curl -s "http://localhost:11434/api/version" > /dev/null; then
  echo -e "${RED}Ollama server not detected at http://localhost:11434${NC}"
  echo -e "${YELLOW}For the best experience, we recommend installing Ollama:${NC}"
  echo -e "${YELLOW}  curl https://ollama.ai/install.sh | sh${NC}"
  echo -e "${YELLOW}  ollama pull deepseek-r1:1.5b${NC}"
  echo -e "${YELLOW}Otherwise, set TEST_MODE=true in your .env file.${NC}"
  
  # Ask user if they want to continue anyway
  read -p "Continue without Ollama? (y/n) " -n 1 -r
  echo
  if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo -e "${RED}Exiting.${NC}"
    exit 1
  fi
fi

# Build the server
echo -e "${YELLOW}Building Open Sonar...${NC}"
go build -o opensonar ./cmd/server

echo -e "${GREEN}Build successful!${NC}"

# Set default port if not specified
PORT=$(grep "PORT" .env | cut -d '=' -f2)
if [ -z "$PORT" ]; then
  PORT=8080
fi

echo -e "${GREEN}Starting server on port ${PORT}...${NC}"
echo -e "${GREEN}Access the API at http://localhost:${PORT}${NC}"
echo -e "${YELLOW}Press Ctrl+C to stop the server${NC}"

# Run the server
./opensonar

# Clean up on exit
function cleanup {
  echo -e "${YELLOW}Cleaning up...${NC}"
  rm -f opensonar
}

# Register cleanup function
trap cleanup EXIT
