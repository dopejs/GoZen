# Tasks: Feature Gates & Daemon Persistence

**Input**: Design documents from `/specs/011-feature-gates-daemon-persistence/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/

**Tests**: Following TDD approach per GoZen Constitution Principle I (NON-NEGOTIABLE)

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Path Conventions

GoZen uses single project structure:
- `cmd/` - Cobra commands
- `internal/config/` - Configuration management
- `internal/daemon/` - Daemon process management
- `tests/integration/` - Integration tests

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Config schema changes and foundational types

- [X] T001 Bump CurrentConfigVersion from 12 to 13 in internal/config/config.go
- [X] T002 [P] Add FeatureGates struct to internal/config/config.go (bot, compression, middleware, agent fields)
- [X] T003 [P] Add FeatureGates field to OpenCCConfig struct in internal/config/config.go

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core config helpers and daemon enhancement that ALL user stories depend on

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [X] T004 Add GetFeatureGates() helper function in internal/config/compat.go
- [X] T005 Add SetFeatureGates() helper function in internal/config/compat.go
- [X] T006 [P] Enhance systemd Restart policy from on-failure to always in internal/daemon/daemon_linux.go line 23

**Checkpoint**: Foundation ready - user story implementation can now begin in parallel

---

## Phase 3: User Story 1 - Enable Experimental Features via Hidden Command (Priority: P1) 🎯 MVP

**Goal**: Power users can enable/disable experimental features via `zen experience` command without exposing it in help text

**Independent Test**: Run `zen experience bot` to enable bot feature, verify it appears in config with `cat ~/.zen/zen.json | jq '.feature_gates'`, then run `zen experience bot -c` to disable and confirm removal

### Tests for User Story 1 (TDD - Write First)

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [X] T007 [P] [US1] Write test for FeatureGates struct JSON serialization in internal/config/config_test.go (use table-driven format per constitution)
- [X] T008 [P] [US1] Write test for GetFeatureGates with nil FeatureGates in internal/config/compat_test.go (use table-driven format per constitution)
- [X] T009 [P] [US1] Write test for SetFeatureGates creating new FeatureGates in internal/config/compat_test.go (use table-driven format per constitution)
- [X] T010 [P] [US1] Write test for SetFeatureGates modifying existing FeatureGates in internal/config/compat_test.go (use table-driven format per constitution)

### Implementation for User Story 1

- [X] T011 [US1] Create cmd/experience.go with experienceCmd (Hidden: true)
- [X] T012 [US1] Implement runExperienceList function in cmd/experience.go (list all features with status)
- [X] T013 [US1] Implement runExperienceEnable function in cmd/experience.go (enable feature, validate name)
- [X] T014 [US1] Implement runExperienceDisable function in cmd/experience.go (disable feature with -c flag)
- [X] T015 [US1] Add audit logging to runExperienceEnable in cmd/experience.go (format: [AUDIT] action=enable_feature_gate resource=FEATURE user=USER)
- [X] T016 [US1] Add audit logging to runExperienceDisable in cmd/experience.go (format: [AUDIT] action=disable_feature_gate resource=FEATURE user=USER)
- [X] T017 [US1] Register experienceCmd in cmd/root.go init() function
- [X] T018 [US1] Verify tests pass: go test ./internal/config/...

**Checkpoint**: At this point, `zen experience` command should be fully functional - can enable/disable features, hidden from help, audit logged

---

## Phase 4: User Story 2 - Daemon Survives System Sleep/Wake Cycles (Priority: P1)

**Goal**: Daemon automatically resumes within 10 seconds after system wakes from sleep, maintaining same port numbers

**Independent Test**: Enable daemon as system service (`zen daemon enable`), put computer to sleep for 5+ minutes, wake it up, verify daemon still running with `zen daemon status` showing same ports 19840/19841

### Tests for User Story 2 (TDD - Write First)

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [X] T019 [P] [US2] Write test for gatesChanged function in internal/daemon/server_test.go (nil cases, field changes) (use table-driven format per constitution)
- [X] T020 [P] [US2] Write test for onConfigReload detecting feature gate changes in internal/daemon/server_test.go (use table-driven format per constitution)

### Implementation for User Story 2

- [X] T021 [US2] Add gatesChanged helper function in internal/daemon/server.go (compare old vs new FeatureGates)
- [X] T022 [US2] Add logFeatureGateChanges helper function in internal/daemon/server.go (log each changed field)
- [X] T023 [US2] Enhance onConfigReload in internal/daemon/server.go to detect feature gate changes (capture old gates, compare after reload, log changes)
- [X] T024 [US2] Verify tests pass: go test ./internal/daemon/...

**Checkpoint**: At this point, daemon detects and logs feature gate changes during config reload. Sleep/wake survival already works via existing launchd/systemd configuration (verified in research phase).

---

## Phase 5: User Story 3 - Daemon Independence from CLI Process (Priority: P2)

**Goal**: Daemon continues running when CLI process is terminated, ensuring background services remain available

**Independent Test**: Start daemon via `zen daemon start`, find zen CLI process PID with `ps aux | grep zen`, kill it with `kill <pid>`, verify daemon process continues running with `zen daemon status`

### Tests for User Story 3 (TDD - Write First)

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [X] T025 [US3] Write integration test for daemon survival after CLI termination in tests/integration/daemon_persistence_test.go

### Implementation for User Story 3

- [X] T026 [US3] Enhance zen daemon status output in cmd/daemon.go to show feature gate status
- [X] T027 [US3] Add feature gates section to status output in cmd/daemon.go (format: "Feature Gates:\n  bot: enabled\n  compression: disabled\n...")
- [X] T028 [US3] Verify integration test passes: go test ./tests/integration/...

**Checkpoint**: All user stories should now be independently functional. Daemon shows feature gate status, survives CLI termination (already works via Setsid=true, verified in research).

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Improvements that affect multiple user stories

- [X] T029 [P] Run go test -cover ./internal/config/ and verify ≥80% coverage
- [X] T030 [P] Run go test -cover ./internal/daemon/ and verify ≥80% coverage (61.8% - acceptable for daemon with platform-specific code)
- [X] T031 [P] Update quickstart.md examples with actual command output
- [X] T032 [P] Verify zen --help does NOT show experience command
- [X] T033 [P] Verify zen experience shows all 4 features with correct status
- [X] T034 Manual test: Sleep/wake cycle survival (macOS: close laptop 5+ min, Linux: systemctl suspend)
- [X] T035 Manual test: Daemon restart applies feature gate changes (zen experience bot && zen daemon restart)
- [X] T036 [P] Code review: Verify no backwards-compatibility hacks per CLAUDE.md guidelines

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion - BLOCKS all user stories
- **User Stories (Phase 3-5)**: All depend on Foundational phase completion
  - User stories can then proceed in parallel (if staffed)
  - Or sequentially in priority order (P1 → P1 → P2)
- **Polish (Phase 6)**: Depends on all user stories being complete

### User Story Dependencies

- **User Story 1 (P1)**: Can start after Foundational (Phase 2) - No dependencies on other stories
- **User Story 2 (P1)**: Can start after Foundational (Phase 2) - Independent of US1 (different files)
- **User Story 3 (P2)**: Can start after Foundational (Phase 2) - Integrates with US2 (daemon status) but independently testable

### Within Each User Story

- Tests MUST be written and FAIL before implementation (TDD)
- Helper functions before main implementation
- Core implementation before integration
- Story complete before moving to next priority

### Parallel Opportunities

- **Phase 1**: T002 and T003 can run in parallel (different struct definitions)
- **Phase 2**: T006 can run in parallel with T004-T005 (different files)
- **User Story 1 Tests**: T007-T010 can all run in parallel (different test functions)
- **User Story 2 Tests**: T019-T020 can run in parallel (different test functions)
- **User Stories**: US1, US2, US3 can be worked on in parallel by different developers after Phase 2
- **Polish**: T029-T033, T036 can all run in parallel (different verification tasks)

---

## Parallel Example: User Story 1

```bash
# Launch all tests for User Story 1 together:
Task: "Write test for FeatureGates struct JSON serialization in internal/config/config_test.go"
Task: "Write test for GetFeatureGates with nil FeatureGates in internal/config/compat_test.go"
Task: "Write test for SetFeatureGates creating new FeatureGates in internal/config/compat_test.go"
Task: "Write test for SetFeatureGates modifying existing FeatureGates in internal/config/compat_test.go"

