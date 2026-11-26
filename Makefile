.PHONY: run up down reset migrate-up migrate-down test tidy vet

run:
	go run ./cmd/api

up:
	docker compose up -d

down:
	docker compose down

reset:
	docker compose down -v
	docker compose up -d

migrate-up:
	migrate -path migrations \
		-database "postgres://postgres:postgres@localhost:5432/webhook?sslmode=disable" up

migrate-down:
	migrate -path migrations \
		-database "postgres://postgres:postgres@localhost:5432/webhook?sslmode=disable" down 1

test:
	go test ./... -race -count=1

vet:
	go vet ./...

tidy:
	go mod tidy