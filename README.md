# 🌐 FEM Protocol: Secure Hosted Embodiment for AI Agents

> **A new paradigm for collaborative AI: where guest "minds" can securely inhabit "bodies" offered by host environments, enabling rich applications through Secure Delegated Control.**

[![Release](https://img.shields.io/github/v/release/chazmaniandinkle/FEP-FEM)](https://github.com/chazmaniandinkle/FEP-FEM/releases)
[![Go Tests](https://github.com/chazmaniandinkle/FEP-FEM/workflows/Build%20and%20Release/badge.svg)](https://github.com/chazmaniandinkle/FEP-FEM/actions)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE-CODE)
[![Documentation](https://img.shields.io/badge/Docs-CC%20BY--SA%204.0-lightgrey.svg)](LICENSE-DOCS)

## 🎯 What is the FEM Protocol?

The **FEM Protocol** enables **Secure Hosted Embodiment**, a revolutionary approach where a host environment can securely offer a "body"—a sandboxed set of capabilities—for a guest agent's "mind" to inhabit and control.

This moves beyond simple tool federation to a model of **Secure Delegated Control**, unlocking powerful new applications:

- 🎭 **Collaborative Virtual Presence**: Your agent inhabits a friend's virtual avatar
- 🎲 **Collaborative Application Control**: External agents co-pilot your local applications  
- 📱 **Cross-Device Embodiment**: Your phone's agent seamlessly controls your laptop's capabilities

**Core Innovation**: Transform isolated MCP tool servers into a global network of embodied experiences where agents don't just call functions—they inhabit and control digital environments.

## 🚀 Three Flagship Use Cases

### 1. 🎭 The Live2D Guest System (Collaborative Virtual Presence)

**Scenario**: Your AI agent inhabits a virtual avatar in a friend's Live2D application for shared virtual presence.

```
Your Agent (Guest Mind) → Friend's Live2D App (Host Environment) → Avatar (Body)
```

- **Host's Role**: Runs Live2D application, offers `Live2D_Puppet_Body` for embodiment
- **Guest's Role**: Agent discovers and requests to inhabit the avatar body
- **Delegated Control**: Guest calls `avatar.set_expression("happy")`, `avatar.speak("Hello!")`, `chat.send_message()`
- **FEM Advantage**: Secure isolation—guest controls avatar state only, zero access to host's files or application code

### 2. 🎲 The Interactive Storyteller Co-op (Collaborative Application Control)

**Scenario**: An external agent acts as "Dungeon Master" for your locally running storytelling application.

```json
{
  "bodyId": "storyteller-coop-v1",
  "description": "Delegates control over Interactive Storyteller application state",
  "mcpTools": [
    {"name": "update_character", "description": "Updates player character attributes"},
    {"name": "update_world", "description": "Updates game world description"},
    {"name": "add_inventory", "description": "Adds items to player inventory"},
    {"name": "add_npc", "description": "Introduces new NPCs"}
  ]
}
```

- **Host's Role**: Runs storytelling app, creates body definition exposing game state controls
- **Guest's Role**: External narrative AI embodies the co-pilot body, influences story and UI
- **Delegated Control**: Guest modifies game state through validated tools, UI automatically re-renders
- **FEM Advantage**: Guest never accesses database directly—only specific, validated state changes

### 3. 📱 Cross-Device Embodiment (Seamless Multi-Device Control)

**Scenario**: Your phone's chat agent discovers and inhabits your laptop's terminal agent for seamless cross-device control.

```
Phone Agent (Guest) → Laptop Terminal Agent (Host) → Development Tools (Body)
```

- **Host's Role**: Laptop runs terminal-based FEM agent offering "developer-workstation" body
- **Guest's Role**: Phone's chat agent discovers laptop and requests embodiment
- **Delegated Control**: Phone agent executes `file.read()`, `shell.execute()`, `git.status()`, `code.run()` on laptop
- **FEM Advantage**: Zero-trust security, no VPN/SSH setup, automatic discovery, fine-grained permissions

## 🏃 Quick Start: Experience Hosted Embodiment

### 30-Second Demo Setup

```bash
# Download and extract
wget https://github.com/chazmaniandinkle/FEP-FEM/releases/latest/download/fem-v0.3.0-linux-amd64.tar.gz
tar -xzf fem-*.tar.gz

# Start broker (coordinates embodiment and discovery)
./fem-broker --listen :8443 &

# Terminal 1: Host offers "code-executor" body
./fem-coder --broker https://localhost:8443 --agent host-laptop \
  --body-id "developer-workstation" --offer-embodiment

# Terminal 2: Guest discovers and inhabits the body  
./fem-coder --broker https://localhost:8443 --agent guest-phone \
  --discover-bodies --embody "developer-workstation"

# ✨ Guest can now securely control host's development tools!
```

### Building a Host Environment

1. **Define Your Body**: Create a body definition specifying what capabilities you're willing to share:

```go
body := BodyDefinition{
    BodyID: "my-terminal-body",
    Description: "Secure terminal access with file operations",
    MCPTools: []MCPToolDef{
        {Name: "file.read", Handler: secureFileRead},
        {Name: "file.write", Handler: secureFileWrite},
        {Name: "shell.execute", Handler: sandboxedShell},
    },
    SecurityPolicy: SecurityPolicy{
        AllowedPaths: []string{"/home/user/projects/*"},
        DeniedCommands: []string{"rm -rf", "sudo"},
    },
}
```

2. **Offer for Embodiment**: Register your body with the FEM broker:

```bash
./fem-agent --broker https://localhost:8443 --offer-body my-terminal-body
```

### Building a Guest Agent

1. **Discover Available Bodies**: Find environments you can inhabit:

```bash
curl -X POST https://localhost:8443/discover-bodies \
  -d '{"capability": "terminal.*"}'
# Returns: Available bodies with terminal capabilities
```

2. **Request Embodiment**: Inhabit a discovered body:

```bash
./fem-agent --broker https://localhost:8443 --embody my-terminal-body
```

## 🛠️ Building From Source

```bash
git clone https://github.com/chazmaniandinkle/FEP-FEM.git
cd FEP-FEM

# Build all components
make build

# Generate development certificates
make gen-certs

# Run the embodiment demo
./demo-hosted-embodiment.sh
```

## 🔑 Core Concepts

### Host, Guest, and Bodies

- **Host**: Environment that offers "bodies" (sandboxed capabilities) for embodiment
- **Guest**: Agent "mind" that can discover and inhabit offered bodies  
- **Body**: Secure, sandboxed set of tools/capabilities offered by a host
- **Embodiment**: Process of a guest mind inhabiting and controlling a host's body

### Secure Delegated Control

The FEM Protocol's revolutionary security model:

```
Traditional RPC: Client → Server (Direct function calls)
FEM Protocol: Guest Mind → Host Body (Delegated control within boundaries)
```

- Guests exercise **delegated control** over specific capabilities
- Hosts retain **ultimate security boundaries** through body definitions
- All interactions are **cryptographically signed** and **capability-verified**

### Broker-as-Agent Architecture  

The FEM broker isn't just infrastructure—it's a first-class agent:

- **Broker's Mind**: Federation management, health checking, load balancing
- **Broker's Body**: Network-level tools for admin and embodiment management
- **Broker's Environment**: Production vs development embodiment policies

## 💼 Real-World Applications

### 1. **Remote Work Revolution**
```
Scenario: Work seamlessly across all your devices
Host: Desktop development environment  
Guest: Phone/tablet agents accessing development tools
Result: Full development capabilities from any device, securely
```

### 2. **Collaborative Gaming**
```
Scenario: AI-powered game masters and co-pilots
Host: Local gaming application
Guest: External narrative AI controlling game state
Result: Dynamic, AI-enhanced gameplay experiences
```

### 3. **Virtual Presence Networks**
```
Scenario: Shared virtual spaces with embodied AI
Host: Virtual world or social application
Guest: Friends' AI agents controlling avatars
Result: Rich, embodied social experiences with AI participants
```

### 4. **Enterprise Tool Ecosystem**
```
Scenario: Secure cross-team capability sharing
Host: Department-specific tools and environments
Guest: Other teams' agents accessing bounded capabilities
Result: Collaboration without exposing sensitive systems
```

## 🛡️ Security-First Embodiment

Every embodiment session is:
- ✅ **Cryptographically Secured** (Ed25519 signatures)
- ✅ **Capability-Bounded** (Fine-grained permissions)
- ✅ **Environment-Isolated** (Sandboxed execution)
- ✅ **Audit-Logged** (Complete embodiment trail)

**Zero-Trust Model**: Guests cryptographically prove identity and receive minimum required capabilities within host-defined boundaries.

## 🏗️ Architecture: Beyond Tool Federation

```
┌─────────────────────────────────────────────────────────────┐
│                   FEM Protocol Network                     │
│                                                             │
│  ┌─────────────────┐              ┌─────────────────┐      │
│  │   Host A        │              │   Host B        │      │
│  │ ┌─────────────┐ │              │ ┌─────────────┐ │      │
│  │ │   Body 1    │ │◄────────────►│ │   Body 3    │ │      │
│  │ │   Body 2    │ │              │ │   Body 4    │ │      │
│  │ └─────────────┘ │              │ └─────────────┘ │      │
│  └─────────┬───────┘              └─────────┬───────┘      │
│            │                                │              │
│       ┌────▼────────────────────────────────▼────┐         │
│       │            FEM Broker                    │         │
│       │     (Embodiment Discovery &              │         │
│       │      Security Coordination)              │         │
│       └────┬────────────────────────────────┬────┘         │
│            │                                │              │
│  ┌─────────▼───────┐              ┌─────────▼───────┐      │
│  │   Guest C       │              │   Guest D       │      │
│  │ (Mobile Agent)  │              │ (Desktop Agent) │      │
│  └─────────────────┘              └─────────────────┘      │
└─────────────────────────────────────────────────────────────┘
```

## 📦 What's Included

- **fem-broker** - Embodiment discovery and security coordinator
- **fem-router** - Mesh networking for multi-broker embodiment
- **fem-coder** - Reference implementation with host/guest capabilities
- **FEM Protocol** - Complete embodiment specification
- **Go SDK** - Build embodied agents
- **Body Templates** - Pre-built embodiment patterns

## 🚧 Current State & Roadmap

### ✅ Available Now (v0.3.0)
- Complete FEM protocol with cryptographic security
- Basic broker and agent implementations
- MCP tool integration and federation
- Environment detection and embodiment
- Cross-platform releases (Linux, macOS, Windows)

### 🔄 Active Development - The Four Phases

#### **Phase 1: The Ubiquitous Agent** (SDKs & Usability)
- Official Client SDKs (Go, Python, TypeScript)
- Registry for "Body Templates"  
- Broker as Secure Proxy model

#### **Phase 2: The Sentient Network** (Broker-as-Agent)
- Broker's "Mind" with Ed25519 identity and policies
- Broker's "Body" with network-level tools
- fem-admin Host Dashboard

#### **Phase 3: The Resilient Mesh** (Scaling & Intelligence)  
- Broker-to-Broker Federation
- LoadBalancer & HealthChecker
- SemanticIndex for AI-powered discovery

#### **Phase 4: Ecosystem & Polish** (Production Ready)
- Advanced embodiment permissions
- Production-grade observability
- Community body template marketplace

## 📚 Documentation

- **[Flagship Use Cases](docs/Flagship-Use-Cases.md)** - Detailed examples of hosted embodiment
- **[Hosted Embodiment Guide](docs/Hosted-Embodiment-Guide.md)** - Build host/guest applications
- **[Implementation Roadmap](docs/Implementation-Roadmap.md)** - Four-phase development plan
- **[Protocol Specification](docs/Protocol-Specification.md)** - Technical FEM protocol details
- **[Framework Architecture](docs/FEM-Framework.md)** - System design and components
- **[Security Model](docs/Security.md)** - Cryptography and trust model

## 🤝 Community & Contributing

The FEM Protocol is open source and we welcome contributions!

- **GitHub Issues** - Report bugs or request features
- **Pull Requests** - Submit improvements  
- **Discussions** - Share your embodiment use cases

See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## 🙏 Acknowledgments

The FEM Protocol builds on ideas from:
- **Model Context Protocol (MCP)** by Anthropic - The foundation for AI tool interfaces
- **Embodied Cognition** theory - Intelligence through environment interaction
- **Capability-based Security** - Fine-grained access control for delegation
- **Actor Model** - Distributed computation through message passing

See [Attribution](docs/Attribution.md) for full credits.

## 📄 License

- **Code**: [Apache 2.0](LICENSE-CODE)
- **Documentation**: [CC-BY-SA 4.0](LICENSE-DOCS)

---

**Ready to experience Secure Hosted Embodiment?** [Get Started →](docs/Quick-Start.md)

*FEM Protocol: Where AI agents don't just call functions—they inhabit worlds.*