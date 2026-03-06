# Implementation Plan: Feature Gates & Daemon Persistence

**Branch**: `011-feature-gates-daemon-persistence` | **Date**: 2026-03-05 | **Spec**: [spec.md](./spec.md)

## Summary

Implement a hidden feature gate mechanism (`zen experience` command) to control experimental features (bot, compression, middleware, agent) and verify/enhance daemon persistence to ensure survival across sleep/wake cycles and process independence. The feature gates will be stored in `~/.zen/zen.json` and changes will trigger daemon config reload with audit logging.

## Technical Context

**Language/Version**: Go 1.21+
**Primary Dependencies**: Cobra (CLI), Bubble Tea + Lip Gloss (TUI), React + TypeScript + Vite (Web UI)
**Storage**: JSON config at `~/.zen/zen.json` (version 12 → 13)
**Testing**: Go test (table-driven), vitest (Web UI)
**Target Platform**: macOS (launchd), Linux (systemd)
**Project Type**: CLI tool with daemon process
**Performance Goals**: Feature toggle < 5 seconds, daemon resume < 10 seconds after wake
**Constraints**: Hidden command (not in help text), zero downtime config reload, audit logging required
**Scale/Scope**: 4 experimental features, 2 platform-specific daemon implementations

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### Pre-Research Check ✅

- **Principle I (TDD)**: ✅ Will write tests first for feature gate CRUD operations, config migration, and daemon reload
- **Principle II (Simplicity)**: ✅ Minimal implementation - flat boolean struct, no abstractions, reuses existing config reload
- **Principle III (Config Migration)**: ✅ Version bump 12→13, additive change (no migration logic needed)
- **Principle IV (Branch Protection)**: ✅ Feature branch workflow, atomic commits per task
- **Principle V (Minimal Artifacts)**: ✅ No summary docs, planning artifacts in `.dev/` only
- **Principle VI (Test Coverage)**: ✅ Must maintain 80%+ for internal/config, internal/daemon

**Violations**: None

### Post-Design Check ✅

**Re-evaluation after Phase 1 design:**

- **Principle I (TDD)**: ✅ Test structure defined in contracts, unit tests for config/daemon packages
- **Principle II (Simplicity)**: ✅ Design remains minimal - flat struct, hidden command, existing reload pattern
- **Principle III (Config Migration)**: ✅ Version 12→13 confirmed additive, no migration logic needed
- **Principle IV (Branch Protection)**: ✅ Feature branch workflow maintained
- **Principle V (Minimal Artifacts)**: ✅ Planning docs in specs/, no summary files created
- **Principle VI (Test Coverage)**: ✅ Test strategy defined, targets 80%+ for config/daemon packages

**Violations**: None

**Design Complexity Assessment**:
- Config schema: Simple flat boolean struct (4 fields)
- CLI interface: Single hidden command with 2 operations (enable/disable)
- Daemon changes: One-line systemd enhancement, existing reload pattern
- No new abstractions, no speculative features, no unnecessary complexity

**Ready for Phase 2 (Task Generation)**: ✅

## Project Structure

### Documentation (this feature)

```text
specs/011-feature-gates-daemon-persistence/
├── plan.md              # This file
├── research.md          # Phase 0 output (feature gates + daemon persistence research)
├── data-model.md        # Phase 1 output (FeatureGates entity)
├── quickstart.md        # Phase 1 output (user guide for zen experience command)
└── contracts/           # Phase 1 output (CLI command interface)
    └── experience-command.md
```

### Source Code (repository root)

```text
cmd/
├── experience.go        # NEW: Hidden zen experience command (enable/disable/list)
├── daemon.go            # MODIFY: Add feature gate change detection to status output
└── root.go              # MODIFY: Register experience command (hidden)

internal/
├── config/
│   ├── config.go        # MODIFY: Add FeatureGates struct, bump version to 13
│   ├── config_test.go   # MODIFY: Add feature gate tests
│   ├── compat.go        # MODIFY: Add GetFeatureGates/SetFeatureGates helpers
│   └── compat_test.go   # MODIFY: Add helper tests
├── daemon/
│   ├── daemon_darwin.go # VERIFY: KeepAlive=true is sufficient (no changes needed)
│   ├── daemon_linux.go  # MODIFY: Change Restart=on-failure to Restart=always
│   ├── server.go        # MODIFY: Add feature gate change detection to onConfigReload
│   └── server_test.go   # MODIFY: Add config reload tests with feature gate changes

tests/
└── integration/
    └── daemon_persistence_test.go  # NEW: Integration tests for sleep/wake simulation
```

**Structure Decision**: Single Go project with existing cmd/ and internal/ structure. Feature gates integrate into existing config system. Daemon persistence enhancements are minimal (one-line change for Linux systemd).

## Complexity Tracking

No violations - all changes align with constitution principles.

