package main

import (
	"fmt"
	"math"
	"math/rand"
	"sort"
	"sync"
	"time"
)

// NewLoadBalancer creates a new load balancer with all strategies
func NewLoadBalancer() *LoadBalancer {
	lb := &LoadBalancer{
		strategies: make(map[LoadBalanceMode]LoadBalanceStrategy),
	}

	// Register all load balancing strategies
	lb.strategies[LoadBalanceRoundRobin] = &RoundRobinStrategy{}
	lb.strategies[LoadBalanceLeastLoaded] = &LeastLoadedStrategy{}
	lb.strategies[LoadBalanceWeightedRound] = &WeightedRoundRobinStrategy{}
	lb.strategies[LoadBalanceBestPerformance] = &BestPerformanceStrategy{}
	lb.strategies[LoadBalanceAffinityBased] = &AffinityBasedStrategy{}

	return lb
}

// SelectAgent selects the best agent using the specified load balancing mode
func (lb *LoadBalancer) SelectAgent(agents []string, metrics map[string]*AgentMetrics, context *RequestContext, mode LoadBalanceMode) (string, error) {
	lb.mutex.RLock()
	strategy, exists := lb.strategies[mode]
	lb.mutex.RUnlock()

	if !exists {
		return "", fmt.Errorf("unknown load balance mode: %s", mode)
	}

	if len(agents) == 0 {
		return "", fmt.Errorf("no agents available")
	}

	return strategy.SelectAgent(agents, metrics, context)
}

// RoundRobinStrategy implements simple round-robin load balancing
type RoundRobinStrategy struct {
	counter uint64
	mutex   sync.Mutex
}

func (rr *RoundRobinStrategy) SelectAgent(agents []string, metrics map[string]*AgentMetrics, context *RequestContext) (string, error) {
	if len(agents) == 0 {
		return "", fmt.Errorf("no agents available")
	}

	rr.mutex.Lock()
	defer rr.mutex.Unlock()

	// Filter healthy agents
	healthyAgents := make([]string, 0, len(agents))
	for _, agent := range agents {
		if metric, exists := metrics[agent]; !exists || metric.HealthScore > 0.5 {
			healthyAgents = append(healthyAgents, agent)
		}
	}

	if len(healthyAgents) == 0 {
		// Fall back to any agent if none are healthy
		healthyAgents = agents
	}

	rr.counter++
	selectedIndex := int(rr.counter) % len(healthyAgents)
	return healthyAgents[selectedIndex], nil
}

// LeastLoadedStrategy selects the agent with the lowest current load
type LeastLoadedStrategy struct{}

func (ll *LeastLoadedStrategy) SelectAgent(agents []string, metrics map[string]*AgentMetrics, context *RequestContext) (string, error) {
	if len(agents) == 0 {
		return "", fmt.Errorf("no agents available")
	}

	type agentLoad struct {
		agentID string
		load    float64
		health  float64
	}

	agentLoads := make([]agentLoad, 0, len(agents))

	for _, agent := range agents {
		metric, exists := metrics[agent]
		if !exists {
			// Unknown agent, assign neutral load
			agentLoads = append(agentLoads, agentLoad{
				agentID: agent,
				load:    0.5,
				health:  1.0,
			})
			continue
		}

		// Calculate combined load score (lower is better)
		loadScore := metric.LoadScore
		healthPenalty := (1.0 - metric.HealthScore) * 0.5 // Health issues increase effective load
		combinedLoad := loadScore + healthPenalty

		agentLoads = append(agentLoads, agentLoad{
			agentID: agent,
			load:    combinedLoad,
			health:  metric.HealthScore,
		})
	}

	// Sort by load (ascending) and health (descending)
	sort.Slice(agentLoads, func(i, j int) bool {
		if math.Abs(agentLoads[i].load-agentLoads[j].load) < 0.1 {
			return agentLoads[i].health > agentLoads[j].health
		}
		return agentLoads[i].load < agentLoads[j].load
	})

	return agentLoads[0].agentID, nil
}

// WeightedRoundRobinStrategy implements weighted round-robin based on agent capabilities
type WeightedRoundRobinStrategy struct {
	weights map[string]int
	current map[string]int
	mutex   sync.Mutex
}

