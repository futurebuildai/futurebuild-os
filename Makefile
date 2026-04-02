.PHONY: build build-server test lint db-up db-down clean

build: build-server

build-server:
	go build -o bin/server ./cmd/server

test:
	go test ./... -v -count=1

lint:
	golangci-lint run ./...

db-up:
	docker compose up -d

db-down:
	docker compose down

clean:
	rm -rf bin/

audit: lint test
	@echo "Audit passed."
