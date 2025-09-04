# Build stage
FROM golang:1.25-alpine AS builder

# Install build dependencies
RUN apk add --no-cache ca-certificates git tzdata

# Install sqlc
RUN go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy sqlc configuration and SQL files
COPY sqlc.yaml ./
COPY migrations/ ./migrations/
COPY internal/db/queries/ ./internal/db/queries/

# Generate sqlc code
RUN sqlc generate

# Copy the rest of the source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/server

# Final stage
FROM alpine:3.18

# Install ca-certificates for HTTPS requests and create non-root user
RUN apk --no-cache add ca-certificates tzdata && \
    adduser -D -s /bin/sh appuser

# Set working directory
WORKDIR /root/

# Copy the binary from builder stage
COPY --from=builder /app/main .

# Copy any static files if needed (migrations, etc.)
COPY --from=builder /app/migrations ./migrations

# Change ownership to appuser
RUN chown -R appuser:appuser /root/
USER appuser

# Expose port
EXPOSE 8080

# Run the application
CMD ["./main"]
