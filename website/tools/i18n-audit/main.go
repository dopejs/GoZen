package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	version = "1.0.0"
)

var rootCmd = &cobra.Command{
	Use:   "i18n-audit",
	Short: "Audit translation coverage for documentation",
	Long: `i18n-audit is a tool to scan documentation directories and generate
reports on translation coverage across multiple locales.

It helps identify missing translations, detect outdated content, and
maintain translation consistency for internationalized documentation.`,
	Version: version,
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
