package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"net/http"
	"sync"
	"time"
)

// NewHealthChecker creates a new health checker
func NewHealthChecker(checkInterval time.Duration, healthThreshold float64) *HealthChecker {
	return &HealthChecker{
		checkInterval:     checkInterval,
		healthThreshold:   healthThreshold,
		degradedThreshold: healthThreshold * 0.7,
		stopChan:         make(chan struct{}),
	}
}

// Start begins the health checking process
func (hc *HealthChecker) Start(fm *FederationManager) {
	go hc.healthCheckLoop(fm)
}

// Stop stops the health checking process
func (hc *HealthChecker) Stop() {
	close(hc.stopChan)
}

// healthCheckLoop runs the periodic health checks
func (hc *HealthChecker) healthCheckLoop(fm *FederationManager) {
	ticker := time.NewTicker(hc.checkInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			hc.performHealthChecks(fm)
		case <-hc.stopChan:
			return
		}
	}
}

// performHealthChecks executes health checks for all agents and brokers
func (hc *HealthChecker) performHealthChecks(fm *FederationManager) {
	// Check agent health
	hc.checkAgentHealth(fm)
	
	// Check federated broker health
	hc.checkBrokerHealth(fm)
}

// checkAgentHealth performs health checks on all registered agents
func (hc *HealthChecker) checkAgentHealth(fm *FederationManager) {
	agents := fm.mcpRegistry.ListTools()
	
	// Group tools by agent
	agentEndpoints := make(map[string]string)
	for _, tool := range agents {
		if _, exists := agentEndpoints[tool.AgentID]; !exists {
			agentEndpoints[tool.AgentID] = tool.MCPEndpoint
		}
	}

	var wg sync.WaitGroup
	for agentID, endpoint := range agentEndpoints {
		wg.Add(1)
		go func(id, ep string) {
			defer wg.Done()
			hc.checkSingleAgent(fm, id, ep)
		}(agentID, endpoint)
	}
	
	wg.Wait()
}

// checkSingleAgent performs a health check on a single agent
func (hc *HealthChecker) checkSingleAgent(fm *FederationManager, agentID, endpoint string) {
	startTime := time.Now()
	healthScore := 0.0
	
	// Perform basic connectivity check
	isReachable := hc.checkAgentConnectivity(endpoint)
	if isReachable {
		healthScore += 0.4
	}
	
	// Perform capability verification
	capabilityScore := hc.checkAgentCapabilities(endpoint)
	healthScore += capabilityScore * 0.3
	
	// Check response time
	responseTime := time.Since(startTime)
	timeScore := hc.calculateTimeScore(responseTime)
	healthScore += timeScore * 0.3
	
	// Update agent metrics
	fm.metricsMutex.Lock()
	metrics, exists := fm.agentMetrics[agentID]
	if !exists {
		metrics = &AgentMetrics{
			AgentID: agentID,
		}
		fm.agentMetrics[agentID] = metrics
	}
	
	metrics.HealthScore = healthScore
	metrics.LastHealthCheck = time.Now()
	metrics.LastResponseTime = responseTime
	
	// Update availability tracking
	if isReachable {
		metrics.SuccessfulRequests++
	} else {
		metrics.FailedRequests++
	}
	
	total := metrics.SuccessfulRequests + metrics.FailedRequests
	if total > 0 {
		metrics.Availability = float64(metrics.SuccessfulRequests) / float64(total)
		metrics.ErrorRate = float64(metrics.FailedRequests) / float64(total)
	}
	
	// Update average response time
	if metrics.AverageResponseTime == 0 {
		metrics.AverageResponseTime = responseTime
	} else {
		// Exponential moving average
		alpha := 0.3
		metrics.AverageResponseTime = time.Duration(float64(metrics.AverageResponseTime)*(1-alpha) + float64(responseTime)*alpha)
	}
	
	metrics.LastUpdated = time.Now()
	fm.metricsMutex.Unlock()
}

