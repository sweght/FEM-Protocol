package protocol

import (
	"encoding/json"
	"testing"
	"time"
)

func TestDiscoverToolsEnvelope(t *testing.T) {
	pubKey, privKey, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("Failed to generate key pair: %v", err)
	}
	
	envelope := &DiscoverToolsEnvelope{
		BaseEnvelope: BaseEnvelope{
			Type: EnvelopeDiscoverTools,
			CommonHeaders: CommonHeaders{
				Agent: "test-agent",
				TS:    time.Now().UnixMilli(),
				Nonce: "test-nonce",
			},
		},
		Body: DiscoverToolsBody{
			Query: ToolQuery{
				Capabilities:    []string{"file.*", "code.execute"},
				MaxResults:      10,
				IncludeMetadata: true,
			},
			RequestID: "test-request",
		},
	}
	
	// Test signing
	err = envelope.Sign(privKey)
	if err != nil {
		t.Fatalf("Failed to sign envelope: %v", err)
	}
	
	if envelope.Sig == "" {
		t.Fatal("Signature should not be empty")
	}
	
	// Test JSON marshaling
	data, err := json.Marshal(envelope)
	if err != nil {
		t.Fatalf("Failed to marshal envelope: %v", err)
	}
	
	// Test JSON unmarshaling
	var unmarshaled DiscoverToolsEnvelope
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal envelope: %v", err)
	}
	
	if unmarshaled.Body.RequestID != envelope.Body.RequestID {
		t.Errorf("RequestID mismatch: got %s, want %s", 
			unmarshaled.Body.RequestID, envelope.Body.RequestID)
	}
	
	if len(unmarshaled.Body.Query.Capabilities) != 2 {
		t.Errorf("Capabilities length mismatch: got %d, want 2", 
			len(unmarshaled.Body.Query.Capabilities))
	}
	
	t.Logf("✅ DiscoverToolsEnvelope test passed")
	_ = pubKey // avoid unused variable
}

func TestToolsDiscoveredEnvelope(t *testing.T) {
	_, privKey, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("Failed to generate key pair: %v", err)
	}
	
	envelope := &ToolsDiscoveredEnvelope{
		BaseEnvelope: BaseEnvelope{
			Type: EnvelopeToolsDiscovered,
			CommonHeaders: CommonHeaders{
				Agent: "broker-001",
				TS:    time.Now().UnixMilli(),
				Nonce: "discovered-nonce",
			},
		},
		Body: ToolsDiscoveredBody{
			RequestID: "test-request-001",
			Tools: []DiscoveredTool{
				{
					AgentID:         "math-agent-001",
					MCPEndpoint:     "http://localhost:8080",
					Capabilities:    []string{"math.add", "math.multiply"},
					EnvironmentType: "local",
					MCPTools: []MCPTool{
						{
							Name:        "math.add",
							Description: "Add two numbers",
							InputSchema: map[string]interface{}{
								"type": "object",
								"properties": map[string]interface{}{
									"a": map[string]interface{}{"type": "number"},
									"b": map[string]interface{}{"type": "number"},
								},
							},
						},
					},
					Metadata: ToolMetadata{
						LastSeen:            time.Now().UnixMilli(),
						AverageResponseTime: 150,
						TrustScore:          0.95,
					},
				},
			},
			TotalResults: 1,
			HasMore:      false,
		},
	}
	
	// Test signing
	err = envelope.Sign(privKey)
	if err != nil {
		t.Fatalf("Failed to sign envelope: %v", err)
	}
	
	// Test JSON marshaling
	data, err := json.Marshal(envelope)
	if err != nil {
		t.Fatalf("Failed to marshal envelope: %v", err)
	}
	
	// Test JSON unmarshaling
	var unmarshaled ToolsDiscoveredEnvelope
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal envelope: %v", err)
	}
	
	if len(unmarshaled.Body.Tools) != 1 {
		t.Errorf("Tools length mismatch: got %d, want 1", len(unmarshaled.Body.Tools))
	}
	
	tool := unmarshaled.Body.Tools[0]
	if tool.AgentID != "math-agent-001" {
		t.Errorf("AgentID mismatch: got %s, want math-agent-001", tool.AgentID)
	}
	
	t.Logf("✅ ToolsDiscoveredEnvelope test passed")
}

func TestEmbodimentUpdateEnvelope(t *testing.T) {
	_, privKey, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("Failed to generate key pair: %v", err)
	}
	
	envelope := &EmbodimentUpdateEnvelope{
		BaseEnvelope: BaseEnvelope{
			Type: EnvelopeEmbodimentUpdate,
			CommonHeaders: CommonHeaders{
				Agent: "adaptive-agent-001",
				TS:    time.Now().UnixMilli(),
				Nonce: "embodiment-nonce",
			},
		},
		Body: EmbodimentUpdateBody{
			EnvironmentType: "cloud",
			BodyDefinition: BodyDefinition{
				Name:        "cloud-body",
				Environment: "cloud",
				Capabilities: []string{"s3.read", "s3.write"},
				MCPTools: []MCPTool{
					{
						Name:        "s3.read",
						Description: "Read from S3 bucket",
						InputSchema: map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"bucket": map[string]interface{}{"type": "string"},
								"key":    map[string]interface{}{"type": "string"},
							},
						},
					},
				},
			},
			MCPEndpoint:  "http://localhost:8081",
			UpdatedTools: []string{"s3.read"},
		},
	}
	
	// Test signing
	err = envelope.Sign(privKey)
	if err != nil {
		t.Fatalf("Failed to sign envelope: %v", err)
	}
	
	// Test JSON marshaling
	data, err := json.Marshal(envelope)
	if err != nil {
		t.Fatalf("Failed to marshal envelope: %v", err)
	}
	
	// Test JSON unmarshaling
	var unmarshaled EmbodimentUpdateEnvelope
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal envelope: %v", err)
	}
	
	if unmarshaled.Body.EnvironmentType != "cloud" {
		t.Errorf("EnvironmentType mismatch: got %s, want cloud", 
			unmarshaled.Body.EnvironmentType)
	}
	
	if len(unmarshaled.Body.BodyDefinition.MCPTools) != 1 {
		t.Errorf("MCPTools length mismatch: got %d, want 1", 
			len(unmarshaled.Body.BodyDefinition.MCPTools))
	}
	
	t.Logf("✅ EmbodimentUpdateEnvelope test passed")
}

func TestBodyDefinition(t *testing.T) {
	bodyDef := BodyDefinition{
		Name:        "cloud-worker",
		Environment: "cloud",
		Capabilities: []string{"s3.read", "lambda.invoke"},
		MCPTools: []MCPTool{
			{
				Name:        "s3.read",
				Description: "Read from S3 bucket",
				InputSchema: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"bucket": map[string]interface{}{"type": "string"},
						"key":    map[string]interface{}{"type": "string"},
					},
				},
			},
		},
	}
	
	// Test JSON marshaling/unmarshaling
	data, err := json.Marshal(bodyDef)
	if err != nil {
		t.Fatalf("Failed to marshal BodyDefinition: %v", err)
	}
	
	var unmarshaled BodyDefinition
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal BodyDefinition: %v", err)
	}
	
	if unmarshaled.Name != bodyDef.Name {
		t.Errorf("Name mismatch: got %s, want %s", unmarshaled.Name, bodyDef.Name)
	}
	
	if len(unmarshaled.MCPTools) != 1 {
		t.Errorf("MCPTools length mismatch: got %d, want 1", len(unmarshaled.MCPTools))
	}
	
	t.Logf("✅ BodyDefinition test passed")
}