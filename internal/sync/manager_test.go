package sync

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/dopejs/gozen/internal/config"
)

// mockBackend is an in-memory Backend for testing.
type mockBackend struct {
	mu   sync.Mutex
	data []byte
	name string
}

func newMockBackend() *mockBackend {
	return &mockBackend{name: "mock"}
}

func (b *mockBackend) Name() string { return b.name }

func (b *mockBackend) Download(_ context.Context) ([]byte, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.data == nil {
		return nil, nil
	}
	cp := make([]byte, len(b.data))
	copy(cp, b.data)
	return cp, nil
}

func (b *mockBackend) Upload(_ context.Context, data []byte) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.data = make([]byte, len(data))
	copy(b.data, data)
	return nil
}

// setupTestEnv sets HOME to a temp dir and resets the default store.
func setupTestEnv(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	config.ResetDefaultStore()
	t.Cleanup(func() { config.ResetDefaultStore() })
	return dir
}

// newTestManager creates a SyncManager with a mock backend for testing.
func newTestManager(t *testing.T, passphrase string) (*SyncManager, *mockBackend) {
	t.Helper()
	home := setupTestEnv(t)

	// Ensure config dir exists
	os.MkdirAll(filepath.Join(home, ".zen"), 0755)

	mock := newMockBackend()
	meta := NewSyncMeta("test-device")
	cfg := &config.SyncConfig{
		Backend:    "mock",
		Passphrase: passphrase,
	}
	mgr := &SyncManager{
		backend: mock,
		cfg:     cfg,
		meta:    meta,
	}
	return mgr, mock
}

func TestManagerIsPulling(t *testing.T) {
	mgr, _ := newTestManager(t, "")
	if mgr.IsPulling() {
		t.Fatal("should not be pulling initially")
	}
}

func TestManagerStatus(t *testing.T) {
	mgr, _ := newTestManager(t, "")
	status := mgr.Status()
	if !status.Configured {
		t.Fatal("should be configured")
	}
	if status.Backend != "mock" {
		t.Fatalf("expected mock backend, got %s", status.Backend)
	}
	if status.DeviceID != "test-device" {
		t.Fatalf("expected test-device, got %s", status.DeviceID)
	}
}

func TestManagerTestConnection(t *testing.T) {
	mgr, _ := newTestManager(t, "")
	err := mgr.TestConnection(context.Background())
	if err != nil {
		t.Fatalf("TestConnection should succeed with mock: %v", err)
	}
}

func TestManagerPushEmpty(t *testing.T) {
	mgr, mock := newTestManager(t, "")

	// Set up a provider in the config store
	store := config.DefaultStore()
	store.SetProvider("test-prov", &config.ProviderConfig{
		BaseURL:   "https://api.example.com",
		AuthToken: "sk-test-123",
	})
	store.SetProfileOrder("default", []string{"test-prov"})

	ctx := context.Background()
	if err := mgr.Push(ctx); err != nil {
		t.Fatalf("Push failed: %v", err)
	}

	// Verify data was uploaded
	if mock.data == nil {
		t.Fatal("expected data to be uploaded")
	}

	// Verify payload structure
	var payload SyncPayload
	if err := json.Unmarshal(mock.data, &payload); err != nil {
		t.Fatalf("invalid payload JSON: %v", err)
	}
	if _, ok := payload.Providers["test-prov"]; !ok {
		t.Fatal("provider test-prov should be in payload")
	}
	if _, ok := payload.Profiles["default"]; !ok {
		t.Fatal("profile default should be in payload")
	}
	if payload.DefaultProfile == nil {
		t.Fatal("default_profile should be set")
	}

	// Verify meta was updated
	if mgr.meta.LastPushAt.IsZero() {
		t.Fatal("LastPushAt should be set after push")
	}
}

func TestManagerPullEmpty(t *testing.T) {
	mgr, _ := newTestManager(t, "")
	// Pull with no remote data should succeed silently
	if err := mgr.Pull(context.Background()); err != nil {
		t.Fatalf("Pull with no remote data should succeed: %v", err)
	}
}

func TestManagerPullWithRemoteData(t *testing.T) {
	mgr, mock := newTestManager(t, "")

	// Set up local provider
	store := config.DefaultStore()
	store.SetProvider("local-prov", &config.ProviderConfig{
		BaseURL:   "https://local.example.com",
		AuthToken: "sk-local",
	})
	store.SetProfileOrder("default", []string{"local-prov"})

	// Create remote payload with a different provider
	remoteProv, _ := json.Marshal(&config.ProviderConfig{
		BaseURL:   "https://remote.example.com",
		AuthToken: "sk-remote",
	})
	remotePayload := NewSyncPayload("remote-device")
	remotePayload.Providers["remote-prov"] = &SyncEntity{
		ModifiedAt: time.Now().UTC(),
		Config:     remoteProv,
	}
	remoteData, _ := json.MarshalIndent(remotePayload, "", "  ")
	mock.data = remoteData

	ctx := context.Background()
	if err := mgr.Pull(ctx); err != nil {
		t.Fatalf("Pull failed: %v", err)
	}

	// Verify remote provider was applied locally
	p := store.GetProvider("remote-prov")
	if p == nil {
		t.Fatal("remote-prov should exist after pull")
	}
	if p.BaseURL != "https://remote.example.com" {
		t.Fatalf("expected remote URL, got %s", p.BaseURL)
	}

	// Verify meta was updated
	if mgr.meta.LastPullAt.IsZero() {
		t.Fatal("LastPullAt should be set after pull")
	}
}

