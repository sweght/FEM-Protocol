# Federated Embodiment Protocol (FEP) & Federated Embodied Mesh (FEM)

**FEP** is the wire-level protocol for secure federated AI agent communication. **FEM** is the reference framework that implements it.

[![Release](https://img.shields.io/github/v/release/chazmaniandinkle/FEP-FEM)](https://github.com/chazmaniandinkle/FEP-FEM/releases)
[![Go Tests](https://github.com/chazmaniandinkle/FEP-FEM/workflows/Build%20and%20Release/badge.svg)](https://github.com/chazmaniandinkle/FEP-FEM/actions)
[![License](https://img.shields.io/badge/License-CC%20BY--SA%204.0-lightgrey.svg)](https://creativecommons.org/licenses/by-sa/4.0/)

## Quick Start

### Download Pre-built Binaries

Download the latest release for your platform from [GitHub Releases](https://github.com/chazmaniandinkle/FEP-FEM/releases/latest).

### Start a FEM Network

```bash
# Start the broker
./fem-broker --listen :8443

# In another terminal, start an agent
./fem-coder --broker https://localhost:8443 --agent my-coding-agent
```

### Build from Source

```bash
# Clone the repository
git clone https://github.com/chazmaniandinkle/FEP-FEM.git
cd FEP-FEM

# Build all components
make build

# Run the test network
./test-network.sh
```

## Project Status

🚀 **Production Ready** - Complete implementation with cross-platform releases, comprehensive documentation, and proven functionality.

### Core Components

- **fem-broker** - Message broker for agent coordination and federation
- **fem-router** - Mesh networking router for multi-broker deployments  
- **fem-coder** - Sandboxed code execution agent with tool capabilities

### Features

✅ **Complete FEP Protocol** - All 7 envelope types implemented with Ed25519 signatures  
✅ **Cryptographic Security** - End-to-end message signing and verification  
✅ **Capability-Based Authorization** - Fine-grained permission system  
✅ **Cross-Platform Releases** - Linux, macOS, Windows (amd64, arm64)  
✅ **Comprehensive Documentation** - Protocol specification and guides  
✅ **Automated Testing** - 24+ unit tests with CI/CD integration  
✅ **Docker Support** - Container images for all components  

## Documentation

- 📋 [Protocol Specification](docs/Protocol-Specification.md) - Complete FEP protocol documentation
- 🚀 [Quick Start Guide](docs/Quick-Start.md) - Get up and running fast
- 🏗️ [Framework Architecture](docs/FEM-Framework.md) - Deep dive into FEM design
- 🔐 [Security Guide](docs/Security.md) - Cryptography and security model
- 🛠️ [Agent Development](docs/Agent-Development.md) - Build custom agents
- 🌐 [Deployment Guide](docs/Deployment.md) - Production deployment

## Repository Structure

```
├── protocol/go/          # Core FEP protocol implementation (Go)
├── broker/               # FEP message broker
├── router/               # Mesh networking router  
├── bodies/coder/         # Code execution agent
├── docs/                 # Complete documentation
├── .github/workflows/    # CI/CD automation
└── Makefile             # Build system
```

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for development setup, code style, and submission guidelines.

## License & Attribution

This project is licensed under [CC-BY-SA 4.0](LICENSE-DOCS). See [Attribution](docs/Attribution.md) for upstream dependencies and credits.

© 2025 FEM Authors — CC-BY-SA 4.0