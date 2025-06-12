# Agent Development Guide: Host & Guest Patterns

Build custom FEM Protocol agents for **Secure Hosted Embodiment**. This guide covers developing both host and guest agents that enable the core FEM Protocol paradigm where guest "minds" inhabit host-offered "bodies" with delegated control.

## Table of Contents
- [Development Overview](#development-overview)
- [Host Agent Development](#host-agent-development)
- [Guest Agent Development](#guest-agent-development)
- [Security Implementation](#security-implementation)
- [Session Management](#session-management)
- [Testing Embodiment](#testing-embodiment)
- [Production Deployment](#production-deployment)
- [Example Implementations](#example-implementations)

## Development Overview

### FEM Protocol Agent Types

**Host Agents**: Offer "bodies" (sandboxed capability sets) for guest embodiment
- Define and secure body definitions with MCP tools
- Manage embodiment sessions with cryptographic validation
- Enforce security policies and resource limits
- Provide audit logging for all guest actions

**Guest Agents**: Discover and inhabit host-offered bodies
- Search for suitable bodies based on capability needs
- Request embodiment with specific intentions
- Exercise delegated control within granted permissions
- Maintain session state and handle expiration

### Core Embodiment Flow

```
┌─────────────┐    ┌──────────────┐    ┌─────────────┐    ┌──────────────┐
│   Guest     │───►│  Discovery   │───►│ Embodiment  │───►│ Delegated    │
│ Discovery   │    │   & Eval     │    │  Request    │    │  Control     │
└─────────────┘    └──────────────┘    └─────────────┘    └──────────────┘
        │                                                           │
        ▼                                                           ▼
┌─────────────┐    ┌──────────────┐    ┌─────────────┐    ┌──────────────┐
│  Host Body  │───►│   Security   │───►│   Session   │───►│   Session    │
│ Advertising │    │ Validation   │    │   Grant     │    │ Termination  │
└─────────────┘    └──────────────┘    └─────────────┘    └──────────────┘
```

### Key Development Principles

1. **Secure Delegated Control** - Hosts retain ultimate authority while granting specific control
2. **Cryptographic Identity** - All agents verified via Ed25519 signatures
3. **Session Isolation** - Each embodiment session is cryptographically isolated
4. **MCP Integration** - Bodies are defined through MCP tool schemas
5. **Environment Awareness** - Bodies adapt to deployment contexts

## Host Agent Development

Host agents offer "bodies" for guest embodiment, managing security, permissions, and session lifecycle.

### Host Agent Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                     Host Agent                             │
│  ┌─────────────────────────────────────────────────────┐   │
│  │              Body Definitions                       │   │
│  │  • MCP Tool Schemas                                 │   │
│  │  • Security Policies                               │   │
│  │  • Resource Limits                                 │   │
│  └─────────────────────────────────────────────────────┘   │
│  ┌─────────────────────────────────────────────────────┐   │
│  │          Embodiment Sessions                        │   │
│  │  • Session Token Management                         │   │
│  │  • Permission Validation                            │   │
│  │  • Resource Monitoring                              │   │
│  └─────────────────────────────────────────────────────┘   │
│  ┌─────────────────────────────────────────────────────┐   │
│  │                MCP Server                           │   │
│  │  • Tool Implementation                              │   │
│  │  • Session-Scoped Endpoints                        │   │
│  │  • Audit Logging                                   │   │
│  └─────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
```

### Prerequisites

- **Go 1.21+** for Go agents
- **FEM Protocol Package** from this repository  
- **MCP Server Library** for tool implementation
- **Test Broker** for development and testing

### Host Project Structure

```
my-host-agent/
├── main.go              # Host agent entry point
├── host/
│   ├── host.go          # Core host implementation
│   ├── bodies.go        # Body definition management
│   ├── sessions.go      # Embodiment session management
│   └── security.go      # Security policy enforcement
├── mcp/
│   ├── server.go        # MCP server implementation
│   ├── tools.go         # Tool implementations
│   └── validation.go    # Input validation
├── config/
│   └── config.go        # Configuration management
├── go.mod
└── README.md
```

### Host Agent Implementation

#### 1. Core Host Structure

```go
// host/host.go
package host

import (
    "crypto/ed25519"
    "encoding/json"
    "fmt"
    "net/http"
    "sync"
    "time"
    
    "github.com/fem-protocol/protocol"
    "github.com/fem-protocol/mcp"
)

type HostAgent struct {
    ID          string
    BrokerURL   string
    PrivKey     ed25519.PrivateKey
    PubKey      ed25519.PublicKey
    
    // Host-specific fields
    Bodies      map[string]*BodyDefinition
    Sessions    map[string]*EmbodimentSession
    MCPServer   *mcp.Server
    
    // Security and monitoring
    SecurityPolicy SecurityPolicy
    AuditLogger   AuditLogger
    
    // Synchronization
    sessionsMux sync.RWMutex
    client      *http.Client
}

type BodyDefinition struct {
    BodyID          string           `json:"bodyId"`
    Description     string           `json:"description"`
    EnvironmentType string           `json:"environmentType"`
    MCPTools        []MCPToolDef     `json:"mcpTools"`
    SecurityPolicy  SecurityPolicy   `json:"securityPolicy"`
    
    // Session management
    MaxConcurrentGuests    int           `json:"maxConcurrentGuests"`
    DefaultSessionDuration time.Duration `json:"defaultSessionDuration"`
    MaxSessionDuration     time.Duration `json:"maxSessionDuration"`
    
    // Trust requirements
    RequireApproval    bool   `json:"requireApproval"`
    TrustLevelRequired string `json:"trustLevelRequired"`
}

type EmbodimentSession struct {
    SessionToken  string
    GuestID      string
    BodyID       string
    
    // Session state
    StartTime    time.Time
    ExpiryTime   time.Time
    Permissions  []string
    
    // Security and monitoring
    ResourceUsage ResourceUsage
    ActionCount   int
    ViolationCount int
    
    // MCP endpoint
    MCPEndpoint string
}

func NewHostAgent(id, brokerURL string) (*HostAgent, error) {
    // Generate cryptographic identity
    pubKey, privKey, err := ed25519.GenerateKey(nil)
    if err != nil {
        return nil, fmt.Errorf("failed to generate keys: %w", err)
    }
    
    // Initialize MCP server
    mcpServer := mcp.NewServer()
    
    return &HostAgent{
        ID:        id,
        BrokerURL: brokerURL,
        PrivKey:   privKey,
        PubKey:    pubKey,
        Bodies:    make(map[string]*BodyDefinition),
        Sessions:  make(map[string]*EmbodimentSession),
        MCPServer: mcpServer,
        client:    &http.Client{},
    }, nil
}
```

#### 2. Body Definition Creation

```go
// host/bodies.go
func (h *HostAgent) DefineBody(bodyDef *BodyDefinition) error {
    // Validate body definition
    if err := h.validateBodyDefinition(bodyDef); err != nil {
        return fmt.Errorf("invalid body definition: %w", err)
    }
    
    // Register MCP tools for this body
    for _, toolDef := range bodyDef.MCPTools {
        if err := h.registerMCPTool(bodyDef.BodyID, toolDef); err != nil {
            return fmt.Errorf("failed to register tool %s: %w", toolDef.Name, err)
        }
    }
    
    // Store body definition
    h.Bodies[bodyDef.BodyID] = bodyDef
    
    log.Printf("Body '%s' defined with %d tools", bodyDef.BodyID, len(bodyDef.MCPTools))
    return nil
}

func (h *HostAgent) validateBodyDefinition(bodyDef *BodyDefinition) error {
    if bodyDef.BodyID == "" {
        return fmt.Errorf("bodyId is required")
    }
    
    if len(bodyDef.MCPTools) == 0 {
        return fmt.Errorf("at least one MCP tool is required")
    }
    
    // Validate security policy
    if err := h.validateSecurityPolicy(&bodyDef.SecurityPolicy); err != nil {
        return fmt.Errorf("invalid security policy: %w", err)
    }
    
    return nil
}

// Example body definition for cross-device embodiment
func (h *HostAgent) DefineDeveloperWorkstation() error {
    bodyDef := &BodyDefinition{
        BodyID:          "developer-workstation-v1",
        Description:     "Secure development environment with terminal and file access",
        EnvironmentType: "local-development",
        MCPTools: []MCPToolDef{
            {
                Name:        "shell.execute",
                Description: "Execute shell commands in sandboxed environment",
                InputSchema: map[string]interface{}{
                    "type": "object",
                    "properties": map[string]interface{}{
                        "command": map[string]interface{}{"type": "string"},
                        "workdir": map[string]interface{}{"type": "string", "default": "/home/alice/projects"},
                        "timeout": map[string]interface{}{"type": "number", "default": 30},
                    },
                },
            },
            {
                Name:        "file.read",
                Description: "Read files from allowed project directories",
                InputSchema: map[string]interface{}{
                    "type": "object",
                    "properties": map[string]interface{}{
                        "path":     map[string]interface{}{"type": "string"},
                        "encoding": map[string]interface{}{"type": "string", "default": "utf-8"},
                    },
                },
            },
        },
        SecurityPolicy: SecurityPolicy{
            AllowedPaths:      []string{"/home/alice/projects/*", "/tmp/fem-workspace/*"},
            DeniedPaths:       []string{"/home/alice/.ssh/*", "/etc/*", "/root/*"},
            AllowedCommands:   []string{"git", "npm", "yarn", "python", "node", "go", "ls", "cat"},
            DeniedCommands:    []string{"rm -rf", "sudo", "su", "curl", "wget", "ssh"},
            ResourceLimits: ResourceLimits{
                MaxCPUPercent:    25,
                MaxMemoryMB:      512,
                MaxDiskWriteMB:   100,
                MaxNetworkKbps:   0, // No network access
            },
            MaxSessionDuration: 3600 * time.Second, // 1 hour
        },
        MaxConcurrentGuests:    2,
        DefaultSessionDuration: 1800 * time.Second, // 30 minutes
        MaxSessionDuration:     3600 * time.Second, // 1 hour
        RequireApproval:        false,
        TrustLevelRequired:     "personal-device",
    }
    
    return h.DefineBody(bodyDef)
}
```

#### 3. Host Registration with Broker

```go
func (h *HostAgent) Register() error {
    // Prepare body definitions for registration
    offeredBodies := make([]interface{}, 0, len(h.Bodies))
    for _, bodyDef := range h.Bodies {
        offeredBodies = append(offeredBodies, map[string]interface{}{
            "bodyId":          bodyDef.BodyID,
            "description":     bodyDef.Description,
            "environmentType": bodyDef.EnvironmentType,
            "mcpTools":        bodyDef.MCPTools,
            "securityPolicy":  bodyDef.SecurityPolicy,
        })
    }
    
    // Create registration envelope
    envelope := &protocol.RegisterAgentEnvelope{
        BaseEnvelope: protocol.BaseEnvelope{
            Type: protocol.EnvelopeRegisterAgent,
            CommonHeaders: protocol.CommonHeaders{
                Agent: h.ID,
                TS:    time.Now().UnixMilli(),
                Nonce: generateNonce(),
            },
        },
        Body: protocol.RegisterAgentBody{
            PubKey:       protocol.EncodePublicKey(h.PubKey),
            AgentType:    "host",
            Capabilities: h.extractCapabilities(),
            OfferedBodies: offeredBodies,
            MCPEndpoint:   fmt.Sprintf("https://%s:%d/mcp", h.getHostname(), h.MCPServer.Port),
            Metadata: map[string]interface{}{
                "version":     "1.0.0",
                "description": "FEM Protocol Host Agent",
                "trustLevel":  "personal-device",
            },
        },
    }
    
    // Sign and send
    if err := envelope.Sign(h.PrivKey); err != nil {
        return fmt.Errorf("failed to sign envelope: %w", err)
    }
    
    return h.sendEnvelope(envelope)
}

func (h *HostAgent) extractCapabilities() []string {
    capabilities := make([]string, 0)
    for _, bodyDef := range h.Bodies {
        for _, tool := range bodyDef.MCPTools {
            capabilities = append(capabilities, tool.Name)
        }
    }
    return capabilities
}
```

#### 4. Embodiment Request Handling

```go
// host/sessions.go
func (h *HostAgent) HandleEmbodimentRequest(envelope *protocol.RequestEmbodimentEnvelope) error {
    guestID := envelope.Agent
    bodyID := envelope.Body.BodyID
    requestID := envelope.Body.RequestID
    
    // Validate request
    bodyDef, exists := h.Bodies[bodyID]
    if !exists {
        return h.sendEmbodimentDenied(requestID, guestID, "BODY_NOT_FOUND", 
            "Requested body does not exist")
    }
    
    // Check concurrent session limits
    if h.getActiveSessionsForBody(bodyID) >= bodyDef.MaxConcurrentGuests {
        return h.sendEmbodimentDenied(requestID, guestID, "SESSION_LIMIT_EXCEEDED", 
            "Maximum concurrent guests reached")
    }
    
    // Validate guest trust level (in production, check against guest's reputation)
    // For development, we'll allow the request
    
    // Create session
    session, err := h.createEmbodimentSession(guestID, bodyID, envelope.Body.RequestedDuration)
    if err != nil {
        return h.sendEmbodimentDenied(requestID, guestID, "SESSION_CREATION_FAILED", 
            "Failed to create embodiment session")
    }
    
    // Grant embodiment
    return h.sendEmbodimentGranted(requestID, session)
}

func (h *HostAgent) createEmbodimentSession(guestID, bodyID string, requestedDuration int) (*EmbodimentSession, error) {
    bodyDef := h.Bodies[bodyID]
    
    // Generate unique session token
    sessionToken := generateSessionToken()
    
    // Calculate session duration
    duration := time.Duration(requestedDuration) * time.Second
    if duration > bodyDef.MaxSessionDuration {
        duration = bodyDef.MaxSessionDuration
    }
    if duration == 0 {
        duration = bodyDef.DefaultSessionDuration
    }
    
    // Create session
    session := &EmbodimentSession{
        SessionToken: sessionToken,
        GuestID:     guestID,
        BodyID:      bodyID,
        StartTime:   time.Now(),
        ExpiryTime:  time.Now().Add(duration),
        Permissions: h.generatePermissions(bodyDef),
        MCPEndpoint: fmt.Sprintf("https://%s:%d/mcp/sessions/%s", 
            h.getHostname(), h.MCPServer.Port, sessionToken),
    }
    
    // Store session
    h.sessionsMux.Lock()
    h.Sessions[sessionToken] = session
    h.sessionsMux.Unlock()
    
    // Set up session-specific MCP endpoint
    h.MCPServer.CreateSessionEndpoint(sessionToken, session)
    
    // Log session creation
    h.AuditLogger.LogSessionStart(session)
    
    return session, nil
}

func (h *HostAgent) generatePermissions(bodyDef *BodyDefinition) []string {
    permissions := make([]string, 0)
    
    for _, tool := range bodyDef.MCPTools {
        for _, path := range bodyDef.SecurityPolicy.AllowedPaths {
            permissions = append(permissions, fmt.Sprintf("%s:%s", tool.Name, path))
        }
    }
    
    return permissions
}

func generateSessionToken() string {
    // Generate cryptographically secure session token
    tokenBytes := make([]byte, 32)
    _, err := rand.Read(tokenBytes)
    if err != nil {
        panic("Failed to generate session token")
    }
    return base64.URLEncoding.EncodeToString(tokenBytes)
}
```

## Guest Agent Development

Guest agents discover and inhabit host-offered bodies to exercise delegated control.

### Guest Agent Architecture

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                           Guest Agent                                      │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │                    Discovery Engine                                 │   │
│  │  • Body Search & Evaluation                                         │   │
│  │  • Capability Matching                                              │   │
│  │  • Trust Assessment                                                 │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │                  Embodiment Client                                  │   │
│  │  • Session Management                                               │   │
│  │  • Permission Tracking                                              │   │
│  │  • Tool Call Execution                                              │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │                     Agent Mind                                      │   │
│  │  • Goal-Oriented Logic                                              │   │
│  │  • Environment Adaptation                                           │   │
│  │  • Context Maintenance                                              │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Guest Project Structure

```
my-guest-agent/
├── main.go              # Guest agent entry point
├── guest/
│   ├── guest.go          # Core guest implementation
│   ├── discovery.go      # Body discovery logic
│   ├── embodiment.go     # Embodiment session management
│   └── mind.go           # Agent logic and goals
├── mcp/
│   ├── client.go         # MCP client for tool calls
│   └── session.go        # Session-aware MCP client
├── config/
│   └── config.go         # Configuration management
├── go.mod
└── README.md
```

### Guest Agent Implementation

#### 1. Core Guest Structure

```go
// guest/guest.go
package guest

import (
    "crypto/ed25519"
    "fmt"
    "net/http"
    "sync"
    "time"
    
    "github.com/fem-protocol/protocol"
    "github.com/fem-protocol/mcp"
)

type GuestAgent struct {
    ID        string
    BrokerURL string
    PrivKey   ed25519.PrivateKey
    PubKey    ed25519.PublicKey
    
    // Guest-specific fields
    CurrentSessions map[string]*EmbodimentSession
    Preferences     DiscoveryPreferences
    Mind           AgentMind
    
    // MCP client for tool calls
    MCPClient *mcp.Client
    
    // Synchronization
    sessionsMux sync.RWMutex
    client      *http.Client
}

type DiscoveryPreferences struct {
    RequiredCapabilities []string
    PreferredEnvironments []string
    MaxSessionDuration   time.Duration
    TrustLevelRequired   string
}

type EmbodimentSession struct {
    SessionToken  string
    HostID        string
    BodyID        string
    MCPEndpoint   string
    
    // Session state
    StartTime     time.Time
    ExpiryTime    time.Time
    Permissions   []string
    
    // Session context
    Goals         []string
    CurrentTask   string
    Context       map[string]interface{}
}

type AgentMind struct {
    Goals         []Goal
    CurrentGoal   *Goal
    Context       map[string]interface{}
    MemoryStore   map[string]interface{}
}

type Goal struct {
    ID          string
    Description string
    Actions     []string
    Completed   bool
}

func NewGuestAgent(id, brokerURL string, preferences DiscoveryPreferences) (*GuestAgent, error) {
    // Generate cryptographic identity
    pubKey, privKey, err := ed25519.GenerateKey(nil)
    if err != nil {
        return nil, fmt.Errorf("failed to generate keys: %w", err)
    }
    
    return &GuestAgent{
        ID:              id,
        BrokerURL:       brokerURL,
        PrivKey:         privKey,
        PubKey:          pubKey,
        CurrentSessions: make(map[string]*EmbodimentSession),
        Preferences:     preferences,
        MCPClient:       mcp.NewClient(),
        client:          &http.Client{},
    }, nil
}
```

#### 2. Body Discovery Implementation

```go
// guest/discovery.go
func (g *GuestAgent) DiscoverBodies() ([]BodyDiscoveryResult, error) {
    // Create discovery request
    envelope := &protocol.DiscoverBodiesEnvelope{
        BaseEnvelope: protocol.BaseEnvelope{
            Type: protocol.EnvelopeDiscoverBodies,
            CommonHeaders: protocol.CommonHeaders{
                Agent: g.ID,
                TS:    time.Now().UnixMilli(),
                Nonce: generateNonce(),
            },
        },
        Body: protocol.DiscoverBodiesBody{
            Query: protocol.DiscoveryQuery{
                Capabilities:          g.Preferences.RequiredCapabilities,
                EnvironmentType:       g.getPreferredEnvironment(),
                TrustLevel:           g.Preferences.TrustLevelRequired,
                MaxResults:           10,
                IncludeSecurityPolicies: true,
            },
            GuestProfile: protocol.GuestProfile{
                GuestID:                g.ID,
                PreferredSessionDuration: int(g.Preferences.MaxSessionDuration.Seconds()),
                IntendedUse:             g.getCurrentGoalDescription(),
            },
            RequestID: generateRequestID(),
        },
    }
    
    // Sign and send
    if err := envelope.Sign(g.PrivKey); err != nil {
        return nil, fmt.Errorf("failed to sign discovery request: %w", err)
    }
    
    // Send to broker and wait for response
    response, err := g.sendAndWaitForResponse(envelope)
    if err != nil {
        return nil, fmt.Errorf("discovery request failed: %w", err)
    }
    
    // Parse response
    return g.parseDiscoveryResponse(response)
}

func (g *GuestAgent) EvaluateBodies(bodies []BodyDiscoveryResult) *BodyDiscoveryResult {
    var bestBody *BodyDiscoveryResult
    bestScore := 0.0
    
    for _, body := range bodies {
        score := g.scoreBody(&body)
        if score > bestScore {
            bestScore = score
            bestBody = &body
        }
    }
    
    return bestBody
}

func (g *GuestAgent) scoreBody(body *BodyDiscoveryResult) float64 {
    score := 0.0
    
    // Score based on capability match
    capabilityMatch := g.calculateCapabilityMatch(body.Capabilities)
    score += capabilityMatch * 0.4
    
    // Score based on trust level
    trustScore := g.calculateTrustScore(body.Availability.TrustScore)
    score += trustScore * 0.3
    
    // Score based on availability
    availabilityScore := g.calculateAvailabilityScore(body.Availability)
    score += availabilityScore * 0.2
    
    // Score based on security policy alignment
    securityScore := g.calculateSecurityScore(body.SecurityPolicy)
    score += securityScore * 0.1
    
    return score
}
```

#### 3. Embodiment Request and Session Management

```go
// guest/embodiment.go
func (g *GuestAgent) RequestEmbodiment(body *BodyDiscoveryResult, goals []string) (*EmbodimentSession, error) {
    // Create embodiment request
    envelope := &protocol.RequestEmbodimentEnvelope{
        BaseEnvelope: protocol.BaseEnvelope{
            Type: protocol.EnvelopeRequestEmbodiment,
            CommonHeaders: protocol.CommonHeaders{
                Agent: g.ID,
                TS:    time.Now().UnixMilli(),
                Nonce: generateNonce(),
            },
        },
        Body: protocol.RequestEmbodimentBody{
            HostAgentID:       body.HostAgentID,
            BodyID:           body.BodyID,
            RequestedDuration: int(g.Preferences.MaxSessionDuration.Seconds()),
            IntendedActions:   goals,
            GuestCredentials: protocol.GuestCredentials{
                GuestID:              g.ID,
                TrustLevel:           "verified-user",
                PreviousSessions:     g.getPreviousSessionCount(),
                AverageSessionRating: g.getAverageRating(),
            },
            RequestID: generateRequestID(),
        },
    }
    
    // Sign and send
    if err := envelope.Sign(g.PrivKey); err != nil {
        return nil, fmt.Errorf("failed to sign embodiment request: %w", err)
    }
    
    // Send and wait for response
    response, err := g.sendAndWaitForResponse(envelope)
    if err != nil {
        return nil, fmt.Errorf("embodiment request failed: %w", err)
    }
    
    // Handle response
    switch response.Type {
    case protocol.EnvelopeEmbodimentGranted:
        return g.handleEmbodimentGranted(response.(*protocol.EmbodimentGrantedEnvelope))
    case protocol.EnvelopeEmbodimentDenied:
        return nil, g.handleEmbodimentDenied(response.(*protocol.EmbodimentDeniedEnvelope))
    default:
        return nil, fmt.Errorf("unexpected response type: %s", response.Type)
    }
}

func (g *GuestAgent) handleEmbodimentGranted(envelope *protocol.EmbodimentGrantedEnvelope) (*EmbodimentSession, error) {
    session := &EmbodimentSession{
        SessionToken: envelope.Body.SessionToken,
        HostID:      envelope.Agent,
        BodyID:      envelope.Body.BodyID,
        MCPEndpoint: envelope.Body.MCPEndpoint,
        StartTime:   time.Now(),
        ExpiryTime:  time.Unix(envelope.Body.SessionExpiry/1000, 0),
        Permissions: envelope.Body.GrantedPermissions,
        Goals:       g.Mind.getCurrentGoals(),
        Context:     make(map[string]interface{}),
    }
    
    // Store session
    g.sessionsMux.Lock()
    g.CurrentSessions[session.SessionToken] = session
    g.sessionsMux.Unlock()
    
    // Configure MCP client for this session
    g.MCPClient.ConfigureSession(session.SessionToken, session.MCPEndpoint)
    
    log.Printf("Embodiment granted: session %s expires at %v", 
        session.SessionToken, session.ExpiryTime)
    
    return session, nil
}

func (g *GuestAgent) ExecuteToolInSession(sessionToken, toolName string, parameters map[string]interface{}) (interface{}, error) {
    // Validate session
    session, exists := g.CurrentSessions[sessionToken]
    if !exists {
        return nil, fmt.Errorf("session not found: %s", sessionToken)
    }
    
    if time.Now().After(session.ExpiryTime) {
        return nil, fmt.Errorf("session expired: %s", sessionToken)
    }
    
    // Check permissions
    if !g.hasPermission(session, toolName, parameters) {
        return nil, fmt.Errorf("insufficient permissions for tool %s", toolName)
    }
    
    // Execute tool via MCP client
    result, err := g.MCPClient.CallTool(sessionToken, toolName, parameters)
    if err != nil {
        return nil, fmt.Errorf("tool execution failed: %w", err)
    }
    
    // Update session context
    g.updateSessionContext(session, toolName, parameters, result)
    
    return result, nil
}
```

#### 4. Agent Mind and Goal Management

```go
// guest/mind.go
func (g *GuestAgent) SetGoals(goals []Goal) {
    g.Mind.Goals = goals
    if len(goals) > 0 {
        g.Mind.CurrentGoal = &goals[0]
    }
}

func (g *GuestAgent) ExecuteCurrentGoal() error {
    if g.Mind.CurrentGoal == nil {
        return fmt.Errorf("no current goal set")
    }
    
    goal := g.Mind.CurrentGoal
    
    // Find suitable body for this goal
    bodies, err := g.DiscoverBodies()
    if err != nil {
        return fmt.Errorf("failed to discover bodies: %w", err)
    }
    
    if len(bodies) == 0 {
        return fmt.Errorf("no suitable bodies found for goal: %s", goal.Description)
    }
    
    // Select best body
    bestBody := g.EvaluateBodies(bodies)
    
    // Request embodiment
    session, err := g.RequestEmbodiment(bestBody, goal.Actions)
    if err != nil {
        return fmt.Errorf("failed to request embodiment: %w", err)
    }
    
    // Execute goal actions
    for _, action := range goal.Actions {
        if err := g.executeAction(session, action); err != nil {
            log.Printf("Action failed: %s - %v", action, err)
            continue
        }
    }
    
    // Mark goal as completed
    goal.Completed = true
    
    // Move to next goal
    g.advanceToNextGoal()
    
    return nil
}

func (g *GuestAgent) executeAction(session *EmbodimentSession, action string) error {
    // Parse action into tool call
    toolCall := g.parseActionToToolCall(action)
    
    // Execute tool
    result, err := g.ExecuteToolInSession(session.SessionToken, toolCall.Tool, toolCall.Parameters)
    if err != nil {
        return fmt.Errorf("failed to execute %s: %w", action, err)
    }
    
    // Update context with result
    session.Context[action] = result
    
    log.Printf("Action completed: %s", action)
    return nil
}

// Example: Cross-device development goal
func (g *GuestAgent) CreateDevelopmentGoals() {
    goals := []Goal{
        {
            ID:          "check-project-status",
            Description: "Check the status of development projects",
            Actions: []string{
                "Execute: git status in /home/alice/projects/my-app",
                "Read: /home/alice/projects/my-app/package.json",
                "Execute: npm list --depth=0",
            },
        },
        {
            ID:          "run-development-server",
            Description: "Start development server for testing",
            Actions: []string{
                "Execute: npm run dev in /home/alice/projects/my-app",
                "Check: server status on port 3000",
            },
        },
    }
    
    g.SetGoals(goals)
}
```

## Security Implementation

### Cryptographic Security

#### Ed25519 Signature Implementation

```go
// security.go
package security

import (
    "crypto/ed25519"
    "crypto/rand"
    "encoding/base64"
    "encoding/json"
    "fmt"
)

type SecurityManager struct {
    privKey ed25519.PrivateKey
    pubKey  ed25519.PublicKey
}

func NewSecurityManager() (*SecurityManager, error) {
    pubKey, privKey, err := ed25519.GenerateKey(rand.Reader)
    if err != nil {
        return nil, fmt.Errorf("failed to generate keys: %w", err)
    }
    
    return &SecurityManager{
        privKey: privKey,
        pubKey:  pubKey,
    }, nil
}

func (sm *SecurityManager) SignEnvelope(envelope interface{}) (string, error) {
    // Remove signature field for signing
    envelopeMap, err := sm.envelopeToMap(envelope)
    if err != nil {
        return "", fmt.Errorf("failed to convert envelope: %w", err)
    }
    
    delete(envelopeMap, "sig")
    
    // Serialize canonically
    canonical, err := json.Marshal(envelopeMap)
    if err != nil {
        return "", fmt.Errorf("failed to marshal envelope: %w", err)
    }
    
    // Sign
    signature := ed25519.Sign(sm.privKey, canonical)
    
    // Encode to base64
    return base64.StdEncoding.EncodeToString(signature), nil
}

func (sm *SecurityManager) VerifySignature(envelope interface{}, signature string, pubKey ed25519.PublicKey) error {
    // Decode signature
    sigBytes, err := base64.StdEncoding.DecodeString(signature)
    if err != nil {
        return fmt.Errorf("failed to decode signature: %w", err)
    }
    
    // Prepare envelope for verification
    envelopeMap, err := sm.envelopeToMap(envelope)
    if err != nil {
        return fmt.Errorf("failed to convert envelope: %w", err)
    }
    
    delete(envelopeMap, "sig")
    
    // Serialize canonically
    canonical, err := json.Marshal(envelopeMap)
    if err != nil {
        return fmt.Errorf("failed to marshal envelope: %w", err)
    }
    
    // Verify
    if !ed25519.Verify(pubKey, canonical, sigBytes) {
        return fmt.Errorf("signature verification failed")
    }
    
    return nil
}
```

### Permission Validation

#### Host Permission Enforcement

```go
// host/security.go
type SecurityValidator struct {
    policy SecurityPolicy
}

func NewSecurityValidator(policy SecurityPolicy) *SecurityValidator {
    return &SecurityValidator{policy: policy}
}

func (sv *SecurityValidator) ValidateToolCall(session *EmbodimentSession, toolName string, parameters map[string]interface{}) error {
    // Check session validity
    if time.Now().After(session.ExpiryTime) {
        return fmt.Errorf("session expired")
    }
    
    // Validate tool permissions
    if !sv.hasToolPermission(session, toolName) {
        return fmt.Errorf("tool %s not permitted in this session", toolName)
    }
    
    // Validate path restrictions
    if err := sv.validatePaths(toolName, parameters); err != nil {
        return fmt.Errorf("path validation failed: %w", err)
    }
    
    // Validate command restrictions
    if err := sv.validateCommands(toolName, parameters); err != nil {
        return fmt.Errorf("command validation failed: %w", err)
    }
    
    return nil
}

func (sv *SecurityValidator) validatePaths(toolName string, parameters map[string]interface{}) error {
    if toolName == "file.read" || toolName == "file.write" {
        path, ok := parameters["path"].(string)
        if !ok {
            return fmt.Errorf("missing path parameter")
        }
        
        // Check allowed paths
        allowed := false
        for _, allowedPath := range sv.policy.AllowedPaths {
            if matchesPattern(path, allowedPath) {
                allowed = true
                break
            }
        }
        
        if !allowed {
            return fmt.Errorf("path not allowed: %s", path)
        }
        
        // Check denied paths
        for _, deniedPath := range sv.policy.DeniedPaths {
            if matchesPattern(path, deniedPath) {
                return fmt.Errorf("path explicitly denied: %s", path)
            }
        }
    }
    
    return nil
}

func (sv *SecurityValidator) validateCommands(toolName string, parameters map[string]interface{}) error {
    if toolName == "shell.execute" {
        command, ok := parameters["command"].(string)
        if !ok {
            return fmt.Errorf("missing command parameter")
        }
        
        // Check denied commands
        for _, deniedCmd := range sv.policy.DeniedCommands {
            if strings.Contains(command, deniedCmd) {
                return fmt.Errorf("command contains denied pattern: %s", deniedCmd)
            }
        }
        
        // Extract base command
        baseCmd := strings.Fields(command)[0]
        
        // Check if base command is in allowed list
        if len(sv.policy.AllowedCommands) > 0 {
            allowed := false
            for _, allowedCmd := range sv.policy.AllowedCommands {
                if baseCmd == allowedCmd {
                    allowed = true
                    break
                }
            }
            
            if !allowed {
                return fmt.Errorf("command not in allowed list: %s", baseCmd)
            }
        }
    }
    
    return nil
}
```

## Session Management

### Session Lifecycle

#### Host Session Management

```go
// host/sessions.go
type SessionManager struct {
    sessions    map[string]*EmbodimentSession
    sessionsMux sync.RWMutex
    
    // Monitoring
    auditLogger AuditLogger
    metrics     *SessionMetrics
}

func NewSessionManager(auditLogger AuditLogger) *SessionManager {
    return &SessionManager{
        sessions:    make(map[string]*EmbodimentSession),
        auditLogger: auditLogger,
        metrics:     NewSessionMetrics(),
    }
}

func (sm *SessionManager) CreateSession(guestID, bodyID string, duration time.Duration) (*EmbodimentSession, error) {
    sessionToken := generateSecureToken()
    
    session := &EmbodimentSession{
        SessionToken: sessionToken,
        GuestID:     guestID,
        BodyID:      bodyID,
        StartTime:   time.Now(),
        ExpiryTime:  time.Now().Add(duration),
        Permissions: sm.generatePermissions(bodyID),
        Context:     make(map[string]interface{}),
    }
    
    sm.sessionsMux.Lock()
    sm.sessions[sessionToken] = session
    sm.sessionsMux.Unlock()
    
    // Start session monitoring
    go sm.monitorSession(session)
    
    // Log session creation
    sm.auditLogger.LogSessionCreated(session)
    
    return session, nil
}

func (sm *SessionManager) ValidateSession(sessionToken string) (*EmbodimentSession, error) {
    sm.sessionsMux.RLock()
    session, exists := sm.sessions[sessionToken]
    sm.sessionsMux.RUnlock()
    
    if !exists {
        return nil, fmt.Errorf("session not found")
    }
    
    if time.Now().After(session.ExpiryTime) {
        sm.TerminateSession(sessionToken, "expired")
        return nil, fmt.Errorf("session expired")
    }
    
    return session, nil
}

func (sm *SessionManager) monitorSession(session *EmbodimentSession) {
    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()
    
    for {
        select {
        case <-ticker.C:
            if time.Now().After(session.ExpiryTime) {
                sm.TerminateSession(session.SessionToken, "expired")
                return
            }
            
            // Check resource usage
            if session.ResourceUsage.ExceedsLimits() {
                sm.TerminateSession(session.SessionToken, "resource_limit_exceeded")
                return
            }
            
            // Check violation count
            if session.ViolationCount > 5 {
                sm.TerminateSession(session.SessionToken, "too_many_violations")
                return
            }
        }
    }
}

func (sm *SessionManager) TerminateSession(sessionToken, reason string) {
    sm.sessionsMux.Lock()
    session, exists := sm.sessions[sessionToken]
    if exists {
        delete(sm.sessions, sessionToken)
    }
    sm.sessionsMux.Unlock()
    
    if exists {
        // Log session termination
        sm.auditLogger.LogSessionTerminated(session, reason)
        
        // Clean up session resources
        sm.cleanupSessionResources(session)
    }
}
```

#### Guest Session Management

```go
// guest/sessions.go
type GuestSessionManager struct {
    sessions    map[string]*EmbodimentSession
    sessionsMux sync.RWMutex
    
    // MCP clients per session
    mcpClients map[string]*mcp.Client
}

func NewGuestSessionManager() *GuestSessionManager {
    return &GuestSessionManager{
        sessions:   make(map[string]*EmbodimentSession),
        mcpClients: make(map[string]*mcp.Client),
    }
}

func (gsm *GuestSessionManager) AddSession(session *EmbodimentSession) error {
    gsm.sessionsMux.Lock()
    defer gsm.sessionsMux.Unlock()
    
    // Create MCP client for this session
    mcpClient := mcp.NewClient()
    if err := mcpClient.Connect(session.MCPEndpoint, session.SessionToken); err != nil {
        return fmt.Errorf("failed to connect to MCP endpoint: %w", err)
    }
    
    gsm.sessions[session.SessionToken] = session
    gsm.mcpClients[session.SessionToken] = mcpClient
    
    // Start session renewal monitoring
    go gsm.monitorSessionExpiry(session)
    
    return nil
}

func (gsm *GuestSessionManager) ExecuteTool(sessionToken, toolName string, parameters map[string]interface{}) (interface{}, error) {
    gsm.sessionsMux.RLock()
    session, sessionExists := gsm.sessions[sessionToken]
    mcpClient, clientExists := gsm.mcpClients[sessionToken]
    gsm.sessionsMux.RUnlock()
    
    if !sessionExists || !clientExists {
        return nil, fmt.Errorf("session not found: %s", sessionToken)
    }
    
    if time.Now().After(session.ExpiryTime) {
        return nil, fmt.Errorf("session expired")
    }
    
    // Execute tool via MCP
    result, err := mcpClient.CallTool(toolName, parameters)
    if err != nil {
        return nil, fmt.Errorf("tool execution failed: %w", err)
    }
    
    // Update session context
    session.ActionCount++
    session.Context[fmt.Sprintf("action_%d", session.ActionCount)] = map[string]interface{}{
        "tool":       toolName,
        "parameters": parameters,
        "result":     result,
        "timestamp":  time.Now(),
    }
    
    return result, nil
}

func (gsm *GuestSessionManager) monitorSessionExpiry(session *EmbodimentSession) {
    // Calculate warning time (5 minutes before expiry)
    warningTime := session.ExpiryTime.Add(-5 * time.Minute)
    
    // Wait until warning time
    time.Sleep(time.Until(warningTime))
    
    log.Printf("Session %s will expire in 5 minutes", session.SessionToken)
    
    // Wait until expiry
    time.Sleep(time.Until(session.ExpiryTime))
    
    // Clean up expired session
    gsm.RemoveSession(session.SessionToken)
}

func (gsm *GuestSessionManager) RemoveSession(sessionToken string) {
    gsm.sessionsMux.Lock()
    defer gsm.sessionsMux.Unlock()
    
    if mcpClient, exists := gsm.mcpClients[sessionToken]; exists {
        mcpClient.Disconnect()
        delete(gsm.mcpClients, sessionToken)
    }
    
    delete(gsm.sessions, sessionToken)
    log.Printf("Session %s removed", sessionToken)
}
```

## Production Deployment

### Host Agent Deployment

```dockerfile
# Dockerfile.host
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o fem-host-agent ./cmd/host

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/

COPY --from=builder /app/fem-host-agent .
COPY --from=builder /app/configs/ ./configs/

EXPOSE 8080 8443

CMD ["./fem-host-agent"]
```

### Guest Agent Deployment

```dockerfile
# Dockerfile.guest
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o fem-guest-agent ./cmd/guest

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/

COPY --from=builder /app/fem-guest-agent .
COPY --from=builder /app/configs/ ./configs/

CMD ["./fem-guest-agent"]
```

## Example Implementations

### Complete Cross-Device Development Host

```go
// cmd/host/main.go
package main

import (
    "flag"
    "log"
    "os"
    "os/signal"
    "syscall"
    "time"
    
    "my-host-agent/host"
)

func main() {
    var (
        agentID   = flag.String("agent", "laptop-host-alice", "Host agent ID")
        brokerURL = flag.String("broker", "https://localhost:8443", "Broker URL")
        mcpPort   = flag.Int("mcp-port", 8080, "MCP server port")
    )
    flag.Parse()
    
    // Create host agent
    hostAgent, err := host.NewHostAgent(*agentID, *brokerURL)
    if err != nil {
        log.Fatalf("Failed to create host agent: %v", err)
    }
    
    // Define development workstation body
    if err := hostAgent.DefineDeveloperWorkstation(); err != nil {
        log.Fatalf("Failed to define developer workstation: %v", err)
    }
    
    // Start MCP server
    if err := hostAgent.StartMCPServer(*mcpPort); err != nil {
        log.Fatalf("Failed to start MCP server: %v", err)
    }
    
    // Register with broker
    if err := hostAgent.Register(); err != nil {
        log.Fatalf("Failed to register with broker: %v", err)
    }
    
    log.Printf("Host agent %s started, offering developer workstation body", *agentID)
    
    // Wait for interrupt
    c := make(chan os.Signal, 1)
    signal.Notify(c, os.Interrupt, syscall.SIGTERM)
    <-c
    
    log.Println("Host agent shutting down...")
    hostAgent.Shutdown()
}
```

### Complete Mobile Guest Implementation

```go
// cmd/guest/main.go
package main

import (
    "flag"
    "log"
    "time"
    
    "my-guest-agent/guest"
)

func main() {
    var (
        agentID   = flag.String("agent", "phone-guest-bob", "Guest agent ID")
        brokerURL = flag.String("broker", "https://localhost:8443", "Broker URL")
        goal      = flag.String("goal", "check-projects", "Goal to execute")
    )
    flag.Parse()
    
    // Create guest agent with development preferences
    preferences := guest.DiscoveryPreferences{
        RequiredCapabilities:  []string{"shell.execute", "file.read"},
        PreferredEnvironments: []string{"local-development"},
        MaxSessionDuration:    1800 * time.Second, // 30 minutes
        TrustLevelRequired:    "personal-device",
    }
    
    guestAgent, err := guest.NewGuestAgent(*agentID, *brokerURL, preferences)
    if err != nil {
        log.Fatalf("Failed to create guest agent: %v", err)
    }
    
    // Set development goals based on command line
    switch *goal {
    case "check-projects":
        guestAgent.CreateDevelopmentGoals()
    case "deploy-app":
        guestAgent.CreateDeploymentGoals()
    default:
        log.Fatalf("Unknown goal: %s", *goal)
    }
    
    // Execute current goal
    if err := guestAgent.ExecuteCurrentGoal(); err != nil {
        log.Fatalf("Failed to execute goal: %v", err)
    }
    
    log.Printf("Guest agent %s completed goal: %s", *agentID, *goal)
}
```

### Live2D Host Example

```go
// Example Live2D host body definition
func (h *HostAgent) DefineLive2DAvatarBody() error {
    bodyDef := &BodyDefinition{
        BodyID:          "live2d-puppet-v1",
        Description:     "Virtual avatar control with expression and animation",
        EnvironmentType: "interactive-virtual-world",
        MCPTools: []MCPToolDef{
            {
                Name:        "avatar.set_expression",
                Description: "Change avatar's facial expression",
                InputSchema: map[string]interface{}{
                    "type": "object",
                    "properties": map[string]interface{}{
                        "expression": map[string]interface{}{
                            "type": "string",
                            "enum": []string{"happy", "sad", "surprised", "angry", "neutral"},
                        },
                    },
                },
            },
            {
                Name:        "avatar.speak",
                Description: "Make avatar speak with text-to-speech",
                InputSchema: map[string]interface{}{
                    "type": "object",
                    "properties": map[string]interface{}{
                        "text":    map[string]interface{}{"type": "string", "maxLength": 200},
                        "emotion": map[string]interface{}{"type": "string", "enum": []string{"normal", "excited", "calm"}},
                    },
                },
            },
        },
        SecurityPolicy: SecurityPolicy{
            MaxSessionDuration: 3600 * time.Second,
            ResourceLimits: ResourceLimits{
                MaxAnimationsPerMinute: 30,
                MaxSpeechPerMinute:     10,
            },
        },
        MaxConcurrentGuests: 1,
        RequireApproval:     false,
    }
    
    return h.DefineBody(bodyDef)
}
```

### Configuration Files

```yaml
# configs/host.yaml
host:
  id: "laptop-host-alice"
  broker_url: "https://fem-broker:8443"
  mcp_port: 8080
  
security:
  trust_level: "personal-device"
  require_tls: true
  
bodies:
  developer_workstation:
    enabled: true
    max_concurrent_guests: 2
    session_duration: 3600
  
logging:
  level: "info"
  audit_enabled: true
  audit_path: "/var/log/fem-host-audit.log"
```

```yaml
# configs/guest.yaml
guest:
  id: "phone-guest-bob"
  broker_url: "https://fem-broker:8443"
  
preferences:
  required_capabilities:
    - "shell.execute"
    - "file.read"
  preferred_environments:
    - "local-development"
  max_session_duration: 1800
  trust_level_required: "personal-device"
  
goals:
  default_goal: "check-projects"
  
logging:
  level: "info"
```

This guide provides a complete foundation for building production-ready FEM Protocol agents that enable **Secure Hosted Embodiment**. The examples demonstrate real-world patterns for cross-device development, virtual presence, and collaborative applications.