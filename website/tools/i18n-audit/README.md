# i18n Documentation Audit Tool

A Go-based CLI tool to audit translation coverage for internationalized documentation. Designed for Docusaurus websites but adaptable to other i18n structures.

## Features

- **Coverage Reports**: Generate comprehensive translation coverage reports across all locales
- **Missing File Detection**: Identify which documentation pages need translation
- **Priority Sorting**: Sort missing files by size to prioritize translation work
- **Staleness Detection**: Detect outdated translations using git timestamps
- **Multiple Output Formats**: Table (default), JSON, and Markdown
- **CI/CD Integration**: Exit codes and JSON output for automated workflows

## Installation

### From Source

```bash
cd website/tools/i18n-audit
go build -o i18n-audit
```

### Run Directly

```bash
go run . [command] [flags]
```

## Quick Start

### Generate Coverage Report

```bash
./i18n-audit audit --docs-path ../../docs --i18n-path ../../i18n
```

Output:
```
Translation Coverage Report
Generated: 2026-03-06 10:30:00

┌──────────┬───────────┬─────────┬─────────┬──────────┐
│ Locale   │ Translated│ Missing │ Outdated│ Coverage │
├──────────┼───────────┼─────────┼─────────┼──────────┤
│ zh-Hans  │ 9         │ 7       │ 0       │   56.2% │
│ zh-Hant  │ 9         │ 7       │ 0       │   56.2% │
│ es       │ 9         │ 7       │ 0       │   56.2% │
│ ja       │ 0         │ 16      │ 0       │    0.0% │
│ ko       │ 0         │ 16      │ 0       │    0.0% │
└──────────┴───────────┴─────────┴─────────┴──────────┘

Overall Coverage: 33.8%
Total Source Files: 16
```

### List Missing Files with Priority

```bash
./i18n-audit list-missing --locale zh-Hans --priority
```

Output:
```
Missing Translations for zh-Hans
Total: 7 files

  middleware.md
  → Create: ../../i18n/zh-Hans/docusaurus-plugin-content-docs/current/middleware.md

  agent-infrastructure.md
  → Create: ../../i18n/zh-Hans/docusaurus-plugin-content-docs/current/agent-infrastructure.md

  ...
```

### Check for Outdated Translations

```bash
./i18n-audit sync-check --locale zh-Hans
```

Output:
```
Outdated Translations Report

No outdated translations found.
All translations are up-to-date! ✓
```

## Commands

### `audit`

Generate translation coverage report.

**Flags:**
- `--docs-path <path>` - Path to source docs directory (default: "./docs")
- `--i18n-path <path>` - Path to i18n directory (default: "./i18n")
- `--locale <code>` - Audit specific locale only (default: all)
- `--format <format>` - Output format: table|json|markdown (default: "table")
- `--output <file>` - Write output to file instead of stdout
- `--min-coverage <percent>` - Exit with error if coverage below threshold

**Examples:**
```bash
# Basic audit
./i18n-audit audit

# Specific locale
./i18n-audit audit --locale zh-Hans

# JSON output for CI/CD
./i18n-audit audit --format json --output coverage.json

# Enforce minimum coverage
./i18n-audit audit --min-coverage 80  # exits 1 if below 80%
```

### `list-missing`

List missing translation files with suggested paths.

**Flags:**
- `--docs-path <path>` - Path to source docs directory (default: "./docs")
- `--i18n-path <path>` - Path to i18n directory (default: "./i18n")
- `--locale <code>` - Locale to list missing files for (required)
- `--priority` - Sort by priority (file size)

**Examples:**
```bash
# List missing files
./i18n-audit list-missing --locale zh-Hans

# Prioritized list (largest files first)
./i18n-audit list-missing --locale zh-Hans --priority
```

### `sync-check`

Detect outdated translations by comparing modification times.

**Flags:**
- `--docs-path <path>` - Path to source docs directory (default: "./docs")
- `--i18n-path <path>` - Path to i18n directory (default: "./i18n")
- `--locale <code>` - Check specific locale only (default: all)

**Examples:**
```bash
# Check all locales
./i18n-audit sync-check

# Check specific locale
./i18n-audit sync-check --locale zh-Hans
```

**Exit Codes:**
- `0` - No outdated translations found
- `1` - Outdated translations detected

## CI/CD Integration

### GitHub Actions

```yaml
name: Check Translation Coverage

on:
  pull_request:
    paths:
      - 'website/docs/**'
      - 'website/i18n/**'

jobs:
  i18n-audit:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'

      - name: Run i18n audit
        run: |
          cd website/tools/i18n-audit
          go run . audit --format json --output coverage.json --min-coverage 80

      - name: Upload coverage report
        if: always()
        uses: actions/upload-artifact@v3
        with:
          name: i18n-coverage
          path: website/tools/i18n-audit/coverage.json
```

### Pre-commit Hook

```bash
#!/bin/bash
# .git/hooks/pre-commit

cd website/tools/i18n-audit
if ! go run . audit --min-coverage 80 --quiet; then
  echo "Translation coverage below 80%"
  echo "Run 'i18n-audit' to see details"
  exit 1
fi
```

## Output Formats

### Table (Default)

Human-readable table with colored output using lipgloss.

### JSON

Machine-readable format for CI/CD integration:

```json
{
  "timestamp": "2026-03-06T10:30:00Z",
  "source_docs_path": "./docs",
  "total_source_files": 16,
  "overall_coverage": 33.8,
  "locales": {
    "zh-Hans": {
      "code": "zh-Hans",
      "name": "简体中文",
      "translated_files": 9,
      "missing_files": 7,
      "outdated_files": 0,
      "coverage_percent": 56.2,
      "missing_file_list": [
        "agent-infrastructure.md",
        "bot.md",
        ...
      ]
    }
  }
}
```

### Markdown

Documentation-friendly format for tracking progress:

```markdown
# Translation Coverage Report

**Generated**: 2026-03-06 10:30:00
**Overall Coverage**: 33.8%

## Summary

| Locale | Translated | Missing | Outdated | Coverage |
|--------|-----------|---------|----------|----------|
| zh-Hans | 9 | 7 | 0 | 56.2% |
...
```

## Development

### Run Tests

```bash
go test ./...
```

### Run with Coverage

```bash
go test -cover ./...
```

### Build

```bash
go build -o i18n-audit
```

## Architecture

- **scanner.go**: File system scanning and frontmatter parsing
- **analyzer.go**: Coverage calculation and missing file detection
- **reporter.go**: Report generation (table, JSON, markdown)
- **git.go**: Git integration for timestamp comparison
- **cmd_*.go**: CLI command implementations

## Requirements

- Go 1.21+
- Git (for staleness detection)

## Dependencies

- `github.com/spf13/cobra` - CLI framework
- `github.com/charmbracelet/lipgloss` - Terminal styling
- `github.com/yuin/goldmark` - Markdown parsing
- `gopkg.in/yaml.v3` - YAML frontmatter parsing

## License

Part of the GoZen project.