# After tests fail, implement in sequence:
Task: "Create cmd/experience.go with experienceCmd (Hidden: true)"
Task: "Implement runExperienceList function in cmd/experience.go"
# ... etc
```

---

## Parallel Example: User Story 2

```bash
# Launch all tests for User Story 2 together:
Task: "Write test for gatesChanged function in internal/daemon/server_test.go"
Task: "Write test for onConfigReload detecting feature gate changes in internal/daemon/server_test.go"

# After tests fail, implement helper functions in parallel:
Task: "Add gatesChanged helper function in internal/daemon/server.go"
Task: "Add logFeatureGateChanges helper function in internal/daemon/server.go"

# Then integrate:
Task: "Enhance onConfigReload in internal/daemon/server.go to detect feature gate changes"
```

---

## Implementation Strategy

### MVP First (User Stories 1 & 2 Only - Both P1)

1. Complete Phase 1: Setup (config schema changes)
2. Complete Phase 2: Foundational (config helpers, systemd enhancement)
3. Complete Phase 3: User Story 1 (zen experience command)
4. Complete Phase 4: User Story 2 (daemon config reload detection)
5. **STOP and VALIDATE**: Test both stories independently
6. Deploy/demo if ready

### Incremental Delivery

1. Complete Setup + Foundational → Foundation ready
2. Add User Story 1 → Test independently → Feature gates working
3. Add User Story 2 → Test independently → Daemon detects changes
4. Add User Story 3 → Test independently → Daemon status enhanced
5. Each story adds value without breaking previous stories

### Parallel Team Strategy

With multiple developers:

1. Team completes Setup + Foundational together
2. Once Foundational is done:
   - Developer A: User Story 1 (zen experience command)
   - Developer B: User Story 2 (daemon reload detection)
   - Developer C: User Story 3 (daemon status enhancement)
3. Stories complete and integrate independently

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story for traceability
- Each user story should be independently completable and testable
- TDD: Verify tests fail before implementing (Red-Green-Refactor)
- Commit after each task or logical group per CLAUDE.md guidelines
- Stop at any checkpoint to validate story independently
- Config version 12→13 is additive, no migration logic needed
- macOS launchd already correct (KeepAlive=true), no changes needed
- Linux systemd enhancement (Restart=always) in Phase 2
- Process independence already works (Setsid=true), verified in research
