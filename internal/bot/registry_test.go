package bot

import (
	"net"
	"testing"
	"time"
)

func TestNewRegistry(t *testing.T) {
	aliases := map[string]string{
		"myproject": "/path/to/project", // alias -> path
	}
	r := NewRegistry(aliases)

	if r == nil {
		t.Fatal("NewRegistry returned nil")
	}
	if len(r.aliases) != 1 {
		t.Errorf("expected 1 alias, got %d", len(r.aliases))
	}
}

func TestRegistry_Register(t *testing.T) {
	r := NewRegistry(nil)

	// Create a mock connection
	server, client := net.Pipe()
	defer server.Close()
	defer client.Close()

	info := &ProcessInfo{
		ID:        "test-1",
		Path:      "/path/to/api",
		PID:       1234,
		Status:    "idle",
		StartTime: time.Now(),
	}

	r.Register(info, client)

	// Check that process was registered
	if len(r.processes) != 1 {
		t.Errorf("expected 1 process, got %d", len(r.processes))
	}

	// Check auto-generated name
	if info.Name != "api" {
		t.Errorf("expected name 'api', got '%s'", info.Name)
	}
}

func TestRegistry_RegisterDuplicatePath(t *testing.T) {
	r := NewRegistry(nil)

	server1, client1 := net.Pipe()
	defer server1.Close()
	defer client1.Close()

	server2, client2 := net.Pipe()
	defer server2.Close()
	defer client2.Close()

	info1 := &ProcessInfo{
		ID:        "test-1",
		Path:      "/path/to/api",
		PID:       1234,
		Status:    "idle",
		StartTime: time.Now(),
	}

	info2 := &ProcessInfo{
		ID:        "test-2",
		Path:      "/path/to/api",
		PID:       5678,
		Status:    "idle",
		StartTime: time.Now(),
	}

	r.Register(info1, client1)
	r.Register(info2, client2)

	// Check that both processes were registered with unique names
	if len(r.processes) != 2 {
		t.Errorf("expected 2 processes, got %d", len(r.processes))
	}

	if info1.Name != "api" {
		t.Errorf("expected first name 'api', got '%s'", info1.Name)
	}

	if info2.Name != "api#2" {
		t.Errorf("expected second name 'api#2', got '%s'", info2.Name)
	}
}

func TestRegistry_RegisterWithAlias(t *testing.T) {
	// aliases map is alias -> path
	aliases := map[string]string{
		"proj": "/path/to/myproject",
	}
	r := NewRegistry(aliases)

	server, client := net.Pipe()
	defer server.Close()
	defer client.Close()

	info := &ProcessInfo{
		ID:        "test-1",
		Path:      "/path/to/myproject",
		PID:       1234,
		Status:    "idle",
		StartTime: time.Now(),
	}

	r.Register(info, client)

	if info.Alias != "proj" {
		t.Errorf("expected alias 'proj', got '%s'", info.Alias)
	}
}

func TestRegistry_Unregister(t *testing.T) {
	r := NewRegistry(nil)

	server, client := net.Pipe()
	defer server.Close()
	defer client.Close()

	info := &ProcessInfo{
		ID:        "test-1",
		Path:      "/path/to/api",
		PID:       1234,
		Status:    "idle",
		StartTime: time.Now(),
	}

	r.Register(info, client)
	r.Unregister("test-1")

	if len(r.processes) != 0 {
		t.Errorf("expected 0 processes after unregister, got %d", len(r.processes))
	}
}

func TestRegistry_Get(t *testing.T) {
	r := NewRegistry(nil)

	server, client := net.Pipe()
	defer server.Close()
	defer client.Close()

	info := &ProcessInfo{
		ID:        "test-1",
		Path:      "/path/to/api",
		PID:       1234,
		Status:    "idle",
		StartTime: time.Now(),
	}

	r.Register(info, client)

	got := r.Get("test-1")
	if got == nil {
		t.Fatal("Get returned nil for existing process")
	}
	if got.ID != "test-1" {
		t.Errorf("expected ID 'test-1', got '%s'", got.ID)
	}

	// Test non-existent
	got = r.Get("non-existent")
	if got != nil {
		t.Error("Get should return nil for non-existent process")
	}
}

func TestRegistry_Find(t *testing.T) {
	// aliases map is alias -> path
	aliases := map[string]string{
		"proj": "/path/to/myproject",
	}
	r := NewRegistry(aliases)

	server, client := net.Pipe()
	defer server.Close()
	defer client.Close()

	info := &ProcessInfo{
		ID:        "test-1",
		Path:      "/path/to/myproject",
		PID:       1234,
		Status:    "idle",
		StartTime: time.Now(),
	}

	r.Register(info, client)

	// Find by name
	got := r.Find("myproject")
	if got == nil {
		t.Fatal("Find by name returned nil")
	}

	// Find by alias
	got = r.Find("proj")
	if got == nil {
		t.Fatal("Find by alias returned nil")
	}

	// Find non-existent
	got = r.Find("non-existent")
	if got != nil {
		t.Error("Find should return nil for non-existent")
	}
}

func TestRegistry_List(t *testing.T) {
	r := NewRegistry(nil)

	// Empty list
	list := r.List()
	if len(list) != 0 {
		t.Errorf("expected empty list, got %d", len(list))
	}

	server, client := net.Pipe()
	defer server.Close()
	defer client.Close()

	info := &ProcessInfo{
		ID:        "test-1",
		Path:      "/path/to/api",
		PID:       1234,
		Status:    "idle",
		StartTime: time.Now(),
	}

	r.Register(info, client)

	list = r.List()
	if len(list) != 1 {
		t.Errorf("expected 1 process, got %d", len(list))
	}
}

