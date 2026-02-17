package sync

import (
	"bytes"
	"testing"
)

func TestEncryptDecryptRoundTrip(t *testing.T) {
	salt, err := GenerateSalt()
	if err != nil {
		t.Fatal(err)
	}
	if len(salt) != saltSize {
		t.Fatalf("expected salt size %d, got %d", saltSize, len(salt))
	}

	key := DeriveKey("test-passphrase", salt)
	if len(key) != keySize {
		t.Fatalf("expected key size %d, got %d", keySize, len(key))
	}

	plaintext := []byte(`{"auth_token":"sk-ant-secret123","base_url":"https://api.example.com"}`)

	encrypted, err := Encrypt(plaintext, key)
	if err != nil {
		t.Fatal(err)
	}
	if encrypted == "" {
		t.Fatal("encrypted string is empty")
	}

	decrypted, err := Decrypt(encrypted, key)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(plaintext, decrypted) {
		t.Fatalf("decrypted does not match original:\n  got:  %s\n  want: %s", decrypted, plaintext)
	}
}

func TestDecryptWrongKey(t *testing.T) {
	salt, _ := GenerateSalt()
	key1 := DeriveKey("passphrase-1", salt)
	key2 := DeriveKey("passphrase-2", salt)

	plaintext := []byte("secret data")
	encrypted, err := Encrypt(plaintext, key1)
	if err != nil {
		t.Fatal(err)
	}

	_, err = Decrypt(encrypted, key2)
	if err == nil {
		t.Fatal("expected error decrypting with wrong key")
	}
}

func TestDecryptInvalidBase64(t *testing.T) {
	key := DeriveKey("pass", []byte("salt"))
	_, err := Decrypt("not-valid-base64!!!", key)
	if err == nil {
		t.Fatal("expected error for invalid base64")
	}
}

func TestDecryptTooShort(t *testing.T) {
	key := DeriveKey("pass", []byte("salt"))
	_, err := Decrypt("AQID", key) // 3 bytes, shorter than nonce
	if err == nil {
		t.Fatal("expected error for too-short ciphertext")
	}
}

func TestDifferentSaltsDifferentKeys(t *testing.T) {
	salt1, _ := GenerateSalt()
	salt2, _ := GenerateSalt()
	key1 := DeriveKey("same-passphrase", salt1)
	key2 := DeriveKey("same-passphrase", salt2)
	if bytes.Equal(key1, key2) {
		t.Fatal("different salts should produce different keys")
	}
}
