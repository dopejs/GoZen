package web

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/dopejs/gozen/internal/config"
)

func TestGenerateKeyPair(t *testing.T) {
	kp, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("GenerateKeyPair: %v", err)
	}
	if kp.Private == nil || kp.Public == nil {
		t.Fatal("key pair should have both private and public keys")
	}
}

func TestPublicKeyPEM(t *testing.T) {
	kp, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("GenerateKeyPair: %v", err)
	}

	pem, err := kp.PublicKeyPEM()
	if err != nil {
		t.Fatalf("PublicKeyPEM: %v", err)
	}
	if len(pem) == 0 {
		t.Fatal("PEM should not be empty")
	}
	if pem[:27] != "-----BEGIN PUBLIC KEY-----\n" {
		t.Errorf("PEM should start with header, got: %s", pem[:27])
	}
}

func TestDecryptToken(t *testing.T) {
	kp, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("GenerateKeyPair: %v", err)
	}

	plaintext := "sk-test-secret-token-1234567890"

	// Encrypt with public key (simulating what the browser does)
	ciphertext, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, kp.Public, []byte(plaintext), nil)
	if err != nil {
		t.Fatalf("EncryptOAEP: %v", err)
	}
	encoded := base64.StdEncoding.EncodeToString(ciphertext)

	// Decrypt with private key
	decrypted, err := kp.DecryptToken(encoded)
	if err != nil {
		t.Fatalf("DecryptToken: %v", err)
	}
	if decrypted != plaintext {
		t.Errorf("decrypted = %q, want %q", decrypted, plaintext)
	}
}

func TestMaybeDecryptToken(t *testing.T) {
	kp, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("GenerateKeyPair: %v", err)
	}

	// Plaintext (no prefix) — should pass through
	plain := "sk-plain-token"
	result, err := kp.MaybeDecryptToken(plain)
	if err != nil {
		t.Fatalf("MaybeDecryptToken plain: %v", err)
	}
	if result != plain {
		t.Errorf("plain token: got %q, want %q", result, plain)
	}

	// Empty string — should pass through
	result, err = kp.MaybeDecryptToken("")
	if err != nil {
		t.Fatal("MaybeDecryptToken empty:", err)
	}
	if result != "" {
		t.Errorf("empty: got %q, want empty", result)
	}

	// Encrypted (with prefix)
	ciphertext, _ := rsa.EncryptOAEP(sha256.New(), rand.Reader, kp.Public, []byte("secret"), nil)
	encoded := "ENC:" + base64.StdEncoding.EncodeToString(ciphertext)

	result, err = kp.MaybeDecryptToken(encoded)
	if err != nil {
		t.Fatalf("MaybeDecryptToken encrypted: %v", err)
	}
	if result != "secret" {
		t.Errorf("encrypted: got %q, want %q", result, "secret")
	}
}

func TestMaybeDecryptTokenInvalidCiphertext(t *testing.T) {
	kp, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("GenerateKeyPair: %v", err)
	}

	_, err = kp.MaybeDecryptToken("ENC:not-valid-base64!!!")
	if err == nil {
		t.Error("expected error for invalid base64")
	}

	_, err = kp.MaybeDecryptToken("ENC:" + base64.StdEncoding.EncodeToString([]byte("garbage")))
	if err == nil {
		t.Error("expected error for invalid ciphertext")
	}
}

func TestPubKeyEndpoint(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	config.ResetDefaultStore()
	t.Cleanup(func() { config.ResetDefaultStore() })

	configDir := filepath.Join(dir, config.ConfigDir)
	os.MkdirAll(configDir, 0755)
	cfg := &config.OpenCCConfig{
		Providers: map[string]*config.ProviderConfig{},
		Profiles:  map[string]*config.ProfileConfig{"default": {Providers: []string{}}},
	}
	data, _ := json.MarshalIndent(cfg, "", "  ")
	os.WriteFile(filepath.Join(configDir, config.ConfigFile), data, 0600)
	config.DefaultStore()

	logger := log.New(io.Discard, "", 0)
	s := NewServer("test", logger, 0)

	req := httptest.NewRequest("GET", "/api/v1/auth/pubkey", nil)
	w := httptest.NewRecorder()
	s.handlePubKey(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("pubkey got %d, want 200", w.Code)
	}

	var resp map[string]string
	json.NewDecoder(w.Body).Decode(&resp)
	if resp["pubkey"] == "" {
		t.Error("pubkey should not be empty")
	}
}

