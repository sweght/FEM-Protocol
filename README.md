# 🌐 Federated Embodiment Protocol (FEP) & Federated Embodied Mesh (FEM)

> **The missing federation layer for MCP tools — Build networks of intelligent agents that discover, share, and adapt their capabilities securely at scale.**

[![Release](https://img.shields.io/github/v/release/chazmaniandinkle/FEP-FEM)](https://github.com/chazmaniandinkle/FEP-FEM/releases)
[![Go Tests](https://github.com/chazmaniandinkle/FEP-FEM/workflows/Build%20and%20Release/badge.svg)](https://github.com/chazmaniandinkle/FEP-FEM/actions)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE-CODE)
[![Documentation](https://img.shields.io/badge/Docs-CC%20BY--SA%204.0-lightgrey.svg)](LICENSE-DOCS)

## 🎯 What is FEP-FEM?

FEP-FEM doesn't replace MCP—it **federates** it. Think of it as the missing piece that transforms MCP from isolated tool servers into a global network of discoverable, secure, and adaptable AI capabilities.

- **FEP (Protocol)**: Secure federation layer for distributed agent communication
- **FEM (Framework)**: Implements both FEP (for federation) AND MCP (for tool interfaces)  
- **MCP Integration**: Every FEM agent can expose tools via MCP server and consume tools via MCP client

**The Vision**: Any MCP tool, anywhere in the network, discoverable and usable by any compatible agent or LLM—with cryptographic security and environment-aware adaptation.

## 🚀 Why FEP-FEM?

### The Problem with MCP Today
- **Isolated Tools**: Each MCP server is a silo; no discovery mechanism
- **Manual Configuration**: Must hardcode every MCP endpoint 
- **No Federation**: Can't share tools across organizations securely
- **Static Deployment**: Tools can't adapt to different environments

### The FEP-FEM Solution
FEP-FEM provides the **federation infrastructure** that MCP needs:
- 🔍 **Dynamic Discovery**: Find any MCP tool across the federated network
- 🤝 **Secure Sharing**: Share tools across organizations with cryptographic guarantees
- 📈 **Adaptive Embodiment**: Same agent, different tools based on deployment environment
- 🔐 **Zero-Trust Federation**: Every message cryptographically signed and verified

## 💡 Core Innovation: Embodied AI Agents

### Mind + Body + Environment = Embodied Agent

- **Mind**: Your agent's core logic and identity (persistent across environments)
- **Body**: Collection of MCP tools available to the agent (changes with environment)
- **Environment**: Where the agent is deployed (local, cloud, browser, mobile, etc.)

### Example: Universal File Agent

```python
# Same mind, different bodies based on environment
class FileAgent(FEMAgent):
    def embody_local(self):
        # Local development environment
        self.mcp_server.register_tool("file.read", read_from_filesystem)
        self.mcp_server.register_tool("file.write", write_to_filesystem)
        
    def embody_cloud(self):
        # Cloud production environment  
        self.mcp_server.register_tool("file.read", read_from_s3)
        self.mcp_server.register_tool("file.write", write_to_s3)
        
    def embody_browser(self):
        # Browser extension environment
        self.mcp_server.register_tool("file.read", read_from_indexeddb)
        self.mcp_server.register_tool("file.download", download_from_url)

# Other agents always call the same interface
file_content = await any_agent.call_tool("file.read", {"path": "data.txt"})
# But the implementation adapts to environment automatically
```

## 🏃 Quick Start

### 30-Second MCP Federation Demo

```bash
# Download and extract
wget https://github.com/chazmaniandinkle/FEP-FEM/releases/latest/download/fem-v0.1.3-linux-amd64.tar.gz
tar -xzf fem-*.tar.gz

# Start broker (coordinates MCP tool discovery)
./fem-broker --listen :8443 &

# Agent 1: Expose calculator tools via MCP
./fem-coder --broker https://localhost:8443 --agent calculator \
  --mcp-server :3001 --mcp-tools "math.add,math.multiply"

# Agent 2: Discover and use calculator tools
./fem-coder --broker https://localhost:8443 --agent consumer \
  --discover-mcp-tools

# ✨ Agent 2 can now discover and use Agent 1's calculator tools!
```

### Your First Federated MCP Network

1. **Start the FEM Broker** (handles MCP tool discovery and federation):
```bash
./fem-broker --listen :8443
```

2. **Connect Agents with MCP Tools**:
```bash
# Terminal 1: Code execution agent
./fem-coder --broker https://localhost:8443 --agent coder-1 \
  --mcp-server :3001 --mcp-tools "code.python,code.javascript"

# Terminal 2: Data analysis agent  
./fem-coder --broker https://localhost:8443 --agent analyzer-1 \
  --mcp-server :3002 --mcp-tools "data.csv,data.json,data.visualize"

# Terminal 3: File management agent
./fem-coder --broker https://localhost:8443 --agent files-1 \
  --mcp-server :3003 --mcp-tools "file.read,file.write,file.list"
```

3. **Use Any Tool from Any Agent**:
```bash
# Any agent can now discover and use tools from others
curl -X POST https://localhost:8443/discover \
  -d '{"capability": "data.*"}' 
# Returns: agents analyzer-1 has data.csv, data.json, data.visualize

# Direct MCP tool invocation
curl -X POST http://localhost:3002/mcp \
  -d '{"method": "tools/call", "params": {"name": "data.csv", "arguments": {"file": "sales.csv"}}}'
```

## 🔑 Key Concepts

### Agents as MCP Federators
Every FEM agent is simultaneously:
- **MCP Server**: Exposes its capabilities as discoverable tools
- **MCP Client**: Can discover and use tools from other agents
- **FEP Participant**: Participates in secure federation protocol

### Environment-Aware Bodies
Agents adapt their MCP tool collections based on deployment environment:

```
Agent Mind: "Data Processor"
├── Local Body → Tools: [file.read, shell.exec, python.run]
├── Cloud Body → Tools: [s3.read, lambda.invoke, athena.query] 
├── Edge Body → Tools: [sensor.read, cache.local, compress.data]
└── Browser Body → Tools: [indexeddb.read, worker.spawn, canvas.draw]
```

### Cross-Organization Federation
Organizations can securely share MCP tools without exposing infrastructure:

```
Organization A          FEM Federation          Organization B
┌─────────────┐       ┌─────────────┐       ┌─────────────┐
│ MCP Tools:  │◄─────►│   Secure    │◄─────►│ MCP Tools:  │
│ - analyze   │       │  Brokering  │       │ - visualize │
│ - validate  │       │             │       │ - export    │
└─────────────┘       └─────────────┘       └─────────────┘
```

Org A can use Org B's visualization tools without raw data ever leaving Org A.

## 💼 Real-World Use Cases

### 1. **Federated AI Workbench**
```
Research Org: "We have specialized ML analysis tools"
University: "We have large datasets but basic tools"
└─► FEM Federation enables secure collaboration without data transfer
```

### 2. **Multi-Cloud Agent Deployment**
```
Agent: "I need file storage"
FEM: "You're in AWS → here are S3 tools"
FEM: "Now you're in GCP → here are Cloud Storage tools"
└─► Same agent logic, environment-appropriate tools
```

### 3. **Enterprise Tool Marketplace**
```
Engineering: Publishes code.* MCP tools to FEM network
Data Science: Discovers and uses code.python for their pipelines  
DevOps: Uses code.docker for deployment automation
└─► Internal tool reuse without custom integrations
```

## 🛡️ Security-First Federation

Every MCP tool call through FEM is:
- ✅ **Cryptographically Signed** (Ed25519)
- ✅ **Capability Verified** (agents only use declared tools)
- ✅ **Environment Isolated** (tools run in appropriate sandbox)
- ✅ **Audit Logged** (complete federation trail)

**Zero-Trust Model**: Agents cryptographically prove their identity and are only granted minimum required capabilities.

## 🏗️ Architecture: MCP + FEP Integration

```
┌─────────────────────────────────────────────────────────────┐
│                    Your Application                         │
└─────────────────────┬───────────────────────────────────────┘
                      │
┌─────────────────────▼───────────────────────────────────────┐
│                   FEM Framework                             │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐         │
│  │   Agent A   │  │   Agent B   │  │   Agent C   │         │
│  │             │  │             │  │             │         │
│  │ MCP Server  │  │ MCP Server  │  │ MCP Server  │         │ 
│  │ MCP Client  │◄─│ MCP Client  │─►│ MCP Client  │         │
│  └─────┬───────┘  └─────┬───────┘  └─────┬───────┘         │
│        │                │                │                 │
│  ┌─────▼────────────────▼────────────────▼───────┐         │
│  │              FEM Broker                       │         │
│  │      (MCP Tool Discovery & Federation)        │         │
│  └───────────────────┬───────────────────────────┘         │
└──────────────────────┼───────────────────────────────────────┘
                       │
┌──────────────────────▼───────────────────────────────────────┐
│                   FEP Protocol                              │
│        (Secure Federation & Agent Communication)            │
└─────────────────────────────────────────────────────────────┘
```

## 📦 What's Included

- **fem-broker** - MCP tool discovery and federation coordinator
- **fem-router** - Mesh networking for multi-broker federation
- **fem-coder** - Reference agent with MCP server/client
- **FEP Protocol** - Complete federation specification
- **Go SDK** - Build federated MCP agents
- **Body Templates** - Environment-specific MCP tool configurations

## 🚧 Current State & Roadmap

### ✅ Available Now (v0.1.3)
- Complete FEP protocol with cryptographic security
- Basic broker and agent implementations with MCP support
- Cross-platform releases (Linux, macOS, Windows)
- Environment detection and basic embodiment

### 🔄 In Development
- [ ] Full MCP server/client integration in agents
- [ ] Dynamic MCP tool discovery via FEP brokers
- [ ] Cross-broker MCP tool federation
- [ ] Body definition templates and RBAC
- [ ] TypeScript and Python SDKs

### 🎯 Future Vision
- Global MCP tool marketplace
- Visual agent workflow builder
- Integration with Claude Desktop and other MCP clients
- Standards adoption across AI agent frameworks

## 📚 Documentation

- **[Ontology](docs/Ontology.md)** - Core concepts: Mind, Body, Environment, Embodiment
- **[Quick Start Guide](docs/Quick-Start.md)** - Get running in 5 minutes
- **[MCP Integration](docs/MCP-Integration.md)** - Migrate existing MCP tools
- **[Embodiment Guide](docs/Embodiment-Guide.md)** - Environment-specific patterns
- **[Framework Architecture](docs/FEM-Framework.md)** - Technical deep dive
- **[Protocol Specification](docs/Protocol-Specification.md)** - FEP wire protocol
- **[Security Model](docs/Security.md)** - Cryptography and trust model

## 🤝 Community & Contributing

FEP-FEM is open source and we welcome contributions!

- **GitHub Issues** - Report bugs or request features
- **Pull Requests** - Submit improvements  
- **Discussions** - Share ideas and get help

See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## 🙏 Acknowledgments

FEP-FEM builds on ideas from:
- **Model Context Protocol (MCP)** by Anthropic - The standard for AI tool interfaces
- **Embodied Cognition** theory - Intelligence through environment interaction
- **Capability-based Security** - Fine-grained access control
- **Event-driven Architectures** - Reactive, scalable systems

See [Attribution](docs/Attribution.md) for full credits.

## 📄 License

- **Code**: [Apache 2.0](LICENSE-CODE)
- **Documentation**: [CC-BY-SA 4.0](LICENSE-DOCS)

---

**Ready to federate your MCP tools?** [Get Started →](docs/Quick-Start.md)

*FEP-FEM: Where MCP tools meet global federation.*