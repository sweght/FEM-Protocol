package main

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/fep-fem/protocol"
)

func TestMCPClientCreation(t *testing.T) {
	_, privKey, err := protocol.GenerateKeyPair()
	if err != nil {
		t.Fatalf("Failed to generate key pair: %v", err)
	}

	config := MCPClientConfig{
		AgentID:        "test-client-001",
		BrokerURL:      "https://broker.example.com",
		PrivateKey:     privKey,
		CacheExpiry:    10 * time.Minute,
		RequestTimeout: 30 * time.Second,
		TLSInsecure:    true,
	}

	client := NewMCPClient(config)

	if client.agentID != config.AgentID {
		t.Errorf("AgentID mismatch: got %s, want %s", client.agentID, config.AgentID)
	}

	if client.brokerURL != config.BrokerURL {
		t.Errorf("BrokerURL mismatch: got %s, want %s", client.brokerURL, config.BrokerURL)
	}

	if client.cacheExpiry != config.CacheExpiry {
		t.Errorf("CacheExpiry mismatch: got %v, want %v", client.cacheExpiry, config.CacheExpiry)
	}

	if len(client.toolCache) != 0 {
		t.Errorf("Expected empty tool cache, got %d entries", len(client.toolCache))
	}
}

func TestMCPClientDefaults(t *testing.T) {
	_, privKey, err := protocol.GenerateKeyPair()
	if err != nil {
		t.Fatalf("Failed to generate key pair: %v", err)
	}

	// Create client with minimal config
	config := MCPClientConfig{
		AgentID:     "test-client",
		BrokerURL:   "https://broker.example.com",
		PrivateKey:  privKey,
		TLSInsecure: true,
	}

	client := NewMCPClient(config)

	// Check defaults
	if client.cacheExpiry != 5*time.Minute {
		t.Errorf("Expected default cache expiry 5m, got %v", client.cacheExpiry)
	}

	if client.httpClient.Timeout != 30*time.Second {
		t.Errorf("Expected default timeout 30s, got %v", client.httpClient.Timeout)
	}
}

func TestMCPClientCacheKey(t *testing.T) {
	_, privKey, err := protocol.GenerateKeyPair()
	if err != nil {
		t.Fatalf("Failed to generate key pair: %v", err)
	}

	client := NewMCPClient(MCPClientConfig{
		AgentID:     "test",
		BrokerURL:   "https://example.com",
		PrivateKey:  privKey,
		TLSInsecure: true,
	})

	tests := []struct {
		name     string
		query    protocol.ToolQuery
		expected string
	}{
		{
			name: "Basic query",
			query: protocol.ToolQuery{
				Capabilities: []string{"math.*"},
				MaxResults:   10,
			},
			expected: "env:,caps:[math.*],max:10",
		},
		{
			name: "Query with environment",
			query: protocol.ToolQuery{
				Capabilities:    []string{"file.*", "code.*"},
				EnvironmentType: "production",
				MaxResults:      20,
			},
			expected: "env:production,caps:[file.* code.*],max:20",
		},
		{
			name: "Empty query",
			query: protocol.ToolQuery{
				Capabilities: []string{},
			},
			expected: "env:,caps:[],max:0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := client.buildCacheKey(tt.query)
			if key != tt.expected {
				t.Errorf("Cache key mismatch: got %s, want %s", key, tt.expected)
			}
		})
	}
}

