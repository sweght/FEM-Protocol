# Federated Embodied Mesh (FEM) & Federated Embodiment Protocol (FEP)

**FEP** is the wire-level protocol. **FEM** is the reference framework that implements it.

## Project Status

This repository contains a complete implementation of the FEP-FEM system with the following components:

### Core Protocol Package (`/protocol/go/`)
- **Complete FEP protocol implementation** with all 7 envelope types
- **Ed25519 cryptographic signing** for envelope integrity
- **JWT-based capability management** for authorization
- **Comprehensive test suite** with 24 passing unit tests

### Components
- **fem-broker** (`/broker/`) - HTTP-based FEP broker for agent registration and message routing
- **fem-router** (`/router/`) - Mesh networking component for broker federation
- **fem-coder** (`/bodies/coder/`) - Sandboxed code execution body with tool capabilities

### Build System
- **Makefile** with targets for building, testing, and Docker deployment
- **GitHub Actions CI/CD** for automated testing and builds
- **Docker configurations** for all components

### Documentation
- **Attribution.md** - Project credits and acknowledgments
- **JSON schemas** for FEP envelope validation

## Getting Started

```bash
# Build all components
make build

# Run tests
make test

# Build Docker images
make docker

# Run broker locally
./bin/fem-broker --listen :8443
```

## Repository Structure

```
├── protocol/go/          # Core FEP protocol implementation
├── broker/               # FEP broker implementation  
├── router/               # Mesh networking router
├── bodies/coder/         # Code execution body
├── spec/schemas/         # JSON schema definitions
└── docs/                # Documentation
```

Status: **Implementation Complete** - Ready for integration testing and example development.

© 2025 FEM Authors — CC-BY-SA 4.0