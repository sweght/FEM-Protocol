# FEP-FEM Implementation Roadmap

## Overview

This document provides a comprehensive technical roadmap for implementing the MCP federation and embodiment features described in the documentation. The roadmap is organized into distinct implementation phases, each representing a logical unit of work that can be completed in a single focused development session.

## Current State Analysis

### What's Implemented (v0.1.2)
- ✅ **Core FEP Protocol**: 7 base envelope types with Ed25519 signatures
- ✅ **Basic Broker**: Agent registration, message routing, TLS support
- ✅ **Basic Agent Framework**: Registration, capability declaration, tool execution
- ✅ **Cryptographic Security**: Ed25519 signing/verification, replay protection
- ✅ **Cross-platform Builds**: Linux, macOS, Windows support

### What's Missing for MCP Integration
- ❌ **MCP Server/Client Integration**: No MCP protocol support in agents
- ❌ **Tool Discovery System**: No broker-level MCP tool registry
- ❌ **Embodiment Framework**: No environment-specific tool adaptation
- ❌ **New Envelope Types**: `discoverTools`, `toolsDiscovered`, `embodimentUpdate`
- ❌ **Body Definitions**: No structured environment templates
- ❌ **Federation Features**: No cross-broker MCP tool sharing

## Implementation Phases

### Phase A: Protocol Foundation - New Envelope Types
**Objective**: Extend FEP protocol with MCP integration envelope types
**Scope**: Single focused addition to protocol layer
**Dependencies**: None

#### A.1: Add MCP Discovery Envelopes
**Files Modified**: `protocol/go/envelopes.go`

Add three new envelope types documented in Protocol-Specification.md:

```go
const (
    // Existing types...
    EnvelopeDiscoverTools     EnvelopeType = "discoverTools"
    EnvelopeToolsDiscovered   EnvelopeType = "toolsDiscovered" 
    EnvelopeEmbodimentUpdate  EnvelopeType = "embodimentUpdate"
)

// DiscoverToolsEnvelope requests MCP tool discovery
type DiscoverToolsEnvelope struct {
    BaseEnvelope
    Body DiscoverToolsBody `json:"body"`
}

type DiscoverToolsBody struct {
    Query     ToolQuery `json:"query"`
    RequestID string    `json:"requestId"`
}

type ToolQuery struct {
    Capabilities    []string `json:"capabilities"`
    EnvironmentType string   `json:"environmentType,omitempty"`
    MaxResults      int      `json:"maxResults,omitempty"`
    IncludeMetadata bool     `json:"includeMetadata,omitempty"`
}

// ToolsDiscoveredEnvelope returns discovered MCP tools
type ToolsDiscoveredEnvelope struct {
    BaseEnvelope
    Body ToolsDiscoveredBody `json:"body"`
}

type ToolsDiscoveredBody struct {
    RequestID    string           `json:"requestId"`
    Tools        []DiscoveredTool `json:"tools"`
    TotalResults int              `json:"totalResults"`
    HasMore      bool             `json:"hasMore"`
}

type DiscoveredTool struct {
    AgentID         string       `json:"agentId"`
    MCPEndpoint     string       `json:"mcpEndpoint"`
    Capabilities    []string     `json:"capabilities"`
    EnvironmentType string       `json:"environmentType"`
    MCPTools        []MCPTool    `json:"mcpTools"`
    Metadata        ToolMetadata `json:"metadata,omitempty"`
}

// EmbodimentUpdateEnvelope notifies of environment changes
type EmbodimentUpdateEnvelope struct {
    BaseEnvelope
    Body EmbodimentUpdateBody `json:"body"`
}

type EmbodimentUpdateBody struct {
    EnvironmentType string         `json:"environmentType"`
    BodyDefinition  BodyDefinition `json:"bodyDefinition"`
    MCPEndpoint     string         `json:"mcpEndpoint"`
    UpdatedTools    []string       `json:"updatedTools"`
}
```

#### A.2: Add Supporting Types
**Files Modified**: `protocol/go/envelopes.go`

```go
type MCPTool struct {
    Name        string                 `json:"name"`
    Description string                 `json:"description"`
    InputSchema map[string]interface{} `json:"inputSchema"`
}

type ToolMetadata struct {
    LastSeen            int64   `json:"lastSeen"`
    AverageResponseTime int     `json:"averageResponseTime"`
    TrustScore          float64 `json:"trustScore"`
}

type BodyDefinition struct {
    Name         string                 `json:"name"`
    Environment  string                 `json:"environment"`
    Capabilities []string               `json:"capabilities"`
    MCPTools     []MCPTool             `json:"mcpTools"`
    Constraints  map[string]interface{} `json:"constraints,omitempty"`
    Metadata     map[string]interface{} `json:"metadata,omitempty"`
}
```

#### A.3: Update RegisterAgent for MCP
**Files Modified**: `protocol/go/envelopes.go`

```go
type RegisterAgentBody struct {
    PubKey          string                 `json:"pubkey"`
    Capabilities    []string               `json:"capabilities"`
    Metadata        map[string]interface{} `json:"metadata,omitempty"`
    // New MCP integration fields
    MCPEndpoint     string                 `json:"mcpEndpoint,omitempty"`
    BodyDefinition  *BodyDefinition        `json:"bodyDefinition,omitempty"`
    EnvironmentType string                 `json:"environmentType,omitempty"`
}
```

#### A.4: Add Signing Methods
**Files Modified**: `protocol/go/envelopes.go`

```go
func (e *DiscoverToolsEnvelope) Sign(privateKey ed25519.PrivateKey) error {
    e.Sig = ""
    data, err := json.Marshal(e)
    if err != nil {
        return err
    }
    signature := ed25519.Sign(privateKey, data)
    e.Sig = base64.StdEncoding.EncodeToString(signature)
    return nil
}

func (e *ToolsDiscoveredEnvelope) Sign(privateKey ed25519.PrivateKey) error {
    e.Sig = ""
    data, err := json.Marshal(e)
    if err != nil {
        return err
    }
    signature := ed25519.Sign(privateKey, data)
    e.Sig = base64.StdEncoding.EncodeToString(signature)
    return nil
}

func (e *EmbodimentUpdateEnvelope) Sign(privateKey ed25519.PrivateKey) error {
    e.Sig = ""
    data, err := json.Marshal(e)
    if err != nil {
        return err
    }
    signature := ed25519.Sign(privateKey, data)
    e.Sig = base64.StdEncoding.EncodeToString(signature)
    return nil
}
```

