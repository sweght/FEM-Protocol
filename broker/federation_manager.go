package main

import (
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/fep-fem/protocol"
)

// FederationManager handles advanced tool federation, routing, and load balancing
type FederationManager struct {
	// Core registries
	mcpRegistry *MCPRegistry
	
	// Federation topology
	federatedBrokers map[string]*FederatedBroker
	routingTable     map[string]*ToolRoute
	topologyMutex    sync.RWMutex
	
	// Load balancing and performance
	agentMetrics     map[string]*AgentMetrics
	loadBalancer     *LoadBalancer
	healthChecker    *HealthChecker
	metricsMutex     sync.RWMutex
	
	// Discovery enhancement
	semanticIndex    *SemanticIndex
	rankingEngine    *RankingEngine
	
	// Configuration
	config *FederationConfig
}

// FederatedBroker represents a peer broker in the federation
type FederatedBroker struct {
	ID               string
	Endpoint         string
	PublicKey        string
	LastSeen         time.Time
	Status           BrokerStatus
	Capabilities     []string
	TrustScore       float64
	ResponseTime     time.Duration
	ToolCount        int
	LoadScore        float64
}

// BrokerStatus represents the status of a federated broker
type BrokerStatus string

const (
	BrokerStatusActive      BrokerStatus = "active"
	BrokerStatusDegraded    BrokerStatus = "degraded"
	BrokerStatusUnreachable BrokerStatus = "unreachable"
	BrokerStatusMaintenance BrokerStatus = "maintenance"
)

// ToolRoute defines how to route requests for specific tools
type ToolRoute struct {
	ToolPattern      string
	PrimaryAgents    []string
	FallbackAgents   []string
	LoadBalanceMode  LoadBalanceMode
	RoutingStrategy  RoutingStrategy
	HealthThreshold  float64
	LastUpdated      time.Time
}

// LoadBalanceMode defines different load balancing strategies
type LoadBalanceMode string

const (
	LoadBalanceRoundRobin    LoadBalanceMode = "round_robin"
	LoadBalanceLeastLoaded   LoadBalanceMode = "least_loaded"
	LoadBalanceWeightedRound LoadBalanceMode = "weighted_round"
	LoadBalanceBestPerformance LoadBalanceMode = "best_performance"
	LoadBalanceAffinityBased LoadBalanceMode = "affinity_based"
)

// RoutingStrategy defines different routing approaches
type RoutingStrategy string

const (
	RoutingLocal        RoutingStrategy = "local_first"
	RoutingFederated    RoutingStrategy = "federated_first"
	RoutingBestFit      RoutingStrategy = "best_fit"
	RoutingMulticast    RoutingStrategy = "multicast"
	RoutingGeographicAware RoutingStrategy = "geographic_aware"
)

// AgentMetrics tracks performance and health metrics for agents
type AgentMetrics struct {
	AgentID              string
	TotalRequests        int64
	SuccessfulRequests   int64
	FailedRequests       int64
	AverageResponseTime  time.Duration
	LastResponseTime     time.Duration
	ErrorRate            float64
	Availability         float64
	ThroughputPerSecond  float64
	LastHealthCheck      time.Time
	HealthScore          float64
	LoadScore            float64
	GeographicRegion     string
	LastUpdated          time.Time
}

// LoadBalancer handles intelligent load distribution
type LoadBalancer struct {
	strategies map[LoadBalanceMode]LoadBalanceStrategy
	mutex      sync.RWMutex
}

// LoadBalanceStrategy interface for different load balancing algorithms
type LoadBalanceStrategy interface {
	SelectAgent(agents []string, metrics map[string]*AgentMetrics, context *RequestContext) (string, error)
}

// RequestContext provides context for routing and load balancing decisions
type RequestContext struct {
	RequesterID      string
	ToolName         string
	Parameters       map[string]interface{}
	Priority         RequestPriority
	LatencyRequirement time.Duration
	GeographicRegion string
	AffinityPreferences []string
}

// RequestPriority defines request priority levels
type RequestPriority string

const (
	PriorityLow      RequestPriority = "low"
	PriorityNormal   RequestPriority = "normal"
	PriorityHigh     RequestPriority = "high"
	PriorityCritical RequestPriority = "critical"
)

// HealthChecker monitors agent and broker health
type HealthChecker struct {
	checkInterval    time.Duration
	healthThreshold  float64
	degradedThreshold float64
	stopChan         chan struct{}
	mutex            sync.RWMutex
}

// SemanticIndex provides advanced tool discovery capabilities
type SemanticIndex struct {
	toolVectors    map[string][]float64
	categoryIndex  map[string][]string
	similarityCache map[string][]SimilarityResult
	mutex          sync.RWMutex
}

