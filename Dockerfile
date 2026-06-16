# Builder stage
FROM golang:1.25-bookworm AS builder

WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the entire source code
COPY . .

# Build the Go app with CGO enabled (required for go-sqlite3)
RUN CGO_ENABLED=1 GOOS=linux go build -o server ./cmd/server

# Final stage
FROM debian:bookworm-slim

WORKDIR /app

# Copy the compiled binary from the builder stage
COPY --from=builder /app/server .

# Install necessary libraries (ca-certificates for HTTPS/APIs, sqlite3 for DB)
RUN apt-get update && \
    apt-get install -y ca-certificates sqlite3 && \
    rm -rf /var/lib/apt/lists/*

# Expose port 7860 (Hugging Face Spaces default port)
EXPOSE 7860

# Run the server
CMD ["./server"]
