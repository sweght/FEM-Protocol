package main

import (
	"math"
	"sort"
	"strings"
	"sync" // Used for mutex in SemanticIndex and RankingEngine

	"github.com/fep-fem/protocol"
)

// NewSemanticIndex creates a new semantic index
func NewSemanticIndex() *SemanticIndex {
	si := &SemanticIndex{
		toolVectors:     make(map[string][]float64),
		categoryIndex:   make(map[string][]string),
		similarityCache: make(map[string][]SimilarityResult),
		mutex:           sync.RWMutex{},
	}
	return si
}

// IndexTool adds a tool to the semantic index
func (si *SemanticIndex) IndexTool(agentID string, tool protocol.MCPTool) {
	si.mutex.Lock()
	defer si.mutex.Unlock()

	toolKey := agentID + "/" + tool.Name
	
	// Generate semantic vector for the tool
	vector := si.generateSemanticVector(tool)
	si.toolVectors[toolKey] = vector

	// Categorize the tool
	categories := si.categorizeTool(tool)
	si.categoryIndex[toolKey] = categories

	// Clear similarity cache as it's now outdated
	si.similarityCache = make(map[string][]SimilarityResult)
}

// generateSemanticVector creates a semantic vector representation of a tool
func (si *SemanticIndex) generateSemanticVector(tool protocol.MCPTool) []float64 {
	// This is a simplified semantic vector generation
	// In practice, you might use word embeddings, TF-IDF, or ML models
	
	vector := make([]float64, 100) // 100-dimensional vector
	
	// Extract features from tool name and description
	text := strings.ToLower(tool.Name + " " + tool.Description)
	words := strings.Fields(text)
	
	// Simple keyword-based feature extraction
	keywords := map[string]int{
		"file": 0, "read": 1, "write": 2, "create": 3, "delete": 4,
		"math": 5, "calculate": 6, "compute": 7, "add": 8, "subtract": 9,
		"code": 10, "execute": 11, "run": 12, "compile": 13, "debug": 14,
		"data": 15, "process": 16, "transform": 17, "filter": 18, "sort": 19,
		"network": 20, "http": 21, "api": 22, "request": 23, "response": 24,
		"database": 25, "query": 26, "insert": 27, "update": 28, "select": 29,
		"text": 30, "parse": 31, "format": 32, "search": 33, "replace": 34,
		"image": 35, "resize": 36, "convert": 37, "crop": 38, "rotate": 39,
		"security": 40, "encrypt": 41, "decrypt": 42, "hash": 43, "verify": 44,
		"time": 45, "date": 46, "schedule": 47, "timer": 48, "wait": 49,
	}
	
	// Set vector values based on keyword presence
	for _, word := range words {
		if index, exists := keywords[word]; exists && index < len(vector) {
			vector[index] = 1.0
		}
	}
	
	// Add some random variation to make vectors more unique
	for i := 50; i < len(vector); i++ {
		if len(tool.Name) > i-50 {
			vector[i] = float64(tool.Name[i-50]) / 255.0
		}
	}
	
	return si.normalizeVector(vector)
}

// normalizeVector normalizes a vector to unit length
func (si *SemanticIndex) normalizeVector(vector []float64) []float64 {
	var magnitude float64
	for _, v := range vector {
		magnitude += v * v
	}
	magnitude = math.Sqrt(magnitude)
	
	if magnitude == 0 {
		return vector
	}
	
	normalized := make([]float64, len(vector))
	for i, v := range vector {
		normalized[i] = v / magnitude
	}
	
	return normalized
}

