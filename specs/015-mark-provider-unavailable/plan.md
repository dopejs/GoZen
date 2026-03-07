# Implementation Plan: Manual Provider Unavailability Marking

**Branch**: `015-mark-provider-unavailable` | **Date**: 2026-03-07 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/015-mark-provider-unavailable/spec.md`

## Summary

Add manual provider unavailability marking with three duration types (today, month, permanent). Users mark providers via Web UI or CLI (`zen disable/enable`). The proxy skips marked providers during routing; returns 503 error when all providers are unavailable. Markings persist in `zen.json` config (new `disabled_providers` map, config version 14). Scenario routes fall back to profile defaults when all route providers are disabled.

## Technical Context

**Language/Version**: Go 1.21+
**Primary Dependencies**: Cobra (CLI), net/http (proxy/web), React + TypeScript + Vite (Web UI)
**Storage**: JSON config at `~/.zen/zen.json` (version 13 → 14)
**Testing**: `go test ./...` (Go), `pnpm run test:coverage` (Web UI, vitest)
**Target Platform**: macOS/Linux CLI + Web UI (local daemon)
**Project Type**: CLI + Web service (proxy daemon with management interfaces)
**Performance Goals**: Zero overhead on provider selection (single map lookup per provider)
**Constraints**: No background timers for expiration; lazy evaluation at request time
**Scale/Scope**: Typically 2-10 providers per user; negligible data size

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| I. TDD (NON-NEGOTIABLE) | PASS | Tests written first for config, proxy, web API, CLI |
| II. Simplicity & YAGNI | PASS | Minimal new types; lazy expiration; no background workers |
| III. Config Migration Safety | PASS | Version 13→14; new `disabled_providers` field with omitempty; no migration code needed (empty map default) |
| IV. Branch Protection & Commit Discipline | PASS | Feature branch `015-mark-provider-unavailable`; individual commits per task |
| V. Minimal Artifacts | PASS | No summary docs; design docs in specs/ (gitignored or feature-specific) |
| VI. Test Coverage (NON-NEGOTIABLE) | PASS | New code in config (80%), proxy (80%), web (80%+) must meet CI thresholds |

**Post-Phase 1 Re-check**: All principles still satisfied. No complexity violations.

## Project Structure

### Documentation (this feature)

```text
specs/015-mark-provider-unavailable/
├── plan.md              # This file
├── spec.md              # Feature specification
├── research.md          # Phase 0: research decisions
├── data-model.md        # Phase 1: entity definitions
├── quickstart.md        # Phase 1: dev quickstart guide
├── contracts/
│   └── api.md           # Phase 1: API & CLI contracts
├── checklists/
│   └── requirements.md  # Specification quality checklist
└── tasks.md             # Phase 2 output (created by /speckit.tasks)
```

### Source Code (repository root)

```text
internal/config/
├── config.go            # MODIFY: Add UnavailableMarking struct, bump version 13→14, add DisabledProviders to OpenCCConfig
├── store.go             # MODIFY: Add DisableProvider(), EnableProvider(), GetDisabledProviders(), IsProviderDisabled() methods
└── config_test.go       # MODIFY: Add tests for new marking types and store methods

internal/proxy/
├── server.go            # MODIFY: Add unavailability filtering in tryProviders() via config.IsProviderDisabled() and all-disabled error in ServeHTTP()
└── server_test.go       # MODIFY: Add tests for disabled provider skipping and error response

internal/web/
├── server.go            # MODIFY: Register new disable/enable API routes
├── api_providers.go     # MODIFY: Add disable/enable handlers, extend provider list response
└── api_providers_test.go # MODIFY: Add tests for new endpoints

cmd/
├── disable.go           # NEW: zen disable command
├── enable.go            # NEW: zen enable command
└── root.go              # MODIFY: Register disable/enable commands

web/src/
├── hooks/use-providers.ts   # MODIFY: Add disable/enable API calls and disabled status
└── [pages]                  # MODIFY: Add disable/enable UI controls and status indicators
```

**Structure Decision**: Follows existing project structure. All modifications are in existing directories. Two new files (`cmd/disable.go`, `cmd/enable.go`) follow the established Cobra command pattern.
