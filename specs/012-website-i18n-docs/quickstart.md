# Quickstart Guide: i18n Documentation Audit Tool

**Feature**: 012-website-i18n-docs
**Date**: 2026-03-06

## Overview

This guide demonstrates how to use the i18n documentation audit tool to identify missing translations, detect outdated content, and maintain translation coverage for the GoZen website.

## Prerequisites

- Go 1.21+ installed
- GoZen website repository cloned
- Basic familiarity with command line

## Installation

### Option 1: Run from source (Development)

```bash
cd website/tools/i18n-audit
go run . [command] [flags]
```

### Option 2: Build binary (Production)

```bash
cd website/tools/i18n-audit
go build -o i18n-audit
./i18n-audit [command] [flags]
```

### Option 3: Install globally

```bash
cd website/tools/i18n-audit
go install
i18n-audit [command] [flags]
```

## Quick Start Scenarios

### Scenario 1: First-time Audit

**Goal**: Get an overview of translation coverage across all languages.

**Steps**:

1. Navigate to website directory:
```bash
cd website
```

2. Run the audit:
```bash
go run tools/i18n-audit
```

**Expected Output**:
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

**Interpretation**:
- Chinese and Spanish have 56% coverage (9 of 16 files)
- Japanese and Korean have 0% coverage (no translations yet)
- 7 files are missing translations for Chinese/Spanish
- Overall coverage is 33.8% across all languages

---

### Scenario 2: Identify Missing Translations for a Specific Language

**Goal**: Get a prioritized list of files to translate for Chinese (Simplified).

**Steps**:

1. List missing files with priority:
```bash
cd website
go run tools/i18n-audit list-missing --locale zh-Hans --priority
```

**Expected Output**:
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
  health-monitoring.md
  load-balancing.md
  usage-tracking.md
```

**Next Steps**:
- Start with Priority 1 files (largest impact)
- Create translation files at the suggested paths
- Copy content from source and translate

---

### Scenario 3: Create a Translation

**Goal**: Translate a missing documentation file.

**Steps**:

1. Identify missing file:
```bash
cd website
go run tools/i18n-audit list-missing --locale zh-Hans
```

2. Create the translation directory (if needed):
```bash
mkdir -p i18n/zh-Hans/docusaurus-plugin-content-docs/current
```

3. Copy source file:
```bash
cp docs/agent-infrastructure.md \
   i18n/zh-Hans/docusaurus-plugin-content-docs/current/agent-infrastructure.md
```

4. Translate the content:
```bash
# Edit the file and translate:
# - Keep frontmatter structure
# - Translate title, description
# - Translate all text content
# - Keep code examples unchanged
vim i18n/zh-Hans/docusaurus-plugin-content-docs/current/agent-infrastructure.md
```

5. Verify the translation is detected:
```bash
go run tools/i18n-audit --locale zh-Hans
```

**Expected Output**:
```
Translation Coverage Report
Generated: 2026-03-06 11:00:00

┌──────────┬───────────┬─────────┬─────────┬──────────┐
│ Locale   │ Translated│ Missing │ Outdated│ Coverage │
├──────────┼───────────┼─────────┼─────────┼──────────┤
│ zh-Hans  │ 10        │ 6       │ 0       │ 62.5%    │
└──────────┴───────────┴─────────┴─────────┴──────────┘

Overall Coverage: 62.5%
Total Source Files: 16
```

**Success**: Coverage increased from 56.3% to 62.5%!

---

### Scenario 4: Detect Outdated Translations

**Goal**: Find translations that need updating after source files changed.

**Steps**:

1. Run sync check:
```bash
cd website
go run tools/i18n-audit sync-check
```

**Expected Output (if outdated translations exist)**:
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

**Expected Output (if all up-to-date)**:
```
Outdated Translations Report
Generated: 2026-03-06 10:30:00

No outdated translations found.
All translations are up-to-date!
```

2. Update outdated translations:
```bash
# Compare source and translation to see what changed
diff docs/getting-started.md \
     i18n/zh-Hans/docusaurus-plugin-content-docs/current/getting-started.md

# Update the translation
vim i18n/zh-Hans/docusaurus-plugin-content-docs/current/getting-started.md
```

3. Verify update:
```bash
go run tools/i18n-audit sync-check --locale zh-Hans
```

---

### Scenario 5: CI/CD Integration

**Goal**: Enforce minimum translation coverage in CI pipeline.

**Steps**:

1. Create GitHub Actions workflow (`.github/workflows/i18n-check.yml`):
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
          cd website
          go run tools/i18n-audit --format json --output coverage.json --min-coverage 80

      - name: Upload coverage report
        if: always()
        uses: actions/upload-artifact@v3
        with:
          name: i18n-coverage
          path: website/coverage.json

      - name: Comment on PR
        if: failure()
        uses: actions/github-script@v6
        with:
          script: |
            const fs = require('fs');
            const coverage = JSON.parse(fs.readFileSync('website/coverage.json', 'utf8'));
            const comment = `## Translation Coverage Report\n\n` +
              `Overall Coverage: ${coverage.overall_coverage.toFixed(1)}%\n` +
              `Minimum Required: 80%\n\n` +
              `Please add missing translations before merging.`;
            github.rest.issues.createComment({
              issue_number: context.issue.number,
              owner: context.repo.owner,
              repo: context.repo.repo,
              body: comment
            });