// checkAgentConnectivity checks if an agent endpoint is reachable
func (hc *HealthChecker) checkAgentConnectivity(endpoint string) bool {
	client := &http.Client{
		Timeout: 5 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	
	// Try a simple health check endpoint
	healthURL := endpoint + "/health"
	resp, err := client.Get(healthURL)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	
	return resp.StatusCode == http.StatusOK
}

// checkAgentCapabilities verifies that an agent can respond to capability queries
func (hc *HealthChecker) checkAgentCapabilities(endpoint string) float64 {
	client := &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	
	// Create a simple capability check request
	checkReq := map[string]interface{}{
		"method": "tools/list",
		"id":     "health-check",
	}
	
	reqData, err := json.Marshal(checkReq)
	if err != nil {
		return 0.0
	}
	
	resp, err := client.Post(endpoint, "application/json", bytes.NewReader(reqData))
	if err != nil {
		return 0.0
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return 0.5
	}
	
	// Try to parse response
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0.7
	}
	
	// Full capability response received
	return 1.0
}

// calculateTimeScore converts response time to a score (0-1)
func (hc *HealthChecker) calculateTimeScore(responseTime time.Duration) float64 {
	// Score based on response time thresholds
	if responseTime <= 100*time.Millisecond {
		return 1.0
	} else if responseTime <= 500*time.Millisecond {
		return 0.8
	} else if responseTime <= 1*time.Second {
		return 0.6
	} else if responseTime <= 5*time.Second {
		return 0.4
	} else {
		return 0.2
	}
}

// checkBrokerHealth performs health checks on federated brokers
func (hc *HealthChecker) checkBrokerHealth(fm *FederationManager) {
	fm.topologyMutex.RLock()
	brokers := make([]*FederatedBroker, 0, len(fm.federatedBrokers))
	for _, broker := range fm.federatedBrokers {
		brokers = append(brokers, broker)
	}
	fm.topologyMutex.RUnlock()
	
	var wg sync.WaitGroup
	for _, broker := range brokers {
		wg.Add(1)
		go func(b *FederatedBroker) {
			defer wg.Done()
			hc.checkSingleBroker(fm, b)
		}(broker)
	}
	
	wg.Wait()
}

// checkSingleBroker performs a health check on a single federated broker
func (hc *HealthChecker) checkSingleBroker(fm *FederationManager, broker *FederatedBroker) {
	startTime := time.Now()
	
	client := &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	
	// Check broker health endpoint
	healthURL := broker.Endpoint + "/health"
	resp, err := client.Get(healthURL)
	
	responseTime := time.Since(startTime)
	
	fm.topologyMutex.Lock()
	defer fm.topologyMutex.Unlock()
	
	if err != nil {
		broker.Status = BrokerStatusUnreachable
		broker.ResponseTime = responseTime
		return
	}
	defer resp.Body.Close()
	
	broker.ResponseTime = responseTime
	broker.LastSeen = time.Now()
	
	if resp.StatusCode == http.StatusOK {
		// Try to get additional broker stats
		statsURL := broker.Endpoint + "/federation/stats"
		statsResp, err := client.Get(statsURL)
		
		if err == nil && statsResp.StatusCode == http.StatusOK {
			var stats struct {
				ToolCount   int     `json:"toolCount"`
				LoadScore   float64 `json:"loadScore"`
				AgentCount  int     `json:"agentCount"`
			}
			
			if json.NewDecoder(statsResp.Body).Decode(&stats) == nil {
				broker.ToolCount = stats.ToolCount
				broker.LoadScore = stats.LoadScore
			}
			statsResp.Body.Close()
		}
		
		// Determine status based on response time and other factors
		if responseTime < 1*time.Second {
			broker.Status = BrokerStatusActive
		} else if responseTime < 5*time.Second {
			broker.Status = BrokerStatusDegraded
		} else {
			broker.Status = BrokerStatusDegraded
		}
		
		// Update trust score based on performance
		hc.updateBrokerTrustScore(broker, responseTime)
	} else {
		broker.Status = BrokerStatusDegraded
	}
}

// updateBrokerTrustScore updates a broker's trust score based on recent performance
func (hc *HealthChecker) updateBrokerTrustScore(broker *FederatedBroker, responseTime time.Duration) {
	// Simple trust score calculation based on response time and availability
	timeScore := hc.calculateTimeScore(responseTime)
	
	// Exponential moving average for trust score
	alpha := 0.2
	broker.TrustScore = broker.TrustScore*(1-alpha) + timeScore*alpha
	
	// Ensure trust score stays within bounds
	if broker.TrustScore < 0 {
		broker.TrustScore = 0
	} else if broker.TrustScore > 1 {
		broker.TrustScore = 1
	}
}

