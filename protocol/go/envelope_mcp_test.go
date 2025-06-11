package protocol

import (
	"encoding/json"
	"fmt"
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

// Phase B: Comprehensive Protocol Testing

func TestDiscoverToolsEnvelopeEdgeCases(t *testing.T) {
	_, privKey, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("Failed to generate key pair: %v", err)
	}

	tests := []struct {
		name     string
		envelope *DiscoverToolsEnvelope
		wantErr  bool
	}{
		{
			name: "Empty capabilities",
			envelope: &DiscoverToolsEnvelope{
				BaseEnvelope: BaseEnvelope{
					Type: EnvelopeDiscoverTools,
					CommonHeaders: CommonHeaders{
						Agent: "test-agent",
						TS:    time.Now().UnixMilli(),
						Nonce: "test-nonce",
					},
				},
				Body: DiscoverToolsBody{
					Query:     ToolQuery{Capabilities: []string{}},
					RequestID: "empty-caps",
				},
			},
			wantErr: false,
		},
		{
			name: "Wildcard capabilities",
			envelope: &DiscoverToolsEnvelope{
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
						Capabilities:    []string{"*", "file.*", "code.*"},
						EnvironmentType: "production",
						MaxResults:      100,
					},
					RequestID: "wildcard-test",
				},
			},
			wantErr: false,
		},
		{
			name: "Max results zero",
			envelope: &DiscoverToolsEnvelope{
				BaseEnvelope: BaseEnvelope{
					Type: EnvelopeDiscoverTools,
					CommonHeaders: CommonHeaders{
						Agent: "test-agent",
						TS:    time.Now().UnixMilli(),
						Nonce: "test-nonce",
					},
				},
				Body: DiscoverToolsBody{
					Query:     ToolQuery{MaxResults: 0},
					RequestID: "zero-max",
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test signing
			err := tt.envelope.Sign(privKey)
			if (err != nil) != tt.wantErr {
				t.Errorf("Sign() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Test JSON roundtrip
				data, err := json.Marshal(tt.envelope)
				if err != nil {
					t.Fatalf("Failed to marshal: %v", err)
				}

				var unmarshaled DiscoverToolsEnvelope
				err = json.Unmarshal(data, &unmarshaled)
				if err != nil {
					t.Fatalf("Failed to unmarshal: %v", err)
				}

				if unmarshaled.Body.RequestID != tt.envelope.Body.RequestID {
					t.Errorf("RequestID mismatch: got %s, want %s",
						unmarshaled.Body.RequestID, tt.envelope.Body.RequestID)
				}
			}
		})
	}
}

