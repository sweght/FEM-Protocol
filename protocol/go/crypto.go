package protocol

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"fmt"
)

// GenerateKeyPair generates a new Ed25519 key pair
func GenerateKeyPair() (ed25519.PublicKey, ed25519.PrivateKey, error) {
	return ed25519.GenerateKey(rand.Reader)
}

// EncodePublicKey encodes a public key to base64
func EncodePublicKey(pubKey ed25519.PublicKey) string {
	return base64.StdEncoding.EncodeToString(pubKey)
}

// DecodePublicKey decodes a base64 public key
func DecodePublicKey(encoded string) (ed25519.PublicKey, error) {
	data, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, fmt.Errorf("invalid public key encoding: %w", err)
	}
	
	if len(data) != ed25519.PublicKeySize {
		return nil, fmt.Errorf("invalid public key size: got %d, want %d", len(data), ed25519.PublicKeySize)
	}
	
	return ed25519.PublicKey(data), nil
}

// EncodePrivateKey encodes a private key to base64
func EncodePrivateKey(privKey ed25519.PrivateKey) string {
	return base64.StdEncoding.EncodeToString(privKey)
}

// DecodePrivateKey decodes a base64 private key
func DecodePrivateKey(encoded string) (ed25519.PrivateKey, error) {
	data, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, fmt.Errorf("invalid private key encoding: %w", err)
	}
	
	if len(data) != ed25519.PrivateKeySize {
		return nil, fmt.Errorf("invalid private key size: got %d, want %d", len(data), ed25519.PrivateKeySize)
	}
	
	return ed25519.PrivateKey(data), nil
}