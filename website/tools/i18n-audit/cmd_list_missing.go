package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	listLocale   string
	listPriority bool
)

var listMissingCmd = &cobra.Command{
	Use:   "list-missing",
	Short: "List missing translation files with suggested paths",
	Long: `List missing translation files for a specific locale.

This command shows which files need to be translated and provides
suggested file paths where translations should be created.`,
	RunE: runListMissing,
}

func init() {
	rootCmd.AddCommand(listMissingCmd)

	// Flags
	listMissingCmd.Flags().StringVar(&docsPath, "docs-path", "./docs", "Path to source docs directory")
	listMissingCmd.Flags().StringVar(&i18nPath, "i18n-path", "./i18n", "Path to i18n directory")
	listMissingCmd.Flags().StringVar(&listLocale, "locale", "", "Locale to list missing files for (required)")
	listMissingCmd.Flags().BoolVar(&listPriority, "priority", false, "Sort by priority (file size)")
	listMissingCmd.MarkFlagRequired("locale")
}

func runListMissing(cmd *cobra.Command, args []string) error {
	// Scan source documentation
	sourceDocs, err := ScanSourceDocs(docsPath)
	if err != nil {
		return fmt.Errorf("failed to scan source docs: %w", err)
	}

	if len(sourceDocs) == 0 {
		return fmt.Errorf("no documentation files found in %s", docsPath)
	}

	// Scan translations for the specified locale
	translations, err := ScanTranslations(i18nPath, listLocale, sourceDocs)
	if err != nil {
		return fmt.Errorf("failed to scan translations for %s: %w", listLocale, err)
	}

	// Detect missing files
	missing := DetectMissingFiles(translations)

	// Prioritize if requested
	if listPriority {
		missing = PrioritizeMissingFiles(sourceDocs, missing)
	}

	// Generate report
	report := GenerateMissingReport(listLocale, missing, i18nPath)
	fmt.Print(report)

	return nil
}