func TestToolsDiscoveredEnvelopeComprehensive(t *testing.T) {
	_, privKey, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("Failed to generate key pair: %v", err)
	}

	tests := []struct {
		name     string
		envelope *ToolsDiscoveredEnvelope
		wantErr  bool
	}{
		{
			name: "Empty tools list",
			envelope: &ToolsDiscoveredEnvelope{
				BaseEnvelope: BaseEnvelope{
					Type: EnvelopeToolsDiscovered,
					CommonHeaders: CommonHeaders{
						Agent: "broker-001",
						TS:    time.Now().UnixMilli(),
						Nonce: "empty-tools",
					},
				},
				Body: ToolsDiscoveredBody{
					RequestID:    "req-001",
					Tools:        []DiscoveredTool{},
					TotalResults: 0,
					HasMore:      false,
				},
			},
			wantErr: false,
		},
		{
			name: "Multiple agents with complex tools",
			envelope: &ToolsDiscoveredEnvelope{
				BaseEnvelope: BaseEnvelope{
					Type: EnvelopeToolsDiscovered,
					CommonHeaders: CommonHeaders{
						Agent: "broker-001",
						TS:    time.Now().UnixMilli(),
						Nonce: "multi-tools",
					},
				},
				Body: ToolsDiscoveredBody{
					RequestID: "req-multi",
					Tools: []DiscoveredTool{
						{
							AgentID:         "data-agent-001",
							MCPEndpoint:     "https://data.example.com:8080",
							Capabilities:    []string{"data.read", "data.write", "data.transform"},
							EnvironmentType: "cloud",
							MCPTools: []MCPTool{
								{
									Name:        "data.read",
									Description: "Read data from various sources",
									InputSchema: map[string]interface{}{
										"type": "object",
										"properties": map[string]interface{}{
											"source": map[string]interface{}{
												"type": "string",
												"enum": []string{"file", "db", "api"},
											},
											"path": map[string]interface{}{"type": "string"},
										},
										"required": []string{"source", "path"},
									},
								},
								{
									Name:        "data.transform",
									Description: "Transform data using various operations",
									InputSchema: map[string]interface{}{
										"type": "object",
										"properties": map[string]interface{}{
											"operation": map[string]interface{}{
												"type": "string",
												"enum": []string{"filter", "map", "reduce"},
											},
											"expression": map[string]interface{}{"type": "string"},
										},
									},
								},
							},
							Metadata: ToolMetadata{
								LastSeen:            time.Now().UnixMilli(),
								AverageResponseTime: 250,
								TrustScore:          0.92,
							},
						},
						{
							AgentID:         "ml-agent-002",
							MCPEndpoint:     "http://ml.local:8081",
							Capabilities:    []string{"ml.train", "ml.predict"},
							EnvironmentType: "local",
							MCPTools: []MCPTool{
								{
									Name:        "ml.predict",
									Description: "Run model inference",
									InputSchema: map[string]interface{}{
										"type": "object",
										"properties": map[string]interface{}{
											"model": map[string]interface{}{"type": "string"},
											"input": map[string]interface{}{"type": "array"},
										},
									},
								},
							},
							Metadata: ToolMetadata{
								LastSeen:            time.Now().UnixMilli() - 5000,
								AverageResponseTime: 500,
								TrustScore:          0.88,
							},
						},
					},
					TotalResults: 2,
					HasMore:      false,
				},
			},
			wantErr: false,
		},
		{
			name: "Large result set with pagination",
			envelope: &ToolsDiscoveredEnvelope{
				BaseEnvelope: BaseEnvelope{
					Type: EnvelopeToolsDiscovered,
					CommonHeaders: CommonHeaders{
						Agent: "broker-001",
						TS:    time.Now().UnixMilli(),
						Nonce: "large-set",
					},
				},
				Body: ToolsDiscoveredBody{
					RequestID:    "req-large",
					Tools:        generateLargeToolSet(50), // Helper function
					TotalResults: 500,
					HasMore:      true,
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test signing
			err := tt.envelope.Sign(privKey)
			if (err != nil) != tt.wantErr {
				t.Errorf("Sign() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Test JSON roundtrip
				data, err := json.Marshal(tt.envelope)
				if err != nil {
					t.Fatalf("Failed to marshal: %v", err)
				}

				var unmarshaled ToolsDiscoveredEnvelope
				err = json.Unmarshal(data, &unmarshaled)
				if err != nil {
					t.Fatalf("Failed to unmarshal: %v", err)
				}

				if len(unmarshaled.Body.Tools) != len(tt.envelope.Body.Tools) {
					t.Errorf("Tools length mismatch: got %d, want %d",
						len(unmarshaled.Body.Tools), len(tt.envelope.Body.Tools))
				}

				if unmarshaled.Body.TotalResults != tt.envelope.Body.TotalResults {
					t.Errorf("TotalResults mismatch: got %d, want %d",
						unmarshaled.Body.TotalResults, tt.envelope.Body.TotalResults)
				}
			}
		})
	}
}

func TestEmbodimentUpdateEnvelopeEdgeCases(t *testing.T) {
	_, privKey, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("Failed to generate key pair: %v", err)
	}

	tests := []struct {
		name     string
		envelope *EmbodimentUpdateEnvelope
		wantErr  bool
	}{
		{
			name: "Environment change with constraints",
			envelope: &EmbodimentUpdateEnvelope{
				BaseEnvelope: BaseEnvelope{
					Type: EnvelopeEmbodimentUpdate,
					CommonHeaders: CommonHeaders{
						Agent: "adaptive-agent",
						TS:    time.Now().UnixMilli(),
						Nonce: "constrained",
					},
				},
				Body: EmbodimentUpdateBody{
					EnvironmentType: "secure-cloud",
					BodyDefinition: BodyDefinition{
						Name:        "secure-body",
						Environment: "secure-cloud",
						Capabilities: []string{"secure.read", "secure.process"},
						MCPTools: []MCPTool{
							{
								Name:        "secure.read",
								Description: "Read from secure storage",
								InputSchema: map[string]interface{}{
									"type": "object",
									"properties": map[string]interface{}{
										"vault": map[string]interface{}{"type": "string"},
										"key":   map[string]interface{}{"type": "string"},
									},
								},
							},
						},
						Constraints: map[string]interface{}{
							"encryption":    "required",
							"audit_logging": true,
							"max_data_size": 10485760, // 10MB
						},
						Metadata: map[string]interface{}{
							"compliance": "SOC2",
							"region":     "us-east-1",
						},
					},
					MCPEndpoint:  "https://secure.agent.com:8443",
					UpdatedTools: []string{"secure.read", "secure.process"},
				},
			},
			wantErr: false,
		},
		{
			name: "Empty body definition",
			envelope: &EmbodimentUpdateEnvelope{
				BaseEnvelope: BaseEnvelope{
					Type: EnvelopeEmbodimentUpdate,
					CommonHeaders: CommonHeaders{
						Agent: "minimal-agent",
						TS:    time.Now().UnixMilli(),
						Nonce: "minimal",
					},
				},
				Body: EmbodimentUpdateBody{
					EnvironmentType: "minimal",
					BodyDefinition: BodyDefinition{
						Name:         "minimal-body",
						Environment:  "minimal",
						Capabilities: []string{},
						MCPTools:     []MCPTool{},
					},
					MCPEndpoint:  "http://localhost:8080",
					UpdatedTools: []string{},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test signing
			err := tt.envelope.Sign(privKey)
			if (err != nil) != tt.wantErr {
				t.Errorf("Sign() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Test JSON roundtrip
				data, err := json.Marshal(tt.envelope)
				if err != nil {
					t.Fatalf("Failed to marshal: %v", err)
				}

				var unmarshaled EmbodimentUpdateEnvelope
				err = json.Unmarshal(data, &unmarshaled)
				if err != nil {
					t.Fatalf("Failed to unmarshal: %v", err)
				}

				if unmarshaled.Body.EnvironmentType != tt.envelope.Body.EnvironmentType {
					t.Errorf("EnvironmentType mismatch: got %s, want %s",
						unmarshaled.Body.EnvironmentType, tt.envelope.Body.EnvironmentType)
				}
			}
		})
	}
}

func TestMCPEnvelopeSignatureVerification(t *testing.T) {
	pubKey, privKey, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("Failed to generate key pair: %v", err)
	}

	// Test DiscoverTools signature verification
	discoverEnv := &DiscoverToolsEnvelope{
		BaseEnvelope: BaseEnvelope{
			Type: EnvelopeDiscoverTools,
			CommonHeaders: CommonHeaders{
				Agent: "test-agent",
				TS:    time.Now().UnixMilli(),
				Nonce: "sig-test",
			},
		},
		Body: DiscoverToolsBody{
			Query:     ToolQuery{Capabilities: []string{"test.*"}},
			RequestID: "sig-test-req",
		},
	}

	err = discoverEnv.Sign(privKey)
	if err != nil {
		t.Fatalf("Failed to sign DiscoverTools envelope: %v", err)
	}

	// Test verification through generic envelope
	data, err := json.Marshal(discoverEnv)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	var genericEnv Envelope
	err = json.Unmarshal(data, &genericEnv)
	if err != nil {
		t.Fatalf("Failed to unmarshal to generic envelope: %v", err)
	}

	err = genericEnv.Verify(pubKey)
	if err != nil {
		t.Errorf("Signature verification failed: %v", err)
	}

	// Test with wrong key should fail
	wrongPubKey, _, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("Failed to generate wrong key: %v", err)
	}

	err = genericEnv.Verify(wrongPubKey)
	if err == nil {
		t.Error("Expected signature verification to fail with wrong key")
	}
}

