package protocol

import (
	"testing"
	"time"
	
	"github.com/golang-jwt/jwt/v5"
)

func TestNewCapabilityManager(t *testing.T) {
	key := []byte("test-signing-key")
	cm := NewCapabilityManager(key)

	if cm == nil {
		t.Fatal("Expected non-nil capability manager")
	}

	if len(cm.signingKey) != len(key) {
		t.Error("Signing key not properly stored")
	}
}

func TestCreateCapability(t *testing.T) {
	cm := NewCapabilityManager([]byte("test-key"))
	
	scope := "scope:local"
	issuer := "broker.test"
	subject := "agent.test"
	permissions := []string{"tool.execute", "event.emit"}
	duration := time.Hour

	token, err := cm.CreateCapability(scope, issuer, subject, permissions, duration)
	if err != nil {
		t.Fatalf("Failed to create capability: %v", err)
	}

	if token == "" {
		t.Error("Expected non-empty token")
	}

	// Token should be a valid JWT format (three parts separated by dots)
	parts := splitString(token, ".")
	if len(parts) != 3 {
		t.Errorf("Expected JWT with 3 parts, got %d", len(parts))
	}
}

func TestValidateCapability(t *testing.T) {
	cm := NewCapabilityManager([]byte("test-key"))
	
	scope := "scope:trusted"
	issuer := "broker.test"
	subject := "agent.test"
	permissions := []string{"tool.execute", "event.emit", "broker.register"}
	duration := time.Hour

	// Create capability
	token, err := cm.CreateCapability(scope, issuer, subject, permissions, duration)
	if err != nil {
		t.Fatalf("Failed to create capability: %v", err)
	}

	// Validate capability
	capability, err := cm.ValidateCapability(token)
	if err != nil {
		t.Fatalf("Failed to validate capability: %v", err)
	}

	// Check fields
	if capability.Scope != scope {
		t.Errorf("Expected scope %s, got %s", scope, capability.Scope)
	}

	if capability.Issuer != issuer {
		t.Errorf("Expected issuer %s, got %s", issuer, capability.Issuer)
	}

	if capability.Subject != subject {
		t.Errorf("Expected subject %s, got %s", subject, capability.Subject)
	}

	if len(capability.Permissions) != len(permissions) {
		t.Errorf("Expected %d permissions, got %d", len(permissions), len(capability.Permissions))
	}

	for i, perm := range permissions {
		if capability.Permissions[i] != perm {
			t.Errorf("Expected permission %s, got %s", perm, capability.Permissions[i])
		}
	}
}

func TestCapabilityHasPermission(t *testing.T) {
	capability := &Capability{
		Permissions: []string{"tool.execute", "event.emit", "broker.register"},
	}

	tests := []struct {
		permission string
		expected   bool
	}{
		{"tool.execute", true},
		{"event.emit", true},
		{"broker.register", true},
		{"admin.revoke", false},
		{"tool.read", false},
	}

	for _, tt := range tests {
		t.Run(tt.permission, func(t *testing.T) {
			result := capability.HasPermission(tt.permission)
			if result != tt.expected {
				t.Errorf("Expected %v for permission %s, got %v", tt.expected, tt.permission, result)
			}
		})
	}
}

func TestCapabilityWildcardPermission(t *testing.T) {
	capability := &Capability{
		Permissions: []string{"*"},
	}

	tests := []string{
		"tool.execute",
		"event.emit",
		"broker.register",
		"admin.revoke",
		"anything.else",
	}

	for _, permission := range tests {
		t.Run(permission, func(t *testing.T) {
			if !capability.HasPermission(permission) {
				t.Errorf("Expected wildcard to grant permission %s", permission)
			}
		})
	}
}

