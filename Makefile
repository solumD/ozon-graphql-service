include .env

LOCAL_BIN:=$(CURDIR)/bin

run:
	make install-deps
	docker compose up -d --build
	make wait-db
	make local-migration-up

run-in-memory:
	STORAGE_TYPE=in_memory PG_DSN= LOCAL_MIGRATION_DSN= docker compose up -d --build --no-deps app

stop:
	docker compose down

wait-db:
	docker compose exec -T pg sh -c 'until pg_isready -U "$$POSTGRES_USER" -d "$$POSTGRES_DB"; do sleep 1; done'

install-deps:
	GOBIN=$(LOCAL_BIN) go install github.com/pressly/goose/v3/cmd/goose@v3.14.0

test:
	go test ./internal/usecase/tests/... ./internal/delivery/graphql/tests/...

local-migration-status:
	${LOCAL_BIN}/goose -dir "${LOCAL_MIGRATION_DIR}" postgres "${LOCAL_MIGRATION_DSN}" status -v

local-migration-up:
	${LOCAL_BIN}/goose -dir "${LOCAL_MIGRATION_DIR}" postgres "${LOCAL_MIGRATION_DSN}" up -v

local-migration-down:
	${LOCAL_BIN}/goose -dir "${LOCAL_MIGRATION_DIR}" postgres "${LOCAL_MIGRATION_DSN}" down -v