# Quick Start Guide: Secure Hosted Embodiment in Minutes

Experience the power of **Secure Hosted Embodiment** with FEM Protocol. This guide walks you through the three flagship use cases where guest "minds" inhabit host-offered "bodies" to enable new forms of collaborative AI interaction.

## Prerequisites

- **Linux/macOS/Windows** (amd64 or arm64)
- **TLS certificates** (auto-generated for development)
- **MCP-compatible tools** (optional, for advanced scenarios)

## Choose Your Embodiment Journey

### 🎭 Live2D Guest System (2 minutes)
**Virtual avatar control through hosted embodiment**

### 💻 Cross-Device Development (3 minutes)  
**Access your laptop's development environment from your phone**

### 📚 Interactive Storytelling (4 minutes)
**AI storytellers controlling game worlds through secure embodiment**

---

## 🎭 Live2D Guest System

Transform any virtual avatar into a host body that AI agents can inhabit and control.

### 1. Download & Setup

```bash
# Download latest release
wget https://github.com/chazmaniandinkle/FEP-FEM/releases/latest/download/fem-v0.3.0-linux-amd64.tar.gz
tar -xzf fem-v0.3.0-linux-amd64.tar.gz
cd fem-v0.3.0-linux-amd64/
```

### 2. Start the Live2D Embodiment Demo

```bash
# Terminal 1: Start embodiment coordination broker
./fem-broker --listen :8443 --embodiment-enabled

# Terminal 2: Start Live2D host offering avatar body
./fem-live2d-host --broker https://localhost:8443 \
  --agent avatar-host-maya \
  --body live2d-puppet-v1 \
  --avatar-model ./avatars/maya.live2d

# Terminal 3: Start AI guest that wants to control the avatar
./fem-guest-agent --broker https://localhost:8443 \
  --agent ai-storyteller \
  --target-body live2d-puppet-v1 \
  --personality "cheerful anime character"
```

### 3. Watch the Magic ✨

```bash
# The AI guest automatically:
# 1. Discovers the Live2D avatar body
# 2. Requests embodiment access  
# 3. Starts controlling avatar expressions and speech
# 4. Responds to user interactions through the avatar

# Test avatar control manually:
curl -k -X POST https://localhost:8443/fem \
  -H "Content-Type: application/json" \
  -d '{
    "type": "requestEmbodiment",
    "agent": "test-guest",
    "body": {
      "hostAgentId": "avatar-host-maya",
      "bodyId": "live2d-puppet-v1",
      "intendedActions": ["Make avatar wave and say hello"]
    }
  }'
```

**What just happened?**
- 🏠 **Host** offered a Live2D avatar as an embodiment "body"
- 🧠 **Guest** AI discovered and inhabited the avatar body
- 🎮 **Security** ensured guest can only control avatar, not host system
- 🎭 **Result** AI can now control virtual avatar expressions, speech, and animations

---

## 💻 Cross-Device Development

Enable your phone to securely control your laptop's development environment.

### 1. Setup Laptop Host

```bash
# Terminal 1: Broker (can be on laptop or separate server)
./fem-broker --listen :8443

# Terminal 2: Laptop offers development environment as embodiment body
./fem-host-agent --broker https://localhost:8443 \
  --agent laptop-host-alice \
  --body developer-workstation-v1 \
  --mcp-port 8080 \
  --security-policy ./configs/dev-security.json
```

**Security Policy Example** (`configs/dev-security.json`):
```json
{
  "allowedPaths": ["/home/alice/projects/*", "/tmp/fem-workspace/*"],
  "deniedCommands": ["rm -rf", "sudo", "curl", "ssh"],
  "maxSessionDuration": 3600,
  "maxConcurrentGuests": 2
}
```

### 2. Connect from Phone

```bash
# On your phone (using Termux or similar)
./fem-guest-agent --broker https://your-laptop-ip:8443 \
  --agent phone-guest-bob \
  --target-capabilities "shell.execute,file.read,file.write" \
  --device-type mobile
```

### 3. Secure Development from Anywhere

```bash
# Your phone can now securely:

# Check git status
echo "shell.execute git status /home/alice/projects/my-app" | \
  nc localhost 8080

# Read configuration files  
echo "file.read /home/alice/projects/my-app/package.json" | \
  nc localhost 8080

# Start development server (accessible to your phone's browser)
echo "shell.execute 'cd /home/alice/projects/my-app && npm run dev'" | \
  nc localhost 8080
```

**What just happened?**
- 🏠 **Laptop** offered secure development environment as embodiment body
- 📱 **Phone** inhabited the body with delegated control capabilities
- 🔒 **Security** restricted guest to safe paths and commands only
- 🌐 **Result** Full development access from mobile device with laptop security

---

## 📚 Interactive Storytelling 

AI storytellers control game state through secure embodiment sessions.

### 1. Setup Story World Host

```bash
# Terminal 1: Broker
./fem-broker --listen :8443

# Terminal 2: Game/story application offers world control body
./fem-story-host --broker https://localhost:8443 \
  --agent story-world-host \
  --body interactive-story-v1 \
  --world-state ./worlds/fantasy-tavern.json \
  --ui-port 3000
```

### 2. Connect AI Storyteller

