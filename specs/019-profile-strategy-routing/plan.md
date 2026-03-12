# Implementation Plan: Profile Strategy-Aware Provider Routing

**Branch**: `019-profile-strategy-routing` | **Date**: 2026-03-09 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/019-profile-strategy-routing/spec.md`

**Note**: This template is filled in by the `/speckit.plan` command. See `.specify/templates/plan-template.md` for the execution workflow.

## Summary

Connect profile strategy configuration to real provider selection by implementing strategy-aware routing that evaluates profile strategy (least-latency, least-cost, round-robin, weighted) before selecting a provider for each request. The system will track provider latency metrics (average over last 100 requests, minimum 10 samples required), maintain round-robin state per profile (in-memory, resets on restart), and use read-only metric snapshots for concurrent request safety. Each strategy decision will be logged with provider name, strategy type, and selection reason.

## Technical Context

**Language/Version**: Go 1.21+
**Primary Dependencies**: `net/http`, `sync`, `time`, `encoding/json` (stdlib only); existing `internal/config`, `internal/proxy` packages
**Storage**: SQLite (existing LogDB at `~/.zen/logs.db`) for latency metrics persistence; in-memory ring buffer for round-robin state
**Testing**: Go testing (`go test`), table-driven tests, race detector (`-race`), coverage target ≥80%
**Target Platform**: macOS, Linux, Windows (cross-platform daemon)
**Project Type**: CLI tool with embedded HTTP proxy daemon
**Performance Goals**: Strategy evaluation <5ms per request, support 100 concurrent requests, 24-hour uptime
**Constraints**: No external dependencies beyond stdlib, backward-compatible config migration, zero downtime during config reload
**Scale/Scope**: 10-50 providers per profile, 1000 requests/hour typical load, 100 concurrent requests peak

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### Principle I: Test-Driven Development (TDD) - ✅ PASS
- **Requirement**: Write tests before implementation, maintain ≥80% coverage
- **Compliance**: Plan includes comprehensive test strategy with unit tests for strategy evaluation, integration tests for provider selection, and race condition tests for concurrent access
- **Evidence**: Test coverage targets defined in Technical Context, TDD workflow enforced in task ordering

### Principle II: Simplicity First - ✅ PASS
- **Requirement**: Avoid premature abstraction, prefer direct solutions
- **Compliance**: Strategy evaluation uses simple switch statement, no complex patterns or frameworks introduced
- **Evidence**: LoadBalancer already exists with strategy enum, extending existing pattern rather than creating new abstraction

### Principle III: Explicit Over Implicit - ✅ PASS
- **Requirement**: Clear naming, explicit error handling, no magic
- **Compliance**: Strategy decisions logged explicitly with provider name and reason, metric snapshots explicitly created per request
- **Evidence**: FR-016 requires logging each decision, FR-017 requires explicit read-only snapshots

### Principle IV: Fail Fast, Fail Clearly - ✅ PASS
- **Requirement**: Validate early, return descriptive errors
- **Compliance**: Insufficient sample size (< 10 requests) causes provider exclusion with clear logging, invalid strategy falls back to ordered failover
- **Evidence**: FR-011 defines fallback behavior, edge cases document validation rules

### Principle V: Minimize State, Maximize Clarity - ✅ PASS
- **Requirement**: Prefer stateless, document state carefully
- **Compliance**: Round-robin state is minimal (single counter), explicitly documented as in-memory only, resets on restart
- **Evidence**: FR-014 explicitly states "in-memory only, resets on daemon restart", clarification Q4 confirms no persistence

### Principle VI: Composition Over Inheritance - ✅ PASS
- **Requirement**: Favor interfaces and composition
- **Compliance**: Strategy evaluation composes existing LoadBalancer with LogDB metrics, no inheritance introduced
- **Evidence**: LoadBalancer.Select() already exists, extending with profile-aware routing via composition

### Principle VII: Document Decisions, Not Code - ✅ PASS
- **Requirement**: Explain why, not what
- **Compliance**: Clarifications document time window choice (100 requests), persistence decision (in-memory), concurrency approach (snapshots)
- **Evidence**: Clarifications section in spec.md explains rationale for each decision

### Principle VIII: Daemon Proxy Stability Priority (NON-NEGOTIABLE) - ✅ PASS
- **Requirement**: All proxy issues are blocking, 24-hour uptime, 100 concurrent requests
- **Compliance**: Strategy evaluation <5ms target ensures no latency regression, read-only snapshots prevent race conditions, existing failover preserved
- **Evidence**: SC-005 defines 5ms target, FR-017 ensures concurrency safety, FR-009 preserves failover behavior

**GATE STATUS**: ✅ ALL PRINCIPLES SATISFIED - Proceed to Phase 0

## Project Structure

### Documentation (this feature)

```text
specs/019-profile-strategy-routing/
├── plan.md              # This file (/speckit.plan command output)
├── research.md          # Phase 0 output (/speckit.plan command)
├── data-model.md        # Phase 1 output (/speckit.plan command)
├── quickstart.md        # Phase 1 output (/speckit.plan command)
├── contracts/           # Phase 1 output (/speckit.plan command)
│   └── strategy-api.md  # Strategy evaluation interface contract
└── tasks.md             # Phase 2 output (/speckit.tasks command - NOT created by /speckit.plan)
```

### Source Code (repository root)

```text
internal/
├── config/
│   └── config.go                    # ProfileConfig.Strategy already exists (LoadBalanceStrategy enum)
├── proxy/
│   ├── loadbalancer.go              # MODIFY: Add profile-aware Select() method
│   ├── loadbalancer_test.go         # MODIFY: Add strategy evaluation tests
│   ├── profile_proxy.go             # MODIFY: Pass profile strategy to LoadBalancer
│   ├── profile_proxy_test.go        # MODIFY: Add integration tests for strategy routing
│   ├── logdb.go                     # MODIFY: Add GetProviderLatencyMetrics() method
│   ├── logdb_test.go                # MODIFY: Add latency query tests
│   ├── metrics.go                   # EXISTING: Already tracks latency per provider
│   └── provider.go                  # NO CHANGE: Provider struct unchanged
└── web/
    └── api.go                       # NO CHANGE: Strategy configured via existing profile API

tests/
├── integration/
│   └── strategy_routing_test.go     # NEW: End-to-end strategy routing tests
└── unit/
    └── strategy_evaluation_test.go  # NEW: Unit tests for strategy logic
```

**Structure Decision**: Single project structure (Option 1) with modifications to existing `internal/proxy` package. No new packages required - strategy evaluation logic integrates into existing LoadBalancer. Config schema already supports `ProfileConfig.Strategy` field (added in v1.4.0), so no migration needed. Tests follow existing pattern: unit tests in `*_test.go` files alongside implementation, integration tests in `tests/integration/`.

## Complexity Tracking

> **No violations - this section intentionally left empty**

All constitution principles satisfied without exceptions. No complexity justification required.
