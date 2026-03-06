package main

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

var (
	// Table styles
	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#5eead4"))

	cellStyle = lipgloss.NewStyle().
			Padding(0, 1)

	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#c4b5fd")).
			MarginBottom(1)
)

// GenerateTableReport generates a formatted table report using lipgloss
func GenerateTableReport(report *AuditReport) string {
	var output strings.Builder

	// Title
	output.WriteString(titleStyle.Render("Translation Coverage Report"))
	output.WriteString("\n")
	output.WriteString(fmt.Sprintf("Generated: %s\n\n", report.Timestamp.Format("2006-01-02 15:04:05")))

	// Table header
	output.WriteString("┌──────────┬───────────┬─────────┬─────────┬──────────┐\n")
	output.WriteString("│ Locale   │ Translated│ Missing │ Outdated│ Coverage │\n")
	output.WriteString("├──────────┼───────────┼─────────┼─────────┼──────────┤\n")

	// Sort locales for consistent output
	sortedLocales := SortLocalesByName(report.Locales)

	// Table rows
	for _, localeReport := range sortedLocales {
		locale := localeReport.Locale
		output.WriteString(fmt.Sprintf("│ %-8s │ %-9d │ %-7d │ %-7d │ %7s │\n",
			locale.Code,
			locale.TranslatedFiles,
			locale.MissingFiles,
			locale.OutdatedFiles,
			FormatCoveragePercent(locale.CoveragePercent),
		))
	}

	output.WriteString("└──────────┴───────────┴─────────┴─────────┴──────────┘\n\n")

	// Summary
	output.WriteString(fmt.Sprintf("Overall Coverage: %.1f%%\n", report.OverallCoverage))
	output.WriteString(fmt.Sprintf("Total Source Files: %d\n", report.TotalSourceFiles))

	return output.String()
}

// FormatCoveragePercent formats a coverage percentage for display
func FormatCoveragePercent(coverage float64) string {
	return fmt.Sprintf("%.1f%%", coverage)
}

// SortLocalesByName sorts locale reports by locale code alphabetically
func SortLocalesByName(locales map[string]*LocaleReport) []*LocaleReport {
	var sorted []*LocaleReport

	// Extract to slice
	for _, report := range locales {
		sorted = append(sorted, report)
	}

	// Sort by code
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Locale.Code < sorted[j].Locale.Code
	})

	return sorted
}

// GenerateJSONReport generates a JSON report (placeholder for future implementation)
func GenerateJSONReport(report *AuditReport) string {
	// Build JSON structure
	type JSONLocale struct {
		Code            string   `json:"code"`
		Name            string   `json:"name"`
		TranslatedFiles int      `json:"translated_files"`
		MissingFiles    int      `json:"missing_files"`
		OutdatedFiles   int      `json:"outdated_files"`
		CoveragePercent float64  `json:"coverage_percent"`
		MissingFileList []string `json:"missing_file_list"`
	}

	type JSONReport struct {
		Timestamp        string                 `json:"timestamp"`
		SourceDocsPath   string                 `json:"source_docs_path"`
		TotalSourceFiles int                    `json:"total_source_files"`
		OverallCoverage  float64                `json:"overall_coverage"`
		Locales          map[string]JSONLocale  `json:"locales"`
	}

	jsonReport := JSONReport{
		Timestamp:        report.Timestamp.Format(time.RFC3339),
		SourceDocsPath:   report.SourceDocsPath,
		TotalSourceFiles: report.TotalSourceFiles,
		OverallCoverage:  report.OverallCoverage,
		Locales:          make(map[string]JSONLocale),
	}

	for code, localeReport := range report.Locales {
		jsonReport.Locales[code] = JSONLocale{
			Code:            localeReport.Locale.Code,
			Name:            localeReport.Locale.Name,
			TranslatedFiles: localeReport.Locale.TranslatedFiles,
			MissingFiles:    localeReport.Locale.MissingFiles,
			OutdatedFiles:   localeReport.Locale.OutdatedFiles,
			CoveragePercent: localeReport.Locale.CoveragePercent,
			MissingFileList: localeReport.MissingFiles,
		}
	}

	// Marshal to JSON
	jsonBytes, err := json.MarshalIndent(jsonReport, "", "  ")
	if err != nil {
		return fmt.Sprintf(`{"error": "failed to marshal JSON: %s"}`, err.Error())
	}

	return string(jsonBytes)
}

