package main

import (
	"testing"
	"time"

	"github.com/fep-fem/protocol"
)

func TestFederationManagerCreation(t *testing.T) {
	mcpRegistry := NewMCPRegistry()
	
	// Test with default config
	fm := NewFederationManager(mcpRegistry, nil)
	
	if fm.mcpRegistry != mcpRegistry {
		t.Error("MCP registry not set correctly")
	}
	
	if fm.config == nil {
		t.Error("Config should not be nil")
	}
	
	if fm.loadBalancer == nil {
		t.Error("Load balancer should be initialized")
	}
	
	if fm.healthChecker == nil {
		t.Error("Health checker should be initialized")
	}
	
	// Test default config values
	if fm.config.DefaultLoadBalanceMode != LoadBalanceBestPerformance {
		t.Errorf("Expected default load balance mode %s, got %s", 
			LoadBalanceBestPerformance, fm.config.DefaultLoadBalanceMode)
	}
	
	if fm.config.EnableSemanticSearch != true {
		t.Error("Semantic search should be enabled by default")
	}
}

func TestFederationManagerWithCustomConfig(t *testing.T) {
	mcpRegistry := NewMCPRegistry()
	
	config := &FederationConfig{
		MaxBrokers:             5,
		DefaultLoadBalanceMode: LoadBalanceRoundRobin,
		DefaultRoutingStrategy: RoutingLocal,
		EnableSemanticSearch:   false,
		EnableRanking:          false,
		HealthThreshold:        0.9,
	}
	
	fm := NewFederationManager(mcpRegistry, config)
	
	if fm.config.MaxBrokers != 5 {
		t.Errorf("Expected MaxBrokers 5, got %d", fm.config.MaxBrokers)
	}
	
	if fm.config.DefaultLoadBalanceMode != LoadBalanceRoundRobin {
		t.Errorf("Expected load balance mode %s, got %s", 
			LoadBalanceRoundRobin, fm.config.DefaultLoadBalanceMode)
	}
	
	if fm.semanticIndex != nil {
		t.Error("Semantic index should be nil when disabled")
	}
	
	if fm.rankingEngine != nil {
		t.Error("Ranking engine should be nil when disabled")
	}
}

