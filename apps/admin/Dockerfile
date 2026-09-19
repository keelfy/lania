# Build from the repository root; the admin binary shares domain code with the API.
FROM golang:1.27-alpine AS build
WORKDIR /src
COPY apps/api/go.mod apps/api/go.sum ./
RUN go mod download
COPY apps/api/ ./
RUN CGO_ENABLED=0 go build -trimpath -ldflags='-s -w' -o /admin ./cmd/admin

FROM alpine:3.22
RUN apk add --no-cache ca-certificates && adduser -D -u 10001 admin
USER admin
COPY --from=build /admin /admin
EXPOSE 8081
ENTRYPOINT ["/admin"]
