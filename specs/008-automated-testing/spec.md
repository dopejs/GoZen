# Feature Specification: Comprehensive Automated Testing Infrastructure

**Feature Branch**: `008-automated-testing`
**Created**: 2026-03-04
**Status**: Draft
**Input**: User description: "建立充分的自动化测试渠道，添加适合的skill，和足够的集成测试及 e2e 测试用例，主要测试范围包括：1. web ui 功能可用、配置保存、配置热加载到运行中的proxy; 2. proxy 稳定性（daemon不异常终止、终止后自动重启、port不变; provider fallback顺利完成; 基于场景的routing稳定健壮不降智）; 3. web ui/bot 进程稳定，不异常终止，终止后自动重启（服务器部署场景，机器不休眠不关机）"

## Problem Statement

GoZen has extensive unit test coverage (80%+ for core modules) but lacks comprehensive integration and end-to-end testing that validates the system behaves correctly as a whole. The existing e2e tests (in `tests/e2e_daemon_test.go`) cover daemon stability scenarios from spec 007, and `test/integration/` covers proxy routing and web API individually, but critical gaps remain:

1. **Web UI integration**: No tests verify that configuration changes made through the Web UI are persisted correctly and hot-reloaded into the running proxy. Frontend page components have zero test coverage — only hooks and API client functions are tested.

2. **Proxy end-to-end reliability**: No tests exercise the full provider fallback chain through the actual proxy binary (current proxy tests use in-process mock servers). Scenario-based routing (think/image/long-context) is unit-tested in `server_test.go` but never validated against the real daemon with actual request routing.

3. **Process stability for server deployments**: GoZen is increasingly deployed on always-on servers where the Web UI and bot processes must remain accessible 24/7. There is no mechanism to ensure these processes auto-restart after unexpected termination, and no tests validate long-running process resilience.

## Clarifications

### Session 2026-03-04

- Q: Should new integration and e2e tests be added to the CI pipeline, and if so, should e2e test failures block merge? → A: Add both `test/integration/` and `tests/` (e2e) to CI. Integration tests are blocking (failures block merge). E2E tests run as a separate non-blocking job (failures are reported but do not block merge).

## Current Testing Landscape

| Area | Unit Tests | Integration Tests | E2E Tests | Gap |
|------|-----------|------------------|-----------|-----|
| Config store | 80%+ | - | - | No hot-reload verification |
| Proxy routing | 80%+ (58 tests) | 8 tests (`test/integration/proxy_test.go`) | - | No full-binary failover test |
| Scenario routing | Covered in `server_test.go` | - | - | No real daemon scenario test |
| Web API | 200+ tests | 12 tests (`test/integration/web_test.go`) | - | No config-change-to-proxy flow |
| Daemon lifecycle | 40+ tests | 12 tests (`test/integration/daemon_test.go`) | 6 tests (`tests/e2e_daemon_test.go`) | Covered well for port stability |
| Web UI pages | 0 tests | - | - | No page component tests |
| Web UI hooks | 10 test files | - | - | Hooks covered, pages not |
| Bot gateway | 13 test files | - | - | No process stability tests |
| Process supervisor | N/A | N/A | N/A | Does not exist yet |

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Web UI Config Persistence and Hot-Reload (Priority: P0)

As a user managing GoZen through the Web UI, when I add/edit/remove a provider or change profile settings, those changes must be persisted to disk and immediately take effect in the running proxy — without restarting the daemon.

**Why this priority**: This is the core user workflow. If config changes don't persist or don't take effect in the running proxy, the Web UI is unreliable as a management interface. Users currently have no automated confidence that their config changes actually work end-to-end.

**Independent Test**: Start daemon, make config changes via Web API, verify changes are persisted in `zen.json`, then verify the running proxy reflects the changes (e.g., a newly added provider receives traffic).

**Acceptance Scenarios**:

1. **Given** a running daemon with provider "A", **When** provider "B" is added via the Web API, **Then** `zen.json` contains provider "B", AND the proxy routes requests to provider "B" when provider "A" is unavailable.
2. **Given** a running daemon with profile "default" using providers ["A", "B"], **When** the profile is updated to ["B", "A"] via the Web API, **Then** subsequent proxy requests route to provider "B" first (failover order changed).
3. **Given** a running daemon, **When** a provider is removed via the Web API, **Then** the provider is removed from `zen.json` AND the proxy no longer routes to it.
4. **Given** a running daemon, **When** settings are updated via the Web API (e.g., compression enabled), **Then** the setting is persisted AND takes effect immediately in the proxy.
5. **Given** a running daemon with provider "A" (base_url pointing to mock server), **When** provider "A" is edited to point to a different mock server via the Web API, **Then** subsequent proxy requests are forwarded to the new URL.

