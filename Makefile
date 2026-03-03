.PHONY: all build web test clean

all: build

# Build Web UI (required before Go build)
web:
	cd web && npm ci && npm run build

# Build Go binary (requires web to be built first)
build: web
	go build -o bin/zen .

# Run all tests (requires web to be built first)
test: web
	go test ./...

# Run tests with race detector
test-race: web
	go test -race -count=1 ./...

# Clean build artifacts
clean:
	rm -rf bin/ internal/web/dist/