// SimilarityResult represents semantic similarity between tools
type SimilarityResult struct {
	ToolName   string
	AgentID    string
	Similarity float64
}

// RankingEngine provides intelligent tool ranking
type RankingEngine struct {
	rankingFactors map[string]float64
	userPreferences map[string]UserPreferences
	mutex          sync.RWMutex
}

// UserPreferences stores user-specific ranking preferences
type UserPreferences struct {
	PreferredAgents      []string
	PreferredRegions     []string
	PerformanceWeight    float64
	ReliabilityWeight    float64
	CostWeight           float64
	LatencyWeight        float64
}

// FederationConfig holds configuration for the federation manager
type FederationConfig struct {
	// Topology management
	MaxBrokers           int
	BrokerSyncInterval   time.Duration
	TopologyUpdateInterval time.Duration
	
	// Load balancing
	DefaultLoadBalanceMode LoadBalanceMode
	DefaultRoutingStrategy RoutingStrategy
	HealthCheckInterval    time.Duration
	HealthThreshold        float64
	
	// Discovery enhancement
	EnableSemanticSearch   bool
	EnableRanking          bool
	SimilarityThreshold    float64
	
	// Performance
	MetricsRetentionPeriod time.Duration
	CacheUpdateInterval    time.Duration
}

// NewFederationManager creates a new federation manager
func NewFederationManager(mcpRegistry *MCPRegistry, config *FederationConfig) *FederationManager {
	if config == nil {
		config = &FederationConfig{
			MaxBrokers:             10,
			BrokerSyncInterval:     30 * time.Second,
			TopologyUpdateInterval: 60 * time.Second,
			DefaultLoadBalanceMode: LoadBalanceBestPerformance,
			DefaultRoutingStrategy: RoutingBestFit,
			HealthCheckInterval:    15 * time.Second,
			HealthThreshold:        0.8,
			EnableSemanticSearch:   true,
			EnableRanking:          true,
			SimilarityThreshold:    0.7,
			MetricsRetentionPeriod: 24 * time.Hour,
			CacheUpdateInterval:    5 * time.Minute,
		}
	}

	fm := &FederationManager{
		mcpRegistry:      mcpRegistry,
		federatedBrokers: make(map[string]*FederatedBroker),
		routingTable:     make(map[string]*ToolRoute),
		agentMetrics:     make(map[string]*AgentMetrics),
		config:           config,
	}

	// Initialize subsystems
	fm.loadBalancer = NewLoadBalancer()
	fm.healthChecker = NewHealthChecker(config.HealthCheckInterval, config.HealthThreshold)
	
	if config.EnableSemanticSearch {
		fm.semanticIndex = NewSemanticIndex()
	}
	
	if config.EnableRanking {
		fm.rankingEngine = NewRankingEngine()
	}

	// Start background processes
	if config.TopologyUpdateInterval > 0 {
		go fm.startTopologyManager()
	}
	if config.CacheUpdateInterval > 0 {
		go fm.startMetricsCollector()
	}

	return fm
}

// DiscoverToolsAdvanced performs enhanced tool discovery with ranking and routing
func (fm *FederationManager) DiscoverToolsAdvanced(query protocol.ToolQuery, context *RequestContext) (*AdvancedDiscoveryResult, error) {
	// Get base discovery results
	baseTools, err := fm.mcpRegistry.DiscoverTools(query)
	if err != nil {
		return nil, fmt.Errorf("base discovery failed: %w", err)
	}

	result := &AdvancedDiscoveryResult{
		BaseResults:    baseTools,
		RequestContext: context,
		Timestamp:      time.Now(),
	}

	// Apply semantic enhancement if enabled
	if fm.config.EnableSemanticSearch && fm.semanticIndex != nil {
		semanticResults := fm.enhanceWithSemanticSearch(baseTools, query)
		result.SemanticResults = semanticResults
	}

	// Apply ranking if enabled
	if fm.config.EnableRanking && fm.rankingEngine != nil {
		rankedResults := fm.rankingEngine.RankTools(baseTools, context)
		result.RankedResults = rankedResults
	}

	// Generate routing recommendations
	routingRecommendations := fm.generateRoutingRecommendations(baseTools, context)
	result.RoutingRecommendations = routingRecommendations

	// Add federation-wide statistics
	result.FederationStats = fm.getFederationStats()

	return result, nil
}

// AdvancedDiscoveryResult contains enhanced discovery results
type AdvancedDiscoveryResult struct {
	BaseResults             []protocol.DiscoveredTool
	SemanticResults         []SemanticDiscoveryResult
	RankedResults           []RankedTool
	RoutingRecommendations  []RoutingRecommendation
	FederationStats         *FederationStats
	RequestContext          *RequestContext
	Timestamp               time.Time
}

