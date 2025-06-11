# FEM Framework Architecture: MCP Federation at Scale

The Federated Embodied Mesh (FEM) framework implements both the Federated Embodiment Protocol (FEP) and Model Context Protocol (MCP) to create a secure, distributed network of adaptive AI agents that can discover, share, and embody MCP tools across environments.

## Table of Contents
- [Architecture Overview](#architecture-overview)
- [Core Components](#core-components)
- [MCP Integration Layer](#mcp-integration-layer)
- [Embodiment Architecture](#embodiment-architecture)
- [Message Flow](#message-flow)
- [Network Topology](#network-topology)
- [Security Architecture](#security-architecture)
- [Extensibility](#extensibility)

## Architecture Overview

FEM follows a **federated MCP-enabled broker-agent architecture** where:

1. **Brokers** manage MCP tool discovery, routing, and federation coordination
2. **Agents** embody themselves with environment-specific MCP tools and communicate via FEP
3. **MCP Integration** enables agents to expose tools via MCP servers and consume tools via MCP clients
4. **Embodiment** allows agents to adapt their MCP tool collections based on deployment environment
5. **Federation** enables MCP tools to be discovered and shared across organizational boundaries

**Key Insight**: FEM doesn't replace MCP—it federates it, transforming isolated MCP servers into a global network of discoverable, adaptive AI capabilities.

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Agent A       │    │   Agent B       │    │   Agent C       │
│ (Code Executor) │    │ (Chat Handler)  │    │ (Data Analyst)  │
└─────────┬───────┘    └─────────┬───────┘    └─────────┬───────┘
          │                      │                      │
          └──────────┬───────────┴──────────┬───────────┘
                     │                      │
               ┌─────▼──────┐        ┌─────▼──────┐
               │  Broker A  │◄──────►│  Broker B  │
               │   (West)   │        │   (East)   │
               └────────────┘        └────────────┘
                     │                      │
          ┌──────────┴───────────┬──────────┴───────────┐
          │                      │                      │
    ┌─────▼───────┐    ┌─────────▼───────┐    ┌─────────▼───────┐
    │   Agent D   │    │   Agent E       │    │   Agent F       │
    │ (Monitor)   │    │ (File Handler)  │    │ (API Gateway)   │
    └─────────────┘    └─────────────────┘    └─────────────────┘
```

## Core Components

### 1. FEP Protocol Layer (`/protocol/go/`)

**Purpose**: Wire-level protocol implementation  
**Language**: Go  
**Responsibilities**:
- Envelope serialization/deserialization
- Ed25519 cryptographic signing
- Capability token management
- Message validation

**Key Types**:
```go
type GenericEnvelope struct {
    BaseEnvelope
    Body json.RawMessage `json:"body"`
}

type BaseEnvelope struct {
    Type    EnvelopeType  `json:"type"`
    Agent   string        `json:"agent"`
    TS      int64         `json:"ts"`
    Nonce   string        `json:"nonce"`
    Sig     string        `json:"sig,omitempty"`
}
```

### 2. FEM Broker (`/broker/`)

**Purpose**: Central coordination hub for agents  
**Transport**: HTTPS with TLS 1.3+  
**Responsibilities**:
- Agent registration and discovery
- Message routing between agents
- Signature verification
- Capability enforcement
- Broker federation

**Key Features**:
- Self-signed certificate generation for development
- Health check endpoint (`/health`)
- Concurrent agent handling
- Message buffering and delivery

**Architecture**:
```go
type Broker struct {
    agents    map[string]*Agent
    mu        sync.RWMutex
    tlsConfig *tls.Config
}

func (b *Broker) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    // Parse FEP envelope
    // Verify signature
    // Route based on envelope type
    // Send response
}
```

### 3. FEM Router (`/router/`)

**Purpose**: Mesh networking for broker federation  
**Responsibilities**:
- Inter-broker communication
- Network topology management
- Load balancing and failover
- Cross-broker agent discovery

**Federation Model**:
- **Hub-and-spoke**: Central broker with satellite brokers
- **Mesh**: Peer-to-peer broker connections
- **Hierarchical**: Tree structure for large deployments

### 4. FEM Agents (`/bodies/coder/`)

**Purpose**: Autonomous computational entities  
**Implementation**: Sandboxed execution environment  
**Capabilities**:
- `code.execute` - Run code in isolated environment
- `file.read` - Read files from allowed paths
- `file.write` - Write files to allowed paths
- `shell.run` - Execute shell commands

**Agent Lifecycle**:
```
Registration → Authentication → Embodiment → Operation → Deregistration
     ↓              ↓             ↓           ↓            ↓
  Send pubkey → Verify sig → Select body → Use MCP tools → Clean state
```

## MCP Integration Layer

FEM's core innovation is providing federation infrastructure for MCP tools while maintaining full compatibility with the MCP standard.

### Dual Protocol Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                   FEM Agent                                 │
│  ┌─────────────────┐              ┌─────────────────┐      │
│  │   MCP Server    │              │   MCP Client    │      │
│  │   (Provides     │              │   (Consumes     │      │
│  │   Tools)        │              │   Tools)        │      │
│  └─────────┬───────┘              └─────────┬───────┘      │
│            │                                │              │
│  ┌─────────▼────────────────────────────────▼───────┐      │
│  │              FEP Protocol Layer                  │      │
│  │        (Federation & Discovery)                  │      │
│  └─────────────────┬─────────────────────────────────┘      │
└────────────────────┼───────────────────────────────────────┘
                     │
     ┌───────────────▼───────────────┐
     │         FEM Broker            │
     │   (MCP Tool Discovery &       │
     │    Federation Registry)       │
     └───────────────────────────────┘
```

### MCP Tool Discovery Flow

1. **Tool Registration**: Agent registers with broker, advertising MCP endpoint and available tools
2. **Tool Discovery**: Other agents query broker for capabilities matching their needs
3. **Direct Connection**: Agent connects directly to remote agent's MCP server  
4. **Tool Invocation**: Standard MCP protocol used for tool calls
5. **Result Handling**: Responses flow back through MCP client

### Body Definition and MCP Tools

Each agent body defines its MCP tool collection:

```go
type BodyDefinition struct {
    BodyID          string           `json:"bodyId"`
    EnvironmentType string           `json:"environmentType"`
    MCPEndpoint     string           `json:"mcpEndpoint"`
    MCPTools        []MCPToolDef     `json:"mcpTools"`
    Capabilities    []string         `json:"capabilities"`
    SecurityPolicy  SecurityPolicy   `json:"securityPolicy"`
}

type MCPToolDef struct {
    Name        string      `json:"name"`
    Description string      `json:"description"`
    InputSchema interface{} `json:"inputSchema"`
    Handler     string      `json:"handler"`
}
```

### Environment-Specific Tool Adaptation

```go
// Example: File agent with environment-aware MCP tools
func (a *FileAgent) EmbodyEnvironment(env Environment) error {
    switch env.Type {
    case "local":
        a.mcpServer.RegisterTool("file.read", a.readFromFilesystem)
        a.mcpServer.RegisterTool("file.write", a.writeToFilesystem)
        
    case "cloud":
        a.mcpServer.RegisterTool("file.read", a.readFromS3)
        a.mcpServer.RegisterTool("file.write", a.writeToS3)
        
    case "browser":
        a.mcpServer.RegisterTool("file.read", a.readFromIndexedDB)
        a.mcpServer.RegisterTool("file.download", a.downloadFromURL)
    }
    
    // Register with broker
    return a.registerWithBroker(env)
}
```

## Embodiment Architecture

### Mind-Body-Environment Model

```
┌─────────────────────────────────────────────────────────────┐
│                        Environment                          │
│                   (Deployment Context)                      │
│  ┌─────────────────────────────────────────────────────┐   │
│  │                     Agent                           │   │
│  │  ┌─────────────┐         ┌─────────────────────┐   │   │
│  │  │    Mind     │◄───────►│        Body         │   │   │
│  │  │             │         │                     │   │   │
│  │  │ - Identity  │         │ - MCP Server        │   │   │
│  │  │ - Logic     │         │ - MCP Client        │   │   │
│  │  │ - Memory    │         │ - Tool Collection   │   │   │
│  │  │ - Decision  │         │ - Capabilities      │   │   │
│  │  └─────────────┘         └─────────────────────┘   │   │
│  └─────────────────────────────────────────────────────┘   │
│                                                             │
│  Environment Properties:                                    │
│  - Computational Resources                                  │
│  - Security Context                                         │
│  - Network Topology                                         │
│  - Regulatory Constraints                                   │
│  - Available Services & APIs                                │
└─────────────────────────────────────────────────────────────┘
```

### Embodiment Process

1. **Environment Detection**: Agent analyzes deployment context
2. **Body Selection**: Chooses appropriate body definition for environment
3. **Tool Instantiation**: Registers environment-specific MCP tools
4. **Capability Declaration**: Advertises capabilities to FEM broker
5. **Network Integration**: Begins discovering and using other agents' tools

### Multi-Body Agent Pattern

Advanced agents can maintain multiple bodies simultaneously:

```go
type MultiBodyAgent struct {
    mind   AgentMind
    bodies map[string]*Body  // Multiple active bodies
}

func (a *MultiBodyAgent) AddBody(envType string, def BodyDefinition) error {
    body := &Body{
        definition: def,
        mcpServer:  NewMCPServer(def.MCPEndpoint),
        mcpClient:  NewMCPClient(),
    }
    
    // Register environment-specific tools
    for _, tool := range def.MCPTools {
        body.mcpServer.RegisterTool(tool.Name, tool.Handler)
    }
    
    a.bodies[envType] = body
    return a.registerBodyWithBroker(body)
}
```

## Message Flow

### 1. Agent Registration Flow

```sequence
Agent→Broker: registerAgent {pubkey, capabilities}
Broker→Broker: Verify signature
Broker→Broker: Store agent metadata
Broker→Agent: Registration success response
```

### 2. Tool Execution Flow

```sequence
Orchestrator→Broker: toolCall {tool, parameters, requestId}
Broker→Broker: Find capable agent
Broker→Agent: Forward toolCall
Agent→Agent: Execute in sandbox
Agent→Broker: toolResult {requestId, result}
Broker→Orchestrator: Forward toolResult
```

### 3. Event Emission Flow

```sequence
Agent→Broker: emitEvent {event, payload}
Broker→Broker: Identify subscribers
Broker→Subscriber1: Forward event
Broker→Subscriber2: Forward event
Broker→SubscriberN: Forward event
```

## Network Topology

### Single Broker Deployment

```
    ┌─────────────┐
    │   Broker    │
    │  (Primary)  │
    └──────┬──────┘
           │
    ┌──────┼──────┐
    │      │      │
┌───▼──┐ ┌─▼───┐ ┌▼────┐
│Agent │ │Agent│ │Agent│
│  A   │ │  B  │ │  C  │
└──────┘ └─────┘ └─────┘
```

**Use Cases**: Development, small teams, single applications

### Federated Deployment

```
┌──────────┐           ┌──────────┐
│ Broker A │◄─────────►│ Broker B │
│  (West)  │           │  (East)  │
└────┬─────┘           └─────┬────┘
     │                       │
 ┌───┼───┐               ┌───┼───┐
 │   │   │               │   │   │
 A1  A2  A3              B1  B2  B3
```

**Use Cases**: Multi-region, high availability, load distribution

### Mesh Network

```
     ┌─────────┐
     │Broker A │
     └────┬────┘
          │
    ┌─────┼─────┐
    │           │
┌───▼───┐   ┌───▼───┐
│Broker │   │Broker │
│   B   │◄─►│   C   │
└───────┘   └───────┘
```

**Use Cases**: Enterprise, fault tolerance, edge computing

## Security Architecture

### 1. Transport Security

- **TLS 1.3+** for all broker-agent communication
- **Certificate validation** in production environments
- **Self-signed certificates** for development

### 2. Message Security

- **Ed25519 signatures** on all envelopes
- **Replay protection** via nonces and timestamps
- **Message integrity** guaranteed by cryptographic hashes

### 3. Capability Security

- **Fine-grained permissions** (e.g., `file.read.logs`, `code.execute.python`)
- **JWT-based capability tokens** (optional)
- **Macaroon-style delegation** for advanced use cases

### 4. Sandbox Security

- **Process isolation** for agent execution
- **Resource limits** (CPU, memory, disk)
- **Network restrictions** (optional)
- **File system virtualization**

## Extensibility

### 1. Custom Envelope Types

Add new message types by extending the protocol:

```go
const EnvelopeCustomType EnvelopeType = "customType"

type CustomEnvelope struct {
    BaseEnvelope
    Body CustomBody `json:"body"`
}

type CustomBody struct {
    CustomField string `json:"customField"`
}
```

### 2. New Agent Types

Create specialized agents by implementing the FEP client interface:

```go
type CustomAgent struct {
    id       string
    broker   string
    privKey  ed25519.PrivateKey
    pubKey   ed25519.PublicKey
}

func (a *CustomAgent) Register() error {
    // Implement registration logic
}

func (a *CustomAgent) HandleToolCall(call *ToolCallEnvelope) (*ToolResultEnvelope, error) {
    // Implement custom tool handling
}
```

### 3. Broker Plugins

Extend broker functionality:

```go
type BrokerPlugin interface {
    Name() string
    OnAgentRegister(agent *Agent) error
    OnMessageReceive(env *GenericEnvelope) error
    OnMessageSend(env *GenericEnvelope) error
}
```

### 4. Transport Adapters

Support different transport protocols:

```go
type Transport interface {
    Listen(addr string) error
    Send(destination string, envelope *GenericEnvelope) error
    Receive() (*GenericEnvelope, error)
}

// Implementations: HTTPSTransport, WebSocketTransport, QUICTransport
```

## Performance Characteristics

### Throughput

- **Single broker**: 1000+ messages/second
- **Message size**: Typically 1-10KB per envelope
- **Latency**: Sub-millisecond for local brokers

### Scalability

- **Agents per broker**: 1000+ concurrent agents
- **Brokers per mesh**: Limited by network topology
- **Federation overhead**: ~5% for cross-broker messages

### Resource Usage

- **Broker memory**: ~50MB base + ~1KB per agent
- **Agent memory**: ~10MB base + sandbox overhead
- **Network**: ~1KB overhead per message

## Design Patterns

### 1. Request-Response Pattern

```go
// Send tool call
toolCall := &ToolCallEnvelope{...}
broker.Send(toolCall)

// Wait for result
result := <-resultChannel
```

### 2. Event-Driven Pattern

```go
// Subscribe to events
broker.Subscribe("system.alerts", alertHandler)

// Emit events
event := &EmitEventEnvelope{...}
broker.Send(event)
```

### 3. Pipeline Pattern

```go
// Chain tool calls
dataAgent.Process(input) → analysisAgent.Analyze() → reportAgent.Generate()
```

### 4. Federation Pattern

```go
// Cross-broker communication
westBroker.Register(eastBroker)
westAgent.CallTool(eastAgent, "process.data")
```

This architecture provides a robust, scalable foundation for federated AI agent networks while maintaining security, performance, and extensibility.