func TestManagerPushThenPull(t *testing.T) {
	// Push from one manager, pull from another
	mgr1, mock := newTestManager(t, "")

	store := config.DefaultStore()
	store.SetProvider("shared-prov", &config.ProviderConfig{
		BaseURL:   "https://shared.example.com",
		AuthToken: "sk-shared",
	})
	store.SetProfileOrder("default", []string{"shared-prov"})

	ctx := context.Background()
	if err := mgr1.Push(ctx); err != nil {
		t.Fatalf("Push failed: %v", err)
	}

	// Create a second manager (simulating another device) with same mock backend
	home2 := t.TempDir()
	t.Setenv("HOME", home2)
	config.ResetDefaultStore()
	os.MkdirAll(filepath.Join(home2, ".zen"), 0755)

	meta2 := NewSyncMeta("device-2")
	mgr2 := &SyncManager{
		backend: mock,
		cfg:     &config.SyncConfig{Backend: "mock"},
		meta:    meta2,
	}

	if err := mgr2.Pull(ctx); err != nil {
		t.Fatalf("Pull failed: %v", err)
	}

	// Verify the provider was synced
	store2 := config.DefaultStore()
	p := store2.GetProvider("shared-prov")
	if p == nil {
		t.Fatal("shared-prov should exist after pull on device 2")
	}
	if p.AuthToken != "sk-shared" {
		t.Fatalf("expected sk-shared, got %s", p.AuthToken)
	}
}

func TestManagerEncryptDecryptRoundTrip(t *testing.T) {
	mgr, _ := newTestManager(t, "my-secret-passphrase")

	payload := NewSyncPayload("test-device")
	provCfg, _ := json.Marshal(&config.ProviderConfig{
		BaseURL:   "https://api.example.com",
		AuthToken: "sk-super-secret-token",
	})
	payload.Providers["encrypted-prov"] = &SyncEntity{
		ModifiedAt: time.Now().UTC(),
		Config:     provCfg,
	}

	// Encrypt
	encrypted, err := mgr.encryptPayload(payload)
	if err != nil {
		t.Fatalf("encryptPayload failed: %v", err)
	}

	// Verify raw encrypted data doesn't contain the token
	if json.Valid(encrypted) {
		raw := string(encrypted)
		if contains(raw, "sk-super-secret-token") {
			t.Fatal("encrypted payload should not contain plaintext token")
		}
	}

	// Decrypt
	decrypted, err := mgr.decryptPayload(encrypted)
	if err != nil {
		t.Fatalf("decryptPayload failed: %v", err)
	}

	var got config.ProviderConfig
	if err := json.Unmarshal(decrypted.Providers["encrypted-prov"].Config, &got); err != nil {
		t.Fatalf("unmarshal decrypted provider: %v", err)
	}
	if got.AuthToken != "sk-super-secret-token" {
		t.Fatalf("expected sk-super-secret-token, got %s", got.AuthToken)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchString(s, substr)
}

func searchString(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}

func TestManagerEncryptNoPassphrase(t *testing.T) {
	mgr, _ := newTestManager(t, "")

	payload := NewSyncPayload("test-device")
	provCfg, _ := json.Marshal(&config.ProviderConfig{AuthToken: "sk-plain"})
	payload.Providers["p1"] = &SyncEntity{
		ModifiedAt: time.Now().UTC(),
		Config:     provCfg,
	}

	data, err := mgr.encryptPayload(payload)
	if err != nil {
		t.Fatal(err)
	}

	// Without passphrase, config should be plaintext JSON object
	decrypted, err := mgr.decryptPayload(data)
	if err != nil {
		t.Fatal(err)
	}
	var got config.ProviderConfig
	json.Unmarshal(decrypted.Providers["p1"].Config, &got)
	if got.AuthToken != "sk-plain" {
		t.Fatalf("expected sk-plain, got %s", got.AuthToken)
	}
}

func TestManagerPushSkipsDuringPull(t *testing.T) {
	mgr, mock := newTestManager(t, "")

	// Simulate pulling state
	mgr.mu.Lock()
	mgr.isPulling = true
	mgr.mu.Unlock()

	ctx := context.Background()
	if err := mgr.Push(ctx); err != nil {
		t.Fatalf("Push during pull should return nil, got: %v", err)
	}

	// Nothing should be uploaded
	if mock.data != nil {
		t.Fatal("no data should be uploaded during pull")
	}
}

func TestManagerMarkDeleted(t *testing.T) {
	mgr, _ := newTestManager(t, "")
	mgr.meta.Providers["to-delete"] = time.Now().UTC()

	mgr.MarkDeleted("provider", "to-delete")

	if _, ok := mgr.meta.Tombstones["provider:to-delete"]; !ok {
		t.Fatal("tombstone should be created")
	}
	if _, ok := mgr.meta.Providers["to-delete"]; ok {
		t.Fatal("provider timestamp should be removed from meta")
	}
}

func TestManagerMarkDeletedProfile(t *testing.T) {
	mgr, _ := newTestManager(t, "")
	mgr.meta.Profiles["old-profile"] = time.Now().UTC()

	mgr.MarkDeleted("profile", "old-profile")

	if _, ok := mgr.meta.Tombstones["profile:old-profile"]; !ok {
		t.Fatal("tombstone should be created")
	}
	if _, ok := mgr.meta.Profiles["old-profile"]; ok {
		t.Fatal("profile timestamp should be removed from meta")
	}
}

func TestManagerMarkModified(t *testing.T) {
	mgr, _ := newTestManager(t, "")

	mgr.MarkModified("provider", "p1")
	if mgr.meta.Providers["p1"].IsZero() {
		t.Fatal("provider timestamp should be set")
	}

	mgr.MarkModified("profile", "default")
	if mgr.meta.Profiles["default"].IsZero() {
		t.Fatal("profile timestamp should be set")
	}

	mgr.MarkModified("default_profile", "")
	if mgr.meta.DefaultProfile.IsZero() {
		t.Fatal("default_profile timestamp should be set")
	}
}

func TestGenerateDeviceID(t *testing.T) {
	id := generateDeviceID()
	if id == "" {
		t.Fatal("device ID should not be empty")
	}
	if len(id) != 8 { // 4 bytes = 8 hex chars
		t.Fatalf("expected 8 hex chars, got %d: %s", len(id), id)
	}
	// Should be different each time
	id2 := generateDeviceID()
	if id == id2 {
		t.Fatal("two generated IDs should differ")
	}
}

func TestNewBackendFactory(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *config.SyncConfig
		wantErr bool
		backend string
	}{
		{"nil config", nil, true, ""},
		{"empty backend", &config.SyncConfig{}, true, ""},
		{"unknown backend", &config.SyncConfig{Backend: "ftp"}, true, ""},
		{"webdav", &config.SyncConfig{Backend: "webdav", Endpoint: "https://dav.example.com/f"}, false, "webdav"},
		{"gist", &config.SyncConfig{Backend: "gist", GistID: "abc", Token: "ghp_x"}, false, "gist"},
		{"repo", &config.SyncConfig{Backend: "repo", RepoOwner: "u", RepoName: "r", Token: "ghp_x"}, false, "repo"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, err := NewBackend(tt.cfg)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if b.Name() != tt.backend {
				t.Fatalf("expected %s, got %s", tt.backend, b.Name())
			}
		})
	}
}

