package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/fep-fem/protocol"
)

func TestBrokerMCPIntegration(t *testing.T) {
	// Create test broker
	broker := NewBroker()
	server := httptest.NewTLSServer(broker)
	defer server.Close()

	// Create HTTP client that accepts self-signed certs
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	// Test 1: Register an agent with MCP capabilities
	t.Run("RegisterAgentWithMCP", func(t *testing.T) {
		_, privKey, err := protocol.GenerateKeyPair()
		if err != nil {
			t.Fatalf("Failed to generate key pair: %v", err)
		}

		envelope := &protocol.RegisterAgentEnvelope{
			BaseEnvelope: protocol.BaseEnvelope{
				Type: protocol.EnvelopeRegisterAgent,
				CommonHeaders: protocol.CommonHeaders{
					Agent: "test-agent-001",
					TS:    time.Now().UnixMilli(),
					Nonce: "test-register",
				},
			},
			Body: protocol.RegisterAgentBody{
				PubKey:       "test-public-key",
				Capabilities: []string{"math.add", "math.multiply"},
				MCPEndpoint:  "http://localhost:8080",
				BodyDefinition: &protocol.BodyDefinition{
					Name:        "math-body",
					Environment: "test",
					Capabilities: []string{"math.add", "math.multiply"},
					MCPTools: []protocol.MCPTool{
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
						{
							Name:        "math.multiply",
							Description: "Multiply two numbers",
							InputSchema: map[string]interface{}{
								"type": "object",
								"properties": map[string]interface{}{
									"a": map[string]interface{}{"type": "number"},
									"b": map[string]interface{}{"type": "number"},
								},
							},
						},
					},
				},
				EnvironmentType: "test",
			},
		}

		err = envelope.Sign(privKey)
		if err != nil {
			t.Fatalf("Failed to sign envelope: %v", err)
		}

		data, err := json.Marshal(envelope)
		if err != nil {
			t.Fatalf("Failed to marshal envelope: %v", err)
		}

		resp, err := client.Post(server.URL+"/", "application/json", bytes.NewReader(data))
		if err != nil {
			t.Fatalf("Failed to send request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}

		var response map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&response)
		if err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if response["status"] != "registered" {
			t.Errorf("Expected status 'registered', got %v", response["status"])
		}

		// Verify agent is registered in MCP registry
		agent, exists := broker.mcpRegistry.GetAgent("test-agent-001")
		if !exists {
			t.Fatal("Agent should be registered in MCP registry")
		}

		if len(agent.Tools) != 2 {
			t.Errorf("Expected 2 tools, got %d", len(agent.Tools))
		}
	})

	// Test 2: Discover tools
	t.Run("DiscoverTools", func(t *testing.T) {
		_, privKey, err := protocol.GenerateKeyPair()
		if err != nil {
			t.Fatalf("Failed to generate key pair: %v", err)
		}

		envelope := &protocol.DiscoverToolsEnvelope{
			BaseEnvelope: protocol.BaseEnvelope{
				Type: protocol.EnvelopeDiscoverTools,
				CommonHeaders: protocol.CommonHeaders{
					Agent: "discovery-client",
					TS:    time.Now().UnixMilli(),
					Nonce: "test-discover",
				},
			},
			Body: protocol.DiscoverToolsBody{
				Query: protocol.ToolQuery{
					Capabilities: []string{"math.*"},
					MaxResults:   10,
				},
				RequestID: "discovery-test-001",
			},
		}

		err = envelope.Sign(privKey)
		if err != nil {
			t.Fatalf("Failed to sign envelope: %v", err)
		}

		data, err := json.Marshal(envelope)
		if err != nil {
			t.Fatalf("Failed to marshal envelope: %v", err)
		}

		resp, err := client.Post(server.URL+"/", "application/json", bytes.NewReader(data))
		if err != nil {
			t.Fatalf("Failed to send request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}

		var response map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&response)
		if err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if response["status"] != "success" {
			t.Errorf("Expected status 'success', got %v", response["status"])
		}

		tools, ok := response["tools"].([]interface{})
		if !ok {
			t.Fatal("Response should contain tools array")
		}

		if len(tools) == 0 {
			t.Error("Should find at least one tool matching math.*")
		}

		// Verify tool structure
		if len(tools) > 0 {
			tool := tools[0].(map[string]interface{})
			if tool["agentId"] != "test-agent-001" {
				t.Errorf("Expected agentId 'test-agent-001', got %v", tool["agentId"])
			}

			mcpTools, ok := tool["mcpTools"].([]interface{})
			if !ok || len(mcpTools) == 0 {
				t.Error("Tool should have mcpTools")
			}
		}
	})

	// Test 3: Update agent embodiment
	t.Run("EmbodimentUpdate", func(t *testing.T) {
		_, privKey, err := protocol.GenerateKeyPair()
		if err != nil {
			t.Fatalf("Failed to generate key pair: %v", err)
		}

		envelope := &protocol.EmbodimentUpdateEnvelope{
			BaseEnvelope: protocol.BaseEnvelope{
				Type: protocol.EnvelopeEmbodimentUpdate,
				CommonHeaders: protocol.CommonHeaders{
					Agent: "test-agent-001", // Same agent as registered above
					TS:    time.Now().UnixMilli(),
					Nonce: "test-update",
				},
			},
			Body: protocol.EmbodimentUpdateBody{
				EnvironmentType: "production",
				BodyDefinition: protocol.BodyDefinition{
					Name:        "prod-math-body",
					Environment: "production",
					Capabilities: []string{"math.add", "math.subtract", "math.divide"},
					MCPTools: []protocol.MCPTool{
						{
							Name:        "math.add",
							Description: "Add numbers with precision",
							InputSchema: map[string]interface{}{
								"type": "object",
								"properties": map[string]interface{}{
									"a":         map[string]interface{}{"type": "number"},
									"b":         map[string]interface{}{"type": "number"},
									"precision": map[string]interface{}{"type": "integer"},
								},
							},
						},
						{
							Name:        "math.divide",
							Description: "Divide numbers",
							InputSchema: map[string]interface{}{
								"type": "object",
								"properties": map[string]interface{}{
									"dividend": map[string]interface{}{"type": "number"},
									"divisor":  map[string]interface{}{"type": "number"},
								},
							},
						},
					},
				},
				MCPEndpoint:  "http://localhost:8080",
				UpdatedTools: []string{"math.add", "math.divide"},
			},
		}

		err = envelope.Sign(privKey)
		if err != nil {
			t.Fatalf("Failed to sign envelope: %v", err)
		}

		data, err := json.Marshal(envelope)
		if err != nil {
			t.Fatalf("Failed to marshal envelope: %v", err)
		}

		resp, err := client.Post(server.URL+"/", "application/json", bytes.NewReader(data))
		if err != nil {
			t.Fatalf("Failed to send request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}

		var response map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&response)
		if err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if response["status"] != "updated" {
			t.Errorf("Expected status 'updated', got %v", response["status"])
		}

		// Verify agent was updated in registry
		agent, exists := broker.mcpRegistry.GetAgent("test-agent-001")
		if !exists {
			t.Fatal("Agent should still exist in registry")
		}

		if agent.EnvironmentType != "production" {
			t.Errorf("Expected environment 'production', got %s", agent.EnvironmentType)
		}

		if len(agent.Tools) != 2 {
			t.Errorf("Expected 2 updated tools, got %d", len(agent.Tools))
		}
	})

	// Test 4: Health check still works
	t.Run("HealthCheck", func(t *testing.T) {
		resp, err := client.Get(server.URL + "/health")
		if err != nil {
			t.Fatalf("Failed to send health check: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}
	})
}

func TestBrokerBackwardsCompatibility(t *testing.T) {
	// Test that old-style agent registration still works
	broker := NewBroker()
	server := httptest.NewTLSServer(broker)
	defer server.Close()

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	// Register agent without MCP fields (backwards compatibility)
	t.Run("OldStyleRegistration", func(t *testing.T) {
		_, privKey, err := protocol.GenerateKeyPair()
		if err != nil {
			t.Fatalf("Failed to generate key pair: %v", err)
		}

		envelope := &protocol.RegisterAgentEnvelope{
			BaseEnvelope: protocol.BaseEnvelope{
				Type: protocol.EnvelopeRegisterAgent,
				CommonHeaders: protocol.CommonHeaders{
					Agent: "old-style-agent",
					TS:    time.Now().UnixMilli(),
					Nonce: "old-register",
				},
			},
			Body: protocol.RegisterAgentBody{
				PubKey:       "test-public-key",
				Capabilities: []string{"legacy.tool"},
				// No MCP fields
			},
		}

		err = envelope.Sign(privKey)
		if err != nil {
			t.Fatalf("Failed to sign envelope: %v", err)
		}

		data, err := json.Marshal(envelope)
		if err != nil {
			t.Fatalf("Failed to marshal envelope: %v", err)
		}

		resp, err := client.Post(server.URL+"/", "application/json", bytes.NewReader(data))
		if err != nil {
			t.Fatalf("Failed to send request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}

		var response map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&response)
		if err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if response["status"] != "registered" {
			t.Errorf("Expected status 'registered', got %v", response["status"])
		}

		// Verify agent is NOT in MCP registry (since no MCP endpoint)
		_, exists := broker.mcpRegistry.GetAgent("old-style-agent")
		if exists {
			t.Error("Agent should not be in MCP registry without MCP endpoint")
		}

		// Verify agent is in regular agent registry
		broker.mu.RLock()
		_, exists = broker.agents["old-style-agent"]
		broker.mu.RUnlock()

		if !exists {
			t.Error("Agent should be in regular agent registry")
		}
	})
}

func TestBrokerErrorHandling(t *testing.T) {
	broker := NewBroker()
	server := httptest.NewTLSServer(broker)
	defer server.Close()

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	// Test invalid envelope type
	t.Run("InvalidEnvelopeType", func(t *testing.T) {
		invalidEnvelope := map[string]interface{}{
			"type":  "invalidType",
			"agent": "test-agent",
			"ts":    time.Now().UnixMilli(),
			"nonce": "test",
			"body":  map[string]interface{}{},
		}

		data, _ := json.Marshal(invalidEnvelope)
		resp, err := client.Post(server.URL+"/", "application/json", bytes.NewReader(data))
		if err != nil {
			t.Fatalf("Failed to send request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", resp.StatusCode)
		}
	})

	// Test discovery with invalid body
	t.Run("InvalidDiscoveryBody", func(t *testing.T) {
		invalidEnvelope := map[string]interface{}{
			"type":  "discoverTools",
			"agent": "test-agent",
			"ts":    time.Now().UnixMilli(),
			"nonce": "test",
			"body":  "invalid-body",
		}

		data, _ := json.Marshal(invalidEnvelope)
		resp, err := client.Post(server.URL+"/", "application/json", bytes.NewReader(data))
		if err != nil {
			t.Fatalf("Failed to send request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", resp.StatusCode)
		}
	})
}

func TestFullMCPFederationLoop(t *testing.T) {
	// Start a test broker
	broker := NewBroker()
	server := httptest.NewTLSServer(broker)
	defer server.Close()

	// HTTP client for the agents and tests
	testClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	// Simulate agent registration with MCP endpoint
	agent1ID := "integration-agent-001"
	pubKey1, privKey1, _ := protocol.GenerateKeyPair()
	mcpPort1 := 8090
	
	// Create and send registration envelope for agent 1
	mcpTools := []protocol.MCPTool{
		{Name: "math.add", Description: "Add two numbers"},
		{Name: "code.execute", Description: "Execute commands"},
	}
	
	regBody1 := protocol.RegisterAgentBody{
		PubKey:          protocol.EncodePublicKey(pubKey1),
		Capabilities:    []string{"math.add", "code.execute"},
		MCPEndpoint:     fmt.Sprintf("http://localhost:%d/mcp", mcpPort1),
		BodyDefinition:  &protocol.BodyDefinition{Name: "test-body", MCPTools: mcpTools},
		EnvironmentType: "test",
	}
	
	regEnv1 := &protocol.RegisterAgentEnvelope{
		BaseEnvelope: protocol.BaseEnvelope{
			Type: protocol.EnvelopeRegisterAgent, 
			CommonHeaders: protocol.CommonHeaders{
				Agent: agent1ID, 
				TS: time.Now().UnixMilli(), 
				Nonce: "nonce1",
			},
		},
		Body: regBody1,
	}
	regEnv1.Sign(privKey1)
	regData1, _ := json.Marshal(regEnv1)
	
	resp, err := testClient.Post(server.URL+"/", "application/json", bytes.NewReader(regData1))
	if err != nil {
		t.Fatalf("Agent 1 registration failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Agent 1 registration returned non-200 status: %d", resp.StatusCode)
	}

	// Verify agent is registered in broker
	agent, exists := broker.mcpRegistry.GetAgent(agent1ID)
	if !exists {
		t.Fatal("Agent should be registered in MCP registry")
	}
	
	if len(agent.Tools) != 2 {
		t.Errorf("Expected 2 tools, got %d", len(agent.Tools))
	}

	// Test tool discovery
	_, clientPrivKey, _ := protocol.GenerateKeyPair()
	
	discoverEnv := &protocol.DiscoverToolsEnvelope{
		BaseEnvelope: protocol.BaseEnvelope{
			Type: protocol.EnvelopeDiscoverTools,
			CommonHeaders: protocol.CommonHeaders{
				Agent: "test-mcp-client",
				TS:    time.Now().UnixMilli(),
				Nonce: "discover-nonce",
			},
		},
		Body: protocol.DiscoverToolsBody{
			Query: protocol.ToolQuery{
				Capabilities: []string{"math.add"},
			},
			RequestID: "discovery-req-1",
		},
	}
	discoverEnv.Sign(clientPrivKey)
	discoverData, _ := json.Marshal(discoverEnv)
	
	resp, err = testClient.Post(server.URL+"/", "application/json", bytes.NewReader(discoverData))
	if err != nil {
		t.Fatalf("Tool discovery failed: %v", err)
	}
	
	var discoveryResponse map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&discoveryResponse)
	
	if discoveryResponse["status"] != "success" {
		t.Errorf("Discovery should succeed, got: %v", discoveryResponse["status"])
	}
	
	tools, ok := discoveryResponse["tools"].([]interface{})
	if !ok || len(tools) == 0 {
		t.Fatal("Should discover at least one tool")
	}
	
	discoveredTool := tools[0].(map[string]interface{})
	if discoveredTool["agentId"] != agent1ID {
		t.Errorf("Discovered tool from wrong agent. Expected %s, got %s", agent1ID, discoveredTool["agentId"])
	}
	
	t.Log("Successfully discovered agent's tool via the broker.")
}