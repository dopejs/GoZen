package main

import (
	"fmt"
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
	// TODO: Implement JSON marshaling
	return "{}"
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
