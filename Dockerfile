# Dockerfile for developers-conferences-agenda-mcp
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy go.mod and go.sum first for caching dependencies
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN go build -ldflags="-s -w" -o server .

FROM alpine:latest

WORKDIR /app

# Copy the built binary from the builder stage
COPY --from=builder /app/server ./

# Set entrypoint for the MCP server
ENTRYPOINT ["./server"]
