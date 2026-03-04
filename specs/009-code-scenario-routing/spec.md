# Feature Specification: Code Scenario Routing

**Feature Branch**: `009-code-scenario-routing`
**Created**: 2026-03-04
**Status**: Draft
**Input**: User description: "如果我通过场景路由来要实现，深度思考、规划、planning使用某一个provider的其中一个模型，coding使用另一个provider的另一个模型，现在的功能够用吗？如果不够用，需要添加什么样的功能。"

## Background & Gap Analysis

GoZen currently supports 6 scenario types for routing: `think`, `image`, `longContext`, `webSearch`, `background`, and `default`. Each scenario can route to different providers with per-provider model overrides.

**What already works**: The `think` scenario correctly detects requests with `thinking.type=enabled` and routes them to a designated provider/model. This covers the "deep thinking / planning" use case.

**What's missing**: There is no `code` scenario. When Claude Code sends a regular coding request (no extended thinking, not a background Haiku task, no images), it falls through to `default`. Users cannot configure a separate provider or model specifically for coding tasks because there's no way to distinguish "coding" requests from other default traffic.

**The gap**: To route "thinking/planning" to one provider/model and "coding" to another, users need a `code` scenario that is detected automatically based on characteristics of coding-oriented requests.

## User Scenarios & Testing

### User Story 1 — Route Coding Requests to a Dedicated Provider/Model (Priority: P0) 🎯 MVP

A power user configures their GoZen profile so that extended thinking requests (planning, deep analysis) go to Provider A with an expensive, high-capability model, while regular coding requests (tool use, code generation without extended thinking) go to Provider B with a faster, cheaper model. This reduces cost while maintaining quality where it matters.

**Why this priority**: This is the core use case the user described — without this, the feature delivers no value.

**Independent Test**: Configure a profile with `think` route pointing to Provider A/Model X and `code` route pointing to Provider B/Model Y. Send a request with `thinking.type=enabled` and verify it goes to Provider A. Send a regular request with tool_use and verify it goes to Provider B.

**Acceptance Scenarios**:

1. **Given** a profile with `code` scenario routed to Provider B / `claude-sonnet-4`, **When** a regular API request (no thinking, no image, no web search) is sent, **Then** the request is routed to Provider B with model override `claude-sonnet-4`.

2. **Given** a profile with `think` scenario routed to Provider A / `claude-opus-4` and `code` scenario routed to Provider B / `claude-sonnet-4`, **When** a request with `thinking.type=enabled` is sent, **Then** the request is routed to Provider A with model override `claude-opus-4` (not Provider B).

3. **Given** a profile with only `code` scenario configured (no `think` route), **When** a request with extended thinking is sent, **Then** the request falls through to the default provider chain (not the `code` route).

4. **Given** a profile with `code` scenario routed to Provider B, **When** a background Haiku request is sent, **Then** the `background` scenario takes precedence and the request is NOT routed to Provider B.

5. **Given** a profile with `code` scenario configured, **When** a request containing image content is sent, **Then** the `image` scenario takes precedence over `code`.

---

### User Story 2 — Configure Code Scenario via Web UI (Priority: P1)

Users can configure the `code` scenario route through the Web UI profile editor, just like existing scenarios (think, image, longContext, etc.), selecting which providers and model overrides to use for coding requests.

**Why this priority**: Users need a way to configure this without editing JSON manually, but the feature works via JSON config even without the Web UI.

**Independent Test**: Open the profile edit page, navigate to the Routing tab, expand the "Code" scenario, add a provider with a model override, save, and verify the config is persisted to `zen.json`.

**Acceptance Scenarios**:

1. **Given** the profile edit page with the Routing tab open, **When** the user views available scenarios, **Then** a "Code" scenario appears alongside the existing scenarios (Think, Image, Long Context, Web Search, Background).

2. **Given** the Code scenario expanded, **When** the user adds Provider B with model override `claude-sonnet-4` and saves, **Then** the `zen.json` file contains the routing configuration for the `code` scenario with the specified provider and model.

---

### Edge Cases

- What happens when both `think` and `code` scenarios are configured and a request has extended thinking? → The `think` scenario has higher priority and is matched first (existing priority order applies).
- What happens when no `code` scenario is configured? → Requests fall through to `default` as they do today — fully backward compatible.
- What happens when a request has tool_use but also has thinking enabled? → The `think` scenario takes priority because thinking is checked before code in the priority order: webSearch > think > image > longContext > code > background > default.
- What happens when all providers in the `code` scenario fail? → Existing fallback behavior applies: the proxy falls back to the profile's default provider chain.

## Requirements

### Functional Requirements

- **FR-001**: System MUST support a new `code` scenario type alongside existing scenarios (think, image, longContext, webSearch, background, default).
- **FR-002**: System MUST automatically detect coding requests — defined as requests that are NOT think, NOT image, NOT webSearch, NOT longContext, and NOT background. The `code` detection MUST explicitly exclude background (Haiku) requests. The `code` scenario captures all "regular" non-specialized requests when configured.
- **FR-003**: System MUST maintain the existing scenario priority order, inserting `code` between `longContext` and `background`: webSearch > think > image > longContext > code > background > default. The `code` detection explicitly filters out Haiku requests before matching, so `background` requests are never captured by `code` even though `code` appears earlier in the priority chain.
- **FR-004**: When a `code` scenario route is configured, requests that would otherwise fall through to `default` MUST be routed to the `code` providers instead.
- **FR-005**: When NO `code` scenario route is configured, request routing MUST behave identically to today — full backward compatibility.
- **FR-006**: System MUST support per-provider model overrides for the `code` scenario, consistent with all other scenarios.
- **FR-007**: The Web UI profile editor MUST display the `code` scenario in the Routing tab with appropriate labeling (e.g., "Code / Default Coding").
- **FR-008**: The `code` scenario MUST be persisted in `zen.json` using the key `"code"` within the profile's `routing` map.

### Key Entities

- **Scenario**: Extended with a new `code` variant. Represents a detected request type that determines which provider chain handles the request.
- **ScenarioRoute**: Unchanged — already supports per-provider model overrides via `ProviderRoute{name, model}`.
- **ProfileConfig**: Unchanged structurally — the `routing` map already supports arbitrary `Scenario` keys.

## Success Criteria

### Measurable Outcomes

- **SC-001**: Users can configure separate providers/models for thinking vs. coding within a single profile, and requests are routed correctly 100% of the time based on scenario detection.
- **SC-002**: Feature is fully backward compatible — profiles without `code` routing configured behave identically to before.
- **SC-003**: The `code` scenario appears in the Web UI profile editor alongside existing scenarios.
- **SC-004**: All existing scenario routing tests continue to pass without modification.

## Assumptions

- The `code` scenario acts as a "catch-all" for non-specialized requests. It is only active when explicitly configured in the profile's routing map.
- The priority position of `code` is just above `default` — all other specialized scenarios (think, image, webSearch, longContext, background) take precedence.
- The existing `background` scenario detection (Haiku model requests) remains unchanged. `code` does NOT capture Haiku background tasks even though they could be considered "coding."
- No config version migration is needed because the `routing` map already uses `Scenario` (string) keys — adding a new scenario key `"code"` is fully backward compatible.

## Clarifications

### Session 2026-03-04

- Q: Priority order — should `code` be before or after `background` in the detection chain? → A: `code` before `background` (Option B). The `code` detection must explicitly exclude Haiku requests so background requests are never captured by the code catch-all. Priority: webSearch > think > image > longContext > code > background > default.
