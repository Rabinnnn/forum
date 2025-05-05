# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod ./
# Copy the entire project
COPY . .

# Build the application
RUN go build -o forum ./cmd

# Final stage
FROM alpine:latest

WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /app/forum /app/
# Copy web templates and static files
COPY --from=builder /app/internal/web /app/internal/web

# Create uploads directory
RUN mkdir -p /app/internal/web/static/uploads

# Expose the port the app runs on
EXPOSE 8080

# Command to run the executable
CMD ["./forum"]
