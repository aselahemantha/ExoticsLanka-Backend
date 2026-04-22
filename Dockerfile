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

# Build the modular monolith
RUN CGO_ENABLED=0 GOOS=linux go build -o monolith ./cmd/api/main.go

# Run Stage
FROM alpine:3.19

WORKDIR /app

# Install certificates for HTTPS
RUN apk --no-cache add ca-certificates

# Copy binary from builder
COPY --from=builder /app/monolith .

# Copy migration files (Monolith needs all migrations)
COPY migrations ./migrations

# Expose port
EXPOSE 8080

# Environment variables
ENV PORT=8080

# Command to run
CMD ["./monolith"]