**Completion Criteria**:
- All new envelope types compile without errors
- New envelope types can be marshaled/unmarshaled to JSON
- Signing and verification works for new envelope types

### Phase B: Protocol Testing
**Objective**: Add comprehensive tests for new envelope types
**Scope**: Test coverage for Phase A changes
**Dependencies**: Phase A

**Files Modified**: `protocol/go/envelopes_test.go`

```go
func TestDiscoverToolsEnvelope(t *testing.T) {
    pubKey, privKey, err := GenerateKeyPair()
    require.NoError(t, err)
    
    envelope := &DiscoverToolsEnvelope{
        BaseEnvelope: BaseEnvelope{
            Type: EnvelopeDiscoverTools,
            CommonHeaders: CommonHeaders{
                Agent: "test-agent",
                TS:    time.Now().UnixMilli(),
                Nonce: "test-nonce",
            },
        },
        Body: DiscoverToolsBody{
            Query: ToolQuery{
                Capabilities: []string{"file.*", "code.execute"},
                MaxResults:   10,
            },
            RequestID: "test-request",
        },
    }
    
    // Test signing
    err = envelope.Sign(privKey)
    require.NoError(t, err)
    require.NotEmpty(t, envelope.Sig)
    
    // Test JSON marshaling
    data, err := json.Marshal(envelope)
    require.NoError(t, err)
    
    // Test JSON unmarshaling
    var unmarshaled DiscoverToolsEnvelope
    err = json.Unmarshal(data, &unmarshaled)
    require.NoError(t, err)
    require.Equal(t, envelope.Body.RequestID, unmarshaled.Body.RequestID)
}

func TestToolsDiscoveredEnvelope(t *testing.T) {
    // Similar comprehensive test for ToolsDiscoveredEnvelope
}

func TestEmbodimentUpdateEnvelope(t *testing.T) {
    // Similar comprehensive test for EmbodimentUpdateEnvelope
}

func TestBodyDefinition(t *testing.T) {
    bodyDef := BodyDefinition{
        Name:        "cloud-worker",
        Environment: "cloud",
        Capabilities: []string{"s3.read", "lambda.invoke"},
        MCPTools: []MCPTool{
            {
                Name:        "s3.read",
                Description: "Read from S3 bucket",
                InputSchema: map[string]interface{}{
                    "type": "object",
                    "properties": map[string]interface{}{
                        "bucket": map[string]interface{}{"type": "string"},
                        "key":    map[string]interface{}{"type": "string"},
                    },
                },
            },
        },
    }
    
    // Test JSON marshaling/unmarshaling
    data, err := json.Marshal(bodyDef)
    require.NoError(t, err)
    
    var unmarshaled BodyDefinition
    err = json.Unmarshal(data, &unmarshaled)
    require.NoError(t, err)
    require.Equal(t, bodyDef.Name, unmarshaled.Name)
    require.Len(t, unmarshaled.MCPTools, 1)
}
```

**Completion Criteria**:
- All tests pass
- Test coverage >90% for new code
- Tests validate JSON serialization and envelope signing

### Phase C: Broker MCP Registry Core
**Objective**: Add basic MCP tool registry to broker
**Scope**: Core data structures and in-memory storage
**Dependencies**: Phases A, B

#### C.1: Create MCP Registry
**Files Created**: `broker/mcp_registry.go`

```go
package main

import (
    "fmt"
    "sync"
    "time"
    "github.com/fep-fem/protocol"
)

type MCPRegistry struct {
    tools  map[string]*RegisteredTool
    agents map[string]*MCPAgent
    mu     sync.RWMutex
}

type RegisteredTool struct {
    AgentID         string
    Tool            protocol.MCPTool
    MCPEndpoint     string
    EnvironmentType string
    RegisteredAt    time.Time
    LastSeen        time.Time
}

type MCPAgent struct {
    ID              string
    MCPEndpoint     string
    BodyDefinition  *protocol.BodyDefinition
    EnvironmentType string
    Tools           []protocol.MCPTool
    LastHeartbeat   time.Time
}

func NewMCPRegistry() *MCPRegistry {
    return &MCPRegistry{
        tools:  make(map[string]*RegisteredTool),
        agents: make(map[string]*MCPAgent),
    }
}

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

func (r *MCPRegistry) GetAgent(agentID string) (*MCPAgent, bool) {
    r.mu.RLock()
    defer r.mu.RUnlock()
    agent, exists := r.agents[agentID]
    return agent, exists
}

func (r *MCPRegistry) ListTools() []*RegisteredTool {
    r.mu.RLock()
    defer r.mu.RUnlock()
    
    tools := make([]*RegisteredTool, 0, len(r.tools))
    for _, tool := range r.tools {
        tools = append(tools, tool)
    }
    return tools
}

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
```

#### C.2: Add Basic Tool Discovery
**Files Modified**: `broker/mcp_registry.go`

```go
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

func (r *MCPRegistry) extractCapabilities(tools []protocol.MCPTool) []string {
    capabilities := make([]string, 0, len(tools))
    for _, tool := range tools {
        capabilities = append(capabilities, tool.Name)
    }
    return capabilities
}
```

**Completion Criteria**:
- Registry can store and retrieve agent information
- Basic tool discovery works with simple pattern matching
- Registry handles agent registration and unregistration
- Thread-safe operations with proper locking

### Phase D: Broker Handler Integration
**Objective**: Integrate MCP registry with broker HTTP handlers
**Scope**: Connect registry to broker request processing
**Dependencies**: Phase C

**Files Modified**: `broker/main.go`

