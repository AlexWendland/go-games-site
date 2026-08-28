.DEFAULT_GOAL := build

.PHONY: fmt vet lint test build run clean ui-build ui-dev

# Frontend
ui-build:
	cd ui && npm run build

ui-dev:
	cd ui && npm run dev

# Go linting
fmt:
	go fmt ./...

vet: fmt
	go vet ./...

lint: vet
	golangci-lint run ./...

# Tests
test:
	go test ./...

# Build (requires ui/dist to exist)
build: ui-build
	go build -o bin/server ./cmd/server

# Run the built binary
run: build
	./bin/server

# Clean build artifacts
clean:
	go clean
	rm -rf bin/ ui/dist/
