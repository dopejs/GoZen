package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestScanSourceDocs(t *testing.T) {
	tests := []struct {
		name        string
		docsPath    string
		wantCount   int
		wantErr     bool
		description string
	}{
		{
			name:        "valid docs directory",
			docsPath:    "testdata/sample-docs",
			wantCount:   4,
			wantErr:     false,
			description: "should find all 4 markdown files",
		},
		{
			name:        "empty directory",
			docsPath:    "testdata/empty",
			wantCount:   0,
			wantErr:     false,
			description: "should return empty list for empty directory",
		},
		{
			name:        "non-existent directory",
			docsPath:    "testdata/nonexistent",
			wantCount:   0,
			wantErr:     true,
			description: "should return error for non-existent directory",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create empty directory for empty test case
			if tt.name == "empty directory" {
				os.MkdirAll(tt.docsPath, 0755)
				defer os.RemoveAll(tt.docsPath)
			}

			docs, err := ScanSourceDocs(tt.docsPath)

			if (err != nil) != tt.wantErr {
				t.Errorf("ScanSourceDocs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && len(docs) != tt.wantCount {
				t.Errorf("ScanSourceDocs() got %d docs, want %d", len(docs), tt.wantCount)
			}

			// Verify structure of returned docs
			if !tt.wantErr && len(docs) > 0 {
				for _, doc := range docs {
					if doc.Path == "" {
						t.Error("DocumentationPage.Path should not be empty")
					}
					if doc.FullPath == "" {
						t.Error("DocumentationPage.FullPath should not be empty")
					}
					if doc.Size <= 0 {
						t.Error("DocumentationPage.Size should be positive")
					}
					if !filepath.IsAbs(doc.FullPath) {
						t.Errorf("DocumentationPage.FullPath should be absolute, got %s", doc.FullPath)
					}
				}
			}
		})
	}
}

func TestScanTranslations(t *testing.T) {
	tests := []struct {
		name        string
		i18nPath    string
		locale      string
		sourceDocs  []DocumentationPage
		wantCount   int
		wantErr     bool
		description string
	}{
		{
			name:     "zh-Hans locale with partial translations",
			i18nPath: "testdata/sample-i18n",
			locale:   "zh-Hans",
			sourceDocs: []DocumentationPage{
				{Path: "getting-started.md"},
				{Path: "configuration.md"},
				{Path: "examples.md"},
				{Path: "api-reference.md"},
			},
			wantCount:   1,
			wantErr:     false,
			description: "should find 1 translated file for zh-Hans",
		},
		{
			name:     "es locale with partial translations",
			i18nPath: "testdata/sample-i18n",
			locale:   "es",
			sourceDocs: []DocumentationPage{
				{Path: "getting-started.md"},
				{Path: "configuration.md"},
			},
			wantCount:   1,
			wantErr:     false,
			description: "should find 1 translated file for es",
		},
		{
			name:     "non-existent locale",
			i18nPath: "testdata/sample-i18n",
			locale:   "ja",
			sourceDocs: []DocumentationPage{
				{Path: "getting-started.md"},
			},
			wantCount:   0,
			wantErr:     false,
			description: "should return empty list for non-existent locale",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			translations, err := ScanTranslations(tt.i18nPath, tt.locale, tt.sourceDocs)

			if (err != nil) != tt.wantErr {
				t.Errorf("ScanTranslations() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			translatedCount := 0
			for _, trans := range translations {
				if trans.Status == StatusExists {
					translatedCount++
				}
			}

			if translatedCount != tt.wantCount {
				t.Errorf("ScanTranslations() found %d translated files, want %d", translatedCount, tt.wantCount)
			}
		})
	}
}
