# Research: Website Documentation Internationalization

**Date**: 2026-03-06
**Status**: Complete

## Executive Summary

GoZen website uses Docusaurus 3.9.2 with 6 supported locales (en, zh-Hans, zh-Hant, es, ja, ko). Current translation coverage is 56% for Chinese/Spanish (9 of 16 files) and 0% for Japanese/Korean. We will build a Go-based CLI tool to audit translation coverage and detect outdated translations.

## Technology Decisions

### Decision 1: Tool Language - Go

**Chosen**: Go

**Rationale**:
- Project consistency: GoZen is a Go project
- Single binary distribution (no runtime dependencies)
- Native file system operations (`path/filepath`)
- Can integrate as `zen docs audit` subcommand later
- Excellent performance for file scanning

**Alternatives Considered**:
- **Node.js/TypeScript**: Rich i18n ecosystem, better Docusaurus integration
  - Rejected: Adds `node_modules` complexity to Go-centric project
- **Python**: Simple scripting, good file handling
  - Rejected: Requires Python runtime, less consistent with project

### Decision 2: Website Structure - Docusaurus i18n

**Finding**: Website uses standard Docusaurus i18n structure

**Structure**:
```
website/
├── docs/                    # Source (English) - 16 files
├── versioned_docs/
│   └── version-3.0/        # 16 files
└── i18n/
    ├── zh-Hans/docusaurus-plugin-content-docs/current/  # 9 files (56%)
    ├── zh-Hant/docusaurus-plugin-content-docs/current/  # 9 files (56%)
    ├── es/docusaurus-plugin-content-docs/current/       # 9 files (56%)
    ├── ja/docusaurus-plugin-content-docs/version-3.0/   # 0 files (0%)
    └── ko/docusaurus-plugin-content-docs/version-3.0/   # 0 files (0%)
```

**Supported Locales**:
- `en` (default/source)
- `zh-Hans` (简体中文)
- `zh-Hant` (繁體中文)
- `es` (Español)
- `ja` (日本語)
- `ko` (한국어)

### Decision 3: Key Dependencies

**Chosen**:
- `path/filepath` - File system traversal and glob matching
- `github.com/yuin/goldmark` - Markdown parsing
- `gopkg.in/yaml.v3` - YAML frontmatter parsing
- `github.com/spf13/cobra` - CLI framework (already in GoZen)
- `github.com/charmbracelet/lipgloss` - Pretty output (already in GoZen)

**Rationale**: All dependencies either built-in or already used by GoZen

### Decision 4: Testing Framework - Go testing

**Chosen**: Built-in `testing` package + `testify/assert`

**Test Strategy**:
- Table-driven tests (Go idiom)
- Test fixtures in `testdata/` directory
- Unit tests for scanner, analyzer, reporter
- Integration tests with sample doc structures

## Current State Analysis

### Translation Coverage Gap

**Missing Translations** (7 new v3.0 files not translated):
1. `agent-infrastructure.md`
2. `bot.md`
3. `compression.md`
4. `health-monitoring.md`
5. `load-balancing.md`
6. `middleware.md`
7. `usage-tracking.md`
8. `webhooks.md`

**Coverage by Locale**:
- zh-Hans: 56% (9/16 files)
- zh-Hant: 56% (9/16 files)
- es: 56% (9/16 files)
- ja: 0% (0/16 files)
- ko: 0% (0/16 files)

### Additional Translation Files

**Root README translations** (in `/docs/`):
- `README.zh-CN.md`
- `README.zh-TW.md`
- `README.es.md`

**UI strings**: `i18n/{locale}/code.json`

## Implementation Approach

### Tool Architecture

```go
type AuditReport struct {
    SourceFiles      []string
    Locales          map[string]LocaleStatus
    MissingFiles     map[string][]string
    CoveragePercent  map[string]float64
}

type LocaleStatus struct {
    TranslatedFiles  []string
    MissingFiles     []string
    OutdatedFiles    []string  // Based on git mtime comparison
}
```

### Algorithm

1. Scan source docs directory (`docs/`, `versioned_docs/version-3.0/`)
2. Extract list of all `.md` files
3. For each configured locale:
   - Check `i18n/{locale}/docusaurus-plugin-content-docs/current/`
   - Compare file lists
   - Identify missing translations
   - Optional: Compare git timestamps for staleness
4. Generate report with coverage percentage and missing files

### Output Formats

1. **Console** (default): Colored table with lipgloss
2. **JSON**: For CI/CD integration
3. **Markdown**: For documentation/issues

## Best Practices

Based on Docusaurus documentation and i18n best practices:

1. **Version Control**: Keep translations in git (current approach ✓)
2. **Atomic Updates**: Update all languages when adding new docs
3. **Frontmatter Consistency**: Translate `title` and maintain `sidebar_position`
4. **CI/CD Integration**: Block releases if coverage drops below threshold
5. **Staleness Detection**: Flag translations older than source files

## Existing Tools Evaluated

**Reviewed**:
- `i18n-unused` (npm): Focuses on JSON translations, not markdown
- `i18next-scanner`: Extracts keys from code, not docs
- `appcheck-cli`: TypeScript, checks unused translations

**Verdict**: No existing tool fits GoZen's needs (Docusaurus markdown-based i18n + Go ecosystem). Custom tool is justified.

## Performance Considerations

**Expected Performance**:
- File count: ~100 files (16 source × 6 locales)
- Scan time: <1 second (Go's concurrent file operations)
- Report generation: <1 second
- **Total**: Well under 30-second requirement

## Integration Opportunities

**Future Enhancements**:
1. Integrate as `zen docs audit` subcommand
2. Pre-release hook to enforce 100% coverage
3. GitHub Actions workflow to comment on PRs with coverage changes
4. Auto-generate translation task issues

## References

- [Docusaurus i18n Introduction](https://docusaurus.io/docs/i18n/introduction)
- [Docusaurus i18n with Git](https://www.docusaurus.io/docs/next/i18n/git)
- [Understanding File Globbing in Go](https://leapcell.io/blog/understanding-file-globbing-in-go)
- GoZen website config: `website/docusaurus.config.ts`
- GoZen website structure: `website/` directory
