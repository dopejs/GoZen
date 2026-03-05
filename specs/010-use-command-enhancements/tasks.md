# Tasks: zen use Command Enhancements

**Input**: Design documents from `/specs/010-use-command-enhancements/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/

**Tests**: TDD is required per constitution. Tests MUST be written first and verified to fail before implementation.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Path Conventions

This is a CLI tool with embedded web UI:
- **Backend (Go)**: `cmd/`, `internal/config/`
- **Frontend (React)**: `web/src/`
- **Tests**: `*_test.go` files alongside implementation, `web/tests/` for frontend
- **Documentation**: `README.md`, `docs/README.*.md`

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization and config schema foundation

- [X] T001 Read current config version from internal/config/config.go
- [X] T002 [P] Review existing prependAutoApproveArgs() function in cmd/root.go
- [X] T003 [P] Review existing config migration patterns in internal/config/

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core config schema changes that ALL user stories depend on

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

### Tests for Config Schema (TDD - Write First)

- [X] T004 [P] Write test for AutoPermissionConfig type validation in internal/config/config_test.go
- [X] T005 [P] Write test for config version migration (v11→v12) in internal/config/config_test.go
- [X] T006 [P] Write test for backward compatibility (old configs without new fields) in internal/config/config_test.go
- [X] T007 [P] Write test for forward compatibility (new configs read by old code) in internal/config/config_test.go
- [X] T008 [P] Write test for config round-trip marshaling in internal/config/config_test.go

### Config Schema Implementation

- [X] T009 Add AutoPermissionConfig type to internal/config/config.go
- [X] T010 Add claude_auto_permission field to OpenCCConfig in internal/config/config.go
- [X] T011 Add codex_auto_permission field to OpenCCConfig in internal/config/config.go
- [X] T012 Add opencode_auto_permission field to OpenCCConfig in internal/config/config.go
- [X] T013 Bump CurrentConfigVersion to 12 in internal/config/config.go
- [X] T014 Add config migration logic for v11→v12 in internal/config/store.go or UnmarshalJSON
- [X] T015 Run tests to verify config schema changes (T004-T008 should now pass)

**Checkpoint**: Config schema ready - user story implementation can now begin in parallel

---

## Phase 3: User Story 1 - Auto-Approve Client Permissions with --yes Flag (Priority: P1) 🎯 MVP

**Goal**: Fix `--yes` flag to use `bypassPermissions` for Claude Code and `-a never` for Codex, eliminating all permission prompts

**Independent Test**: Run `zen --yes` and verify Claude Code receives `--permission-mode bypassPermissions` (or Codex receives `-a never`), resulting in zero permission prompts

### Tests for User Story 1 (TDD - Write First)

- [X] T016 [P] [US1] Write test for prependAutoApproveArgs with --yes for Claude Code in cmd/root_test.go
- [X] T017 [P] [US1] Write test for prependAutoApproveArgs with --yes for Codex in cmd/root_test.go
- [X] T018 [P] [US1] Write test for prependAutoApproveArgs with --yes for OpenCode in cmd/root_test.go
- [ ] T019 [P] [US1] Write integration test for zen --yes command execution in cmd/root_test.go
- [ ] T019a [P] [US1] Write test for error messages when client not found in PATH in cmd/root_test.go
- [ ] T019b [P] [US1] Write test for exit status codes (0 for success, non-zero for errors) in cmd/root_test.go

### Implementation for User Story 1

- [X] T020 [US1] Modify prependAutoApproveArgs() in cmd/root.go to use bypassPermissions for Claude Code
- [X] T021 [US1] Modify prependAutoApproveArgs() in cmd/root.go to use -a never for Codex
- [X] T022 [US1] Handle OpenCode permission flags in prependAutoApproveArgs() in cmd/root.go (research Q1)
- [ ] T023 [US1] Run tests to verify --yes flag behavior (T016-T019b should now pass)
- [ ] T024 [US1] Manual test: Run zen --yes with Claude Code and verify no prompts
- [ ] T025 [US1] Manual test: Run zen --yes with Codex and verify no prompts

**Checkpoint**: User Story 1 complete - `--yes` flag now bypasses all permission prompts

---

## Phase 4: User Story 2 - Persistent Auto-Permission Configuration (Priority: P2)

**Goal**: Enable auto-permission mode by default through Web UI configuration, without typing `--yes` every time

**Independent Test**: Enable auto-permission mode in Web UI for Claude Code with `bypassPermissions`, then run `zen` and verify it automatically passes the configured permission flags

### Tests for User Story 2 (TDD - Write First)

#### Backend Tests

- [ ] T026 [P] [US2] Write test for getAutoPermissionConfig() helper in internal/config/config_test.go
- [ ] T027 [P] [US2] Write test for permission priority resolution (Web UI config) in cmd/root_test.go
- [ ] T028 [P] [US2] Write test for client-specific permission mode retrieval in cmd/root_test.go

#### Frontend Tests

- [ ] T029 [P] [US2] Write test for PermissionConfig component rendering in web/tests/components/PermissionConfig.test.tsx
- [ ] T030 [P] [US2] Write test for client-specific dropdown options in web/tests/components/PermissionConfig.test.tsx
- [ ] T031 [P] [US2] Write test for save configuration API call in web/tests/components/PermissionConfig.test.tsx

### Implementation for User Story 2

#### Backend Implementation

- [ ] T032 [P] [US2] Add getAutoPermissionConfig() helper to internal/config/compat.go
- [ ] T033 [US2] Modify startViaDaemon() in cmd/root.go to check Web UI auto-permission config
- [ ] T034 [US2] Implement permission priority resolution (Web UI config < --yes) in cmd/root.go
- [ ] T035 [US2] Run backend tests to verify Web UI config integration (T026-T028 should now pass)

#### Frontend Implementation

- [ ] T036 [P] [US2] Create PermissionConfig component in web/src/components/Settings/PermissionConfig.tsx
- [ ] T037 [P] [US2] Add client selector dropdown to PermissionConfig component
- [ ] T038 [P] [US2] Add enable/disable toggle to PermissionConfig component
- [ ] T039 [P] [US2] Add permission mode dropdown with client-specific options to PermissionConfig component
- [ ] T040 [US2] Wire PermissionConfig component to Settings page in web/src/pages/Settings.tsx
- [ ] T041 [US2] Add TypeScript types for auto-permission config in web/src/types/config.ts
- [ ] T042 [US2] Update config API client to handle auto-permission fields in web/src/api/config.ts
- [ ] T043 [US2] Run frontend tests to verify PermissionConfig component (T029-T031 should now pass)
- [ ] T044 [US2] Manual test: Enable auto-permission in Web UI and verify zen command uses it

**Checkpoint**: User Story 2 complete - Web UI auto-permission configuration working

---

## Phase 5: User Story 3 - Client Parameter Pass-Through (Priority: P3)

**Goal**: Support `--` separator to pass arbitrary parameters to underlying clients with proper priority handling

**Independent Test**: Run `zen -- --verbose --debug` and verify flags are passed to client. Run `zen --yes -- --permission-mode acceptEdits` and verify `acceptEdits` is used (not `bypassPermissions`)

### Tests for User Story 3 (TDD - Write First)

- [ ] T045 [P] [US3] Write test for detectPermissionFlags() helper in cmd/root_test.go
- [ ] T046 [P] [US3] Write test for -- parameter parsing in cmd/root_test.go
- [ ] T047 [P] [US3] Write test for permission priority resolution (-- > --yes > Web UI) in cmd/root_test.go
- [ ] T048 [P] [US3] Write test for -- with zen use command in cmd/use_test.go
- [ ] T049 [P] [US3] Write integration test for -- parameter pass-through in cmd/root_test.go

### Implementation for User Story 3

- [ ] T050 [P] [US3] Add detectPermissionFlags() helper to cmd/root.go
- [ ] T051 [US3] Modify startViaDaemon() to detect permission flags in -- parameters
- [ ] T052 [US3] Implement permission priority resolution (-- > --yes > Web UI > default) in cmd/root.go
- [ ] T053 [US3] Update zen use command to support -- separator in cmd/use.go
- [ ] T054 [US3] Run tests to verify -- parameter handling (T045-T049 should now pass)
- [ ] T055 [US3] Manual test: Run zen -- --verbose and verify flag passed to client
- [ ] T056 [US3] Manual test: Run zen --yes -- --permission-mode acceptEdits and verify acceptEdits used

**Checkpoint**: User Story 3 complete - Client parameter pass-through working with correct priority

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Documentation, help text, and final integration

### Documentation Updates

- [ ] T057 [P] Update zen --help text to document --yes behavior change in cmd/root.go
- [ ] T058 [P] Update zen --help text to document -- separator usage in cmd/root.go
- [ ] T059 [P] Update README.md (English) with new features and priority order
- [ ] T060 [P] Update docs/README.zh-CN.md (简体中文) with new features
- [ ] T061 [P] Update docs/README.zh-TW.md (繁體中文) with new features
- [ ] T062 [P] Update docs/README.es.md (Español) with new features

### Final Testing & Verification

- [ ] T063 Run go test ./... and verify all tests pass
- [ ] T064 Run go test -cover ./internal/config and verify ≥80% coverage
- [ ] T065 Run pnpm run test:coverage in web/ and verify ≥70% branch coverage
- [ ] T066 [P] Test config migration: Create v11 config, run zen, verify migration to v12
- [ ] T067 [P] Test backward compatibility: Use old zen with v12 config, verify graceful handling
- [ ] T068 [P] Test all three user stories end-to-end in sequence
- [ ] T069 Update cmd/root.go Version constant if preparing for release

**Checkpoint**: Feature complete and ready for PR

---

## Dependencies & Execution Strategy

### User Story Dependencies

```
Phase 1 (Setup) → Phase 2 (Foundational)
                      ↓
        ┌─────────────┼─────────────┐
        ↓             ↓             ↓
    Phase 3       Phase 4       Phase 5
    (US1: P1)     (US2: P2)     (US3: P3)
    --yes flag    Web UI config -- separator
        ↓             ↓             ↓
        └─────────────┼─────────────┘
                      ↓
                  Phase 6 (Polish)