// categorizeTool assigns categories to a tool based on its characteristics
func (si *SemanticIndex) categorizeTool(tool protocol.MCPTool) []string {
	categories := make([]string, 0)
	
	text := strings.ToLower(tool.Name + " " + tool.Description)
	
	categoryKeywords := map[string][]string{
		"file_management": {"file", "read", "write", "create", "delete", "copy", "move", "rename"},
		"mathematics":     {"math", "calculate", "compute", "add", "subtract", "multiply", "divide", "equation"},
		"code_execution":  {"code", "execute", "run", "compile", "debug", "script", "program"},
		"data_processing": {"data", "process", "transform", "filter", "sort", "parse", "format"},
		"networking":      {"network", "http", "api", "request", "response", "url", "web"},
		"database":        {"database", "query", "insert", "update", "select", "sql", "table"},
		"text_processing": {"text", "string", "search", "replace", "regex", "pattern"},
		"image_processing": {"image", "photo", "picture", "resize", "convert", "crop", "rotate"},
		"security":        {"security", "encrypt", "decrypt", "hash", "verify", "authenticate"},
		"time_operations": {"time", "date", "schedule", "timer", "wait", "delay", "timestamp"},
		"ai_ml":          {"ai", "ml", "machine", "learning", "model", "predict", "classify"},
		"system":         {"system", "process", "memory", "cpu", "disk", "monitor"},
	}
	
	for category, keywords := range categoryKeywords {
		for _, keyword := range keywords {
			if strings.Contains(text, keyword) {
				categories = append(categories, category)
				break
			}
		}
	}
	
	if len(categories) == 0 {
		categories = append(categories, "general")
	}
	
	return categories
}

// calculateSemanticScore calculates semantic similarity between a tool and query
func (si *SemanticIndex) calculateSemanticScore(tool protocol.MCPTool, query protocol.ToolQuery) float64 {
	si.mutex.RLock()
	defer si.mutex.RUnlock()
	
	// Generate query vector
	queryTool := protocol.MCPTool{
		Name:        strings.Join(query.Capabilities, " "),
		Description: query.EnvironmentType,
	}
	queryVector := si.generateSemanticVector(queryTool)
	
	// Get tool vector
	// For simplicity, assume we can generate it on the fly
	toolVector := si.generateSemanticVector(tool)
	
	// Calculate cosine similarity
	return si.cosineSimilarity(toolVector, queryVector)
}

// cosineSimilarity calculates cosine similarity between two vectors
func (si *SemanticIndex) cosineSimilarity(vec1, vec2 []float64) float64 {
	if len(vec1) != len(vec2) {
		return 0.0
	}
	
	var dotProduct, magnitude1, magnitude2 float64
	
	for i := 0; i < len(vec1); i++ {
		dotProduct += vec1[i] * vec2[i]
		magnitude1 += vec1[i] * vec1[i]
		magnitude2 += vec2[i] * vec2[i]
	}
	
	magnitude1 = math.Sqrt(magnitude1)
	magnitude2 = math.Sqrt(magnitude2)
	
	if magnitude1 == 0 || magnitude2 == 0 {
		return 0.0
	}
	
	return dotProduct / (magnitude1 * magnitude2)
}

// findSimilarTools finds tools similar to the given tool
func (si *SemanticIndex) findSimilarTools(toolName string) []SimilarityResult {
	si.mutex.RLock()
	defer si.mutex.RUnlock()
	
	// Check cache first
	if cached, exists := si.similarityCache[toolName]; exists {
		return cached
	}
	
	// Find the tool vector
	var targetVector []float64
	targetKey := ""
	for key := range si.toolVectors {
		if strings.HasSuffix(key, "/"+toolName) {
			targetVector = si.toolVectors[key]
			targetKey = key
			break
		}
	}
	
	if targetVector == nil {
		return nil
	}
	
	// Calculate similarities with all other tools
	similarities := make([]SimilarityResult, 0)
	
	for key, vector := range si.toolVectors {
		if key == targetKey {
			continue
		}
		
		similarity := si.cosineSimilarity(targetVector, vector)
		if similarity > 0.3 { // Threshold for similarity
			parts := strings.Split(key, "/")
			if len(parts) == 2 {
				similarities = append(similarities, SimilarityResult{
					ToolName:   parts[1],
					AgentID:    parts[0],
					Similarity: similarity,
				})
			}
		}
	}
	
	// Sort by similarity
	sort.Slice(similarities, func(i, j int) bool {
		return similarities[i].Similarity > similarities[j].Similarity
	})
	
	// Keep top 10
	if len(similarities) > 10 {
		similarities = similarities[:10]
	}
	
	// Cache the result
	si.similarityCache[toolName] = similarities
	
	return similarities
}

