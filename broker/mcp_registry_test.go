package main

import (
	"testing"
	"time"

	"github.com/fep-fem/protocol"
)

func TestMCPRegistryBasics(t *testing.T) {
	registry := NewMCPRegistry()

	// Test initial state
	if registry.GetAgentCount() != 0 {
		t.Errorf("Expected 0 agents, got %d", registry.GetAgentCount())
	}

	if registry.GetToolCount() != 0 {
		t.Errorf("Expected 0 tools, got %d", registry.GetToolCount())
	}

	// Create test agent
	agent := &MCPAgent{
		ID:              "test-agent-001",
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
		LastHeartbeat: time.Now(),
	}

	// Test agent registration
	err := registry.RegisterAgent(agent.ID, agent)
	if err != nil {
		t.Fatalf("Failed to register agent: %v", err)
	}

	// Verify counts
	if registry.GetAgentCount() != 1 {
		t.Errorf("Expected 1 agent, got %d", registry.GetAgentCount())
	}

	if registry.GetToolCount() != 2 {
		t.Errorf("Expected 2 tools, got %d", registry.GetToolCount())
	}

	// Test agent retrieval
	retrievedAgent, exists := registry.GetAgent(agent.ID)
	if !exists {
		t.Fatal("Agent should exist")
	}

	if retrievedAgent.ID != agent.ID {
		t.Errorf("Agent ID mismatch: got %s, want %s", retrievedAgent.ID, agent.ID)
	}

	if len(retrievedAgent.Tools) != 2 {
		t.Errorf("Expected 2 tools, got %d", len(retrievedAgent.Tools))
	}
}

func TestMCPRegistryDiscovery(t *testing.T) {
	registry := NewMCPRegistry()

	// Register test agents with different tool types
	mathAgent := &MCPAgent{
		ID:              "math-agent",
		MCPEndpoint:     "http://localhost:8080",
		EnvironmentType: "local",
		Tools: []protocol.MCPTool{
			{
				Name:        "math.add",
				Description: "Add numbers",
			},
			{
				Name:        "math.subtract",
				Description: "Subtract numbers",
			},
		},
		LastHeartbeat: time.Now(),
	}

	fileAgent := &MCPAgent{
		ID:              "file-agent",
		MCPEndpoint:     "http://localhost:8081",
		EnvironmentType: "local",
		Tools: []protocol.MCPTool{
			{
				Name:        "file.read",
				Description: "Read files",
			},
			{
				Name:        "file.write",
				Description: "Write files",
			},
		},
		LastHeartbeat: time.Now(),
	}

	registry.RegisterAgent(mathAgent.ID, mathAgent)
	registry.RegisterAgent(fileAgent.ID, fileAgent)

	tests := []struct {
		name           string
		query          protocol.ToolQuery
		expectedAgents int
		expectedTools  int
	}{
		{
			name: "Find all tools",
			query: protocol.ToolQuery{
				Capabilities: []string{"*"},
			},
			expectedAgents: 2,
			expectedTools:  4,
		},
		{
			name: "Find math tools",
			query: protocol.ToolQuery{
				Capabilities: []string{"math.*"},
			},
			expectedAgents: 1,
			expectedTools:  2,
		},
		{
			name: "Find file tools",
			query: protocol.ToolQuery{
				Capabilities: []string{"file.*"},
			},
			expectedAgents: 1,
			expectedTools:  2,
		},
		{
			name: "Find specific tool",
			query: protocol.ToolQuery{
				Capabilities: []string{"math.add"},
			},
			expectedAgents: 1,
			expectedTools:  1,
		},
		{
			name: "Find multiple patterns",
			query: protocol.ToolQuery{
				Capabilities: []string{"math.add", "file.read"},
			},
			expectedAgents: 2,
			expectedTools:  2,
		},
		{
			name: "Find nothing",
			query: protocol.ToolQuery{
				Capabilities: []string{"nonexistent.*"},
			},
			expectedAgents: 0,
			expectedTools:  0,
		},
		{
			name: "With max results",
			query: protocol.ToolQuery{
				Capabilities: []string{"*"},
				MaxResults:   1,
			},
			expectedAgents: 1,
			expectedTools:  -1, // Variable based on which agent comes first
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			discovered, err := registry.DiscoverTools(tt.query)
			if err != nil {
				t.Fatalf("Discovery failed: %v", err)
			}

			if len(discovered) != tt.expectedAgents {
				t.Errorf("Expected %d agents, got %d", tt.expectedAgents, len(discovered))
			}

			if tt.expectedTools >= 0 {
				totalTools := 0
				for _, agent := range discovered {
					totalTools += len(agent.MCPTools)
				}

				if totalTools != tt.expectedTools {
					t.Errorf("Expected %d tools, got %d", tt.expectedTools, totalTools)
				}
			}

			// Verify structure
			for _, agent := range discovered {
				if agent.AgentID == "" {
					t.Error("AgentID should not be empty")
				}
				if agent.MCPEndpoint == "" {
					t.Error("MCPEndpoint should not be empty")
				}
				if len(agent.MCPTools) == 0 {
					t.Error("MCPTools should not be empty")
				}
			}
		})
	}
}