func TestMCPEnvelopeInvalidJSON(t *testing.T) {
	invalidJSONTests := []struct {
		name string
		json string
	}{
		{
			name: "Invalid DiscoverTools JSON",
			json: `{"type":"discoverTools","agent":"test","ts":123,"nonce":"test","body":{"invalidField":true}}`,
		},
		{
			name: "Missing required fields",
			json: `{"type":"discoverTools"}`,
		},
		{
			name: "Invalid enum values",
			json: `{"type":"invalidType","agent":"test","ts":123,"nonce":"test","body":{}}`,
		},
	}

	for _, tt := range invalidJSONTests {
		t.Run(tt.name, func(t *testing.T) {
			var env DiscoverToolsEnvelope
			err := json.Unmarshal([]byte(tt.json), &env)
			// We expect these to either fail or result in zero values
			// The important thing is they don't panic
			t.Logf("Unmarshal result: err=%v, env=%+v", err, env)
		})
	}
}

func TestRegisterAgentWithMCPFields(t *testing.T) {
	_, privKey, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("Failed to generate key pair: %v", err)
	}

	envelope := &RegisterAgentEnvelope{
		BaseEnvelope: BaseEnvelope{
			Type: EnvelopeRegisterAgent,
			CommonHeaders: CommonHeaders{
				Agent: "mcp-agent-001",
				TS:    time.Now().UnixMilli(),
				Nonce: "mcp-reg",
			},
		},
		Body: RegisterAgentBody{
			PubKey:       "test-public-key",
			Capabilities: []string{"code.execute", "file.read"},
			MCPEndpoint:  "https://agent.example.com:8080/mcp",
			BodyDefinition: &BodyDefinition{
				Name:        "development-body",
				Environment: "development",
				Capabilities: []string{"code.execute", "file.read"},
				MCPTools: []MCPTool{
					{
						Name:        "code.execute",
						Description: "Execute code in sandbox",
						InputSchema: map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"code":     map[string]interface{}{"type": "string"},
								"language": map[string]interface{}{"type": "string"},
							},
						},
					},
				},
			},
			EnvironmentType: "development",
		},
	}

	// Test signing
	err = envelope.Sign(privKey)
	if err != nil {
		t.Fatalf("Failed to sign envelope: %v", err)
	}

	// Test JSON roundtrip
	data, err := json.Marshal(envelope)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	var unmarshaled RegisterAgentEnvelope
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	// Verify MCP fields are preserved
	if unmarshaled.Body.MCPEndpoint != envelope.Body.MCPEndpoint {
		t.Errorf("MCPEndpoint mismatch: got %s, want %s",
			unmarshaled.Body.MCPEndpoint, envelope.Body.MCPEndpoint)
	}

	if unmarshaled.Body.EnvironmentType != envelope.Body.EnvironmentType {
		t.Errorf("EnvironmentType mismatch: got %s, want %s",
			unmarshaled.Body.EnvironmentType, envelope.Body.EnvironmentType)
	}

	if unmarshaled.Body.BodyDefinition == nil {
		t.Fatal("BodyDefinition should not be nil")
	}

	if len(unmarshaled.Body.BodyDefinition.MCPTools) != 1 {
		t.Errorf("MCPTools length mismatch: got %d, want 1",
			len(unmarshaled.Body.BodyDefinition.MCPTools))
	}
}

