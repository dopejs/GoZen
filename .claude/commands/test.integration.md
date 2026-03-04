---
description: Build binary and run Go integration and e2e tests against a real daemon.
---

## User Input

```text
$ARGUMENTS
```

You **MUST** consider the user input before proceeding (if not empty).

## Goal

Build the GoZen binary and run integration and e2e tests that exercise the real daemon, proxy, and web server.

## Steps

1. **Build the binary**:
   ```bash
   cd web && pnpm install && pnpm build && cd ..
   go build -o bin/zen .
   ```

2. **Run integration tests** (test/integration/):
   ```bash
   go test -tags=integration -v -timeout 120s ./test/integration/...
   ```

3. **Run e2e tests** (tests/):
   ```bash
   go test -tags=integration -v -timeout 180s ./tests/...
   ```

4. **Report results**:
   - Show daemon startup/shutdown status for each test
   - List each test: PASS/FAIL with name and duration
   - If any test fails, show the full failure output
   - Report which mock providers were used and request counts
   - Show overall summary

## Notes

- Integration tests require a built binary at the project root or `bin/zen`
- Each test starts its own daemon on random ports — no port conflicts
- Tests use `//go:build integration` build tag — they won't run with plain `go test ./...`
- The e2e stress test (TestE2E_StressTest) may run longer; it sends 500+ requests
- If a test hangs, it likely means the daemon didn't start — check for port conflicts or build errors
