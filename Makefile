# Convenience targets for local development.
DB_URL ?= postgres://overvu:overvu@localhost:5432/overvu?sslmode=disable

.PHONY: db migrate backend frontend test test-legacy generate up down

db:            ## start PostgreSQL
	docker compose up -d postgres

migrate: db    ## apply Flyway migrations
	docker compose run --rm flyway

backend:       ## run the Go API on :8080
	cd backend && APP_ENV=dev DATABASE_URL="$(DB_URL)" go run ./cmd/server

frontend:      ## run the Vue dev server on :5173
	cd frontend && npm run dev

test:          ## unit tests + API contract suite against the local database
	cd backend && TEST_DATABASE_URL="$(DB_URL)" go test -count=1 ./...
	cd frontend && npm run lint && npm run build

test-legacy:   ## run the contract suite against a server on :8080 (e.g. the old Spring backend)
	cd backend && CONTRACT_BASE_URL=http://localhost:8080 CONTRACT_LEGACY=1 go test -count=1 ./contracttest/

generate:      ## regenerate the sqlc database code
	cd backend && sqlc generate

up:            ## full stack in Docker
	docker compose up --build

down:
	docker compose down
