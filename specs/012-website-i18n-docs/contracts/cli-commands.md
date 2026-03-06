# CLI Command Contract: i18n-audit

**Feature**: 012-website-i18n-docs
**Date**: 2026-03-06
**Type**: Command-line interface

## Overview

This document defines the command-line interface contract for the i18n documentation audit tool. The tool provides commands to audit translation coverage, detect outdated translations, and generate reports.

## Command Structure

```bash
i18n-audit [command] [flags]
```

## Commands

### 1. audit (default)

Scan documentation and generate translation coverage report.

**Usage**:
```bash
i18n-audit audit [flags]
i18n-audit [flags]  # audit is default command
```

**Flags**:
- `--docs-path <path>` - Path to source docs directory (default: "./docs")
- `--i18n-path <path>` - Path to i18n directory (default: "./i18n")
- `--format <format>` - Output format: table|json|markdown (default: "table")
- `--output <file>` - Write output to file instead of stdout
- `--locale <code>` - Audit specific locale only (default: all)
- `--min-coverage <percent>` - Exit with error if coverage below threshold (default: 0)
- `--verbose` - Show detailed file-by-file analysis

**Exit Codes**:
- `0` - Success (coverage meets threshold)
- `1` - Coverage below threshold
- `2` - Invalid arguments or file system error

**Output (table format)**:
```
Translation Coverage Report
Generated: 2026-03-06 10:30:00

┌──────────┬───────────┬─────────┬─────────┬──────────┐
│ Locale   │ Translated│ Missing │ Outdated│ Coverage │
├──────────┼───────────┼─────────┼─────────┼──────────┤
│ zh-Hans  │ 9         │ 7       │ 0       │ 56.3%    │
│ zh-Hant  │ 9         │ 7       │ 0       │ 56.3%    │
│ es       │ 9         │ 7       │ 0       │ 56.3%    │
│ ja       │ 0         │ 16      │ 0       │ 0.0%     │
│ ko       │ 0         │ 16      │ 0       │ 0.0%     │
└──────────┴───────────┴─────────┴─────────┴──────────┘

Overall Coverage: 33.8%
Total Source Files: 16
```

**Output (JSON format)**:
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
      "coverage_percent": 56.3,
      "missing_file_list": [
        "agent-infrastructure.md",
        "bot.md",
        "compression.md",
        "health-monitoring.md",
        "load-balancing.md",
        "middleware.md",
        "usage-tracking.md",
        "webhooks.md"
      ]
    }
  }
}
```

**Output (markdown format)**:
```markdown
# Translation Coverage Report

**Generated**: 2026-03-06 10:30:00
**Overall Coverage**: 33.8%

## Summary

| Locale | Translated | Missing | Outdated | Coverage |
|--------|-----------|---------|----------|----------|
| zh-Hans | 9 | 7 | 0 | 56.3% |
| zh-Hant | 9 | 7 | 0 | 56.3% |
| es | 9 | 7 | 0 | 56.3% |
| ja | 0 | 16 | 0 | 0.0% |
| ko | 0 | 16 | 0 | 0.0% |

## Missing Translations

### zh-Hans (7 files)
- agent-infrastructure.md
- bot.md
- compression.md
...
```

**Examples**:
```bash
# Basic audit
i18n-audit

# Audit specific locale
i18n-audit --locale zh-Hans

# JSON output for CI/CD
i18n-audit --format json --output coverage.json

# Enforce minimum coverage
i18n-audit --min-coverage 80  # exits 1 if below 80%

# Verbose output
i18n-audit --verbose
```

---

### 2. sync-check

Detect outdated translations by comparing modification times.

**Usage**:
```bash
i18n-audit sync-check [flags]
```

**Flags**:
- `--docs-path <path>` - Path to source docs directory (default: "./docs")
- `--i18n-path <path>` - Path to i18n directory (default: "./i18n")
- `--format <format>` - Output format: table|json|markdown (default: "table")
- `--locale <code>` - Check specific locale only (default: all)

**Exit Codes**:
- `0` - No outdated translations found
- `1` - Outdated translations detected
- `2` - Invalid arguments or file system error

**Output (table format)**:
```
Outdated Translations Report
Generated: 2026-03-06 10:30:00

┌──────────┬────────────────────────┬─────────────┬──────────────┐
│ Locale   │ File                   │ Source Date │ Translation  │
├──────────┼────────────────────────┼─────────────┼──────────────┤
│ zh-Hans  │ getting-started.md     │ 2026-03-01  │ 2026-02-15   │
│ zh-Hans  │ providers.md           │ 2026-03-05  │ 2026-02-20   │
└──────────┴────────────────────────┴─────────────┴──────────────┘

Total Outdated: 2 files across 1 locale
```

**Examples**:
```bash
# Check all locales
i18n-audit sync-check

# Check specific locale
i18n-audit sync-check --locale zh-Hans

# JSON output
i18n-audit sync-check --format json
```

---

### 3. list-missing

List missing translation files with suggested paths.

**Usage**:
```bash
i18n-audit list-missing [flags]
```

**Flags**:
- `--docs-path <path>` - Path to source docs directory (default: "./docs")
- `--i18n-path <path>` - Path to i18n directory (default: "./i18n")
- `--locale <code>` - List for specific locale only (required)
- `--priority` - Sort by priority (file size/importance)

**Exit Codes**:
- `0` - Success
- `2` - Invalid arguments

**Output**:
```
Missing Translations for zh-Hans
Total: 7 files