func TestMCPRegistryPatternMatching(t *testing.T) {
	registry := NewMCPRegistry()

	tests := []struct {
		toolName string
		pattern  string
		expected bool
	}{
		{"math.add", "*", true},
		{"math.add", "math.*", true},
		{"math.add", "math.add", true},
		{"math.add", "file.*", false},
		{"file.read", "file.*", true},
		{"complex.namespace.tool", "complex.*", true},
		{"complex.namespace.tool", "complex.namespace.*", true},
		{"complex.namespace.tool", "complex.namespace.tool", true},
		{"complex.namespace.tool", "complex.other.*", false},
	}

	for _, tt := range tests {
		t.Run(tt.toolName+"_"+tt.pattern, func(t *testing.T) {
			result := registry.matchCapability(tt.toolName, tt.pattern)
			if result != tt.expected {
				t.Errorf("matchCapability(%q, %q) = %v, want %v",
					tt.toolName, tt.pattern, result, tt.expected)
			}
		})
	}
}

func TestMCPRegistryUnregister(t *testing.T) {
	registry := NewMCPRegistry()

	// Register agent
	agent := &MCPAgent{
		ID:              "temp-agent",
		MCPEndpoint:     "http://localhost:8080",
		EnvironmentType: "test",
		Tools: []protocol.MCPTool{
			{
				Name:        "temp.tool",
				Description: "Temporary tool",
			},
		},
		LastHeartbeat: time.Now(),
	}

	registry.RegisterAgent(agent.ID, agent)

	// Verify registration
	if registry.GetAgentCount() != 1 {
		t.Errorf("Expected 1 agent, got %d", registry.GetAgentCount())
	}
	if registry.GetToolCount() != 1 {
		t.Errorf("Expected 1 tool, got %d", registry.GetToolCount())
	}

	// Unregister
	registry.UnregisterAgent(agent.ID)

	// Verify cleanup
	if registry.GetAgentCount() != 0 {
		t.Errorf("Expected 0 agents, got %d", registry.GetAgentCount())
	}
	if registry.GetToolCount() != 0 {
		t.Errorf("Expected 0 tools, got %d", registry.GetToolCount())
	}

	// Verify agent is gone
	_, exists := registry.GetAgent(agent.ID)
	if exists {
		t.Error("Agent should not exist after unregistration")
	}
}

func TestMCPRegistryHeartbeat(t *testing.T) {
	registry := NewMCPRegistry()

	agent := &MCPAgent{
		ID:              "heartbeat-agent",
		MCPEndpoint:     "http://localhost:8080",
		EnvironmentType: "test",
		Tools: []protocol.MCPTool{
			{
				Name:        "test.tool",
				Description: "Test tool",
			},
		},
		LastHeartbeat: time.Now().Add(-time.Hour), // Old heartbeat
	}

	registry.RegisterAgent(agent.ID, agent)

	// Check initial heartbeat
	retrievedAgent, _ := registry.GetAgent(agent.ID)
	oldHeartbeat := retrievedAgent.LastHeartbeat

	// Update heartbeat
	time.Sleep(time.Millisecond) // Ensure time difference
	registry.UpdateAgentHeartbeat(agent.ID)

	// Verify heartbeat was updated
	retrievedAgent, _ = registry.GetAgent(agent.ID)
	if !retrievedAgent.LastHeartbeat.After(oldHeartbeat) {
		t.Error("Heartbeat should have been updated")
	}
}