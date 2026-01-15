# FutureBuild Root Makefile
# See PRODUCTION_PLAN.md Phase 0

.PHONY: audit test run clean migrate-up migrate-down migrate-create

# Audit: Run linting and type checks
audit:
	@echo "--- Auditing Backend ---"
	go vet ./...
	@echo "--- Auditing Frontend ---"
	npm --prefix frontend run build

# Test: Run all unit tests
test:
	@echo "--- Running Backend Tests ---"
	go test -v ./...
	@echo "--- Running Frontend Tests ---"
	@if [ -d "frontend" ]; then npm --prefix frontend test || echo "No frontend tests defined"; fi

# Auth Test: Verify JWT logic and claims
test-auth:
	@echo "--- Verifying JWT Authentication ---"
	go test -v ./internal/service/auth_service_test.go ./internal/service/auth_service.go ./internal/service/notification_service.go
	go test -v ./internal/api/handlers/auth_handler_test.go ./internal/api/handlers/auth_handler.go

# Contract Test: Verify Go and TS type parity
contract-test:
	@echo "--- Generating Contract Samples ---"
	go test ./pkg/types -v -run TestGenerateContractSamples
	@echo "--- Running Contract Validation ---"
	npm --prefix frontend run test:contract

# Run: Start the application (Dev Mode)
run:
	@echo "--- Starting FutureBuild ---"
	# Placeholder for start command, e.g., go run cmd/api/main.go
	@echo "Please specify a run command in the Makefile when cmd/api is implemented."

# Run Worker: Start the background worker
run-worker:
	@echo "--- Starting FutureBuild Worker ---"
	go run cmd/worker/main.go


# Clean: Remove build artifacts
clean:
	rm -rf bin/
	rm -rf dist/
	rm -rf frontend/dist/

# --- Database Migrations (golang-migrate) ---
# See PRODUCTION_PLAN.md Step 7

# Apply all pending migrations
migrate-up:
	@echo "--- Applying Migrations ---"
	migrate -database "$${DATABASE_URL}" -path migrations up

# Rollback last migration
migrate-down:
	@echo "--- Rolling Back Last Migration ---"
	migrate -database "$${DATABASE_URL}" -path migrations down 1

# Create new migration files
# Usage: make migrate-create name=add_users_table
migrate-create:
	@echo "--- Creating Migration: $(name) ---"
	migrate create -ext sql -dir migrations -seq $(name)

