package cmd

import "testing"

func TestMatchVersionPrefix(t *testing.T) {
	tests := []struct {
		version string
		prefix  string
		want    bool
	}{
		{"1.2.3", "1", true},
		{"1.2.3", "1.2", true},
		{"1.2.3", "1.2.3", true},
		{"1.2.3", "2", false},
		{"1.2.3", "1.3", false},
		{"1.2.3", "1.2.4", false},
		{"2.0.0", "2", true},
		{"2.0.0", "2.0", true},
		{"2.0.0", "2.0.0", true},
		{"2.0.0", "2.1", false},
		{"1.10.0", "1.1", false},
		{"1.10.0", "1.10", true},
		// prefix longer than version
		{"1.2", "1.2.3", false},
	}

	for _, tt := range tests {
		got := matchVersionPrefix(tt.version, tt.prefix)
		if got != tt.want {
			t.Errorf("matchVersionPrefix(%q, %q) = %v, want %v", tt.version, tt.prefix, got, tt.want)
		}
	}
}

func TestCompareVersions(t *testing.T) {
	tests := []struct {
		a, b string
		want int
	}{
		{"1.0.0", "1.0.0", 0},
		{"1.0.0", "1.0.1", -1},
		{"1.0.1", "1.0.0", 1},
		{"1.1.0", "1.0.9", 1},
		{"2.0.0", "1.9.9", 1},
		{"1.10.0", "1.9.0", 1},
		{"1.0", "1.0.0", 0},
		{"1", "1.0.0", 0},
	}

	for _, tt := range tests {
		got := compareVersions(tt.a, tt.b)
		if got != tt.want {
			t.Errorf("compareVersions(%q, %q) = %d, want %d", tt.a, tt.b, got, tt.want)
		}
	}
}

func TestSortVersions(t *testing.T) {
	versions := []string{"2.0.0", "1.0.0", "1.10.0", "1.2.0", "1.1.0", "3.0.0"}
	sortVersions(versions)

	want := []string{"1.0.0", "1.1.0", "1.2.0", "1.10.0", "2.0.0", "3.0.0"}
	for i, v := range versions {
		if v != want[i] {
			t.Errorf("sortVersions[%d] = %q, want %q", i, v, want[i])
		}
	}
}

func TestResolveVersionFromList(t *testing.T) {
	// Test the matching + sorting logic together
	versions := []string{"1.0.0", "1.1.0", "1.1.1", "1.2.0", "2.0.0", "2.1.0"}

	tests := []struct {
		prefix string
		want   string
	}{
		{"", "2.1.0"},
		{"1", "1.2.0"},
		{"1.1", "1.1.1"},
		{"2", "2.1.0"},
		{"2.0", "2.0.0"},
		{"1.0.0", "1.0.0"},
	}

	for _, tt := range tests {
		var matched []string
		if tt.prefix == "" {
			matched = append(matched, versions...)
		} else {
			for _, v := range versions {
				if matchVersionPrefix(v, tt.prefix) {
					matched = append(matched, v)
				}
			}
		}
		sortVersions(matched)
		got := matched[len(matched)-1]
		if got != tt.want {
			t.Errorf("resolve(%q) = %q, want %q", tt.prefix, got, tt.want)
		}
	}
}
