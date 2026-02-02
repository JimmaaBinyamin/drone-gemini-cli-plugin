# Build stage
FROM golang:1.22-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /build

# Copy dependency files first for better caching
COPY go.mod go.sum* ./
RUN go mod download

# Copy source code
COPY . .

# Build the plugin binary with optimizations
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -trimpath -o drone-gemini-cli-plugin .

# Runtime stage - using Node.js Alpine for gemini CLI
FROM node:20-alpine

# Install Gemini CLI globally (official Google package)
RUN npm install -g @google/gemini-cli

# Install required tools
RUN apk add --no-cache ca-certificates tzdata git bash curl

# Add non-root user for security
RUN addgroup -S appuser && adduser -S appuser -G appuser

WORKDIR /app

# Copy the compiled plugin binary with proper ownership
COPY --from=builder --chown=appuser:appuser /build/drone-gemini-cli-plugin /bin/drone-gemini-cli-plugin

# Switch to non-root user
USER appuser

# Set the entrypoint
ENTRYPOINT ["/bin/drone-gemini-cli-plugin"]
