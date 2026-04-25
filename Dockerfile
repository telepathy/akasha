# Multi-stage build for smaller image
FROM --platform=$TARGETPLATFORM golang:1.25-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build for target platform
RUN CGO_ENABLED=0 GOOS=linux GOARCH=$TARGETPLATFORM go build -mod=vendor -o /akasha cmd/server/main.go

# Runtime stage
FROM --platform=$TARGETPLATFORM alpine:latest

WORKDIR /app

# Install certificates for HTTPS
RUN apk --no-cache add ca-certificates tzdata

# Copy binary
COPY --from=builder /akasha .

EXPOSE 8080

CMD ["./akasha"]

# Build argument for multi-platform
ARG TARGETPLATFORM
ARG TARGETOS
ARG TARGETARCH