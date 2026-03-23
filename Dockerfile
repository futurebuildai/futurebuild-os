# FutureBuild API — builds the HTTP server + frontend assets
# For the background worker, see Dockerfile.worker

# --- Stage 1: Frontend Builder ---
FROM node:20-alpine AS frontend-builder
WORKDIR /app/frontend
ARG VITE_CLERK_PUBLISHABLE_KEY
COPY frontend/package*.json ./
RUN npm ci
COPY frontend/ ./
RUN npm run build

# --- Stage 2: Backend Builder ---
FROM golang:1.24-alpine AS backend-builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/bin/api ./cmd/api/main.go

# Download golang-migrate for database migrations
RUN wget -q https://github.com/golang-migrate/migrate/releases/download/v4.17.0/migrate.linux-amd64.tar.gz \
    && tar -xzf migrate.linux-amd64.tar.gz \
    && mv migrate /app/bin/migrate \
    && rm migrate.linux-amd64.tar.gz

# --- Stage 3: Runtime ---
FROM alpine:3.20
WORKDIR /app

LABEL org.opencontainers.image.source=https://github.com/futurebuildai/futurebuild-repo
LABEL org.opencontainers.image.description="FutureBuild API"

RUN apk --no-cache add ca-certificates tzdata && \
    addgroup -S appgroup && adduser -S appuser -G appgroup

COPY --from=backend-builder /app/bin/api /app/api
COPY --from=backend-builder /app/bin/migrate /app/migrate
COPY migrations /app/migrations
COPY --from=frontend-builder /app/frontend/dist /app/frontend/dist

COPY scripts/entrypoint.sh /app/entrypoint.sh
RUN chmod +x /app/entrypoint.sh

RUN chown -R appuser:appgroup /app
USER appuser
EXPOSE 8080

ENTRYPOINT ["/app/entrypoint.sh"]