```bash
# Terminal 3: AI storyteller requests world control embodiment
./fem-storyteller-agent --broker https://localhost:8443 \
  --agent narrative-ai \
  --target-body interactive-story-v1 \
  --story-style "fantasy-adventure" \
  --personality "mysterious tavern keeper"
```

### 3. Interactive Storytelling Session

```bash
# Open story interface
open http://localhost:3000

# The AI storyteller can now:
# - Control NPC dialogue and actions
# - Modify world state (weather, lighting, mood)
# - Respond to player choices
# - Advance storylines dynamically

# Test story control manually:
curl -X POST http://localhost:3000/api/story/action \
  -H "Content-Type: application/json" \
  -d '{
    "action": "update_world",
    "parameters": {
      "weather": "thunderstorm",
      "mood": "ominous", 
      "npc_dialogue": "A mysterious stranger enters the tavern, rain dripping from their cloak..."
    }
  }'
```

**What just happened?**
- 🏠 **Story App** offered world control as an embodiment body
- 🧠 **AI Storyteller** inhabited the world with narrative control
- 🎮 **Security** ensured AI can only control story elements, not system
- 📖 **Result** Dynamic, AI-driven interactive storytelling experience

---

## Understanding Hosted Embodiment

### The Core Innovation

Traditional approach: **Function calls**
```
AI Agent → Function Call → Tool Response
```

FEM Protocol approach: **Hosted Embodiment**
```
Guest Mind → Discovery → Embodiment Request → Host Body → Delegated Control
```

### Security Model

**Secure Delegated Control** means:
- 🔐 **Cryptographic Identity** - Ed25519 signatures for all agents
- ⏱️ **Time-Limited Sessions** - Embodiment expires automatically
- 🛡️ **Permission Boundaries** - Hosts define exactly what guests can control
- 📝 **Audit Logging** - Every action is logged for review
- 🚫 **Isolation** - Guests cannot access host system beyond granted permissions

### Key Components

- **🧠 Guest Agent** - The "mind" that wants to inhabit and control
- **🏠 Host Agent** - Offers "bodies" (sandboxed capability sets) for embodiment
- **🎭 Body** - A secure collection of MCP tools representing capabilities
- **🌐 Broker** - Coordinates discovery and embodiment requests
- **🎫 Session** - Time-bounded period of delegated control with audit logging

---

## Advanced Scenarios

### Multi-Guest Embodiment

```bash
# One Live2D avatar, multiple AI personalities
./fem-guest-agent --agent cheerful-ai --target-body live2d-puppet-v1 &
./fem-guest-agent --agent serious-ai --target-body live2d-puppet-v1 &

# Host manages personality switching and session coordination
```

### Cross-Device Development Team

```bash
# Team laptop offers development body
./fem-host-agent --body team-dev-env --max-guests 3

# Multiple developers can access from different devices
./fem-guest-agent --agent alice-phone --target-body team-dev-env &
./fem-guest-agent --agent bob-tablet --target-body team-dev-env &
./fem-guest-agent --agent charlie-laptop --target-body team-dev-env &
```

### Collaborative Story Worlds

```bash
# Multiple AIs controlling different aspects of the same story
./fem-storyteller-agent --role "narrator" --target-body story-world &
./fem-storyteller-agent --role "npc-controller" --target-body story-world &
./fem-storyteller-agent --role "environment" --target-body story-world &
```

---

## Next Steps

### Learn the Architecture
- **[FEM Framework](FEM-Framework.md)** - Understanding Broker-as-Agent and hosted embodiment
- **[Protocol Specification](Protocol-Specification.md)** - Complete technical specification

### Build Your Own
- **[Agent Development](Agent-Development.md)** - Create custom host and guest agents
- **[MCP Integration](MCP-Integration.md)** - Transform MCP tools into embodiment bodies
- **[Hosted Embodiment Guide](Hosted-Embodiment-Guide.md)** - Deep dive into host/guest patterns

### Deploy Securely  
- **[Security Guide](Security.md)** - Secure Delegated Control implementation
- **[Deployment Guide](Deployment.md)** - Production deployment patterns

---

## Troubleshooting

### Connection Issues
```bash
# Check broker health
curl -k https://localhost:8443/health

# Verify TLS certificates
openssl s_client -connect localhost:8443
```

### Embodiment Failures
```bash
# Check available bodies
curl -k -X POST https://localhost:8443/fem \
  -d '{"type": "discoverBodies", "body": {"query": {"capabilities": ["*"]}}}'

# Monitor broker logs
tail -f broker.log | grep embodiment
```

### Permission Denied
```bash
# Check security policies
cat configs/dev-security.json

# Review session permissions  
tail -f host.log | grep permission
```

## Get Help

- **GitHub Issues**: [Report problems](https://github.com/chazmaniandinkle/FEP-FEM/issues)
- **Documentation**: [Complete guides](../README.md#documentation) 
- **Examples**: Run `./demo-embodiment.sh` for comprehensive demos

---

## What's Different?

**Before FEM Protocol:**
- Static tool integrations
- Direct function calls
- No session management
- Limited security boundaries

**With FEM Protocol:**
- Dynamic embodiment discovery
- Secure delegated control
- Time-bounded sessions with audit logging
- Cryptographic identity and permission enforcement

You now have **Secure Hosted Embodiment** working! 🎉

The future of AI interaction isn't just about calling functions—it's about minds inhabiting bodies with secure, delegated control.