func (wrr *WeightedRoundRobinStrategy) SelectAgent(agents []string, metrics map[string]*AgentMetrics, context *RequestContext) (string, error) {
	if len(agents) == 0 {
		return "", fmt.Errorf("no agents available")
	}

	wrr.mutex.Lock()
	defer wrr.mutex.Unlock()

	if wrr.weights == nil {
		wrr.weights = make(map[string]int)
		wrr.current = make(map[string]int)
	}

	// Calculate weights based on agent performance and health
	for _, agent := range agents {
		if _, exists := wrr.weights[agent]; !exists {
			metric, exists := metrics[agent]
			if !exists {
				wrr.weights[agent] = 1
				continue
			}

			// Weight based on performance (higher performance = higher weight)
			performanceScore := (1.0 - metric.ErrorRate) * metric.Availability * metric.HealthScore
			weight := int(performanceScore * 10)
			if weight < 1 {
				weight = 1
			}
			wrr.weights[agent] = weight
		}
	}

	// Find agent with highest current weight
	var selectedAgent string
	maxCurrent := -1

	for _, agent := range agents {
		wrr.current[agent] += wrr.weights[agent]
		if wrr.current[agent] > maxCurrent {
			maxCurrent = wrr.current[agent]
			selectedAgent = agent
		}
	}

	// Reduce current weight for selected agent
	if selectedAgent != "" {
		wrr.current[selectedAgent] -= maxCurrent
	}

	return selectedAgent, nil
}

// BestPerformanceStrategy selects the agent with the best overall performance
type BestPerformanceStrategy struct{}

func (bp *BestPerformanceStrategy) SelectAgent(agents []string, metrics map[string]*AgentMetrics, context *RequestContext) (string, error) {
	if len(agents) == 0 {
		return "", fmt.Errorf("no agents available")
	}

	type agentScore struct {
		agentID string
		score   float64
	}

	agentScores := make([]agentScore, 0, len(agents))

	for _, agent := range agents {
		metric, exists := metrics[agent]
		if !exists {
			// Unknown agent, assign neutral score
			agentScores = append(agentScores, agentScore{
				agentID: agent,
				score:   0.5,
			})
			continue
		}

		// Calculate comprehensive performance score
		score := bp.calculatePerformanceScore(metric, context)
		agentScores = append(agentScores, agentScore{
			agentID: agent,
			score:   score,
		})
	}

	// Sort by score (descending)
	sort.Slice(agentScores, func(i, j int) bool {
		return agentScores[i].score > agentScores[j].score
	})

	return agentScores[0].agentID, nil
}

func (bp *BestPerformanceStrategy) calculatePerformanceScore(metric *AgentMetrics, context *RequestContext) float64 {
	// Base performance metrics
	successRate := 1.0 - metric.ErrorRate
	availability := metric.Availability
	healthScore := metric.HealthScore
	
	// Latency score (lower latency is better)
	latencyScore := 1.0
	if metric.AverageResponseTime > 0 {
		// Normalize latency to a 0-1 score (assuming 5 seconds is very poor)
		maxAcceptableLatency := 5 * time.Second
		latencyScore = math.Max(0, 1.0-(float64(metric.AverageResponseTime)/float64(maxAcceptableLatency)))
	}

	// Load score (lower load is better)
	loadScore := math.Max(0, 1.0-metric.LoadScore)

	// Weighted combination based on context priority
	var weights struct {
		success   float64
		latency   float64
		health    float64
		load      float64
		available float64
	}

	// Adjust weights based on request priority
	switch context.Priority {
	case PriorityCritical:
		weights = struct {
			success   float64
			latency   float64
			health    float64
			load      float64
			available float64
		}{0.4, 0.3, 0.2, 0.05, 0.05}
	case PriorityHigh:
		weights = struct {
			success   float64
			latency   float64
			health    float64
			load      float64
			available float64
		}{0.3, 0.25, 0.2, 0.15, 0.1}
	default:
		weights = struct {
			success   float64
			latency   float64
			health    float64
			load      float64
			available float64
		}{0.25, 0.2, 0.2, 0.2, 0.15}
	}

	score := successRate*weights.success +
		latencyScore*weights.latency +
		healthScore*weights.health +
		loadScore*weights.load +
		availability*weights.available

	return math.Max(0, math.Min(1, score))
}

// AffinityBasedStrategy considers user preferences and geographic affinity
type AffinityBasedStrategy struct{}

