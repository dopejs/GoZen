package main

import (
	"time"
)

// TranslationStatus represents the state of a translation
type TranslationStatus int

const (
	StatusMissing TranslationStatus = iota
	StatusExists
	StatusOutdated
	StatusNotRequired
)

// DocumentationPage represents a single documentation file in the source language
type DocumentationPage struct {
	Path         string                 // Relative path from docs root (e.g., "getting-started.md")
	FullPath     string                 // Absolute file system path
	Version      string                 // Documentation version ("current", "version-3.0", etc.)
	Size         int64                  // File size in bytes
	ModifiedTime time.Time              // Last modification timestamp (from git)
	Frontmatter  map[string]interface{} // Parsed YAML frontmatter
}

// Translation represents a translated version of a documentation page
type Translation struct {
	SourcePath     string            // Path of the source document
	Locale         string            // Target language code (e.g., "zh-Hans", "es")
	TranslatedPath string            // Path to translated file (if exists)
	Status         TranslationStatus // Current state of translation
	ModifiedTime   time.Time         // Last modification timestamp (if exists)
	IsOutdated     bool              // True if source modified after translation
}

// Locale represents a supported language in the documentation
type Locale struct {
	Code            string  // Language code (e.g., "zh-Hans", "es", "ja")
	Name            string  // Display name (e.g., "简体中文", "Español")
	I18nPath        string  // Path to i18n directory for this locale
	TotalFiles      int     // Total number of source files
	TranslatedFiles int     // Number of translated files
	MissingFiles    int     // Number of missing translations
	OutdatedFiles   int     // Number of outdated translations
	CoveragePercent float64 // Translation coverage percentage
}

// LocaleReport contains detailed report for a single locale
type LocaleReport struct {
	Locale        Locale        // Locale information
	Translations  []Translation // All translations for this locale
	MissingFiles  []string      // List of missing file paths
	OutdatedFiles []string      // List of outdated file paths
	Priority      []string      // Prioritized list of files to translate (by size/importance)
}

// AuditReport represents the complete audit results for all locales
type AuditReport struct {
	Timestamp        time.Time                // When the audit was run
	SourceDocsPath   string                   // Path to source documentation
	SourceFiles      []DocumentationPage      // List of all source files
	Locales          map[string]*LocaleReport // Report per locale
	TotalSourceFiles int                      // Total number of source files
	OverallCoverage  float64                  // Average coverage across all locales
}

// ExclusionRule represents a rule to exclude certain files from translation requirements
type ExclusionRule struct {
	Pattern   string   // Glob pattern (e.g., "internal/*.md", "draft-*.md")
	Reason    string   // Why this file doesn't need translation
	AppliesTo []string // List of locale codes (empty = all locales)
}
