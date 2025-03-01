#!/bin/bash
# Install langchain dependencies and update go.sum

set -e  # Exit on error

echo "Downloading langchaingo dependencies..."

# Clear any outdated go.sum entries to prevent checksum mismatches
rm -f go.sum

# Tidy up modules
go mod tidy

# Explicitly get langchaingo
go get github.com/tmc/langchaingo@v0.1.13

# Verify installations
go mod verify

echo "Successfully updated langchaingo dependencies!"
