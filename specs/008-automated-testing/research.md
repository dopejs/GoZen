# Research: Automated Testing Infrastructure

**Branch**: `008-automated-testing` | **Date**: 2026-03-04

## R1: Test Helper Consolidation Strategy

**Decision**: Extract shared helpers (`findProjectRoot`, `findFreePort`, `setupTest` patterns) into `test/integration/helpers_test.go` and refactor the three existing config structs (`TestConfig`, `ProxyTestConfig`, `WebTestConfig`) to embed a common `BaseTestConfig`.

**Rationale**: The three existing test files in `test/integration/` each independently define `findProjectRoot()`, `findFreePort()`, binary build, config writing, and daemon startup. This duplication will worsen as new tests are added. A shared helper file reduces maintenance cost and makes patterns discoverable.

**Alternatives considered**:
- Leave duplicated code as-is: Rejected because adding 15+ new test functions would compound the duplication further.
- Create a `testutil` package: Rejected (YAGNI) — helpers are only used within `test/integration/`, so a package-level export is unnecessary.

## R2: Mock Provider Server Design

**Decision**: Create a configurable `MockProvider` struct in `test/integration/mock_provider_test.go` that wraps `httptest.Server` and supports: (a) configurable response status codes per request, (b) request counting, (c) latency injection, (d) response body templates for Anthropic `/v1/messages` format.

**Rationale**: Existing tests create ad-hoc `httptest.NewServer` with inline handlers. For config hot-reload and failover testing, we need mock servers that can change behavior at runtime (e.g., "return 500 for first 2 requests, then 200").

**Alternatives considered**:
- Use existing inline `httptest.NewServer`: Insufficient for dynamic behavior changes needed by hot-reload tests.
- Use a third-party mock library (e.g., `go-testdeep`): Rejected (Principle II — simplicity). stdlib `httptest` with a thin wrapper is sufficient.

## R3: E2E Proxy Test Architecture

**Decision**: New e2e proxy tests go in `tests/e2e_proxy_test.go` (alongside existing `e2e_daemon_test.go`). They share the same `testEnv` infrastructure: build binary, create isolated config dir with ephemeral ports, start daemon, send requests through proxy. Mock provider servers run as `httptest.Server` instances in the test process, and their URLs are written into the config.

**Rationale**: The `tests/` directory already has the `testEnv` pattern and the `integration` build tag. Adding proxy-specific e2e tests here keeps the e2e scope clear (full binary tests) vs `test/integration/` (component integration tests).

**Alternatives considered**:
- Put all new tests in `test/integration/`: Rejected because the distinction between "component integration" and "full binary e2e" is valuable for CI (e2e is non-blocking per clarification).
- Create a separate `tests/e2e/` subdirectory: Rejected (YAGNI) — flat structure in `tests/` is sufficient for <10 test files.

## R4: Config Hot-Reload Test Approach

**Decision**: Hot-reload tests verify the flow: (1) start daemon with provider A pointing to mock-server-1, (2) send request → verify it reaches mock-server-1, (3) update provider A's `base_url` via `PUT /api/v1/providers/A` to point to mock-server-2, (4) wait up to 5 seconds for config watcher to trigger reload, (5) send another request → verify it reaches mock-server-2.

**Rationale**: The config watcher (`internal/daemon/server.go::onConfigReload`) is triggered by file system events. The Web API writes to `zen.json`, which triggers the watcher. The test verifies the full chain: API → file write → watcher → proxy reload.

**Alternatives considered**:
- Directly modify `zen.json` file instead of using Web API: Rejected because spec US1 requires testing the Web API path specifically.
- Use `POST /api/v1/reload` to force reload: Could be used as a fallback if file watcher timing is unreliable, but the primary test should exercise the natural watcher path.

## R5: Stress Test Memory Monitoring

**Decision**: The stress test sends 500+ requests in a tight loop, then queries `/proc/self/status` (Linux) or `runtime.MemStats` via a debug endpoint. Since the daemon is a separate process, memory monitoring uses `ps -o rss= -p <PID>` to read RSS before and after the test run.

**Rationale**: Go's `runtime.MemStats` is only accessible from within the process. Since e2e tests run the daemon as a child process, external memory monitoring via `ps` is the only option without adding a debug API.

**Alternatives considered**:
- Add a `/debug/memstats` endpoint to the daemon: Rejected (Principle II — adding production code for test convenience). `ps` provides sufficient accuracy for a 50MB threshold check.
- Use `/proc/<pid>/status` directly: Only works on Linux, not macOS. `ps -o rss=` works on both.

## R6: Frontend Page Test Strategy

**Decision**: Page tests use the existing vitest + @testing-library/react + MSW setup. Each page test file (`*.test.tsx`) imports the page component, wraps it in necessary providers (React Router, TanStack Query, i18n), and verifies rendering and interactions against MSW-mocked API responses.

**Rationale**: The existing `web/src/test/` setup already configures MSW with handlers for all API endpoints. Page tests just need to render components and assert on DOM output.

**Alternatives considered**:
- Playwright/Cypress for browser-based testing: Rejected per spec (out of scope). vitest + testing-library provides sufficient component-level coverage.
- Storybook + visual regression testing: Rejected (YAGNI) — not requested, adds complexity.

## R7: CI Pipeline Changes

**Decision**: Add a new job `e2e` in `.github/workflows/ci.yml` that runs `go test -tags=integration -timeout 180s ./tests/...`. This job uses `continue-on-error: true` to make it non-blocking. The existing `Integration Tests` step in the `go` job continues to run `test/integration/` as a blocking step.

**Rationale**: Per clarification, integration tests block merge but e2e tests do not. A separate job provides clear visibility into e2e results without affecting the merge gate.

**Alternatives considered**:
- Single job with `|| true` on e2e step: Rejected because it hides failures in the job summary.
- Separate workflow file for e2e: Rejected (YAGNI) — a job within the existing workflow is simpler.

## R8: Testing Skills Format

**Decision**: Each skill is a `.claude/commands/test.*.md` file with YAML frontmatter (`description`, optional `handoffs`) and a Markdown body that instructs Claude Code to run specific test commands, parse output, and report results. Skills use `$ARGUMENTS` for optional flags (e.g., `/test.run internal/proxy` to test a specific package).

**Rationale**: This exactly follows the existing `speckit.*` skill format already used in the project. No new infrastructure is needed.

**Alternatives considered**:
- Shell scripts instead of skills: Rejected because the spec requires Claude Code slash commands that can be invoked interactively and provide intelligent feedback (coverage analysis, test skeleton generation).
- Combined single `/test` skill with subcommands: Rejected in favor of separate skills for discoverability and independent handoff chains.