// Helper function for generating large tool sets for testing
func generateLargeToolSet(count int) []DiscoveredTool {
	tools := make([]DiscoveredTool, count)
	for i := 0; i < count; i++ {
		tools[i] = DiscoveredTool{
			AgentID:         fmt.Sprintf("agent-%03d", i),
			MCPEndpoint:     fmt.Sprintf("http://agent-%03d.example.com:8080", i),
			Capabilities:    []string{fmt.Sprintf("tool.%d", i)},
			EnvironmentType: "test",
			MCPTools: []MCPTool{
				{
					Name:        fmt.Sprintf("tool.%d", i),
					Description: fmt.Sprintf("Test tool %d", i),
					InputSchema: map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"input": map[string]interface{}{"type": "string"},
						},
					},
				},
			},
			Metadata: ToolMetadata{
				LastSeen:            time.Now().UnixMilli(),
				AverageResponseTime: 100 + i,
				TrustScore:          0.5 + float64(i)/float64(count*2),
			},
		}
	}
	return tools
}

func TestToolQueryValidation(t *testing.T) {
	tests := []struct {
		name  string
		query ToolQuery
		valid bool
	}{
		{
			name: "Valid query with all fields",
			query: ToolQuery{
				Capabilities:    []string{"file.*", "code.execute"},
				EnvironmentType: "production",
				MaxResults:      10,
				IncludeMetadata: true,
			},
			valid: true,
		},
		{
			name: "Valid query with minimal fields",
			query: ToolQuery{
				Capabilities: []string{"*"},
			},
			valid: true,
		},
		{
			name: "Valid query with empty capabilities",
			query: ToolQuery{
				Capabilities: []string{},
				MaxResults:   100,
			},
			valid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test JSON marshaling
			data, err := json.Marshal(tt.query)
			if err != nil && tt.valid {
				t.Errorf("Expected valid query to marshal, got error: %v", err)
			}

			if tt.valid {
				var unmarshaled ToolQuery
				err = json.Unmarshal(data, &unmarshaled)
				if err != nil {
					t.Errorf("Failed to unmarshal valid query: %v", err)
				}

				if len(unmarshaled.Capabilities) != len(tt.query.Capabilities) {
					t.Errorf("Capabilities length mismatch: got %d, want %d",
						len(unmarshaled.Capabilities), len(tt.query.Capabilities))
				}
			}
		})
	}
}