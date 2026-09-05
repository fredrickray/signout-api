.PHONY: run tidy migrate-up docker-up docker-down

run:
	go run ./cmd/api

tidy:
	go mod tidy

docker-up:
	docker compose up -d

docker-down:
	docker compose down

test:
	go test ./...
