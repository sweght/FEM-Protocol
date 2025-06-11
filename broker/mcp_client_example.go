package main

import (
	"fmt"
	"log"
	"time"

	"github.com/fep-fem/protocol"
)

// ExampleMCPClientUsage demonstrates how to use the MCP client library
func ExampleMCPClientUsage() {
	// Generate key pair for the client
	_, privKey, err := protocol.GenerateKeyPair()
	if err != nil {
		log.Fatalf("Failed to generate key pair: %v", err)
	}

	// Create MCP client configuration
	config := MCPClientConfig{
		AgentID:        "example-client-001",
		BrokerURL:      "https://broker.example.com:4433",
		PrivateKey:     privKey,
		CacheExpiry:    10 * time.Minute,
		RequestTimeout: 30 * time.Second,
		TLSInsecure:    true, // Only for development
	}

	// Create the client
	client := NewMCPClient(config)

	// Example 1: Discover all available tools
	fmt.Println("=== Discovering All Available Tools ===")
	allAgents, err := client.GetAvailableAgents()
	if err != nil {
		log.Printf("Failed to discover agents: %v", err)
		return
	}

	fmt.Printf("Found %d agents with MCP tools:\n", len(allAgents))
	for _, agent := range allAgents {
		fmt.Printf("  Agent: %s (%s)\n", agent.AgentID, agent.EnvironmentType)
		fmt.Printf("    Endpoint: %s\n", agent.MCPEndpoint)
		fmt.Printf("    Tools: %d\n", len(agent.MCPTools))
		for _, tool := range agent.MCPTools {
			fmt.Printf("      - %s: %s\n", tool.Name, tool.Description)
		}
		fmt.Println()
	}

	// Example 2: Find specific tools by capability
	fmt.Println("=== Finding Math Tools ===")
	mathTools, err := client.FindToolsByCapability([]string{"math.*"})
	if err != nil {
		log.Printf("Failed to find math tools: %v", err)
		return
	}

	for _, agent := range mathTools {
		fmt.Printf("Math tools from %s:\n", agent.AgentID)
		for _, tool := range agent.MCPTools {
			fmt.Printf("  %s: %s\n", tool.Name, tool.Description)
			if schema, ok := tool.InputSchema["properties"]; ok {
				fmt.Printf("    Parameters: %v\n", schema)
			}
		}
	}

	// Example 3: Find tools in specific environment
	fmt.Println("=== Finding Production Tools ===")
	prodTools, err := client.FindToolsInEnvironment("production", 5)
	if err != nil {
		log.Printf("Failed to find production tools: %v", err)
		return
	}

	fmt.Printf("Found %d production tools\n", len(prodTools))

	// Example 4: Custom tool discovery with advanced query
	fmt.Println("=== Custom Tool Discovery ===")
	customQuery := protocol.ToolQuery{
		Capabilities:    []string{"file.*", "code.*"},
		EnvironmentType: "development",
		MaxResults:      10,
		IncludeMetadata: true,
	}

	customTools, err := client.DiscoverTools(customQuery)
	if err != nil {
		log.Printf("Failed custom discovery: %v", err)
		return
	}

	fmt.Printf("Custom query found %d agents\n", len(customTools))
	for _, agent := range customTools {
		fmt.Printf("  %s: %d tools (trust: %.2f)\n", 
			agent.AgentID, 
			len(agent.MCPTools),
			agent.Metadata.TrustScore)
	}

	// Example 5: Tool invocation
	fmt.Println("=== Invoking Tools ===")
	if len(allAgents) > 0 && len(allAgents[0].MCPTools) > 0 {
		agent := allAgents[0]
		tool := agent.MCPTools[0]
		
		parameters := map[string]interface{}{
			"input": "test data",
		}

		fmt.Printf("Calling %s on agent %s...\n", tool.Name, agent.AgentID)
		result, err := client.CallTool(agent.AgentID, tool.Name, parameters)
		if err != nil {
			log.Printf("Tool call failed: %v", err)
		} else {
			fmt.Printf("Tool call result: %v\n", result)
		}
	}

	// Example 6: Cache management
	fmt.Println("=== Cache Statistics ===")
	stats := client.GetCacheStats()
	for key, value := range stats {
		fmt.Printf("%s: %v\n", key, value)
	}

	// Refresh cache if needed
	fmt.Println("Refreshing tool cache...")
	client.RefreshCache()

	newStats := client.GetCacheStats()
	fmt.Printf("Cache entries after refresh: %v\n", newStats["cached_queries"])
}

