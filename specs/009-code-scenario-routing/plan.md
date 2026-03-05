# Implementation Plan: Code Scenario Routing

**Branch**: `009-code-scenario-routing` | **Date**: 2026-03-04 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/009-code-scenario-routing/spec.md`

## Summary

Add a new `code` scenario type to GoZen's scenario-based routing system. The `code` scenario acts as a configurable catch-all for regular coding requests (non-think, non-image, non-webSearch, non-longContext, non-background), enabling users to route thinking/planning requests to one provider/model and coding requests to another. The implementation touches the Go scenario detection engine, config types, TUI routing editor, Web UI profile editor, and i18n labels.

## Technical Context

**Language/Version**: Go 1.21+
**Primary Dependencies**: Cobra (CLI), Bubble Tea + Lip Gloss (TUI), React + TypeScript + Vite (Web UI)
**Storage**: JSON at `~/.zen/zen.json` — no schema migration needed (string-keyed routing map)
**Testing**: `go test ./...` (Go), `pnpm run test:coverage` in `web/` (frontend)
**Target Platform**: macOS, Linux (CLI + local web server)
**Project Type**: CLI tool with embedded web UI
**Performance Goals**: N/A — `code` detection is zero-cost (negative check: "nothing else matched")
**Constraints**: No config version bump needed. Backward compatible.
**Scale/Scope**: ~8 files modified, ~50 net lines of Go, ~20 net lines of TypeScript

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| I. TDD (NON-NEGOTIABLE) | ✅ PASS | Tests first for `DetectScenario` code path and priority. Add to existing `scenario_test.go`. |
| II. Simplicity & YAGNI | ✅ PASS | Minimal change: 1 const, 1 detection check, labels. No new abstractions. |
| III. Config Migration Safety | ✅ PASS | No schema change. `routing` map uses `Scenario` (string) keys — `"code"` is just a new key value. No version bump. |
| IV. Branch Protection & Commit | ✅ PASS | Work on feature branch, atomic commits per task. |
| V. Minimal Artifacts | ✅ PASS | No summary docs. Only spec-required plan artifacts. |
| VI. Test Coverage (NON-NEGOTIABLE) | ✅ PASS | Must maintain 80%+ for `internal/proxy`, `internal/config`. Frontend coverage ≥70%. |

No violations. No complexity tracking needed.

## Project Structure

### Documentation (this feature)

```text
specs/009-code-scenario-routing/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
├── contracts/           # Phase 1 output (no new APIs — internal changes only)
└── tasks.md             # Phase 2 output (via /speckit.tasks)
```

### Source Code (repository root)

```text
internal/
├── config/
│   └── config.go          # Add ScenarioCode constant (line ~244)
└── proxy/
    ├── scenario.go        # Add isCodeRequest() check, update DetectScenario priority
    └── scenario_test.go   # TDD: new tests for code scenario detection

tui/
└── routing.go             # Add code entry to knownScenarios slice (line ~55)

web/src/
├── types/api.ts           # Add 'code' to Scenario type, SCENARIOS array, SCENARIO_LABELS
├── i18n/locales/
│   ├── en.json            # Add scenarioCode label
│   ├── zh-CN.json         # Add scenarioCode label
│   └── zh-TW.json         # Add scenarioCode label
└── pages/profiles/edit.tsx # No changes needed (already renders from SCENARIOS array)
```

**Structure Decision**: All changes fit within existing directory structure. No new files needed except tests (added to existing `scenario_test.go`). The web UI profile editor (`edit.tsx`) dynamically renders from the `SCENARIOS` array and `SCENARIO_LABELS` map — adding `'code'` to these automatically makes it appear in the UI.

## Complexity Tracking

> No violations to justify. All changes are minimal and follow existing patterns.