```go
// Add MCP registry to Broker struct
type Broker struct {
    agents      map[string]*Agent
    mu          sync.RWMutex
    tlsConfig   *tls.Config
    mcpRegistry *MCPRegistry  // New field
}

func NewBroker() *Broker {
    return &Broker{
        agents:      make(map[string]*Agent),
        mcpRegistry: NewMCPRegistry(),  // Initialize registry
    }
}

// Update ServeHTTP to handle new envelope types
func (b *Broker) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    // Existing health check and method validation...
    
    envelope, err := protocol.ParseEnvelope(body)
    if err != nil {
        http.Error(w, fmt.Sprintf("Invalid envelope: %v", err), http.StatusBadRequest)
        return
    }

    log.Printf("Received %s envelope from %s", envelope.Type, envelope.Agent)

    switch envelope.Type {
    case protocol.EnvelopeRegisterAgent:
        b.handleRegisterAgent(w, envelope)
    case protocol.EnvelopeDiscoverTools:        // New handler
        b.handleDiscoverTools(w, envelope)
    case protocol.EnvelopeEmbodimentUpdate:     // New handler
        b.handleEmbodimentUpdate(w, envelope)
    // ... existing cases
    default:
        http.Error(w, "Unknown envelope type", http.StatusBadRequest)
        return
    }
}

// Enhanced handleRegisterAgent to support MCP
func (b *Broker) handleRegisterAgent(w http.ResponseWriter, env *protocol.GenericEnvelope) {
    var body protocol.RegisterAgentBody
    if err := env.GetBodyAs(&body); err != nil {
        http.Error(w, "Invalid body", http.StatusBadRequest)
        return
    }

    // Existing agent registration
    b.mu.Lock()
    b.agents[env.Agent] = &Agent{
        ID:           env.Agent,
        Capabilities: body.Capabilities,
        RegisteredAt: time.Now(),
    }
    b.mu.Unlock()

    // New MCP registration if MCP endpoint provided
    if body.MCPEndpoint != "" {
        mcpAgent := &MCPAgent{
            ID:              env.Agent,
            MCPEndpoint:     body.MCPEndpoint,
            BodyDefinition:  body.BodyDefinition,
            EnvironmentType: body.EnvironmentType,
            LastHeartbeat:   time.Now(),
        }
        
        // Extract MCP tools from body definition
        if body.BodyDefinition != nil {
            mcpAgent.Tools = body.BodyDefinition.MCPTools
        }
        
        if err := b.mcpRegistry.RegisterAgent(env.Agent, mcpAgent); err != nil {
            log.Printf("Failed to register MCP agent: %v", err)
        } else {
            log.Printf("Registered MCP agent %s with endpoint %s", env.Agent, body.MCPEndpoint)
        }
    }

    log.Printf("Registered agent %s with capabilities %v", env.Agent, body.Capabilities)

    response := map[string]interface{}{
        "status": "registered",
        "agent":  env.Agent,
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
}

// New handler for tool discovery
func (b *Broker) handleDiscoverTools(w http.ResponseWriter, env *protocol.GenericEnvelope) {
    var discoverBody protocol.DiscoverToolsBody
    if err := env.GetBodyAs(&discoverBody); err != nil {
        http.Error(w, "Invalid discovery request", http.StatusBadRequest)
        return
    }
    
    log.Printf("Tool discovery request from %s: %+v", env.Agent, discoverBody.Query)
    
    discoveredTools, err := b.mcpRegistry.DiscoverTools(discoverBody.Query)
    if err != nil {
        http.Error(w, "Discovery failed", http.StatusInternalServerError)
        return
    }
    
    log.Printf("Found %d tools matching query", len(discoveredTools))
    
    response := map[string]interface{}{
        "status":       "success",
        "requestId":    discoverBody.RequestID,
        "tools":        discoveredTools,
        "totalResults": len(discoveredTools),
        "hasMore":      false,
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
}

// New handler for embodiment updates
func (b *Broker) handleEmbodimentUpdate(w http.ResponseWriter, env *protocol.GenericEnvelope) {
    var updateBody protocol.EmbodimentUpdateBody
    if err := env.GetBodyAs(&updateBody); err != nil {
        http.Error(w, "Invalid embodiment update", http.StatusBadRequest)
        return
    }
    
    log.Printf("Embodiment update from %s: environment=%s", env.Agent, updateBody.EnvironmentType)
    
    // Update MCP registry with new embodiment
    if agent, exists := b.mcpRegistry.GetAgent(env.Agent); exists {
        agent.EnvironmentType = updateBody.EnvironmentType
        agent.BodyDefinition = &updateBody.BodyDefinition
        agent.MCPEndpoint = updateBody.MCPEndpoint
        agent.Tools = updateBody.BodyDefinition.MCPTools
        agent.LastHeartbeat = time.Now()
        
        // Re-register to update tool index
        b.mcpRegistry.RegisterAgent(env.Agent, agent)
        
        log.Printf("Updated embodiment for agent %s", env.Agent)
    }
    
    response := map[string]interface{}{
        "status": "updated",
        "agent":  env.Agent,
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
}
```

**Completion Criteria**:
- Broker can handle new MCP envelope types
- Agent registration includes MCP endpoint and body definition
- Tool discovery returns properly formatted responses
- Embodiment updates modify registry state correctly

### Phase E: Basic MCP Client Library
**Objective**: Create MCP client for agents to consume federated tools
**Scope**: Standalone MCP client that works with FEM discovery
**Dependencies**: Phase D

**Files Created**: `protocol/go/mcp_client.go`

