#!/bin/bash

# Base directory for services
SERVICES_DIR="services"

# Loop through each service directory
for SERVICE in "$SERVICES_DIR"/*; do
  if [ -d "$SERVICE" ]; then
    SERVICE_NAME=$(basename "$SERVICE")
    echo "Processing $SERVICE_NAME..."

    # Determine migration path
    MIGRATION_COPY_CMD=""
    if [ -d "$SERVICE/migrations" ]; then
      MIGRATION_COPY_CMD="COPY migrations ./migrations"
    elif [ -d "$SERVICE/sql/migrations" ]; then
        MIGRATION_COPY_CMD="COPY sql/migrations ./sql/migrations"
    fi

    # Create Dockerfile content
    cat > "$SERVICE/Dockerfile" <<EOF
# Build Stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Copy Go module files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
# CGO_ENABLED=0 creates a statically linked binary
RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/api

# Run Stage
FROM alpine:3.19

WORKDIR /app

# Install certificates for HTTPS
RUN apk --no-cache add ca-certificates

# Copy binary from builder
COPY --from=builder /app/main .

# Copy migration files (if they exist)
$MIGRATION_COPY_CMD

# Expose port
EXPOSE 8080

# Command to run
CMD ["./main"]
EOF

    echo "Created Dockerfile for $SERVICE_NAME"
  fi
done
