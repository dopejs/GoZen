# Quickstart: Code Scenario Routing

**Feature**: 009-code-scenario-routing
**Date**: 2026-03-04

## Development Setup

```bash
# Ensure you're on the feature branch
git checkout 009-code-scenario-routing

# Verify Go builds
go build ./...

# Run existing tests (should all pass before changes)
go test ./...

# Start dev daemon for manual testing
./scripts/dev.sh restart
```

## Implementation Order

1. **Go config constant** → `internal/config/config.go`
   - Add `ScenarioCode Scenario = "code"` to the const block

2. **Go detection logic (TDD)** → `internal/proxy/scenario.go` + `scenario_test.go`
   - Write tests first in `scenario_test.go`
   - Add `isCodeRequest()` function
   - Update `DetectScenario()` priority chain

3. **TUI label** → `tui/routing.go`
   - Add entry to `knownScenarios` slice

4. **Web UI types + labels** → `web/src/types/api.ts` + i18n files
   - Add `'code'` to `Scenario` type, `SCENARIOS` array, `SCENARIO_LABELS`
   - Add `scenarioCode` to all i18n locale files

## Verification

```bash
# Go tests (after implementation)
go test ./internal/proxy/ -run TestDetectScenarioCode -v
go test ./internal/proxy/ -run TestDetectScenario -v
go test ./...

# Coverage check
go test -cover ./internal/proxy/
go test -cover ./internal/config/

# Frontend tests
cd web && pnpm run test:coverage

# Manual test via dev daemon
./scripts/dev.sh restart
# Open http://localhost:29840 → Profiles → Edit → Routing tab
# Verify "Code" scenario appears in the list
```

## Configuration Example

After implementation, users configure the `code` scenario in `zen.json`:

```json
{
  "profiles": {
    "my-profile": {
      "providers": ["provider-a", "provider-b"],
      "routing": {
        "think": {
          "providers": [{"name": "provider-a", "model": "claude-opus-4"}]
        },
        "code": {
          "providers": [{"name": "provider-b", "model": "claude-sonnet-4"}]
        }
      }
    }
  }
}
```

Result: thinking requests → provider-a with claude-opus-4, regular coding → provider-b with claude-sonnet-4.
