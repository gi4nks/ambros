# Multi-stage Dockerfile for Ambros

# Stage 1: Build
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache make git

# Copy source and build
COPY . .
RUN go mod download
RUN make build

# Stage 2: Final Image
FROM alpine:latest

WORKDIR /app

# Install runtime dependencies (like git/curl which Ambros uses)
RUN apk add --no-cache git curl ca-certificates

# Copy binary from builder
COPY --from=builder /app/ambros /usr/local/bin/ambros

# Create data directory
RUN mkdir -p /root/.ambros

EXPOSE 8080

# Run in server mode by default (dashboard)
ENTRYPOINT ["ambros"]
CMD ["server", "--host", "0.0.0.0"]
