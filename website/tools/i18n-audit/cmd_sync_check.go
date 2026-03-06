package main

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
)

var (
	syncLocale string
)

var syncCheckCmd = &cobra.Command{
	Use:   "sync-check",
	Short: "Detect outdated translations by comparing modification times",
	Long: `Check if translations are outdated by comparing source file modification
times with translation file modification times.

This command uses git log to determine when files were last modified,
making it more reliable than file system timestamps.`,
	RunE: runSyncCheck,
}

func init() {
	rootCmd.AddCommand(syncCheckCmd)

	// Flags
	syncCheckCmd.Flags().StringVar(&docsPath, "docs-path", "./docs", "Path to source docs directory")
	syncCheckCmd.Flags().StringVar(&i18nPath, "i18n-path", "./i18n", "Path to i18n directory")
	syncCheckCmd.Flags().StringVar(&syncLocale, "locale", "", "Check specific locale only (default: all)")
}

func runSyncCheck(cmd *cobra.Command, args []string) error {
	// Scan source documentation
	sourceDocs, err := ScanSourceDocs(docsPath)
	if err != nil {
		return fmt.Errorf("failed to scan source docs: %w", err)
	}

	if len(sourceDocs) == 0 {
		return fmt.Errorf("no documentation files found in %s", docsPath)
	}

	// Get git modification times for source files
	sourceModTimes := make(map[string]time.Time)
	for _, doc := range sourceDocs {
		modTime, err := GetGitModTime(doc.FullPath)
		if err != nil {
			// Fallback to file system mtime if git fails
			modTime = doc.ModifiedTime
		}
		sourceModTimes[doc.Path] = modTime
	}

	// Determine which locales to check
	locales := []string{"zh-Hans", "zh-Hant", "es", "ja", "ko"}
	if syncLocale != "" {
		locales = []string{syncLocale}
	}

	// Check each locale for outdated translations
	outdatedByLocale := make(map[string][]OutdatedFile)

	for _, loc := range locales {
		// Scan translations for this locale
		translations, err := ScanTranslations(i18nPath, loc, sourceDocs)
		if err != nil {
			return fmt.Errorf("failed to scan translations for %s: %w", loc, err)
		}

		// Check each translation
		for _, trans := range translations {
			if trans.Status != StatusExists {
				continue
			}

			// Get translation modification time
			transModTime, err := GetGitModTime(trans.TranslatedPath)
			if err != nil {
				// Fallback to file system mtime
				transModTime = trans.ModifiedTime
			}

			// Compare with source modification time
			sourceModTime := sourceModTimes[trans.SourcePath]
			if sourceModTime.After(transModTime) {
				outdatedByLocale[loc] = append(outdatedByLocale[loc], OutdatedFile{
					Path:                trans.SourcePath,
					SourceModified:      sourceModTime,
					TranslationModified: transModTime,
				})
			}
		}
	}

	// Generate report
	report := GenerateSyncReport(outdatedByLocale)
	fmt.Print(report)

	// Exit with error if outdated translations found
	totalOutdated := 0
	for _, files := range outdatedByLocale {
		totalOutdated += len(files)
	}

	if totalOutdated > 0 {
		os.Exit(1)
	}

	return nil
}
