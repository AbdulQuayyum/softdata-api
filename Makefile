.PHONY: migrate-up migrate-down migrate-version migrate-force migrate-create sqlc-generate sqlc-compile database-check

SQLC_VERSION ?= v1.31.1
MIGRATE_VERSION ?= v4.17.1

SQLC := go run github.com/sqlc-dev/sqlc/cmd/sqlc@$(SQLC_VERSION)
MIGRATE := go run github.com/golang-migrate/migrate/v4/cmd/migrate@$(MIGRATE_VERSION)

migrate-up:
	@if [ -z "$(DATABASE_URL)" ]; then echo "DATABASE_URL is required"; exit 1; fi
	@$(MIGRATE) -path database/migrations -database "$(DATABASE_URL)" up

migrate-down:
	@if [ -z "$(DATABASE_URL)" ]; then echo "DATABASE_URL is required"; exit 1; fi
	@if [ -z "$(STEPS)" ]; then echo "STEPS is required"; exit 1; fi
	@if ! printf '%s' "$(STEPS)" | grep -Eq '^[1-9][0-9]*$$'; then echo "STEPS must be a positive integer"; exit 1; fi
	@$(MIGRATE) -path database/migrations -database "$(DATABASE_URL)" down "$(STEPS)"

migrate-version:
	@if [ -z "$(DATABASE_URL)" ]; then echo "DATABASE_URL is required"; exit 1; fi
	@$(MIGRATE) -path database/migrations -database "$(DATABASE_URL)" version

migrate-force:
	@if [ -z "$(DATABASE_URL)" ]; then echo "DATABASE_URL is required"; exit 1; fi
	@if [ -z "$(VERSION)" ]; then echo "VERSION is required"; exit 1; fi
	@if ! printf '%s' "$(VERSION)" | grep -Eq '^-?[0-9]+$$'; then echo "VERSION must be an integer migration version"; exit 1; fi
	@echo "Warning: force changes migration metadata and does not execute SQL."
	@$(MIGRATE) -path database/migrations -database "$(DATABASE_URL)" force "$(VERSION)"

migrate-create:
	@if [ -z "$(NAME)" ]; then echo "NAME is required"; exit 1; fi
	@$(MIGRATE) create -ext sql -dir database/migrations -seq "$(NAME)"

sqlc-generate:
	@$(SQLC) generate

sqlc-compile:
	@$(SQLC) compile

database-check:
	@$(MAKE) sqlc-generate
	@$(MAKE) sqlc-compile
	@go test ./...
	@go vet ./...
	@go build ./cmd/api