```go
package protocol

import (
    "bytes"
    "encoding/json"
    "fmt"
    "net/http"
    "time"
)

type MCPClient struct {
    httpClient *http.Client
    endpoints  map[string]string // agent ID -> MCP endpoint
    brokerURL  string
}

func NewMCPClient(brokerURL string) *MCPClient {
    return &MCPClient{
        httpClient: &http.Client{
            Timeout: 30 * time.Second,
        },
        endpoints: make(map[string]string),
        brokerURL: brokerURL,
    }
}

// DiscoverTools finds MCP tools matching the query
func (c *MCPClient) DiscoverTools(query ToolQuery, requesterID string) ([]DiscoveredTool, error) {
    // Create discovery request envelope
    envelope := &DiscoverToolsEnvelope{
        BaseEnvelope: BaseEnvelope{
            Type: EnvelopeDiscoverTools,
            CommonHeaders: CommonHeaders{
                Agent: requesterID,
                TS:    time.Now().UnixMilli(),
                Nonce: fmt.Sprintf("discover-%d", time.Now().UnixNano()),
            },
        },
        Body: DiscoverToolsBody{
            Query:     query,
            RequestID: fmt.Sprintf("req-%d", time.Now().UnixNano()),
        },
    }
    
    // Marshal and send to broker
    data, err := json.Marshal(envelope)
    if err != nil {
        return nil, fmt.Errorf("failed to marshal discovery request: %w", err)
    }
    
    resp, err := c.httpClient.Post(c.brokerURL, "application/json", bytes.NewReader(data))
    if err != nil {
        return nil, fmt.Errorf("failed to send discovery request: %w", err)
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("broker returned status %d", resp.StatusCode)
    }
    
    var response struct {
        Tools []DiscoveredTool `json:"tools"`
    }
    
    if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
        return nil, fmt.Errorf("failed to decode discovery response: %w", err)
    }
    
    // Cache endpoints for direct MCP calls
    for _, tool := range response.Tools {
        c.endpoints[tool.AgentID] = tool.MCPEndpoint
    }
    
    return response.Tools, nil
}

// CallTool executes an MCP tool on a remote agent
func (c *MCPClient) CallTool(agentID, toolName string, params map[string]interface{}) (interface{}, error) {
    endpoint, exists := c.endpoints[agentID]
    if !exists {
        return nil, fmt.Errorf("no MCP endpoint known for agent %s", agentID)
    }
    
    // Create MCP tool call request
    mcpRequest := map[string]interface{}{
        "method": "tools/call",
        "params": map[string]interface{}{
            "name":      toolName,
            "arguments": params,
        },
        "id": fmt.Sprintf("call-%d", time.Now().UnixNano()),
    }
    
    jsonData, err := json.Marshal(mcpRequest)
    if err != nil {
        return nil, fmt.Errorf("failed to marshal MCP request: %w", err)
    }
    
    resp, err := c.httpClient.Post(endpoint, "application/json", bytes.NewReader(jsonData))
    if err != nil {
        return nil, fmt.Errorf("failed to call MCP tool: %w", err)
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("MCP server returned status %d", resp.StatusCode)
    }
    
    var mcpResponse struct {
        Result interface{} `json:"result"`
        Error  interface{} `json:"error"`
    }
    
    if err := json.NewDecoder(resp.Body).Decode(&mcpResponse); err != nil {
        return nil, fmt.Errorf("failed to decode MCP response: %w", err)
    }
    
    if mcpResponse.Error != nil {
        return nil, fmt.Errorf("MCP tool error: %v", mcpResponse.Error)
    }
    
    return mcpResponse.Result, nil
}

// FindTool discovers and returns the first matching tool
func (c *MCPClient) FindTool(capability, requesterID string) (*DiscoveredTool, error) {
    tools, err := c.DiscoverTools(ToolQuery{
        Capabilities: []string{capability},
        MaxResults:   1,
    }, requesterID)
    
    if err != nil {
        return nil, err
    }
    
    if len(tools) == 0 {
        return nil, fmt.Errorf("no tools found for capability: %s", capability)
    }
    
    return &tools[0], nil
}
```

**Completion Criteria**:
- MCP client can discover tools via FEM broker
- MCP client can call tools on remote agents
- Error handling for network and MCP protocol errors
- Endpoint caching for efficiency

### Phase F: Basic MCP Server Library
**Objective**: Create MCP server for agents to expose tools
**Scope**: Embeddable MCP server that integrates with FEM agents
**Dependencies**: Phase E

**Files Created**: `protocol/go/mcp_server.go`

```go
package protocol

import (
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "sync"
)

type MCPServer struct {
    tools     map[string]ToolHandler
    port      int
    server    *http.Server
    mu        sync.RWMutex
    isRunning bool
}

type ToolHandler func(params map[string]interface{}) (interface{}, error)

func NewMCPServer(port int) *MCPServer {
    return &MCPServer{
        tools: make(map[string]ToolHandler),
        port:  port,
    }
}

func (s *MCPServer) RegisterTool(name string, handler ToolHandler) {
    s.mu.Lock()
    defer s.mu.Unlock()
    s.tools[name] = handler
    log.Printf("Registered MCP tool: %s", name)
}

func (s *MCPServer) GetTools() []MCPTool {
    s.mu.RLock()
    defer s.mu.RUnlock()
    
    tools := make([]MCPTool, 0, len(s.tools))
    for name := range s.tools {
        tools = append(tools, MCPTool{
            Name:        name,
            Description: fmt.Sprintf("Tool: %s", name),
            InputSchema: map[string]interface{}{
                "type": "object",
                "properties": map[string]interface{}{},
            },
        })
    }
    return tools
}

func (s *MCPServer) Start() error {
    s.mu.Lock()
    defer s.mu.Unlock()
    
    if s.isRunning {
        return fmt.Errorf("server already running")
    }
    
    mux := http.NewServeMux()
    mux.HandleFunc("/", s.handleMCPRequest)
    mux.HandleFunc("/tools/list", s.handleToolsList)
    mux.HandleFunc("/tools/call", s.handleToolCall)
    
    s.server = &http.Server{
        Addr:    fmt.Sprintf(":%d", s.port),
        Handler: mux,
    }
    
    s.isRunning = true
    
    go func() {
        log.Printf("MCP server starting on port %d", s.port)
        if err := s.server.ListenAndServe(); err != http.ErrServerClosed {
            log.Printf("MCP server error: %v", err)
        }
    }()
    
    return nil
}

func (s *MCPServer) Stop() error {
    s.mu.Lock()
    defer s.mu.Unlock()
    
    if !s.isRunning || s.server == nil {
        return nil
    }
    
    s.isRunning = false
    return s.server.Close()
}

func (s *MCPServer) handleMCPRequest(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }
    
    var request struct {
        Method string                 `json:"method"`
        Params map[string]interface{} `json:"params"`
        ID     interface{}            `json:"id"`
    }
    
    if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
        http.Error(w, "Invalid JSON", http.StatusBadRequest)
        return
    }
    
    switch request.Method {
    case "tools/list":
        s.handleToolsListMCP(w, request.ID)
    case "tools/call":
        s.handleToolCallMCP(w, request.Params, request.ID)
    default:
        s.sendMCPError(w, fmt.Sprintf("Unknown method: %s", request.Method), request.ID)
    }
}

func (s *MCPServer) handleToolsList(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }
    
    tools := s.GetTools()
    
    response := map[string]interface{}{
        "tools": tools,
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
}

func (s *MCPServer) handleToolCall(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }
    
    var request struct {
        Name      string                 `json:"name"`
        Arguments map[string]interface{} `json:"arguments"`
    }
    
    if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
        http.Error(w, "Invalid request", http.StatusBadRequest)
        return
    }
    
    s.mu.RLock()
    handler, exists := s.tools[request.Name]
    s.mu.RUnlock()
    
    if !exists {
        http.Error(w, "Tool not found", http.StatusNotFound)
        return
    }
    
    result, err := handler(request.Arguments)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    
    response := map[string]interface{}{
        "result": result,
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
}

func (s *MCPServer) handleToolsListMCP(w http.ResponseWriter, id interface{}) {
    tools := s.GetTools()
    
    response := map[string]interface{}{
        "jsonrpc": "2.0",
        "result": map[string]interface{}{
            "tools": tools,
        },
        "id": id,
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
}

func (s *MCPServer) handleToolCallMCP(w http.ResponseWriter, params map[string]interface{}, id interface{}) {
    toolName, ok := params["name"].(string)
    if !ok {
        s.sendMCPError(w, "Missing tool name", id)
        return
    }
    
    arguments, ok := params["arguments"].(map[string]interface{})
    if !ok {
        arguments = make(map[string]interface{})
    }
    
    s.mu.RLock()
    handler, exists := s.tools[toolName]
    s.mu.RUnlock()
    
    if !exists {
        s.sendMCPError(w, "Tool not found", id)
        return
    }
    
    result, err := handler(arguments)
    if err != nil {
        s.sendMCPError(w, err.Error(), id)
        return
    }
    
    response := map[string]interface{}{
        "jsonrpc": "2.0",
        "result":  result,
        "id":      id,
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
}

func (s *MCPServer) sendMCPError(w http.ResponseWriter, message string, id interface{}) {
    response := map[string]interface{}{
        "jsonrpc": "2.0",
        "error": map[string]interface{}{
            "code":    -1,
            "message": message,
        },
        "id": id,
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
}

func (s *MCPServer) GetEndpoint() string {
    return fmt.Sprintf("http://localhost:%d", s.port)
}
```