func TestSyncMetaSaveLoad(t *testing.T) {
	home := setupTestEnv(t)
	os.MkdirAll(filepath.Join(home, ".zen"), 0755)

	meta := NewSyncMeta("dev-123")
	meta.Providers["p1"] = time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC)
	meta.LastPushAt = time.Date(2025, 6, 2, 0, 0, 0, 0, time.UTC)

	if err := meta.Save(); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	loaded, err := LoadSyncMeta()
	if err != nil {
		t.Fatalf("LoadSyncMeta failed: %v", err)
	}
	if loaded == nil {
		t.Fatal("loaded meta should not be nil")
	}
	if loaded.DeviceID != "dev-123" {
		t.Fatalf("expected dev-123, got %s", loaded.DeviceID)
	}
	if loaded.Providers["p1"].IsZero() {
		t.Fatal("provider timestamp should be preserved")
	}
}

func TestSyncMetaLoadNotExist(t *testing.T) {
	setupTestEnv(t)
	meta, err := LoadSyncMeta()
	if err != nil {
		t.Fatalf("LoadSyncMeta should not error for missing file: %v", err)
	}
	if meta != nil {
		t.Fatal("meta should be nil when file doesn't exist")
	}
}

func TestUpdateMetaTimestamps(t *testing.T) {
	mgr, _ := newTestManager(t, "")
	mgr.meta.Providers["old-prov"] = time.Now().UTC()

	now := time.Now().UTC()
	payload := NewSyncPayload("dev")
	payload.Providers["new-prov"] = &SyncEntity{ModifiedAt: now}
	payload.Profiles["default"] = &SyncEntity{ModifiedAt: now}
	payload.DefaultProfile = &SyncScalar{ModifiedAt: now, Value: "default"}

	mgr.updateMetaTimestamps(payload)

	if _, ok := mgr.meta.Providers["old-prov"]; ok {
		t.Fatal("old-prov should be cleaned up")
	}
	if mgr.meta.Providers["new-prov"].IsZero() {
		t.Fatal("new-prov should have timestamp")
	}
	if mgr.meta.Profiles["default"].IsZero() {
		t.Fatal("default profile should have timestamp")
	}
}
