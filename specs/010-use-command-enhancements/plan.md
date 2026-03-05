# Implementation Plan: zen use Command Enhancements

**Branch**: `010-use-command-enhancements` | **Date**: 2026-03-05 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/010-use-command-enhancements/spec.md`

**Note**: This template is filled in by the `/speckit.plan` command. See `.specify/templates/plan-template.md` for the execution workflow.

## Summary

Enhance the `zen` CLI to support improved permission control and client parameter pass-through. The primary requirements are:

1. **Fix `--yes` flag**: Change from `acceptEdits` to `bypassPermissions` for Claude Code, and `-a never` for Codex to eliminate all permission prompts
2. **Web UI auto-permission config**: Add per-client permission mode configuration in Web UI with client-specific options (no abstraction)
3. **Client parameter pass-through**: Support `--` separator to pass arbitrary parameters to underlying clients with proper priority handling

**Priority order**: `--` parameters > `--yes` flag > Web UI config > default behavior

## Technical Context

**Language/Version**: Go 1.21+
**Primary Dependencies**: Cobra (CLI framework), existing `cmd/root.go` permission handling
**Storage**: JSON config at `~/.zen/zen.json` (schema version bump required)
**Testing**: Go standard testing (`go test ./...`), table-driven tests in existing `*_test.go` files
**Target Platform**: macOS, Linux, Windows (cross-platform CLI)
**Project Type**: CLI tool
**Performance Goals**: Command execution <1 second, no impact on client startup time
**Constraints**: Must maintain backward compatibility with existing configs, config migration required
**Scale/Scope**: Single-user CLI tool, affects 3 client types (Claude Code, Codex, OpenCode)

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### Principle I: Test-Driven Development ✅
- **Status**: PASS
- **Plan**: Write tests first for permission flag logic, config migration, and parameter priority handling
- **Coverage targets**: `internal/config` (80%), `cmd/root.go` permission logic (covered by integration tests)

### Principle II: Simplicity & YAGNI ✅
- **Status**: PASS
- **Approach**: Minimal changes to existing `prependAutoApproveArgs()` function, no new abstractions
- **Justification**: Reusing existing permission flag injection pattern, only changing the mode values

### Principle III: Config Migration Safety ✅
- **Status**: REQUIRES ATTENTION
- **Action**: Bump `CurrentConfigVersion` from 11 to 12 (current version is 11 per existing codebase)
- **Migration**: Add new fields `claude_auto_permission`, `codex_auto_permission`, `opencode_auto_permission` with default values
- **Tests**: Must cover old config parsing, new field defaults, round-trip marshaling

### Principle IV: Branch Protection & Commit Discipline ✅
- **Status**: PASS
- **Plan**: Feature branch `010-use-command-enhancements` → PR to main → merge after review

### Principle V: Minimal Artifacts ✅
- **Status**: PASS
- **Plan**: No summary docs, no example configs in root, design docs in `.dev/` if needed

### Principle VI: Test Coverage Enforcement ✅
- **Status**: PASS
- **Targets**:
  - `internal/config`: 80% (currently 85.6%, adding new fields)
  - `cmd/root.go`: Covered by integration tests
  - Web UI (`web/`): 70% branch coverage (adding permission config UI)
- **Verification**: Run `go test -cover ./internal/config` and `pnpm run test:coverage` in `web/` before PR

**Overall Gate Status**: ✅ PASS (with config migration attention required)

## Project Structure

### Documentation (this feature)

```text
specs/010-use-command-enhancements/
├── spec.md              # Feature specification (completed)
├── plan.md              # This file (/speckit.plan command output)
├── research.md          # Phase 0 output (to be generated)
├── data-model.md        # Phase 1 output (to be generated)
├── quickstart.md        # Phase 1 output (to be generated)
├── contracts/           # Phase 1 output (CLI command contracts)
├── checklists/          # Quality validation checklists
│   └── requirements.md  # Spec quality checklist (completed)
└── tasks.md             # Phase 2 output (/speckit.tasks command - NOT created by /speckit.plan)
```

### Source Code (repository root)

```text
cmd/
├── root.go              # Main CLI entry, permission flag handling (MODIFY)
├── use.go               # zen use command (MODIFY if needed)
└── [other commands]

internal/
├── config/
│   ├── config.go        # Config types, version bump (MODIFY)
│   ├── store.go         # Config persistence (MODIFY)
│   ├── config_test.go   # Config tests (ADD tests)
│   └── migration logic  # Config migration (ADD if needed)
└── [other packages]

