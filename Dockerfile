# Build stage for Go binary
FROM golang:1.22-alpine AS go-builder

RUN apk add --no-cache git ca-certificates

WORKDIR /build

COPY go.mod go.sum* ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -trimpath -o drone-gemini-cli-plugin .

# Build stage for Node.js (install gemini-cli with build tools)
FROM node:20-alpine AS node-builder

# Install build dependencies for native modules
RUN apk add --no-cache python3 make g++ git

# Install gemini-cli
RUN npm install -g @google/gemini-cli --omit=dev

# Clean up aggressively
RUN npm cache clean --force \
    && rm -rf /root/.npm /tmp/* \
    && find /usr/local/lib/node_modules -name "*.md" -delete 2>/dev/null || true \
    && find /usr/local/lib/node_modules -name "*.map" -delete 2>/dev/null || true \
    && find /usr/local/lib/node_modules -name "LICENSE*" -delete 2>/dev/null || true \
    && find /usr/local/lib/node_modules -name "CHANGELOG*" -delete 2>/dev/null || true \
    && find /usr/local/lib/node_modules -name "HISTORY*" -delete 2>/dev/null || true \
    && find /usr/local/lib/node_modules -name "README*" -delete 2>/dev/null || true \
    && find /usr/local/lib/node_modules -type d -name "test" -exec rm -rf {} + 2>/dev/null || true \
    && find /usr/local/lib/node_modules -type d -name "tests" -exec rm -rf {} + 2>/dev/null || true \
    && find /usr/local/lib/node_modules -type d -name "docs" -exec rm -rf {} + 2>/dev/null || true \
    && find /usr/local/lib/node_modules -type d -name "doc" -exec rm -rf {} + 2>/dev/null || true \
    && find /usr/local/lib/node_modules -type d -name "example" -exec rm -rf {} + 2>/dev/null || true \
    && find /usr/local/lib/node_modules -type d -name "examples" -exec rm -rf {} + 2>/dev/null || true \
    && find /usr/local/lib/node_modules -type d -name ".github" -exec rm -rf {} + 2>/dev/null || true

# Runtime stage - minimal Alpine (no build tools needed)
FROM node:20-alpine

# Install only essential runtime dependencies
RUN apk add --no-cache git ca-certificates \
    && rm -rf /var/cache/apk/*

# Copy cleaned node_modules
COPY --from=node-builder /usr/local/lib/node_modules /usr/local/lib/node_modules

# Create wrapper script for gemini command
RUN printf '#!/bin/sh\nexec node /usr/local/lib/node_modules/@google/gemini-cli/dist/index.js "$@"\n' > /usr/local/bin/gemini \
    && chmod +x /usr/local/bin/gemini

# Copy Go plugin binary
COPY --from=go-builder /build/drone-gemini-cli-plugin /bin/drone-gemini-cli-plugin

WORKDIR /drone/src

ENTRYPOINT ["/bin/drone-gemini-cli-plugin"]
