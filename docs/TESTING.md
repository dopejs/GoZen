# Testing Strategy

This document describes the testing strategy for GoZen, with a focus on maintaining high reliability for the daemon proxy (P0 component) while avoiding CI instability.

## Test Layers

### Layer 1: Unit & Component Tests (Blocking CI)

**Location**: `internal/*/` test files
**Run on**: Every PR, every push to main
**Characteristics**:
- Fast (< 5 minutes total)
- Stable (no flakiness)
- No external dependencies
- Race detection enabled

**Coverage Requirements**:
- `internal/config`: ≥80%
- `internal/proxy`: ≥80%
- `internal/proxy/transform`: ≥80%
- `internal/web`: ≥80%
- `internal/bot`: ≥80%
- `internal/daemon`: ≥50%
- `internal/update`: ≥50%
- `internal/sync`: ≥50%

**Run locally**:
```bash
go test -race -short ./...
```

### Layer 2: Integration Tests (Blocking CI)

**Location**: `tests/integration/`
**Run on**: Every PR, every push to main
**Characteristics**:
- Moderate speed (< 3 minutes)
- Stable and controlled
- Minimal external dependencies
- Skips flaky tests in CI (via `CI=true` env var)

**What's tested**:
- Proxy failover and load balancing
- Health checking and metrics
- Configuration management
- Provider disable/enable
- Connection pool cleanup
- Timeout handling

**Run locally**:
```bash
go test -race ./tests/integration/...
```

**Run in CI mode** (skips flaky tests):
```bash
SKIP_FLAKY_TESTS=true go test -race ./tests/integration/...
```

### Layer 3: E2E Tests (Non-blocking)

**Location**: `tests/integration/daemon_*_test.go`
**Run on**:
- Manual trigger (workflow_dispatch)
- Nightly (2 AM UTC)
- After merge to main (workflow_run)

**Characteristics**:
- Slow (up to 10 minutes)
- May be flaky due to:
  - Process spawning and signals
  - Port binding races
  - Timing-dependent behavior
  - GitHub runner environment variations
- Tests real daemon binary behavior

**What's tested**:
- Daemon auto-restart after crash
- Signal handling (SIGTERM, SIGINT)
- Port takeover logic
- PID file management
- Full daemon lifecycle

**Current limitations** (documented in test comments):
- `TestAutoRestart`: Only tests daemon startup, not actual restart after crash
- `TestDaemonAutoRestart`: Tests fatal error/signal handling, not crash recovery
- `TestDaemonCrashRecovery`: Only tests error classification, not real crash injection

**Future work**:
- Real crash injection and recovery verification
- Restart loop with exponential backoff validation
- Max restart limit enforcement testing

These limitations are acceptable because:
- Core daemon stability is validated by Layer 1 & 2 tests
- Crash detection logic (IsFatalError) is tested
- Port takeover and process management are tested
- Real crash recovery requires complex process injection

**Run locally**:
```bash
go test -v -timeout 600s ./tests/...
```

**Run specific test**:
```bash
go test -v -run TestDaemonAutoRestart ./tests/integration/
```

## CI Workflows

### Main CI (`.github/workflows/ci.yml`)

**Jobs**:
1. **Unit Tests** - Fast unit tests with coverage checks
2. **Integration Tests** - Stable integration tests (flaky tests skipped via `CI=true`)
3. **Web UI Tests** - Frontend tests with coverage
4. **Website Build** - Documentation build verification

**Status**: All jobs are **required checks** for PR merge

### E2E Workflow (`.github/workflows/e2e.yml`)

**Triggers**:
- Manual: Go to Actions → E2E Tests → Run workflow
- Nightly: Runs at 2 AM UTC daily
- Post-merge: Runs after CI passes on main branch

**Status**: **Not a required check** - failures don't block PRs

**Notifications**:
- Failed E2E runs after merge will comment on the merged PR
- Check Actions tab for detailed results

## Test Annotations

Tests that are flaky in CI should check the `SKIP_FLAKY_TESTS` environment variable:

```go
func TestDaemonAutoRestart(t *testing.T) {
    // Skip in CI environment - these tests are flaky on GitHub runners
    if os.Getenv("SKIP_FLAKY_TESTS") == "true" {
        t.Skip("skipping daemon auto-restart test (SKIP_FLAKY_TESTS=true)")
    }

    // Test implementation...
}
```

## Running Tests Locally

### Quick validation (before commit):
```bash
go test -short ./...
```

### Full test suite (before PR):
```bash
go test -race ./...
```

### Integration tests only:
```bash
go test -race ./tests/integration/...
```

### E2E tests (may be slow):
```bash
go test -v -timeout 600s ./tests/...
```

### Specific package with coverage:
```bash
go test -cover ./internal/proxy/
```

## Adding New Tests

### Unit/Component Test
- Add to `internal/<package>/<file>_test.go`
- Must be fast and stable
- No external dependencies
- Will run on every PR

### Integration Test
- Add to `tests/integration/<feature>_test.go`
- Should be stable and controlled
- If potentially flaky, add skip check:
  ```go
  if os.Getenv("SKIP_FLAKY_TESTS") == "true" {
      t.Skip("skipping in CI environment (SKIP_FLAKY_TESTS=true)")
  }
  ```

### E2E Test
- Add to `tests/integration/daemon_*_test.go` or similar
- Can be slow and timing-dependent
- Should skip in main CI:
  ```go
  if os.Getenv("SKIP_FLAKY_TESTS") == "true" {
      t.Skip("skipping E2E test (SKIP_FLAKY_TESTS=true)")
  }
  ```

## Philosophy

For the daemon proxy (P0 component):

1. **Confidence comes from stable tests**, not flaky E2E tests
2. **Main CI must be green = mergeable** - no yellow/red noise
3. **E2E tests provide additional signal** but don't block development
4. **Local testing is the primary validation** - CI is a safety net

This approach ensures:
- Fast feedback on PRs
- No false positives blocking merges
- Comprehensive coverage without CI instability
- Clear signal when tests fail (real issues, not flakiness)
