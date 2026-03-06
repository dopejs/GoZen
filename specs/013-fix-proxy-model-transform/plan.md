# Implementation Plan: Fix Proxy Model Transform for Cross-Format Providers

**Branch**: `013-fix-proxy-model-transform` | **Date**: 2026-03-06 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/013-fix-proxy-model-transform/spec.md`

## Summary

Fix two independent bugs in the reverse proxy that prevent Claude clients from working with OpenAI-type providers:
1. **Path duplication**: `TransformPath` returns `/v1/chat/completions` which gets double-joined with `base_url` ending in `/v1`, producing `/v1/v1/chat/completions` → 404.
2. **Default model filling**: `buildProviders()` fills empty tier-specific models with Anthropic defaults for ALL providers, including `type: "openai"` ones, preventing `mapModel()` from falling through to the provider's default model.

## Technical Context

**Language/Version**: Go 1.21+
**Primary Dependencies**: `net/http`, `encoding/json`, `net/url` (proxy); existing `internal/proxy/transform` package
**Storage**: JSON config at `~/.zen/zen.json` (no schema changes needed)
**Testing**: `go test ./internal/proxy/...`, `go test ./internal/proxy/transform/...`
**Target Platform**: darwin/linux (CLI tool)
**Project Type**: CLI with embedded HTTP proxy
**Performance Goals**: N/A (bug fix, no performance change)
**Constraints**: Must maintain backward compatibility with existing Anthropic provider configs
**Scale/Scope**: 3 files changed, ~30 lines modified, ~100 lines of new tests

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| I. TDD (NON-NEGOTIABLE) | ✅ PASS | Tests written first for both bugs, then implementation |
| II. Simplicity & YAGNI | ✅ PASS | Minimal changes: condition guard in buildProviders, path dedup in forwardRequest |
| III. Config Migration Safety | ✅ PASS | No config schema changes needed |
| IV. Branch Protection | ✅ PASS | Working on feature branch, will PR to main |
| V. Minimal Artifacts | ✅ PASS | No extra docs beyond spec artifacts |
| VI. Test Coverage (NON-NEGOTIABLE) | ✅ PASS | New tests for both bugs; coverage targets: proxy ≥80%, transform ≥80% |

## Project Structure

### Documentation (this feature)

```text
specs/013-fix-proxy-model-transform/
├── spec.md              # Feature specification
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
└── contracts/           # N/A (internal bug fix)
```

### Source Code (files to modify)

```text
internal/proxy/
├── profile_proxy.go     # Bug 2: buildProviders() default model filling
├── server.go            # Bug 1: forwardRequest() path deduplication
├── server_test.go       # Tests for both bugs
└── transform/
    ├── transform.go     # TransformPath() — may need adjustment
    └── transform_test.go # Path transform tests
```

**Structure Decision**: Existing file structure. No new files needed — changes go into existing source and test files.

## Complexity Tracking

No constitution violations. No complexity tracking needed.

## Design Decisions

### Bug 1: Path Deduplication

**Approach**: Strip overlapping `/v1` prefix from `targetPath` when `base_url` already ends with `/v1`. This is done in `forwardRequest()` before calling `singleJoiningSlash()`.

**Why here and not in TransformPath**: `TransformPath` is format-agnostic and shouldn't know about base_url. The deduplication belongs in the URL construction logic that combines base_url + path.

**Algorithm**:
```
baseURLPath = p.BaseURL.Path  (e.g., "/v1" or "/api/v1")
if targetPath starts with baseURLPath:
    targetPath = targetPath[len(baseURLPath):]  (strip the overlapping prefix)
```

This handles:
- `base_url: .../v1` + `/v1/chat/completions` → `/chat/completions` → joined = `.../v1/chat/completions` ✓
- `base_url: ...` + `/v1/chat/completions` → no overlap → joined = `.../v1/chat/completions` ✓
- `base_url: .../v1/` + `/v1/chat/completions` → overlap after trimming slash → `.../v1/chat/completions` ✓
- `base_url: .../api/v1` + `/v1/chat/completions` → `/v1` overlaps with `/api/v1`? No, `/v1/chat/completions` doesn't start with `/api/v1` → no strip → `.../api/v1/v1/chat/completions`... This is still wrong.

**Revised approach**: Use `strings.TrimPrefix` to strip the known `/v1` prefix from `targetPath` when the base URL path already ends with `/v1`:

```
basePath = strings.TrimSuffix(p.BaseURL.Path, "/")
if strings.HasSuffix(basePath, "/v1") && strings.HasPrefix(targetPath, "/v1") {
    targetPath = targetPath[3:]  // strip "/v1", keep "/chat/completions"
}
```

This is more targeted and only deduplicates the `/v1` version prefix specifically.

### Bug 2: Type-Aware Default Model Filling

**Approach**: In `buildProviders()`, only fill Anthropic defaults when provider type is `"anthropic"` (or empty, which defaults to anthropic). For `type: "openai"` providers, leave empty tier-specific fields as empty strings. This lets `mapModel()` fall through to `p.Model` correctly.

**Code change**:
```go
isAnthropic := pc.GetType() == config.ProviderTypeAnthropic

reasoningModel := pc.ReasoningModel
if reasoningModel == "" && isAnthropic {
    reasoningModel = "claude-sonnet-4-5-thinking"
}
// ... same pattern for haiku, opus, sonnet
```

**Backward compatibility**: Anthropic providers keep the same default behavior. Only OpenAI providers change.

## Implementation Order

1. **Write tests first** (TDD per constitution):
   - Test: path dedup with various base_url patterns
   - Test: buildProviders type-aware defaults
   - Test: model mapping fallthrough for OpenAI providers
   - Verify tests fail (Red phase)

2. **Fix Bug 1**: Path deduplication in `server.go:forwardRequest()`

3. **Fix Bug 2**: Type-aware defaults in `profile_proxy.go:buildProviders()`

4. **Verify tests pass** (Green phase)

5. **Dev environment E2E test**: Restart dev daemon, test with curl

6. **Coverage check**: `go test -cover ./internal/proxy/...`
