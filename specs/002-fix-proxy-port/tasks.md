# Tasks: Fix Proxy Port Stability

**Input**: Design documents from `/specs/002-fix-proxy-port/`
**Prerequisites**: plan.md, spec.md

**Tests**: Included per constitution Principle I (TDD) and Principle VI (Coverage Enforcement).

**Organization**: Tasks grouped by user story. US1 is the core bug fix; US2 is verification-only (no code changes beyond US1).

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2)

---

## Phase 1: Setup

**Purpose**: No setup needed — existing project, existing branch.

_(No tasks — project structure and dependencies already in place.)_

---

## Phase 2: Foundational

**Purpose**: Verify preconditions before making changes.

- [x] T001 Verify `config.GetProxyPort()` returns default 19841 when config value is 0 or unset in `internal/config/compat.go`
- [x] T002 Verify `internal/daemon/server.go` daemon path reads configured port correctly (reference implementation, read-only)

**Checkpoint**: Preconditions confirmed — implementation can begin.

---

## Phase 3: User Story 1 — Stable Proxy Port on Legacy Path (Priority: P1) 🎯 MVP

**Goal**: Legacy proxy startup reads the configured port instead of hardcoding `:0`.

**Independent Test**: Start GoZen via legacy proxy path, verify it listens on configured port (default 19841 or custom). Restart and confirm same port.

### Tests for User Story 1 (TDD — write first, verify they fail)

- [x] T003 [P] [US1] Write table-driven test for `GetProxyPort()` in `internal/config/config_test.go` — cases: unset→19841, zero→19841, custom value preserved, port-out-of-range (negative, >65535) returns error or default
- [x] T004 [P] [US1] Write test verifying port-conflict error message includes port number when binding fails, in `cmd/root_test.go`

### Implementation for User Story 1

- [x] T005 [US1] Replace hardcoded `"127.0.0.1:0"` with `config.GetProxyPort()` address in `startLegacyProxy()` in `cmd/root.go`
- [x] T006 [US1] Verify logged proxy address and env vars match the configured port in `cmd/root.go`
- [x] T007 [US1] Run tests and verify T003/T004 pass — `go test ./internal/config/ ./cmd/`

**Checkpoint**: Legacy proxy binds to configured port. Core bug is fixed.

---

## Phase 4: User Story 2 — Consistent Port Across Startup Modes (Priority: P2)

**Goal**: Daemon and legacy paths produce identical port behavior for the same config.

**Independent Test**: Start via daemon mode, note port. Stop daemon. Start via legacy mode, confirm same port.

### Verification for User Story 2

- [x] T008 [US2] Verify daemon path in `internal/daemon/server.go` and legacy path in `cmd/root.go` both call `config.GetProxyPort()` with same resolution logic

_(No code changes needed — US1 fix aligns legacy path with daemon path. This phase is verification-only.)_

**Checkpoint**: Both startup modes use identical port resolution.

---

## Phase 5: Polish & Cross-Cutting Concerns

- [x] T009 Run full test suite — `go test ./...`
- [x] T010 Verify `internal/proxy` coverage meets 80% threshold — `go test -cover ./internal/proxy/`
- [x] T011 Verify no other packages dropped below CI thresholds — `go test -cover ./internal/config/ ./internal/web/ ./internal/bot/`

---

## Dependencies & Execution Order

### Phase Dependencies

- **Foundational (Phase 2)**: No dependencies — verification only
- **User Story 1 (Phase 3)**: Depends on Phase 2 confirmation
- **User Story 2 (Phase 4)**: Depends on US1 completion (Phase 3)
- **Polish (Phase 5)**: Depends on all phases complete

### Within User Story 1

- T003, T004 (tests) MUST be written and FAIL before T005 (implementation)
- T005, T006 (implementation) before T007 (test verification)

### Parallel Opportunities

- T001 and T002 can run in parallel (both read-only verification)
- T003 and T004 can run in parallel (different test cases, same file but independent)

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 2: Verify preconditions
2. Write failing tests (T003, T004)
3. Implement fix (T005) and verify log/env output (T006) — ~5 lines changed in `cmd/root.go`
4. Verify tests pass (T007)
5. **STOP and VALIDATE**: Legacy proxy uses configured port

### Incremental Delivery

1. US1 fixes the core bug → MVP complete
2. US2 is verification-only → confirms consistency
3. Polish verifies coverage thresholds

---

## Notes

- Total scope: ~10 lines changed in 1 file (`cmd/root.go`)
- `internal/proxy/server.go` needs no changes — bug is in the caller
- Config schema unchanged — `proxy_port` field already exists
- Commit after T007 (US1 complete) and after T011 (all verified)
