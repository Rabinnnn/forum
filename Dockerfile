# Stage 1: Build the Go application
FROM golang:1.23-alpine AS builder

# Install build dependencies
RUN apk add --no-cache gcc musl-dev


WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod go.sum ./
RUN go mod download

# Copy the entire project
COPY . .

# Enable CGO 
ENV CGO_ENABLED=1

# Build the application
RUN go build -o forum ./main.go

# Stage 2: Create a minimal image with the built binary
FROM alpine:latest

# Install runtime dependencies
RUN apk add --no-cache ca-certificates

WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /app/forum /app/
# Copy web templates and static files
COPY --from=builder /app/internal/web /app/internal/web

# Create uploads directory
RUN mkdir -p /app/internal/web/static/uploads

# Create templates directory
RUN mkdir -p /app/internal/web/static/templates

# Expose the port the app runs on
EXPOSE 8080

# Command to run the executable
CMD ["./forum"]
