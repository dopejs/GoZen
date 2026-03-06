package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// ScanSourceDocs scans the documentation directory and returns all markdown files
func ScanSourceDocs(docsPath string) ([]DocumentationPage, error) {
	var docs []DocumentationPage

	// Check if directory exists
	if _, err := os.Stat(docsPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("docs path does not exist: %s", docsPath)
	}

	// Get absolute path
	absDocsPath, err := filepath.Abs(docsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Walk the directory tree
	err = filepath.Walk(absDocsPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Only process .md and .mdx files
		ext := filepath.Ext(path)
		if ext != ".md" && ext != ".mdx" {
			return nil
		}

		// Get relative path from docs root
		relPath, err := filepath.Rel(absDocsPath, path)
		if err != nil {
			return fmt.Errorf("failed to get relative path: %w", err)
		}

		// Parse frontmatter
		frontmatter, err := parseFrontmatter(path)
		if err != nil {
			// Non-fatal: continue without frontmatter
			frontmatter = make(map[string]interface{})
		}

		doc := DocumentationPage{
			Path:         relPath,
			FullPath:     path,
			Version:      "current",
			Size:         info.Size(),
			ModifiedTime: info.ModTime(),
			Frontmatter:  frontmatter,
		}

		docs = append(docs, doc)
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to scan docs: %w", err)
	}

	return docs, nil
}

// ScanTranslations scans the i18n directory for translations of source documents
func ScanTranslations(i18nPath, locale string, sourceDocs []DocumentationPage) ([]Translation, error) {
	var translations []Translation

	// Build locale path (Docusaurus structure)
	localePath := filepath.Join(i18nPath, locale, "docusaurus-plugin-content-docs", "current")

	// Check if locale directory exists
	localeExists := false
	if _, err := os.Stat(localePath); err == nil {
		localeExists = true
	}

	// For each source doc, check if translation exists
	for _, sourceDoc := range sourceDocs {
		translatedPath := filepath.Join(localePath, sourceDoc.Path)

		trans := Translation{
			SourcePath: sourceDoc.Path,
			Locale:     locale,
		}

		if localeExists {
			if info, err := os.Stat(translatedPath); err == nil {
				// Translation exists
				trans.TranslatedPath = translatedPath
				trans.Status = StatusExists
				trans.ModifiedTime = info.ModTime()
			} else {
				// Translation missing
				trans.Status = StatusMissing
			}
		} else {
			// Locale directory doesn't exist
			trans.Status = StatusMissing
		}

		translations = append(translations, trans)
	}

	return translations, nil
}

// parseFrontmatter extracts YAML frontmatter from a markdown file
func parseFrontmatter(filePath string) (map[string]interface{}, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	// Check for frontmatter delimiters (---)
	lines := strings.Split(string(content), "\n")
	if len(lines) < 3 || lines[0] != "---" {
		return nil, fmt.Errorf("no frontmatter found")
	}

	// Find closing delimiter
	endIdx := -1
	for i := 1; i < len(lines); i++ {
		if lines[i] == "---" {
			endIdx = i
			break
		}
	}

	if endIdx == -1 {
		return nil, fmt.Errorf("frontmatter not closed")
	}

	// Parse YAML
	frontmatterText := strings.Join(lines[1:endIdx], "\n")
	var frontmatter map[string]interface{}
	if err := yaml.Unmarshal([]byte(frontmatterText), &frontmatter); err != nil {
		return nil, fmt.Errorf("failed to parse frontmatter: %w", err)
	}

	return frontmatter, nil
}
