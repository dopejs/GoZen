# Data Model: Website Documentation Internationalization

**Feature**: 012-website-i18n-docs
**Date**: 2026-03-06

## Overview

This document defines the data structures for the i18n documentation audit tool. The tool scans the Docusaurus website structure and generates reports on translation coverage.

## Core Entities

### DocumentationPage

Represents a single documentation file in the source language.

**Attributes**:
- `Path` (string): Relative path from docs root (e.g., "getting-started.md")
- `FullPath` (string): Absolute file system path
- `Version` (string): Documentation version ("current", "version-3.0", etc.)
- `Size` (int64): File size in bytes
- `ModifiedTime` (time.Time): Last modification timestamp (from git)
- `Frontmatter` (map[string]interface{}): Parsed YAML frontmatter

**Relationships**:
- Has many `Translation` (one per locale)

**Validation Rules**:
- Path must end with `.md` or `.mdx`
- Path must be relative to docs root
- ModifiedTime used for staleness detection

**Example**:
```go
type DocumentationPage struct {
    Path         string
    FullPath     string
    Version      string
    Size         int64
    ModifiedTime time.Time
    Frontmatter  map[string]interface{}
}
```

---

### Translation

Represents a translated version of a documentation page.

**Attributes**:
- `SourcePath` (string): Path of the source document
- `Locale` (string): Target language code (e.g., "zh-Hans", "es")
- `TranslatedPath` (string): Path to translated file (if exists)
- `Status` (TranslationStatus): Current state of translation
- `ModifiedTime` (time.Time): Last modification timestamp (if exists)
- `IsOutdated` (bool): True if source modified after translation

**Relationships**:
- Belongs to one `DocumentationPage`
- Belongs to one `Locale`

**State Transitions**:
```
[missing] → [exists] (when translation file created)
[exists] → [outdated] (when source file modified)
[outdated] → [exists] (when translation updated)
[exists] → [not-required] (when marked as excluded)
```

**Example**:
```go
type Translation struct {
    SourcePath     string
    Locale         string
    TranslatedPath string
    Status         TranslationStatus
    ModifiedTime   time.Time
    IsOutdated     bool
}

type TranslationStatus int

const (
    StatusMissing TranslationStatus = iota
    StatusExists
    StatusOutdated
    StatusNotRequired
)
```

---

### Locale

Represents a supported language in the documentation.

**Attributes**:
- `Code` (string): Language code (e.g., "zh-Hans", "es", "ja")
- `Name` (string): Display name (e.g., "简体中文", "Español")
- `I18nPath` (string): Path to i18n directory for this locale
- `TotalFiles` (int): Total number of source files
- `TranslatedFiles` (int): Number of translated files
- `MissingFiles` (int): Number of missing translations
- `OutdatedFiles` (int): Number of outdated translations
- `CoveragePercent` (float64): Translation coverage percentage

**Relationships**:
- Has many `Translation`

**Calculated Fields**:
- `CoveragePercent = (TranslatedFiles / TotalFiles) * 100`
- `MissingFiles = TotalFiles - TranslatedFiles - NotRequiredFiles`

**Example**:
```go
type Locale struct {
    Code             string
    Name             string
    I18nPath         string
    TotalFiles       int
    TranslatedFiles  int
    MissingFiles     int
    OutdatedFiles    int
    CoveragePercent  float64
}
```

---

### AuditReport

Represents the complete audit results for all locales.

**Attributes**:
- `Timestamp` (time.Time): When the audit was run
- `SourceDocsPath` (string): Path to source documentation
- `SourceFiles` ([]DocumentationPage): List of all source files
- `Locales` (map[string]*LocaleReport): Report per locale
- `TotalSourceFiles` (int): Total number of source files
- `OverallCoverage` (float64): Average coverage across all locales

**Relationships**:
- Has many `LocaleReport` (one per locale)

**Calculated Fields**:
- `OverallCoverage = average(Locales[*].CoveragePercent)`

