package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/fep-fem/protocol"
)

// MCPRegistry manages MCP tool discovery and agent embodiment
type MCPRegistry struct {
	tools  map[string]*RegisteredTool
	agents map[string]*MCPAgent
	mu     sync.RWMutex
}

// RegisteredTool represents a tool that's been indexed for discovery
type RegisteredTool struct {
	AgentID         string
	Tool            protocol.MCPTool
	MCPEndpoint     string
	EnvironmentType string
	RegisteredAt    time.Time
	LastSeen        time.Time
}

// MCPAgent represents an agent with MCP capabilities
type MCPAgent struct {
	ID              string
	MCPEndpoint     string
	BodyDefinition  *protocol.BodyDefinition
	EnvironmentType string
	Tools           []protocol.MCPTool
	LastHeartbeat   time.Time
}

// NewMCPRegistry creates a new MCP registry instance
func NewMCPRegistry() *MCPRegistry {
	return &MCPRegistry{
		tools:  make(map[string]*RegisteredTool),
		agents: make(map[string]*MCPAgent),
	}
}

// RegisterAgent registers an agent and indexes its MCP tools
func (r *MCPRegistry) RegisterAgent(agentID string, agent *MCPAgent) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.agents[agentID] = agent

	// Index all tools for discovery
	for _, tool := range agent.Tools {
		toolKey := fmt.Sprintf("%s/%s", agentID, tool.Name)
		r.tools[toolKey] = &RegisteredTool{
			AgentID:         agentID,
			Tool:            tool,
			MCPEndpoint:     agent.MCPEndpoint,
			EnvironmentType: agent.EnvironmentType,
			RegisteredAt:    time.Now(),
			LastSeen:        time.Now(),
		}
	}

	return nil
}

// GetAgent retrieves an agent by ID
func (r *MCPRegistry) GetAgent(agentID string) (*MCPAgent, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	agent, exists := r.agents[agentID]
	return agent, exists
}

// ListTools returns all registered tools
func (r *MCPRegistry) ListTools() []*RegisteredTool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tools := make([]*RegisteredTool, 0, len(r.tools))
	for _, tool := range r.tools {
		tools = append(tools, tool)
	}
	return tools
}

// UnregisterAgent removes an agent and all its tools
func (r *MCPRegistry) UnregisterAgent(agentID string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Remove agent
	delete(r.agents, agentID)

	// Remove all tools for this agent
	for toolKey, tool := range r.tools {
		if tool.AgentID == agentID {
			delete(r.tools, toolKey)
		}
	}
}

// DiscoverTools finds tools matching the given query
func (r *MCPRegistry) DiscoverTools(query protocol.ToolQuery) ([]protocol.DiscoveredTool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Simple matching logic - will be enhanced in later phases
	var matchingTools []*RegisteredTool

	for _, tool := range r.tools {
		// Match capabilities
		if r.matchesCapabilities(tool, query.Capabilities) {
			// Filter by environment if specified
			if query.EnvironmentType == "" || tool.EnvironmentType == query.EnvironmentType {
				matchingTools = append(matchingTools, tool)
			}
		}
	}

	// Apply max results limit
	if query.MaxResults > 0 && len(matchingTools) > query.MaxResults {
		matchingTools = matchingTools[:query.MaxResults]
	}

	// Group tools by agent
	agentTools := make(map[string][]protocol.MCPTool)
	agentInfo := make(map[string]*RegisteredTool)

	for _, tool := range matchingTools {
		agentTools[tool.AgentID] = append(agentTools[tool.AgentID], tool.Tool)
		agentInfo[tool.AgentID] = tool // Store agent info
	}

	// Build discovery response
	var discovered []protocol.DiscoveredTool
	for agentID, tools := range agentTools {
		info := agentInfo[agentID]
		discovered = append(discovered, protocol.DiscoveredTool{
			AgentID:         agentID,
			MCPEndpoint:     info.MCPEndpoint,
			Capabilities:    r.extractCapabilities(tools),
			EnvironmentType: info.EnvironmentType,
			MCPTools:        tools,
			Metadata: protocol.ToolMetadata{
				LastSeen:            info.LastSeen.UnixMilli(),
				AverageResponseTime: 150, // Placeholder
				TrustScore:          0.95, // Placeholder
			},
		})
	}

	return discovered, nil
}

// matchesCapabilities checks if a tool matches any of the capability patterns
func (r *MCPRegistry) matchesCapabilities(tool *RegisteredTool, capabilities []string) bool {
	if len(capabilities) == 0 {
		return true // No filter means match all
	}

	toolName := tool.Tool.Name
	for _, cap := range capabilities {
		if r.matchCapability(toolName, cap) {
			return true
		}
	}
	return false
}

// matchCapability performs pattern matching for a single capability
func (r *MCPRegistry) matchCapability(toolName, pattern string) bool {
	// Simple pattern matching - supports wildcards like "file.*"
	if pattern == "*" {
		return true
	}

	if len(pattern) > 0 && pattern[len(pattern)-1] == '*' {
		prefix := pattern[:len(pattern)-1]
		return len(toolName) >= len(prefix) && toolName[:len(prefix)] == prefix
	}

	return toolName == pattern
}

// extractCapabilities extracts capability names from tools
func (r *MCPRegistry) extractCapabilities(tools []protocol.MCPTool) []string {
	capabilities := make([]string, 0, len(tools))
	for _, tool := range tools {
		capabilities = append(capabilities, tool.Name)
	}
	return capabilities
}

// UpdateAgentHeartbeat updates the last seen time for an agent
func (r *MCPRegistry) UpdateAgentHeartbeat(agentID string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if agent, exists := r.agents[agentID]; exists {
		agent.LastHeartbeat = time.Now()

		// Update tool last seen times
		for _, tool := range r.tools {
			if tool.AgentID == agentID {
				tool.LastSeen = time.Now()
			}
		}
	}
}

// GetToolCount returns the total number of registered tools
func (r *MCPRegistry) GetToolCount() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.tools)
}

// GetAgentCount returns the total number of registered MCP agents
func (r *MCPRegistry) GetAgentCount() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.agents)
}