// SemanticDiscoveryResult represents semantically enhanced tool discovery
type SemanticDiscoveryResult struct {
	Tool            protocol.DiscoveredTool
	SemanticScore   float64
	RelatedTools    []SimilarityResult
	Categories      []string
	ConceptVector   []float64
}

// RankedTool represents a tool with calculated ranking score
type RankedTool struct {
	Tool              protocol.DiscoveredTool
	OverallScore      float64
	PerformanceScore  float64
	ReliabilityScore  float64
	LatencyScore      float64
	CostScore         float64
	AffinityScore     float64
	RankingFactors    map[string]float64
}

// RoutingRecommendation suggests optimal routing for tool invocation
type RoutingRecommendation struct {
	ToolName            string
	RecommendedAgent    string
	AlternativeAgents   []string
	RoutingStrategy     RoutingStrategy
	LoadBalanceMode     LoadBalanceMode
	ExpectedLatency     time.Duration
	ConfidenceScore     float64
	Justification       string
}

// FederationStats provides federation-wide statistics
type FederationStats struct {
	TotalBrokers        int
	ActiveBrokers       int
	TotalAgents         int
	TotalTools          int
	AverageResponseTime time.Duration
	OverallHealthScore  float64
	LoadDistribution    map[string]float64
	TopPerformingAgents []string
	GeographicDistribution map[string]int
	LastUpdated         time.Time
}

// RouteToolInvocation intelligently routes tool invocations
func (fm *FederationManager) RouteToolInvocation(toolName string, agentID string, context *RequestContext) (*RoutingDecision, error) {
	fm.topologyMutex.RLock()
	route, exists := fm.routingTable[toolName]
	fm.topologyMutex.RUnlock()

	if !exists {
		// Create default route
		route = &ToolRoute{
			ToolPattern:     toolName,
			LoadBalanceMode: fm.config.DefaultLoadBalanceMode,
			RoutingStrategy: fm.config.DefaultRoutingStrategy,
			HealthThreshold: fm.config.HealthThreshold,
			LastUpdated:     time.Now(),
		}
	}

	// Get available agents for this tool
	availableAgents := fm.getAvailableAgentsForTool(toolName, agentID)
	if len(availableAgents) == 0 {
		return nil, fmt.Errorf("no available agents for tool %s", toolName)
	}

	// Select best agent using load balancer
	selectedAgent, err := fm.loadBalancer.SelectAgent(availableAgents, fm.agentMetrics, context, route.LoadBalanceMode)
	if err != nil {
		return nil, fmt.Errorf("agent selection failed: %w", err)
	}

	decision := &RoutingDecision{
		SelectedAgent:     selectedAgent,
		RoutingStrategy:   route.RoutingStrategy,
		LoadBalanceMode:   route.LoadBalanceMode,
		AlternativeAgents: availableAgents,
		Justification:     fmt.Sprintf("Selected using %s strategy", route.LoadBalanceMode),
		Timestamp:         time.Now(),
	}

	// Update metrics
	fm.updateRoutingMetrics(toolName, selectedAgent, context)

	return decision, nil
}

// RoutingDecision represents the result of intelligent routing
type RoutingDecision struct {
	SelectedAgent     string
	RoutingStrategy   RoutingStrategy
	LoadBalanceMode   LoadBalanceMode
	AlternativeAgents []string
	ExpectedLatency   time.Duration
	ConfidenceScore   float64
	Justification     string
	Timestamp         time.Time
}

// Helper methods

func (fm *FederationManager) enhanceWithSemanticSearch(tools []protocol.DiscoveredTool, query protocol.ToolQuery) []SemanticDiscoveryResult {
	if fm.semanticIndex == nil {
		return nil
	}

	results := make([]SemanticDiscoveryResult, 0, len(tools))
	
	for _, tool := range tools {
		for _, mcpTool := range tool.MCPTools {
			// Calculate semantic score (simplified implementation)
			semanticScore := fm.semanticIndex.calculateSemanticScore(mcpTool, query)
			
			if semanticScore > fm.config.SimilarityThreshold {
				result := SemanticDiscoveryResult{
					Tool:          tool,
					SemanticScore: semanticScore,
					RelatedTools:  fm.semanticIndex.findSimilarTools(mcpTool.Name),
					Categories:    fm.semanticIndex.getToolCategories(mcpTool.Name),
				}
				results = append(results, result)
			}
		}
	}

	// Sort by semantic score
	sort.Slice(results, func(i, j int) bool {
		return results[i].SemanticScore > results[j].SemanticScore
	})

	return results
}