// GenerateMarkdownReport generates a Markdown report (placeholder for future implementation)
func GenerateMarkdownReport(report *AuditReport) string {
	var output strings.Builder

	output.WriteString("# Translation Coverage Report\n\n")
	output.WriteString(fmt.Sprintf("**Generated**: %s\n", report.Timestamp.Format("2006-01-02 15:04:05")))
	output.WriteString(fmt.Sprintf("**Overall Coverage**: %.1f%%\n\n", report.OverallCoverage))

	output.WriteString("## Summary\n\n")
	output.WriteString("| Locale | Translated | Missing | Outdated | Coverage |\n")
	output.WriteString("|--------|-----------|---------|----------|----------|\n")

	sortedLocales := SortLocalesByName(report.Locales)
	for _, localeReport := range sortedLocales {
		locale := localeReport.Locale
		output.WriteString(fmt.Sprintf("| %s | %d | %d | %d | %.1f%% |\n",
			locale.Code,
			locale.TranslatedFiles,
			locale.MissingFiles,
			locale.OutdatedFiles,
			locale.CoveragePercent,
		))
	}

	return output.String()
}

// FormatTimestamp formats a timestamp for display
func FormatTimestamp(t time.Time) string {
	return t.Format("2006-01-02")
}

// GenerateMissingReport generates a report of missing translations with suggested paths
func GenerateMissingReport(locale string, missing []string, i18nPath string) string {
	var output strings.Builder

	output.WriteString(titleStyle.Render(fmt.Sprintf("Missing Translations for %s", locale)))
	output.WriteString("\n")

	if len(missing) == 0 {
		output.WriteString("\nNo missing translations found.\n")
		output.WriteString("All files are translated! ✓\n")
		return output.String()
	}

	output.WriteString(fmt.Sprintf("Total: %d files\n\n", len(missing)))

	for _, filePath := range missing {
		suggestedPath := filepath.Join(i18nPath, locale, "docusaurus-plugin-content-docs", "current", filePath)
		output.WriteString(fmt.Sprintf("  %s\n", filePath))
		output.WriteString(fmt.Sprintf("  → Create: %s\n\n", suggestedPath))
	}

	return output.String()
}

// CategorizePriority categorizes a file by size into priority tiers
func CategorizePriority(size int64) string {
	if size >= 5000 {
		return "Priority 1"
	} else if size >= 1000 {
		return "Priority 2"
	}
	return "Priority 3"
}

// GenerateSyncReport generates a report of outdated translations
func GenerateSyncReport(outdated map[string][]OutdatedFile) string {
	var output strings.Builder

	output.WriteString(titleStyle.Render("Outdated Translations Report"))
	output.WriteString("\n\n")

	// Check if there are any outdated files
	totalOutdated := 0
	for _, files := range outdated {
		totalOutdated += len(files)
	}

	if totalOutdated == 0 {
		output.WriteString("No outdated translations found.\n")
		output.WriteString("All translations are up-to-date! ✓\n")
		return output.String()
	}

	// Table header
	output.WriteString("┌──────────┬────────────────────────┬─────────────┬──────────────┐\n")
	output.WriteString("│ Locale   │ File                   │ Source Date │ Translation  │\n")
	output.WriteString("├──────────┼────────────────────────┼─────────────┼──────────────┤\n")

	// Table rows
	for locale, files := range outdated {
		for _, file := range files {
			output.WriteString(fmt.Sprintf("│ %-8s │ %-22s │ %-11s │ %-12s │\n",
				locale,
				truncateString(file.Path, 22),
				FormatTimestamp(file.SourceModified),
				FormatTimestamp(file.TranslationModified),
			))
		}
	}

	output.WriteString("└──────────┴────────────────────────┴─────────────┴──────────────┘\n\n")

	// Summary
	localeCount := len(outdated)
	output.WriteString(fmt.Sprintf("Total Outdated: %d files across %d locale(s)\n", totalOutdated, localeCount))

	return output.String()
}

// truncateString truncates a string to a maximum length
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
