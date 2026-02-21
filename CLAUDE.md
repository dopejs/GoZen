# GoZen - Claude Code Environment Switcher

## Project Overview

Go CLI tool for managing multiple Claude API provider configurations with proxy failover and TUI/Web interfaces.

## Tech Stack

- **Language**: Go
- **TUI**: Bubble Tea (charmbracelet/bubbletea) + Lip Gloss
- **Web**: Embedded static files (HTML/CSS/JS, no framework), Go HTTP server
- **CLI**: Cobra
- **Config**: JSON store at `~/.zen/zen.json`

## Project Structure

```
cmd/           # Cobra commands (root, config, pick, web, upgrade, bind)
internal/
  config/      # Config store, types, legacy migration, compat helpers
  daemon/      # Web server daemon management (start/stop, platform-specific)
  proxy/       # Reverse proxy with failover, token calculation, session cache
  update/      # Non-blocking version update checker with 24h cache
  web/         # HTTP API server + embedded static frontend
    static/    # app.js, index.html, style.css (vanilla JS, no build step)
tui/           # All TUI models (editor, pick, config_main, fallback, routing, etc.)
```

## Build & Test

```sh
go build ./...
go test ./...
```

## Release Process

Push a git tag to trigger GitHub Actions release workflow:

```sh
git tag v2.x.0
git push origin v2.x.0
```

Do NOT use `gh release create` — the CI pipeline handles release creation automatically.

## Workflow Rules

- **Commit often**: Each completed task/item should be committed individually, not batched into one large commit. After finishing a feature or fix, commit immediately before moving to the next task.
- **Branch protection**: The `main` branch is protected. All changes MUST go through a pull request — never push directly to main. Create a feature branch, commit there, then use `gh pr create` to open a PR.
- **Pre-release check**: Before tagging a release, check for unpushed commits (`git log origin/main..HEAD`) and push them first.
- **Update version constant**: Before releasing, update `Version` in `cmd/root.go` to match the release tag.
- **Update README**: Before releasing, check that `README.md` reflects all new features and changes. **All four language versions must be updated together**: `README.md` (English), `docs/README.zh-CN.md` (简体中文), `docs/README.zh-TW.md` (繁體中文), `docs/README.es.md` (Español).
- **No summary/explanation docs**: Do NOT create markdown files to summarize or explain completed work. No "implementation notes", no "changes summary", no "feature documentation". The commit message and code comments are sufficient.
- **Keep planning docs**: Architecture planning and design docs should be kept for context across sessions. Store them in `.dev/` folder (gitignored, for internal dev use only). `docs/` is for user-facing content only.
- **No example files**: Do NOT create example config files (*.json, *.yaml, etc.) in the repository root. Examples belong in README.md or `docs/`.
- **Minimal test files**: Only add tests for new public APIs or complex logic. Do not create excessive test files for simple functions. Prefer table-driven tests in existing *_test.go files.
- **No unnecessary files**: Before committing, review `git status` and remove any generated, temporary, or example files that should not be in the repository.
- **TDD for new features**: Use Test-Driven Development (TDD) for new feature development to ensure code quality. Write tests first, then implement the feature to make tests pass.

## Release Checklist

Before tagging a release, complete the following checklist:

1. **Bug check**: Review all code for unresolved bugs. Run `go test ./...` and ensure all tests pass.
2. **Version number**: Verify `Version` in `cmd/root.go` matches the release tag.
3. **Website documentation**:
   - Ensure website contains documentation for the release version
   - Verify documentation is accurate and complete
   - Confirm all new features are documented
   - Remove or update documentation for changed/removed features
4. **README files**: Update all README files to reflect the latest version:
   - `README.md` (English)
   - `docs/README.zh-CN.md` (简体中文)
   - `docs/README.zh-TW.md` (繁體中文)
   - `docs/README.es.md` (Español)

## Config Migration Rules

When modifying `OpenCCConfig` or its nested types in a way that changes the JSON schema (e.g. changing a field's type, renaming a field, restructuring data):

1. Bump `CurrentConfigVersion` in `internal/config/config.go`.
2. Add migration logic so older config files are parsed correctly. The current pattern uses a custom `UnmarshalJSON` on `OpenCCConfig` with a fallback path that parses changed fields as `json.RawMessage` and converts them to the new format.
3. The version check in `loadLocked()` (`store.go`) automatically upgrades `cfg.Version < CurrentConfigVersion` to the current version after unmarshal.
4. Add tests covering: old format parsing, mixed old/new format, field preservation on the fallback path, and marshal round-trip.

## Brand Colors

Primary palette for website, Web UI, and marketing materials:

| Name | Dark Mode | Light Mode | Usage |
|------|-----------|------------|-------|
| Teal (Primary) | `#5eead4` | `#0d9488` | Primary accent, buttons, links |
| Lavender (Secondary) | `#c4b5fd` | `#7c3aed` | Secondary accent, highlights |
| Sage | `#86efac` | - | Success states |
| Red | `#fb7185` | - | Error states |
| Amber | `#fbbf24` | - | Warning states |
| Blue | `#93c5fd` | - | Info states |

Background (Dark): `#0f1117` → `#181a24` → `#1e2030` → `#252837`
Background (Light): `#f8fafc` → `#ffffff` → `#f1f5f9` → `#e2e8f0`

## Key Conventions

- Config convenience functions in `internal/config/compat.go` wrap `DefaultStore()` methods
- TUI models follow Bubble Tea pattern: `newXxxModel()`, `Init()`, `Update()`, `View()`
- Standalone TUI entry points: `RunXxx()` functions in tui package
- Inline config sub-editors use wrapper types implementing `tea.Model` (e.g. `editorWrapper`, `fallbackWrapper`)
- Web API routes: `/api/v1/providers`, `/api/v1/profiles`, `/api/v1/health`, `/api/v1/reload`
- Web frontend uses vanilla JS (no build tools), CSS custom properties for theming
- Model IDs in autocomplete must be verified against official Anthropic docs
- Environment variable prefix: `GOZEN_` (e.g., `GOZEN_WEB_DAEMON`)

## Version History

- v1.0.0: Initial release
- v1.1.0: Go rewrite with proxy failover and TUI
- v1.1.1: Upgrade command, sorted configs
- v1.2.0: Fallback profiles, pick command, installer
- v1.3.0: Web UI, profile assignment on provider add, model autocomplete, CLI name args
- v1.3.1: Download progress bar, README refresh
- v1.3.2: Fix progress bar display (show downloaded/total size)
- v1.4.0: Scenario routing, token calculation, session cache, project bindings
- v1.5.0: Dashboard TUI, project binding CLI support, centralized CLI list
- v1.5.1: Fix symlink dedup in bindings, Web UI dropdown style, --port restriction, config v3→v5 migration, SQLite log storage
- v1.5.2: Allow reinstalling same version in upgrade command
- v1.5.3: Per-binary PID files to avoid multi-binary conflicts
- v2.0.0: Rename to GoZen (opencc → zen), config migration from ~/.opencc/ to ~/.zen/, non-blocking version update check
- v3.0.0: Usage tracking & budget control, provider health monitoring, smart load balancing, webhooks, context compression, middleware pipeline, agent infrastructure
