# Repository Guidelines

## Project Structure & Module Organization

- `cmd/`: CLI entrypoints such as `zen`, daemon commands, and helper binaries.
- `internal/`: core application packages. Key areas include `internal/proxy/` (daemon proxy, routing, transforms), `internal/config/`, `internal/daemon/`, `internal/web/`, and `internal/middleware/`.
- `tests/`: integration and e2e-style Go tests. Stable integration tests live in `tests/integration/`.
- `web/`: Vite + React frontend source. Built assets are emitted to `internal/web/dist/`; do not edit `dist/` by hand.
- `docs/` and `specs/`: design notes, contributor docs, and feature specs. Put architecture-heavy changes here when behavior changes.

## Build, Test, and Development Commands

- `make build`: installs web deps, builds frontend, then builds `bin/zen`.
- `go test -race -short ./...`: fast local validation across Go packages.
- `go test -race ./tests/integration/...`: stable integration coverage.
- `go test -v -timeout 600s ./tests/...`: broader e2e-style validation; slower and sometimes environment-sensitive.
- `cd web && pnpm build`: build frontend only.
- `cd web && pnpm test -- --run --coverage`: run frontend tests with coverage.

## Coding Style & Naming Conventions

- Go code must be `gofmt`-formatted; keep packages focused and fix root causes instead of layering workarounds.
- Follow Go naming conventions: exported identifiers in `CamelCase`, internal helpers in `camelCase`, tests as `TestXxx`.
- Frontend code is TypeScript; use ESLint (`cd web && pnpm lint`) and keep React components/types consistent with existing patterns.
- Avoid editing generated artifacts directly, especially `internal/web/dist/`.

## Testing Guidelines

- Prefer targeted tests first, then broader suites.
- Add unit/component tests next to code in `internal/<pkg>/*_test.go`.
- Add integration tests in `tests/integration/`; if a test is CI-sensitive, guard it with `SKIP_FLAKY_TESTS=true` as documented in `docs/TESTING.md`.
- Maintain strong coverage in critical packages, especially `internal/proxy`, `internal/proxy/transform`, `internal/config`, and `internal/web`.

## Commit & Pull Request Guidelines

- Match existing history: use concise conventional prefixes such as `feat:`, `fix:`, `docs:`, `refactor:`, or scoped forms like `feat(spec):`.
- Keep commits focused and explain behavior changes, not just file edits.
- PRs should include: purpose, risk level, affected packages, test evidence, and linked issue/spec when relevant.
- For proxy/routing/transform changes, call out protocol impacts and fallback behavior explicitly.
