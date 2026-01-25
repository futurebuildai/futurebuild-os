# --- Stage 1: Frontend Builder ---
FROM node:20-alpine AS frontend-builder
WORKDIR /app/frontend
COPY frontend/package*.json ./
RUN npm install
COPY frontend/ ./
RUN npm run build

# --- Stage 2: Backend Builder ---
FROM golang:1.24-alpine AS backend-builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
# Build the API binary
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/api ./cmd/api/main.go

# --- Stage 3: Runtime ---
FROM alpine:latest
WORKDIR /app

# Add metadata
LABEL org.opencontainers.image.source=https://github.com/futurebuildai/futurebuild-repo
LABEL org.opencontainers.image.description="FutureBuild API & Frontend"
LABEL org.opencontainers.image.licenses=UNLICENSED

# Install security updates and ca-certificates
RUN apk --no-cache add ca-certificates tzdata && \
    addgroup -S appgroup && adduser -S appuser -G appgroup

# Copy backend binaries
COPY --from=backend-builder /app/bin/api /app/api

# Copy frontend static assets (for serving via Chi)
COPY --from=frontend-builder /app/frontend/dist /app/frontend/dist

# Set ownership to non-root user
RUN chown -R appuser:appgroup /app

# Switch to non-root user
USER appuser

# Expose API port
EXPOSE 8080

# Run the API
ENTRYPOINT ["/app/api"]
