# ==========================================
# Stage 1: Build Web UI Frontend (Vue 3 + Vite)
# ==========================================
FROM node:22-alpine AS web-builder
WORKDIR /app/web

COPY web/package*.json ./
RUN npm install

COPY web/ ./
RUN npm run build

# ==========================================
# Stage 2: Build Go Static Binary
# ==========================================
FROM golang:alpine AS go-builder
WORKDIR /app

RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

# Copy backend source
COPY internal/ ./internal/
COPY cmd/ ./cmd/

# Copy built frontend assets to web/dist for go:embed
COPY --from=web-builder /app/web/dist ./web/dist
COPY web/embed.go ./web/embed.go

# Compile pure-Go static binary (no CGO required with modernc.org/sqlite)
ENV CGO_ENABLED=0
ENV GOOS=linux
RUN go build -ldflags="-s -w" -o /app/bin/syncwave ./cmd/server/main.go

# ==========================================
# Stage 3: Minimal Production Runtime
# ==========================================
FROM alpine:3.21 AS runner

LABEL maintainer="SyncWave Team"
LABEL description="Self-hosted YouTube Music Sync & Streaming Daemon"

# Install yt-dlp, ffmpeg, ca-certificates, and utilities
RUN apk add --no-cache \
    yt-dlp \
    ffmpeg \
    ca-certificates \
    tzdata \
    dumb-init

WORKDIR /app

# Copy compiled single static binary
COPY --from=go-builder /app/bin/syncwave /usr/local/bin/syncwave

# Create data and config directories
RUN mkdir -p /data/music /data/covers

ENV PORT=8080
ENV HOST=0.0.0.0
ENV DATA_DIR=/data
ENV APP_ENV=production

EXPOSE 8080

VOLUME ["/data"]

ENTRYPOINT ["/usr/bin/dumb-init", "--"]
CMD ["/usr/local/bin/syncwave"]