func TestPubKeyEndpointMethodNotAllowed(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	config.ResetDefaultStore()
	t.Cleanup(func() { config.ResetDefaultStore() })

	configDir := filepath.Join(dir, config.ConfigDir)
	os.MkdirAll(configDir, 0755)
	cfg := &config.OpenCCConfig{
		Providers: map[string]*config.ProviderConfig{},
		Profiles:  map[string]*config.ProfileConfig{"default": {Providers: []string{}}},
	}
	data, _ := json.MarshalIndent(cfg, "", "  ")
	os.WriteFile(filepath.Join(configDir, config.ConfigFile), data, 0600)
	config.DefaultStore()

	logger := log.New(io.Discard, "", 0)
	s := NewServer("test", logger, 0)

	req := httptest.NewRequest("POST", "/api/v1/auth/pubkey", nil)
	w := httptest.NewRecorder()
	s.handlePubKey(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("pubkey POST got %d, want 405", w.Code)
	}
}

func TestCreateProviderWithEncryptedToken(t *testing.T) {
	s := setupTestServer(t)

	// Encrypt a token using the server's public key
	plainToken := "sk-encrypted-test-token"
	ciphertext, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, s.keys.Public, []byte(plainToken), nil)
	if err != nil {
		t.Fatalf("EncryptOAEP: %v", err)
	}
	encToken := "ENC:" + base64.StdEncoding.EncodeToString(ciphertext)

	body := map[string]interface{}{
		"name": "encrypted-provider",
		"config": map[string]string{
			"base_url":   "https://api.encrypted.com",
			"auth_token": encToken,
		},
	}

	w := doRequest(s, "POST", "/api/v1/providers", body)
	if w.Code != http.StatusCreated {
		t.Fatalf("create provider got %d, want 201", w.Code)
	}

	// Verify the token was decrypted and stored correctly
	p := config.GetProvider("encrypted-provider")
	if p == nil {
		t.Fatal("provider should exist")
	}
	if p.AuthToken != plainToken {
		t.Errorf("auth_token = %q, want %q", p.AuthToken, plainToken)
	}
}

func TestUpdateProviderWithEncryptedToken(t *testing.T) {
	s := setupTestServer(t)

	// Encrypt a new token using the server's public key
	newToken := "sk-new-encrypted-token"
	ciphertext, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, s.keys.Public, []byte(newToken), nil)
	if err != nil {
		t.Fatalf("EncryptOAEP: %v", err)
	}
	encToken := "ENC:" + base64.StdEncoding.EncodeToString(ciphertext)

	body := map[string]string{
		"auth_token": encToken,
		"base_url":   "https://api.updated.com",
	}

	w := doRequest(s, "PUT", "/api/v1/providers/test-provider", body)
	if w.Code != http.StatusOK {
		t.Fatalf("update provider got %d, want 200", w.Code)
	}

	// Verify the token was decrypted and stored correctly
	p := config.GetProvider("test-provider")
	if p == nil {
		t.Fatal("provider should exist")
	}
	if p.AuthToken != newToken {
		t.Errorf("auth_token = %q, want %q", p.AuthToken, newToken)
	}
}

func TestCreateProviderWithPlainToken(t *testing.T) {
	s := setupTestServer(t)

	body := map[string]interface{}{
		"name": "plain-provider",
		"config": map[string]string{
			"base_url":   "https://api.plain.com",
			"auth_token": "sk-plain-token-no-encryption",
		},
	}

	w := doRequest(s, "POST", "/api/v1/providers", body)
	if w.Code != http.StatusCreated {
		t.Fatalf("create provider got %d, want 201", w.Code)
	}

	p := config.GetProvider("plain-provider")
	if p == nil {
		t.Fatal("provider should exist")
	}
	if p.AuthToken != "sk-plain-token-no-encryption" {
		t.Errorf("auth_token = %q, want plain token", p.AuthToken)
	}
}
