# Use a standard Go image - this will automatically match your host architecture
FROM golang:1.22

# Create workspace
WORKDIR /app

# Copy Go module files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN go build -o open-sonar ./cmd/server

# Set up runtime
EXPOSE 8080
CMD ["/app/open-sonar"]