func TestMCPClientCaching(t *testing.T) {
	_, privKey, err := protocol.GenerateKeyPair()
	if err != nil {
		t.Fatalf("Failed to generate key pair: %v", err)
	}

	client := NewMCPClient(MCPClientConfig{
		AgentID:     "cache-test",
		BrokerURL:   "https://example.com",
		PrivateKey:  privKey,
		CacheExpiry: 100 * time.Millisecond, // Short expiry for testing
		TLSInsecure: true,
	})

	// Test data
	tools := []protocol.DiscoveredTool{
		{
			AgentID:      "test-agent",
			MCPEndpoint:  "http://localhost:8080",
			Capabilities: []string{"test.tool"},
			MCPTools: []protocol.MCPTool{
				{
					Name:        "test.tool",
					Description: "Test tool",
				},
			},
		},
	}

	// Cache some tools
	cacheKey := "test-key"
	client.cacheResult(cacheKey, tools)

	// Verify cache hit
	cached := client.getCachedResult(cacheKey)
	if cached == nil {
		t.Fatal("Expected cache hit, got nil")
	}

	if len(cached.Tools) != 1 {
		t.Errorf("Expected 1 cached tool, got %d", len(cached.Tools))
	}

	if cached.Tools[0].AgentID != "test-agent" {
		t.Errorf("Cached tool AgentID mismatch: got %s, want test-agent", cached.Tools[0].AgentID)
	}

	// Wait for cache expiry
	time.Sleep(150 * time.Millisecond)

	// Verify cache miss after expiry
	expired := client.getCachedResult(cacheKey)
	if expired != nil {
		t.Error("Expected cache miss after expiry, got hit")
	}
}

func TestMCPClientRequestIDGeneration(t *testing.T) {
	_, privKey, err := protocol.GenerateKeyPair()
	if err != nil {
		t.Fatalf("Failed to generate key pair: %v", err)
	}

	client := NewMCPClient(MCPClientConfig{
		AgentID:     "req-test",
		BrokerURL:   "https://example.com",
		PrivateKey:  privKey,
		TLSInsecure: true,
	})

	// Generate multiple request IDs
	ids := make(map[string]bool)
	for i := 0; i < 10; i++ {
		id := client.generateRequestID()
		
		// Check uniqueness
		if ids[id] {
			t.Errorf("Duplicate request ID generated: %s", id)
		}
		ids[id] = true

		// Check format
		expectedPrefix := "req-test-req-"
		if len(id) < len(expectedPrefix) {
			t.Errorf("Request ID too short: %s", id)
		}
	}
}

