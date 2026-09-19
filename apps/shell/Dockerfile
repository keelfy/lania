# syntax=docker/dockerfile:1

# Build stage
FROM golang:1.27-alpine AS buildx
WORKDIR /app

# Download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY . .

# Build binary with limited memory usage
RUN GOOS=linux CGO_ENABLED=0 GOARCH=amd64 GOMAXPROCS=2 go build -ldflags='-s' -o shell ./cmd

# Runtime stage
FROM debian:bookworm-slim
WORKDIR /app

# Install ca-certificates and apply security updates
RUN apt-get update && apt-get upgrade -y && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*

# Run as a non-root user
RUN useradd --system --no-create-home --uid 10001 shell
USER shell

COPY --from=buildx /app/shell .

EXPOSE 9090

CMD ["./shell"]
