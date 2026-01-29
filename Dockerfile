# Build stage for development tools
FROM golang:1.25.6-bookworm AS dev-tools

# Install build dependencies
RUN apt-get update && \
    apt-get install -y --no-install-recommends \
    git \
    make \
    python3-pip \
    python3-setuptools \
    && rm -rf /var/lib/apt/lists/* && \
    pip3 install --no-cache-dir --break-system-packages pre-commit

# Install Go tools
RUN go install mvdan.cc/gofumpt@latest && \
    go install github.com/kisielk/errcheck@latest && \
    go install github.com/go-delve/delve/cmd/dlv@latest && \
    go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest && \
    go install github.com/securego/gosec/v2/cmd/gosec@latest

# Builder stage
FROM golang:1.25.6-bookworm AS builder

WORKDIR /app

# Copy go mod and sum files first for better caching
COPY go.mod go.sum ./

# Download all dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /go/bin/vulnerable-target ./cmd/vt

# Final stage for application
FROM alpine:3.21

# Install runtime dependencies
RUN apk --no-cache add \
    docker-cli \
    git \
    bash \
    curl \
    jq \
    yamllint \
    && rm -rf /var/cache/apk/*

# Copy Go tools from dev-tools stage
COPY --from=dev-tools /go/bin/* /usr/local/bin/

WORKDIR /app

# Copy the binary from builder
COPY --from=builder /go/bin/vulnerable-target .

# Set the entrypoint to bash by default
ENTRYPOINT ["/bin/bash"]
