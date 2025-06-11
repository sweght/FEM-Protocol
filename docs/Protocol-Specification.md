# Federated Embodiment Protocol (FEP) Specification

**Author**: Chaz Dinkle  
**Version**: 0.1.3  
**License**: CC-BY-SA 4.0

## Table of Contents
- [Overview](#overview)
- [Protocol Fundamentals](#protocol-fundamentals)
- [Message Envelopes](#message-envelopes)
- [Security Model](#security-model)
- [Transport Layer](#transport-layer)
- [Agent Lifecycle](#agent-lifecycle)
- [Broker Operations](#broker-operations)
- [Error Handling](#error-handling)
- [Examples](#examples)

## Overview

The Federated Embodiment Protocol (FEP) is a wire-level protocol designed for secure communication between autonomous AI agents in a federated network. FEP enables agents to register with brokers, emit events, execute tools, and collaborate across distributed systems while maintaining cryptographic security and capability-based authorization.

### Key Design Principles

1. **Decentralization**: No single point of control; agents can communicate across multiple brokers
2. **Security-First**: All messages are cryptographically signed using Ed25519
3. **Capability-Based**: Agents declare and are granted specific capabilities
4. **Transport Agnostic**: Works over any reliable transport (HTTPS, WebSockets, etc.)
5. **Extensible**: New envelope types and capabilities can be added
6. **Federated**: Brokers can connect to form larger networks

### Protocol Version

Current version: **v0.1.2**

## Protocol Fundamentals

### Core Concepts

**Agent**: An autonomous entity that can execute code, process data, or perform other computational tasks. Agents have unique identifiers and declare their capabilities upon registration.

**Broker**: A network node that facilitates communication between agents. Brokers handle message routing, capability matching, and federation with other brokers.

**Envelope**: A structured message format that wraps all FEP communications. Each envelope contains headers, body, and cryptographic signature.

**Capability**: A declared ability of an agent (e.g., "code.execute", "file.read", "chat.respond"). Capabilities enable fine-grained access control.

**Federation**: The process of connecting multiple brokers to form a larger network, allowing agents on different brokers to interact.

### Message Flow

1. **Registration**: Agent connects to broker and sends registerAgent envelope
2. **Capability Declaration**: Agent declares what it can do (capabilities)
3. **Authentication**: Broker verifies agent's cryptographic signature
4. **Operation**: Agent can now emit events, receive instructions, execute tools
5. **Federation**: Broker can route messages to agents on other connected brokers

## Message Envelopes

All FEP communication uses structured message envelopes. Every envelope shares common headers and contains a type-specific body.

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
- **nonce**: Unique string to prevent replay attacks
- **sig**: Base64-encoded Ed25519 signature of the entire envelope (excluding sig field)
- **body**: Type-specific message content

### Envelope Types

FEP defines seven core envelope types:

#### 1. registerAgent

Registers a new agent with a broker.

```json
{
  "type": "registerAgent",
  "agent": "my-coding-agent-001",
  "ts": 1641234567890,
  "nonce": "reg-12345-67890",
  "sig": "Mf8B7tKqE...",
  "body": {
    "pubkey": "base64-encoded-ed25519-public-key",
    "capabilities": ["code.execute", "file.read", "shell.run"],
    "metadata": {
      "version": "1.0.0",
      "description": "Python code execution agent"
    }
  }
}
```

**Body Fields**:
- `pubkey`: Agent's Ed25519 public key for signature verification
- `capabilities`: Array of capability strings the agent provides
- `metadata`: Optional additional information about the agent

#### 2. registerBroker

Registers a broker with another broker for federation.

```json
{
  "type": "registerBroker",
  "agent": "broker-west-coast",
  "ts": 1641234567890,
  "nonce": "broker-reg-98765",
  "sig": "Kl9A8uLpF...",
  "body": {
    "brokerId": "broker-west-coast",
    "endpoint": "https://west.example.com:8443",
    "pubkey": "base64-encoded-ed25519-public-key",
    "capabilities": ["federation", "routing", "discovery"]
  }
}
```

**Body Fields**:
- `brokerId`: Unique identifier for the broker
- `endpoint`: TLS endpoint where the broker can be reached
- `pubkey`: Broker's Ed25519 public key
- `capabilities`: Broker-level capabilities for federation

#### 3. emitEvent

Emits an event that other agents can observe.

```json
{
  "type": "emitEvent",
  "agent": "monitoring-agent-001",
  "ts": 1641234567890,
  "nonce": "event-54321-09876",
  "sig": "Np7C2vMrG...",
  "body": {
    "event": "system.resource.warning",
    "payload": {
      "resource": "memory",
      "usage": 0.85,
      "threshold": 0.80,
      "timestamp": "2024-01-03T10:30:00Z"
    }
  }
}
```

**Body Fields**:
- `event`: Event type identifier (hierarchical naming recommended)
- `payload`: Event-specific data

#### 4. renderInstruction

Sends an instruction for an agent to process or execute.

```json
{
  "type": "renderInstruction",
  "agent": "orchestrator-001",
  "ts": 1641234567890,
  "nonce": "instr-11111-22222",
  "sig": "Ql8D3wNsH...",
  "body": {
    "instruction": "Analyze the uploaded CSV file and generate a summary report",
    "parameters": {
      "target_agent": "data-analyst-001",
      "file_path": "/uploads/data.csv",
      "format": "markdown"
    }
  }
}
```

**Body Fields**:
- `instruction`: Human-readable instruction text
- `parameters`: Optional structured parameters for the instruction

#### 5. toolCall

Requests execution of a specific tool or capability.

```json
{
  "type": "toolCall",
  "agent": "orchestrator-001",
  "ts": 1641234567890,
  "nonce": "tool-call-33333",
  "sig": "Rm9E4xOtI...",
  "body": {
    "tool": "code.execute",
    "parameters": {
      "language": "python",
      "code": "print('Hello, FEP!')\nresult = 2 + 2\nprint(f'Result: {result}')"
    },
    "requestId": "req-python-exec-001"
  }
}
```

**Body Fields**:
- `tool`: Capability identifier for the tool to execute
- `parameters`: Tool-specific parameters
- `requestId`: Unique identifier to correlate with toolResult

#### 6. toolResult

Returns the result of a tool execution.

```json
{
  "type": "toolResult",
  "agent": "coding-agent-001",
  "ts": 1641234567890,
  "nonce": "tool-result-44444",
  "sig": "Sn0F5yPuJ...",
  "body": {
    "requestId": "req-python-exec-001",
    "success": true,
    "result": {
      "stdout": "Hello, FEP!\nResult: 4",
      "stderr": "",
      "exit_code": 0
    }
  }
}
```

**Body Fields**:
- `requestId`: Correlates with the original toolCall
- `success`: Boolean indicating if tool execution succeeded
- `result`: Tool execution results (success case)
- `error`: Error message (failure case)

#### 7. revoke

Revokes an agent's registration or specific capabilities.

```json
{
  "type": "revoke",
  "agent": "admin-agent-001",
  "ts": 1641234567890,
  "nonce": "revoke-55555",
  "sig": "To1G6zQuK...",
  "body": {
    "target": "suspicious-agent-999",
    "reason": "Security policy violation"
  }
}
```

**Body Fields**:
- `target`: Agent identifier to revoke
- `reason`: Optional human-readable reason for revocation

## Security Model

FEP implements a comprehensive security model based on cryptographic signatures and capability-based authorization.

### Cryptographic Foundation

**Algorithm**: Ed25519 (Edwards-curve Digital Signature Algorithm)
- **Key Size**: 32 bytes (256 bits)
- **Signature Size**: 64 bytes
- **Security Level**: ~128-bit security
- **Performance**: Fast signing and verification

### Signature Process

1. **Envelope Creation**: Agent creates envelope with all fields except `sig`
2. **Serialization**: Envelope is serialized to canonical JSON
3. **Signing**: Agent signs the serialized data with its Ed25519 private key
4. **Encoding**: Signature is base64-encoded and added to `sig` field

### Verification Process

1. **Signature Extraction**: Broker extracts `sig` field from envelope
2. **Envelope Reconstruction**: Temporarily removes `sig` field
3. **Serialization**: Envelope is serialized to canonical JSON
4. **Verification**: Signature is verified against agent's known public key

### Replay Protection

- **Nonce**: Each envelope must include a unique nonce
- **Timestamp**: Recent timestamp required (configurable window)
- **Broker Tracking**: Brokers can track recent nonces to prevent replays

### Capability-Based Authorization

- **Declaration**: Agents declare capabilities during registration
- **Verification**: Brokers verify agents only use declared capabilities
- **Scoping**: Capabilities can be hierarchical (e.g., "file.read.logs")
- **Revocation**: Capabilities can be revoked by authorized agents

## Transport Layer

FEP is designed to be transport-agnostic but mandates certain security requirements.

### Required Transport Properties

1. **Reliability**: Messages must be delivered in order
2. **Security**: Transport must provide confidentiality and integrity
3. **Binary-Safe**: Must handle arbitrary binary data in JSON strings

### Recommended Transports

#### HTTPS (Primary)

- **TLS Version**: 1.3 or higher recommended
- **Method**: POST for all envelope submissions
- **Content-Type**: `application/json`
- **Endpoint**: Broker-defined (commonly `/fep` or `/`)

#### WebSocket (Real-time)

- **Protocol**: WSS (WebSocket Secure)
- **Framing**: Each envelope is a separate text frame
- **Keepalive**: Recommended for long-lived connections

#### Message Queues (Asynchronous)

- **Protocols**: AMQP, MQTT over TLS
- **Durability**: Messages should be persisted
- **Ordering**: FIFO delivery required

### HTTP-Specific Considerations

When using HTTPS transport:

**Request Format**:
```http
POST /fep HTTP/1.1
Host: broker.example.com
Content-Type: application/json
Content-Length: [length]

{envelope-json}
```

**Response Format**:
```http
HTTP/1.1 200 OK
Content-Type: application/json

{
  "status": "success",
  "message": "Envelope processed"
}
```

**Error Responses**:
- `400 Bad Request`: Malformed envelope
- `401 Unauthorized`: Invalid signature
- `403 Forbidden`: Insufficient capabilities
- `429 Too Many Requests`: Rate limiting
- `500 Internal Server Error`: Broker error

## Agent Lifecycle

### Registration Phase

1. **Key Generation**: Agent generates Ed25519 key pair
2. **Capability Declaration**: Agent determines what capabilities it provides
3. **Broker Discovery**: Agent finds broker endpoint (DNS, config, etc.)
4. **Registration Request**: Agent sends `registerAgent` envelope
5. **Verification**: Broker validates signature and capabilities
6. **Confirmation**: Broker responds with registration status

### Operational Phase

1. **Event Emission**: Agent can emit events for others to observe
2. **Instruction Processing**: Agent receives and processes instructions
3. **Tool Execution**: Agent executes tools and returns results
4. **Collaboration**: Agent works with other agents through broker

### Deregistration Phase

1. **Graceful Shutdown**: Agent notifies broker of pending disconnect
2. **Resource Cleanup**: Broker cleans up agent state
3. **Revocation**: Admin agents can forcibly revoke problematic agents

## Broker Operations

### Core Responsibilities

1. **Agent Registry**: Maintain registry of connected agents and capabilities
2. **Message Routing**: Route envelopes between agents
3. **Signature Verification**: Validate all incoming envelope signatures
4. **Capability Enforcement**: Ensure agents only use declared capabilities
5. **Federation**: Connect with other brokers for larger networks

### Message Processing Pipeline

1. **Transport Reception**: Receive envelope over transport
2. **Parsing**: Parse JSON envelope structure
3. **Signature Verification**: Verify Ed25519 signature
4. **Agent Lookup**: Find sending agent in registry
5. **Capability Check**: Verify agent has required capabilities
6. **Processing**: Handle envelope based on type
7. **Response**: Send response to agent
8. **Routing**: Forward to other agents if needed

### Federation Protocol

Brokers can connect to form federated networks:

1. **Discovery**: Brokers discover each other (DNS, configuration)
2. **Registration**: Broker A registers with Broker B using `registerBroker`
3. **Capability Sharing**: Brokers share available agent capabilities
4. **Message Routing**: Route messages across broker boundaries
5. **Health Monitoring**: Monitor connection health and failover

## Error Handling

### Client Errors (4xx)

- **Malformed Envelope**: Invalid JSON or missing required fields
- **Invalid Signature**: Signature verification failed
- **Unknown Agent**: Agent not registered with broker
- **Insufficient Capabilities**: Agent lacks required capability

### Server Errors (5xx)

- **Broker Overload**: Too many concurrent requests
- **Storage Failure**: Unable to persist agent state
- **Federation Error**: Error communicating with federated broker
- **Internal Error**: Unexpected broker error

### Error Response Format

```json
{
  "status": "error",
  "code": "INVALID_SIGNATURE",
  "message": "Envelope signature verification failed",
  "details": {
    "agent": "suspicious-agent-001",
    "timestamp": "2024-01-03T10:30:00Z"
  }
}
```

## Examples

### Complete Agent Registration Flow

```json
// 1. Agent generates key pair and sends registration
{
  "type": "registerAgent",
  "agent": "example-coder-001",
  "ts": 1641234567890,
  "nonce": "reg-abc123-def456",
  "sig": "MEUCIQDxLrWZ...",
  "body": {
    "pubkey": "MCowBQYDK2VwAyEA...",
    "capabilities": ["code.execute", "file.read"],
    "metadata": {
      "language": "python",
      "version": "1.0.0"
    }
  }
}

// 2. Broker responds with success
{
  "status": "success",
  "agent": "example-coder-001",
  "capabilities_granted": ["code.execute", "file.read"],
  "broker_id": "main-broker-001"
}
```

### Tool Execution Flow

```json
// 1. Orchestrator requests code execution
{
  "type": "toolCall",
  "agent": "orchestrator-001",
  "ts": 1641234567890,
  "nonce": "tool-xyz789",
  "sig": "MEUCIQDyMsXa...",
  "body": {
    "tool": "code.execute",
    "parameters": {
      "language": "python",
      "code": "import math\nprint(f'Pi is approximately {math.pi:.2f}')"
    },
    "requestId": "req-001"
  }
}

// 2. Coding agent executes and returns result
{
  "type": "toolResult",
  "agent": "example-coder-001",
  "ts": 1641234567900,
  "nonce": "result-xyz790",
  "sig": "MEUCIQDzNtYb...",
  "body": {
    "requestId": "req-001",
    "success": true,
    "result": {
      "stdout": "Pi is approximately 3.14",
      "stderr": "",
      "exit_code": 0,
      "execution_time": 0.123
    }
  }
}
```

This completes the FEP Protocol Specification. The protocol provides a secure, extensible foundation for federated AI agent communication with strong cryptographic guarantees and flexible capability management.