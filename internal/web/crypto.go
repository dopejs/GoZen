package web

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"strings"
)

const encPrefix = "ENC:"

// KeyPair holds an RSA key pair generated at server startup.
type KeyPair struct {
	Private *rsa.PrivateKey
	Public  *rsa.PublicKey
}

// GenerateKeyPair creates a new RSA-2048 key pair.
func GenerateKeyPair() (*KeyPair, error) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, fmt.Errorf("generate RSA key: %w", err)
	}
	return &KeyPair{Private: priv, Public: &priv.PublicKey}, nil
}

// PublicKeyPEM returns the PEM-encoded public key.
func (kp *KeyPair) PublicKeyPEM() (string, error) {
	der, err := x509.MarshalPKIXPublicKey(kp.Public)
	if err != nil {
		return "", err
	}
	block := &pem.Block{Type: "PUBLIC KEY", Bytes: der}
	return string(pem.EncodeToMemory(block)), nil
}

// DecryptToken decrypts a token that was encrypted with the public key.
// The ciphertext should be base64-encoded (after removing the "ENC:" prefix).
func (kp *KeyPair) DecryptToken(ciphertext string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", fmt.Errorf("base64 decode: %w", err)
	}
	plaintext, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, kp.Private, data, nil)
	if err != nil {
		return "", fmt.Errorf("RSA decrypt: %w", err)
	}
	return string(plaintext), nil
}

// MaybeDecryptToken checks if the value has the "ENC:" prefix and decrypts it.
// If no prefix, returns the value as-is (backward compatible).
func (kp *KeyPair) MaybeDecryptToken(value string) (string, error) {
	if !strings.HasPrefix(value, encPrefix) {
		return value, nil
	}
	return kp.DecryptToken(value[len(encPrefix):])
}
