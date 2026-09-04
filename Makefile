.DEFAULT_GOAL := build

.PHONY: fmt vet lint test build run clean ui-build ui-dev generate run-ui-backend run-ui-frontend docker-build

# Frontend
ui-build:
	cd ui && npm run build

#   Run backend only, expecting frontend on :5173 (ui-dev mode)
run-ui-backend: build
	GAMES_PRODUCTION=false \
	GAMES_BASE_URL=http://localhost:8080 \
	GAMES_ALLOWED_ORIGINS=http://localhost:5173 \
	./bin/server

#   Run frontend dev server pointing at backend on :8080
run-ui-frontend:
	cd ui && VITE_BACKEND_URL=$${BACKEND_URL:-http://localhost:8080} npm run dev

# Code generation
generate: ui-build
	sqlc generate
	oapi-codegen --config oapi-codegen.yaml internal/api/openapi.yaml
	cd ui && npx openapi-typescript ../internal/api/openapi.yaml -o src/api/types.ts

# Go linting
lint:
	fd -e go --exec goimports -w
	go fmt ./...
	go vet ./...
	go mod tidy
	golangci-lint run ./...

# Tests
test:
	go test ./...

# Build (requires ui/dist to exist)
build: ui-build generate
	go build -o bin/server ./cmd/server


# Run everything off the same server (production-like)
run: build
	GAMES_PRODUCTION=false \
	GAMES_BASE_URL=http://localhost:8080 \
	GAMES_ALLOWED_ORIGINS=http://localhost:8080 \
	./bin/server

# Build and test the docker image locally
docker-build:
	docker build -t go-games-site:local .

# Clean build artifacts
clean:
	go clean
	rm -rf bin/ ui/dist/
	mkdir ui/dist
	touch ui/dist/.gitkeep
	rm -f games.db*
