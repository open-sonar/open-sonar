#!/bin/bash
# Update dependencies

echo "Updating Go dependencies..."

# Initialize the module and download dependencies
go mod tidy

# Explicitly get the new packages
go get github.com/PuerkitoBio/goquery
go get github.com/go-shiori/go-readability
go get github.com/gorilla/mux
go get github.com/joho/godotenv

echo "Dependencies updated successfully!"

# Check dependency status
echo "Current dependency status:"
go mod verify