func TestMCPClientDiscoverToolsIntegration(t *testing.T) {
	// Create test broker
	broker := NewBroker()
	server := httptest.NewTLSServer(broker)
	defer server.Close()

	// Register a test MCP agent in the broker
	testAgent := &MCPAgent{
		ID:              "math-agent",
		MCPEndpoint:     "http://localhost:8080",
		EnvironmentType: "test",
		Tools: []protocol.MCPTool{
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
		LastHeartbeat: time.Now(),
	}
	broker.mcpRegistry.RegisterAgent(testAgent.ID, testAgent)

	// Create MCP client
	_, privKey, err := protocol.GenerateKeyPair()
	if err != nil {
		t.Fatalf("Failed to generate key pair: %v", err)
	}

	client := NewMCPClient(MCPClientConfig{
		AgentID:     "client-test",
		BrokerURL:   server.URL,
		PrivateKey:  privKey,
		TLSInsecure: true,
	})

	// Test tool discovery
	t.Run("DiscoverMathTools", func(t *testing.T) {
		tools, err := client.FindToolsByCapability([]string{"math.*"})
		if err != nil {
			t.Fatalf("Discovery failed: %v", err)
		}

		if len(tools) != 1 {
			t.Errorf("Expected 1 agent with tools, got %d", len(tools))
		}

		agent := tools[0]
		if agent.AgentID != "math-agent" {
			t.Errorf("AgentID mismatch: got %s, want math-agent", agent.AgentID)
		}

		if len(agent.MCPTools) != 1 {
			t.Errorf("Expected 1 tool, got %d", len(agent.MCPTools))
		}

		tool := agent.MCPTools[0]
		if tool.Name != "math.add" {
			t.Errorf("Tool name mismatch: got %s, want math.add", tool.Name)
		}
	})

	t.Run("DiscoverAllTools", func(t *testing.T) {
		agents, err := client.GetAvailableAgents()
		if err != nil {
			t.Fatalf("Failed to get available agents: %v", err)
		}

		if len(agents) == 0 {
			t.Error("Expected at least one agent, got none")
		}
	})

	t.Run("DiscoverToolsInEnvironment", func(t *testing.T) {
		tools, err := client.FindToolsInEnvironment("test", 10)
		if err != nil {
			t.Fatalf("Environment discovery failed: %v", err)
		}

		if len(tools) != 1 {
			t.Errorf("Expected 1 agent in test environment, got %d", len(tools))
		}
	})

	t.Run("CacheWorking", func(t *testing.T) {
		// First call - should hit broker
		tools1, err := client.FindToolsByCapability([]string{"math.*"})
		if err != nil {
			t.Fatalf("First discovery failed: %v", err)
		}

		// Second call - should hit cache
		tools2, err := client.FindToolsByCapability([]string{"math.*"})
		if err != nil {
			t.Fatalf("Second discovery failed: %v", err)
		}

		// Results should be identical
		if len(tools1) != len(tools2) {
			t.Errorf("Cache results differ: %d vs %d tools", len(tools1), len(tools2))
		}

		// Check cache stats
		stats := client.GetCacheStats()
		if cached, ok := stats["cached_queries"].(int); !ok || cached == 0 {
			t.Error("Expected cache to have entries")
		}
	})
}

func TestMCPClientCacheRefresh(t *testing.T) {
	_, privKey, err := protocol.GenerateKeyPair()
	if err != nil {
		t.Fatalf("Failed to generate key pair: %v", err)
	}

	client := NewMCPClient(MCPClientConfig{
		AgentID:     "refresh-test",
		BrokerURL:   "https://example.com",
		PrivateKey:  privKey,
		TLSInsecure: true,
	})

	// Add some cache entries
	tools := []protocol.DiscoveredTool{
		{AgentID: "test-agent", MCPEndpoint: "http://test"},
	}
	
	client.cacheResult("key1", tools)
	client.cacheResult("key2", tools)

	// Verify cache has entries
	stats := client.GetCacheStats()
	if cached := stats["cached_queries"].(int); cached != 2 {
		t.Errorf("Expected 2 cached queries, got %d", cached)
	}

	// Refresh cache
	client.RefreshCache()

	// Verify cache is empty
	statsAfter := client.GetCacheStats()
	if cached := statsAfter["cached_queries"].(int); cached != 0 {
		t.Errorf("Expected 0 cached queries after refresh, got %d", cached)
	}
}

func TestMCPClientToolCallFormat(t *testing.T) {
	// Create test broker that logs tool calls
	broker := NewBroker()
	server := httptest.NewTLSServer(broker)
	defer server.Close()

	_, privKey, err := protocol.GenerateKeyPair()
	if err != nil {
		t.Fatalf("Failed to generate key pair: %v", err)
	}

	client := NewMCPClient(MCPClientConfig{
		AgentID:     "tool-call-test",
		BrokerURL:   server.URL,
		PrivateKey:  privKey,
		TLSInsecure: true,
	})

	// Test tool call (will fail but we're testing the format)
	parameters := map[string]interface{}{
		"a": 5,
		"b": 3,
	}

	result, err := client.CallTool("math-agent", "add", parameters)
	
	// We expect this to return a "processing" status from our broker
	if err != nil {
		t.Fatalf("Tool call failed: %v", err)
	}

	// Check result format
	if result == nil {
		t.Fatal("Expected result, got nil")
	}

	resultMap, ok := result.(map[string]interface{})
	if !ok {
		t.Fatal("Expected result to be a map")
	}

	if status, ok := resultMap["status"].(string); !ok || status != "processing" {
		t.Errorf("Expected status 'processing', got %v", resultMap["status"])
	}
}