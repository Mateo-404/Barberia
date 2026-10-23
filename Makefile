# Tareas de todo el repo. Las del backend delegan en backend-go/Makefile.
# Usa '>' en vez de TAB para las recetas (GNU make >= 3.82).
.RECIPEPREFIX := >
SHELL := /bin/bash

# Defaults = PostgreSQL de docker-compose.yml.
DB_SUPER_URL ?= postgres://postgres:postgres@localhost:5432/barber?sslmode=disable
MIGRATE_DATABASE_URL ?= $(DB_SUPER_URL)

.PHONY: up down reset migrate test test-db test-frontend lint-sql fmt vet verify

up:
> docker compose up -d --wait

down:
> docker compose down -v

reset: down up migrate

migrate:
> cd backend-go && MIGRATE_DATABASE_URL="$(MIGRATE_DATABASE_URL)" go run ./cmd/migrate

test:
> $(MAKE) -C backend-go test

# Tests SQL de RLS y doble reserva (requieren BD migrada y el rol app_user de db/init).
test-db:
> psql "$(DB_SUPER_URL)" -v ON_ERROR_STOP=1 -q -f db/tests/rls_isolation.sql
> psql "$(DB_SUPER_URL)" -v ON_ERROR_STOP=1 -q -f db/tests/booking_exclusion.sql

test-frontend:
> cd frontend && bun run lint && bun run typecheck && bun run test && bun run build

lint-sql:
> @! grep -RInE --include='*.go' 'fmt\.Sprintf\("[^"]*(SELECT|INSERT|UPDATE|DELETE)[^"]*%' backend-go || (echo "lint-sql: SQL armado con Sprintf (usá parámetros)"; exit 1)

fmt:
> $(MAKE) -C backend-go fmt

vet:
> cd backend-go && go vet ./...

verify: fmt vet lint-sql migrate test test-db test-frontend
> @echo "verify: OK"
