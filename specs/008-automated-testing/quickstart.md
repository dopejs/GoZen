# Quickstart: Automated Testing Infrastructure

**Branch**: `008-automated-testing` | **Date**: 2026-03-04

## Manual Verification Scenarios

After implementation, verify these scenarios manually to confirm the automated tests work correctly.

### Scenario 1: Run All Test Tiers via Makefile

```bash
# Run each tier individually
make test-unit          # Should run go test ./... and report pass/fail
make test-integration   # Should build binary, run test/integration/ with integration tag
make test-e2e           # Should build binary, run tests/ with integration tag
make test-web           # Should run pnpm test in web/

# Run everything
make test-all           # Should run all tiers in sequence, report consolidated result
```

**Expected**: All commands complete successfully. `test-all` takes <3 minutes.

### Scenario 2: Config Hot-Reload Integration Test

```bash
go test -tags=integration -run TestIntegration_ConfigHotReload -v ./test/integration/...
```

**Expected**: Test starts daemon, adds provider via Web API, verifies proxy routes to the new provider. Output shows:
- Daemon started on ephemeral ports
- Provider added via POST /api/v1/providers
- Request routed to mock server after config reload

### Scenario 3: Provider Failover E2E Test

```bash
go test -tags=integration -run TestE2E_ProviderFailover -v ./tests/...
```

**Expected**: Test starts daemon with 3 mock providers (2 failing, 1 healthy), sends request, verifies failover completes. Output shows:
- Provider A returned 503
- Provider B returned 503
- Provider C returned 200 (success)

### Scenario 4: Stress Test

```bash
go test -tags=integration -run TestE2E_StressTest -v -timeout 180s ./tests/...
```

**Expected**: Test sends 500+ requests through proxy, monitors memory. Output shows:
- Request count and success/failure breakdown
- Memory before and after
- No memory growth exceeding 50MB

### Scenario 5: Frontend Page Tests

```bash
cd web && pnpm test -- --run
```

**Expected**: All page component tests pass (monitoring, providers, profiles, settings). Coverage remains above 70%.

### Scenario 6: Testing Skills

```bash
# In Claude Code CLI:
/test.run
/test.integration
/test.web
/test.all
/test.write
```

**Expected**: Each skill executes the appropriate tests and reports results. `/test.write` analyzes modified files and suggests test skeletons.

### Scenario 7: CI Pipeline

Push branch to remote, open PR. Verify:
1. `go` job runs unit tests, integration tests, and coverage checks (blocking)
2. `e2e` job runs e2e tests from `tests/` (non-blocking, `continue-on-error: true`)
3. If e2e fails, PR is still mergeable but the job shows as failed for visibility