**Completion Criteria**:
- MCP server exposes tools via standard MCP protocol
- Server supports both REST and JSON-RPC endpoints
- Tool registration and execution works correctly
- Server can be started/stopped programmatically

### Phase G: Agent MCP Integration
**Objective**: Update example agent to use MCP server/client
**Scope**: Modify fem-coder to demonstrate MCP federation
**Dependencies**: Phases E, F

**Files Modified**: `bodies/coder/cmd/fem-coder/main.go`

```go
// Add MCP imports and fields to Agent struct
import (
    // existing imports...
    "strconv"
)

type Agent struct {
    ID        string
    BrokerURL string
    PubKey    ed25519.PublicKey
    PrivKey   ed25519.PrivateKey
    client    *http.Client
    // New MCP integration fields
    mcpServer *protocol.MCPServer
    mcpClient *protocol.MCPClient
    mcpPort   int
}

func main() {
    brokerURL := flag.String("broker", "https://localhost:4433", "Broker URL to connect to")
    agentID := flag.String("agent", "fem-coder-001", "Agent identifier")
    mcpPort := flag.Int("mcp-port", 8080, "Port for MCP server")
    flag.Parse()

    log.Printf("fem-coder starting - Agent ID: %s, Broker: %s, MCP Port: %d", 
               *agentID, *brokerURL, *mcpPort)

    pubKey, privKey, err := protocol.GenerateKeyPair()
    if err != nil {
        log.Fatalf("Failed to generate key pair: %v", err)
    }

    agent := &Agent{
        ID:        *agentID,
        BrokerURL: *brokerURL,
        PubKey:    pubKey,
        PrivKey:   privKey,
        mcpPort:   *mcpPort,
        client: &http.Client{
            Transport: &http.Transport{
                TLSClientConfig: &tls.Config{
                    InsecureSkipVerify: true,
                },
            },
            Timeout: 10 * time.Second,
        },
    }

    // Initialize MCP components
    if err := agent.initializeMCP(); err != nil {
        log.Fatalf("Failed to initialize MCP: %v", err)
    }

    // Register with broker
    if err := agent.registerWithBroker(); err != nil {
        log.Fatalf("Failed to register with broker: %v", err)
    }

    log.Println("Successfully registered with broker. MCP server running. Waiting for tool calls...")

    // Keep the agent running
    select {}
}

func (a *Agent) initializeMCP() error {
    // Initialize MCP server
    a.mcpServer = protocol.NewMCPServer(a.mcpPort)
    
    // Register tools
    a.mcpServer.RegisterTool("code.execute", a.handleCodeExecution)
    a.mcpServer.RegisterTool("shell.run", a.handleShellExecution)
    a.mcpServer.RegisterTool("math.add", a.handleMathAdd)
    a.mcpServer.RegisterTool("math.multiply", a.handleMathMultiply)
    
    // Start MCP server
    if err := a.mcpServer.Start(); err != nil {
        return fmt.Errorf("failed to start MCP server: %w", err)
    }
    
    // Initialize MCP client
    a.mcpClient = protocol.NewMCPClient(a.BrokerURL)
    
    log.Printf("MCP server started on port %d", a.mcpPort)
    return nil
}

func (a *Agent) registerWithBroker() error {
    // Create body definition for current environment
    bodyDef := &protocol.BodyDefinition{
        Name:        "coder-body",
        Environment: "local",
        Capabilities: []string{"code.execute", "shell.run", "math.add", "math.multiply"},
        MCPTools:    a.mcpServer.GetTools(),
    }
    
    envelope := &protocol.RegisterAgentEnvelope{
        BaseEnvelope: protocol.BaseEnvelope{
            Type: protocol.EnvelopeRegisterAgent,
            CommonHeaders: protocol.CommonHeaders{
                Agent: a.ID,
                TS:    time.Now().UnixMilli(),
                Nonce: fmt.Sprintf("%d", time.Now().UnixNano()),
            },
        },
        Body: protocol.RegisterAgentBody{
            Capabilities:    []string{"code.execute", "shell.run", "math.add", "math.multiply"},
            PubKey:          protocol.EncodePublicKey(a.PubKey),
            MCPEndpoint:     a.mcpServer.GetEndpoint(),
            BodyDefinition:  bodyDef,
            EnvironmentType: "local",
        },
    }

    if err := envelope.Sign(a.PrivKey); err != nil {
        return fmt.Errorf("failed to sign envelope: %w", err)
    }

    data, err := json.Marshal(envelope)
    if err != nil {
        return fmt.Errorf("failed to marshal envelope: %w", err)
    }

    resp, err := a.client.Post(a.BrokerURL+"/fep", "application/json", bytes.NewReader(data))
    if err != nil {
        return fmt.Errorf("failed to send registration: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return fmt.Errorf("broker returned status %d", resp.StatusCode)
    }

    log.Printf("Registration successful - Agent %s registered with MCP endpoint %s", 
               a.ID, a.mcpServer.GetEndpoint())
    return nil
}

// MCP tool handlers
func (a *Agent) handleCodeExecution(params map[string]interface{}) (interface{}, error) {
    command, ok := params["command"].(string)
    if !ok {
        return nil, fmt.Errorf("missing 'command' parameter")
    }
    
    output, err := a.executeCode("sh", []string{"-c", command})
    if err != nil {
        return nil, err
    }
    
    return map[string]interface{}{
        "output": output,
        "status": "success",
    }, nil
}

func (a *Agent) handleShellExecution(params map[string]interface{}) (interface{}, error) {
    command, ok := params["command"].(string)
    if !ok {
        return nil, fmt.Errorf("missing 'command' parameter")
    }
    
    output, err := a.executeCode("sh", []string{"-c", command})
    if err != nil {
        return nil, err
    }
    
    return map[string]interface{}{
        "output": output,
        "status": "success",
    }, nil
}

func (a *Agent) handleMathAdd(params map[string]interface{}) (interface{}, error) {
    aVal, aOk := params["a"]
    bVal, bOk := params["b"]
    
    if !aOk || !bOk {
        return nil, fmt.Errorf("missing 'a' or 'b' parameter")
    }
    
    // Convert to float64 for calculation
    var a, b float64
    var err error
    
    switch v := aVal.(type) {
    case float64:
        a = v
    case int:
        a = float64(v)
    case string:
        a, err = strconv.ParseFloat(v, 64)
        if err != nil {
            return nil, fmt.Errorf("invalid number for 'a': %v", v)
        }
    default:
        return nil, fmt.Errorf("invalid type for 'a': %T", v)
    }
    
    switch v := bVal.(type) {
    case float64:
        b = v
    case int:
        b = float64(v)
    case string:
        b, err = strconv.ParseFloat(v, 64)
        if err != nil {
            return nil, fmt.Errorf("invalid number for 'b': %v", v)
        }
    default:
        return nil, fmt.Errorf("invalid type for 'b': %T", v)
    }
    
    result := a + b
    return map[string]interface{}{
        "result": result,
        "operation": "addition",
    }, nil
}

func (a *Agent) handleMathMultiply(params map[string]interface{}) (interface{}, error) {
    aVal, aOk := params["a"]
    bVal, bOk := params["b"]
    
    if !aOk || !bOk {
        return nil, fmt.Errorf("missing 'a' or 'b' parameter")
    }
    
    // Similar conversion logic as add
    var a, b float64
    var err error
    
    switch v := aVal.(type) {
    case float64:
        a = v
    case int:
        a = float64(v)
    case string:
        a, err = strconv.ParseFloat(v, 64)
        if err != nil {
            return nil, fmt.Errorf("invalid number for 'a': %v", v)
        }
    default:
        return nil, fmt.Errorf("invalid type for 'a': %T", v)
    }
    
    switch v := bVal.(type) {
    case float64:
        b = v
    case int:
        b = float64(v)
    case string:
        b, err = strconv.ParseFloat(v, 64)
        if err != nil {
            return nil, fmt.Errorf("invalid number for 'b': %v", v)
        }
    default:
        return nil, fmt.Errorf("invalid type for 'b': %T", v)
    }
    
    result := a * b
    return map[string]interface{}{
        "result": result,
        "operation": "multiplication",
    }, nil
}

// Demo method to show tool discovery and calling
func (a *Agent) demonstrateToolDiscovery() error {
    log.Println("Demonstrating tool discovery...")
    
    // Discover math tools
    tools, err := a.mcpClient.DiscoverTools(protocol.ToolQuery{
        Capabilities: []string{"math.*"},
        MaxResults:   10,
    }, a.ID)
    
    if err != nil {
        return fmt.Errorf("tool discovery failed: %w", err)
    }
    
    log.Printf("Found %d math tools", len(tools))
    for _, tool := range tools {
        log.Printf("  Agent: %s, Tools: %v", tool.AgentID, tool.MCPTools)
    }
    
    // Try to call a math tool on another agent (if available)
    if len(tools) > 0 {
        for _, tool := range tools {
            if tool.AgentID != a.ID { // Don't call ourselves
                result, err := a.mcpClient.CallTool(tool.AgentID, "math.add", map[string]interface{}{
                    "a": 5,
                    "b": 3,
                })
                if err != nil {
                    log.Printf("Failed to call remote tool: %v", err)
                } else {
                    log.Printf("Remote math.add result: %v", result)
                }
                break
            }
        }
    }
    
    return nil
}
```

