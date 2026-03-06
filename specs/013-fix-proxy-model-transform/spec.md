# Feature Specification: Fix Proxy Model Transform for Cross-Format Providers

**Feature Branch**: `013-fix-proxy-model-transform`
**Created**: 2026-03-06
**Status**: Draft
**Input**: User description: "反向代理在做 transform 时存在问题。cctq-codex provider 配置了 type: openai，使用 zen -p codex -c codex 可以正常工作，但使用 zen -p codex（默认 claude 客户端）时返回错误：'There's an issue with the selected model (claude-sonnet-4-6)'"

## Clarifications

### Session 2026-03-06

- Q: cctq-codex 后端 API 是什么格式？ → A: OpenAI 兼容（type 配置为 "openai"）
- Q: 根因确认 → A: Dev 环境实测发现两个独立 bug：(1) 路径拼接双重 /v1，(2) buildProviders 为 OpenAI provider 填充 Anthropic 默认模型名

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Claude Client Through OpenAI Provider with base_url Containing /v1 (Priority: P1)

A user configures an OpenAI-type provider with `base_url` that already includes `/v1` (e.g., `https://code.b886.top/v1`). When the Claude client sends a request to `/v1/messages`, the proxy transforms the path to `/v1/chat/completions`. The path concatenation must not produce a double `/v1` in the final URL.

**Why this priority**: This causes a 404 error that completely breaks the provider. Dev environment testing confirmed the URL becomes `https://code.b886.top/v1/v1/chat/completions`.

**Independent Test**: Send a request through a provider with `/v1` in base_url and verify the final URL is `https://code.b886.top/v1/chat/completions` (single `/v1`).

**Acceptance Scenarios**:

1. **Given** a provider with `base_url: "https://example.com/v1"` and `type: "openai"`, **When** Claude sends to `/v1/messages`, **Then** the final URL is `https://example.com/v1/chat/completions` (no path duplication).
2. **Given** a provider with `base_url: "https://example.com"` and `type: "openai"`, **When** Claude sends to `/v1/messages`, **Then** the final URL is `https://example.com/v1/chat/completions` (correctly appended).
3. **Given** a provider with `base_url: "https://example.com/v1"` and `type: "openai"`, **When** Codex sends to `/v1/chat/completions`, **Then** the final URL is `https://example.com/v1/chat/completions` (no path duplication).

---

### User Story 2 - OpenAI Provider Without Tier-Specific Models (Priority: P1)

A user configures an OpenAI-type provider with only a default `model` field (no `sonnet_model`, `haiku_model`, etc.). When any client sends a request, the model should map to the provider's default `model`, not to an Anthropic default model name.

**Why this priority**: `buildProviders()` currently fills empty tier-specific fields with Anthropic model names (e.g., `claude-sonnet-4-5`) regardless of provider type. This causes `mapModel()` to return an Anthropic model name to an OpenAI backend, which rejects it.

**Independent Test**: Send a request with `model: claude-sonnet-4-6` through a provider that only has `model: "gpt-5.3-codex"` and verify the model is mapped to `gpt-5.3-codex`.

**Acceptance Scenarios**:

1. **Given** an OpenAI provider with only `model: "gpt-5.3-codex"` (no tier-specific models), **When** Claude sends `model: "claude-sonnet-4-6"`, **Then** the model maps to `gpt-5.3-codex` (not `claude-sonnet-4-5`).
2. **Given** an OpenAI provider with only `model: "gpt-5.3-codex"`, **When** Claude sends `model: "claude-opus-4-6"`, **Then** the model maps to `gpt-5.3-codex`.
3. **Given** an OpenAI provider with `model: "gpt-5.3-codex"` AND `sonnet_model: "gpt-5.4"`, **When** Claude sends `model: "claude-sonnet-4-6"`, **Then** the model maps to `gpt-5.4` (explicit tier takes priority).
4. **Given** an Anthropic provider with no tier-specific models, **When** Claude sends `model: "claude-sonnet-4-6"`, **Then** the default Anthropic model names are still used (backward compatibility).

---

### User Story 3 - End-to-End Cross-Format Request/Response (Priority: P2)

