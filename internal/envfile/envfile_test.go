package envfile

import (
	"os"
	"path/filepath"
	"sort"
	"testing"
)

// setTestHome overrides HOME to a temp dir and returns cleanup func.
func setTestHome(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	return dir
}

func TestEnvsPath(t *testing.T) {
	home := setTestHome(t)
	got := EnvsPath()
	want := filepath.Join(home, ".cc_envs")
	if got != want {
		t.Errorf("EnvsPath() = %q, want %q", got, want)
	}
}

func TestConfigGetSet(t *testing.T) {
	cfg := &Config{Name: "test", Entries: []Entry{
		{Key: "A", Value: "1"},
		{Key: "B", Value: "2"},
	}}

	if v := cfg.Get("A"); v != "1" {
		t.Errorf("Get(A) = %q, want %q", v, "1")
	}
	if v := cfg.Get("B"); v != "2" {
		t.Errorf("Get(B) = %q, want %q", v, "2")
	}
	if v := cfg.Get("C"); v != "" {
		t.Errorf("Get(C) = %q, want empty", v)
	}

	// Set existing key
	cfg.Set("A", "10")
	if v := cfg.Get("A"); v != "10" {
		t.Errorf("after Set(A,10), Get(A) = %q, want %q", v, "10")
	}

	// Set new key
	cfg.Set("C", "3")
	if v := cfg.Get("C"); v != "3" {
		t.Errorf("after Set(C,3), Get(C) = %q, want %q", v, "3")
	}
	if len(cfg.Entries) != 3 {
		t.Errorf("expected 3 entries, got %d", len(cfg.Entries))
	}
}

func TestLoadAndSave(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "work.env")

	content := "ANTHROPIC_BASE_URL=https://api.example.com\nANTHROPIC_AUTH_TOKEN=sk-123\nANTHROPIC_MODEL=opus\n"
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if cfg.Name != "work" {
		t.Errorf("Name = %q, want %q", cfg.Name, "work")
	}
	if v := cfg.Get("ANTHROPIC_BASE_URL"); v != "https://api.example.com" {
		t.Errorf("ANTHROPIC_BASE_URL = %q", v)
	}
	if v := cfg.Get("ANTHROPIC_AUTH_TOKEN"); v != "sk-123" {
		t.Errorf("ANTHROPIC_AUTH_TOKEN = %q", v)
	}
	if v := cfg.Get("ANTHROPIC_MODEL"); v != "opus" {
		t.Errorf("ANTHROPIC_MODEL = %q", v)
	}

	// Modify and save
	cfg.Set("ANTHROPIC_MODEL", "sonnet")
	if err := cfg.Save(); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	// Reload and verify
	cfg2, err := Load(path)
	if err != nil {
		t.Fatalf("Load() after save error: %v", err)
	}
	if v := cfg2.Get("ANTHROPIC_MODEL"); v != "sonnet" {
		t.Errorf("after save, ANTHROPIC_MODEL = %q, want %q", v, "sonnet")
	}
}

func TestLoadSkipsCommentsAndBlanks(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.env")

	content := `# This is a comment

KEY1=val1
KEY2=val2 # inline comment
# another comment
KEY3=val3
`
	os.WriteFile(path, []byte(content), 0644)

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if len(cfg.Entries) != 3 {
		t.Fatalf("expected 3 entries, got %d: %+v", len(cfg.Entries), cfg.Entries)
	}
	if v := cfg.Get("KEY1"); v != "val1" {
		t.Errorf("KEY1 = %q", v)
	}
	if v := cfg.Get("KEY2"); v != "val2" {
		t.Errorf("KEY2 = %q, want %q (inline comment should be stripped)", v, "val2")
	}
	if v := cfg.Get("KEY3"); v != "val3" {
		t.Errorf("KEY3 = %q", v)
	}
}

func TestLoadSkipsLinesWithoutEquals(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.env")

	content := "GOOD=value\nno_equals_here\nALSO_GOOD=yes\n"
	os.WriteFile(path, []byte(content), 0644)

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if len(cfg.Entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(cfg.Entries))
	}
}

func TestLoadNonexistent(t *testing.T) {
	_, err := Load("/nonexistent/path.env")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestDelete(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "del.env")
	os.WriteFile(path, []byte("K=V\n"), 0644)

	cfg, _ := Load(path)
	if err := cfg.Delete(); err != nil {
		t.Fatalf("Delete() error: %v", err)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Error("file should be deleted")
	}
}

