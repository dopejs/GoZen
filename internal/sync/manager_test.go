package sync

import (
	"context"
	"encoding/json"
	"sync"
	"testing"
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

func TestEncryptDecryptPayloadRoundTrip(t *testing.T) {
	// We can't easily test the full manager without a real config store,
	// but we can test encrypt/decrypt directly.
	payload := NewSyncPayload("dev1")
	provCfg, _ := json.Marshal(map[string]string{
		"auth_token": "sk-ant-secret",
		"base_url":   "https://api.anthropic.com",
	})
	payload.Providers["anthropic"] = &SyncEntity{Config: provCfg}

	// Encrypt with passphrase
	salt, _ := GenerateSalt()
	key := DeriveKey("test-pass", salt)

	for name, ent := range payload.Providers {
		enc, err := Encrypt(ent.Config, key)
		if err != nil {
			t.Fatal(err)
		}
		quoted, _ := json.Marshal(enc)
		payload.Providers[name] = &SyncEntity{
			ModifiedAt: ent.ModifiedAt,
			Config:     quoted,
		}
	}

	// Serialize
	data, err := json.Marshal(payload)
	if err != nil {
		t.Fatal(err)
	}

	// Deserialize and decrypt
	var restored SyncPayload
	if err := json.Unmarshal(data, &restored); err != nil {
		t.Fatal(err)
	}

	for name, ent := range restored.Providers {
		var encrypted string
		if err := json.Unmarshal(ent.Config, &encrypted); err != nil {
			t.Fatalf("provider %s config is not a string: %v", name, err)
		}
		decrypted, err := Decrypt(encrypted, key)
		if err != nil {
			t.Fatal(err)
		}
		var got map[string]string
		if err := json.Unmarshal(decrypted, &got); err != nil {
			t.Fatal(err)
		}
		if got["auth_token"] != "sk-ant-secret" {
			t.Fatalf("expected sk-ant-secret, got %s", got["auth_token"])
		}
	}
}