**Example**:
```go
type AuditReport struct {
    Timestamp        time.Time
    SourceDocsPath   string
    SourceFiles      []DocumentationPage
    Locales          map[string]*LocaleReport
    TotalSourceFiles int
    OverallCoverage  float64
}
```

---

### LocaleReport

Detailed report for a single locale.

**Attributes**:
- `Locale` (Locale): Locale information
- `Translations` ([]Translation): All translations for this locale
- `MissingFiles` ([]string): List of missing file paths
- `OutdatedFiles` ([]string): List of outdated file paths
- `Priority` ([]string): Prioritized list of files to translate (by size/importance)

**Relationships**:
- Belongs to one `AuditReport`
- Has one `Locale`
- Has many `Translation`

**Example**:
```go
type LocaleReport struct {
    Locale        Locale
    Translations  []Translation
    MissingFiles  []string
    OutdatedFiles []string
    Priority      []string
}
```

---

### ExclusionRule

Represents a rule to exclude certain files from translation requirements.

**Attributes**:
- `Pattern` (string): Glob pattern (e.g., "internal/*.md", "draft-*.md")
- `Reason` (string): Why this file doesn't need translation
- `AppliesTo` ([]string): List of locale codes (empty = all locales)

**Example**:
```go
type ExclusionRule struct {
    Pattern   string
    Reason    string
    AppliesTo []string
}
```

**Usage**:
- Stored in `.i18n-exclusions.yaml` in docs root
- Checked during audit to mark files as `StatusNotRequired`

---

## Data Flow

```
1. Scanner reads source docs directory
   → Produces []DocumentationPage

2. For each Locale:
   a. Scanner reads i18n/{locale}/docusaurus-plugin-content-docs/
   b. Analyzer compares source vs translated files
   c. Analyzer checks git timestamps for staleness
   → Produces LocaleReport

3. Reporter aggregates all LocaleReports
   → Produces AuditReport

4. Reporter formats output (console/JSON/markdown)
   → Displays to user
```

## File System Mapping

### Source Files
```
docs/getting-started.md
→ DocumentationPage{
    Path: "getting-started.md",
    Version: "current",
    ...
}
```

### Translated Files
```
i18n/zh-Hans/docusaurus-plugin-content-docs/current/getting-started.md
→ Translation{
    SourcePath: "getting-started.md",
    Locale: "zh-Hans",
    Status: StatusExists,
    ...
}
```

### Missing Translation
```
docs/agent-infrastructure.md exists
i18n/zh-Hans/.../agent-infrastructure.md does NOT exist
→ Translation{
    SourcePath: "agent-infrastructure.md",
    Locale: "zh-Hans",
    Status: StatusMissing,
    ...
}
```

## Validation Rules

### DocumentationPage
- Path must be relative and within docs directory
- File must exist and be readable
- Frontmatter must be valid YAML (if present)

### Translation
- SourcePath must reference a valid DocumentationPage
- Locale must be in supported locales list
- If Status = StatusExists, TranslatedPath must point to existing file
- IsOutdated = true only if both source and translation exist

### Locale
- Code must match Docusaurus locale configuration
- I18nPath must exist in file system
- CoveragePercent must be between 0 and 100

### AuditReport
- Timestamp must be set
- SourceFiles must not be empty
- Locales must contain at least one locale
- OverallCoverage must be between 0 and 100

## Performance Considerations

**File Scanning**:
- Use `filepath.Walk` with concurrent processing
- Cache file stats to avoid repeated syscalls
- Expected: <1 second for ~100 files

**Git Operations**:
- Use `git log -1 --format=%ct <file>` for modification time
- Batch git operations where possible
- Expected: <1 second for ~100 files

**Memory Usage**:
- All data structures fit in memory (~1MB for 100 files)
- No need for database or persistent storage
- Report generated on-demand

## Future Enhancements

1. **Translation Memory**: Track common phrases across files
2. **Importance Scoring**: Weight files by page views or size
3. **Auto-prioritization**: ML-based ranking of translation urgency
4. **Diff Detection**: Show what changed in source files (not just timestamp)
5. **Translation Suggestions**: LLM-assisted translation drafts
