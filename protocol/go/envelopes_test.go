package protocol

import (
	"crypto/ed25519"
	"encoding/json"
	"testing"
	"time"
)

func TestEnvelopeTypes(t *testing.T) {
	tests := []struct {
		name     string
		envType  EnvelopeType
		expected string
	}{
		{"RegisterAgent", EnvelopeRegisterAgent, "registerAgent"},
		{"RegisterBroker", EnvelopeRegisterBroker, "registerBroker"},
		{"EmitEvent", EnvelopeEmitEvent, "emitEvent"},
		{"RenderInstruction", EnvelopeRenderInstruction, "renderInstruction"},
		{"ToolCall", EnvelopeToolCall, "toolCall"},
		{"ToolResult", EnvelopeToolResult, "toolResult"},
		{"Revoke", EnvelopeRevoke, "revoke"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.envType) != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, string(tt.envType))
			}
		})
	}
}

func TestNewEnvelope(t *testing.T) {
	envType := EnvelopeRegisterAgent
	agent := "test.agent"

	envelope := NewEnvelope(envType, agent)

	if envelope.Type != envType {
		t.Errorf("Expected type %s, got %s", envType, envelope.Type)
	}

	if envelope.Agent != agent {
		t.Errorf("Expected agent %s, got %s", agent, envelope.Agent)
	}

	if envelope.TS == 0 {
		t.Error("Expected non-zero timestamp")
	}

	if envelope.Nonce == "" {
		t.Error("Expected non-empty nonce")
	}

	if envelope.Sig != "" {
		t.Error("Expected empty signature for new envelope")
	}
}

func TestEnvelopeSignAndVerify(t *testing.T) {
	// Generate test key pair
	pubKey, privKey, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatalf("Failed to generate key pair: %v", err)
	}

	// Create test envelope
	envelope := NewEnvelope(EnvelopeRegisterAgent, "test.agent")
	envelope.Body = json.RawMessage(`{"test": "data"}`)

	// Sign the envelope
	err = envelope.Sign(privKey)
	if err != nil {
		t.Fatalf("Failed to sign envelope: %v", err)
	}

	if envelope.Sig == "" {
		t.Error("Expected signature after signing")
	}

	// Verify the signature
	err = envelope.Verify(pubKey)
	if err != nil {
		t.Errorf("Failed to verify signature: %v", err)
	}

	// Test verification with wrong key
	_, wrongPrivKey, _ := ed25519.GenerateKey(nil)
	wrongPubKey := wrongPrivKey.Public().(ed25519.PublicKey)
	
	err = envelope.Verify(wrongPubKey)
	if err == nil {
		t.Error("Expected verification to fail with wrong key")
	}
}

func TestEnvelopeSerialization(t *testing.T) {
	envelope := NewEnvelope(EnvelopeEmitEvent, "test.agent")
	envelope.Body = json.RawMessage(`{"event": "test", "payload": {"key": "value"}}`)

	// Marshal to JSON
	data, err := json.Marshal(envelope)
	if err != nil {
		t.Fatalf("Failed to marshal envelope: %v", err)
	}

	// Unmarshal from JSON
	var unmarshaled Envelope
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal envelope: %v", err)
	}

	// Compare fields
	if unmarshaled.Type != envelope.Type {
		t.Errorf("Type mismatch: expected %s, got %s", envelope.Type, unmarshaled.Type)
	}

	if unmarshaled.Agent != envelope.Agent {
		t.Errorf("Agent mismatch: expected %s, got %s", envelope.Agent, unmarshaled.Agent)
	}

	if unmarshaled.TS != envelope.TS {
		t.Errorf("Timestamp mismatch: expected %d, got %d", envelope.TS, unmarshaled.TS)
	}
}

func TestRegisterAgentEnvelope(t *testing.T) {
	pubKey, _, _ := ed25519.GenerateKey(nil)
	
	body := RegisterAgentBody{
		PubKey:       EncodePublicKey(pubKey),
		Capabilities: []string{"tool.execute", "event.emit"},
		Metadata: map[string]interface{}{
			"version": "1.0.0",
			"name":    "test-agent",
		},
	}

	envelope := &RegisterAgentEnvelope{
		BaseEnvelope: BaseEnvelope{
			Type: EnvelopeRegisterAgent,
			CommonHeaders: CommonHeaders{
				Agent: "test.agent",
				TS:    time.Now().UnixMilli(),
				Nonce: "test-nonce-123",
			},
		},
		Body: body,
	}

	// Test serialization
	data, err := json.Marshal(envelope)
	if err != nil {
		t.Fatalf("Failed to marshal RegisterAgentEnvelope: %v", err)
	}

	// Test deserialization
	var unmarshaled RegisterAgentEnvelope
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal RegisterAgentEnvelope: %v", err)
	}

	if len(unmarshaled.Body.Capabilities) != 2 {
		t.Errorf("Expected 2 capabilities, got %d", len(unmarshaled.Body.Capabilities))
	}
}