// GetAgentHealthStatus returns the current health status of all agents
func (hc *HealthChecker) GetAgentHealthStatus(fm *FederationManager) map[string]*AgentHealthStatus {
	fm.metricsMutex.RLock()
	defer fm.metricsMutex.RUnlock()
	
	status := make(map[string]*AgentHealthStatus)
	
	for agentID, metrics := range fm.agentMetrics {
		healthStatus := &AgentHealthStatus{
			AgentID:          agentID,
			HealthScore:      metrics.HealthScore,
			Status:           hc.determineAgentStatus(metrics.HealthScore),
			LastCheck:        metrics.LastHealthCheck,
			ResponseTime:     metrics.LastResponseTime,
			Availability:     metrics.Availability,
			ErrorRate:        metrics.ErrorRate,
			TotalRequests:    metrics.TotalRequests,
			FailedRequests:   metrics.FailedRequests,
		}
		
		status[agentID] = healthStatus
	}
	
	return status
}

// AgentHealthStatus represents the health status of an agent
type AgentHealthStatus struct {
	AgentID        string        `json:"agentId"`
	HealthScore    float64       `json:"healthScore"`
	Status         AgentStatus   `json:"status"`
	LastCheck      time.Time     `json:"lastCheck"`
	ResponseTime   time.Duration `json:"responseTime"`
	Availability   float64       `json:"availability"`
	ErrorRate      float64       `json:"errorRate"`
	TotalRequests  int64         `json:"totalRequests"`
	FailedRequests int64         `json:"failedRequests"`
}

// AgentStatus represents the status of an agent
type AgentStatus string

const (
	AgentStatusHealthy   AgentStatus = "healthy"
	AgentStatusDegraded  AgentStatus = "degraded"
	AgentStatusUnhealthy AgentStatus = "unhealthy"
	AgentStatusUnknown   AgentStatus = "unknown"
)

// determineAgentStatus determines agent status based on health score
func (hc *HealthChecker) determineAgentStatus(healthScore float64) AgentStatus {
	if healthScore >= hc.healthThreshold {
		return AgentStatusHealthy
	} else if healthScore >= hc.degradedThreshold {
		return AgentStatusDegraded
	} else if healthScore > 0 {
		return AgentStatusUnhealthy
	} else {
		return AgentStatusUnknown
	}
}

// GetBrokerHealthStatus returns the current health status of all federated brokers
func (hc *HealthChecker) GetBrokerHealthStatus(fm *FederationManager) map[string]*BrokerHealthStatus {
	fm.topologyMutex.RLock()
	defer fm.topologyMutex.RUnlock()
	
	status := make(map[string]*BrokerHealthStatus)
	
	for brokerID, broker := range fm.federatedBrokers {
		healthStatus := &BrokerHealthStatus{
			BrokerID:     brokerID,
			Endpoint:     broker.Endpoint,
			Status:       broker.Status,
			LastSeen:     broker.LastSeen,
			ResponseTime: broker.ResponseTime,
			TrustScore:   broker.TrustScore,
			ToolCount:    broker.ToolCount,
			LoadScore:    broker.LoadScore,
		}
		
		status[brokerID] = healthStatus
	}
	
	return status
}

// BrokerHealthStatus represents the health status of a federated broker
type BrokerHealthStatus struct {
	BrokerID     string        `json:"brokerId"`
	Endpoint     string        `json:"endpoint"`
	Status       BrokerStatus  `json:"status"`
	LastSeen     time.Time     `json:"lastSeen"`
	ResponseTime time.Duration `json:"responseTime"`
	TrustScore   float64       `json:"trustScore"`
	ToolCount    int           `json:"toolCount"`
	LoadScore    float64       `json:"loadScore"`
}