func (ab *AffinityBasedStrategy) SelectAgent(agents []string, metrics map[string]*AgentMetrics, context *RequestContext) (string, error) {
	if len(agents) == 0 {
		return "", fmt.Errorf("no agents available")
	}

	type agentAffinity struct {
		agentID       string
		affinityScore float64
		performance   float64
	}

	agentAffinities := make([]agentAffinity, 0, len(agents))

	for _, agent := range agents {
		metric, exists := metrics[agent]
		
		affinityScore := ab.calculateAffinityScore(agent, metric, context)
		performanceScore := 0.5 // Default for unknown agents
		
		if exists {
			bp := &BestPerformanceStrategy{}
			performanceScore = bp.calculatePerformanceScore(metric, context)
		}

		agentAffinities = append(agentAffinities, agentAffinity{
			agentID:       agent,
			affinityScore: affinityScore,
			performance:   performanceScore,
		})
	}

	// Sort by combined affinity and performance score
	sort.Slice(agentAffinities, func(i, j int) bool {
		scoreI := agentAffinities[i].affinityScore*0.6 + agentAffinities[i].performance*0.4
		scoreJ := agentAffinities[j].affinityScore*0.6 + agentAffinities[j].performance*0.4
		return scoreI > scoreJ
	})

	return agentAffinities[0].agentID, nil
}

func (ab *AffinityBasedStrategy) calculateAffinityScore(agentID string, metric *AgentMetrics, context *RequestContext) float64 {
	score := 0.0

	// Check preferred agents
	for _, preferred := range context.AffinityPreferences {
		if agentID == preferred {
			score += 0.5
			break
		}
	}

	// Geographic affinity
	if metric != nil && context.GeographicRegion != "" {
		if metric.GeographicRegion == context.GeographicRegion {
			score += 0.3
		} else if metric.GeographicRegion != "" {
			// Same continent/region gets partial score
			// This is simplified - in practice you'd use proper geographic distance
			score += 0.1
		}
	}

	// Tool specialization (if agent frequently handles this type of tool)
	// This would require tracking tool usage history
	score += 0.2

	return math.Max(0, math.Min(1, score))
}

// AdaptiveStrategy adjusts selection based on historical performance
type AdaptiveStrategy struct {
	performanceHistory map[string]*PerformanceHistory
	mutex              sync.RWMutex
}

type PerformanceHistory struct {
	AgentID           string
	RecentSelections  []SelectionResult
	SuccessRate       float64
	AverageLatency    time.Duration
	LastUpdated       time.Time
}

type SelectionResult struct {
	Timestamp   time.Time
	Success     bool
	Latency     time.Duration
	ErrorType   string
}

func NewAdaptiveStrategy() *AdaptiveStrategy {
	return &AdaptiveStrategy{
		performanceHistory: make(map[string]*PerformanceHistory),
	}
}

func (as *AdaptiveStrategy) SelectAgent(agents []string, metrics map[string]*AgentMetrics, context *RequestContext) (string, error) {
	if len(agents) == 0 {
		return "", fmt.Errorf("no agents available")
	}

	// Use best performance strategy as base, then apply adaptive adjustments
	bp := &BestPerformanceStrategy{}
	
	type adaptiveScore struct {
		agentID    string
		baseScore  float64
		adaptiveAdjustment float64
		finalScore float64
	}

	scores := make([]adaptiveScore, 0, len(agents))

	as.mutex.RLock()
	defer as.mutex.RUnlock()

	for _, agent := range agents {
		metric, exists := metrics[agent]
		
		baseScore := 0.5
		if exists {
			baseScore = bp.calculatePerformanceScore(metric, context)
		}

		// Apply adaptive adjustment based on recent performance
		adaptiveAdjustment := as.getAdaptiveAdjustment(agent, context)
		finalScore := baseScore + adaptiveAdjustment

		scores = append(scores, adaptiveScore{
			agentID:            agent,
			baseScore:          baseScore,
			adaptiveAdjustment: adaptiveAdjustment,
			finalScore:         math.Max(0, math.Min(1, finalScore)),
		})
	}

	// Sort by final score
	sort.Slice(scores, func(i, j int) bool {
		return scores[i].finalScore > scores[j].finalScore
	})

	return scores[0].agentID, nil
}

