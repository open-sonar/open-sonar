#!/bin/bash
# Clean dependencies and rebuild module

echo "Cleaning Go module cache..."

# Clear the module cache
go clean -modcache

# Remove go.sum file
rm -f go.sum

# Initialize the modules
go mod tidy

# Get required packages
go get github.com/PuerkitoBio/goquery@v1.8.1
go get github.com/go-shiori/go-readability@v0.0.0-20231029095239-6b97d5aba789
go get github.com/gorilla/mux@v1.8.1
go get github.com/joho/godotenv@v1.5.1

# Verify the modules
go mod verify

echo "Dependencies cleaned and rebuilt successfully!"
