# Implementation Plan: Provider & Model Tag in Proxy Responses

**Branch**: `005-provider-model-tag` | **Date**: 2026-03-02 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/005-provider-model-tag/spec.md`

## Summary

Add optional provider/model tag injection to proxy responses. When enabled via a global config setting (`show_provider_tag`), the proxy prepends `[provider: <name>, model: <model>]\n` to the first text content in successful responses. Supports both Anthropic Messages and OpenAI Chat Completions formats, in both streaming (SSE) and non-streaming modes. The tag is injected in `copyResponse()` after any format transformation. A Web UI toggle in General Settings controls the feature (default: OFF).

## Technical Context

**Language/Version**: Go 1.21+
**Primary Dependencies**: `net/http`, `encoding/json`, `bufio` (SSE parsing), React + TypeScript (Web UI)
**Storage**: JSON config at `~/.zen/zen.json` (new `show_provider_tag` boolean field, version 10 → 11)
**Testing**: `go test ./...` (table-driven, TDD per constitution), `npm run test` (Web UI)
**Target Platform**: macOS, Linux (CLI + daemon + Web UI)
**Project Type**: CLI + daemon (reverse proxy) + React Web UI
**Performance Goals**: Tag injection adds <5ms latency (SC-003)
**Constraints**: Tag injection after format transformation (FR-010), only on 2xx responses with text content (FR-007)
**Scale/Scope**: ~5 Go files modified, ~2 Web UI files modified, ~150-200 lines of production code + tests

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| I. TDD (NON-NEGOTIABLE) | PASS | Tests written first for tag injection (non-streaming, streaming), config, API, Web UI |
| II. Simplicity & YAGNI | PASS | Single boolean config, tag logic in existing `copyResponse()`, no new abstractions |
| III. Config Migration Safety | PASS | Version bump 10→11, `omitempty` bool defaults to `false`, no migration logic needed |
| IV. Branch Protection & Commit | PASS | Feature branch `005-provider-model-tag`, individual commits per user story |
| V. Minimal Artifacts | PASS | No summary docs created |
| VI. Test Coverage (NON-NEGOTIABLE) | PASS | Targets: internal/proxy ≥80%, internal/config ≥80%, internal/web ≥80%, web UI branch ≥70% |

**Post-design re-check**: All gates still pass. No new entities beyond a single boolean field. Tag injection logic lives in existing `copyResponse()` — no new packages or abstractions.

## Project Structure

### Documentation (this feature)

```text
specs/005-provider-model-tag/
├── plan.md              # This file
├── research.md          # Phase 0 — 7 research questions resolved
├── data-model.md        # Phase 1 — config field, API types
├── quickstart.md        # Phase 1 — build/test/manual verification
└── tasks.md             # Phase 2 output (/speckit.tasks)
```

### Source Code (files modified)

```text
internal/proxy/
├── server.go              # Tag injection in copyResponse() — non-streaming + SSE streaming
└── server_test.go         # Tests for tag injection (both formats, both modes)

internal/config/
├── config.go              # ShowProviderTag field in OpenCCConfig, version bump 10→11
├── store.go               # GetShowProviderTag() / SetShowProviderTag() methods
├── compat.go              # GetShowProviderTag() / SetShowProviderTag() convenience functions
└── config_test.go         # Config migration test (v10→v11)

internal/web/
├── api_settings.go        # Add ShowProviderTag to settings GET/PUT
└── api_settings_test.go   # Test settings API with new field

web/src/
├── types/api.ts           # Add show_provider_tag to Settings interface
├── pages/settings/tabs/GeneralSettings.tsx  # Add toggle Switch
└── i18n/locales/*.json    # Translation keys for toggle label
```

**Structure Decision**: Modifications to existing files only. Tag injection logic in `copyResponse()` (server.go). Config field in existing `OpenCCConfig`. Web UI toggle in existing GeneralSettings tab. No new packages or files.

## Key Design Decisions

### Non-Streaming Tag Injection

After format transformation in `copyResponse()`:
1. Check if `config.GetShowProviderTag()` is enabled
2. Check if response is 2xx (only inject on success)
3. Parse response body JSON
4. For Anthropic format: find first `{type: "text", text: "..."}` block in `content` array, prepend tag
5. For OpenAI format: prepend tag to `choices[0].message.content`
6. Re-marshal and write with updated Content-Length

### SSE Streaming Tag Injection

After optional StreamTransformer in `copyResponse()`:
1. Wrap the reader with a tag-injecting reader
2. The wrapper buffers SSE events, extracting the model from early events (`message_start` for Anthropic, first chunk for OpenAI)
3. On first text delta event, prepend tag to the text content
4. Pass all subsequent events through unmodified

### Model Extraction

- **Non-streaming**: Extract `model` field from response body JSON (both Anthropic and OpenAI include it)
- **Streaming Anthropic**: Extract from `message_start` event's `message.model`
- **Streaming OpenAI**: Extract from first chunk's `model` field
- Provider name always from `p.Name`