**Completion Criteria**:
- Agent exposes tools via MCP server
- Agent registers with broker including MCP endpoint and body definition
- Agent can discover and call tools on other agents
- Math tools work correctly with proper parameter validation

### Phase H: Simple End-to-End Example
**Objective**: Create working demonstration of MCP federation
**Scope**: Runnable example showing two agents federating tools
**Dependencies**: Phase G

**Files Created**: `examples/mcp-federation/README.md`

```markdown
# MCP Federation Demo

This example demonstrates two FEM agents federating their MCP tools through a broker.

## Setup

1. Start the broker:
```bash
./fem-broker --listen :8443
```

2. Start the calculator agent:
```bash
./fem-coder --broker https://localhost:8443 --agent calculator-001 --mcp-port 8080
```

3. Start the consumer agent:
```bash
./fem-coder --broker https://localhost:8443 --agent consumer-001 --mcp-port 8081
```

## What Happens

1. Both agents register with the broker, advertising their MCP tools
2. The broker indexes all available tools
3. Either agent can discover and use tools from the other agent
4. Tools are called via standard MCP protocol through the federated network

## Testing

You can test tool discovery by calling the broker directly:

```bash
curl -k -X POST https://localhost:8443/fep \
  -H "Content-Type: application/json" \
  -d '{
    "type": "discoverTools",
    "agent": "test-client",
    "ts": 1641234567890,
    "nonce": "test-123",
    "body": {
      "query": {
        "capabilities": ["math.*"],
        "maxResults": 10
      },
      "requestId": "test-discovery"
    }
  }'
