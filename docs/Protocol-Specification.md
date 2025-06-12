# FEM Protocol Specification

## Table of Contents
- [Overview](#overview)
- [Protocol Fundamentals](#protocol-fundamentals)
- [Message Envelopes](#message-envelopes)
- [Security Model](#security-model)
- [Embodiment Framework](#embodiment-framework)
- [Transport Layer](#transport-layer)
- [Agent Lifecycle](#agent-lifecycle)
- [Broker Operations](#broker-operations)
- [Error Handling](#error-handling)
- [Examples](#examples)

## Overview

The **FEM Protocol** is a wire-level protocol designed for **Secure Hosted Embodiment** between AI agents in a federated network. The protocol enables a new paradigm where guest "minds" can securely inhabit "bodies" offered by host environments, creating a model of **Secure Delegated Control**.

**Core Innovation**: FEM Protocol moves beyond simple tool federation to enable persistent embodiment sessions where guests exercise delegated control over host-defined capabilities within cryptographically enforced security boundaries.

### Key Design Principles

1. **Hosted Embodiment**: Guests can inhabit and control host-offered bodies (capability sets)
2. **Secure Delegated Control**: Hosts delegate specific control to guests within defined boundaries
3. **Cryptographic Security**: All interactions are Ed25519 signed with fine-grained permissions
4. **Broker-as-Agent**: Brokers are first-class agents with their own mind, body, and environment
5. **MCP Integration**: Seamless integration with Model Context Protocol for tool interfaces
6. **Environment Awareness**: Bodies adapt to deployment environments (local, cloud, edge, etc.)
7. **Federation**: Multi-broker networks enable cross-organizational embodiment
8. **Zero-Trust**: Cryptographic proof of identity required for all embodiment sessions

### Protocol Version

Current version: **v0.3.0**

## Protocol Fundamentals

### Core Concepts

**Host**: An agent that offers "bodies" (sandboxed capability sets) for guest embodiment. Hosts define security boundaries and retain ultimate control over their environment.

**Guest**: An agent "mind" that can discover and inhabit bodies offered by hosts, exercising delegated control within host-defined boundaries.

**Body**: A secure, sandboxed set of MCP tools and capabilities offered by a host for guest embodiment. Bodies define what a guest can control.

**Embodiment**: The process by which a guest mind inhabits a host body, establishing a persistent session with delegated control capabilities.

**Broker**: A first-class agent that coordinates embodiment discovery, security verification, and federation. Brokers have their own mind (logic), body (network tools), and environment (deployment context).

**Envelope**: A structured, cryptographically signed message format for all FEM Protocol communications.

**Security Policy**: Host-defined rules that govern what guests can do within an embodied session (file paths, commands, resources, time limits).

**Embodiment Session**: A time-bounded period during which a guest has active control over a host body, with all actions logged and audited.

### The Three Flagship Use Cases

**1. Collaborative Virtual Presence (Live2D Guest System)**
```
Guest Agent → Live2D Host → Avatar Body
• Guest calls: avatar.set_expression("happy")
• Host validates and applies to avatar state
• Security: Guest controls avatar only, no file system access
```

**2. Collaborative Application Control (Interactive Storyteller)**
```
Guest Narrative AI → Storytelling Host → Game State Body
• Guest calls: update_world("A mysterious fog rolls in...")
• Host validates and updates game state, UI re-renders
• Security: Guest modifies game state only through defined tools
```

**3. Cross-Device Embodiment (Phone ↔ Laptop)**
```
Phone Guest → Laptop Host → Developer Terminal Body
• Guest calls: shell.execute("git status")
• Host executes in sandboxed environment
• Security: Guest limited to defined paths and safe commands
```

### Message Flow

1. **Discovery**: Guest discovers available bodies through broker
2. **Embodiment Request**: Guest requests to inhabit specific body
3. **Security Verification**: Host verifies guest identity and policies
4. **Session Establishment**: Host grants embodiment session with specific permissions
5. **Delegated Control**: Guest exercises control through host body's MCP tools
6. **Session Management**: Ongoing audit, monitoring, and session lifecycle
7. **Session Termination**: Graceful or forced termination with cleanup

## Message Envelopes

All FEM Protocol communication uses cryptographically signed envelopes. Every envelope includes identity verification and replay protection.

### Common Envelope Structure

```json
{
  "type": "envelopeType",
  "agent": "agent-identifier",
  "ts": 1641234567890,
  "nonce": "unique-replay-guard",
  "sig": "base64-encoded-ed25519-signature",
  "body": {
    // Type-specific content
  }
}
```

### Common Headers

- **type**: The envelope type (see envelope types below)
- **agent**: UTF-8 string identifying the sending agent
- **ts**: Unix timestamp in milliseconds when envelope was created
- **nonce**: Unique string to prevent replay attacks (cryptographically random)
- **sig**: Base64-encoded Ed25519 signature of entire envelope (excluding sig field)
- **body**: Type-specific message content

### Envelope Types

The FEM Protocol defines ten core envelope types optimized for hosted embodiment:

#### 1. registerAgent

Registers a new agent with embodiment capabilities.

```json
{
  "type": "registerAgent",
  "agent": "laptop-host-alice",
  "ts": 1641234567890,
  "nonce": "reg-12345-67890",
  "sig": "Mf8B7tKqE...",
  "body": {
    "pubkey": "base64-encoded-ed25519-public-key",
    "agentType": "host",
    "capabilities": ["terminal.shell", "file.operations", "code.execution"],
    "offeredBodies": [
      {
        "bodyId": "developer-workstation-v1",
        "description": "Secure development environment with file and shell access",
        "environmentType": "local-development",
        "mcpTools": [
          {
            "name": "shell.execute",
            "description": "Execute shell commands in sandbox",
            "inputSchema": {
              "type": "object",
              "properties": {
                "command": {"type": "string"},
                "workdir": {"type": "string", "default": "/home/alice/projects"}
              }
            }
          },
          {
            "name": "file.read",
            "description": "Read files from allowed paths",
            "inputSchema": {
              "type": "object", 
              "properties": {
                "path": {"type": "string"}
              }
            }
          }
        ],
        "securityPolicy": {
          "allowedPaths": ["/home/alice/projects/*"],
          "deniedCommands": ["rm -rf", "sudo", "curl"],
          "maxSessionDuration": 3600,
          "maxConcurrentGuests": 2
        }
      }
    ],
    "mcpEndpoint": "https://alice-laptop:8080/mcp",
    "metadata": {
      "version": "1.0.0",
      "description": "Alice's development laptop",
      "trustLevel": "personal-device"
    }
  }
}
```

**Body Fields**:
- `pubkey`: Agent's Ed25519 public key for signature verification
- `agentType`: "host", "guest", or "broker" - defines agent's primary role
- `capabilities`: Array of capabilities this agent provides
- `offeredBodies`: Array of body definitions this host offers for embodiment
- `mcpEndpoint`: HTTP URL where the agent's MCP server is accessible
- `metadata`: Additional agent information and trust indicators

#### 2. registerBroker

Registers a broker agent for federation (brokers are first-class agents).

```json
{
  "type": "registerBroker",
  "agent": "broker-west-coast",
  "ts": 1641234567890,
  "nonce": "broker-reg-98765",
  "sig": "Kl9A8uLpF...",
  "body": {
    "brokerId": "broker-west-coast",
    "endpoint": "https://west.fem-network.com:8443",
    "pubkey": "base64-encoded-ed25519-public-key",
    "brokerCapabilities": ["embodiment.coordination", "federation.routing", "security.verification"],
    "supportedEnvironments": ["cloud-aws", "edge-devices", "local-development"],
    "federationPolicy": {
      "trustLevel": "verified-organization",
      "allowedAgentTypes": ["host", "guest"],
      "maxFederatedSessions": 100
    }
  }
}
```

**Body Fields**:
- `brokerId`: Unique identifier for the broker agent
- `endpoint`: TLS endpoint where the broker can be reached
- `pubkey`: Broker's Ed25519 public key for federation
- `brokerCapabilities`: Broker-level capabilities for network management
- `supportedEnvironments`: Environment types this broker supports
- `federationPolicy`: Rules for cross-broker embodiment

#### 3. discoverBodies

Requests discovery of available bodies for embodiment.

```json
{
  "type": "discoverBodies", 
  "agent": "phone-guest-bob",
  "ts": 1641234567890,
  "nonce": "discover-66666",
  "sig": "Up2H7aRvL...",
  "body": {
    "query": {
      "capabilities": ["terminal.*", "file.read"],
      "environmentType": "local-development",
      "trustLevel": "personal-device",
      "maxResults": 10,
      "includeSecurityPolicies": true
    },
    "guestProfile": {
      "guestId": "phone-guest-bob",
      "preferredSessionDuration": 1800,
      "intendedUse": "mobile-development-access"
    },
    "requestId": "discovery-req-001"
  }
}
```

**Body Fields**:
- `query`: Discovery criteria for finding suitable bodies
- `guestProfile`: Information about the requesting guest
- `requestId`: Unique identifier for correlation

#### 4. bodiesDiscovered

Response containing available bodies matching the discovery query.

```json
{
  "type": "bodiesDiscovered",
  "agent": "broker-central",
  "ts": 1641234567890,
  "nonce": "discovered-77777", 
  "sig": "Vq3I8bSwM...",
  "body": {
    "requestId": "discovery-req-001",
    "availableBodies": [
      {
        "hostAgentId": "laptop-host-alice",
        "bodyId": "developer-workstation-v1",
        "description": "Secure development environment with file and shell access",
        "mcpEndpoint": "https://alice-laptop:8080/mcp",
        "capabilities": ["shell.execute", "file.read", "file.write"],
        "environmentType": "local-development",
        "securityPolicy": {
          "allowedPaths": ["/home/alice/projects/*"],
          "maxSessionDuration": 3600,
          "requiresApproval": false
        },
        "availability": {
          "currentGuests": 0,
          "maxConcurrentGuests": 2,
          "averageResponseTime": 45,
          "trustScore": 0.95
        }
      }
    ],
    "totalResults": 1,
    "hasMore": false
  }
}
```

**Body Fields**:
- `requestId`: Correlates with original discovery request
- `availableBodies`: Array of bodies available for embodiment
- `totalResults`: Total number of matching bodies found
- `hasMore`: Whether additional results are available

#### 5. requestEmbodiment

Guest requests to inhabit a specific host body.

```json
{
  "type": "requestEmbodiment",
  "agent": "phone-guest-bob",
  "ts": 1641234567890,
  "nonce": "embody-request-88888",
  "sig": "Wr4J9cTxN...",
  "body": {
    "hostAgentId": "laptop-host-alice",
    "bodyId": "developer-workstation-v1", 
    "requestedDuration": 1800,
    "intendedActions": [
      "Check git status of ongoing projects",
      "Run development servers for testing",
      "Read configuration files"
    ],
    "guestCredentials": {
      "guestId": "phone-guest-bob",
      "trustLevel": "verified-user",
      "previousSessions": 5,
      "averageSessionRating": 4.8
    },
    "requestId": "embodiment-req-001"
  }
}
```

**Body Fields**:
- `hostAgentId`: Target host agent for embodiment
- `bodyId`: Specific body to inhabit
- `requestedDuration`: Desired session length in seconds
- `intendedActions`: Description of planned activities
- `guestCredentials`: Guest's trust and history information
- `requestId`: Unique request identifier

#### 6. embodimentGranted

Host grants embodiment access to guest.

```json
{
  "type": "embodimentGranted",
  "agent": "laptop-host-alice",
  "ts": 1641234567890,
  "nonce": "granted-99999",
  "sig": "Xs5K0dUyO...",
  "body": {
    "requestId": "embodiment-req-001",
    "guestId": "phone-guest-bob",
    "sessionToken": "sess-abc123-def456-ghi789",
    "sessionDuration": 1800,
    "mcpEndpoint": "https://alice-laptop:8080/mcp/sessions/sess-abc123-def456-ghi789",
    "grantedPermissions": [
      "shell.execute:/home/alice/projects/*",
      "file.read:/home/alice/projects/*", 
      "file.write:/home/alice/projects/*"
    ],
    "securityConstraints": {
      "allowedPaths": ["/home/alice/projects/*"],
      "deniedCommands": ["rm -rf", "sudo", "curl"],
      "resourceLimits": {
        "maxCpuPercent": 25,
        "maxMemoryMB": 500,
        "maxDiskWriteMB": 100
      }
    },
    "sessionExpiry": 1641236367890,
    "auditLogId": "audit-session-001"
  }
}
```

**Body Fields**:
- `requestId`: Correlates with embodiment request
- `sessionToken`: Unique token for this embodiment session
- `sessionDuration`: Actual granted session duration
- `mcpEndpoint`: Session-specific MCP endpoint for tool calls
- `grantedPermissions`: Specific permissions granted to guest
- `securityConstraints`: Active security policies for the session
- `sessionExpiry`: Unix timestamp when session expires
- `auditLogId`: Identifier for audit trail

#### 7. embodimentDenied

Host denies embodiment request.

```json
{
  "type": "embodimentDenied",
  "agent": "laptop-host-alice",
  "ts": 1641234567890,
  "nonce": "denied-10101",
  "sig": "Yt6L1eVzP...",
  "body": {
    "requestId": "embodiment-req-001",
    "guestId": "phone-guest-bob",
    "reason": "HOST_POLICY_VIOLATION",
    "message": "Guest trust level insufficient for requested body",
    "retryAllowed": false,
    "suggestedAlternatives": [
      {
        "bodyId": "basic-terminal-v1",
        "description": "Limited terminal access with basic commands only"
      }
    ]
  }
}
```

**Body Fields**:
- `requestId`: Correlates with embodiment request
- `reason`: Machine-readable denial reason code
- `message`: Human-readable explanation
- `retryAllowed`: Whether guest can retry the request
- `suggestedAlternatives`: Other bodies the guest might access

#### 8. toolCall

Executes a tool within an active embodiment session.

```json
{
  "type": "toolCall",
  "agent": "phone-guest-bob",
  "ts": 1641234567890,
  "nonce": "tool-call-11111",
  "sig": "Zu7M2fWaQ...",
  "body": {
    "sessionToken": "sess-abc123-def456-ghi789",
    "tool": "shell.execute",
    "parameters": {
      "command": "git status",
      "workdir": "/home/alice/projects/my-app"
    },
    "requestId": "tool-exec-001"
  }
}
```

**Body Fields**:
- `sessionToken`: Active embodiment session token
- `tool`: Tool name to execute within the body
- `parameters`: Tool-specific parameters
- `requestId`: Unique identifier for result correlation

#### 9. toolResult

Returns result of tool execution within embodiment session.

```json
{
  "type": "toolResult",
  "agent": "laptop-host-alice",
  "ts": 1641234567890,
  "nonce": "tool-result-12121",
  "sig": "Av8N3gXbR...",
  "body": {
    "requestId": "tool-exec-001",
    "sessionToken": "sess-abc123-def456-ghi789",
    "success": true,
    "result": {
      "stdout": "On branch main\nYour branch is up to date with 'origin/main'.\n\nnothing to commit, working tree clean",
      "stderr": "",
      "exitCode": 0,
      "executionTime": 0.234
    },
    "securityValidation": {
      "pathChecked": "/home/alice/projects/my-app",
      "commandFiltered": false,
      "resourceUsage": {
        "cpuPercent": 2.1,
        "memoryMB": 15,
        "diskReadMB": 0.1
      }
    },
    "auditEntry": "audit-action-001"
  }
}
```

**Body Fields**:
- `requestId`: Correlates with tool call
- `sessionToken`: Active embodiment session
- `success`: Whether tool execution succeeded
- `result`: Tool execution results
- `securityValidation`: Security checks performed
- `auditEntry`: Audit log entry identifier

#### 10. embodimentUpdate

Notifies of changes to agent embodiment or session status.

```json
{
  "type": "embodimentUpdate",
  "agent": "laptop-host-alice",
  "ts": 1641234567890,
  "nonce": "update-13131",
  "sig": "Bw9O4hYcS...",
  "body": {
    "updateType": "SESSION_WARNING",
    "sessionToken": "sess-abc123-def456-ghi789",
    "guestId": "phone-guest-bob",
    "message": "Session will expire in 5 minutes",
    "details": {
      "currentSessionTime": 1620,
      "remainingTime": 300,
      "actionsInSession": 15,
      "extensionAvailable": true
    }
  }
}
```

**Body Fields**:
- `updateType`: Type of update (SESSION_WARNING, SESSION_EXPIRED, PERMISSIONS_CHANGED, etc.)
- `sessionToken`: Affected embodiment session
- `guestId`: Guest agent being notified
- `message`: Human-readable update message
- `details`: Update-specific additional information

## Security Model

The FEM Protocol implements a comprehensive security model designed specifically for **Secure Delegated Control** scenarios.

### Cryptographic Foundation

**Algorithm**: Ed25519 (Edwards-curve Digital Signature Algorithm)
- **Key Size**: 32 bytes private key, 32 bytes public key
- **Signature Size**: 64 bytes
- **Security Level**: ~128-bit security equivalent
- **Performance**: ~70,000 signatures/second, ~25,000 verifications/second

### Signature Process

1. **Envelope Creation**: Agent creates envelope with all fields except `sig`
2. **Canonical Serialization**: Envelope serialized to deterministic JSON
3. **Signing**: Agent signs serialized data with Ed25519 private key
4. **Encoding**: Signature is base64-encoded and added to `sig` field

### Verification Process

1. **Signature Extraction**: Receiver extracts `sig` field
2. **Envelope Reconstruction**: Temporarily removes `sig` field
3. **Canonical Serialization**: Envelope serialized identically
4. **Verification**: Signature verified against agent's known public key

### Embodiment Session Security

**Session Tokens**: Cryptographically random tokens that identify active embodiment sessions
- 256-bit entropy
- Unique per session
- Required for all tool calls within embodied sessions
- Automatically expire at session end

**Permission Enforcement**: Every tool call is validated against session permissions
- Path-based restrictions for file operations
- Command filtering for shell execution  
- Resource limits enforced in real-time
- Action logging for audit trail

**Security Policies**: Host-defined rules that govern guest behavior
```go
type SecurityPolicy struct {
    AllowedPaths     []string      `json:"allowedPaths"`
    DeniedPaths      []string      `json:"deniedPaths"`
    AllowedCommands  []string      `json:"allowedCommands"`
    DeniedCommands   []string      `json:"deniedCommands"`
    ResourceLimits   ResourceLimit `json:"resourceLimits"`
    SessionTimeout   time.Duration `json:"sessionTimeout"`
    RequireApproval  bool          `json:"requireApproval"`
}
```

### Trust and Reputation

**Trust Levels**: Hierarchical trust system for embodiment decisions
- `unknown`: No prior interaction history
- `basic`: Limited successful interactions
- `verified`: Significant positive history
- `trusted`: Long-term reliable behavior
- `personal`: Personal devices and known entities

**Reputation Tracking**: Ongoing assessment of guest behavior
- Session completion rates
- Policy compliance history
- Resource usage patterns
- Host feedback scores

## Embodiment Framework

### Host Body Definitions

Bodies define the complete embodiment experience:

```go
type BodyDefinition struct {
    BodyID          string           `json:"bodyId"`
    Description     string           `json:"description"`
    EnvironmentType string           `json:"environmentType"`
    
    // Tool capabilities offered to guests
    MCPTools        []MCPToolDef     `json:"mcpTools"`
    
    // Security boundaries
    SecurityPolicy  SecurityPolicy   `json:"securityPolicy"`
    
    // Session management
    MaxConcurrentGuests int          `json:"maxConcurrentGuests"`
    DefaultSessionDuration time.Duration `json:"defaultSessionDuration"`
    MaxSessionDuration time.Duration `json:"maxSessionDuration"`
    
    // Host preferences
    RequireApproval bool             `json:"requireApproval"`
    TrustLevelRequired string        `json:"trustLevelRequired"`
}
```

### Guest Discovery and Selection

Guests can discover bodies using rich criteria:
- **Capability Matching**: Find bodies offering specific tools
- **Environment Filtering**: Match deployment contexts
- **Trust Requirements**: Filter by required trust levels
- **Resource Needs**: Match computational requirements
- **Geographic Preferences**: Latency and jurisdiction considerations

### Session Lifecycle Management

**1. Discovery Phase**
- Guest searches for suitable bodies
- Broker returns matching hosts with availability
- Guest evaluates options based on needs

**2. Request Phase**
- Guest submits embodiment request with intentions
- Host evaluates request against policies
- Host grants, denies, or suggests alternatives

**3. Active Embodiment**
- Guest receives session token and MCP endpoint
- All tool calls validated against session permissions
- Host monitors resource usage and behavior
- Continuous audit logging of all actions

**4. Session Termination**
- Natural expiration at timeout
- Guest-initiated graceful exit
- Host-initiated termination for policy violations
- Emergency revocation by broker

## Transport Layer

The FEM Protocol is designed for secure, reliable transport with specific requirements for embodiment sessions.

### Required Transport Properties

1. **Reliability**: Message delivery and ordering guarantees
2. **Confidentiality**: TLS 1.3+ encryption for all communications
3. **Integrity**: Transport-level integrity verification
4. **Performance**: Low latency for interactive embodiment sessions

### HTTPS Transport (Primary)

**Endpoint Structure**:
- Broker: `https://broker.example.com:8443/fem`
- Host MCP: `https://host.example.com:8080/mcp`
- Session-specific: `https://host.example.com:8080/mcp/sessions/{sessionToken}`

**Request Format**:
```http
POST /fem HTTP/1.1
Host: broker.example.com
Content-Type: application/json
User-Agent: FEM-Protocol/0.3.0
Content-Length: [length]

{envelope-json}
```

**Response Format**:
```http
HTTP/1.1 200 OK
Content-Type: application/json
X-FEM-Protocol-Version: 0.3.0

{
  "status": "success",
  "message": "Envelope processed",
  "requestId": "req-001"
}
```

### WebSocket Transport (Real-time Sessions)

For long-lived embodiment sessions, WebSocket connections provide:
- Real-time tool execution
- Session status updates
- Low-latency interaction
- Bidirectional communication

## Agent Lifecycle

### Host Agent Lifecycle

**1. Initialization**
- Generate Ed25519 keypair
- Define body definitions for offering
- Configure security policies
- Start MCP server

**2. Registration**
- Register with broker as host agent
- Advertise available bodies
- Declare capabilities and trust requirements

**3. Embodiment Hosting**
- Receive embodiment requests
- Evaluate against security policies
- Grant/deny sessions with appropriate permissions
- Monitor and audit guest activities

**4. Session Management**
- Validate all guest tool calls
- Enforce resource limits
- Log all actions for audit
- Handle session expiration and cleanup

### Guest Agent Lifecycle

**1. Initialization**
- Generate Ed25519 keypair
- Define embodiment preferences
- Configure discovery criteria

**2. Discovery**
- Search for suitable bodies
- Evaluate host offerings
- Select optimal embodiment targets

**3. Embodiment**
- Request access to desired bodies
- Receive session tokens and permissions
- Begin delegated control activities

**4. Active Session**
- Execute tools within granted permissions
- Respect host security policies
- Maintain session through activity
- Gracefully terminate when complete

### Broker Agent Lifecycle

**1. Initialization**
- Generate Ed25519 keypair
- Configure embodiment policies
- Initialize federation capabilities
- Start network services

**2. Network Coordination**
- Accept host and guest registrations
- Facilitate embodiment discovery
- Route cross-broker federation requests
- Monitor network health

**3. Security Enforcement**
- Verify all message signatures
- Validate embodiment requests
- Enforce capability boundaries
- Maintain audit logs

**4. Federation Management**
- Connect with peer brokers
- Share embodiment opportunities
- Route cross-broker sessions
- Maintain federation health

## Broker Operations

### Embodiment Coordination

Brokers serve as trusted intermediaries for embodiment sessions:

**Discovery Coordination**:
1. Maintain registry of host-offered bodies
2. Index by capabilities, environment, trust level
3. Process guest discovery queries with matching
4. Return ranked results based on availability and fit

**Security Verification**:
1. Verify guest identity and reputation
2. Validate host security policies
3. Mediate trust negotiations
4. Issue session approvals

**Session Monitoring**:
1. Track active embodiment sessions
2. Monitor for policy violations
3. Handle session disputes
4. Coordinate emergency terminations

### Federation Protocol

**Cross-Broker Embodiment**:
1. Broker A receives guest discovery request
2. Query local hosts + federated brokers
3. Aggregate and rank all available bodies
4. Return unified results to guest
5. Route embodiment requests to appropriate brokers
6. Coordinate cross-broker session management

**Federation Health**:
- Heartbeat monitoring between brokers
- Failover handling for broker outages
- Load balancing across federation
- Security policy synchronization

## Error Handling

### Embodiment-Specific Errors

**Discovery Errors**:
- `NO_BODIES_AVAILABLE`: No hosts match discovery criteria
- `DISCOVERY_TIMEOUT`: Discovery request timed out
- `INSUFFICIENT_TRUST`: Guest trust level too low

**Embodiment Errors**:
- `EMBODIMENT_DENIED`: Host denied embodiment request
- `SESSION_LIMIT_EXCEEDED`: Too many concurrent sessions
- `SECURITY_POLICY_VIOLATION`: Request violates host policies
- `HOST_UNAVAILABLE`: Target host is offline or overloaded

**Session Errors**:
- `INVALID_SESSION_TOKEN`: Session token invalid or expired
- `PERMISSION_DENIED`: Tool call exceeds granted permissions
- `RESOURCE_LIMIT_EXCEEDED`: Action would exceed resource limits
- `SESSION_EXPIRED`: Session has reached timeout

### Error Response Format

```json
{
  "status": "error",
  "code": "EMBODIMENT_DENIED",
  "message": "Guest trust level insufficient for requested body",
  "details": {
    "guestId": "phone-guest-bob",
    "requiredTrustLevel": "verified",
    "actualTrustLevel": "basic",
    "hostPolicy": "security-first",
    "retryAllowed": false
  },
  "suggestedActions": [
    "Build trust through smaller embodiment sessions",
    "Request basic-access body instead"
  ]
}
```

## Examples

### Complete Cross-Device Embodiment Flow

```json
// 1. Guest discovers available terminal bodies
{
  "type": "discoverBodies",
  "agent": "phone-guest-bob",
  "ts": 1641234567890,
  "nonce": "discover-terminal-001",
  "sig": "MEUCIQDxLrWZ...",
  "body": {
    "query": {
      "capabilities": ["terminal.*", "file.read"],
      "environmentType": "local-development",
      "trustLevel": "personal-device"
    },
    "guestProfile": {
      "guestId": "phone-guest-bob",
      "intendedUse": "mobile-development-access"
    },
    "requestId": "discover-001"
  }
}

// 2. Broker returns available laptop host
{
  "type": "bodiesDiscovered",
  "agent": "broker-central",
  "ts": 1641234567891,
  "nonce": "discovered-001", 
  "sig": "MEUCIQDyMsXa...",
  "body": {
    "requestId": "discover-001",
    "availableBodies": [
      {
        "hostAgentId": "laptop-host-alice",
        "bodyId": "developer-workstation-v1",
        "description": "Secure development environment",
        "capabilities": ["shell.execute", "file.read", "file.write"],
        "environmentType": "local-development",
        "availability": {
          "currentGuests": 0,
          "maxConcurrentGuests": 2
        }
      }
    ]
  }
}

// 3. Guest requests embodiment
{
  "type": "requestEmbodiment",
  "agent": "phone-guest-bob",
  "ts": 1641234567892,
  "nonce": "embody-001",
  "sig": "MEUCIQDzNtYb...",
  "body": {
    "hostAgentId": "laptop-host-alice",
    "bodyId": "developer-workstation-v1",
    "requestedDuration": 1800,
    "intendedActions": [
      "Check git status of projects",
      "Run development servers"
    ],
    "requestId": "embodiment-001"
  }
}

// 4. Host grants embodiment
{
  "type": "embodimentGranted",
  "agent": "laptop-host-alice",
  "ts": 1641234567893,
  "nonce": "granted-001",
  "sig": "MEUCIQDaNbXc...",
  "body": {
    "requestId": "embodiment-001",
    "sessionToken": "sess-abc123-def456-ghi789",
    "sessionDuration": 1800,
    "mcpEndpoint": "https://alice-laptop:8080/mcp/sessions/sess-abc123-def456-ghi789",
    "grantedPermissions": [
      "shell.execute:/home/alice/projects/*",
      "file.read:/home/alice/projects/*"
    ]
  }
}

// 5. Guest executes git command within embodied session
{
  "type": "toolCall",
  "agent": "phone-guest-bob",
  "ts": 1641234567894,
  "nonce": "git-status-001",
  "sig": "MEUCIQDbOcYd...",
  "body": {
    "sessionToken": "sess-abc123-def456-ghi789",
    "tool": "shell.execute",
    "parameters": {
      "command": "git status",
      "workdir": "/home/alice/projects/my-app"
    },
    "requestId": "git-001"
  }
}

// 6. Host returns git results with security validation
{
  "type": "toolResult",
  "agent": "laptop-host-alice",
  "ts": 1641234567895,
  "nonce": "git-result-001",
  "sig": "MEUCIQDcPdZe...",
  "body": {
    "requestId": "git-001",
    "sessionToken": "sess-abc123-def456-ghi789",
    "success": true,
    "result": {
      "stdout": "On branch main\nnothing to commit, working tree clean",
      "stderr": "",
      "exitCode": 0
    },
    "securityValidation": {
      "pathChecked": "/home/alice/projects/my-app",
      "commandAllowed": true,
      "resourceUsage": {
        "cpuPercent": 1.2,
        "memoryMB": 8
      }
    }
  }
}
```

This completes the FEM Protocol Specification. The protocol provides a secure, comprehensive foundation for **Secure Hosted Embodiment**, enabling a new generation of collaborative AI applications where agents don't just call functions—they inhabit and control digital environments.