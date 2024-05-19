.PHONY: run-db stop-db migrate-db build run test

# Database environment variables
DB_USER ?= user
DB_PASSWORD ?= password
DB_NAME ?= itsdb
DB_HOST ?= localhost
DB_PORT ?= 5432

# Docker command to spin up the PostgreSQL database
run-db:
	docker run --name its-db \
	-e POSTGRES_USER=$(DB_USER) \
	-e POSTGRES_PASSWORD=$(DB_PASSWORD) \
	-e POSTGRES_DB=$(DB_NAME) \
	-p $(DB_PORT):5432 \
	--cpus="0.15" \
	--memory="150m" \
	-d postgres:alpine

# Docker command to remove the PostgreSQL database
remove-db:
	docker rm -f its-db

# Command to run the SQL script to instantiate the tables
migrate-db:
	docker run --rm \
	--network host \
	-e PGPASSWORD=$(DB_PASSWORD) \
	-v $(shell pwd)/schema.sql:/schema.sql \
	postgres:alpine \
	psql -h $(DB_HOST) -U $(DB_USER) -d $(DB_NAME) -f /schema.sql

# Command to run the Go application
run:
	DB_USER=$(DB_USER) DB_PASSWORD=$(DB_PASSWORD) DB_NAME=$(DB_NAME) DB_HOST=$(DB_HOST) DB_PORT=$(DB_PORT) go run cmd/main.go

# Command to run all tests (unit and integration)
test:
	@echo "Running tests with coverage..."
	@go test -race -coverprofile=coverage.out ./...
	@go tool cover -func=coverage.out

