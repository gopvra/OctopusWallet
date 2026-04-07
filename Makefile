.PHONY: build run-server run-worker test migrate clean web-install web-dev web-build

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
	psql "$(DATABASE_URL)" -f migrations/002_enterprise_features.sql
	psql "$(DATABASE_URL)" -f migrations/003_invoices_refunds_ledger.sql
	psql "$(DATABASE_URL)" -f migrations/004_payment_links_audit.sql

docker-up:
	docker-compose up -d

docker-down:
	docker-compose down

clean:
	rm -rf bin/ web/dist/

# Frontend
web-install:
	cd web && npm install

web-dev:
	cd web && npm run dev

web-build:
	cd web && npm run build
