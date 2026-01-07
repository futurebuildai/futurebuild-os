# Step 8: Multi-stage Dockerfile for FutureBuild
# See PRODUCTION_PLAN.md Section Phase 0

# --- Stage 1: Frontend Builder ---
FROM node:20-alpine AS frontend-builder
WORKDIR /app/frontend
COPY frontend/package*.json ./
RUN npm install
COPY frontend/ ./
RUN npm run build

# --- Stage 2: Backend Builder ---
FROM golang:1.22-alpine AS backend-builder
WORKDIR /app
COPY go.mod ./
# COPY go.sum ./
RUN go mod download
COPY . .
# Build the API binary
RUN go build -o /app/bin/api ./cmd/api/main.go
# Build the Worker binary (if applicable, placeholder for now)
# RUN go build -o /app/bin/worker ./cmd/worker/main.go

# --- Stage 3: Runtime ---
FROM alpine:latest
WORKDIR /app

# Install security updates and ca-certificates
RUN apk --no-cache add ca-certificates tzdata

# Copy backend binaries
COPY --from=backend-builder /app/bin/api /app/api
# COPY --from=backend-builder /app/bin/worker /app/worker

# Copy frontend static assets (for serving via Chi)
COPY --from=frontend-builder /app/frontend/dist /app/frontend/dist

# Expose API port
EXPOSE 8080

# Run the API
ENTRYPOINT ["/app/api"]
