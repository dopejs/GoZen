.PHONY: all build web test clean test-unit test-integration test-e2e test-web test-all test-race

all: build

# Build Web UI (required before Go build)
web:
	cd web && pnpm install && pnpm build

# Build Go binary (requires web to be built first)
build: web
	go build -o bin/zen .

# Run Go unit tests (no build tags, fast)
test-unit:
	go test -race -count=1 ./...

# Run Go integration tests (requires built binary)
test-integration: build
	go test -tags=integration -v -timeout 120s ./test/integration/...

# Run Go e2e tests (requires built binary)
test-e2e: build
	go test -tags=integration -v -timeout 180s ./tests/...

# Run frontend tests with coverage
test-web:
	cd web && pnpm install && pnpm test -- --run --coverage

# Run all test tiers in sequence
test-all: test-unit test-integration test-e2e test-web
	@echo "All test tiers complete."

# Legacy alias
test: test-unit

# Run tests with race detector (legacy alias)
test-race:
	go test -race -count=1 ./...

# Clean build artifacts
clean:
	rm -rf bin/ internal/web/dist/
