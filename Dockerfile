# Build stage
FROM golang:1.22-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /build

COPY go.mod go.sum* ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -trimpath -o drone-gemini-cli-plugin .

# Runtime stage - using Node.js slim for smaller image
FROM node:20-slim

# Install only essential packages
RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    git \
    && rm -rf /var/lib/apt/lists/* \
    && apt-get clean

# Install gemini CLI globally
RUN npm install -g @google/gemini-cli && npm cache clean --force

# Copy plugin binary
COPY --from=builder /build/drone-gemini-cli-plugin /bin/drone-gemini-cli-plugin

# Set working directory for drone
WORKDIR /drone/src

ENTRYPOINT ["/bin/drone-gemini-cli-plugin"]