func TestAdvancedToolDiscovery(t *testing.T) {
	mcpRegistry := NewMCPRegistry()
	fm := NewFederationManager(mcpRegistry, nil)
	
	// Register a test agent
	testAgent := &MCPAgent{
		ID:              "test-agent",
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
	
	err := mcpRegistry.RegisterAgent(testAgent.ID, testAgent)
	if err != nil {
		t.Fatalf("Failed to register test agent: %v", err)
	}
	
	// Add metrics for the agent so routing recommendations can be generated
	fm.agentMetrics["test-agent"] = &AgentMetrics{
		AgentID:             "test-agent",
		HealthScore:         0.9,
		AverageResponseTime: 100 * time.Millisecond,
		Availability:        0.95,
		ErrorRate:           0.05,
		LastUpdated:         time.Now(),
	}
	
	// Test advanced discovery
	query := protocol.ToolQuery{
		Capabilities:    []string{"math.*"},
		MaxResults:      10,
		IncludeMetadata: true,
	}
	
	context := &RequestContext{
		RequesterID:      "test-client",
		Priority:         PriorityNormal,
		GeographicRegion: "us-east",
	}
	
	result, err := fm.DiscoverToolsAdvanced(query, context)
	if err != nil {
		t.Fatalf("Advanced discovery failed: %v", err)
	}
	
	if result == nil {
		t.Fatal("Result should not be nil")
	}
	
	if len(result.BaseResults) == 0 {
		t.Error("Should have found at least one tool")
	}
	
	if result.FederationStats == nil {
		t.Error("Federation stats should be included")
	}
	
	if len(result.RoutingRecommendations) == 0 {
		t.Error("Should have routing recommendations")
	}
}

func TestToolRouting(t *testing.T) {
	mcpRegistry := NewMCPRegistry()
	fm := NewFederationManager(mcpRegistry, nil)
	
	// Register test agents
	agent1 := &MCPAgent{
		ID:              "agent-1",
		MCPEndpoint:     "http://localhost:8080",
		EnvironmentType: "test",
		Tools: []protocol.MCPTool{
			{Name: "math.add", Description: "Add numbers"},
		},
		LastHeartbeat: time.Now(),
	}
	
	agent2 := &MCPAgent{
		ID:              "agent-2", 
		MCPEndpoint:     "http://localhost:8081",
		EnvironmentType: "test",
		Tools: []protocol.MCPTool{
			{Name: "math.add", Description: "Add numbers"},
		},
		LastHeartbeat: time.Now(),
	}
	
	mcpRegistry.RegisterAgent(agent1.ID, agent1)
	mcpRegistry.RegisterAgent(agent2.ID, agent2)
	
	// Add some metrics to make agents selectable
	fm.agentMetrics["agent-1"] = &AgentMetrics{
		AgentID:             "agent-1",
		HealthScore:         0.9,
		AverageResponseTime: 100 * time.Millisecond,
		Availability:        0.95,
		ErrorRate:           0.05,
	}
	
	fm.agentMetrics["agent-2"] = &AgentMetrics{
		AgentID:             "agent-2",
		HealthScore:         0.8,
		AverageResponseTime: 200 * time.Millisecond,
		Availability:        0.90,
		ErrorRate:           0.10,
	}
	
	// Test tool routing
	context := &RequestContext{
		RequesterID: "test-client",
		ToolName:    "math.add",
		Priority:    PriorityNormal,
	}
	
	decision, err := fm.RouteToolInvocation("math.add", "", context)
	if err != nil {
		t.Fatalf("Tool routing failed: %v", err)
	}
	
	if decision == nil {
		t.Fatal("Routing decision should not be nil")
	}
	
	if decision.SelectedAgent == "" {
		t.Error("Should have selected an agent")
	}
	
	// Agent 1 should be selected due to better metrics
	if decision.SelectedAgent != "agent-1" {
		t.Errorf("Expected agent-1 to be selected, got %s", decision.SelectedAgent)
	}
	
	if len(decision.AlternativeAgents) == 0 {
		t.Error("Should have alternative agents")
	}
}

func TestFederationStats(t *testing.T) {
	mcpRegistry := NewMCPRegistry()
	fm := NewFederationManager(mcpRegistry, nil)
	
	// Register test agents
	agent := &MCPAgent{
		ID:              "test-agent",
		MCPEndpoint:     "http://localhost:8080",
		EnvironmentType: "test",
		Tools: []protocol.MCPTool{
			{Name: "test.tool", Description: "Test tool"},
		},
		LastHeartbeat: time.Now(),
	}
	
	mcpRegistry.RegisterAgent(agent.ID, agent)
	
	// Add test metrics
	fm.agentMetrics["test-agent"] = &AgentMetrics{
		AgentID:             "test-agent",
		HealthScore:         0.85,
		AverageResponseTime: 150 * time.Millisecond,
		Availability:        0.92,
	}
	
	// Add a federated broker
	fm.federatedBrokers["broker-1"] = &FederatedBroker{
		ID:       "broker-1",
		Endpoint: "https://broker1.example.com",
		Status:   BrokerStatusActive,
		LastSeen: time.Now(),
	}
	
	stats := fm.getFederationStats()
	
	if stats.TotalAgents != 1 {
		t.Errorf("Expected 1 agent, got %d", stats.TotalAgents)
	}
	
	if stats.TotalTools != 1 {
		t.Errorf("Expected 1 tool, got %d", stats.TotalTools)
	}
	
	if stats.TotalBrokers != 1 {
		t.Errorf("Expected 1 broker, got %d", stats.TotalBrokers)
	}
	
	if stats.ActiveBrokers != 1 {
		t.Errorf("Expected 1 active broker, got %d", stats.ActiveBrokers)
	}
	
	if stats.AverageResponseTime != 150*time.Millisecond {
		t.Errorf("Expected average response time 150ms, got %v", stats.AverageResponseTime)
	}
	
	if stats.OverallHealthScore != 0.85 {
		t.Errorf("Expected health score 0.85, got %f", stats.OverallHealthScore)
	}
}

func TestLoadBalancerStrategies(t *testing.T) {
	lb := NewLoadBalancer()
	
	agents := []string{"agent-1", "agent-2", "agent-3"}
	metrics := map[string]*AgentMetrics{
		"agent-1": {
			AgentID:             "agent-1",
			HealthScore:         0.9,
			LoadScore:           0.3,
			AverageResponseTime: 100 * time.Millisecond,
			ErrorRate:           0.05,
			Availability:        0.95,
		},
		"agent-2": {
			AgentID:             "agent-2",
			HealthScore:         0.8,
			LoadScore:           0.7,
			AverageResponseTime: 200 * time.Millisecond,
			ErrorRate:           0.10,
			Availability:        0.90,
		},
		"agent-3": {
			AgentID:             "agent-3",
			HealthScore:         0.95,
			LoadScore:           0.1,
			AverageResponseTime: 80 * time.Millisecond,
			ErrorRate:           0.02,
			Availability:        0.98,
		},
	}
	
	context := &RequestContext{
		RequesterID: "test-client",
		Priority:    PriorityHigh,
	}
	
	// Test best performance strategy
	agent, err := lb.SelectAgent(agents, metrics, context, LoadBalanceBestPerformance)
	if err != nil {
		t.Fatalf("Best performance selection failed: %v", err)
	}
	
	// Agent 3 should be selected (best overall performance)
	if agent != "agent-3" {
		t.Errorf("Expected agent-3 for best performance, got %s", agent)
	}
	
	// Test least loaded strategy
	agent, err = lb.SelectAgent(agents, metrics, context, LoadBalanceLeastLoaded)
	if err != nil {
		t.Fatalf("Least loaded selection failed: %v", err)
	}
	
	// Agent 3 should be selected (lowest load)
	if agent != "agent-3" {
		t.Errorf("Expected agent-3 for least loaded, got %s", agent)
	}
	
	// Test round robin
	selections := make(map[string]int)
	for i := 0; i < 30; i++ {
		agent, err := lb.SelectAgent(agents, metrics, context, LoadBalanceRoundRobin)
		if err != nil {
			t.Fatalf("Round robin selection failed: %v", err)
		}
		selections[agent]++
	}
	
	// Should have distributed selections across agents
	if len(selections) == 0 {
		t.Error("Round robin should have selected agents")
	}
}

func TestSemanticIndex(t *testing.T) {
	si := NewSemanticIndex()
	
	// Test tool indexing
	tool1 := protocol.MCPTool{
		Name:        "math.add",
		Description: "Add two numbers together",
	}
	
	tool2 := protocol.MCPTool{
		Name:        "math.subtract",
		Description: "Subtract one number from another",
	}
	
	tool3 := protocol.MCPTool{
		Name:        "file.read",
		Description: "Read contents of a file",
	}
	
	si.IndexTool("agent-1", tool1)
	si.IndexTool("agent-1", tool2)
	si.IndexTool("agent-2", tool3)
	
	// Test semantic similarity
	query := protocol.ToolQuery{
		Capabilities: []string{"math", "calculate"},
	}
	
	score1 := si.calculateSemanticScore(tool1, query)
	score2 := si.calculateSemanticScore(tool2, query)
	score3 := si.calculateSemanticScore(tool3, query)
	
	// Math tools should have higher similarity to math query
	if score1 <= score3 || score2 <= score3 {
		t.Error("Math tools should have higher semantic similarity to math query")
	}
	
	// Test similar tools finding
	similar := si.findSimilarTools("math.add")
	found := false
	for _, sim := range similar {
		if sim.ToolName == "math.subtract" {
			found = true
			break
		}
	}
	
	if !found {
		t.Error("Should find math.subtract as similar to math.add")
	}
	
	// Test categories
	categories := si.getToolCategories("math.add")
	hasCategory := false
	for _, cat := range categories {
		if cat == "mathematics" {
			hasCategory = true
			break
		}
	}
	
	if !hasCategory {
		t.Error("Math tool should be categorized as mathematics")
	}
}

func TestRankingEngine(t *testing.T) {
	re := NewRankingEngine()
	
	// Test tool ranking
	tools := []protocol.DiscoveredTool{
		{
			AgentID:         "agent-1",
			MCPEndpoint:     "http://localhost:8080",
			Capabilities:    []string{"math.add"},
			EnvironmentType: "local",
			MCPTools: []protocol.MCPTool{
				{Name: "math.add", Description: "Add numbers"},
			},
			Metadata: protocol.ToolMetadata{
				AverageResponseTime: 100,
				TrustScore:          0.9,
			},
		},
		{
			AgentID:         "agent-2",
			MCPEndpoint:     "http://remote.example.com:8080",
			Capabilities:    []string{"math.add"},
			EnvironmentType: "cloud",
			MCPTools: []protocol.MCPTool{
				{Name: "math.add", Description: "Add numbers"},
			},
			Metadata: protocol.ToolMetadata{
				AverageResponseTime: 300,
				TrustScore:          0.7,
			},
		},
	}
	
	context := &RequestContext{
		RequesterID:      "test-client",
		Priority:         PriorityHigh,
		GeographicRegion: "local",
	}
	
	ranked := re.RankTools(tools, context)
	
	if len(ranked) != 2 {
		t.Errorf("Expected 2 ranked tools, got %d", len(ranked))
	}
	
	// First tool should be agent-1 (better performance, local)
	if ranked[0].Tool.AgentID != "agent-1" {
		t.Errorf("Expected agent-1 to be ranked first, got %s", ranked[0].Tool.AgentID)
	}
	
	// Check that scores are calculated
	if ranked[0].OverallScore <= 0 {
		t.Error("Overall score should be greater than 0")
	}
	
	if ranked[0].PerformanceScore <= 0 {
		t.Error("Performance score should be greater than 0")
	}
}

func TestHealthChecker(t *testing.T) {
	hc := NewHealthChecker(1*time.Second, 0.8)
	
	if hc.checkInterval != 1*time.Second {
		t.Errorf("Expected check interval 1s, got %v", hc.checkInterval)
	}
	
	if hc.healthThreshold != 0.8 {
		t.Errorf("Expected health threshold 0.8, got %f", hc.healthThreshold)
	}
	
	// Test status determination
	status := hc.determineAgentStatus(0.9)
	if status != AgentStatusHealthy {
		t.Errorf("Expected healthy status for score 0.9, got %s", status)
	}
	
	status = hc.determineAgentStatus(0.7)
	if status != AgentStatusDegraded {
		t.Errorf("Expected degraded status for score 0.7, got %s", status)
	}
	
	status = hc.determineAgentStatus(0.3)
	if status != AgentStatusUnhealthy {
		t.Errorf("Expected unhealthy status for score 0.3, got %s", status)
	}
}

func TestRequestContext(t *testing.T) {
	context := &RequestContext{
		RequesterID:         "test-client",
		ToolName:           "math.add",
		Priority:           PriorityHigh,
		LatencyRequirement: 100 * time.Millisecond,
		GeographicRegion:   "us-east",
		AffinityPreferences: []string{"preferred-agent"},
	}
	
	if context.RequesterID != "test-client" {
		t.Error("RequesterID not set correctly")
	}
	
	if context.Priority != PriorityHigh {
		t.Error("Priority not set correctly")
	}
	
	if context.LatencyRequirement != 100*time.Millisecond {
		t.Error("Latency requirement not set correctly")
	}
}

func TestFederationConfigDefaults(t *testing.T) {
	mcpRegistry := NewMCPRegistry()
	fm := NewFederationManager(mcpRegistry, nil)
	
	config := fm.config
	
	if config.MaxBrokers != 10 {
		t.Errorf("Expected default MaxBrokers 10, got %d", config.MaxBrokers)
	}
	
	if config.BrokerSyncInterval != 30*time.Second {
		t.Errorf("Expected default BrokerSyncInterval 30s, got %v", config.BrokerSyncInterval)
	}
	
	if config.DefaultLoadBalanceMode != LoadBalanceBestPerformance {
		t.Errorf("Expected default load balance mode %s, got %s", 
			LoadBalanceBestPerformance, config.DefaultLoadBalanceMode)
	}
	
	if config.HealthThreshold != 0.8 {
		t.Errorf("Expected default health threshold 0.8, got %f", config.HealthThreshold)
	}
	
	if !config.EnableSemanticSearch {
		t.Error("Semantic search should be enabled by default")
	}
	
	if !config.EnableRanking {
		t.Error("Ranking should be enabled by default")
	}
}