func (fm *FederationManager) generateRoutingRecommendations(tools []protocol.DiscoveredTool, context *RequestContext) []RoutingRecommendation {
	recommendations := make([]RoutingRecommendation, 0)

	for _, tool := range tools {
		for _, mcpTool := range tool.MCPTools {
			// Get metrics for this agent
			fm.metricsMutex.RLock()
			metrics, exists := fm.agentMetrics[tool.AgentID]
			fm.metricsMutex.RUnlock()

			if !exists {
				continue
			}

			recommendation := RoutingRecommendation{
				ToolName:         mcpTool.Name,
				RecommendedAgent: tool.AgentID,
				RoutingStrategy:  fm.config.DefaultRoutingStrategy,
				LoadBalanceMode:  fm.config.DefaultLoadBalanceMode,
				ExpectedLatency:  metrics.AverageResponseTime,
				ConfidenceScore:  metrics.HealthScore,
				Justification:    fmt.Sprintf("Agent health: %.2f, avg latency: %v", metrics.HealthScore, metrics.AverageResponseTime),
			}

			recommendations = append(recommendations, recommendation)
		}
	}

	return recommendations
}

func (fm *FederationManager) getAvailableAgentsForTool(toolName string, preferredAgent string) []string {
	agents := make([]string, 0)
	
	// Add preferred agent first if available and healthy
	if preferredAgent != "" {
		fm.metricsMutex.RLock()
		if metrics, exists := fm.agentMetrics[preferredAgent]; exists && metrics.HealthScore > fm.config.HealthThreshold {
			agents = append(agents, preferredAgent)
		}
		fm.metricsMutex.RUnlock()
	}

	// Add other healthy agents
	allTools := fm.mcpRegistry.ListTools()
	for _, tool := range allTools {
		if tool.Tool.Name == toolName && tool.AgentID != preferredAgent {
			fm.metricsMutex.RLock()
			if metrics, exists := fm.agentMetrics[tool.AgentID]; exists && metrics.HealthScore > fm.config.HealthThreshold {
				agents = append(agents, tool.AgentID)
			}
			fm.metricsMutex.RUnlock()
		}
	}

	return agents
}

func (fm *FederationManager) updateRoutingMetrics(toolName, agentID string, context *RequestContext) {
	fm.metricsMutex.Lock()
	defer fm.metricsMutex.Unlock()

	metrics, exists := fm.agentMetrics[agentID]
	if !exists {
		metrics = &AgentMetrics{
			AgentID:     agentID,
			LastUpdated: time.Now(),
		}
		fm.agentMetrics[agentID] = metrics
	}

	metrics.TotalRequests++
	metrics.LastUpdated = time.Now()
}

func (fm *FederationManager) getFederationStats() *FederationStats {
	fm.topologyMutex.RLock()
	totalBrokers := len(fm.federatedBrokers)
	activeBrokers := 0
	for _, broker := range fm.federatedBrokers {
		if broker.Status == BrokerStatusActive {
			activeBrokers++
		}
	}
	fm.topologyMutex.RUnlock()

	totalAgents := fm.mcpRegistry.GetAgentCount()
	totalTools := fm.mcpRegistry.GetToolCount()

	// Calculate average response time and health score
	fm.metricsMutex.RLock()
	var totalResponseTime time.Duration
	var totalHealthScore float64
	agentCount := 0
	
	for _, metrics := range fm.agentMetrics {
		totalResponseTime += metrics.AverageResponseTime
		totalHealthScore += metrics.HealthScore
		agentCount++
	}
	fm.metricsMutex.RUnlock()

	var avgResponseTime time.Duration
	var avgHealthScore float64
	if agentCount > 0 {
		avgResponseTime = totalResponseTime / time.Duration(agentCount)
		avgHealthScore = totalHealthScore / float64(agentCount)
	}

	return &FederationStats{
		TotalBrokers:       totalBrokers,
		ActiveBrokers:      activeBrokers,
		TotalAgents:        totalAgents,
		TotalTools:         totalTools,
		AverageResponseTime: avgResponseTime,
		OverallHealthScore: avgHealthScore,
		LastUpdated:        time.Now(),
	}
}

// Background processes

func (fm *FederationManager) startTopologyManager() {
	ticker := time.NewTicker(fm.config.TopologyUpdateInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			fm.updateTopology()
		}
	}
}

func (fm *FederationManager) startMetricsCollector() {
	ticker := time.NewTicker(fm.config.CacheUpdateInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			fm.collectMetrics()
		}
	}
}

func (fm *FederationManager) updateTopology() {
	// Update federated broker status and topology
	// This would typically involve pinging other brokers, updating routing tables, etc.
	// Simplified implementation for now
}

func (fm *FederationManager) collectMetrics() {
	// Collect performance metrics from agents and brokers
	// Update health scores, response times, etc.
	// Simplified implementation for now
}