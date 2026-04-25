# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -mod=vendor -o /akasha cmd/server/main.go

# Runtime stage
FROM alpine:latest

WORKDIR /app

# Install certificates for HTTPS
RUN apk --no-cache add ca-certificates tzdata

# Copy binary and necessary files
COPY --from=builder /akasha .
COPY config.yaml .
COPY templates/ ./templates/
COPY static/ ./static/

EXPOSE 8080

CMD ["./akasha"]