```

2. Test locally before pushing:
```bash
cd website
go run tools/i18n-audit --min-coverage 80
```

**Expected Behavior**:
- If coverage ≥ 80%: Exit code 0 (success)
- If coverage < 80%: Exit code 1 (failure), blocks PR merge

---

### Scenario 6: Generate Report for Documentation

**Goal**: Create a markdown report to track translation progress.

**Steps**:

1. Generate markdown report:
```bash
cd website
go run tools/i18n-audit --format markdown --output i18n-status.md
```

2. View the report:
```bash
cat i18n-status.md
```

**Expected Output**:
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
- health-monitoring.md
- load-balancing.md
- middleware.md
- usage-tracking.md
- webhooks.md

### zh-Hant (7 files)
...
```

3. Commit the report:
```bash
git add i18n-status.md
git commit -m "docs: update i18n coverage status"
```

---

### Scenario 7: Exclude Files from Translation

**Goal**: Mark certain files as "translation not required".

**Steps**:

1. Create exclusion config (`.i18n-audit.yaml`):
```yaml
docs_path: ./docs
i18n_path: ./i18n

exclusions:
  - pattern: "internal/*.md"
    reason: "Internal documentation not for public translation"

  - pattern: "draft-*.md"
    reason: "Draft files not ready for translation"

  - pattern: "api-reference.md"
    reason: "Auto-generated API docs"
    applies_to: ["ja", "ko"]  # Only exclude for Japanese/Korean
```

2. Run audit:
```bash
cd website
go run tools/i18n-audit
```

**Expected Behavior**:
- Excluded files won't appear in "Missing" count
- Coverage percentage calculated without excluded files
- Exclusions logged in verbose mode

---

## Common Workflows

### Weekly Translation Review

```bash
#!/bin/bash
# scripts/weekly-i18n-review.sh

cd website

echo "=== Translation Coverage Report ==="
go run tools/i18n-audit

echo ""
echo "=== Outdated Translations ==="
go run tools/i18n-audit sync-check

echo ""
echo "=== Priority Files to Translate ==="
for locale in zh-Hans zh-Hant es ja ko; do
  echo ""
  echo "--- $locale ---"
  go run tools/i18n-audit list-missing --locale $locale --priority | head -20
done
```

### Pre-release Checklist

```bash
#!/bin/bash
# scripts/pre-release-i18n-check.sh

cd website

# Require 100% coverage for primary languages
for locale in zh-Hans zh-Hant es; do
  echo "Checking $locale coverage..."
  if ! go run tools/i18n-audit --locale $locale --min-coverage 100 --quiet; then
    echo "ERROR: $locale coverage below 100%"
    exit 1
  fi
done

echo "All primary languages have 100% coverage!"
```

## Troubleshooting

### Issue: "No documentation files found"

**Cause**: Wrong directory or incorrect path

**Solution**:
```bash
# Check current directory
pwd

# Should be in website/ directory
cd website

# Or specify paths explicitly
go run tools/i18n-audit --docs-path ./docs --i18n-path ./i18n
```

### Issue: "Locale not found"

**Cause**: Locale code doesn't match Docusaurus config

**Solution**:
```bash
# Check supported locales in docusaurus.config.ts
cat docusaurus.config.ts | grep -A 10 "i18n:"

# Use correct locale code (e.g., zh-Hans not zh-CN)
go run tools/i18n-audit --locale zh-Hans
```

### Issue: Coverage seems incorrect

**Cause**: Cached or stale data

**Solution**:
```bash
# Run with verbose flag to see file-by-file analysis
go run tools/i18n-audit --verbose

# Check file structure matches expected pattern
ls -R i18n/zh-Hans/docusaurus-plugin-content-docs/
```

## Next Steps

- **For translators**: Use `list-missing --priority` to find high-impact files
- **For maintainers**: Set up CI/CD integration with `--min-coverage`
- **For reviewers**: Use `sync-check` to find outdated translations
- **For project managers**: Generate markdown reports for tracking progress

## Additional Resources

- [Docusaurus i18n Documentation](https://docusaurus.io/docs/i18n/introduction)
- [GoZen Translation Guidelines](../../../docs/TRANSLATION.md)
- [CLI Command Reference](./contracts/cli-commands.md)
- [Data Model Documentation](./data-model.md)