func TestCapabilityIsValid(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name     string
		expiry   *time.Time
		expected bool
	}{
		{
			name:     "Valid - future expiry",
			expiry:   func() *time.Time { t := now.Add(time.Hour); return &t }(),
			expected: true,
		},
		{
			name:     "Invalid - past expiry",
			expiry:   func() *time.Time { t := now.Add(-time.Hour); return &t }(),
			expected: false,
		},
		{
			name:     "Valid - no expiry",
			expiry:   nil,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			capability := &Capability{}
			if tt.expiry != nil {
				capability.ExpiresAt = jwt.NewNumericDate(*tt.expiry)
			}

			result := capability.IsValid()
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestValidateCapabilityWithWrongKey(t *testing.T) {
	cm1 := NewCapabilityManager([]byte("key1"))
	cm2 := NewCapabilityManager([]byte("key2"))

	// Create capability with first manager
	token, err := cm1.CreateCapability("scope:local", "issuer", "subject", []string{"test"}, time.Hour)
	if err != nil {
		t.Fatalf("Failed to create capability: %v", err)
	}

	// Try to validate with second manager (different key)
	_, err = cm2.ValidateCapability(token)
	if err == nil {
		t.Error("Expected validation to fail with wrong signing key")
	}
}

func TestValidateInvalidToken(t *testing.T) {
	cm := NewCapabilityManager([]byte("test-key"))

	tests := []string{
		"invalid.token.here",
		"not-a-jwt",
		"",
		"too.many.parts.in.this.token",
		"header.payload", // missing signature
	}

	for _, token := range tests {
		t.Run(token, func(t *testing.T) {
			_, err := cm.ValidateCapability(token)
			if err == nil {
				t.Error("Expected validation to fail for invalid token")
			}
		})
	}
}

func TestCapabilityExpiration(t *testing.T) {
	cm := NewCapabilityManager([]byte("test-key"))
	
	// Create capability with short duration
	token, err := cm.CreateCapability(
		"scope:local",
		"broker.test",
		"agent.test",
		[]string{"test"},
		time.Second*2,
	)
	if err != nil {
		t.Fatalf("Failed to create capability: %v", err)
	}

	// Should be valid immediately
	capability, err := cm.ValidateCapability(token)
	if err != nil {
		t.Fatalf("Failed to validate fresh capability: %v", err)
	}

	if !capability.IsValid() {
		t.Error("Fresh capability should be valid")
	}

	// Wait for expiration
	time.Sleep(time.Second * 3)

	// Should fail to parse because JWT library validates expiration
	_, err = cm.ValidateCapability(token)
	if err == nil {
		t.Error("Expected validation to fail for expired capability")
	}
}

func TestCapabilityRoundTrip(t *testing.T) {
	cm := NewCapabilityManager([]byte("round-trip-test-key"))
	
	originalScope := "scope:trusted"
	originalIssuer := "broker.roundtrip"
	originalSubject := "agent.roundtrip"
	originalPermissions := []string{"tool.execute", "event.emit", "admin.manage"}
	duration := time.Hour * 24

	// Create
	token, err := cm.CreateCapability(originalScope, originalIssuer, originalSubject, originalPermissions, duration)
	if err != nil {
		t.Fatalf("Failed to create capability: %v", err)
	}

	// Validate
	capability, err := cm.ValidateCapability(token)
	if err != nil {
		t.Fatalf("Failed to validate capability: %v", err)
	}

	// Check all fields match
	if capability.Scope != originalScope {
		t.Errorf("Scope mismatch: expected %s, got %s", originalScope, capability.Scope)
	}

	if capability.Issuer != originalIssuer {
		t.Errorf("Issuer mismatch: expected %s, got %s", originalIssuer, capability.Issuer)
	}

	if capability.Subject != originalSubject {
		t.Errorf("Subject mismatch: expected %s, got %s", originalSubject, capability.Subject)
	}

	if len(capability.Permissions) != len(originalPermissions) {
		t.Errorf("Permissions length mismatch: expected %d, got %d", len(originalPermissions), len(capability.Permissions))
	}

	for i, perm := range originalPermissions {
		if capability.Permissions[i] != perm {
			t.Errorf("Permission %d mismatch: expected %s, got %s", i, perm, capability.Permissions[i])
		}
	}

	// Check that it has all the expected permissions
	for _, perm := range originalPermissions {
		if !capability.HasPermission(perm) {
			t.Errorf("Capability should have permission %s", perm)
		}
	}

	// Check timestamps are reasonable
	if capability.IssuedAt == nil {
		t.Error("IssuedAt should be set")
	}

	if capability.ExpiresAt == nil {
		t.Error("ExpiresAt should be set")
	}

	if capability.ID == "" {
		t.Error("ID should be set")
	}
}

// Helper function to split string (simplified)
func splitString(s, sep string) []string {
	if s == "" {
		return []string{}
	}
	
	var result []string
	start := 0
	
	for i := 0; i <= len(s)-len(sep); i++ {
		if s[i:i+len(sep)] == sep {
			result = append(result, s[start:i])
			start = i + len(sep)
			i += len(sep) - 1
		}
	}
	result = append(result, s[start:])
	
	return result
}