func (as *AdaptiveStrategy) getAdaptiveAdjustment(agentID string, context *RequestContext) float64 {
	history, exists := as.performanceHistory[agentID]
	if !exists {
		return 0.0 // No history, no adjustment
	}

	// Calculate adjustment based on recent performance trends
	recentWindow := 10 // Last 10 selections
	if len(history.RecentSelections) < 3 {
		return 0.0 // Not enough data
	}

	recentSelections := history.RecentSelections
	if len(recentSelections) > recentWindow {
		recentSelections = recentSelections[len(recentSelections)-recentWindow:]
	}

	// Calculate recent success rate
	successes := 0
	totalLatency := time.Duration(0)
	for _, selection := range recentSelections {
		if selection.Success {
			successes++
		}
		totalLatency += selection.Latency
	}

	recentSuccessRate := float64(successes) / float64(len(recentSelections))
	avgRecentLatency := totalLatency / time.Duration(len(recentSelections))

	// Compare to historical averages
	successRateDiff := recentSuccessRate - history.SuccessRate
	latencyImprovement := 0.0
	if history.AverageLatency > 0 && avgRecentLatency > 0 {
		latencyImprovement = float64(history.AverageLatency-avgRecentLatency) / float64(history.AverageLatency)
	}

	// Calculate adjustment (-0.2 to +0.2)
	adjustment := (successRateDiff * 0.15) + (latencyImprovement * 0.05)
	return math.Max(-0.2, math.Min(0.2, adjustment))
}

// RecordSelection records the result of an agent selection for adaptive learning
func (as *AdaptiveStrategy) RecordSelection(agentID string, success bool, latency time.Duration, errorType string) {
	as.mutex.Lock()
	defer as.mutex.Unlock()

	history, exists := as.performanceHistory[agentID]
	if !exists {
		history = &PerformanceHistory{
			AgentID:          agentID,
			RecentSelections: make([]SelectionResult, 0),
		}
		as.performanceHistory[agentID] = history
	}

	// Add new selection result
	result := SelectionResult{
		Timestamp: time.Now(),
		Success:   success,
		Latency:   latency,
		ErrorType: errorType,
	}

	history.RecentSelections = append(history.RecentSelections, result)

	// Keep only recent history (last 50 selections)
	maxHistory := 50
	if len(history.RecentSelections) > maxHistory {
		history.RecentSelections = history.RecentSelections[len(history.RecentSelections)-maxHistory:]
	}

	// Update aggregated metrics
	as.updateAggregatedMetrics(history)
}

func (as *AdaptiveStrategy) updateAggregatedMetrics(history *PerformanceHistory) {
	if len(history.RecentSelections) == 0 {
		return
	}

	successes := 0
	totalLatency := time.Duration(0)

	for _, selection := range history.RecentSelections {
		if selection.Success {
			successes++
		}
		totalLatency += selection.Latency
	}

	history.SuccessRate = float64(successes) / float64(len(history.RecentSelections))
	history.AverageLatency = totalLatency / time.Duration(len(history.RecentSelections))
	history.LastUpdated = time.Now()
}

// MultiCriteriaStrategy combines multiple strategies with configurable weights
type MultiCriteriaStrategy struct {
	strategies map[string]LoadBalanceStrategy
	weights    map[string]float64
	mutex      sync.RWMutex
}

func NewMultiCriteriaStrategy(weights map[string]float64) *MultiCriteriaStrategy {
	strategies := map[string]LoadBalanceStrategy{
		"performance": &BestPerformanceStrategy{},
		"load":        &LeastLoadedStrategy{},
		"affinity":    &AffinityBasedStrategy{},
		"adaptive":    NewAdaptiveStrategy(),
	}

	// Default weights if none provided
	if weights == nil {
		weights = map[string]float64{
			"performance": 0.4,
			"load":        0.3,
			"affinity":    0.2,
			"adaptive":    0.1,
		}
	}

	return &MultiCriteriaStrategy{
		strategies: strategies,
		weights:    weights,
	}
}

func (mcs *MultiCriteriaStrategy) SelectAgent(agents []string, metrics map[string]*AgentMetrics, context *RequestContext) (string, error) {
	if len(agents) == 0 {
		return "", fmt.Errorf("no agents available")
	}

	agentScores := make(map[string]float64)

	mcs.mutex.RLock()
	defer mcs.mutex.RUnlock()

	// Get scores from each strategy
	for strategyName, strategy := range mcs.strategies {
		weight, exists := mcs.weights[strategyName]
		if !exists || weight <= 0 {
			continue
		}

		// This is a simplified approach - in practice, each strategy would return scores for all agents
		selectedAgent, err := strategy.SelectAgent(agents, metrics, context)
		if err != nil {
			continue
		}

		// Award points to selected agent (simplified scoring)
		agentScores[selectedAgent] += weight
	}

	// Find agent with highest combined score
	var bestAgent string
	var bestScore float64

	for agent, score := range agentScores {
		if score > bestScore {
			bestScore = score
			bestAgent = agent
		}
	}

	if bestAgent == "" {
		// Fallback to random selection
		return agents[rand.Intn(len(agents))], nil
	}

	return bestAgent, nil
}