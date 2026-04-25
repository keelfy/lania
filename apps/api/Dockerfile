# syntax=docker/dockerfile:1

# Build stage
FROM golang:1.24-bullseye AS buildx
WORKDIR /app

# Download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Install wire
RUN go install github.com/google/wire/cmd/wire@latest
ENV PATH="/go/bin:${PATH}"

# Copy source
COPY . .

# Generate wire file in the correct directory, but only if it doesn't exist
RUN if [ ! -f cmd/wire_gen.go ]; then cd cmd && wire; fi

# Build binary with limited memory usage
RUN GOOS=linux CGO_ENABLED=0 GOARCH=amd64 GOMAXPROCS=2 go build -ldflags='-s' -o backend ./cmd

# Runtime stage
FROM debian:bookworm-slim
WORKDIR /app

# Install ca-certificates
RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*

# Copy binary from buildx
COPY --from=buildx /app/backend .

# Expose port (adjust if needed)
EXPOSE 8080

# Run the binary
CMD ["./backend"]