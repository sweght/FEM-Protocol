# Quick Start Guide: MCP Federation in Minutes

Experience the power of federated MCP tools with FEP-FEM. This guide shows you how to transform isolated MCP servers into a discoverable, collaborative network.

## Prerequisites

- **Linux/macOS/Windows** (amd64 or arm64)
- **TLS certificates** (auto-generated for development)

## Option 1: Download Pre-built Binaries

### 1. Download Release

Visit [GitHub Releases](https://github.com/chazmaniandinkle/FEP-FEM/releases/latest) and download the appropriate archive for your platform:

- `fem-v{version}-linux-amd64.tar.gz`
- `fem-v{version}-linux-arm64.tar.gz`
- `fem-v{version}-darwin-amd64.tar.gz` (Intel Mac)
- `fem-v{version}-darwin-arm64.tar.gz` (Apple Silicon)
- `fem-v{version}-windows-amd64.zip`

### 2. Extract Binaries

```bash
# Linux/macOS
tar -xzf fem-v*-linux-amd64.tar.gz
cd fem-v*-linux-amd64/

# Windows (PowerShell)
Expand-Archive fem-v*-windows-amd64.zip
cd fem-v*-windows-amd64/
```

### 3. Start Your First MCP Federation Network

```bash
# Terminal 1: Start the FEP broker (handles MCP tool discovery)
./fem-broker --listen :8443

# Terminal 2: Agent with calculator MCP tools
./fem-coder --broker https://localhost:8443 --agent calculator-001 --mcp-port 8080

# Terminal 3: Agent with file processing tools  
./fem-coder --broker https://localhost:8443 --agent processor-001 --mcp-port 8081

# ✨ Now any agent can discover and use tools from other agents!
```

**What just happened?**
- Broker coordinates MCP tool discovery across the network
- Calculator agent exposes `code.execute` and `shell.run` tools via MCP server on port 8080
- Processor agent exposes the same tools via MCP server on port 8081  
- Both agents can discover and use each other's tools through FEP federation

## Option 2: Build from Source

### 1. Clone Repository

```bash
git clone https://github.com/chazmaniandinkle/FEP-FEM.git
cd FEP-FEM
```

### 2. Build All Components

```bash
# Build everything
make build

# Or build individually
cd broker && go build . && cd ..
cd router && go build ./cmd/fem-router && cd ..
cd bodies/coder && go build ./cmd/fem-coder && cd ../..
```

### 3. Run Test Network

```bash
# Automated test
./test-network.sh

# Or comprehensive demo
./demo-fem-network.sh
```

## Basic Usage

### Starting a Broker

```bash
# Development (self-signed cert)
./fem-broker --listen :8443

# Production (with your TLS cert)
./fem-broker --listen :8443 --cert server.crt --key server.key
```

The broker will:
- Generate a self-signed certificate for development
- Listen for FEP agents on the specified port
- Provide a `/health` endpoint for monitoring

### Connecting an Agent

```bash
# Basic agent connection
./fem-coder --broker https://localhost:8443 --agent my-agent-001

# With specific capabilities
./fem-coder --broker https://broker.example.com:8443 \
            --agent production-coder \
            --capabilities "code.execute,file.read,shell.run"
```

### Verifying the Connection

Check broker logs for:
```
Received registerAgent envelope from my-agent-001
Registered agent my-agent-001 with capabilities [code.execute file.read shell.run]
```

Check agent logs for:
```
Registration successful
Agent my-agent-001 connected to broker
```

## Testing Tool Execution

Once connected, the agent can receive tool calls. Here's how to test manually:

### 1. Send a Tool Call (using curl)

```bash
curl -k -X POST https://localhost:8443 \
  -H "Content-Type: application/json" \
  -d '{
    "type": "toolCall",
    "agent": "orchestrator",
    "ts": '$(date +%s000)',
    "nonce": "test-'$(date +%s)'",
    "sig": "",
    "body": {
      "tool": "code.execute",
      "parameters": {
        "language": "python",
        "code": "print(\"Hello from FEM!\")\nresult = 2 + 2\nprint(f\"Result: {result}\")"
      },
      "requestId": "test-exec-001"
    }
  }'
```

### 2. Check Agent Response

The agent will execute the code and return results through the broker.

## MCP Federation Scenarios

### Scenario 1: Instant MCP Tool Federation (2 minutes)

Transform standalone MCP servers into a federated network:

```bash
# Terminal 1: Start FEP broker
./fem-broker --listen :8443

# Terminal 2: Existing MCP server (running on port 8080)
python your_existing_mcp_server.py &

# Terminal 3: Federate the MCP server  
./fem-agent-wrapper --broker https://localhost:8443 \
  --agent "legacy-mcp-server" \
  --mcp-endpoint "http://localhost:8080/mcp" \
  --auto-register-tools

# Now your MCP tools are discoverable across the FEM network!
```

### Scenario 2: Multi-Environment Agent Embodiment (3 minutes)

Same agent, different tools based on environment:

```bash
# Terminal 1: Broker
./fem-broker --listen :8443

# Terminal 2: Local development embodiment
./fem-coder --broker https://localhost:8443 --agent file-agent \
  --environment "local" \
  --mcp-tools "file.read.filesystem,file.write.filesystem,shell.execute"

# Terminal 3: Simulate cloud migration
./fem-coder --broker https://localhost:8443 --agent file-agent \
  --environment "cloud" \
  --mcp-tools "file.read.s3,file.write.s3,lambda.invoke" \
  --update-embodiment

# Same agent logic, different capabilities based on environment
```

### Scenario 3: Cross-Organization MCP Tool Sharing (5 minutes)

Secure tool sharing between organizations:

```bash
# Organization A setup
./fem-broker --listen :8443 --broker-id "org-a" &
./fem-coder --broker https://localhost:8443 --agent "data-validator" \
  --mcp-tools "data.validate,data.clean" \
  --access-policy "public" &

# Organization B setup  
./fem-broker --listen :8444 --broker-id "org-b" &
./fem-coder --broker https://localhost:8444 --agent "ml-processor" \
  --mcp-tools "ml.train,ml.predict" \
  --access-policy "partners:org-a" &

# Connect brokers for federation
./fem-router --connect-brokers org-a:8443 org-b:8444

# Now Org B can use Org A's validation tools, and vice versa
```

### Scenario 4: Dynamic MCP Tool Discovery (1 minute)

Discover and use any available tool:

```bash
# Query available tools
curl -k https://localhost:8443/api/v1/discover \
  -H "Content-Type: application/json" \
  -d '{
    "query": {
      "capabilities": ["file.*"],
      "environmentType": "local",
      "maxResults": 5
    }
  }'

# Response shows all file-related tools with MCP endpoints:
# {
#   "tools": [
#     {
#       "agentId": "file-agent-001",
#       "mcpEndpoint": "https://agent1:8080/mcp",
#       "capabilities": ["file.read", "file.write"],
#       "mcpTools": [...]
#     }
#   ]
# }

# Use discovered tool via standard MCP protocol
curl -X POST https://agent1:8080/mcp \
  -H "Content-Type: application/json" \
  -d '{
    "method": "tools/call",
    "params": {
      "name": "file.read",
      "arguments": {"path": "/tmp/test.txt"}
    }
  }'
```

### Scenario 5: Agent Collaboration Workflow (3 minutes)

Agents discovering and using each other's MCP tools:

```bash
# Start collaborative agent network
./fem-broker --listen :8443 &

# Data processing pipeline agents
./fem-coder --agent "data-ingester" \
  --mcp-tools "data.fetch.api,data.fetch.csv" &

./fem-coder --agent "data-transformer" \
  --mcp-tools "data.clean,data.normalize,data.validate" &

./fem-coder --agent "data-analyzer" \
  --mcp-tools "stats.analyze,ml.cluster,viz.plot" &

./fem-coder --agent "report-generator" \
  --mcp-tools "report.create,report.export.pdf" &

# Orchestrator agent that uses all the above tools
./fem-coder --agent "pipeline-orchestrator" \
  --mcp-client-only \
  --auto-discover-tools \
  --workflow "data_pipeline.json"

# The orchestrator automatically discovers and chains tools:
# ingester.fetch → transformer.clean → analyzer.stats → generator.report
```

### Scenario 6: Environment-Aware Tool Adaptation (2 minutes)

Agent automatically adapts tools when environment changes:

```bash
# Start adaptive agent
./fem-coder --agent "adaptive-storage" \
  --auto-detect-environment \
  --body-templates "storage_bodies.yaml"

# storage_bodies.yaml defines:
# local: file.read.filesystem, file.write.filesystem
# cloud: file.read.s3, file.write.s3  
# edge: file.read.cache, file.write.batch

# Agent detects AWS environment and automatically embodies cloud tools
# Agent detects local environment and automatically embodies filesystem tools
# Same MCP interface, environment-appropriate implementations
```

## Next Steps

- **[Framework Architecture](FEM-Framework.md)** - Understand the system design
- **[Agent Development](Agent-Development.md)** - Build custom agents
- **[Security Guide](Security.md)** - Secure your deployment
- **[Deployment Guide](Deployment.md)** - Production deployment

## Troubleshooting

### Common Issues

**"Connection refused"**
- Ensure broker is running before starting agents
- Check firewall settings for port 8443
- Verify TLS certificate generation

**"Registration failed"**
- Check agent logs for signature errors
- Ensure broker and agent have compatible protocol versions
- Verify network connectivity

**"Permission denied"**
- Check agent capabilities match requested operations
- Verify broker security policies
- Review signature verification

### Getting Help

- **Issues**: [GitHub Issues](https://github.com/chazmaniandinkle/FEP-FEM/issues)
- **Documentation**: [Complete docs](../README.md#documentation)
- **Examples**: Check `test-network.sh` and `demo-fem-network.sh`

## Example Configurations

### Minimal Broker

```bash
./fem-broker --listen :8443
```

### Production Broker

```bash
./fem-broker \
  --listen :8443 \
  --cert /etc/ssl/certs/fem-broker.crt \
  --key /etc/ssl/private/fem-broker.key \
  --log-level info
```

### Multi-Capability Agent

```bash
./fem-coder \
  --broker https://broker.company.com:8443 \
  --agent "production-coder-$(hostname)" \
  --capabilities "code.execute,file.read,file.write,shell.run" \
  --sandbox-level strict
```

You now have a working FEP-FEM network! 🎉

## 🚀 Run the Complete MCP Federation Demo

For a comprehensive demonstration of the MCP federation capabilities, use our included demo script:

```bash
# Run the complete federation demo
./demo-mcp-federation.sh
```

This demo:
1. **Builds** all components (broker and agents)
2. **Starts** the broker on https://localhost:8443
3. **Launches** two agents with MCP servers on different ports
4. **Demonstrates** tool discovery via the broker
5. **Shows** direct MCP tool calls between agents
6. **Validates** the complete federation loop

### What You'll See

The demo script will output:
```
🚀 Starting FEM MCP Federation Demo
📦 Building broker and coder...
🔄 Starting FEM Broker on https://localhost:8443...
🤖 Starting Agent-1 (calculator) on MCP port 8080...
🤖 Starting Agent-2 (executor) on MCP port 8081...
✅ Network is up. Broker and two agents are running.
🔍 Discovering all 'code.execute' tools via the broker...
📞 Calling the 'code.execute' tool directly on Agent-2's MCP endpoint...
🎉 Demo Complete! The full federation loop is working.
```

### Manual Testing

You can also test the federation manually:

```bash
# Discover tools via broker
curl -k -s -X POST https://localhost:8443/ \
    -H "Content-Type: application/json" \
    -d '{
        "type": "discoverTools",
        "agent": "test-client",
        "ts": '$(date +%s%3N)',
        "nonce": "discover-'$(date +%s)'",
        "body": {
            "query": { "capabilities": ["code.execute"] },
            "requestId": "test-discovery"
        }
    }' | jq .

# Call MCP tool directly
curl -s -X POST http://localhost:8080/mcp \
    -H "Content-Type: application/json" \
    -d '{
        "jsonrpc": "2.0",
        "method": "tools/call",
        "params": {
            "name": "code.execute",
            "arguments": {
                "command": "echo Hello from federated MCP tool!"
            }
        },
        "id": 1
    }' | jq .
```

This demonstrates the complete FEP-FEM vision: **MCP tools that are discoverable, federated, and secure**.