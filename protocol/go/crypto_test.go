package protocol

import (
	"crypto/ed25519"
	"encoding/base64"
	"testing"
)

func TestGenerateKeyPair(t *testing.T) {
	pubKey, privKey, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("Failed to generate key pair: %v", err)
	}

	// Check key sizes
	if len(pubKey) != ed25519.PublicKeySize {
		t.Errorf("Expected public key size %d, got %d", ed25519.PublicKeySize, len(pubKey))
	}

	if len(privKey) != ed25519.PrivateKeySize {
		t.Errorf("Expected private key size %d, got %d", ed25519.PrivateKeySize, len(privKey))
	}

	// Check that public key matches private key
	derivedPubKey := privKey.Public().(ed25519.PublicKey)
	if !pubKey.Equal(derivedPubKey) {
		t.Error("Public key doesn't match derived public key from private key")
	}
}

func TestEncodeDecodePublicKey(t *testing.T) {
	// Generate test key
	pubKey, _, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("Failed to generate key pair: %v", err)
	}

	// Encode
	encoded := EncodePublicKey(pubKey)
	if encoded == "" {
		t.Error("Expected non-empty encoded public key")
	}

	// Check it's valid base64
	_, err = base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		t.Errorf("Encoded public key is not valid base64: %v", err)
	}

	// Decode
	decoded, err := DecodePublicKey(encoded)
	if err != nil {
		t.Errorf("Failed to decode public key: %v", err)
	}

	// Compare
	if !pubKey.Equal(decoded) {
		t.Error("Decoded public key doesn't match original")
	}
}

func TestEncodeDecodePrivateKey(t *testing.T) {
	// Generate test key
	_, privKey, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("Failed to generate key pair: %v", err)
	}

	// Encode
	encoded := EncodePrivateKey(privKey)
	if encoded == "" {
		t.Error("Expected non-empty encoded private key")
	}

	// Check it's valid base64
	_, err = base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		t.Errorf("Encoded private key is not valid base64: %v", err)
	}

	// Decode
	decoded, err := DecodePrivateKey(encoded)
	if err != nil {
		t.Errorf("Failed to decode private key: %v", err)
	}

	// Compare
	if !privKey.Equal(decoded) {
		t.Error("Decoded private key doesn't match original")
	}
}

func TestDecodePublicKeyErrors(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Invalid base64",
			input:    "not-base64!",
			expected: "invalid public key encoding",
		},
		{
			name:     "Wrong size",
			input:    base64.StdEncoding.EncodeToString([]byte("too-short")),
			expected: "invalid public key size",
		},
		{
			name:     "Empty string",
			input:    "",
			expected: "invalid public key size",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := DecodePublicKey(tt.input)
			if err == nil {
				t.Error("Expected error for invalid input")
			}
			if err != nil && !containsString(err.Error(), tt.expected) {
				t.Errorf("Expected error containing '%s', got '%s'", tt.expected, err.Error())
			}
		})
	}
}

func TestDecodePrivateKeyErrors(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Invalid base64",
			input:    "not-base64!",
			expected: "invalid private key encoding",
		},
		{
			name:     "Wrong size",
			input:    base64.StdEncoding.EncodeToString([]byte("wrong-size-key")),
			expected: "invalid private key size",
		},
		{
			name:     "Empty string",
			input:    "",
			expected: "invalid private key size",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := DecodePrivateKey(tt.input)
			if err == nil {
				t.Error("Expected error for invalid input")
			}
			if err != nil && !containsString(err.Error(), tt.expected) {
				t.Errorf("Expected error containing '%s', got '%s'", tt.expected, err.Error())
			}
		})
	}
}

func TestKeyPairRoundTrip(t *testing.T) {
	// Generate original key pair
	origPubKey, origPrivKey, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("Failed to generate key pair: %v", err)
	}

	// Encode both keys
	encodedPub := EncodePublicKey(origPubKey)
	encodedPriv := EncodePrivateKey(origPrivKey)

	// Decode both keys
	decodedPub, err := DecodePublicKey(encodedPub)
	if err != nil {
		t.Fatalf("Failed to decode public key: %v", err)
	}

	decodedPriv, err := DecodePrivateKey(encodedPriv)
	if err != nil {
		t.Fatalf("Failed to decode private key: %v", err)
	}

	// Test that decoded keys still work for signing
	message := []byte("test message")
	signature := ed25519.Sign(decodedPriv, message)

	// Verify with decoded public key
	if !ed25519.Verify(decodedPub, message, signature) {
		t.Error("Signature verification failed with decoded keys")
	}

	// Verify that decoded public key matches private key
	derivedPub := decodedPriv.Public().(ed25519.PublicKey)
	if !decodedPub.Equal(derivedPub) {
		t.Error("Decoded public key doesn't match decoded private key")
	}
}

func TestCryptoInteroperability(t *testing.T) {
	// Test that our encoding/decoding works with standard ed25519 operations
	pubKey, privKey, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatalf("Failed to generate key pair: %v", err)
	}

	// Encode with our functions
	encodedPub := EncodePublicKey(pubKey)
	encodedPriv := EncodePrivateKey(privKey)

	// Decode with our functions
	decodedPub, err := DecodePublicKey(encodedPub)
	if err != nil {
		t.Fatalf("Failed to decode public key: %v", err)
	}

	decodedPriv, err := DecodePrivateKey(encodedPriv)
	if err != nil {
		t.Fatalf("Failed to decode private key: %v", err)
	}

	// Test signing and verification
	message := []byte("interoperability test")
	
	// Sign with original private key
	sig1 := ed25519.Sign(privKey, message)
	
	// Sign with decoded private key
	sig2 := ed25519.Sign(decodedPriv, message)

	// Verify both signatures with both public keys
	if !ed25519.Verify(pubKey, message, sig1) {
		t.Error("Failed to verify signature 1 with original public key")
	}

	if !ed25519.Verify(decodedPub, message, sig1) {
		t.Error("Failed to verify signature 1 with decoded public key")
	}

	if !ed25519.Verify(pubKey, message, sig2) {
		t.Error("Failed to verify signature 2 with original public key")
	}

	if !ed25519.Verify(decodedPub, message, sig2) {
		t.Error("Failed to verify signature 2 with decoded public key")
	}
}

// Helper function to check if string contains substring
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && s[:len(substr)] == substr || 
		   (len(s) > len(substr) && s[len(s)-len(substr):] == substr) ||
		   (len(s) > len(substr)*2 && s[len(substr):len(s)-len(substr)] == substr)
}

func TestEncodingConsistency(t *testing.T) {
	// Test that multiple encodings of the same key produce the same result
	pubKey, privKey, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("Failed to generate key pair: %v", err)
	}

	// Encode multiple times
	encoded1 := EncodePublicKey(pubKey)
	encoded2 := EncodePublicKey(pubKey)
	encoded3 := EncodePrivateKey(privKey)
	encoded4 := EncodePrivateKey(privKey)

	if encoded1 != encoded2 {
		t.Error("Multiple encodings of the same public key should be identical")
	}

	if encoded3 != encoded4 {
		t.Error("Multiple encodings of the same private key should be identical")
	}
}