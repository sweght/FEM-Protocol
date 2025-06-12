# Hosted Embodiment Development Guide

## Introduction

This guide provides comprehensive instructions for building applications using the FEM Protocol's **Secure Hosted Embodiment** paradigm. Whether you're creating a host environment that offers bodies for guest inhabitation, or building guest agents that embody into remote environments, this guide covers the patterns, security considerations, and best practices for hosted embodiment development.

## Table of Contents
- [Core Concepts](#core-concepts)
- [Building Host Applications](#building-host-applications)
- [Building Guest Applications](#building-guest-applications)
- [Security Best Practices](#security-best-practices)
- [Body Definition Patterns](#body-definition-patterns)
- [Session Management](#session-management)
- [Testing and Debugging](#testing-and-debugging)
- [Production Deployment](#production-deployment)

## Core Concepts

### The Host-Guest-Body Model

**Host**: An application that offers "bodies" (sandboxed capability sets) for guest embodiment. Hosts retain ultimate control over their environment and define security boundaries.

**Guest**: An agent "mind" that discovers and inhabits bodies to exercise delegated control. Guests operate within host-defined constraints.

**Body**: A secure, sandboxed set of MCP tools that define what a guest can control within a host environment.

**Embodiment Session**: A time-bounded period during which a guest has active delegated control over a host body.

### Key Principles

1. **Delegated Control**: Guests exercise control on behalf of hosts, not independent control
2. **Security Boundaries**: All guest actions are validated against host-defined policies
3. **Audit Trail**: Every action in an embodiment session is logged for security and debugging
4. **Session Isolation**: Each embodiment session is cryptographically isolated
5. **Graceful Degradation**: Systems continue operating even if embodiment sessions fail

## Building Host Applications

### 1. Define Your Body

The first step is defining what capabilities you're willing to delegate to guests.

```go
type BodyDefinition struct {
    BodyID          string           `json:"bodyId"`
    Description     string           `json:"description"`
    EnvironmentType string           `json:"environmentType"`
    MCPTools        []MCPToolDef     `json:"mcpTools"`
    SecurityPolicy  SecurityPolicy   `json:"securityPolicy"`
    SessionLimits   SessionLimits    `json:"sessionLimits"`
}

// Example: Development workstation body
func CreateDeveloperWorkstationBody() *BodyDefinition {
    return &BodyDefinition{
        BodyID: "dev-workstation-v1",
        Description: "Secure development environment with file and shell access",
        EnvironmentType: "local-development",
        MCPTools: []MCPToolDef{
            {
                Name: "file.read",
                Description: "Read files from project directories",
                InputSchema: fileReadSchema,
                Handler: "handleSecureFileRead",
            },
            {
                Name: "shell.execute", 
                Description: "Execute safe shell commands",
                InputSchema: shellExecuteSchema,
                Handler: "handleSandboxedShell",
            },
        },
        SecurityPolicy: SecurityPolicy{
            AllowedPaths: []string{"/home/user/projects/*"},
            DeniedCommands: []string{"rm -rf", "sudo", "curl"},
            ResourceLimits: ResourceLimits{
                MaxCPUPercent: 25,
                MaxMemoryMB: 500,
            },
        },
        SessionLimits: SessionLimits{
            MaxDuration: 3600, // 1 hour
            MaxConcurrentGuests: 2,
        },
    }
}
```

### 2. Implement Security Validation

Every tool call must be validated against your security policy:

```go
type SecurityEnforcer struct {
    policy SecurityPolicy
    session *EmbodimentSession
}

func (se *SecurityEnforcer) ValidateFileAccess(path string) error {
    // Check if path is in allowed list
    for _, allowedPath := range se.policy.AllowedPaths {
        if matched, _ := filepath.Match(allowedPath, path); matched {
            // Check if path is explicitly denied
            for _, deniedPath := range se.policy.DeniedPaths {
                if matched, _ := filepath.Match(deniedPath, path); matched {
                    return fmt.Errorf("path %s is explicitly denied", path)
                }
            }
            return nil
        }
    }
    return fmt.Errorf("path %s not in allowed paths", path)
}

func (se *SecurityEnforcer) ValidateCommand(command string) error {
    // Check denied commands first
    for _, denied := range se.policy.DeniedCommands {
        if strings.Contains(command, denied) {
            return fmt.Errorf("command contains denied pattern: %s", denied)
        }
    }
    
    // If allow list exists, check it
    if len(se.policy.AllowedCommands) > 0 {
        commandParts := strings.Fields(command)
        if len(commandParts) == 0 {
            return fmt.Errorf("empty command")
        }
        
        baseCommand := commandParts[0]
        for _, allowed := range se.policy.AllowedCommands {
            if baseCommand == allowed {
                return nil
            }
        }
        return fmt.Errorf("command %s not in allowed list", baseCommand)
    }
    
    return nil
}
```

### 3. Implement Tool Handlers

Tool handlers execute guest requests within security boundaries:

```go
func (h *HostAgent) HandleSecureFileRead(req *ToolCallRequest) (*ToolResult, error) {
    // Extract parameters
    path, ok := req.Parameters["path"].(string)
    if !ok {
        return nil, fmt.Errorf("path parameter required")
    }
    
    // Validate security
    if err := h.securityEnforcer.ValidateFileAccess(path); err != nil {
        h.auditLogger.LogViolation(req.SessionToken, "file.read", path, err.Error())
        return nil, fmt.Errorf("access denied: %w", err)
    }
    
    // Execute the operation
    content, err := os.ReadFile(path)
    if err != nil {
        return nil, fmt.Errorf("failed to read file: %w", err)
    }
    
    // Log successful access
    h.auditLogger.LogAccess(req.SessionToken, "file.read", path, len(content))
    
    return &ToolResult{
        Success: true,
        Result: map[string]interface{}{
            "content": string(content),
            "size": len(content),
        },
    }, nil
}

func (h *HostAgent) HandleSandboxedShell(req *ToolCallRequest) (*ToolResult, error) {
    command, ok := req.Parameters["command"].(string)
    if !ok {
        return nil, fmt.Errorf("command parameter required")
    }
    
    // Validate security
    if err := h.securityEnforcer.ValidateCommand(command); err != nil {
        h.auditLogger.LogViolation(req.SessionToken, "shell.execute", command, err.Error())
        return nil, fmt.Errorf("command denied: %w", err)
    }
    
    // Execute in sandbox with resource limits
    result, err := h.sandboxExecutor.Execute(command, h.securityEnforcer.policy.ResourceLimits)
    if err != nil {
        return nil, fmt.Errorf("execution failed: %w", err)
    }
    
    // Log execution
    h.auditLogger.LogExecution(req.SessionToken, "shell.execute", command, result.ExitCode)
    
    return &ToolResult{
        Success: result.ExitCode == 0,
        Result: map[string]interface{}{
            "stdout": result.Stdout,
            "stderr": result.Stderr,
            "exitCode": result.ExitCode,
            "executionTime": result.Duration.Milliseconds(),
        },
    }, nil
}
```

### 4. Session Lifecycle Management

Manage embodiment sessions from request to termination:

```go
type EmbodimentSessionManager struct {
    activeSessions map[string]*EmbodimentSession
    hostAgent      *HostAgent
    broker         *BrokerClient
}

func (esm *EmbodimentSessionManager) HandleEmbodimentRequest(req *EmbodimentRequest) (*EmbodimentResponse, error) {
    // Validate guest identity
    guest, err := esm.broker.VerifyGuestIdentity(req.GuestID)
    if err != nil {
        return nil, fmt.Errorf("guest verification failed: %w", err)
    }
    
    // Check trust level
    body := esm.hostAgent.GetBody(req.BodyID)
    if guest.TrustLevel < body.SecurityPolicy.RequiredTrustLevel {
        return &EmbodimentResponse{
            Success: false,
            Reason: "INSUFFICIENT_TRUST_LEVEL",
            Message: fmt.Sprintf("Guest trust level %s insufficient for body requiring %s", 
                guest.TrustLevel, body.SecurityPolicy.RequiredTrustLevel),
        }, nil
    }
    
    // Check session limits
    if len(esm.activeSessions) >= body.SessionLimits.MaxConcurrentGuests {
        return &EmbodimentResponse{
            Success: false,
            Reason: "SESSION_LIMIT_EXCEEDED",
            Message: "Maximum concurrent guests reached",
        }, nil
    }
    
    // Create session
    session := &EmbodimentSession{
        SessionToken: generateSecureToken(),
        GuestID: req.GuestID,
        BodyID: req.BodyID,
        StartTime: time.Now(),
        ExpiryTime: time.Now().Add(time.Duration(req.RequestedDuration) * time.Second),
        Permissions: body.GetGuestPermissions(guest),
        AuditLog: NewAuditLog(),
    }
    
    esm.activeSessions[session.SessionToken] = session
    
    // Start session monitoring
    go esm.monitorSession(session)
    
    return &EmbodimentResponse{
        Success: true,
        SessionToken: session.SessionToken,
        MCPEndpoint: fmt.Sprintf("%s/sessions/%s", esm.hostAgent.MCPEndpoint, session.SessionToken),
        GrantedPermissions: session.Permissions,
        SessionDuration: req.RequestedDuration,
        ExpiryTime: session.ExpiryTime,
    }, nil
}

func (esm *EmbodimentSessionManager) monitorSession(session *EmbodimentSession) {
    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()
    
    for {
        select {
        case <-ticker.C:
            // Check if session expired
            if time.Now().After(session.ExpiryTime) {
                esm.terminateSession(session.SessionToken, "SESSION_EXPIRED")
                return
            }
            
            // Check resource usage
            if session.ResourceUsage.ExceedsLimits() {
                esm.terminateSession(session.SessionToken, "RESOURCE_LIMIT_EXCEEDED")
                return
            }
            
            // Send heartbeat to guest
            esm.sendSessionUpdate(session, "HEARTBEAT")
            
        case <-session.TerminationChannel:
            return
        }
    }
}
```

## Building Guest Applications

### 1. Discovery and Body Selection

Guests must discover suitable bodies for their needs:

```go
type GuestAgent struct {
    identity *AgentIdentity
    broker   *BrokerClient
    mcpClient *MCPClient
    activeSessions map[string]*EmbodimentClient
}

func (ga *GuestAgent) DiscoverBodies(criteria DiscoveryCriteria) ([]*AvailableBody, error) {
    req := &DiscoveryRequest{
        Query: DiscoveryQuery{
            Capabilities: criteria.RequiredCapabilities,
            EnvironmentType: criteria.PreferredEnvironment,
            TrustLevel: criteria.MinimumTrustLevel,
            MaxResults: 10,
        },
        GuestProfile: GuestProfile{
            GuestID: ga.identity.AgentID,
            IntendedUse: criteria.IntendedUse,
            PreferredSessionDuration: criteria.SessionDuration,
        },
    }
    
    response, err := ga.broker.DiscoverBodies(req)
    if err != nil {
        return nil, fmt.Errorf("discovery failed: %w", err)
    }
    
    return response.AvailableBodies, nil
}

func (ga *GuestAgent) SelectOptimalBody(bodies []*AvailableBody, criteria DiscoveryCriteria) *AvailableBody {
    var best *AvailableBody
    var bestScore float64
    
    for _, body := range bodies {
        score := ga.calculateBodyScore(body, criteria)
        if score > bestScore {
            best = body
            bestScore = score
        }
    }
    
    return best
}

func (ga *GuestAgent) calculateBodyScore(body *AvailableBody, criteria DiscoveryCriteria) float64 {
    score := 0.0
    
    // Capability match score
    capabilityScore := ga.calculateCapabilityMatch(body.Capabilities, criteria.RequiredCapabilities)
    score += capabilityScore * 0.4
    
    // Trust score
    score += body.Availability.TrustScore * 0.2
    
    // Performance score (lower response time is better)
    performanceScore := 1.0 - (float64(body.Availability.AverageResponseTime) / 1000.0)
    score += performanceScore * 0.2
    
    // Availability score
    availabilityScore := 1.0 - (float64(body.Availability.CurrentGuests) / float64(body.Availability.MaxConcurrentGuests))
    score += availabilityScore * 0.2
    
    return score
}
```

### 2. Embodiment Request and Session Establishment

```go
func (ga *GuestAgent) RequestEmbodiment(body *AvailableBody, duration int) (*EmbodimentClient, error) {
    req := &EmbodimentRequest{
        HostAgentID: body.HostAgentID,
        BodyID: body.BodyID,
        RequestedDuration: duration,
        IntendedActions: ga.generateIntendedActions(body),
        GuestCredentials: GuestCredentials{
            GuestID: ga.identity.AgentID,
            TrustLevel: ga.identity.TrustLevel,
            PreviousSessions: len(ga.sessionHistory),
        },
    }
    
    response, err := ga.broker.RequestEmbodiment(req)
    if err != nil {
        return nil, fmt.Errorf("embodiment request failed: %w", err)
    }
    
    if !response.Success {
        return nil, fmt.Errorf("embodiment denied: %s", response.Message)
    }
    
    // Create embodiment client
    client := &EmbodimentClient{
        SessionToken: response.SessionToken,
        HostEndpoint: response.MCPEndpoint,
        Permissions: response.GrantedPermissions,
        ExpiryTime: response.ExpiryTime,
        MCPClient: NewMCPClient(response.MCPEndpoint),
    }
    
    ga.activeSessions[response.SessionToken] = client
    
    // Start session monitoring
    go ga.monitorEmbodimentSession(client)
    
    return client, nil
}
```

### 3. Exercising Delegated Control

Once embodied, guests can execute tools within their granted permissions:

```go
func (ec *EmbodimentClient) ExecuteTool(toolName string, parameters map[string]interface{}) (*ToolResult, error) {
    // Check if we have permission for this tool
    if !ec.hasPermission(toolName, parameters) {
        return nil, fmt.Errorf("permission denied for tool %s", toolName)
    }
    
    // Check session validity
    if time.Now().After(ec.ExpiryTime) {
        return nil, fmt.Errorf("embodiment session expired")
    }
    
    // Execute tool via MCP
    req := &MCPToolCallRequest{
        Method: "tools/call",
        Params: map[string]interface{}{
            "name": toolName,
            "arguments": parameters,
        },
    }
    
    // Add session token to request
    req.Headers = map[string]string{
        "X-Embodiment-Session": ec.SessionToken,
    }
    
    response, err := ec.MCPClient.Call(req)
    if err != nil {
        return nil, fmt.Errorf("tool execution failed: %w", err)
    }
    
    // Update session statistics
    ec.updateSessionStats(toolName, response)
    
    return response.Result, nil
}

func (ec *EmbodimentClient) hasPermission(toolName string, parameters map[string]interface{}) bool {
    for _, permission := range ec.Permissions {
        if ec.matchesPermission(permission, toolName, parameters) {
            return true
        }
    }
    return false
}

func (ec *EmbodimentClient) matchesPermission(permission string, toolName string, parameters map[string]interface{}) bool {
    // Parse permission format: "tool.name:constraint"
    parts := strings.Split(permission, ":")
    permittedTool := parts[0]
    
    // Check tool name match (supports wildcards)
    if matched, _ := filepath.Match(permittedTool, toolName); !matched {
        return false
    }
    
    // Check constraints if present
    if len(parts) > 1 {
        constraint := parts[1]
        return ec.validateConstraint(constraint, parameters)
    }
    
    return true
}
```

### 4. Graceful Session Management

```go
func (ga *GuestAgent) monitorEmbodimentSession(client *EmbodimentClient) {
    ticker := time.NewTicker(60 * time.Second)
    defer ticker.Stop()
    
    for {
        select {
        case <-ticker.C:
            // Check session health
            if err := ga.checkSessionHealth(client); err != nil {
                ga.handleSessionError(client, err)
                return
            }
            
            // Send keepalive if needed
            if time.Until(client.ExpiryTime) < 5*time.Minute {
                ga.requestSessionExtension(client)
            }
            
        case update := <-client.UpdateChannel:
            ga.handleSessionUpdate(client, update)
            
        case <-client.TerminationChannel:
            ga.cleanupSession(client)
            return
        }
    }
}

func (ga *GuestAgent) handleSessionUpdate(client *EmbodimentClient, update *SessionUpdate) {
    switch update.Type {
    case "SESSION_WARNING":
        log.Printf("Session %s: %s", client.SessionToken, update.Message)
        
    case "SESSION_EXPIRED":
        log.Printf("Session %s expired", client.SessionToken)
        ga.terminateSession(client)
        
    case "PERMISSION_REVOKED":
        log.Printf("Permissions revoked for session %s", client.SessionToken)
        ga.updateSessionPermissions(client, update.NewPermissions)
        
    case "HOST_SHUTDOWN":
        log.Printf("Host shutting down, terminating session %s", client.SessionToken)
        ga.terminateSession(client)
    }
}
```

## Security Best Practices

### 1. Host Security

**Principle of Least Privilege**: Only grant the minimum permissions necessary for the guest's intended use.

```go
func (h *HostAgent) generateGuestPermissions(guest *GuestIdentity, body *BodyDefinition) []string {
    permissions := []string{}
    
    // Base permissions for all guests
    for _, tool := range body.MCPTools {
        permission := tool.Name
        
        // Add constraints based on trust level
        if guest.TrustLevel == "basic" {
            permission = h.addBasicConstraints(permission, tool)
        } else if guest.TrustLevel == "verified" {
            permission = h.addVerifiedConstraints(permission, tool)
        }
        
        permissions = append(permissions, permission)
    }
    
    return permissions
}

func (h *HostAgent) addBasicConstraints(permission string, tool MCPToolDef) string {
    switch tool.Name {
    case "file.read":
        return permission + ":/tmp/*,/home/guest/*" // Restrict to safe directories
    case "shell.execute":
        return permission + ":safe-commands-only" // Predefined safe command list
    default:
        return permission
    }
}
```

**Input Validation**: Validate all guest inputs rigorously.

```go
func validateShellCommand(command string, policy SecurityPolicy) error {
    // Check command length
    if len(command) > 1000 {
        return fmt.Errorf("command too long")
    }
    
    // Check for shell injection patterns
    dangerousPatterns := []string{";", "&&", "||", "|", "`", "$(", "${"}
    for _, pattern := range dangerousPatterns {
        if strings.Contains(command, pattern) {
            return fmt.Errorf("potentially dangerous shell construct: %s", pattern)
        }
    }
    
    // Validate against policy
    return policy.ValidateCommand(command)
}
```

**Resource Monitoring**: Continuously monitor resource usage during sessions.

```go
type ResourceMonitor struct {
    session *EmbodimentSession
    limits  ResourceLimits
}

func (rm *ResourceMonitor) checkResourceUsage() error {
    usage := rm.getCurrentUsage()
    
    if usage.CPUPercent > rm.limits.MaxCPUPercent {
        return fmt.Errorf("CPU usage %f%% exceeds limit %f%%", usage.CPUPercent, rm.limits.MaxCPUPercent)
    }
    
    if usage.MemoryMB > rm.limits.MaxMemoryMB {
        return fmt.Errorf("memory usage %dMB exceeds limit %dMB", usage.MemoryMB, rm.limits.MaxMemoryMB)
    }
    
    return nil
}
```

### 2. Guest Security

**Session Validation**: Always validate session state before tool calls.

```go
func (ec *EmbodimentClient) validateSession() error {
    if ec.SessionToken == "" {
        return fmt.Errorf("no active session")
    }
    
    if time.Now().After(ec.ExpiryTime) {
        return fmt.Errorf("session expired")
    }
    
    if ec.Terminated {
        return fmt.Errorf("session terminated")
    }
    
    return nil
}
```

**Error Handling**: Handle host errors gracefully without exposing sensitive information.

```go
func (ec *EmbodimentClient) handleToolError(err error) error {
    // Log full error for debugging
    log.Printf("Tool execution error: %v", err)
    
    // Return sanitized error to caller
    if strings.Contains(err.Error(), "permission denied") {
        return fmt.Errorf("access denied for requested operation")
    } else if strings.Contains(err.Error(), "resource limit") {
        return fmt.Errorf("operation would exceed resource limits")
    } else {
        return fmt.Errorf("operation failed")
    }
}
```

## Body Definition Patterns

### 1. Development Environment Body

```json
{
  "bodyId": "dev-env-secure-v1",
  "description": "Secure development environment with git and file access",
  "environmentType": "development",
  "mcpTools": [
    {
      "name": "file.read",
      "description": "Read project files",
      "inputSchema": {
        "type": "object",
        "properties": {
          "path": {"type": "string"},
          "encoding": {"type": "string", "default": "utf-8"}
        },
        "required": ["path"]
      }
    },
    {
      "name": "git.status",
      "description": "Get repository status",
      "inputSchema": {
        "type": "object",
        "properties": {
          "repo_path": {"type": "string", "default": "."}
        }
      }
    }
  ],
  "securityPolicy": {
    "allowedPaths": ["/workspace/*", "/tmp/scratch/*"],
    "deniedPaths": ["/workspace/.git/config", "/workspace/.env"],
    "allowedCommands": ["git", "ls", "cat", "grep", "find"],
    "resourceLimits": {
      "maxCpuPercent": 20,
      "maxMemoryMB": 256
    }
  }
}
```

### 2. Virtual World Avatar Body

```json
{
  "bodyId": "virtual-avatar-v1",
  "description": "3D avatar control in virtual world",
  "environmentType": "virtual-world",
  "mcpTools": [
    {
      "name": "avatar.move",
      "description": "Move avatar in world",
      "inputSchema": {
        "type": "object",
        "properties": {
          "x": {"type": "number"},
          "y": {"type": "number"},
          "z": {"type": "number"},
          "speed": {"type": "number", "minimum": 0, "maximum": 10}
        }
      }
    },
    {
      "name": "avatar.speak",
      "description": "Make avatar speak",
      "inputSchema": {
        "type": "object",
        "properties": {
          "text": {"type": "string", "maxLength": 200},
          "emotion": {"type": "string", "enum": ["neutral", "happy", "sad"]}
        }
      }
    }
  ],
  "securityPolicy": {
    "resourceLimits": {
      "maxMovementsPerMinute": 60,
      "maxSpeechPerMinute": 10
    },
    "worldConstraints": {
      "allowedAreas": ["public-spaces", "user-homes"],
      "deniedAreas": ["admin-zones", "private-meetings"]
    }
  }
}
```

### 3. Data Processing Pipeline Body

```json
{
  "bodyId": "data-processor-v1",
  "description": "Secure data processing and analysis",
  "environmentType": "data-pipeline",
  "mcpTools": [
    {
      "name": "data.transform",
      "description": "Transform data using approved functions",
      "inputSchema": {
        "type": "object",
        "properties": {
          "input_data": {"type": "string"},
          "transform_type": {
            "type": "string",
            "enum": ["filter", "map", "reduce", "sort"]
          },
          "parameters": {"type": "object"}
        }
      }
    },
    {
      "name": "data.export",
      "description": "Export processed data",
      "inputSchema": {
        "type": "object",
        "properties": {
          "data": {"type": "string"},
          "format": {"type": "string", "enum": ["json", "csv", "xml"]},
          "destination": {"type": "string"}
        }
      }
    }
  ],
  "securityPolicy": {
    "dataConstraints": {
      "maxRecordsPerOperation": 10000,
      "allowedDataTypes": ["anonymized", "public"],
      "deniedDataTypes": ["pii", "confidential"]
    },
    "exportLimits": {
      "maxExportsPerHour": 10,
      "allowedDestinations": ["s3://public-bucket/*"]
    }
  }
}
```

## Session Management

### Session State Tracking

```go
type EmbodimentSession struct {
    SessionToken    string                 `json:"sessionToken"`
    GuestID         string                 `json:"guestId"`
    BodyID          string                 `json:"bodyId"`
    StartTime       time.Time              `json:"startTime"`
    ExpiryTime      time.Time              `json:"expiryTime"`
    LastActivity    time.Time              `json:"lastActivity"`
    Permissions     []string               `json:"permissions"`
    ResourceUsage   ResourceUsage          `json:"resourceUsage"`
    ToolCallCount   int                    `json:"toolCallCount"`
    ViolationCount  int                    `json:"violationCount"`
    AuditEntries    []AuditEntry           `json:"auditEntries"`
    Status          SessionStatus          `json:"status"`
    Metadata        map[string]interface{} `json:"metadata"`
}

type SessionStatus string
const (
    SessionActive     SessionStatus = "active"
    SessionWarning    SessionStatus = "warning"
    SessionSuspended  SessionStatus = "suspended"
    SessionExpired    SessionStatus = "expired"
    SessionTerminated SessionStatus = "terminated"
)
```

### Session Extension

```go
func (esm *EmbodimentSessionManager) HandleSessionExtension(req *SessionExtensionRequest) error {
    session, exists := esm.activeSessions[req.SessionToken]
    if !exists {
        return fmt.Errorf("session not found")
    }
    
    // Check if extension is allowed
    body := esm.hostAgent.GetBody(session.BodyID)
    maxExtension := body.SessionLimits.MaxExtensionDuration
    
    if req.ExtensionDuration > maxExtension {
        return fmt.Errorf("extension duration %d exceeds maximum %d", 
            req.ExtensionDuration, maxExtension)
    }
    
    // Check session health
    if session.ViolationCount > body.SessionLimits.MaxViolations {
        return fmt.Errorf("session has too many violations for extension")
    }
    
    // Grant extension
    session.ExpiryTime = session.ExpiryTime.Add(time.Duration(req.ExtensionDuration) * time.Second)
    
    // Notify guest
    esm.sendSessionUpdate(session, &SessionUpdate{
        Type: "SESSION_EXTENDED",
        Message: fmt.Sprintf("Session extended by %d seconds", req.ExtensionDuration),
        NewExpiryTime: session.ExpiryTime,
    })
    
    return nil
}
```

## Testing and Debugging

### 1. Unit Testing Host Bodies

```go
func TestSecureFileRead(t *testing.T) {
    host := NewTestHostAgent()
    session := CreateTestSession("test-guest", "dev-env-v1")
    
    tests := []struct {
        name        string
        path        string
        expectError bool
        errorType   string
    }{
        {
            name: "allowed path",
            path: "/workspace/main.go",
            expectError: false,
        },
        {
            name: "denied path",
            path: "/workspace/.env",
            expectError: true,
            errorType: "access denied",
        },
        {
            name: "path traversal attempt",
            path: "/workspace/../etc/passwd",
            expectError: true,
            errorType: "access denied",
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            req := &ToolCallRequest{
                SessionToken: session.SessionToken,
                Tool: "file.read",
                Parameters: map[string]interface{}{
                    "path": tt.path,
                },
            }
            
            result, err := host.HandleToolCall(req)
            
            if tt.expectError {
                assert.Error(t, err)
                assert.Contains(t, err.Error(), tt.errorType)
            } else {
                assert.NoError(t, err)
                assert.True(t, result.Success)
            }
        })
    }
}
```

### 2. Integration Testing

```go
func TestFullEmbodimentFlow(t *testing.T) {
    // Setup test environment
    broker := NewTestBroker()
    host := NewTestHostAgent()
    guest := NewTestGuestAgent()
    
    // Register host with broker
    err := host.RegisterWithBroker(broker)
    assert.NoError(t, err)
    
    // Guest discovers bodies
    bodies, err := guest.DiscoverBodies(DiscoveryCriteria{
        RequiredCapabilities: []string{"file.read"},
        EnvironmentType: "development",
    })
    assert.NoError(t, err)
    assert.Len(t, bodies, 1)
    
    // Guest requests embodiment
    client, err := guest.RequestEmbodiment(bodies[0], 3600)
    assert.NoError(t, err)
    assert.NotNil(t, client)
    
    // Execute tool call
    result, err := client.ExecuteTool("file.read", map[string]interface{}{
        "path": "/workspace/test.txt",
    })
    assert.NoError(t, err)
    assert.True(t, result.Success)
    
    // Cleanup
    err = guest.TerminateSession(client.SessionToken)
    assert.NoError(t, err)
}
```

### 3. Security Testing

```go
func TestSecurityViolationHandling(t *testing.T) {
    host := NewTestHostAgent()
    session := CreateTestSession("malicious-guest", "dev-env-v1")
    
    // Attempt path traversal
    req := &ToolCallRequest{
        SessionToken: session.SessionToken,
        Tool: "file.read",
        Parameters: map[string]interface{}{
            "path": "../../../../etc/passwd",
        },
    }
    
    result, err := host.HandleToolCall(req)
    
    // Should be denied
    assert.Error(t, err)
    assert.Nil(t, result)
    
    // Check that violation was logged
    violations := host.auditLogger.GetViolations(session.SessionToken)
    assert.Len(t, violations, 1)
    assert.Equal(t, "file.read", violations[0].Tool)
    
    // Check session status
    sessionStatus := host.sessionManager.GetSessionStatus(session.SessionToken)
    assert.Equal(t, SessionWarning, sessionStatus.Status)
}
```

## Production Deployment

### 1. Monitoring and Observability

```go
type EmbodimentMetrics struct {
    ActiveSessions      prometheus.Gauge
    SessionsStarted     prometheus.Counter
    SessionsTerminated  prometheus.Counter
    ToolCallsTotal      prometheus.Counter
    SecurityViolations  prometheus.Counter
    ResourceUsage       prometheus.Histogram
}

func (h *HostAgent) recordMetrics() {
    h.metrics.ActiveSessions.Set(float64(len(h.sessionManager.activeSessions)))
    
    // Record resource usage for each session
    for _, session := range h.sessionManager.activeSessions {
        h.metrics.ResourceUsage.Observe(session.ResourceUsage.CPUPercent)
    }
}

func (h *HostAgent) logSecurityViolation(session *EmbodimentSession, violation SecurityViolation) {
    h.metrics.SecurityViolations.Inc()
    
    log.Warn().
        Str("sessionToken", session.SessionToken).
        Str("guestId", session.GuestID).
        Str("violationType", violation.Type).
        Str("details", violation.Details).
        Msg("Security violation detected")
    
    // Alert security team for serious violations
    if violation.Severity == "critical" {
        h.alertManager.SendAlert(AlertCriticalSecurityViolation, violation)
    }
}
```

### 2. Configuration Management

```yaml
# Production host configuration
embodiment:
  host:
    max_concurrent_sessions: 10
    default_session_duration: 3600
    max_session_duration: 7200
    violation_threshold: 3
    
  security:
    require_tls: true
    min_guest_trust_level: "verified"
    enable_audit_logging: true
    audit_retention_days: 90
    
  resource_limits:
    default_cpu_percent: 10
    default_memory_mb: 256
    default_disk_mb: 100
    
  monitoring:
    metrics_enabled: true
    health_check_interval: 30s
    session_warning_threshold: 300s
```

### 3. Disaster Recovery

```go
type SessionPersistence struct {
    store SessionStore
}

func (sp *SessionPersistence) SaveSession(session *EmbodimentSession) error {
    data, err := json.Marshal(session)
    if err != nil {
        return err
    }
    
    return sp.store.Put(session.SessionToken, data)
}

func (sp *SessionPersistence) RestoreSession(sessionToken string) (*EmbodimentSession, error) {
    data, err := sp.store.Get(sessionToken)
    if err != nil {
        return nil, err
    }
    
    var session EmbodimentSession
    err = json.Unmarshal(data, &session)
    return &session, err
}

func (h *HostAgent) handleBrokerReconnection() {
    // Restore active sessions from persistence
    sessionTokens, err := h.persistence.ListActiveSessions()
    if err != nil {
        log.Error().Err(err).Msg("Failed to list active sessions")
        return
    }
    
    for _, token := range sessionTokens {
        session, err := h.persistence.RestoreSession(token)
        if err != nil {
            log.Error().Str("sessionToken", token).Err(err).Msg("Failed to restore session")
            continue
        }
        
        // Check if session should still be active
        if time.Now().After(session.ExpiryTime) {
            h.persistence.DeleteSession(token)
            continue
        }
        
        // Re-register session
        h.sessionManager.activeSessions[token] = session
        log.Info().Str("sessionToken", token).Msg("Restored embodiment session")
    }
}
```

This guide provides the foundation for building robust, secure hosted embodiment applications using the FEM Protocol. The key is to always maintain strong security boundaries while providing rich, collaborative experiences for both hosts and guests.