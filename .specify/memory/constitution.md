<!--
Sync Impact Report
==================
Version change: 1.1.0 → 1.2.0 (MINOR — updated web frontend constraint, tooling reference)

Modified principles: none

Modified sections:
  - Technology & Architecture Constraints: Web frontend updated from
    "Vanilla JS" to "React + TypeScript + Vite" to reflect migration
    completed in feature 006-revert-tag-add-monitoring.
  - Principle VI: Changed `npm run test:coverage` to
    `pnpm run test:coverage` to match project tooling (CLAUDE.md).

Added sections: none

Removed sections: none

Templates requiring updates:
  - .specify/templates/plan-template.md ✅ no changes needed
  - .specify/templates/spec-template.md ✅ no changes needed
  - .specify/templates/tasks-template.md ✅ no changes needed
  - .specify/templates/checklist-template.md ✅ no changes needed

Follow-up TODOs: none
-->

# GoZen Constitution

## Core Principles

### I. Test-Driven Development (NON-NEGOTIABLE)

- New features MUST use TDD: write tests first, verify they fail,
  then implement to make them pass (Red-Green-Refactor).
- Tests MUST be table-driven and placed in existing `*_test.go` files
  when possible. New test files are only permitted for new public APIs
  or sufficiently complex logic.
- `go test ./...` MUST pass before any commit or release tag.
- Rationale: The codebase manages API proxy failover and config
  migration — correctness is critical. TDD catches regressions early
  and serves as living documentation.

### II. Simplicity & YAGNI

- Every change MUST be the minimum needed for the current task.
  No speculative abstractions, no feature flags for hypothetical
  futures, no helpers for one-time operations.
- Three similar lines of code are preferable to a premature
  abstraction. Error handling and validation MUST only exist at
  system boundaries (user input, external APIs, config file parsing).
- Unused code (variables, exports, compatibility shims) MUST be
  deleted, not commented out or renamed with underscore prefixes.
- Rationale: GoZen is a CLI tool where users expect fast startup and
  predictable behavior. Unnecessary complexity slows development and
  increases the surface area for bugs.

### III. Config Migration Safety

- Any change to `OpenCCConfig` or nested types that alters the JSON
  schema MUST bump `CurrentConfigVersion` in
  `internal/config/config.go`.
- Migration logic MUST be added so older config files parse correctly
  (current pattern: custom `UnmarshalJSON` with `json.RawMessage`
  fallback).
- Tests MUST cover: old format parsing, mixed old/new format, field
  preservation on the fallback path, and marshal round-trip.
- Rationale: Users depend on `~/.zen/zen.json` persisting across
  upgrades. A broken migration means lost provider credentials and
  proxy configurations.

### IV. Branch Protection & Commit Discipline

- The `main` branch is protected. All changes MUST go through a pull
  request — direct pushes to main are prohibited.
- Each completed task or fix MUST be committed individually, not
  batched into large commits.
- Releases MUST be triggered by pushing a git tag (e.g., `v2.x.0`).
  `gh release create` MUST NOT be used — the CI pipeline handles
  release creation.
- Before tagging a release: verify `Version` in `cmd/root.go` matches
  the tag, check for unpushed commits, and update all four README
  files (`README.md`, `docs/README.zh-CN.md`, `docs/README.zh-TW.md`,
  `docs/README.es.md`).
- Rationale: Atomic commits enable clean reverts, and tag-driven CI
  ensures reproducible builds.

### V. Minimal Artifacts

- Summary/explanation markdown files MUST NOT be created. Commit
  messages and code comments are the documentation.
- Example config files (JSON, YAML) MUST NOT exist in the repository
  root — examples belong in README or `docs/`.
- Architecture and design docs for cross-session context MUST be
  stored in `.dev/` (gitignored). The `docs/` directory is reserved
  for user-facing content only.
- Generated, temporary, or example files MUST be removed before
  committing (verify via `git status`).
- Rationale: A clean repository reduces cognitive load and avoids
  stale documentation drifting from the actual implementation.

### VI. Test Coverage Enforcement (NON-NEGOTIABLE)

- After each spec's development is complete, test coverage MUST be
  checked and MUST meet the CI thresholds defined in
  `.github/workflows/ci.yml`.
- Current thresholds (from CI):
  - Go core modules (80%): `internal/config`, `internal/proxy`,
    `internal/proxy/transform`, `internal/web`, `internal/bot`.
  - Go supporting modules (50%): `internal/daemon`,
    `internal/update`, `internal/sync`.
  - Web UI: branch coverage ≥ 70% (configured in `vitest.config.ts`).
- If a feature introduces new code that drops a package below its
  threshold, additional tests MUST be added in the same branch before
  the PR is opened.
- Coverage MUST be verified locally before pushing by running:
  `go test -cover ./internal/<pkg>` for affected Go packages, and
  `pnpm run test:coverage` in `web/` for frontend changes.
- Rationale: CI coverage gates exist to prevent regression. Fixing
  coverage after merge is more expensive and blocks other PRs. Each
  feature owner is responsible for maintaining the coverage bar.

## Technology & Architecture Constraints

- **Language**: Go. All production code MUST be in Go.
- **CLI framework**: Cobra. All commands MUST follow existing patterns
  in `cmd/`.
- **TUI**: Bubble Tea + Lip Gloss. TUI models MUST follow the
  `newXxxModel()` / `Init()` / `Update()` / `View()` convention.
- **Web frontend**: React + TypeScript with Vite build tooling
  (`web/src/`). Built output is embedded in Go binary via
  `internal/web/dist/`. Uses pnpm as package manager.
- **Config store**: JSON at `~/.zen/zen.json`. Schema changes follow
  Principle III.
- **Dev/Prod isolation**: Dev daemon uses ports 29840/29841 and
  config at `~/.zen-dev/zen.json`. Production ports (19840/19841)
  MUST NOT be touched during development.
- **Environment variable prefix**: `GOZEN_`.

## Development Workflow

- After modifying Go code, run `./scripts/dev.sh restart` to rebuild
  and restart the dev daemon.
- `go build ./...` and `go test ./...` MUST succeed before opening a
  pull request.
- Release checklist (see CLAUDE.md) MUST be completed before tagging:
  1. All tests pass.
  2. `Version` in `cmd/root.go` matches the tag.
  3. All four README translations are updated.
  4. Website documentation is current.
- PR reviews MUST verify compliance with this constitution's
  principles. Non-compliant changes require explicit justification
  in the PR description.

## Governance

- This constitution supersedes all other development practices for
  the GoZen project. Conflicts between this document and ad-hoc
  decisions MUST be resolved in favor of the constitution.
- Amendments require:
  1. A pull request modifying this file with a clear rationale.
  2. Version bump following semantic versioning:
     - MAJOR: principle removal or backward-incompatible redefinition.
     - MINOR: new principle or materially expanded guidance.
     - PATCH: wording clarifications, typo fixes.
  3. Update of the Sync Impact Report (HTML comment at file top).
- Compliance review: at the start of each feature branch, verify the
  plan's Constitution Check section against current principles.

**Version**: 1.2.0 | **Ratified**: 2026-02-27 | **Last Amended**: 2026-03-04