```

**Critical Path**: Phase 1 → Phase 2 → Phase 3 (US1) → Phase 6

**Parallel Opportunities**:
- After Phase 2: US1, US2, US3 can be developed in parallel (different code paths)
- Within each phase: Tasks marked [P] can run in parallel

### MVP Scope (Minimum Viable Product)

**Recommended MVP**: Phase 1 + Phase 2 + Phase 3 (User Story 1 only)

**Rationale**:
- US1 (P1) solves the core problem: `--yes` flag not bypassing all prompts
- US2 (P2) and US3 (P3) are enhancements that can be added incrementally
- MVP delivers immediate value for automation use cases

### Incremental Delivery Strategy

1. **Sprint 1**: Phase 1 + Phase 2 + Phase 3 (US1) → MVP release
2. **Sprint 2**: Phase 4 (US2) → Web UI configuration
3. **Sprint 3**: Phase 5 (US3) → Client parameter pass-through
4. **Sprint 4**: Phase 6 → Documentation and polish

---

## Task Summary

**Total Tasks**: 71
**Test Tasks**: 25 (TDD required per constitution)
**Implementation Tasks**: 46

**Tasks by User Story**:
- Setup & Foundational: 15 tasks
- User Story 1 (P1): 12 tasks (8 tests + 4 implementation)
- User Story 2 (P2): 19 tasks (6 tests + 13 implementation)
- User Story 3 (P3): 12 tasks (5 tests + 7 implementation)
- Polish & Documentation: 13 tasks

**Parallel Opportunities**: 35 tasks marked [P] can run in parallel

**Independent Test Criteria**:
- US1: Run `zen --yes`, verify zero prompts
- US2: Enable Web UI config, run `zen`, verify auto-permission applied
- US3: Run `zen -- --verbose`, verify flag passed; run `zen --yes -- --permission-mode acceptEdits`, verify priority

**Estimated Effort**:
- MVP (US1): ~2-3 days
- Full Feature (US1+US2+US3): ~5-7 days
- With Documentation & Polish: ~7-10 days
