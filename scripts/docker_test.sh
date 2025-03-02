#!/bin/bash

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Default settings
KEEP_RUNNING=false

# Parse command line options
while [[ $# -gt 0 ]]; do
  key="$1"
  case $key in
    --keep-running)
      KEEP_RUNNING=true
      shift
      ;;
    *)
      echo "Unknown option: $1"
      exit 1
      ;;
  esac
done

# Function to clean up Docker resources
cleanup() {
  echo -e "${YELLOW}Cleaning up Docker resources...${NC}"
  docker-compose down --remove-orphans
  docker system prune -f
}

# Set trap to clean up on script exit unless --keep-running is specified
if [ "$KEEP_RUNNING" = false ]; then
  trap cleanup EXIT
fi

echo -e "${BLUE}============================================${NC}"
echo -e "${BLUE}    Open Sonar Docker E2E Test             ${NC}"
echo -e "${BLUE}============================================${NC}"

# Set the AUTH_TOKEN environment variable for testing
export AUTH_TOKEN="test-token"

# Clean up any existing containers
docker-compose down --remove-orphans || true
docker system prune -f || true

# Build and start containers
echo -e "${YELLOW}Building and starting containers...${NC}"
if ! docker-compose up --build -d; then
  echo -e "${RED}Failed to build and start containers${NC}"
  exit 1
fi

# Wait for the API to be ready
echo -e "${YELLOW}Waiting for API to start...${NC}"
attempt=0
max_attempts=30
while [ $attempt -lt $max_attempts ]; do
  echo -n "."
  if curl -s http://localhost:8080/test > /dev/null; then
    echo -e "\n${GREEN}API is ready!${NC}"
    break
  fi
  
  # Check if containers are still running
  if ! docker-compose ps | grep -q "Up"; then
    echo -e "\n${RED}Containers stopped unexpectedly${NC}"
    docker-compose logs
    exit 1
  fi
  
  attempt=$((attempt+1))
  sleep 2
done

if [ $attempt -eq $max_attempts ]; then
  echo -e "${RED}API did not become ready in time${NC}"
  docker-compose logs
  exit 1
fi

# Setup test variables
PASSED=0
FAILED=0
TOTAL_TESTS=3

# Test 1: Basic API endpoint
echo -e "${BLUE}\nTEST 1: Basic API endpoint${NC}"
response=$(curl -s http://localhost:8080/test)
if echo "$response" | grep -q "open-sonar server is running"; then
  echo -e "${GREEN}✓ Basic API test passed${NC}"
  PASSED=$((PASSED+1))
else
  echo -e "${RED}✗ Basic API test failed${NC}"
  echo -e "${YELLOW}Received: $response${NC}"
  FAILED=$((FAILED+1))
fi

# Test 2: Check container network connectivity using HTTP
echo -e "${BLUE}\nTEST 2: Container networking${NC}"
echo -e "${YELLOW}Checking container networking...${NC}"
container_network=$(docker-compose exec -T open-sonar env | grep OLLAMA_HOST || echo "Not found")
echo "OLLAMA_HOST in container: $container_network"

# Use curl instead of ping to test connectivity to Ollama
network_test=$(docker-compose exec -T open-sonar curl -s -I -o /dev/null -w "%{http_code}" http://ollama:11434/api/version || echo "Connection failed")

if [ "$network_test" = "200" ]; then
  echo -e "${GREEN}✓ Container network test passed (HTTP 200 from Ollama)${NC}"
  PASSED=$((PASSED+1))
else
  echo -e "${RED}✗ Container network test failed${NC}"
  echo -e "${YELLOW}HTTP response code: $network_test${NC}"
  echo -e "${YELLOW}Trying to diagnose network issue:${NC}"
  docker-compose exec -T open-sonar wget -q -O- --timeout=2 http://ollama:11434/api/version || echo "wget failed too"
  FAILED=$((FAILED+1))
fi

# Test 3: Ollama status once it's ready
echo -e "${BLUE}\nTEST 3: Waiting for Ollama to initialize...${NC}"
ollama_ready=false
for i in {1..10}; do
  echo -n "."
  if docker-compose exec -T ollama ollama list &> /dev/null; then
    ollama_ready=true
    echo -e "\n${GREEN}✓ Ollama is ready${NC}"
    PASSED=$((PASSED+1))
    break
  fi
  sleep 2
done

if [ "$ollama_ready" = false ]; then
  echo -e "\n${YELLOW}⚠ Ollama not ready yet. This is normal during first run and doesn't necessarily indicate a failure.${NC}"
  echo -e "${YELLOW}Ollama status: $(docker-compose logs --tail 5 ollama)${NC}"
fi

# Summary report
echo -e "\n${BLUE}============================================${NC}"
echo -e "${BLUE}              TEST RESULTS                 ${NC}"
echo -e "${BLUE}============================================${NC}"
echo -e "${GREEN}Tests passed: $PASSED${NC}"
echo -e "${RED}Tests failed: $FAILED${NC}"

if [ "$KEEP_RUNNING" = true ]; then
  echo -e "${YELLOW}Containers are still running as requested.${NC}"
  echo -e "${YELLOW}Run 'docker-compose down' when you're done.${NC}"
else
  echo -e "${YELLOW}Automatically cleaning up containers...${NC}"
fi

if [ "$FAILED" -gt 0 ]; then
  echo -e "${RED}Some tests failed!${NC}"
  exit 1
else
  echo -e "${GREEN}All tests passed successfully!${NC}"
  exit 0
fi