func TestToolCallEnvelope(t *testing.T) {
	body := ToolCallBody{
		Tool: "fs.read",
		Parameters: map[string]interface{}{
			"path": "/test/file.txt",
		},
		RequestID: "req-123",
	}

	envelope := &ToolCallEnvelope{
		BaseEnvelope: BaseEnvelope{
			Type: EnvelopeToolCall,
			CommonHeaders: CommonHeaders{
				Agent: "test.agent",
				TS:    time.Now().UnixMilli(),
				Nonce: "test-nonce-456",
			},
		},
		Body: body,
	}

	// Test serialization
	data, err := json.Marshal(envelope)
	if err != nil {
		t.Fatalf("Failed to marshal ToolCallEnvelope: %v", err)
	}

	// Test deserialization
	var unmarshaled ToolCallEnvelope
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal ToolCallEnvelope: %v", err)
	}

	if unmarshaled.Body.Tool != "fs.read" {
		t.Errorf("Expected tool 'fs.read', got '%s'", unmarshaled.Body.Tool)
	}

	if unmarshaled.Body.RequestID != "req-123" {
		t.Errorf("Expected requestId 'req-123', got '%s'", unmarshaled.Body.RequestID)
	}
}

func TestToolResultEnvelope(t *testing.T) {
	body := ToolResultBody{
		RequestID: "req-123",
		Success:   true,
		Result:    "file contents here",
	}

	envelope := &ToolResultEnvelope{
		BaseEnvelope: BaseEnvelope{
			Type: EnvelopeToolResult,
			CommonHeaders: CommonHeaders{
				Agent: "coder.body",
				TS:    time.Now().UnixMilli(),
				Nonce: "test-nonce-789",
			},
		},
		Body: body,
	}

	// Test serialization
	data, err := json.Marshal(envelope)
	if err != nil {
		t.Fatalf("Failed to marshal ToolResultEnvelope: %v", err)
	}

	// Test deserialization
	var unmarshaled ToolResultEnvelope
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal ToolResultEnvelope: %v", err)
	}

	if !unmarshaled.Body.Success {
		t.Error("Expected success to be true")
	}

	if unmarshaled.Body.Result != "file contents here" {
		t.Errorf("Expected result 'file contents here', got '%v'", unmarshaled.Body.Result)
	}
}

func TestRevokeEnvelope(t *testing.T) {
	body := RevokeBody{
		Target: "malicious.agent",
		Reason: "security violation",
	}

	envelope := &RevokeEnvelope{
		BaseEnvelope: BaseEnvelope{
			Type: EnvelopeRevoke,
			CommonHeaders: CommonHeaders{
				Agent: "broker.admin",
				TS:    time.Now().UnixMilli(),
				Nonce: "test-nonce-999",
			},
		},
		Body: body,
	}

	// Test serialization
	data, err := json.Marshal(envelope)
	if err != nil {
		t.Fatalf("Failed to marshal RevokeEnvelope: %v", err)
	}

	// Test deserialization
	var unmarshaled RevokeEnvelope
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal RevokeEnvelope: %v", err)
	}

	if unmarshaled.Body.Target != "malicious.agent" {
		t.Errorf("Expected target 'malicious.agent', got '%s'", unmarshaled.Body.Target)
	}
}

func TestEnvelopeValidation(t *testing.T) {
	// Test empty signature
	envelope := NewEnvelope(EnvelopeRegisterAgent, "test.agent")
	_, wrongPrivKey, _ := ed25519.GenerateKey(nil)
	wrongPubKey := wrongPrivKey.Public().(ed25519.PublicKey)

	err := envelope.Verify(wrongPubKey)
	if err == nil {
		t.Error("Expected verification to fail for envelope without signature")
	}

	// Test invalid signature encoding
	envelope.Sig = "invalid-base64!"
	err = envelope.Verify(wrongPubKey)
	if err == nil {
		t.Error("Expected verification to fail for invalid signature encoding")
	}
}