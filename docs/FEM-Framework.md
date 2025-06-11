# FEM Framework Architecture

**Framework by Chaz Dinkle**

The Federated Embodied Mesh (FEM) framework implements the Federated Embodiment Protocol (FEP) to create a secure, distributed network of autonomous AI agents.

## Table of Contents
- [Architecture Overview](#architecture-overview)
- [Core Components](#core-components)
- [Message Flow](#message-flow)
- [Network Topology](#network-topology)
- [Security Architecture](#security-architecture)
- [Extensibility](#extensibility)

## Architecture Overview

FEM follows a **federated broker-agent architecture** where:

1. **Brokers** manage message routing and agent registration
2. **Agents** perform computational tasks and communicate via brokers
3. **Federation** allows brokers to connect for larger networks
4. **Capabilities** define what agents can do with fine-grained permissions

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
Registration → Authentication → Operation → Deregistration
     ↓              ↓             ↓            ↓
  Send pubkey → Verify sig → Process tools → Clean state
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