// getToolCategories returns categories for a tool
func (si *SemanticIndex) getToolCategories(toolName string) []string {
	si.mutex.RLock()
	defer si.mutex.RUnlock()
	
	for key, categories := range si.categoryIndex {
		if strings.HasSuffix(key, "/"+toolName) {
			return categories
		}
	}
	
	return []string{"general"}
}


// NewRankingEngine creates a new ranking engine
func NewRankingEngine() *RankingEngine {
	return &RankingEngine{
		rankingFactors: map[string]float64{
			"performance":  0.25,
			"reliability":  0.25,
			"latency":      0.20,
			"cost":         0.15,
			"affinity":     0.15,
		},
		userPreferences: make(map[string]UserPreferences),
		mutex:           sync.RWMutex{},
	}
}

// RankTools ranks discovered tools based on multiple criteria
func (re *RankingEngine) RankTools(tools []protocol.DiscoveredTool, context *RequestContext) []RankedTool {
	re.mutex.RLock()
	defer re.mutex.RUnlock()
	
	rankedTools := make([]RankedTool, 0, len(tools))
	
	for _, tool := range tools {
		for range tool.MCPTools {
			rankedTool := RankedTool{
				Tool: tool,
				RankingFactors: make(map[string]float64),
			}
			
			// Calculate individual scores
			rankedTool.PerformanceScore = re.calculatePerformanceScore(tool)
			rankedTool.ReliabilityScore = re.calculateReliabilityScore(tool)
			rankedTool.LatencyScore = re.calculateLatencyScore(tool)
			rankedTool.CostScore = re.calculateCostScore(tool)
			rankedTool.AffinityScore = re.calculateAffinityScore(tool, context)
			
			// Calculate overall score
			rankedTool.OverallScore = re.calculateOverallScore(rankedTool, context)
			
			// Store individual factor contributions
			rankedTool.RankingFactors["performance"] = rankedTool.PerformanceScore
			rankedTool.RankingFactors["reliability"] = rankedTool.ReliabilityScore
			rankedTool.RankingFactors["latency"] = rankedTool.LatencyScore
			rankedTool.RankingFactors["cost"] = rankedTool.CostScore
			rankedTool.RankingFactors["affinity"] = rankedTool.AffinityScore
			
			rankedTools = append(rankedTools, rankedTool)
		}
	}
	
	// Sort by overall score
	sort.Slice(rankedTools, func(i, j int) bool {
		return rankedTools[i].OverallScore > rankedTools[j].OverallScore
	})
	
	return rankedTools
}

// calculatePerformanceScore calculates performance score for a tool
func (re *RankingEngine) calculatePerformanceScore(tool protocol.DiscoveredTool) float64 {
	// Base score on metadata
	if tool.Metadata.AverageResponseTime <= 0 {
		return 0.5 // Unknown performance
	}
	
	// Normalize response time (assume 1 second is excellent, 10 seconds is poor)
	responseTimeScore := math.Max(0, 1.0-(float64(tool.Metadata.AverageResponseTime)/10000.0))
	
	// Factor in other performance indicators
	// This could include throughput, resource usage, etc.
	
	return math.Min(1.0, responseTimeScore)
}

// calculateReliabilityScore calculates reliability score for a tool
func (re *RankingEngine) calculateReliabilityScore(tool protocol.DiscoveredTool) float64 {
	// Use trust score from metadata
	trustScore := tool.Metadata.TrustScore
	
	// Factor in uptime/availability if available
	// This could be enhanced with error rates, failure patterns, etc.
	
	return math.Max(0, math.Min(1, trustScore))
}

// calculateLatencyScore calculates latency score for a tool
func (re *RankingEngine) calculateLatencyScore(tool protocol.DiscoveredTool) float64 {
	responseTime := tool.Metadata.AverageResponseTime
	if responseTime <= 0 {
		return 0.5 // Unknown latency
	}
	
	// Score based on response time (lower is better)
	// 100ms = excellent, 1000ms = good, 5000ms = poor
	if responseTime <= 100 {
		return 1.0
	} else if responseTime <= 1000 {
		return 0.8
	} else if responseTime <= 5000 {
		return 0.6
	} else {
		return 0.2
	}
}

