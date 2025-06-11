# Attribution & Upstream Dependencies
*Federated Embodiment Mesh (FEM) Suite — v0.1.3*

## Authors

**Chaz Dinkle** - *Creator, Protocol Designer, and Lead Developer*
- Federated Embodiment Protocol (FEP) specification
- Federated Embodied Mesh (FEM) framework implementation  
- Core architecture and cryptographic security design
- Cross-platform build and release automation

## License

- **Code**: Apache 2.0 ([LICENSE-CODE](../LICENSE-CODE))
- **Documentation**: CC-BY-SA 4.0 ([LICENSE-DOCS](../LICENSE-DOCS))

This framework builds on a number of open standards and open-source projects. Below is a non-exhaustive list of protocols, libraries, and prior art that influenced or are directly incorporated into FEM/FEP.

| Component / Spec | Upstream Project | License | Where Used |
|------------------|-----------------|---------|------------|
| **JSON (ECMA-404)** | ECMA Intl. | *Public Domain* | Envelope body encoding |
| **WebSocket (RFC 6455)** | IETF | RFC | Primary framing transport |
| **HTTP/2 (RFC 7540)** | IETF | RFC | Optional framing transport |
| **QUIC (RFC 9000)** | IETF | RFC | Optional low-latency transport |
| **TLS 1.3 (RFC 8446)** | IETF | RFC | Mutual authentication & encryption |
| **Ed25519 / EdDSA** | DJB, NIST | *Public Domain* | Envelope signatures |
| **JWT (RFC 7519)** | IETF | RFC | Capability tokens (alternative) |
| **Macaroons** | Google | Apache 2.0 / Paper | Capability tokens (alternative) |
| **Model Context Protocol (MCP)** | Anthropic (OSS spec) | Apache 2.0 | Conceptual foundation; FEP envelopes extend MCP grammar |
| **Embodied Cognition (theory)** | Varela, Thompson, Clark et al. | Academic literature / Fair use | Core philosophical inspiration for "body" + "mind" separation and environment-bound intelligence |
| **BoringSSL / OpenSSL** | Google / OSF | Apache 2.0 / Apache 1.1 | TLS implementation in broker binaries |
| **Docker / OCI** | Docker Inc. / OCI | Apache 2.0 | Container packaging for brokers, routers, bodies |
| **BusyBox / Alpine Linux** | BusyBox | GPL 2.0 / MIT | Base images for fem-broker & fem-coder |
| **gVisor** | Google | Apache 2.0 | Optional sandbox for fem-coder |
| **WASI Preview 1** | Bytecode Alliance | Apache 2.0 | Future WASM body templates |
| **Flow-Based Programming (FBP)** | J. Paul Morrison | Various | Conceptual influence (tool flow) |

## Original Research & Design

The core architectural concepts, protocol design, and implementation of FEP-FEM are original work by **Chaz Dinkle**:

- **FEP Protocol Specification** - Wire-level protocol with 7 envelope types
- **Capability-Based Authorization Model** - Fine-grained permission system
- **Agent-Broker Federation Architecture** - Scalable mesh networking design
- **Ed25519 Signature Integration** - Cryptographic security implementation
- **Sandboxed Execution Framework** - Agent runtime isolation
- **Cross-Platform Release Automation** - CI/CD and distribution system

## Acknowledgments

Special thanks to the open source community and standards organizations whose foundational work enabled this protocol and framework.

If your project or standard is referenced but missing, open a PR against `docs/Attribution.md`.

---
*FEP-FEM © 2025 Chaz Dinkle — Licensed under Apache 2.0 (code) and CC-BY-SA 4.0 (documentation)*