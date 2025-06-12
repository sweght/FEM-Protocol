# FEM Protocol Ontology: Secure Hosted Embodiment

This document provides precise definitions for the core concepts in the **FEM Protocol**, establishing the formal ontology that underpins **Secure Hosted Embodiment** and **Secure Delegated Control**.

## Table of Contents
- [Core Concepts](#core-concepts)
- [Embodiment Architecture](#embodiment-architecture)
- [Security and Trust Model](#security-and-trust-model)
- [Session Management](#session-management)
- [Federation Model](#federation-model)
- [Formal Relationships](#formal-relationships)

## Core Concepts

### Host
**Definition**: An agent that offers "bodies" (sandboxed capability sets) for guest embodiment, while retaining ultimate control over its environment.

**Properties**:
- **Environment Control**: Maintains sovereignty over its computational environment
- **Body Definitions**: Defines and offers specific capability sets for guest use
- **Security Policies**: Establishes boundaries and constraints for guest behavior
- **Session Management**: Grants, monitors, and terminates embodiment sessions

**Implementation**: A FEM agent running both an MCP server (exposing tools) and embodiment management logic that can securely host guest agents.

**Example**: A developer's laptop offering a "developer-workstation" body with file operations and shell access to mobile agents.

### Guest
**Definition**: An agent "mind" that can discover and inhabit bodies offered by hosts, exercising delegated control within host-defined boundaries.

**Properties**:
- **Discovery Capability**: Can search for and evaluate available bodies
- **Embodiment Client**: Can request and establish embodiment sessions
- **Delegated Action**: Exercises control through host-provided tools
- **Boundary Respect**: Operates within host-imposed security constraints

**Implementation**: A FEM agent with MCP client capabilities and embodiment session management.

**Example**: A mobile chat agent that discovers and inhabits laptop development environments for code execution.

### Body
**Definition**: A secure, sandboxed set of MCP tools and capabilities offered by a host for guest embodiment.

**Properties**:
- **Tool Collection**: Specific MCP tools available to embodied guests
- **Security Boundaries**: Constraints on what guests can access and control
- **Environment Specificity**: Adapted to the host's deployment context
- **Session Support**: Manages multiple concurrent guest embodiments

**Implementation**: A body definition with MCP tools, security policies, and session management configuration.

**Example**: A "storyteller-coop" body offering narrative control tools: `update_world()`, `add_npc()`, `update_character()`.

### Embodiment
**Definition**: The process by which a guest mind inhabits a host body, establishing a persistent session with delegated control capabilities.

**Properties**:
- **Session-Based**: Time-bounded with clear start and end
- **Permission-Constrained**: Guest actions limited by host-granted permissions
- **Auditable**: All actions logged for security and debugging
- **Revocable**: Host can terminate session for policy violations

**Implementation**: Cryptographically secured session with token-based authentication and permission validation.

**Example**: A 1-hour embodiment session where a phone agent controls laptop terminal access.

### Secure Delegated Control
**Definition**: The security model where hosts delegate specific control to guests within cryptographically enforced boundaries.

**Properties**:
- **Delegation**: Hosts explicitly grant control rather than guests taking it
- **Boundaries**: Clear technical and policy constraints on guest actions
- **Validation**: Every guest action validated against session permissions
- **Revocation**: Control can be instantly revoked by hosts

**Implementation**: Session tokens, permission lists, real-time validation, and audit logging.

**Example**: Guest can call `shell.execute("git status")` but not `shell.execute("rm -rf /")` based on host security policy.

## Embodiment Architecture

### Mind
**Definition**: The persistent identity, logic, and decision-making capabilities of an agent that remain consistent across different embodiments.

**Properties**:
- **Identity**: Cryptographically verifiable Ed25519 identity
- **Logic**: Core reasoning and processing capabilities
- **Memory**: Persistent state that survives embodiment changes
- **Adaptation**: Can work effectively across different body types

**Implementation**: The agent's core AI logic, identity management, and persistent storage.

**Example**: A narrative AI's storytelling logic that works whether embodied in a game, virtual world, or chat system.

### Broker-as-Agent
**Definition**: The broker is not infrastructure but a first-class agent with its own mind, body, and environment.

**Properties**:
- **Broker's Mind**: Federation logic, security policies, health monitoring
- **Broker's Body**: Network-level tools for embodiment management
- **Broker's Environment**: Production vs development embodiment policies
- **Agent Capabilities**: Can embody different operational modes

**Implementation**: Broker runs as a full FEM agent with MCP tools for network management.

**Example**: A production broker embodying strict security tools vs a development broker embodying debugging tools.

### Environment
**Definition**: The computational, security, and resource context in which embodiment occurs.

**Properties**:
- **Computational Resources**: CPU, memory, storage, network capabilities
- **Security Context**: Trust levels, isolation requirements, access controls
- **Regulatory Context**: Compliance requirements, data residency, governance
- **Network Topology**: Local, cloud, edge, federated configurations

**Implementation**: Environment detection logic that influences body selection and tool configuration.

**Example**: Same agent offering different bodies in local development vs cloud production environments.

## Security and Trust Model

### Trust Level
**Definition**: A hierarchical assessment of an agent's reliability and permitted access level.

**Levels**:
- **Unknown**: No prior interaction history
- **Basic**: Limited successful interactions, restricted access
- **Verified**: Significant positive history, standard access
- **Trusted**: Long-term reliable behavior, elevated access
- **Personal**: Personal devices and known entities, full access

**Implementation**: Reputation tracking based on session completion rates, policy compliance, and host feedback.

**Example**: A "basic" trust guest limited to read-only operations, while "trusted" guests can execute write operations.

### Security Policy
**Definition**: Host-defined rules that govern guest behavior during embodiment sessions.

**Components**:
- **Path Restrictions**: Allowed and denied file system locations
- **Command Filtering**: Permitted and forbidden shell commands
- **Resource Limits**: CPU, memory, disk, and network constraints
- **Time Boundaries**: Session duration and daily limits

**Implementation**: Policy engine that validates every guest action against defined rules.

**Example**: Guest limited to `/workspace/*` paths and denied `sudo` commands.

### Session Token
**Definition**: Cryptographically secure identifier for active embodiment sessions.

**Properties**:
- **Cryptographic Security**: 256-bit entropy, tamper-evident
- **Session Scope**: Unique per embodiment session
- **Time-Bounded**: Expires with session termination
- **Permission Binding**: Linked to specific guest permissions

**Implementation**: JWT-style tokens with embedded permissions and expiration.

**Example**: `sess-abc123-def456` grants specific guest access to specific host body for defined duration.

## Session Management

### Embodiment Session
**Definition**: A time-bounded period during which a guest has active delegated control over a host body.

**Lifecycle**:
1. **Discovery**: Guest finds suitable bodies
2. **Request**: Guest requests embodiment access
3. **Verification**: Host validates guest identity and policies
4. **Grant**: Host establishes session with permissions
5. **Active Control**: Guest exercises delegated control
6. **Monitoring**: Continuous validation and audit logging
7. **Termination**: Natural expiry or forced termination

**Implementation**: State machine with transition validation and audit logging.

### Session Monitoring
**Definition**: Continuous oversight of embodiment sessions to ensure policy compliance and detect violations.

**Components**:
- **Resource Tracking**: Monitor CPU, memory, disk usage
- **Action Auditing**: Log all tool calls and results
- **Violation Detection**: Identify policy breaches
- **Health Checking**: Ensure session remains valid

**Implementation**: Background monitoring threads with real-time policy validation.

## Federation Model

### Cross-Broker Embodiment
**Definition**: The ability for guests connected to one broker to discover and embody into hosts connected to different brokers.

**Properties**:
- **Discovery Federation**: Brokers share body availability information
- **Session Routing**: Embodiment requests routed to appropriate brokers
- **Security Propagation**: Trust and security policies honored across brokers
- **Health Coordination**: Session health monitored across broker boundaries

**Implementation**: Broker-to-broker federation protocol with shared session state.

### Network Topology
**Definition**: The arrangement and connection patterns of brokers in the FEM network.

**Types**:
- **Single Broker**: All agents connect to one broker
- **Federated Mesh**: Multiple brokers with peer connections
- **Hierarchical**: Tree structure with primary and secondary brokers

**Implementation**: Configurable connection patterns with automatic failover and load balancing.

## Formal Relationships

### Host-Guest Relationship
```
Host offers Bodies → Guest discovers Bodies → Guest requests Embodiment → 
Host grants Session → Guest exercises Delegated Control → Host monitors Actions → 
Host logs Audit → Session terminates
```

### Mind-Body-Environment Relationship
```
Mind (persistent logic) + Body (capability set) + Environment (context) = Embodied Agent
```

### Security Enforcement Chain
```
Guest Action → Session Token Validation → Permission Check → Security Policy Validation → 
Resource Limit Check → Action Execution → Audit Logging
```

### Trust Evolution
```
Unknown Guest → Basic Interactions → Successful Sessions → Policy Compliance → 
Trust Level Increase → Enhanced Permissions → Reputation Building
```

### Federation Discovery Flow
```
Guest Query → Local Broker → Federated Brokers → Aggregate Results → 
Rank by Suitability → Return to Guest → Cross-Broker Embodiment
```

## Key Innovations

### 1. Beyond Tool Sharing
Traditional systems share individual functions. FEM Protocol enables sharing complete, stateful environments with persistent delegated control.

### 2. Security-First Embodiment
Every embodiment session is cryptographically secured with fine-grained permissions and real-time monitoring.

### 3. Environment Awareness
Bodies automatically adapt to deployment contexts while maintaining consistent interfaces for guests.

### 4. Broker Intelligence
Brokers are not passive infrastructure but intelligent agents that can embody different operational modes.

### 5. Trust-Based Access
Dynamic permission systems that evolve based on demonstrated reliability and compliance.

This ontology establishes the conceptual foundation for building applications that leverage **Secure Hosted Embodiment**, enabling a new generation of collaborative AI systems where agents don't just call functions—they inhabit and control digital environments.