---

### User Story 2 - Proxy Stability Under Adverse Conditions (Priority: P0)

As a user relying on GoZen as my daily AI proxy, the proxy must remain stable and functional under all normal operating conditions: the daemon should not crash, the proxy port must never change, provider fallback must complete reliably across multiple providers, and scenario-based routing must correctly identify request types to prevent intelligence downgrade.

**Why this priority**: Proxy stability is the fundamental value proposition of GoZen. If the proxy is unreliable, users lose trust and revert to direct API access. The existing e2e tests cover port stability and daemon recovery, but do not cover provider fallback or scenario routing through the real binary.

**Independent Test**: Start daemon with multiple mock providers, send requests that trigger failover and scenario routing, verify correct behavior through the actual binary.

**Acceptance Scenarios**:

1. **Given** a daemon with providers [A(down), B(healthy)], **When** a request is sent through the proxy, **Then** the proxy fails over from A to B and returns a successful response.
2. **Given** a daemon with providers [A(down), B(down), C(healthy)], **When** a request is sent, **Then** the proxy tries A, then B, then C, and returns C's response.
3. **Given** a daemon with all providers returning 500 errors, **When** a request is sent, **Then** the proxy returns a clear error (not a hang or crash) and the daemon remains running.
4. **Given** a daemon with scenario routing configured (think-capable provider and standard provider), **When** a request containing extended thinking parameters is sent, **Then** the proxy routes to the think-capable provider (not the standard one).
5. **Given** a daemon running for an extended period (simulated via rapid request cycling), **When** 500+ requests are sent in sequence with varying success/failure patterns, **Then** the daemon does not leak memory, hang, or crash.
6. **Given** a daemon with providers returning 429 (rate limit), **When** requests continue to arrive, **Then** the proxy backs off from rate-limited providers and routes to available ones without crashing.
7. **Given** a daemon with a provider that takes >30 seconds to respond, **When** a request is sent and the client disconnects (context canceled), **Then** the proxy handles the cancellation gracefully and remains stable for subsequent requests.

---

### User Story 3 - Process Stability for Server Deployment (Priority: P1)

As a server operator running GoZen on an always-on machine, I need the daemon (including web UI and bot gateway) to automatically recover from unexpected process termination, so that the web UI and bot are always accessible without manual intervention.

**Why this priority**: Server deployment is a key use case where GoZen runs unattended. The daemon currently has auto-restart via the `zen` wrapper (spec 007), but there is no supervisor mechanism for the daemon process itself when running on a server without a user interacting via `zen`. The bot gateway and web UI must survive crashes and resource exhaustion.

**Independent Test**: Start daemon in foreground mode (simulating server deployment), kill the process, verify it can be restarted and all services (web UI, proxy, bot) are functional.

**Acceptance Scenarios**:

1. **Given** a daemon running in server mode, **When** the daemon process receives SIGTERM, **Then** it shuts down gracefully (completes in-flight requests, saves state) and exits cleanly.
2. **Given** a daemon that was killed with SIGKILL, **When** a restart is attempted, **Then** the daemon detects and cleans up the stale PID file, acquires the lock, and starts successfully on the same ports.
3. **Given** a daemon with an active bot gateway (e.g., Telegram adapter), **When** the daemon is restarted, **Then** the bot gateway reconnects and resumes handling messages.
4. **Given** a daemon running for an extended period, **When** the config file watcher detects changes, **Then** the proxy reloads without interrupting active connections, and the web UI remains accessible throughout.
5. **Given** a system with a process supervisor (e.g., systemd, launchd, or a simple shell loop), **When** the daemon crashes, **Then** the supervisor restarts it and all services become available again within 10 seconds. *(Note: the supervisor itself is out of scope — this tests that the daemon is supervisor-friendly: clean PID file handling, idempotent startup, no port conflicts on restart.)*

---

### User Story 4 - Testing Skills for Development Workflow (Priority: P1)

As a developer working on GoZen features, I need Claude Code slash commands (skills) that guide me through writing and running automated tests, so that every future feature branch maintains test quality without relying on tribal knowledge of the test infrastructure.