// PerformManualHealthCheck triggers an immediate health check for a specific agent
func (hc *HealthChecker) PerformManualHealthCheck(fm *FederationManager, agentID string) *AgentHealthStatus {
	// Find agent endpoint
	tools := fm.mcpRegistry.ListTools()
	var endpoint string
	
	for _, tool := range tools {
		if tool.AgentID == agentID {
			endpoint = tool.MCPEndpoint
			break
		}
	}
	
	if endpoint == "" {
		return &AgentHealthStatus{
			AgentID: agentID,
			Status:  AgentStatusUnknown,
		}
	}
	
	// Perform health check
	hc.checkSingleAgent(fm, agentID, endpoint)
	
	// Return updated status
	status := hc.GetAgentHealthStatus(fm)
	if agentStatus, exists := status[agentID]; exists {
		return agentStatus
	}
	
	return &AgentHealthStatus{
		AgentID: agentID,
		Status:  AgentStatusUnknown,
	}
}

// GetOverallFederationHealth calculates the overall health of the federation
func (hc *HealthChecker) GetOverallFederationHealth(fm *FederationManager) *FederationHealth {
	agentStatus := hc.GetAgentHealthStatus(fm)
	brokerStatus := hc.GetBrokerHealthStatus(fm)
	
	health := &FederationHealth{
		Timestamp: time.Now(),
	}
	
	// Calculate agent health statistics
	var totalAgentHealth float64
	healthyAgents := 0
	degradedAgents := 0
	unhealthyAgents := 0
	
	for _, status := range agentStatus {
		totalAgentHealth += status.HealthScore
		switch status.Status {
		case AgentStatusHealthy:
			healthyAgents++
		case AgentStatusDegraded:
			degradedAgents++
		case AgentStatusUnhealthy:
			unhealthyAgents++
		}
	}
	
	totalAgents := len(agentStatus)
	if totalAgents > 0 {
		health.AverageAgentHealth = totalAgentHealth / float64(totalAgents)
	}
	
	health.HealthyAgents = healthyAgents
	health.DegradedAgents = degradedAgents
	health.UnhealthyAgents = unhealthyAgents
	health.TotalAgents = totalAgents
	
	// Calculate broker health statistics
	activeBrokers := 0
	degradedBrokers := 0
	unreachableBrokers := 0
	
	for _, status := range brokerStatus {
		switch status.Status {
		case BrokerStatusActive:
			activeBrokers++
		case BrokerStatusDegraded:
			degradedBrokers++
		case BrokerStatusUnreachable:
			unreachableBrokers++
		}
	}
	
	health.ActiveBrokers = activeBrokers
	health.DegradedBrokers = degradedBrokers
	health.UnreachableBrokers = unreachableBrokers
	health.TotalBrokers = len(brokerStatus)
	
	// Calculate overall health score
	agentHealthWeight := 0.7
	brokerHealthWeight := 0.3
	
	agentScore := health.AverageAgentHealth
	brokerScore := 0.0
	if health.TotalBrokers > 0 {
		brokerScore = float64(activeBrokers) / float64(health.TotalBrokers)
	}
	
	health.OverallHealth = agentScore*agentHealthWeight + brokerScore*brokerHealthWeight
	
	// Determine overall status
	if health.OverallHealth >= hc.healthThreshold {
		health.OverallStatus = "healthy"
	} else if health.OverallHealth >= hc.degradedThreshold {
		health.OverallStatus = "degraded"
	} else {
		health.OverallStatus = "unhealthy"
	}
	
	return health
}

// FederationHealth represents the overall health of the federation
type FederationHealth struct {
	Timestamp            time.Time `json:"timestamp"`
	OverallHealth        float64   `json:"overallHealth"`
	OverallStatus        string    `json:"overallStatus"`
	AverageAgentHealth   float64   `json:"averageAgentHealth"`
	TotalAgents          int       `json:"totalAgents"`
	HealthyAgents        int       `json:"healthyAgents"`
	DegradedAgents       int       `json:"degradedAgents"`
	UnhealthyAgents      int       `json:"unhealthyAgents"`
	TotalBrokers         int       `json:"totalBrokers"`
	ActiveBrokers        int       `json:"activeBrokers"`
	DegradedBrokers      int       `json:"degradedBrokers"`
	UnreachableBrokers   int       `json:"unreachableBrokers"`
}