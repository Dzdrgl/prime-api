.PHONY: migrate-up migrate-down migrate-status run build dev

migrate-up:
	go run ./cmd/migration up

migrate-down:
	go run ./cmd/migration down

migrate-status:
	go run ./cmd/migration status

run:
	go run ./cmd/api

build:
	go build ./cmd/api

dev: migrate-up run