web/
├── src/
│   ├── components/
│   │   └── Settings/    # Settings page (MODIFY)
│   │       └── PermissionConfig.tsx  # New permission config component (ADD)
│   ├── types/
│   │   └── config.ts    # TypeScript types (MODIFY)
│   └── api/
│       └── config.ts    # API client (MODIFY if needed)
└── tests/
    └── components/
        └── PermissionConfig.test.tsx  # Component tests (ADD)

docs/
├── README.md            # English docs (UPDATE)
├── README.zh-CN.md      # Chinese simplified (UPDATE)
├── README.zh-TW.md      # Chinese traditional (UPDATE)
└── README.es.md         # Spanish (UPDATE)
```

**Structure Decision**: This is a CLI tool with embedded web UI. Changes span:
1. **Backend (Go)**: `cmd/root.go` for flag handling, `internal/config` for config schema
2. **Frontend (React)**: `web/src/components/Settings` for permission configuration UI
3. **Documentation**: All four README translations must be updated per constitution

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

No violations. All constitution principles are satisfied.

## Phase 0: Research (COMPLETED)

**Output**: `research.md`

**Key Findings**:
1. OpenCode permission behavior needs investigation (marked as TBD)
2. Permission flag detection strategy: Simple string scanning
3. Config migration: Add optional fields with defaults, no complex migration
4. Web UI implementation: Client-aware dropdown with no abstraction

**Resolved Questions**: 3/4
**Remaining**: OpenCode permission handling (to be resolved during implementation)

---

## Phase 1: Design & Contracts (COMPLETED)

**Outputs**:
- `data-model.md`: Config schema with AutoPermissionConfig type, version bump to 12
- `contracts/cli-commands.md`: CLI command contracts, API contracts, priority order guarantees
- `quickstart.md`: User guide and developer implementation checklist

**Key Decisions**:
1. **Config Schema**: Add `claude_auto_permission`, `codex_auto_permission`, `opencode_auto_permission` fields
2. **Priority Order**: `--` > `--yes` > Web UI > default (strictly enforced)
3. **Permission Modes**: No abstraction, client-specific values stored separately
4. **Migration**: Version 11 → 12, backward and forward compatible
5. **OpenCode Handling**: Assume auto-approves by default, verify during implementation (T022)

**Agent Context**: Updated `CLAUDE.md` with new technology stack information

---

## Phase 2: Task Breakdown (NEXT STEP)

**Command**: `/speckit.tasks`

**Expected Output**: `tasks.md` with dependency-ordered implementation tasks

**Task Categories**:
1. Backend (Go): Config schema, permission logic, migration
2. Frontend (React): PermissionConfig component, Settings page integration
3. Testing: Unit tests, integration tests, component tests
4. Documentation: README updates (all 4 languages), help text

---

## Implementation Notes

### Critical Path

1. **Config Schema** (blocking): Must be implemented first, all other work depends on it
2. **Permission Logic** (blocking): Core functionality, required for both CLI and Web UI
3. **Web UI** (parallel): Can be developed in parallel with CLI once config schema is done
4. **Documentation** (final): Update after implementation is complete

### Risk Areas

1. **OpenCode Support**: Permission behavior unknown, may require additional research
2. **Config Migration**: Must be thoroughly tested to avoid breaking existing installations
3. **Priority Resolution**: Complex logic with multiple sources, needs careful testing
4. **Web UI Validation**: Client-specific validation rules must match backend

### Testing Strategy

- **TDD Required**: Per constitution, write tests first
- **Coverage Targets**: 
  - `internal/config`: 80% (currently 85.6%)
  - Web UI: 70% branch coverage
- **Test Types**:
  - Unit: Permission logic, config migration, priority resolution
  - Integration: CLI behavior, end-to-end workflows
  - Component: React components, user interactions

---

## Success Criteria Verification

**From Spec**:
- ✅ SC-001: Zero prompts with `--yes` flag (testable via integration tests)
- ✅ SC-002: <1 second execution (performance test)
- ✅ SC-003: 100% parameter forwarding (integration test)
- ✅ SC-004: Persistent config benefit (manual verification)
- ✅ SC-005: <2 second error feedback (integration test)

**All criteria are measurable and testable.**

---

## Next Steps

1. Run `/speckit.tasks` to generate task breakdown
2. Review and approve tasks
3. Run `/speckit.implement` to execute implementation
4. Verify test coverage meets thresholds
5. Update all four README translations
6. Open PR for review

---

**Planning Phase Complete**: ✅
**Ready for Task Generation**: ✅
