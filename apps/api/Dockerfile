# syntax=docker/dockerfile:1

# Build stage
FROM golang:1.24-bullseye AS buildx
WORKDIR /app

# Download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Install wire and the migrate CLI (mysql driver only, to keep the binary small)
RUN go install github.com/google/wire/cmd/wire@latest
RUN go install -tags 'mysql' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
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
ENV ROOT_DIR=/app

# Install ca-certificates and apply security updates
RUN apt-get update && apt-get upgrade -y && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*

# Run as a non-root user
RUN useradd --system --no-create-home --uid 10001 backend
USER backend

# Copy binary and migrate CLI from buildx
COPY --from=buildx /app/backend .
COPY --from=buildx /go/bin/migrate /usr/local/bin/migrate

# Copy migration files and script
COPY --chmod=755 db ./db

# Expose port (adjust if needed)
EXPOSE 8080

# Run the binary
CMD ["./backend"]