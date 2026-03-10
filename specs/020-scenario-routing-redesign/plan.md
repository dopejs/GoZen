# Implementation Plan: Scenario Routing Architecture Redesign

**Branch**: `020-scenario-routing-redesign` | **Date**: 2026-03-10 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/020-scenario-routing-redesign/spec.md`

**Note**: This template is filled in by the `/speckit.plan` command. See `.specify/templates/plan-template.md` for the execution workflow.

## Summary

**Implementation Strategy**: Complete refactoring of the scenario routing system. Existing scenario detection code (`internal/proxy/scenario.go`) will be replaced with a new architecture designed from the ground up.

Redesign the scenario routing system to be protocol-agnostic (supporting Anthropic Messages, OpenAI Chat, and OpenAI Responses), middleware-extensible (allowing explicit routing decisions via RoutingDecision API), and support open scenario namespaces (custom route keys without code changes). The system will normalize requests into a common semantic representation, allow middleware to drive routing decisions, support per-scenario routing policies (strategy, weights, thresholds), and provide strong config validation with comprehensive observability.

**Key Architectural Changes**:
1. Replace fixed `Scenario` enum with open string-based scenario keys (type alias for backward compatibility)
2. Replace `ScenarioRoute` with new `RoutePolicy` structure supporting per-scenario strategies
3. Add protocol normalization layer for Anthropic Messages, OpenAI Chat, and OpenAI Responses
4. Enable middleware to drive routing decisions via `RoutingDecision` API
5. Migrate config from v14 to v15 with automatic conversion

## Technical Context

**Language/Version**: Go 1.21+
**Primary Dependencies**:
- `net/http` (stdlib) - HTTP server and client
- `encoding/json` (stdlib) - JSON parsing and serialization
- `sync` (stdlib) - Concurrency primitives for routing state
- `github.com/pkoukk/tiktoken-go` - Token counting for long-context detection
- Existing internal packages: `internal/config`, `internal/proxy`, `internal/middleware`

**Storage**:
- JSON config at `~/.zen/zen.json` (existing config store with versioning)
- SQLite LogDB at `~/.zen/logs.db` (existing, for latency metrics)
- In-memory routing state (session cache, round-robin counters)

**Testing**:
- Go stdlib `testing` package
- Table-driven tests in `*_test.go` files
- Integration tests in `tests/integration/`
- Test coverage thresholds: 80% for `internal/proxy`, 80% for `internal/config`

**Target Platform**:
- macOS, Linux, Windows (cross-platform CLI daemon)
- Runs as background daemon process

**Project Type**: CLI tool with embedded HTTP proxy daemon

**Performance Goals**:
- Support 100 concurrent requests (existing limiter)
- Request routing decision p95 < 5ms overhead
- Protocol normalization p95 < 10ms per request
- Total routing overhead p95 < 20ms per request
- 24-hour uptime without degradation

**Constraints**:
- Backward compatibility with existing routing config required (v14 → v15 migration)
- Must not break existing middleware pipeline
- Config migration must be automatic and lossless
- Daemon proxy stability is P0 (all issues blocking per Constitution VIII)
- Complete refactoring: existing scenario detection code will be replaced, not modified

**Scale/Scope**:
- 3 supported protocols (Anthropic Messages, OpenAI Chat, OpenAI Responses)
- 7 builtin scenarios (think, image, longContext, webSearch, code, background, default) + unlimited custom scenarios
- 5 existing load balancing strategies (failover, round-robin, least-latency, least-cost, weighted)
- Complete refactoring: 8-12 new source files, 15-20 modified files
- Config version bump: v14 → v15

**Key Design Decisions** (finalized 2026-03-10):
1. **Scenario Key Naming**: Support camelCase, kebab-case, and snake_case; normalize internally to camelCase (e.g., web-search→webSearch, long_context→longContext)
2. **Scenario Type**: `type Scenario = string` (type alias, not enum) with constants for builtin scenarios
3. **Config Structure**: New `RoutePolicy` type replacing `ScenarioRoute`, includes per-scenario strategy/weights/threshold
4. **Protocol Detection**: Priority order: URL path → X-Zen-Client header → body structure → default to openai_chat
5. **Implementation Strategy**: Complete refactoring (replace existing scenario.go, not modify)

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### Principle I: Test-Driven Development ✅ PASS
- **Requirement**: New features MUST use TDD (write tests first, verify fail, implement)
- **Compliance**: Plan includes comprehensive test strategy with table-driven tests in existing `*_test.go` files
- **Action**: Will write tests for normalization, classification, routing resolution, and config validation before implementation

### Principle II: Simplicity & YAGNI ✅ PASS
- **Requirement**: Minimum needed for current task, no speculative abstractions
- **Compliance**: Design focuses on solving identified problems (protocol-agnostic routing, middleware extensibility) without adding unnecessary features
- **Action**: Will avoid over-engineering; each component serves a clear requirement from spec

### Principle III: Config Migration Safety ✅ PASS
- **Requirement**: Schema changes MUST bump version, add migration logic, include tests
- **Compliance**: Plan includes config version bump (v14 → v15) and migration from `ScenarioRoute` to `RoutePolicy`
- **Action**: Will implement `UnmarshalJSON` with v14 format detection and automatic conversion, add comprehensive migration tests

### Principle IV: Branch Protection & Commit Discipline ✅ PASS
- **Requirement**: All changes via PR, atomic commits, tag-driven releases
- **Compliance**: Working on feature branch `020-scenario-routing-redesign`, will create PR to `feat/v3.0.1`
- **Action**: Will commit each logical unit (normalization, classifier, config, etc.) separately

### Principle V: Minimal Artifacts ✅ PASS
- **Requirement**: No summary docs, no example configs in root, design docs in `.dev/`
- **Compliance**: Architecture doc already in `docs/` (user-facing), plan in `specs/` (standard location)
- **Action**: Will not create unnecessary documentation files

### Principle VI: Test Coverage Enforcement ✅ PASS
- **Requirement**: Must meet CI thresholds (80% for `internal/proxy`, `internal/config`)
- **Compliance**: Plan includes comprehensive test coverage for all new code
- **Action**: Will verify coverage locally before pushing: `go test -cover ./internal/proxy ./internal/config`

### Principle VII: Automated Testing Priority ✅ PASS
- **Requirement**: Automated tests preferred, integration tests for daemon features
- **Compliance**: Plan includes unit tests, integration tests for routing flow, protocol normalization tests
- **Action**: Will write integration tests in `tests/integration/` for end-to-end routing scenarios

### Principle VIII: Daemon Proxy Stability Priority ✅ PASS
- **Requirement**: Daemon proxy is P0, all issues blocking, strictest standards
- **Compliance**: This feature directly impacts daemon proxy routing core; treating all issues as blocking
- **Action**: Will apply strictest review standards, comprehensive test coverage, no shortcuts

**GATE STATUS**: ✅ ALL CHECKS PASS - Proceeding to Phase 0 Research

## Project Structure

### Documentation (this feature)

```text
specs/[###-feature]/
├── plan.md              # This file (/speckit.plan command output)
├── research.md          # Phase 0 output (/speckit.plan command)
├── data-model.md        # Phase 1 output (/speckit.plan command)
├── quickstart.md        # Phase 1 output (/speckit.plan command)
├── contracts/           # Phase 1 output (/speckit.plan command)
└── tasks.md             # Phase 2 output (/speckit.tasks command - NOT created by /speckit.plan)
```

### Source Code (repository root)

```text
internal/
├── proxy/
│   ├── routing_normalize.go          # NEW: Protocol normalization (Anthropic/OpenAI Chat/Responses)
│   ├── routing_normalize_test.go     # NEW: Normalization tests for all protocols
│   ├── routing_classifier.go         # NEW: Builtin scenario classifier on normalized requests
│   ├── routing_classifier_test.go    # NEW: Classifier tests
│   ├── routing_decision.go           # NEW: RoutingDecision and RoutingHints types
│   ├── routing_resolver.go           # NEW: Route policy resolution logic
│   ├── routing_resolver_test.go      # NEW: Resolution tests
│   ├── scenario.go                   # DEPRECATED: Will be replaced by routing_classifier.go
│   ├── scenario_test.go              # DEPRECATED: Will be replaced by routing_classifier_test.go
│   ├── profile_proxy.go              # MODIFIED: Integrate new routing flow
│   ├── profile_proxy_test.go         # MODIFIED: Update tests
│   ├── server.go                     # MODIFIED: Populate RequestContext with routing fields
│   ├── server_test.go                # MODIFIED: Update tests
│   ├── loadbalancer.go               # MODIFIED: Accept route-specific overrides
│   └── loadbalancer_test.go          # MODIFIED: Update tests
│
├── config/
│   ├── config.go                     # MODIFIED: New RoutePolicy type, Scenario as string alias
│   ├── store.go                      # MODIFIED: Config validation for routing, v14→v15 migration
│   ├── compat.go                     # MODIFIED: Legacy config migration helpers
│   └── config_test.go                # MODIFIED: Migration and validation tests
│
└── middleware/
    └── interface.go                  # MODIFIED: Add NormalizedRequest, RoutingDecision, RoutingHints to RequestContext

tests/
└── integration/
    ├── routing_protocol_test.go      # NEW: Protocol-agnostic routing tests
    ├── routing_middleware_test.go    # NEW: Middleware-driven routing tests
    └── routing_policy_test.go        # NEW: Per-scenario policy tests

web/src/
├── types/api.ts                      # MODIFIED: Update Scenario type, add RoutePolicy
└── pages/profiles/edit.tsx           # MODIFIED: Support custom scenario keys

tui/
└── routing.go                        # MODIFIED: Support custom scenario keys
```

**Structure Decision**: Complete refactoring approach. New routing-specific files in `internal/proxy/` (routing_*.go pattern) replace existing `scenario.go`. Config types in `internal/config/` updated to use `RoutePolicy`. Integration tests in `tests/integration/` for end-to-end routing validation. TUI and Web UI updated to support open scenario namespace.

## Complexity Tracking

> **No violations** - This feature follows all constitution principles without requiring exceptions.
