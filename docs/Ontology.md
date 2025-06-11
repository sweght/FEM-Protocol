# FEP-FEM Ontology: Formal Definitions

This document provides precise definitions for the core concepts in the Federated Embodiment Protocol (FEP) and Federated Embodied Mesh (FEM) framework, establishing the formal ontology that underpins the entire system.

## Table of Contents
- [Core Concepts](#core-concepts)
- [Agent Architecture](#agent-architecture)
- [MCP Integration](#mcp-integration)
- [Environment and Embodiment](#environment-and-embodiment)
- [Federation Model](#federation-model)
- [Security and Access Control](#security-and-access-control)
- [Formal Relationships](#formal-relationships)

## Core Concepts

### Mind
**Definition**: The autonomous, persistent identity and decision-making logic of an agent.

**Properties**:
- **Identity**: Cryptographically verifiable Ed25519 key pair
- **Autonomy**: Capable of independent decision-making and action
- **Persistence**: Maintains state and memory across embodiments
- **Logic**: Core reasoning and processing capabilities

**Implementation**: A FEM agent's core logic, including its AI model, decision trees, and persistent state management.

**Example**: A data analysis agent's core reasoning capabilities that remain consistent whether it's embodied in a local environment, cloud environment, or edge device.

### Body
**Definition**: The collection of tools, capabilities, and interfaces that an agent can use to interact with its environment.

**Properties**:
- **Tool Collection**: Set of MCP tools available to the agent
- **Capabilities**: Declared abilities (e.g., "file.read", "code.execute")
- **Interfaces**: Communication endpoints and protocols
- **Adaptability**: Can change based on environment

**Implementation**: MCP server exposing tools + MCP client for consuming tools from other agents.

**Example**: A file management agent might have bodies with tools like:
- Local body: `file.read`, `file.write`, `dir.list`
- Cloud body: `s3.get`, `s3.put`, `s3.list`
- Browser body: `indexeddb.read`, `localstorage.write`

### Environment
**Definition**: The computational, regulatory, and resource context in which an agent operates.

**Properties**:
- **Computational Resources**: CPU, memory, storage, network capabilities
- **Security Context**: Trust level, isolation requirements, access controls
- **Regulatory Context**: Data residency, compliance requirements
- **Network Topology**: Connectivity, latency, bandwidth characteristics
- **Platform Constraints**: Operating system, runtime, available libraries

**Implementation**: Environment detection logic that influences which body definition an agent adopts.

**Examples**:
- **Local Development**: High trust, full filesystem access, development tools
- **Cloud Production**: Scalable resources, container isolation, service mesh
- **Edge Device**: Limited resources, intermittent connectivity, local processing
- **Browser Extension**: Sandboxed execution, DOM access, user interaction
- **Mobile Application**: Touch interface, sensors, offline capabilities

### Embodiment
**Definition**: The process by which a mind adapts its body to suit its environment, creating an embodied agent optimized for its operational context.

**Properties**:
- **Contextual Adaptation**: Tools and capabilities adjust to environment
- **Resource Optimization**: Efficient use of available computational resources
- **Compliance Alignment**: Respects environmental constraints and regulations
- **Performance Tuning**: Optimizes for environment-specific performance characteristics

**Implementation**: Dynamic MCP tool registration based on environment detection and body definition templates.

**Example**: A universal assistant agent embodying differently:
```
Mind: Assistant Logic
├── Local Embodiment → Body: [file.system, shell.exec, app.launch]
├── Cloud Embodiment → Body: [api.call, db.query, scale.compute]
├── Mobile Embodiment → Body: [camera.capture, gps.location, contacts.access]
└── Browser Embodiment → Body: [dom.query, storage.local, history.search]
```

### Agent
**Definition**: The complete entity consisting of a mind embodied within a specific environment, possessing a body of capabilities.

**Formula**: `Agent = Mind + Body + Environment`

**Properties**:
- **Complete Functionality**: Capable of autonomous operation
- **Environmental Specificity**: Optimized for its operational context
- **Collaborative Ability**: Can interact with other agents through FEP
- **Tool Federation**: Can expose and consume MCP tools across the network

## Agent Architecture

### Mind-Body-Environment Relationship

```
┌─────────────────────────────────────────────────────────────┐
│                        Environment                          │
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
│  - Resources (CPU, Memory, Storage)                         │
│  - Security Context                                         │
│  - Network Topology                                         │
│  - Regulatory Constraints                                   │
└─────────────────────────────────────────────────────────────┘
```

### Body Definition Template
**Definition**: A specification that defines what MCP tools and capabilities an agent should expose when embodied in a specific environment type.

**Structure**:
```json
{
  "bodyId": "file-agent-local",
  "environmentType": "local-development",
  "description": "File management agent for local development environment",
  "mcpTools": [
    {
      "name": "file.read",
      "description": "Read file contents from local filesystem",
      "inputSchema": { "type": "object", "properties": { "path": {"type": "string"} } }
    },
    {
      "name": "file.write", 
      "description": "Write content to local filesystem",
      "inputSchema": { "type": "object", "properties": { "path": {"type": "string"}, "content": {"type": "string"} } }
    }
  ],
  "capabilities": ["file.read", "file.write", "file.list"],
  "securityPolicy": {
    "allowedPaths": ["/tmp", "/home/user/workspace"],
    "maxFileSize": "10MB"
  },
  "resourceLimits": {
    "maxMemory": "512MB",
    "maxConcurrentOperations": 5
  }
}
```

## MCP Integration

### MCP-FEM Relationship
**Principle**: FEM federates MCP tools rather than replacing them.

**Integration Points**:
1. **Tool Exposure**: Agents expose their capabilities as MCP servers
2. **Tool Discovery**: Agents discover available tools through FEP protocol
3. **Tool Consumption**: Agents use MCP client protocol to invoke remote tools
4. **Federation**: FEP enables MCP tools to be shared across organizational boundaries

### MCP Server (Tool Provider)
**Definition**: An agent's interface for exposing its capabilities to other agents in the network.

**Implementation**: Each embodied agent runs an MCP server that:
- Registers tools based on current body definition
- Handles tool invocation requests
- Returns results via MCP protocol
- Manages authentication and authorization

### MCP Client (Tool Consumer)
**Definition**: An agent's interface for discovering and using tools from other agents.

**Implementation**: Agents use MCP clients to:
- Discover available tools through FEP brokers
- Invoke tools on remote agents via MCP protocol
- Handle tool results and errors
- Maintain connection pooling and retry logic

### Tool Federation Flow
1. **Registration**: Agent registers with broker, advertising MCP endpoint
2. **Discovery**: Other agents query broker for available capabilities
3. **Connection**: Agent connects to remote agent's MCP server
4. **Invocation**: Tool called via standard MCP protocol
5. **Result**: Response returned through MCP client

## Environment and Embodiment

### Environment Classification

#### Computational Environments
- **Local**: Direct hardware access, full privileges
- **Container**: Isolated but efficient, shared kernel
- **Serverless**: Stateless, event-driven, auto-scaling
- **Edge**: Resource-constrained, intermittent connectivity
- **Mobile**: Battery-conscious, sensor-rich, touch interface

#### Security Environments  
- **Trusted**: Full access, minimal sandboxing
- **Semi-trusted**: Capability restrictions, monitored execution
- **Untrusted**: Heavy sandboxing, strict resource limits
- **Public**: Assume hostile environment, maximum isolation

#### Regulatory Environments
- **GDPR Zone**: EU data protection compliance required
- **HIPAA**: Healthcare data protection requirements
- **SOX**: Financial reporting compliance
- **Classified**: Government security clearance levels

### Embodiment Patterns

#### Static Embodiment
**Definition**: Agent body is determined at deployment time and remains fixed.

**Use Case**: Production services with well-defined operational parameters.

**Example**: Database agent deployed to cloud always has cloud-specific tools.

#### Dynamic Embodiment  
**Definition**: Agent adapts its body based on runtime environment detection.

**Use Case**: Multi-environment deployment, edge-to-cloud migration.

**Example**: Data processing agent that adapts tools based on available resources.

#### Multi-Body Embodiment
**Definition**: Single mind maintains multiple bodies simultaneously across different environments.

**Use Case**: Cross-environment workflows, data synchronization.

**Example**: Sync agent with local and cloud bodies keeping data consistent.

#### Progressive Embodiment
**Definition**: Agent gradually gains new capabilities as it proves trustworthiness.

**Use Case**: Security-conscious environments, capability escalation.

**Example**: New agent starts with read-only tools, gains write access over time.

## Federation Model

### Network Topology

#### Single Broker
```
    Broker
   /   |   \
  A1   A2   A3
```
**Use Case**: Small teams, single organization, development

#### Federated Brokers
```
Broker-A ←→ Broker-B
   |           |
  A1,A2      B1,B2
```
**Use Case**: Multi-organization, geographic distribution

#### Mesh Federation
```
  Broker-A
   /     \
Broker-B—Broker-C
   \     /
  Broker-D
```
**Use Case**: High availability, redundancy, scale

### Cross-Broker Agent Interaction
1. **Discovery**: Agent queries local broker for capabilities
2. **Federation Lookup**: Broker queries federated brokers
3. **Routing**: Request routed to target broker
4. **MCP Connection**: Direct MCP connection established
5. **Tool Execution**: Standard MCP tool invocation

## Security and Access Control

### Identity and Authentication
- **Agent Identity**: Ed25519 public key uniquely identifies each mind
- **Broker Authentication**: Mutual TLS for broker-to-broker communication
- **Message Integrity**: All FEP messages cryptographically signed

### Capability-Based Security
- **Declared Capabilities**: Agents explicitly declare what they can do
- **Capability Verification**: Brokers enforce declared capabilities
- **Least Privilege**: Agents only get minimum required capabilities
- **Capability Revocation**: Capabilities can be dynamically revoked

### Body-Level Security
- **Environment Isolation**: Bodies cannot access unauthorized resources
- **Tool Sandboxing**: Each MCP tool execution is isolated
- **Resource Limits**: CPU, memory, storage, network limits enforced
- **Audit Logging**: All tool executions logged for security analysis

### Role-Based Access Control (RBAC)
**Definition**: Body definitions can include role assignments that determine which agents can use which tools.

**Implementation**:
```json
{
  "bodyId": "database-admin",
  "roles": ["dba", "data-admin"],
  "mcpTools": [
    {
      "name": "db.backup",
      "requiredRoles": ["dba"]
    },
    {
      "name": "db.query", 
      "requiredRoles": ["dba", "data-analyst"]
    }
  ]
}
```

## Formal Relationships

### Composition Relationships
- `Agent ⊇ Mind` (Agent contains Mind)
- `Agent ⊇ Body` (Agent contains Body)  
- `Agent ⊆ Environment` (Agent exists within Environment)
- `Body ⊇ MCPTools` (Body contains MCP Tools)
- `Environment ⊇ Resources` (Environment provides Resources)

### Functional Relationships
- `embodiment: Mind × Environment → Body` (Embodiment function)
- `federation: Agent × Agent → Collaboration` (Federation enables collaboration)
- `toolDiscovery: Agent × Network → AvailableTools` (Discovery function)
- `toolInvocation: Agent × Tool × Parameters → Result` (Invocation function)

### Temporal Relationships
- `Mind` persists across embodiments
- `Body` changes with environment
- `Environment` evolves over time
- `Agent = Mind(t) + Body(Environment(t)) + Environment(t)`

### Cardinality Relationships
- One Mind can have multiple Bodies (1:N)
- One Body can exist in one Environment (1:1)
- One Environment can host multiple Agents (1:N)
- One Agent can use tools from multiple other Agents (N:N)

## Conclusion

This ontology establishes the formal conceptual foundation for FEP-FEM, clarifying how minds, bodies, and environments interact to create adaptive, federated AI agent networks. The integration with MCP provides a standards-based approach to tool federation, while the embodiment model enables agents to optimize their capabilities for their operational context.

The key insight is that FEM doesn't replace existing protocols like MCP—it federates them, creating a global network of discoverable, secure, and adaptable AI capabilities.