```

You can test direct MCP tool calls:

```bash
curl -X POST http://localhost:8080/tools/call \
  -H "Content-Type: application/json" \
  -d '{
    "name": "math.add",
    "arguments": {"a": 5, "b": 3}
  }'
```
```

**Files Created**: `examples/mcp-federation/test-federation.sh`

```bash
#!/bin/bash

# Test script for MCP federation demo

set -e

echo "=== FEM MCP Federation Demo ==="

# Function to check if service is running
check_service() {
    local url=$1
    local name=$2
    echo "Checking $name at $url..."
    if curl -s -f "$url" > /dev/null 2>&1; then
        echo "✓ $name is running"
        return 0
    else
        echo "✗ $name is not responding"
        return 1
    fi
}

# Function to test tool discovery
test_discovery() {
    echo "Testing tool discovery..."
    
    # Create discovery request
    local discovery_request='{
        "type": "discoverTools",
        "agent": "test-client",
        "ts": '$(date +%s%3N)',
        "nonce": "test-'$(date +%s)'",
        "body": {
            "query": {
                "capabilities": ["math.*"],
                "maxResults": 10
            },
            "requestId": "test-discovery-'$(date +%s)'"
        }
    }'
    
    echo "Sending discovery request to broker..."
    response=$(curl -s -k -X POST https://localhost:8443/fep \
        -H "Content-Type: application/json" \
        -d "$discovery_request")
    
    if echo "$response" | grep -q "tools"; then
        echo "✓ Tool discovery successful"
        echo "Found tools:"
        echo "$response" | jq -r '.tools[].mcpTools[].name' 2>/dev/null | sed 's/^/  - /'
    else
        echo "✗ Tool discovery failed"
        echo "Response: $response"
        return 1
    fi
}

# Function to test direct MCP call
test_mcp_call() {
    echo "Testing direct MCP tool call..."
    
    local mcp_request='{
        "name": "math.add",
        "arguments": {"a": 5, "b": 3}
    }'
    
    response=$(curl -s -X POST http://localhost:8080/tools/call \
        -H "Content-Type: application/json" \
        -d "$mcp_request")
    
    if echo "$response" | grep -q "result"; then
        echo "✓ MCP tool call successful"
        result=$(echo "$response" | jq -r '.result.result' 2>/dev/null)
        echo "  math.add(5, 3) = $result"
    else
        echo "✗ MCP tool call failed"
        echo "Response: $response"
        return 1
    fi
}

# Main test sequence
echo "1. Checking if services are running..."
check_service "https://localhost:8443/health" "Broker" || {
    echo "Please start the broker first: ./fem-broker --listen :8443"
    exit 1
}

check_service "http://localhost:8080/tools/list" "Calculator Agent MCP Server" || {
    echo "Please start calculator agent: ./fem-coder --broker https://localhost:8443 --agent calculator-001 --mcp-port 8080"
    exit 1
}

echo
echo "2. Testing tool discovery..."
test_discovery

echo
echo "3. Testing direct MCP tool call..."
test_mcp_call

echo
echo "=== Demo Complete ==="
echo "✓ FEM MCP federation is working correctly!"
echo
echo "Next steps:"
echo "- Start a second agent on port 8081 to see cross-agent tool calls"
echo "- Try discovering tools with different capability patterns"
echo "- Explore the broker's tool registry"
```

**Completion Criteria**:
- Demo script runs successfully with broker and one agent
- Tool discovery returns correct results
- Direct MCP tool calls work
- Documentation is clear and actionable

### Phase I: Testing and Validation
**Objective**: Comprehensive testing of all MCP federation features
**Scope**: Integration tests and validation scenarios
**Dependencies**: Phase H

**Files Created**: `test/integration/mcp_federation_test.go`

