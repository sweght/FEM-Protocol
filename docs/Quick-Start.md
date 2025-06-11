# Quick Start Guide

**FEP-FEM by Chaz Dinkle**

Get up and running with FEP-FEM in minutes.

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

### 3. Start Your First FEM Network

```bash
# Terminal 1: Start the broker
./fem-broker --listen :8443

# Terminal 2: Start a coding agent
./fem-coder --broker https://localhost:8443 --agent my-coding-agent
```

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