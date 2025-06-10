package protocol

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Capability represents a FEP capability token
type Capability struct {
	jwt.RegisteredClaims
	Scope       string   `json:"scope"`
	Permissions []string `json:"permissions"`
	Issuer      string   `json:"iss"`
	Subject     string   `json:"sub"`
}

// CapabilityManager handles capability token creation and validation
type CapabilityManager struct {
	signingKey []byte
}

// NewCapabilityManager creates a new capability manager
func NewCapabilityManager(signingKey []byte) *CapabilityManager {
	return &CapabilityManager{
		signingKey: signingKey,
	}
}

// CreateCapability creates a new capability token
func (cm *CapabilityManager) CreateCapability(scope, issuer, subject string, permissions []string, duration time.Duration) (string, error) {
	now := time.Now()
	claims := Capability{
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(duration)),
			ID:        generateNonce(),
		},
		Scope:       scope,
		Permissions: permissions,
		Issuer:      issuer,
		Subject:     subject,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(cm.signingKey)
}

// ValidateCapability validates a capability token
func (cm *CapabilityManager) ValidateCapability(tokenString string) (*Capability, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Capability{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return cm.signingKey, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Capability); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

// HasPermission checks if the capability has a specific permission
func (c *Capability) HasPermission(permission string) bool {
	for _, p := range c.Permissions {
		if p == permission || p == "*" {
			return true
		}
	}
	return false
}

// IsValid checks if the capability is currently valid
func (c *Capability) IsValid() bool {
	now := time.Now()
	if c.ExpiresAt != nil && now.After(c.ExpiresAt.Time) {
		return false
	}
	return true
}