**Why this priority**: The testing infrastructure built in US1-US3 is only valuable if developers actually use it consistently. Without process-level guidance, new feature development may skip integration/e2e tests or run them incorrectly. Skills encode the testing workflow as repeatable, discoverable commands that enforce quality gates during development.

**Independent Test**: Invoke each skill from the Claude Code CLI and verify it produces the correct output (runs the right tests, generates the right reports, identifies the right coverage gaps).

**Acceptance Scenarios**:

1. **Given** a developer has modified Go code in a feature branch, **When** they invoke `/test.run`, **Then** the skill runs unit tests, checks coverage thresholds, and reports pass/fail with specific failing test names and uncovered packages.
2. **Given** a developer has made changes to proxy, daemon, or web API code, **When** they invoke `/test.integration`, **Then** the skill builds the binary, starts isolated test daemons, runs integration and e2e tests with the `integration` build tag, and reports results.
3. **Given** a developer has modified frontend code in `web/`, **When** they invoke `/test.web`, **Then** the skill runs `pnpm test` with coverage, reports page-level and hook-level results, and flags coverage drops below the 70% threshold.
4. **Given** a developer is about to create a commit or PR, **When** they invoke `/test.all`, **Then** the skill runs all test tiers (unit, integration, e2e, web) in sequence and produces a consolidated pass/fail report suitable for a pre-commit quality gate.
5. **Given** a developer is starting a new feature, **When** they invoke `/test.write` with a description of what they changed, **Then** the skill analyzes the modified files, identifies which test files and test patterns to follow, and generates skeleton test cases following the project's TDD conventions.

---

### User Story 5 - Frontend Page Component Testing (Priority: P2)

As a developer maintaining the GoZen Web UI, I need page-level component tests for critical UI pages to catch regressions when modifying the frontend code, especially for pages with complex state management like monitoring, providers, and profiles.

**Why this priority**: The frontend currently has zero page-level tests. While hooks are tested, the page components contain significant rendering logic, state management, and user interaction flows that are untested. This is lower priority than the backend integration tests because the backend is the critical reliability path, but frontend regressions are still costly.

**Independent Test**: Run `pnpm test` in `web/` and verify that all critical page components render correctly with mocked API data and respond to user interactions.

**Acceptance Scenarios**:

1. **Given** the monitoring page component with mock request data, **When** rendered, **Then** it displays requests in a table with correct timestamps, durations, and status indicators, AND auto-refresh toggles work.
2. **Given** the providers page with mock provider data, **When** a user clicks "Add Provider" and fills in the form, **Then** the form validates inputs and calls the API with correct data.
3. **Given** the profiles page with mock profile data, **When** a user reorders providers via drag-and-drop, **Then** the new order is reflected in the API call.
4. **Given** the settings page, **When** a user views the General tab, **Then** the proxy port is displayed as read-only with the correct value from settings API.

---

### Edge Cases

- What happens when the config file is corrupted (invalid JSON) during a hot-reload? The proxy should continue running with the last good config.
- What happens when a provider's base URL is unreachable (DNS resolution failure)? The proxy should failover to the next provider without hanging.
- What happens when the daemon is started twice simultaneously (race condition)? The file lock should prevent dual startup.
- What happens when the web UI makes a config change while a request is being proxied? The in-flight request should complete on the old config; new requests use the updated config.
- What happens when all providers in the default profile are removed via Web API? The proxy should return a clear error for new requests, not crash.
- What happens when the bot gateway receives messages during a config reload? Messages should be queued or handled without loss.
- What happens when the PID file exists but the process is gone (stale PID)? The daemon should overwrite the PID file and start normally.
- What happens when the daemon is run on a system where the ports are firewalled? The daemon should start and log a warning if the port bind succeeds but external connectivity fails.

## Requirements *(mandatory)*

### Functional Requirements

#### Integration Test Infrastructure

- **FR-001**: The project MUST have a comprehensive integration test suite in `test/integration/` that validates config-change-to-proxy-effect flows by: (a) starting a real daemon with mock providers, (b) making config changes via the Web API, (c) verifying the proxy reflects those changes.
- **FR-002**: Integration tests MUST use isolated config directories and ephemeral ports (following the pattern established in `tests/e2e_daemon_test.go`) to avoid interfering with production or dev instances.
- **FR-003**: Integration tests MUST be gated behind the `integration` build tag so they do not run with `go test ./...`.

#### Web UI Config Persistence Tests

