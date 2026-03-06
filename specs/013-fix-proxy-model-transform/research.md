# Research: Fix Proxy Model Transform

## R1: Path Deduplication Strategy

**Decision**: Strip `/v1` prefix from `targetPath` when `base_url.Path` already ends with `/v1`.

**Rationale**: The path duplication occurs because `TransformPath()` always returns absolute paths like `/v1/chat/completions`, and `singleJoiningSlash()` concatenates without overlap detection. The fix belongs in the URL construction in `forwardRequest()`, not in `TransformPath()`, because:
- `TransformPath` is a pure format converter — it shouldn't need to know about base URLs
- The overlap detection is inherently about URL construction, which is `forwardRequest`'s responsibility
- Non-transform paths (same-format pass-through) can also have this issue

**Alternatives considered**:
1. ~~Make `TransformPath` return relative paths~~ — breaks the `singleJoiningSlash` contract and all existing non-overlapping base URLs
2. ~~Pass base_url into TransformPath~~ — violates single responsibility; TransformPath converts formats, not constructs URLs
3. ~~Strip `/v1` from base_url on provider creation~~ — would break Anthropic providers that need `/v1` in the URL

## R2: Default Model Filling Strategy

**Decision**: Guard default model filling in `buildProviders()` with a `pc.GetType() == "anthropic"` check. Only Anthropic providers get Anthropic default model names.

**Rationale**: The current code unconditionally fills `claude-sonnet-4-5`, `claude-haiku-4-5`, etc. as defaults. For `type: "openai"` providers, these Anthropic model names are meaningless and actively harmful — they prevent `mapModel()` from falling through to the provider's own `model` field.

**Alternatives considered**:
1. ~~Fill OpenAI defaults (gpt-4o, gpt-4o-mini, etc.)~~ — which OpenAI model to default to is unknowable; different backends have different models
2. ~~Remove all defaults~~ — breaks backward compatibility for Anthropic providers where users expect defaults to work
3. ~~Add a separate `default_tier_model` config field~~ — over-engineering; YAGNI per constitution

## R3: Path Overlap for Non-Standard Base URLs

**Decision**: Only handle the `/v1` prefix specifically, not arbitrary path overlaps.

**Rationale**: All known API version prefixes use `/v1`. General path overlap detection (e.g., `base_url: .../api` + `/api/v1/...`) adds complexity for scenarios that don't exist in practice. The `TransformPath` function only ever returns paths starting with `/v1/`, so we only need to handle `/v1` overlap.

**Alternatives considered**:
1. ~~Generic prefix matching~~ — over-engineering for no practical benefit
2. ~~Regex-based path normalization~~ — unnecessary complexity

## R4: Existing Test Patterns

**Decision**: Follow the existing table-driven test pattern in `server_test.go` using `httptest.NewServer` and provider mocks.

**Rationale**: The existing test file has 60+ test functions following a consistent pattern. New tests should use the same structure for consistency and maintainability.

**Key patterns observed**:
- `TestSingleJoiningSlash` — table-driven with `{a, b, want}` tuples
- `TestModelMapping*` — create test HTTP server, construct Provider/ProxyServer, call `applyModelMapping`, verify output
- `TestServeHTTP*` — full integration with `httptest` server, verify responses
