package main

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
)

var (
	docsPath  string
	i18nPath  string
	locale    string
	format    string
	output    string
	minCoverage float64
	verbose   bool
)

var auditCmd = &cobra.Command{
	Use:   "audit",
	Short: "Audit translation coverage for documentation",
	Long: `Scan documentation and generate translation coverage report.

This command scans the source documentation directory and compares it with
translations in the i18n directory to identify missing translations and
calculate coverage percentages for each locale.`,
	RunE: runAudit,
}

func init() {
	rootCmd.AddCommand(auditCmd)

	// Flags
	auditCmd.Flags().StringVar(&docsPath, "docs-path", "./docs", "Path to source docs directory")
	auditCmd.Flags().StringVar(&i18nPath, "i18n-path", "./i18n", "Path to i18n directory")
	auditCmd.Flags().StringVar(&locale, "locale", "", "Audit specific locale only (default: all)")
	auditCmd.Flags().StringVar(&format, "format", "table", "Output format: table|json|markdown")
	auditCmd.Flags().StringVar(&output, "output", "", "Write output to file instead of stdout")
	auditCmd.Flags().Float64Var(&minCoverage, "min-coverage", 0, "Exit with error if coverage below threshold")
	auditCmd.Flags().BoolVar(&verbose, "verbose", false, "Show detailed file-by-file analysis")
}

func runAudit(cmd *cobra.Command, args []string) error {
	// Scan source documentation
	sourceDocs, err := ScanSourceDocs(docsPath)
	if err != nil {
		return fmt.Errorf("failed to scan source docs: %w", err)
	}

	if len(sourceDocs) == 0 {
		return fmt.Errorf("no documentation files found in %s", docsPath)
	}

	// Determine which locales to audit
	locales := []string{"zh-Hans", "zh-Hant", "es", "ja", "ko"}
	if locale != "" {
		locales = []string{locale}
	}

	// Build locale reports
	localeReports := make(map[string]*LocaleReport)

	for _, loc := range locales {
		// Scan translations for this locale
		translations, err := ScanTranslations(i18nPath, loc, sourceDocs)
		if err != nil {
			return fmt.Errorf("failed to scan translations for %s: %w", loc, err)
		}

		// Count translated files
		translatedCount := CountTranslatedFiles(translations)
		missingCount := len(sourceDocs) - translatedCount

		// Calculate coverage
		coverage := AnalyzeCoverage(len(sourceDocs), translatedCount)

		// Build locale info
		localeInfo := Locale{
			Code:            loc,
			Name:            getLocaleName(loc),
			I18nPath:        i18nPath,
			TotalFiles:      len(sourceDocs),
			TranslatedFiles: translatedCount,
			MissingFiles:    missingCount,
			OutdatedFiles:   0, // TODO: Implement staleness detection
			CoveragePercent: coverage,
		}

		// Build locale report
		report := BuildLocaleReport(localeInfo, translations)
		localeReports[loc] = report
	}

	// Build audit report
	auditReport := BuildAuditReport(sourceDocs, localeReports)
	auditReport.Timestamp = time.Now()
	auditReport.SourceDocsPath = docsPath

	// Generate output based on format
	var reportOutput string
	switch format {
	case "json":
		reportOutput = GenerateJSONReport(auditReport)
	case "markdown":
		reportOutput = GenerateMarkdownReport(auditReport)
	default:
		reportOutput = GenerateTableReport(auditReport)
	}

	// Write output
	if output != "" {
		if err := os.WriteFile(output, []byte(reportOutput), 0644); err != nil {
			return fmt.Errorf("failed to write output file: %w", err)
		}
		fmt.Printf("Report written to %s\n", output)
	} else {
		fmt.Print(reportOutput)
	}

	// Check minimum coverage threshold
	if minCoverage > 0 && auditReport.OverallCoverage < minCoverage {
		return fmt.Errorf("coverage %.1f%% is below minimum threshold %.1f%%",
			auditReport.OverallCoverage, minCoverage)
	}

	return nil
}

// getLocaleName returns the display name for a locale code
func getLocaleName(code string) string {
	names := map[string]string{
		"zh-Hans": "简体中文",
		"zh-Hant": "繁體中文",
		"es":      "Español",
		"ja":      "日本語",
		"ko":      "한국어",
	}
	if name, ok := names[code]; ok {
		return name
	}
	return code
}