Priority 1 (Large files):
  agent-infrastructure.md
  → Create: i18n/zh-Hans/docusaurus-plugin-content-docs/current/agent-infrastructure.md

  webhooks.md
  → Create: i18n/zh-Hans/docusaurus-plugin-content-docs/current/webhooks.md

Priority 2 (Medium files):
  bot.md
  → Create: i18n/zh-Hans/docusaurus-plugin-content-docs/current/bot.md

  middleware.md
  → Create: i18n/zh-Hans/docusaurus-plugin-content-docs/current/middleware.md

Priority 3 (Small files):
  compression.md
  → Create: i18n/zh-Hans/docusaurus-plugin-content-docs/current/compression.md

  health-monitoring.md
  → Create: i18n/zh-Hans/docusaurus-plugin-content-docs/current/health-monitoring.md

  load-balancing.md
  → Create: i18n/zh-Hans/docusaurus-plugin-content-docs/current/load-balancing.md

  usage-tracking.md
  → Create: i18n/zh-Hans/docusaurus-plugin-content-docs/current/usage-tracking.md
```

**Examples**:
```bash
# List missing for Chinese
i18n-audit list-missing --locale zh-Hans

# Prioritized list
i18n-audit list-missing --locale zh-Hans --priority
```

---

### 4. version

Show tool version information.

**Usage**:
```bash
i18n-audit version
```

**Output**:
```
i18n-audit version 1.0.0
Built with Go 1.21.5
```

---

## Global Flags

Available for all commands:

- `--help, -h` - Show help for command
- `--version, -v` - Show version information
- `--quiet, -q` - Suppress non-error output
- `--color` - Force colored output (default: auto-detect)
- `--no-color` - Disable colored output

## Environment Variables

- `I18N_AUDIT_DOCS_PATH` - Default docs path
- `I18N_AUDIT_I18N_PATH` - Default i18n path
- `I18N_AUDIT_FORMAT` - Default output format
- `NO_COLOR` - Disable colored output (standard)

## Configuration File

Optional `.i18n-audit.yaml` in project root:

```yaml
docs_path: ./docs
i18n_path: ./i18n
default_format: table
min_coverage: 80

# Locales to audit (default: all from docusaurus.config)
locales:
  - zh-Hans
  - zh-Hant
  - es
  - ja
  - ko

# Files to exclude from audit
exclusions:
  - pattern: "internal/*.md"
    reason: "Internal documentation not for translation"
  - pattern: "draft-*.md"
    reason: "Draft files"
    applies_to: ["ja", "ko"]  # Only exclude for these locales
```

## Error Handling

### Common Errors

**Invalid path**:
```
Error: docs path does not exist: ./invalid-path
```

**No source files found**:
```
Error: no documentation files found in ./docs
Check that --docs-path points to the correct directory
```

**Invalid locale**:
```
Error: locale 'invalid' not found in supported locales
Supported: en, zh-Hans, zh-Hant, es, ja, ko
```

**Permission denied**:
```
Error: permission denied reading ./docs
Check file permissions
```

## Integration Examples

### CI/CD (GitHub Actions)

```yaml
- name: Check translation coverage
  run: |
    cd website
    go run tools/i18n-audit --format json --output coverage.json --min-coverage 80

- name: Upload coverage report
  uses: actions/upload-artifact@v3
  with:
    name: i18n-coverage
    path: website/coverage.json
```

### Pre-commit Hook

```bash
#!/bin/bash
# .git/hooks/pre-commit

cd website
if ! go run tools/i18n-audit --quiet --min-coverage 80; then
  echo "Translation coverage below 80%"
  echo "Run 'i18n-audit' to see details"
  exit 1
fi
```

### Makefile

```makefile
.PHONY: i18n-audit
i18n-audit:
	cd website && go run tools/i18n-audit

.PHONY: i18n-check
i18n-check:
	cd website && go run tools/i18n-audit --min-coverage 80
```

## Testing Contract

### Unit Tests

Test each command with:
- Valid inputs
- Invalid inputs (missing paths, invalid locales)
- Edge cases (empty directories, no translations)
- Output format validation (JSON schema, table formatting)

### Integration Tests

Test with sample documentation structure:
- Create test fixtures in `testdata/`
- Verify correct file discovery
- Verify accurate coverage calculation
- Verify all output formats

### Acceptance Tests

Based on user stories from spec.md:

**US1 - Identify Missing Translations**:
```bash
# Given: website with partial translations
# When: run audit
# Then: see list of missing files per locale
i18n-audit --format json | jq '.locales["zh-Hans"].missing_file_list'
```

**US2 - Translate Missing Documentation**:
```bash
# Given: missing translation identified
# When: create translation file
# Then: audit no longer shows it as missing
i18n-audit list-missing --locale zh-Hans
# Create file...
i18n-audit --locale zh-Hans  # Should show +1 translated
```

**US3 - Maintain Translation Consistency**:
```bash
# Given: source file modified
# When: run sync-check
# Then: translation flagged as outdated
i18n-audit sync-check --locale zh-Hans
```

## Performance Requirements

- Audit command: <2 seconds for 100 files
- Sync-check command: <3 seconds (includes git operations)
- List-missing command: <1 second
- Memory usage: <50MB

## Compatibility

- Go 1.21+
- Works on macOS, Linux, Windows
- No external dependencies beyond Go standard library + specified packages
- Compatible with Docusaurus 2.x and 3.x i18n structure