// ExampleAgentWithMCPClient shows how an agent might use the MCP client
func ExampleAgentWithMCPClient() {
	// This example shows how an agent could use the MCP client
	// to discover and use tools from other agents

	fmt.Println("=== Agent Using MCP Client ===")

	// Agent setup (in real scenario, this would be part of agent initialization)
	_, privKey, err := protocol.GenerateKeyPair()
	if err != nil {
		log.Fatalf("Failed to generate key pair: %v", err)
	}

	client := NewMCPClient(MCPClientConfig{
		AgentID:     "intelligent-agent-001",
		BrokerURL:   "https://broker.example.com:4433",
		PrivateKey:  privKey,
		TLSInsecure: true,
	})

	// Scenario: Agent needs to perform mathematical operations
	fmt.Println("Agent needs to perform math operations...")
	
	mathAgents, err := client.FindToolsByCapability([]string{"math.*"})
	if err != nil {
		log.Printf("Could not find math tools: %v", err)
		return
	}

	if len(mathAgents) == 0 {
		fmt.Println("No math tools available")
		return
	}

	// Use the first available math agent
	mathAgent := mathAgents[0]
	fmt.Printf("Using math tools from agent: %s\n", mathAgent.AgentID)

	// Find addition tool
	var addTool *protocol.MCPTool
	for _, tool := range mathAgent.MCPTools {
		if tool.Name == "math.add" || tool.Name == "add" {
			addTool = &tool
			break
		}
	}

	if addTool != nil {
		fmt.Printf("Found addition tool: %s\n", addTool.Name)
		
		// Call the tool
		result, err := client.CallTool(mathAgent.AgentID, addTool.Name, map[string]interface{}{
			"a": 15,
			"b": 27,
		})
		
		if err != nil {
			log.Printf("Addition failed: %v", err)
		} else {
			fmt.Printf("15 + 27 = %v\n", result)
		}
	}

	// Scenario: Agent needs to work with files
	fmt.Println("\nAgent needs to work with files...")
	
	fileAgents, err := client.FindToolsByCapability([]string{"file.*"})
	if err != nil {
		log.Printf("Could not find file tools: %v", err)
		return
	}

	fmt.Printf("Found %d agents with file capabilities\n", len(fileAgents))
	for _, agent := range fileAgents {
		fmt.Printf("  %s: %v\n", agent.AgentID, agent.Capabilities)
	}
}

// ExampleErrorHandling demonstrates error handling patterns
func ExampleErrorHandling() {
	fmt.Println("=== Error Handling Examples ===")

	// Invalid configuration
	client := NewMCPClient(MCPClientConfig{
		AgentID:   "error-test",
		BrokerURL: "invalid-url",
		// Missing private key will cause issues
	})

	// This will fail due to invalid URL
	_, err := client.FindToolsByCapability([]string{"test.*"})
	if err != nil {
		fmt.Printf("Expected error for invalid URL: %v\n", err)
	}

	// Generate proper key for valid client
	_, privKey, err := protocol.GenerateKeyPair()
	if err != nil {
		log.Fatalf("Failed to generate key pair: %v", err)
	}

	validClient := NewMCPClient(MCPClientConfig{
		AgentID:     "valid-client",
		BrokerURL:   "https://nonexistent.example.com",
		PrivateKey:  privKey,
		TLSInsecure: true,
	})

	// This will fail due to connection error
	_, err = validClient.FindToolsByCapability([]string{"test.*"})
	if err != nil {
		fmt.Printf("Expected connection error: %v\n", err)
	}

	// Tool call with invalid parameters
	_, err = validClient.CallTool("nonexistent-agent", "nonexistent-tool", nil)
	if err != nil {
		fmt.Printf("Expected tool call error: %v\n", err)
	}
}

// Run examples (commented out since this is a library file)
// func main() {
// 	ExampleMCPClientUsage()
// 	ExampleAgentWithMCPClient()  
// 	ExampleErrorHandling()
// }