When a Claude client (Anthropic format) sends a request through an OpenAI-type provider, the full pipeline must work: model mapping, request transformation, path transformation, response transformation — including streaming responses.

**Why this priority**: Even after fixing bugs 1 and 2, the complete Anthropic↔OpenAI transformation pipeline must be validated end-to-end.

**Independent Test**: Send a non-streaming and streaming Anthropic request through an OpenAI provider and verify successful responses are correctly transformed back to Anthropic format.

**Acceptance Scenarios**:

1. **Given** a working OpenAI provider, **When** Claude sends a non-streaming Anthropic request, **Then** the response is successfully transformed to Anthropic format with correct `id`, `model`, `content`, and `usage` fields.
2. **Given** a working OpenAI provider, **When** Claude sends a streaming Anthropic request, **Then** the SSE events are correctly transformed from OpenAI to Anthropic format.

---

### Edge Cases

- What happens when `base_url` ends with trailing slash `https://example.com/v1/`? Path deduplication should still work.
- What happens when `base_url` contains a path beyond `/v1` (e.g., `https://example.com/api/v1`)? Only the matching `/v1` prefix should be deduplicated.
- What happens when a provider has an empty string for `model`? The original model name should be preserved unchanged.
- What happens when the request body contains no `model` field? The body should be passed through unchanged.
- What happens when the upstream returns an error (4xx/5xx)? The error response should be properly transformed and propagated.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: Path transformation MUST NOT produce duplicate path segments when `base_url` already includes the API version prefix (e.g., `/v1`).
- **FR-002**: When a provider of `type: "openai"` has empty tier-specific model fields, the proxy MUST NOT fill them with Anthropic default model names. Instead, they should remain empty so `mapModel()` falls through to the provider's default `model`.
- **FR-003**: For providers of `type: "anthropic"`, existing behavior of filling Anthropic defaults for empty tier fields MUST be preserved (backward compatibility).
- **FR-004**: Model mapping MUST apply before API format transformation in the request processing pipeline.
- **FR-005**: The proxy MUST correctly transform both request and response formats for cross-format scenarios (Anthropic client ↔ OpenAI provider).
- **FR-006**: The proxy MUST log model mapping transformations and path transformations at debug level via the existing `log.Printf` logger for debugging.
- **FR-007**: The proxy MUST correctly handle both streaming (SSE) and non-streaming responses in cross-format transformation.

### Key Entities

- **Provider**: Has a `type` (anthropic/openai), a `base_url`, a default `model`, and optional tier-specific models. Default model filling behavior should vary by provider type.
- **Path Transformation**: Converts between API endpoint paths (`/v1/messages` ↔ `/v1/chat/completions`) while respecting the provider's `base_url` structure.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Claude client can successfully send requests and receive responses through an OpenAI-type provider, regardless of whether `base_url` includes `/v1`.
- **SC-002**: OpenAI-type providers with only a default `model` field correctly use that model for all request types.
- **SC-003**: Existing Anthropic provider configurations continue to work with zero regressions.
- **SC-004**: Both streaming and non-streaming responses are correctly transformed in cross-format scenarios.

## Assumptions

- **Confirmed via dev testing**: Two independent bugs identified:
  1. **Path duplication** (`server.go:473` + `transform.go:TransformPath`): `singleJoiningSlash(base_url, transformedPath)` produces double `/v1` when `base_url` already has `/v1`. Dev test showed `https://code.b886.top/v1/v1/chat/completions` → 404.
  2. **Default model filling** (`profile_proxy.go:186-201`): `buildProviders()` fills empty tier-specific models with Anthropic defaults for ALL providers, including `type: "openai"` ones. This prevents `mapModel()` from falling through to `p.Model`.
- The `cctq-codex` provider config has all tier-specific models explicitly set, so bug 2 doesn't directly affect it. But `yunyi-codex` (only `model: "gpt-5.3-codex"`) is affected.
- `-c codex` works because with OpenAI→OpenAI, no path transformation occurs and the Codex client sends directly to `/v1/chat/completions` which, combined with `singleJoiningSlash`, may still produce a working URL depending on the backend's tolerance.
- Testing done exclusively via dev environment (`~/.zen-dev/zen.json`, ports 29840/29841).
