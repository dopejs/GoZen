package main

import (
	"testing"
)

func TestAnalyzeCoverage(t *testing.T) {
	tests := []struct {
		name             string
		totalFiles       int
		translatedFiles  int
		wantCoverage     float64
		description      string
	}{
		{
			name:            "50% coverage",
			totalFiles:      4,
			translatedFiles: 2,
			wantCoverage:    50.0,
			description:     "2 out of 4 files translated",
		},
		{
			name:            "100% coverage",
			totalFiles:      4,
			translatedFiles: 4,
			wantCoverage:    100.0,
			description:     "all files translated",
		},
		{
			name:            "0% coverage",
			totalFiles:      4,
			translatedFiles: 0,
			wantCoverage:    0.0,
			description:     "no files translated",
		},
		{
			name:            "25% coverage",
			totalFiles:      4,
			translatedFiles: 1,
			wantCoverage:    25.0,
			description:     "1 out of 4 files translated",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			coverage := AnalyzeCoverage(tt.totalFiles, tt.translatedFiles)

			if coverage != tt.wantCoverage {
				t.Errorf("AnalyzeCoverage() = %.1f%%, want %.1f%%", coverage, tt.wantCoverage)
			}
		})
	}
}

func TestDetectMissingFiles(t *testing.T) {
	tests := []struct {
		name         string
		sourceDocs   []DocumentationPage
		translations []Translation
		wantMissing  int
		description  string
	}{
		{
			name: "some files missing",
			sourceDocs: []DocumentationPage{
				{Path: "getting-started.md"},
				{Path: "configuration.md"},
				{Path: "examples.md"},
				{Path: "api-reference.md"},
			},
			translations: []Translation{
				{SourcePath: "getting-started.md", Status: StatusExists},
				{SourcePath: "configuration.md", Status: StatusMissing},
				{SourcePath: "examples.md", Status: StatusMissing},
				{SourcePath: "api-reference.md", Status: StatusMissing},
			},
			wantMissing: 3,
			description: "should detect 3 missing translations",
		},
		{
			name: "all files translated",
			sourceDocs: []DocumentationPage{
				{Path: "getting-started.md"},
				{Path: "configuration.md"},
			},
			translations: []Translation{
				{SourcePath: "getting-started.md", Status: StatusExists},
				{SourcePath: "configuration.md", Status: StatusExists},
			},
			wantMissing: 0,
			description: "should detect 0 missing translations",
		},
		{
			name: "all files missing",
			sourceDocs: []DocumentationPage{
				{Path: "getting-started.md"},
				{Path: "configuration.md"},
			},
			translations: []Translation{
				{SourcePath: "getting-started.md", Status: StatusMissing},
				{SourcePath: "configuration.md", Status: StatusMissing},
			},
			wantMissing: 2,
			description: "should detect 2 missing translations",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			missing := DetectMissingFiles(tt.translations)

			if len(missing) != tt.wantMissing {
				t.Errorf("DetectMissingFiles() found %d missing, want %d", len(missing), tt.wantMissing)
			}
		})
	}
}

func TestApplyExclusionRules(t *testing.T) {
	tests := []struct {
		name        string
		filePath    string
		locale      string
		rules       []ExclusionRule
		wantExclude bool
		description string
	}{
		{
			name:     "file matches exclusion pattern",
			filePath: "internal/debug.md",
			locale:   "zh-Hans",
			rules: []ExclusionRule{
				{Pattern: "internal/*.md", Reason: "Internal docs", AppliesTo: []string{}},
			},
			wantExclude: true,
			description: "should exclude files matching pattern",
		},
		{
			name:     "file does not match pattern",
			filePath: "getting-started.md",
			locale:   "zh-Hans",
			rules: []ExclusionRule{
				{Pattern: "internal/*.md", Reason: "Internal docs", AppliesTo: []string{}},
			},
			wantExclude: false,
			description: "should not exclude files not matching pattern",
		},
		{
			name:     "locale-specific exclusion",
			filePath: "api-reference.md",
			locale:   "ja",
			rules: []ExclusionRule{
				{Pattern: "api-reference.md", Reason: "Auto-generated", AppliesTo: []string{"ja", "ko"}},
			},
			wantExclude: true,
			description: "should exclude for specific locales",
		},
		{
			name:     "locale not in exclusion list",
			filePath: "api-reference.md",
			locale:   "zh-Hans",
			rules: []ExclusionRule{
				{Pattern: "api-reference.md", Reason: "Auto-generated", AppliesTo: []string{"ja", "ko"}},
			},
			wantExclude: false,
			description: "should not exclude for other locales",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			excluded := ApplyExclusionRules(tt.filePath, tt.locale, tt.rules)

			if excluded != tt.wantExclude {
				t.Errorf("ApplyExclusionRules() = %v, want %v", excluded, tt.wantExclude)
			}
		})
	}
}

func TestPrioritizeMissingFiles(t *testing.T) {
	tests := []struct {
		name        string
		sourceDocs  []DocumentationPage
		missing     []string
		wantOrder   []string
		description string
	}{
		{
			name: "sort by file size descending",
			sourceDocs: []DocumentationPage{
				{Path: "small.md", Size: 100},
				{Path: "large.md", Size: 10000},
				{Path: "medium.md", Size: 1000},
			},
			missing:     []string{"small.md", "large.md", "medium.md"},
			wantOrder:   []string{"large.md", "medium.md", "small.md"},
			description: "should prioritize larger files first",
		},
		{
			name: "empty missing list",
			sourceDocs: []DocumentationPage{
				{Path: "doc.md", Size: 100},
			},
			missing:     []string{},
			wantOrder:   []string{},
			description: "should return empty list",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prioritized := PrioritizeMissingFiles(tt.sourceDocs, tt.missing)

			if len(prioritized) != len(tt.wantOrder) {
				t.Errorf("PrioritizeMissingFiles() returned %d files, want %d", len(prioritized), len(tt.wantOrder))
				return
			}

			for i, path := range tt.wantOrder {
				if prioritized[i] != path {
					t.Errorf("PrioritizeMissingFiles() position %d = %s, want %s", i, prioritized[i], path)
				}
			}
		})
	}
}