func TestRegistry_UpdateStatus(t *testing.T) {
	r := NewRegistry(nil)

	server, client := net.Pipe()
	defer server.Close()
	defer client.Close()

	info := &ProcessInfo{
		ID:        "test-1",
		Path:      "/path/to/api",
		PID:       1234,
		Status:    "idle",
		StartTime: time.Now(),
	}

	r.Register(info, client)
	r.UpdateStatus("test-1", "busy", "running tests")

	got := r.Get("test-1")
	if got.Status != "busy" {
		t.Errorf("expected status 'busy', got '%s'", got.Status)
	}
	if got.CurrentTask != "running tests" {
		t.Errorf("expected task 'running tests', got '%s'", got.CurrentTask)
	}
}

func TestRegistry_CleanupStale(t *testing.T) {
	r := NewRegistry(nil)

	server, client := net.Pipe()
	defer server.Close()
	defer client.Close()

	info := &ProcessInfo{
		ID:        "test-1",
		Path:      "/path/to/api",
		PID:       1234,
		Status:    "idle",
		StartTime: time.Now(),
	}

	r.Register(info, client)

	// Manually set LastSeen to old time after registration
	r.mu.Lock()
	r.processes["test-1"].LastSeen = time.Now().Add(-time.Hour)
	r.mu.Unlock()

	// Cleanup with 30 second threshold should remove the stale process
	removed := r.CleanupStale(30 * time.Second)

	if len(removed) != 1 {
		t.Errorf("expected 1 removed, got %d", len(removed))
	}

	if len(r.processes) != 0 {
		t.Errorf("expected 0 processes after cleanup, got %d", len(r.processes))
	}
}

func TestRegistry_CleanupStale_KeepRecent(t *testing.T) {
	r := NewRegistry(nil)

	server, client := net.Pipe()
	defer server.Close()
	defer client.Close()

	info := &ProcessInfo{
		ID:        "test-1",
		Path:      "/path/to/api",
		PID:       1234,
		Status:    "idle",
		StartTime: time.Now(),
	}

	r.Register(info, client)

	// Don't modify LastSeen, it should be recent
	removed := r.CleanupStale(30 * time.Second)

	if len(removed) != 0 {
		t.Errorf("expected 0 removed, got %d", len(removed))
	}

	if len(r.processes) != 1 {
		t.Errorf("expected 1 process to remain, got %d", len(r.processes))
	}
}

func TestRegistry_GetConnection(t *testing.T) {
	r := NewRegistry(nil)

	server, client := net.Pipe()
	defer server.Close()
	defer client.Close()

	info := &ProcessInfo{
		ID:        "test-1",
		Path:      "/path/to/api",
		PID:       1234,
		Status:    "idle",
		StartTime: time.Now(),
	}

	r.Register(info, client)

	conn := r.GetConnection("test-1")
	if conn == nil {
		t.Fatal("GetConnection returned nil for existing process")
	}

	conn = r.GetConnection("non-existent")
	if conn != nil {
		t.Error("GetConnection should return nil for non-existent process")
	}
}

func TestRegistry_Count(t *testing.T) {
	r := NewRegistry(nil)

	if r.Count() != 0 {
		t.Errorf("expected count 0, got %d", r.Count())
	}

	server, client := net.Pipe()
	defer server.Close()
	defer client.Close()

	info := &ProcessInfo{
		ID:        "test-1",
		Path:      "/path/to/api",
		PID:       1234,
		Status:    "idle",
		StartTime: time.Now(),
	}

	r.Register(info, client)

	if r.Count() != 1 {
		t.Errorf("expected count 1, got %d", r.Count())
	}
}

func TestRegistry_SetAlias(t *testing.T) {
	r := NewRegistry(nil)

	server, client := net.Pipe()
	defer server.Close()
	defer client.Close()

	info := &ProcessInfo{
		ID:        "test-1",
		Path:      "/path/to/api",
		PID:       1234,
		Status:    "idle",
		StartTime: time.Now(),
	}

	r.Register(info, client)

	// Set alias after registration
	r.SetAlias("myapi", "/path/to/api")

	got := r.Get("test-1")
	if got.Alias != "myapi" {
		t.Errorf("expected alias 'myapi', got '%s'", got.Alias)
	}

	// Should be findable by alias
	found := r.Find("myapi")
	if found == nil {
		t.Error("Find by new alias returned nil")
	}
}

func TestExtractDirName(t *testing.T) {
	tests := []struct {
		path     string
		expected string
	}{
		{"/path/to/api", "api"},
		{"/path/to/api/", "api"},
		{"api", "api"},
		{"/", ""},
		{"", "unknown"},
	}

	for _, tt := range tests {
		result := extractDirName(tt.path)
		if result != tt.expected {
			t.Errorf("extractDirName(%q) = %q, want %q", tt.path, result, tt.expected)
		}
	}
}

func TestProcessInfo_ToJSON(t *testing.T) {
	info := &ProcessInfo{
		ID:     "test-1",
		Name:   "api",
		Path:   "/path/to/api",
		PID:    1234,
		Status: "idle",
	}

	data, err := info.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON failed: %v", err)
	}

	if len(data) == 0 {
		t.Error("ToJSON returned empty data")
	}
}
