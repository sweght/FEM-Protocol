package protocol

import (
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"
)

// EnvelopeType represents the type of FEP envelope
type EnvelopeType string

const (
	EnvelopeRegisterAgent      EnvelopeType = "registerAgent"
	EnvelopeRegisterBroker     EnvelopeType = "registerBroker"
	EnvelopeEmitEvent          EnvelopeType = "emitEvent"
	EnvelopeRenderInstruction  EnvelopeType = "renderInstruction"
	EnvelopeToolCall           EnvelopeType = "toolCall"
	EnvelopeToolResult         EnvelopeType = "toolResult"
	EnvelopeRevoke             EnvelopeType = "revoke"
)

// CommonHeaders contains headers present in all FEP envelopes
type CommonHeaders struct {
	Agent string `json:"agent"`           // UTF-8 agent identifier
	TS    int64  `json:"ts"`              // Unix timestamp in milliseconds
	Nonce string `json:"nonce"`           // Replay guard
	Sig   string `json:"sig,omitempty"`   // Base64(Ed25519(body))
}

// BaseEnvelope is the base structure for all FEP envelopes
type BaseEnvelope struct {
	Type EnvelopeType `json:"type"`
	CommonHeaders
}

// RegisterAgentEnvelope registers an agent in the system
type RegisterAgentEnvelope struct {
	BaseEnvelope
	Body RegisterAgentBody `json:"body"`
}

type RegisterAgentBody struct {
	PubKey       string   `json:"pubkey"`        // Base64 Ed25519 public key
	Capabilities []string `json:"capabilities"`  // List of capabilities
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// RegisterBrokerEnvelope registers a broker node
type RegisterBrokerEnvelope struct {
	BaseEnvelope
	Body RegisterBrokerBody `json:"body"`
}

type RegisterBrokerBody struct {
	BrokerID     string   `json:"brokerId"`
	Endpoint     string   `json:"endpoint"`      // TLS endpoint
	PubKey       string   `json:"pubkey"`        // Base64 Ed25519 public key
	Capabilities []string `json:"capabilities"`
}

// EmitEventEnvelope emits events from agents
type EmitEventEnvelope struct {
	BaseEnvelope
	Body EmitEventBody `json:"body"`
}

type EmitEventBody struct {
	Event   string                 `json:"event"`
	Payload map[string]interface{} `json:"payload"`
}

// RenderInstructionEnvelope sends rendering instructions
type RenderInstructionEnvelope struct {
	BaseEnvelope
	Body RenderInstructionBody `json:"body"`
}

type RenderInstructionBody struct {
	Instruction string                 `json:"instruction"`
	Parameters  map[string]interface{} `json:"parameters,omitempty"`
}

// ToolCallEnvelope requests tool execution
type ToolCallEnvelope struct {
	BaseEnvelope
	Body ToolCallBody `json:"body"`
}

type ToolCallBody struct {
	Tool       string                 `json:"tool"`
	Parameters map[string]interface{} `json:"parameters"`
	RequestID  string                 `json:"requestId"`
}

// ToolResultEnvelope returns tool execution results
type ToolResultEnvelope struct {
	BaseEnvelope
	Body ToolResultBody `json:"body"`
}

type ToolResultBody struct {
	RequestID string                 `json:"requestId"`
	Success   bool                   `json:"success"`
	Result    interface{}            `json:"result,omitempty"`
	Error     string                 `json:"error,omitempty"`
}

// RevokeEnvelope revokes registrations/capabilities
type RevokeEnvelope struct {
	BaseEnvelope
	Body RevokeBody `json:"body"`
}

type RevokeBody struct {
	Target string `json:"target"` // Agent or broker ID to revoke
	Reason string `json:"reason,omitempty"`
}

// Envelope is a generic envelope that can hold any envelope type
type Envelope struct {
	Type EnvelopeType `json:"type"`
	CommonHeaders
	Body json.RawMessage `json:"body"`
}

// Sign signs the envelope with the given private key
func (e *Envelope) Sign(privateKey ed25519.PrivateKey) error {
	// Remove existing signature
	e.Sig = ""
	
	// Marshal the envelope without signature
	data, err := json.Marshal(e)
	if err != nil {
		return err
	}
	
	// Sign the data
	signature := ed25519.Sign(privateKey, data)
	e.Sig = base64.StdEncoding.EncodeToString(signature)
	
	return nil
}

// Sign methods for specific envelope types
func (e *RegisterAgentEnvelope) Sign(privateKey ed25519.PrivateKey) error {
	// Remove existing signature
	e.Sig = ""
	
	// Marshal the envelope without signature
	data, err := json.Marshal(e)
	if err != nil {
		return err
	}
	
	// Sign the data
	signature := ed25519.Sign(privateKey, data)
	e.Sig = base64.StdEncoding.EncodeToString(signature)
	
	return nil
}

func (e *RegisterBrokerEnvelope) Sign(privateKey ed25519.PrivateKey) error {
	e.Sig = ""
	data, err := json.Marshal(e)
	if err != nil {
		return err
	}
	signature := ed25519.Sign(privateKey, data)
	e.Sig = base64.StdEncoding.EncodeToString(signature)
	return nil
}

func (e *ToolCallEnvelope) Sign(privateKey ed25519.PrivateKey) error {
	e.Sig = ""
	data, err := json.Marshal(e)
	if err != nil {
		return err
	}
	signature := ed25519.Sign(privateKey, data)
	e.Sig = base64.StdEncoding.EncodeToString(signature)
	return nil
}

func (e *ToolResultEnvelope) Sign(privateKey ed25519.PrivateKey) error {
	e.Sig = ""
	data, err := json.Marshal(e)
	if err != nil {
		return err
	}
	signature := ed25519.Sign(privateKey, data)
	e.Sig = base64.StdEncoding.EncodeToString(signature)
	return nil
}

// Verify verifies the envelope signature with the given public key
func (e *Envelope) Verify(publicKey ed25519.PublicKey) error {
	if e.Sig == "" {
		return fmt.Errorf("envelope has no signature")
	}
	
	// Decode signature
	signature, err := base64.StdEncoding.DecodeString(e.Sig)
	if err != nil {
		return fmt.Errorf("invalid signature encoding: %w", err)
	}
	
	// Store and remove signature
	sig := e.Sig
	e.Sig = ""
	defer func() { e.Sig = sig }()
	
	// Marshal envelope without signature
	data, err := json.Marshal(e)
	if err != nil {
		return err
	}
	
	// Verify signature
	if !ed25519.Verify(publicKey, data, signature) {
		return fmt.Errorf("signature verification failed")
	}
	
	return nil
}

// NewEnvelope creates a new envelope with common headers
func NewEnvelope(envType EnvelopeType, agent string) *Envelope {
	return &Envelope{
		Type: envType,
		CommonHeaders: CommonHeaders{
			Agent: agent,
			TS:    time.Now().UnixMilli(),
			Nonce: generateNonce(),
		},
	}
}

// generateNonce generates a random nonce for replay protection
func generateNonce() string {
	// In production, use crypto/rand
	return fmt.Sprintf("%d-%d", time.Now().UnixNano(), time.Now().Unix())
}