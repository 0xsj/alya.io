# backend/Dockerfile
# Stage 1: Build the application
FROM golang:1.24-alpine AS builder

# Set working directory
WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Copy go.mod and go.sum files
COPY go.mod ./
# COPY go.sum ./
# RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o alya ./cmd/server/main.go

# Stage 2: Create a minimal runtime image
FROM alpine:latest

# Add CA certificates for HTTPS
RUN apk --no-cache add ca-certificates tzdata

# Create non-root user for security
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

# Set working directory
WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /app/alya .

# Create config directory
RUN mkdir -p config

# Use non-root user
USER appuser

# Expose the application port
EXPOSE 8080

# Run the application
CMD ["./alya"]