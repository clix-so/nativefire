# Build stage
FROM golang:1.21-alpine AS builder

# Install necessary packages
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o nativefire .

# Runtime stage
FROM alpine:latest

# Install necessary packages for Firebase CLI and other tools
RUN apk add --no-cache \
    ca-certificates \
    nodejs \
    npm \
    git \
    curl \
    bash

# Install Firebase CLI
RUN npm install -g firebase-tools

# Create non-root user
RUN adduser -D -s /bin/bash nativefire

# Copy binary from builder stage
COPY --from=builder /app/nativefire /usr/local/bin/nativefire

# Make binary executable
RUN chmod +x /usr/local/bin/nativefire

# Switch to non-root user
USER nativefire

# Set working directory
WORKDIR /workspace

# Expose any ports if needed (none for this CLI tool)

# Default command
ENTRYPOINT ["nativefire"]
CMD ["--help"]