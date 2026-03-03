# Implementation Plan: Revert Provider Tag & Add Request Monitoring UI

**Branch**: `006-revert-tag-add-monitoring` | **Date**: 2026-03-03 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/006-revert-tag-add-monitoring/spec.md`

**Note**: This template is filled in by the `/speckit.plan` command. See `.specify/templates/plan-template.md` for the execution workflow.

## Summary

Remove the provider tag injection feature (005-provider-model-tag) that causes API errors and persistent data pollution, and replace it with a Web UI request monitoring page. The monitoring page displays real-time request metadata (provider, model, timing, tokens, cost) without modifying API responses, providing visibility into provider usage through a dedicated interface rather than injecting content into responses.

## Technical Context

**Language/Version**: Go 1.21+
**Primary Dependencies**: `net/http`, `encoding/json`, `sync` (for in-memory buffer), Vanilla JS (Web UI)
**Storage**: In-memory ring buffer (default 1000 requests), no database persistence for MVP
**Testing**: `go test ./...` (table-driven tests), manual Web UI testing
**Target Platform**: macOS, Linux (CLI + daemon + Web UI)
**Project Type**: CLI + daemon (reverse proxy) + Vanilla JS Web UI
**Performance Goals**: Request metadata capture adds <5ms overhead, Web UI loads history in <2 seconds
**Constraints**: Must maintain backward compatibility with existing configs containing deprecated `show_provider_tag` field
**Scale/Scope**: ~8 Go files modified (proxy, config, web API), ~3 Web UI files modified (new page + navigation), ~200-300 lines of production code + tests

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| I. TDD (NON-NEGOTIABLE) | PASS | Tests written first for: tag removal verification, request buffer, API endpoint, Web UI integration |
| II. Simplicity & YAGNI | PASS | In-memory buffer (no database), polling (no WebSockets), minimal abstractions |
| III. Config Migration Safety | PASS | Deprecated `show_provider_tag` field ignored on load (no version bump needed, backward compatible) |
| IV. Branch Protection & Commit | PASS | Feature branch `006-revert-tag-add-monitoring`, individual commits per task |
| V. Minimal Artifacts | PASS | No summary docs created, planning docs in specs/ directory |
| VI. Test Coverage (NON-NEGOTIABLE) | PASS | Targets: internal/proxy ≥80%, internal/config ≥80%, internal/web ≥80% |

**Post-design re-check**: All gates still pass. No new entities beyond in-memory request buffer. Removal of existing code reduces complexity.

## Project Structure

### Documentation (this feature)

```text
specs/006-revert-tag-add-monitoring/
├── plan.md              # This file
├── research.md          # Phase 0 — research questions resolved
├── data-model.md        # Phase 1 — request record structure
├── quickstart.md        # Phase 1 — build/test/manual verification
└── tasks.md             # Phase 2 output (/speckit.tasks)
```

### Source Code (files modified)

```text
internal/proxy/
├── server.go              # Remove tag injection code from copyResponse()
├── server_test.go         # Remove tag injection tests
├── request_monitor.go     # NEW: In-memory request buffer + capture logic
└── request_monitor_test.go # NEW: Tests for request monitoring

internal/config/
├── config.go              # Remove ShowProviderTag field (keep for backward compat parsing)
├── store.go               # Remove GetShowProviderTag() / SetShowProviderTag() methods
├── compat.go              # Remove GetShowProviderTag() / SetShowProviderTag() convenience functions
└── config_test.go         # Update tests to verify deprecated field is ignored

internal/web/
├── api_requests.go        # NEW: GET /api/v1/requests endpoint
├── api_requests_test.go   # NEW: Tests for requests API
├── api_settings.go        # Remove ShowProviderTag from settings GET/PUT
└── api_settings_test.go   # Update tests to remove ShowProviderTag

web/src/
├── pages/requests/        # NEW: Requests monitoring page
│   └── RequestsPage.jsx   # NEW: Main monitoring page component
├── components/            # Existing navigation component
│   └── Navigation.jsx     # Update to add "Requests" link
├── types/api.ts           # Remove show_provider_tag from Settings interface
└── App.jsx                # Add route for /requests page
```

**Structure Decision**: Modifications to existing files for removal, new files for monitoring feature. Request monitoring logic in new `internal/proxy/request_monitor.go`. Web UI adds new page under `web/src/pages/requests/`. No new packages or major architectural changes.

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| [e.g., 4th project] | [current need] | [why 3 projects insufficient] |
| [e.g., Repository pattern] | [specific problem] | [why direct DB access insufficient] |
