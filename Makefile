DB_URL ?= postgres://postgres:postgres@localhost:5432/booking?sslmode=disable

.PHONY: up down migrate-naive migrate migrate-down seed gen api test load

up:
	docker compose up -d

down:
	docker compose down

migrate-naive: ## only the tables, no unique index (for the naive-booking demo)
	migrate -path migrations -database "$(DB_URL)" up 1

migrate: ## everything, including the unique index
	migrate -path migrations -database "$(DB_URL)" up

migrate-down:
	migrate -path migrations -database "$(DB_URL)" down 1

seed:
	docker compose exec -T postgres psql -U postgres -d booking < scripts/seed.sql

gen:
	go run github.com/99designs/gqlgen generate

api:
	go run ./cmd/api

test:
	go test -race ./...

load:
	k6 run loadtest/stampede.js
