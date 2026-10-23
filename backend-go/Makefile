# Backend Go — comandos de desarrollo.
# Requiere Go >= 1.27 (ver go.mod). Las versiones de herramientas están fijadas.

GO      ?= go
SQLC    ?= sqlc
# `go run pkg@version` ignora el toolchain de go.mod: lo forzamos para que
# golangci-lint/govulncheck se compilen con la misma versión que el módulo.
GOTOOL  := GOTOOLCHAIN=$(shell $(GO) env GOVERSION) $(GO) run

.PHONY: all build run test test-integration lint vuln fmt tidy migrate-up sqlc docker-build clean

all: fmt lint test build

build:
	$(GO) build -o bin/api ./cmd/api

run: build
	./bin/api

test:
	$(GO) test -race ./...

# Requiere un PostgreSQL accesible (ver README "Tests de integración").
test-integration:
	$(GO) test -race -tags integration ./internal/integration/ -v

lint:
	$(GOTOOL) github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.15.0 run ./...

vuln:
	$(GOTOOL) golang.org/x/vuln/cmd/govulncheck@v1.9.0 ./...

fmt:
	gofmt -w $$(find . -name '*.go' -not -path './internal/platform/database/sqlc/*')

tidy:
	$(GO) mod tidy

# Aplica las migraciones embebidas (MIGRATE_DATABASE_URL o DATABASE_URL).
migrate-up:
	$(GO) run ./cmd/migrate

# Regenera el código de acceso a BD desde db/queries + db/migrations.
sqlc:
	$(SQLC) generate

docker-build:
	docker build -t barber-backend-go:local .

clean:
	rm -rf bin
