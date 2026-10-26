.PHONY: run migrate-up migrate-down generate test lint fmt help

BINARY_NAME=intelligate-api
CMD_PATH=./cmd/api
MIGRATIONS_DIR=./migrations
SQLC_CONFIG=./sqlc.yaml

DATABASE_URL ?= postgres://postgres:postgres@localhost:5433/intelligate?sslmode=disable

help:
	@echo "Available targets:"
	@echo "  run          - Run the API server"
	@echo "  migrate-up   - Run all pending migrations"
	@echo "  migrate-down - Rollback last migration"
	@echo "  generate     - Generate sqlc code"
	@echo "  test         - Run tests"
	@echo "  lint         - Run golangci-lint"
	@echo "  fmt          - Format Go code"

run:
	@go run ./cmd/api

migrate-up:
	@goose -dir $(MIGRATIONS_DIR) postgres "$(DATABASE_URL)" up

migrate-down:
	@goose -dir $(MIGRATIONS_DIR) postgres "$(DATABASE_URL)" down

generate:
	@sqlc generate -f $(SQLC_CONFIG)

test:
	@go test -race ./...

lint:
	@golangci-lint run ./...

fmt:
	@gofmt -w .