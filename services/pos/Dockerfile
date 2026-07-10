# Build Stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Install git and certificates
RUN apk add --no-cache git ca-certificates tzdata

# Copy go mod and sum files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o pos-server .

# Run Stage
FROM alpine:3.21

WORKDIR /app

# Copy timezone and certificates from builder
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy binary from builder
COPY --from=builder /app/pos-server .

# Set environment variables
ENV TZ=Asia/Jakarta
ENV PORT=8080

EXPOSE 8080

CMD ["/app/pos-server"]