func TestCreateAndLoadByName(t *testing.T) {
	setTestHome(t)

	entries := []Entry{
		{Key: "ANTHROPIC_BASE_URL", Value: "https://example.com"},
		{Key: "ANTHROPIC_AUTH_TOKEN", Value: "tok"},
	}
	cfg, err := Create("mytest", entries)
	if err != nil {
		t.Fatalf("Create() error: %v", err)
	}
	if cfg.Name != "mytest" {
		t.Errorf("Name = %q", cfg.Name)
	}

	// LoadByName
	cfg2, err := LoadByName("mytest")
	if err != nil {
		t.Fatalf("LoadByName() error: %v", err)
	}
	if v := cfg2.Get("ANTHROPIC_BASE_URL"); v != "https://example.com" {
		t.Errorf("ANTHROPIC_BASE_URL = %q", v)
	}

	// LoadByName nonexistent
	_, err = LoadByName("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent config")
	}
}

func TestListConfigsAndConfigNames(t *testing.T) {
	home := setTestHome(t)
	dir := filepath.Join(home, ".cc_envs")
	os.MkdirAll(dir, 0755)

	os.WriteFile(filepath.Join(dir, "alpha.env"), []byte("K=V\n"), 0644)
	os.WriteFile(filepath.Join(dir, "beta.env"), []byte("K=V\n"), 0644)
	os.WriteFile(filepath.Join(dir, "notenv.txt"), []byte("K=V\n"), 0644) // should be ignored

	configs, err := ListConfigs()
	if err != nil {
		t.Fatalf("ListConfigs() error: %v", err)
	}
	if len(configs) != 2 {
		t.Fatalf("expected 2 configs, got %d", len(configs))
	}

	names := ConfigNames()
	sort.Strings(names)
	if len(names) != 2 || names[0] != "alpha" || names[1] != "beta" {
		t.Errorf("ConfigNames() = %v, want [alpha beta]", names)
	}
}

func TestListConfigsEmptyDir(t *testing.T) {
	setTestHome(t)
	// Don't create .cc_envs dir

	configs, err := ListConfigs()
	if err != nil {
		t.Fatalf("ListConfigs() error: %v", err)
	}
	if len(configs) != 0 {
		t.Errorf("expected 0 configs, got %d", len(configs))
	}
}

func TestExportToEnv(t *testing.T) {
	cfg := &Config{
		Name: "test",
		Entries: []Entry{
			{Key: "TEST_OPENCC_A", Value: "hello"},
			{Key: "TEST_OPENCC_B", Value: "world"},
		},
	}

	cfg.ExportToEnv()

	if v := os.Getenv("TEST_OPENCC_A"); v != "hello" {
		t.Errorf("TEST_OPENCC_A = %q, want %q", v, "hello")
	}
	if v := os.Getenv("TEST_OPENCC_B"); v != "world" {
		t.Errorf("TEST_OPENCC_B = %q, want %q", v, "world")
	}

	// Cleanup
	os.Unsetenv("TEST_OPENCC_A")
	os.Unsetenv("TEST_OPENCC_B")
}

func TestSaveCreatesCorrectFormat(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "fmt.env")

	cfg := &Config{
		Name: "fmt",
		Path: path,
		Entries: []Entry{
			{Key: "A", Value: "1"},
			{Key: "B", Value: "2"},
		},
	}
	if err := cfg.Save(); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	data, _ := os.ReadFile(path)
	want := "A=1\nB=2\n"
	if string(data) != want {
		t.Errorf("saved content = %q, want %q", string(data), want)
	}
}

func TestCreateErrorBadDir(t *testing.T) {
	// Set HOME to a path that can't be created
	t.Setenv("HOME", "/dev/null/impossible")

	_, err := Create("test", []Entry{{Key: "K", Value: "V"}})
	if err == nil {
		t.Error("expected error when dir can't be created")
	}
}

func TestListConfigsWithBadFile(t *testing.T) {
	home := setTestHome(t)
	dir := filepath.Join(home, ".cc_envs")
	os.MkdirAll(dir, 0755)

	// Create a valid .env and a directory named "bad.env" (will fail to Load)
	os.WriteFile(filepath.Join(dir, "good.env"), []byte("K=V\n"), 0644)
	os.MkdirAll(filepath.Join(dir, "bad.env"), 0755) // directory, not file

	configs, err := ListConfigs()
	if err != nil {
		t.Fatalf("ListConfigs() error: %v", err)
	}
	// Should only get the good one, bad one is skipped
	if len(configs) != 1 {
		t.Errorf("expected 1 config, got %d", len(configs))
	}
	if configs[0].Name != "good" {
		t.Errorf("config name = %q, want %q", configs[0].Name, "good")
	}
}
