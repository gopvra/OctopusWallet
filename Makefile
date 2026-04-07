.PHONY: build run-server run-worker test migrate clean

build:
	go build -o bin/server ./cmd/server
	go build -o bin/worker ./cmd/worker

run-server:
	go run ./cmd/server

run-worker:
	go run ./cmd/worker

test:
	go test ./... -v

migrate:
	psql "$(DATABASE_URL)" -f migrations/001_init.sql

docker-up:
	docker-compose up -d

docker-down:
	docker-compose down

clean:
	rm -rf bin/