```go
package integration

import (
    "bytes"
    "encoding/json"
    "fmt"
    "net/http"
    "testing"
    "time"
    
    "github.com/fep-fem/protocol"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestMCPFederationComplete(t *testing.T) {
    // This test validates the complete MCP federation flow
    
    // Start test broker
    broker := startTestBroker(t, ":8443")
    defer broker.Stop()
    
    // Start test agent with MCP server
    agent1 := startTestAgent(t, "agent1", "https://localhost:8443", 8080)
    defer agent1.Stop()
    
    // Wait for registration
    time.Sleep(500 * time.Millisecond)
    
    // Test 1: Tool discovery
    tools, err := discoverTools(t, "https://localhost:8443", []string{"math.*"})
    require.NoError(t, err)
    require.NotEmpty(t, tools)
    
    foundMathAdd := false
    foundMathMultiply := false
    for _, tool := range tools {
        for _, mcpTool := range tool.MCPTools {
            if mcpTool.Name == "math.add" {
                foundMathAdd = true
            }
            if mcpTool.Name == "math.multiply" {
                foundMathMultiply = true
            }
        }
    }
    
    assert.True(t, foundMathAdd, "Should find math.add tool")
    assert.True(t, foundMathMultiply, "Should find math.multiply tool")
    
    // Test 2: Direct MCP tool call
    result, err := callMCPTool(t, "http://localhost:8080", "math.add", map[string]interface{}{
        "a": 5,
        "b": 3,
    })
    require.NoError(t, err)
    
    resultMap, ok := result.(map[string]interface{})
    require.True(t, ok, "Result should be a map")
    
    mathResult, ok := resultMap["result"].(float64)
    require.True(t, ok, "Result should contain numeric result")
    assert.Equal(t, 8.0, mathResult, "5 + 3 should equal 8")
    
    // Test 3: Cross-agent tool discovery
    agent2 := startTestAgent(t, "agent2", "https://localhost:8443", 8081)
    defer agent2.Stop()
    
    time.Sleep(500 * time.Millisecond)
    
    // Agent2 should be able to discover agent1's tools
    tools2, err := discoverTools(t, "https://localhost:8443", []string{"math.*"})
    require.NoError(t, err)
    
    // Should find tools from both agents
    agentCount := make(map[string]bool)
    for _, tool := range tools2 {
        agentCount[tool.AgentID] = true
    }
    
    assert.Len(t, agentCount, 2, "Should find tools from both agents")
    assert.True(t, agentCount["agent1"], "Should find agent1 tools")
    assert.True(t, agentCount["agent2"], "Should find agent2 tools")
}

func TestEmbodimentUpdate(t *testing.T) {
    broker := startTestBroker(t, ":8444")
    defer broker.Stop()
    
    agent := startTestAgent(t, "adaptive-agent", "https://localhost:8444", 8082)
    defer agent.Stop()
    
    // Test embodiment update
    updateBody := protocol.EmbodimentUpdateBody{
        EnvironmentType: "cloud",
        BodyDefinition: protocol.BodyDefinition{
            Name:        "cloud-body",
            Environment: "cloud",
            Capabilities: []string{"s3.read", "s3.write"},
            MCPTools: []protocol.MCPTool{
                {
                    Name:        "s3.read",
                    Description: "Read from S3",
                },
            },
        },
        MCPEndpoint:  "http://localhost:8082",
        UpdatedTools: []string{"s3.read"},
    }
    
    err := sendEmbodimentUpdate(t, "https://localhost:8444", "adaptive-agent", updateBody)
    require.NoError(t, err)
    
    // Verify tools were updated
    time.Sleep(200 * time.Millisecond)
    
    tools, err := discoverTools(t, "https://localhost:8444", []string{"s3.*"})
    require.NoError(t, err)
    require.NotEmpty(t, tools)
    
    found := false
    for _, tool := range tools {
        if tool.AgentID == "adaptive-agent" && tool.EnvironmentType == "cloud" {
            found = true
            break
        }
    }
    assert.True(t, found, "Should find updated embodiment")
}

// Helper functions
func discoverTools(t *testing.T, brokerURL string, capabilities []string) ([]protocol.DiscoveredTool, error) {
    envelope := protocol.DiscoverToolsEnvelope{
        BaseEnvelope: protocol.BaseEnvelope{
            Type: protocol.EnvelopeDiscoverTools,
            CommonHeaders: protocol.CommonHeaders{
                Agent: "test-client",
                TS:    time.Now().UnixMilli(),
                Nonce: fmt.Sprintf("test-%d", time.Now().UnixNano()),
            },
        },
        Body: protocol.DiscoverToolsBody{
            Query: protocol.ToolQuery{
                Capabilities: capabilities,
                MaxResults:   100,
            },
            RequestID: fmt.Sprintf("req-%d", time.Now().UnixNano()),
        },
    }
    
    data, err := json.Marshal(envelope)
    if err != nil {
        return nil, err
    }
    
    client := &http.Client{
        Transport: &http.Transport{
            TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
        },
    }
    
    resp, err := client.Post(brokerURL+"/fep", "application/json", bytes.NewReader(data))
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    var response struct {
        Tools []protocol.DiscoveredTool `json:"tools"`
    }
    
    if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
        return nil, err
    }
    
    return response.Tools, nil
}

func callMCPTool(t *testing.T, endpoint, toolName string, params map[string]interface{}) (interface{}, error) {
    request := map[string]interface{}{
        "name":      toolName,
        "arguments": params,
    }
    
    data, err := json.Marshal(request)
    if err != nil {
        return nil, err
    }
    
    resp, err := http.Post(endpoint+"/tools/call", "application/json", bytes.NewReader(data))
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    var response struct {
        Result interface{} `json:"result"`
    }
    
    if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
        return nil, err
    }
    
    return response.Result, nil
}

func sendEmbodimentUpdate(t *testing.T, brokerURL, agentID string, body protocol.EmbodimentUpdateBody) error {
    envelope := protocol.EmbodimentUpdateEnvelope{
        BaseEnvelope: protocol.BaseEnvelope{
            Type: protocol.EnvelopeEmbodimentUpdate,
            CommonHeaders: protocol.CommonHeaders{
                Agent: agentID,
                TS:    time.Now().UnixMilli(),
                Nonce: fmt.Sprintf("update-%d", time.Now().UnixNano()),
            },
        },
        Body: body,
    }
    
    data, err := json.Marshal(envelope)
    if err != nil {
        return err
    }
    
    client := &http.Client{
        Transport: &http.Transport{
            TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
        },
    }
    
    resp, err := client.Post(brokerURL+"/fep", "application/json", bytes.NewReader(data))
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != http.StatusOK {
        return fmt.Errorf("broker returned status %d", resp.StatusCode)
    }
    
    return nil
}

// Test infrastructure helpers would be implemented here
func startTestBroker(t *testing.T, addr string) TestBroker { 
    // Implementation for test broker
    return TestBroker{}
}

func startTestAgent(t *testing.T, id, brokerURL string, mcpPort int) TestAgent {
    // Implementation for test agent
    return TestAgent{}
}

type TestBroker struct{}
func (b TestBroker) Stop() {}

type TestAgent struct{}
func (a TestAgent) Stop() {}
```

**Completion Criteria**:
- All integration tests pass
- Tests cover tool discovery, MCP calls, and embodiment updates
- Cross-agent federation works correctly
- Test infrastructure is reusable for future development

## Success Criteria

### Overall Implementation Success
- [ ] All phases complete without breaking existing functionality
- [ ] 3-line MCP tool federation works as documented
- [ ] Agents can discover and call each other's MCP tools
- [ ] Embodiment updates work correctly
- [ ] Integration tests pass consistently
- [ ] Documentation examples are runnable

### Technical Success Metrics
- [ ] Protocol extends cleanly without breaking changes
- [ ] Broker handles >100 concurrent agents with MCP tools
- [ ] Tool discovery responds in <500ms with 1000 indexed tools
- [ ] MCP tool calls complete in <2s end-to-end
- [ ] Memory usage scales linearly with registered tools

### User Experience Success
- [ ] Developer can add MCP federation to existing agent in <5 minutes
- [ ] Tool discovery "just works" without configuration
- [ ] Error messages are clear and actionable
- [ ] Examples run successfully on first try

This phased approach ensures each implementation unit is focused, achievable, and builds logically on previous work. Each phase can be completed in a single development session while maintaining the overall vision of MCP federation through FEM.