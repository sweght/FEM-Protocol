package main

import (
	"bytes"
	"crypto/ed25519"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/fep-fem/protocol"
)

// MCPClient provides high-level interface for discovering and using MCP tools
type MCPClient struct {
	agentID     string
	brokerURL   string
	privateKey  ed25519.PrivateKey
	httpClient  *http.Client
	
	// Tool discovery cache
	toolCache   map[string]*CachedToolResult
	cacheMutex  sync.RWMutex
	cacheExpiry time.Duration
	
	// Request management
	requestID   int64
	requestMutex sync.Mutex
}

// CachedToolResult stores discovered tools with expiration
type CachedToolResult struct {
	Tools      []protocol.DiscoveredTool
	Timestamp  time.Time
	RequestKey string
}

// MCPClientConfig holds configuration for the MCP client
type MCPClientConfig struct {
	AgentID        string
	BrokerURL      string
	PrivateKey     ed25519.PrivateKey
	CacheExpiry    time.Duration
	RequestTimeout time.Duration
	TLSInsecure    bool
}

// NewMCPClient creates a new MCP client instance
func NewMCPClient(config MCPClientConfig) *MCPClient {
	if config.CacheExpiry == 0 {
		config.CacheExpiry = 5 * time.Minute
	}
	if config.RequestTimeout == 0 {
		config.RequestTimeout = 30 * time.Second
	}

	transport := &http.Transport{}
	if config.TLSInsecure {
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	return &MCPClient{
		agentID:     config.AgentID,
		brokerURL:   config.BrokerURL,
		privateKey:  config.PrivateKey,
		toolCache:   make(map[string]*CachedToolResult),
		cacheExpiry: config.CacheExpiry,
		httpClient: &http.Client{
			Transport: transport,
			Timeout:   config.RequestTimeout,
		},
	}
}

// DiscoverTools searches for tools matching the given query
func (c *MCPClient) DiscoverTools(query protocol.ToolQuery) ([]protocol.DiscoveredTool, error) {
	// Check cache first
	cacheKey := c.buildCacheKey(query)
	if cached := c.getCachedResult(cacheKey); cached != nil {
		return cached.Tools, nil
	}

	// Generate request ID
	requestID := c.generateRequestID()

	// Create discovery envelope
	envelope := &protocol.DiscoverToolsEnvelope{
		BaseEnvelope: protocol.BaseEnvelope{
			Type: protocol.EnvelopeDiscoverTools,
			CommonHeaders: protocol.CommonHeaders{
				Agent: c.agentID,
				TS:    time.Now().UnixMilli(),
				Nonce: c.generateNonce(),
			},
		},
		Body: protocol.DiscoverToolsBody{
			Query:     query,
			RequestID: requestID,
		},
	}

	// Sign the envelope
	if err := envelope.Sign(c.privateKey); err != nil {
		return nil, fmt.Errorf("failed to sign discovery request: %w", err)
	}

	// Send request to broker
	response, err := c.sendRequest(envelope)
	if err != nil {
		return nil, fmt.Errorf("failed to send discovery request: %w", err)
	}

	// Parse tools from response
	tools, ok := response["tools"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid response format: missing tools array")
	}

	// Convert to DiscoveredTool structs
	discoveredTools := make([]protocol.DiscoveredTool, 0, len(tools))
	for _, toolData := range tools {
		toolMap, ok := toolData.(map[string]interface{})
		if !ok {
			continue
		}

		// Convert to DiscoveredTool
		toolBytes, _ := json.Marshal(toolMap)
		var discoveredTool protocol.DiscoveredTool
		if err := json.Unmarshal(toolBytes, &discoveredTool); err != nil {
			continue
		}
		discoveredTools = append(discoveredTools, discoveredTool)
	}

	// Cache the result
	c.cacheResult(cacheKey, discoveredTools)

	return discoveredTools, nil
}

// FindToolsByCapability is a convenience method for finding tools by capability pattern
func (c *MCPClient) FindToolsByCapability(capabilities []string) ([]protocol.DiscoveredTool, error) {
	query := protocol.ToolQuery{
		Capabilities:    capabilities,
		IncludeMetadata: true,
	}
	return c.DiscoverTools(query)
}

// FindToolsInEnvironment searches for tools in a specific environment
func (c *MCPClient) FindToolsInEnvironment(environmentType string, maxResults int) ([]protocol.DiscoveredTool, error) {
	query := protocol.ToolQuery{
		Capabilities:    []string{"*"},
		EnvironmentType: environmentType,
		MaxResults:      maxResults,
		IncludeMetadata: true,
	}
	return c.DiscoverTools(query)
}

// CallTool invokes a specific MCP tool through its agent
func (c *MCPClient) CallTool(agentID, toolName string, parameters map[string]interface{}) (interface{}, error) {
	requestID := c.generateRequestID()

	// Create tool call envelope
	envelope := &protocol.ToolCallEnvelope{
		BaseEnvelope: protocol.BaseEnvelope{
			Type: protocol.EnvelopeToolCall,
			CommonHeaders: protocol.CommonHeaders{
				Agent: c.agentID,
				TS:    time.Now().UnixMilli(),
				Nonce: c.generateNonce(),
			},
		},
		Body: protocol.ToolCallBody{
			Tool:       fmt.Sprintf("%s/%s", agentID, toolName),
			Parameters: parameters,
			RequestID:  requestID,
		},
	}

	// Sign the envelope
	if err := envelope.Sign(c.privateKey); err != nil {
		return nil, fmt.Errorf("failed to sign tool call: %w", err)
	}

	// Send request to broker
	response, err := c.sendRequest(envelope)
	if err != nil {
		return nil, fmt.Errorf("failed to send tool call: %w", err)
	}

	// Check for success
	if status, ok := response["status"].(string); ok && status == "processing" {
		// In a real implementation, this would poll for results or use webhooks
		return response, nil
	}

	return nil, fmt.Errorf("tool call failed: %v", response)
}

// GetAvailableAgents returns a list of all agents that have MCP tools
func (c *MCPClient) GetAvailableAgents() ([]protocol.DiscoveredTool, error) {
	return c.FindToolsByCapability([]string{"*"})
}

// RefreshCache clears the tool discovery cache
func (c *MCPClient) RefreshCache() {
	c.cacheMutex.Lock()
	defer c.cacheMutex.Unlock()
	c.toolCache = make(map[string]*CachedToolResult)
}

// sendRequest sends an envelope to the broker and returns the response
func (c *MCPClient) sendRequest(envelope interface{}) (map[string]interface{}, error) {
	// Marshal envelope
	data, err := json.Marshal(envelope)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Send HTTP POST request
	resp, err := c.httpClient.Post(c.brokerURL, "application/json", bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to send HTTP request: %w", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("broker returned status %d", resp.StatusCode)
	}

	// Parse response
	var response map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return response, nil
}

// Cache management methods

func (c *MCPClient) buildCacheKey(query protocol.ToolQuery) string {
	// Create a simple cache key from query parameters
	key := fmt.Sprintf("env:%s,caps:%v,max:%d", 
		query.EnvironmentType, 
		query.Capabilities, 
		query.MaxResults)
	return key
}

func (c *MCPClient) getCachedResult(key string) *CachedToolResult {
	c.cacheMutex.RLock()
	defer c.cacheMutex.RUnlock()

	cached, exists := c.toolCache[key]
	if !exists {
		return nil
	}

	// Check if cache has expired
	if time.Since(cached.Timestamp) > c.cacheExpiry {
		delete(c.toolCache, key)
		return nil
	}

	return cached
}

func (c *MCPClient) cacheResult(key string, tools []protocol.DiscoveredTool) {
	c.cacheMutex.Lock()
	defer c.cacheMutex.Unlock()

	c.toolCache[key] = &CachedToolResult{
		Tools:      tools,
		Timestamp:  time.Now(),
		RequestKey: key,
	}
}

// Request ID generation
func (c *MCPClient) generateRequestID() string {
	c.requestMutex.Lock()
	defer c.requestMutex.Unlock()
	c.requestID++
	return fmt.Sprintf("%s-req-%d", c.agentID, c.requestID)
}

func (c *MCPClient) generateNonce() string {
	return fmt.Sprintf("%s-%d", c.agentID, time.Now().UnixNano())
}

// GetCacheStats returns statistics about the tool cache
func (c *MCPClient) GetCacheStats() map[string]interface{} {
	c.cacheMutex.RLock()
	defer c.cacheMutex.RUnlock()

	stats := map[string]interface{}{
		"cached_queries": len(c.toolCache),
		"cache_expiry":   c.cacheExpiry.String(),
	}

	totalTools := 0
	for _, cached := range c.toolCache {
		totalTools += len(cached.Tools)
	}
	stats["total_cached_tools"] = totalTools

	return stats
}