- **FR-004**: Integration tests MUST verify that adding a provider via `POST /api/v1/providers` persists to `zen.json` and the provider becomes available in the running proxy.
- **FR-005**: Integration tests MUST verify that modifying profile provider order via `PUT /api/v1/profiles/:name` changes the proxy's failover order.
- **FR-006**: Integration tests MUST verify that removing a provider via `DELETE /api/v1/providers/:name` removes it from the proxy's routing table.
- **FR-007**: Integration tests MUST verify that updating settings via `PUT /api/v1/settings` takes effect immediately (e.g., compression toggle).

#### Proxy Stability Tests

- **FR-008**: E2E tests MUST verify provider failover through the real proxy binary: send a request when the first N-1 providers are down, verify the Nth provider receives and handles the request.
- **FR-009**: E2E tests MUST verify that scenario-based routing correctly identifies thinking-mode requests and routes them to the designated provider.
- **FR-010**: E2E tests MUST verify the proxy remains stable after handling 500+ requests with mixed success/failure patterns (no memory leaks, no goroutine leaks, no crashes).
- **FR-011**: E2E tests MUST verify graceful handling of client disconnection (context cancellation) during an in-flight proxy request.

#### Process Stability Tests

- **FR-012**: E2E tests MUST verify that the daemon can be cleanly restarted after SIGTERM (graceful shutdown: PID file removed, port released, lock released).
- **FR-013**: E2E tests MUST verify that the daemon can be restarted after SIGKILL (unclean shutdown: stale PID file, port released by OS, lock released by OS).
- **FR-014**: E2E tests MUST verify that the daemon is idempotent on startup: starting when already running should detect the existing daemon and not start a duplicate.
- **FR-015**: The project SHOULD provide a reference process supervisor configuration (e.g., launchd plist or systemd unit) demonstrating how to run GoZen as a server-mode service. *(Documentation only, not code.)*

#### Frontend Component Tests

- **FR-016**: The Web UI MUST have page-level component tests for the monitoring page (`web/src/pages/monitoring/index.tsx`) covering: rendering with mock data, auto-refresh toggle, filter interactions, and detail modal display.
- **FR-017**: The Web UI MUST have page-level component tests for the providers page (`web/src/pages/providers/`) covering: list rendering, add form validation, edit flow.
- **FR-018**: The Web UI MUST have page-level component tests for the profiles page (`web/src/pages/profiles/`) covering: list rendering, provider reordering, routing config.
- **FR-019**: The Web UI MUST have page-level component tests for the settings page (`web/src/pages/settings/`) covering: general tab rendering with read-only proxy port, password change flow.
- **FR-020**: Frontend test coverage MUST remain above 70% (current CI threshold) after adding page tests.

#### Testing Skills (Claude Code Slash Commands)

- **FR-021**: A `/test.run` skill MUST exist as a Claude Code command (`.claude/commands/test.run.md`) that: (a) detects which Go packages have been modified on the current branch, (b) runs `go test` for those packages with race detection and coverage, (c) compares coverage against CI thresholds (80% for core, 50% for supporting), (d) reports pass/fail with specific test names and uncovered lines.
- **FR-022**: A `/test.integration` skill MUST exist that: (a) builds the `zen` binary, (b) runs integration tests (`go test -tags=integration ./test/integration/...`) and e2e tests (`go test -tags=integration ./tests/...`), (c) reports results including daemon startup/shutdown status and mock provider interactions.
- **FR-023**: A `/test.web` skill MUST exist that: (a) runs `pnpm test` in `web/` with coverage, (b) reports page-level and hook-level test results, (c) flags coverage drops below the 70% threshold.
- **FR-024**: A `/test.all` skill MUST exist that runs all test tiers in sequence (unit, integration, e2e, web) and produces a consolidated pass/fail summary. This skill SHOULD be recommended as a pre-commit/pre-PR quality gate.
- **FR-025**: A `/test.write` skill MUST exist that: (a) analyzes files modified on the current branch (`git diff` against base), (b) identifies corresponding test files and existing test patterns, (c) generates skeleton test cases following the project's conventions (table-driven Go tests, vitest+testing-library for frontend), (d) respects the TDD principle from the constitution (Principle I).
- **FR-026**: All testing skills MUST follow the existing Claude Code command format: a Markdown file in `.claude/commands/` with YAML frontmatter (`description` field) and a structured body with `$ARGUMENTS` support.
- **FR-027**: Testing skills SHOULD include `handoffs` in their frontmatter to suggest logical next steps (e.g., `/test.run` suggests `/test.integration` on success; `/test.all` suggests `/commit` on full pass).

#### Test Tooling

