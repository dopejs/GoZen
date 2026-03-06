package main

import (
	"strings"
	"testing"
	"time"
)

func TestGenerateTableReport(t *testing.T) {
	tests := []struct {
		name        string
		report      *AuditReport
		wantContain []string
		description string
	}{
		{
			name: "basic report with multiple locales",
			report: &AuditReport{
				Timestamp:        time.Now(),
				TotalSourceFiles: 4,
				OverallCoverage:  37.5,
				Locales: map[string]*LocaleReport{
					"zh-Hans": {
						Locale: Locale{
							Code:            "zh-Hans",
							Name:            "简体中文",
							TotalFiles:      4,
							TranslatedFiles: 1,
							MissingFiles:    3,
							CoveragePercent: 25.0,
						},
					},
					"es": {
						Locale: Locale{
							Code:            "es",
							Name:            "Español",
							TotalFiles:      4,
							TranslatedFiles: 2,
							MissingFiles:    2,
							CoveragePercent: 50.0,
						},
					},
				},
			},
			wantContain: []string{
				"Translation Coverage Report",
				"zh-Hans",
				"es",
				"25.0%",
				"50.0%",
				"Overall Coverage: 37.5%",
			},
			description: "should contain locale names and coverage percentages",
		},
		{
			name: "100% coverage report",
			report: &AuditReport{
				Timestamp:        time.Now(),
				TotalSourceFiles: 2,
				OverallCoverage:  100.0,
				Locales: map[string]*LocaleReport{
					"zh-Hans": {
						Locale: Locale{
							Code:            "zh-Hans",
							TotalFiles:      2,
							TranslatedFiles: 2,
							MissingFiles:    0,
							CoveragePercent: 100.0,
						},
					},
				},
			},
			wantContain: []string{
				"100.0%",
				"Overall Coverage: 100.0%",
			},
			description: "should show 100% coverage",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := GenerateTableReport(tt.report)

			for _, want := range tt.wantContain {
				if !strings.Contains(output, want) {
					t.Errorf("GenerateTableReport() output missing %q", want)
				}
			}
		})
	}
}

func TestFormatCoveragePercent(t *testing.T) {
	tests := []struct {
		name     string
		coverage float64
		want     string
	}{
		{
			name:     "whole number",
			coverage: 50.0,
			want:     "50.0%",
		},
		{
			name:     "decimal",
			coverage: 56.3,
			want:     "56.3%",
		},
		{
			name:     "zero",
			coverage: 0.0,
			want:     "0.0%",
		},
		{
			name:     "hundred",
			coverage: 100.0,
			want:     "100.0%",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatCoveragePercent(tt.coverage)
			if got != tt.want {
				t.Errorf("FormatCoveragePercent() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSortLocalesByName(t *testing.T) {
	locales := map[string]*LocaleReport{
		"zh-Hans": {Locale: Locale{Code: "zh-Hans", Name: "简体中文"}},
		"es":      {Locale: Locale{Code: "es", Name: "Español"}},
		"ja":      {Locale: Locale{Code: "ja", Name: "日本語"}},
	}

	sorted := SortLocalesByName(locales)

	if len(sorted) != 3 {
		t.Errorf("SortLocalesByName() returned %d locales, want 3", len(sorted))
	}

	// Verify alphabetical order by code
	expectedOrder := []string{"es", "ja", "zh-Hans"}
	for i, code := range expectedOrder {
		if sorted[i].Locale.Code != code {
			t.Errorf("SortLocalesByName() position %d = %s, want %s", i, sorted[i].Locale.Code, code)
		}
	}
}
