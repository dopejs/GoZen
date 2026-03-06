# Implementation Plan: Website Documentation Internationalization

**Branch**: `012-website-i18n-docs` | **Date**: 2026-03-06 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/012-website-i18n-docs/spec.md`

**Note**: This template is filled in by the `/speckit.plan` command. See `.specify/templates/plan-template.md` for the execution workflow.

## Summary

Create tooling to audit and manage translation coverage for GoZen website documentation. The primary requirement is to identify which documentation pages are missing translations across supported languages (English, Chinese Simplified, Chinese Traditional, Spanish), generate actionable reports, and detect when translations become outdated. This is a standalone script/tool that operates on the website repository's documentation structure.

## Technical Context

**Language/Version**: Go 1.21+
**Primary Dependencies**:
- `path/filepath` (built-in) - File system traversal
- `github.com/yuin/goldmark` - Markdown parsing
- `gopkg.in/yaml.v3` - YAML frontmatter parsing
- `github.com/spf13/cobra` - CLI framework (already in GoZen)
- `github.com/charmbracelet/lipgloss` - Pretty output (already in GoZen)

**Storage**: File system (reads website docs directory structure at `website/`)
**Testing**: Go built-in `testing` package + `github.com/stretchr/testify/assert`
**Target Platform**: Local development environment (macOS/Linux/Windows)
**Project Type**: CLI tool / audit script (standalone or integrated as `zen docs audit`)
**Performance Goals**: Generate complete audit report in <30 seconds (expected <2 seconds)
**Constraints**: Must work with Docusaurus i18n structure without modification
**Scale/Scope**: ~100 files total (16 source docs × 6 locales: en, zh-Hans, zh-Hant, es, ja, ko)

## Constitution Check (Post-Design)

*Re-evaluated after Phase 1 design completion*

### Principle I: Test-Driven Development
**Status**: ✅ PASS
- TDD approach confirmed in design
- Test structure defined: scanner_test.go, analyzer_test.go, reporter_test.go
- Table-driven tests will be used (Go idiom)

### Principle II: Simplicity & YAGNI
**Status**: ✅ PASS
- Minimal scope maintained: audit tool only
- No speculative features added during design
- Simple data structures (no database, no complex abstractions)

### Principle III: Config Migration Safety
**Status**: ✅ N/A
- Tool does not modify GoZen's config schema

### Principle IV: Branch Protection & Commit Discipline
**Status**: ✅ PASS
- Will follow standard PR workflow
- Atomic commits per task

### Principle V: Minimal Artifacts
**Status**: ✅ PASS
- Tool output is ephemeral (console/JSON reports)
- No persistent summary files created

### Principle VI: Test Coverage Enforcement
**Status**: ✅ PASS
- Tool will have test coverage for all core logic
- Expected coverage: >80% for scanner, analyzer, reporter modules

### Technology Constraints
**Status**: ✅ PASS
- Language: Go 1.21+ (matches GoZen core)
- CLI framework: Cobra (already used in GoZen)
- TUI: Lipgloss (already used in GoZen)
- All dependencies align with project standards

**Final Assessment**: No constitution violations. Design is complete and compliant.

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
# Standalone tool in website repository
website/
├── tools/
│   └── i18n-audit/
│       ├── main.go              # CLI entry point
│       ├── scanner.go           # File system scanner
│       ├── analyzer.go          # Translation status analyzer
│       ├── reporter.go          # Report generator
│       ├── scanner_test.go
│       ├── analyzer_test.go
│       ├── reporter_test.go
│       └── testdata/            # Test fixtures
│           ├── sample-docs/
│           └── sample-i18n/
│
├── docs/                        # Source documentation (English)
├── versioned_docs/
│   └── version-3.0/
└── i18n/                        # Translations
    ├── zh-Hans/
    ├── zh-Hant/
    ├── es/
    ├── ja/
    └── ko/

# Alternative: Integrated into GoZen CLI (future enhancement)
cmd/
└── docs.go                      # zen docs audit subcommand

internal/
└── docsaudit/
    ├── scanner.go
    ├── analyzer.go
    └── reporter.go
```

**Structure Decision**:
- **Phase 1 (MVP)**: Standalone tool in `website/tools/i18n-audit/`
  - Simpler to develop and test independently
  - No impact on GoZen core binary size
  - Can be run directly from website repository

- **Phase 2 (Future)**: Integrate as `zen docs audit` subcommand
  - Better developer experience (single tool)
  - Can enforce coverage in release workflow
  - Requires moving code to GoZen core repository

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| [e.g., 4th project] | [current need] | [why 3 projects insufficient] |
| [e.g., Repository pattern] | [specific problem] | [why direct DB access insufficient] |