- **FR-028**: A `Makefile` target or script MUST exist to run all test tiers: `make test-unit` (Go unit tests), `make test-integration` (Go integration tests), `make test-e2e` (Go e2e tests), `make test-web` (frontend tests), `make test-all` (everything).
- **FR-029**: Integration and E2E tests MUST use mock HTTP servers (Go `httptest.Server`) as provider backends to avoid dependency on real API providers.
- **FR-030**: The CI pipeline (`.github/workflows/ci.yml`) MUST be updated to run both integration tests (`test/integration/`) and e2e tests (`tests/`). Integration test failures MUST block merge (existing behavior). E2E test failures MUST be reported but MUST NOT block merge (non-blocking separate job).

### Key Entities

- **Mock Provider Server**: An `httptest.Server` that simulates an Anthropic/OpenAI API provider — responds to `/v1/messages` and `/v1/chat/completions` with configurable responses (success, failure, delay, rate-limit).
- **Test Daemon**: A real `zen` binary running against an isolated config directory with ephemeral ports, started/stopped programmatically by integration tests.
- **Test Web Client**: An HTTP client that calls the Web API endpoints and verifies responses, used by integration tests to simulate Web UI interactions.
- **Testing Skill**: A Claude Code slash command (`.claude/commands/test.*.md`) that automates a specific testing workflow — running tests, checking coverage, generating test skeletons, or orchestrating all tiers as a quality gate.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Config changes via Web API are persisted to disk AND reflected in the proxy within 5 seconds, verified by automated integration tests.
- **SC-002**: Provider failover completes correctly through the real binary for 2-provider and 3-provider chains, verified by automated e2e tests.
- **SC-003**: Scenario routing correctly directs thinking-mode requests to the designated provider 100% of the time, verified by automated e2e tests.
- **SC-004**: The daemon survives 500+ mixed-result requests without memory growth exceeding 50MB above baseline, verified by an automated stress test.
- **SC-005**: The daemon restarts cleanly after both SIGTERM and SIGKILL, verified by automated e2e tests.
- **SC-006**: Frontend page component test coverage covers monitoring, providers, profiles, and settings pages, with overall web UI coverage remaining above 70%.
- **SC-007**: A single command (`make test-all` or equivalent) runs all test tiers and produces a pass/fail result suitable for CI.
- **SC-008**: All integration and e2e tests complete within 3 minutes on a standard development machine.
- **SC-009**: All five testing skills (`/test.run`, `/test.integration`, `/test.web`, `/test.all`, `/test.write`) are installable as `.claude/commands/test.*.md` files and invocable from the Claude Code CLI.
- **SC-010**: A developer following the `/test.write` → `/test.run` → `/test.integration` → `/test.all` workflow produces feature branches that pass CI on the first push 90% of the time.

## Assumptions

- Mock HTTP servers (`httptest.Server`) provide sufficient fidelity for testing provider interactions — real API credentials are never required for automated tests.
- The `integration` build tag convention from the existing test suite is maintained for all new integration and e2e tests.
- The existing daemon fork model (`GOZEN_DAEMON=1` for foreground child process) is stable and testable — new tests build upon the `testEnv` infrastructure from `tests/e2e_daemon_test.go`.
- Frontend page tests use the existing MSW (Mock Service Worker) setup in `web/src/test/` for API mocking.
- Process supervisor integration (launchd/systemd) is out of scope for code changes — the spec only requires that the daemon be "supervisor-friendly" (clean PID files, idempotent startup, no port conflicts on restart).

## Scope

### In Scope

- Integration tests for Web API → config persistence → proxy hot-reload flow
- E2E tests for provider failover through the real binary
- E2E tests for scenario-based routing through the real binary
- E2E tests for proxy stress/stability (high request volume)
- E2E tests for process lifecycle (SIGTERM, SIGKILL, restart, idempotent startup)
- Frontend page component tests for monitoring, providers, profiles, and settings
- Test runner script/Makefile for all test tiers
- Mock provider server infrastructure for integration tests
- Claude Code testing skills (`/test.run`, `/test.integration`, `/test.web`, `/test.all`, `/test.write`) as `.claude/commands/` files

### Out of Scope

- Real API provider testing (all tests use mocks)
- Browser-based E2E testing (Playwright/Cypress) — frontend tests use vitest + testing-library
- Process supervisor implementation (launchd/systemd unit files) — reference docs only
- Performance benchmarking beyond basic stability verification
- Mobile/responsive UI testing
- Network partition simulation (tests run on localhost)