// calculateCostScore calculates cost score for a tool
func (re *RankingEngine) calculateCostScore(tool protocol.DiscoveredTool) float64 {
	// This would typically consider:
	// - Computational cost of running the tool
	// - Network/bandwidth costs
	// - Any usage fees or licensing
	// - Resource consumption
	
	// For now, return a default score
	// Local tools might be cheaper than remote ones
	if strings.Contains(tool.MCPEndpoint, "localhost") || strings.Contains(tool.MCPEndpoint, "127.0.0.1") {
		return 0.9 // Local tools are usually cheaper
	}
	
	return 0.7 // Default for remote tools
}

// calculateAffinityScore calculates affinity score based on user preferences
func (re *RankingEngine) calculateAffinityScore(tool protocol.DiscoveredTool, context *RequestContext) float64 {
	if context == nil {
		return 0.5
	}
	
	score := 0.0
	
	// Check if agent is in preferred list
	for _, preferred := range context.AffinityPreferences {
		if tool.AgentID == preferred {
			score += 0.4
			break
		}
	}
	
	// Check geographic affinity
	if context.GeographicRegion != "" && tool.EnvironmentType == context.GeographicRegion {
		score += 0.3
	}
	
	// Tool specialization bonus
	// This could check if the tool is frequently used for similar tasks
	score += 0.3
	
	return math.Min(1.0, score)
}

// calculateOverallScore combines all factors into an overall score
func (re *RankingEngine) calculateOverallScore(rankedTool RankedTool, context *RequestContext) float64 {
	weights := re.rankingFactors
	
	// Adjust weights based on user preferences if available
	if context != nil {
		if prefs, exists := re.userPreferences[context.RequesterID]; exists {
			weights = map[string]float64{
				"performance":  prefs.PerformanceWeight,
				"reliability":  prefs.ReliabilityWeight,
				"latency":      prefs.LatencyWeight,
				"cost":         prefs.CostWeight,
				"affinity":     prefs.LatencyWeight, // Using latency weight for affinity
			}
		}
	}
	
	// Adjust weights based on request priority
	if context != nil {
		switch context.Priority {
		case PriorityCritical:
			weights["reliability"] *= 1.5
			weights["performance"] *= 1.3
		case PriorityHigh:
			weights["latency"] *= 1.3
			weights["performance"] *= 1.2
		case PriorityLow:
			weights["cost"] *= 1.5
		}
	}
	
	// Normalize weights
	totalWeight := 0.0
	for _, weight := range weights {
		totalWeight += weight
	}
	
	if totalWeight == 0 {
		return 0.5
	}
	
	for factor := range weights {
		weights[factor] /= totalWeight
	}
	
	// Calculate weighted score
	score := rankedTool.PerformanceScore*weights["performance"] +
		rankedTool.ReliabilityScore*weights["reliability"] +
		rankedTool.LatencyScore*weights["latency"] +
		rankedTool.CostScore*weights["cost"] +
		rankedTool.AffinityScore*weights["affinity"]
	
	return math.Max(0, math.Min(1, score))
}

// SetUserPreferences sets ranking preferences for a user
func (re *RankingEngine) SetUserPreferences(userID string, preferences UserPreferences) {
	re.mutex.Lock()
	defer re.mutex.Unlock()
	
	re.userPreferences[userID] = preferences
}

// GetUserPreferences gets ranking preferences for a user
func (re *RankingEngine) GetUserPreferences(userID string) (UserPreferences, bool) {
	re.mutex.RLock()
	defer re.mutex.RUnlock()
	
	prefs, exists := re.userPreferences[userID]
	return prefs, exists
}

// UpdateRankingFactors updates the global ranking factor weights
func (re *RankingEngine) UpdateRankingFactors(factors map[string]float64) {
	re.mutex.Lock()
	defer re.mutex.Unlock()
	
	// Validate and normalize weights
	totalWeight := 0.0
	for _, weight := range factors {
		if weight < 0 {
			return // Invalid weights
		}
		totalWeight += weight
	}
	
	if totalWeight == 0 {
		return
	}
	
	// Normalize and update
	for factor, weight := range factors {
		re.rankingFactors[factor] = weight / totalWeight
	}
}