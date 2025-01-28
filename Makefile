# ==================================================================================== #
# HELPERS
# ==================================================================================== #

include .env
## help: print this help message
.PHONY: help
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'


# ==================================================================================== #
# QUALITY CONTROL
# ==================================================================================== #

## audit: run quality control checks
.PHONY: audit
audit: test
	go mod tidy -diff
	go mod verify
	test -z "$(shell gofmt -l .)" 
	go vet ./...
	go run honnef.co/go/tools/cmd/staticcheck@latest -checks=all,-ST1000,-U1000 ./...
	go run golang.org/x/vuln/cmd/govulncheck@latest ./...

## test: run all tests
.PHONY: test
test:
	go test -v -race -buildvcs ./...

## test/cover: run all tests and display coverage
.PHONY: test/cover
test/cover:
	go test -v -race -buildvcs -coverprofile=/tmp/coverage.out ./...
	go tool cover -html=/tmp/coverage.out

# ==================================================================================== #
# DEVELOPMENT
# ==================================================================================== #

## tidy: tidy modfiles and format .go files
.PHONY: tidy
tidy:
	go mod tidy -v
	go fmt ./...

## build: build the cmd/api application
.PHONY: build
build:
	go build -o=/tmp/bin/api ./cmd/api
	
## run: run the cmd/api application
.PHONY: run
run: build
	/tmp/bin/api

## run/live: run the application with reloading on file changes
.PHONY: run/live
run/live:
	go run github.com/cosmtrek/air@v1.43.0 \
		--build.cmd "make build" --build.bin "/tmp/bin/api" --build.delay "100" \
		--build.exclude_dir "" \
		--build.include_ext "go, tpl, tmpl, html, css, scss, js, ts, sql, jpeg, jpg, gif, png, bmp, svg, webp, ico" \
		--misc.clean_on_exit "true" 

# Create DB container
.PHONY: docker/run
docker/run:
	@if docker compose up 2>/dev/null; then \
		: ; \
	else \
		echo "Falling back to Docker Compose V1"; \
		docker-compose up; \
	fi

# Shutdown DB container
.PHONY: docker/down
docker/down:
	@if docker compose down 2>/dev/null; then \
		: ; \
	else \
		echo "Falling back to Docker Compose V1"; \
		docker-compose down; \
	fi

# ==================================================================================== #
# SQL MIGRATIONS
# ==================================================================================== #

## migrations/up: apply all up database migrations
.PHONY: migrations/up
migrations/up:
	@cd sql/schemas && goose postgres postgres://${DB_USERNAME}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_DATABASE} up

## migrations/up-to version=$1: migrate up to a specific version number
.PHONY: migrations/up-to
migrations/up-to:
	@cd sql/schemas && goose postgres postgres://${DB_USERNAME}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_DATABASE} up-to ${version}

## migrations/up-by-one: migrate up by one version
.PHONY: migrations/up-by-one
migrations/up-by-one:
	@cd sql/schemas && goose postgres postgres://${DB_USERNAME}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_DATABASE} up-by-one

## migrations/down: apply all down database migrations
.PHONY: migrations/down
migrations/down:
	@cd sql/schemas && goose postgres postgres://${DB_USERNAME}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_DATABASE} down-to 0

## migrations/down-to version=$1: migrate down to a specific version number
.PHONY: migrations/down-to
migrations/down-to:
	@cd sql/schemas && goose postgres postgres://${DB_USERNAME}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_DATABASE} down-to ${version}

## migrations/down-by-one: migrate down by one version
.PHONY: migrations/down-by-one
migrations/down-by-one:
	@cd sql/schemas && goose postgres postgres://${DB_USERNAME}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_DATABASE} down

## migrations/status: show the status of the migrations
.PHONY: migrations/status
migrations/status:
	@cd sql/schemas && goose postgres postgres://${DB_USERNAME}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_DATABASE} status

## sqlc/generate: generate sqlc files
.PHONY: sqlc/generate
sqlc/generate:
	sqlc generate

## swag/build: generate swagger documentation
.PHONY: swag/generate
swag/generate:
	@echo "Generating Swagger documentation..."
	@cd ./cmd/api && swag init --parseDependency --parseInternal || { echo "Swagger generation failed"; exit 1; }
	@echo "Swagger documentation generated successfully."