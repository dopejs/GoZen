package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

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
		// prerelease: base prefix matches prerelease versions
		{"2.1.0-alpha.1", "2", true},
		{"2.1.0-alpha.1", "2.1", true},
		{"2.1.0-alpha.1", "2.1.0", true},
		{"2.1.0-alpha.1", "2.2", false},
		// prerelease prefix matches prerelease versions
		{"2.1.0-alpha.1", "2.1.0-alpha", true},
		{"2.1.0-alpha.2", "2.1.0-alpha", true},
		{"2.1.0-beta.1", "2.1.0-alpha", false},
		{"2.1.0-alpha.1", "2.1.0-alpha.1", true},
		{"2.1.0-alpha.1", "2.1.0-beta", false},
		// prerelease prefix does not match stable
		{"2.1.0", "2.1.0-alpha", false},
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
		// prerelease ordering
		{"2.1.0", "2.1.0-beta.1", 1},          // stable > prerelease
		{"2.1.0-alpha.1", "2.1.0", -1},         // prerelease < stable
		{"2.1.0-alpha.1", "2.1.0-alpha.2", -1}, // alpha.1 < alpha.2
		{"2.1.0-alpha.1", "2.1.0-beta.1", -1},  // alpha < beta
		{"2.1.0-beta.1", "2.1.0-alpha.1", 1},   // beta > alpha
		{"2.1.0-beta.2", "2.1.0-beta.1", 1},    // beta.2 > beta.1
		{"2.1.0-alpha.1", "2.1.0-alpha.1", 0},  // equal
		{"2.2.0-alpha.1", "2.1.0", 1},           // higher base wins
		{"2.0.0-rc.1", "2.0.0-beta.2", 1},      // rc > beta
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

func TestResolveVersionFromListPrerelease(t *testing.T) {
	// Simulate a release list with prereleases
	type release struct {
		version    string
		prerelease bool
	}
	releases := []release{
		{"1.0.0", false},
		{"2.0.0", false},
		{"2.1.0", false},
		{"2.1.1-alpha.1", true},
		{"2.1.1-alpha.2", true},
		{"2.1.1-beta.1", true},
		{"2.2.0-alpha.1", true},
	}

	tests := []struct {
		prefix string
		want   string
	}{
		// No prefix → latest stable only
		{"", "2.1.0"},
		// Stable prefix → excludes prereleases
		{"2", "2.1.0"},
		{"2.1", "2.1.0"},
		// Prerelease prefix → includes prereleases
		{"2.1.1-alpha", "2.1.1-alpha.2"},
		{"2.1.1-beta", "2.1.1-beta.1"},
		{"2.2.0-alpha", "2.2.0-alpha.1"},
		// Exact prerelease
		{"2.1.1-alpha.1", "2.1.1-alpha.1"},
	}

	for _, tt := range tests {
		includePrerelease := strings.Contains(tt.prefix, "-")
		var versions []string
		for _, r := range releases {
			if r.prerelease && !includePrerelease {
				continue
			}
			versions = append(versions, r.version)
		}

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

		if len(matched) == 0 {
			t.Errorf("resolve(%q): no matches", tt.prefix)
			continue
		}
		sortVersions(matched)
		got := matched[len(matched)-1]
		if got != tt.want {
			t.Errorf("resolve(%q) = %q, want %q", tt.prefix, got, tt.want)
		}
	}
}

func TestInstallBinaryReplacesFile(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "new-binary")
	dst := filepath.Join(dir, "zen")

	os.WriteFile(src, []byte("new-version"), 0755)
	os.WriteFile(dst, []byte("old-version"), 0755)

	if err := installBinary(src, dst, 0755); err != nil {
		t.Fatalf("installBinary failed: %v", err)
	}

	got, _ := os.ReadFile(dst)
	if string(got) != "new-version" {
		t.Errorf("content = %q, want %q", got, "new-version")
	}
}

func TestInstallBinaryPreservesPermissions(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "new-binary")
	dst := filepath.Join(dir, "zen")

	os.WriteFile(src, []byte("new"), 0755)
	os.WriteFile(dst, []byte("old"), 0700)

	if err := installBinary(src, dst, 0700); err != nil {
		t.Fatalf("installBinary failed: %v", err)
	}

	info, err := os.Stat(dst)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0700 {
		t.Errorf("permissions = %o, want 0700", info.Mode().Perm())
	}
}

func TestInstallBinaryFollowsSymlinks(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "new-binary")
	actual := filepath.Join(dir, "zen-real")
	link := filepath.Join(dir, "zen")

	os.WriteFile(src, []byte("new-version"), 0755)
	os.WriteFile(actual, []byte("old-version"), 0755)
	os.Symlink(actual, link)

	// Resolve expected path the same way (macOS /var → /private/var)
	expectedActual, _ := filepath.EvalSymlinks(actual)

	resolved := resolveInstallPath(link)
	if resolved != expectedActual {
		t.Errorf("resolveInstallPath = %q, want %q", resolved, expectedActual)
	}

	if err := installBinary(src, resolved, 0755); err != nil {
		t.Fatalf("installBinary failed: %v", err)
	}

	got, _ := os.ReadFile(expectedActual)
	if string(got) != "new-version" {
		t.Errorf("content = %q, want %q", got, "new-version")
	}

	// Symlink should still exist
	_, err := os.Readlink(link)
	if err != nil {
		t.Fatal("symlink should still exist:", err)
	}
}

func TestInstallBinaryDestNotExist(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "new-binary")
	dst := filepath.Join(dir, "zen-new")

	os.WriteFile(src, []byte("content"), 0755)

	if err := installBinary(src, dst, 0755); err != nil {
		t.Fatalf("installBinary failed: %v", err)
	}

	got, _ := os.ReadFile(dst)
	if string(got) != "content" {
		t.Errorf("content = %q, want %q", got, "content")
	}
}

func TestShouldUseTarball(t *testing.T) {
	tests := []struct {
		version string
		want    bool
	}{
		{"1.3.0", false},
		{"1.4.0", true},
		{"2.0.0", true},
		{"2.1.0-alpha.1", true},
		{"2.1.0-beta.1", true},
		{"1.3.0-rc.1", false},
		{"1.4.0-alpha.1", true},
	}

	for _, tt := range tests {
		got := shouldUseTarball(tt.version)
		if got != tt.want {
			t.Errorf("shouldUseTarball(%q) = %v, want %v", tt.version, got, tt.want)
		}
	}
}
