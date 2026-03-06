package main

import (
	"path/filepath"
	"sort"
	"strings"
)

// AnalyzeCoverage calculates the translation coverage percentage
func AnalyzeCoverage(totalFiles, translatedFiles int) float64 {
	if totalFiles == 0 {
		return 0.0
	}
	return (float64(translatedFiles) / float64(totalFiles)) * 100.0
}

// DetectMissingFiles returns a list of file paths that are missing translations
func DetectMissingFiles(translations []Translation) []string {
	var missing []string

	for _, trans := range translations {
		if trans.Status == StatusMissing {
			missing = append(missing, trans.SourcePath)
		}
	}

	return missing
}

// ApplyExclusionRules checks if a file should be excluded from translation requirements
func ApplyExclusionRules(filePath, locale string, rules []ExclusionRule) bool {
	for _, rule := range rules {
		// Check if pattern matches
		matched, err := filepath.Match(rule.Pattern, filePath)
		if err != nil || !matched {
			continue
		}

		// If AppliesTo is empty, rule applies to all locales
		if len(rule.AppliesTo) == 0 {
			return true
		}

		// Check if locale is in the AppliesTo list
		for _, targetLocale := range rule.AppliesTo {
			if targetLocale == locale {
				return true
			}
		}
	}

	return false
}

// BuildLocaleReport creates a detailed report for a single locale
func BuildLocaleReport(locale Locale, translations []Translation) *LocaleReport {
	report := &LocaleReport{
		Locale:       locale,
		Translations: translations,
	}

	// Extract missing and outdated files
	for _, trans := range translations {
		if trans.Status == StatusMissing {
			report.MissingFiles = append(report.MissingFiles, trans.SourcePath)
		} else if trans.IsOutdated {
			report.OutdatedFiles = append(report.OutdatedFiles, trans.SourcePath)
		}
	}

	return report
}

// BuildAuditReport creates a complete audit report for all locales
func BuildAuditReport(sourceDocs []DocumentationPage, localeReports map[string]*LocaleReport) *AuditReport {
	report := &AuditReport{
		SourceFiles:      sourceDocs,
		Locales:          localeReports,
		TotalSourceFiles: len(sourceDocs),
	}

	// Calculate overall coverage
	if len(localeReports) > 0 {
		totalCoverage := 0.0
		for _, localeReport := range localeReports {
			totalCoverage += localeReport.Locale.CoveragePercent
		}
		report.OverallCoverage = totalCoverage / float64(len(localeReports))
	}

	return report
}

// CountTranslatedFiles counts how many files have translations
func CountTranslatedFiles(translations []Translation) int {
	count := 0
	for _, trans := range translations {
		if trans.Status == StatusExists {
			count++
		}
	}
	return count
}

// MatchPattern checks if a file path matches a glob pattern
func MatchPattern(pattern, path string) bool {
	// Handle wildcards in pattern
	if strings.Contains(pattern, "*") {
		matched, err := filepath.Match(pattern, path)
		if err != nil {
			return false
		}
		return matched
	}

	// Exact match
	return pattern == path
}

// PrioritizeMissingFiles sorts missing files by size (largest first)
func PrioritizeMissingFiles(sourceDocs []DocumentationPage, missing []string) []string {
	if len(missing) == 0 {
		return []string{}
	}

	// Create a map for quick lookup of file sizes
	sizeMap := make(map[string]int64)
	for _, doc := range sourceDocs {
		sizeMap[doc.Path] = doc.Size
	}

	// Create a slice of missing files with their sizes
	type fileWithSize struct {
		path string
		size int64
	}

	var files []fileWithSize
	for _, path := range missing {
		files = append(files, fileWithSize{
			path: path,
			size: sizeMap[path],
		})
	}

	// Sort by size descending (largest first)
	sort.Slice(files, func(i, j int) bool {
		return files[i].size > files[j].size
	})

	// Extract sorted paths
	var sorted []string
	for _, f := range files {
		sorted = append(sorted, f.path)
	}

	return sorted
}
