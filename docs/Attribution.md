# Attribution & Upstream Dependencies
*Federated Embodiment Mesh (FEM) Suite — Draft 0.1.2*

This framework builds on a number of open standards and open‑source projects.  Below is a non‑exhaustive list of protocols, libraries, and prior art that influenced or are directly incorporated into FEM/FEP.

| Component / Spec | Upstream Project | License | Where Used |
|------------------|-----------------|---------|------------|
| **JSON (ECMA‑404)** | ECMA Intl. | *Public Domain* | Envelope body encoding |
| **WebSocket (RFC 6455)** | IETF | RFC | Primary framing transport |
| **HTTP/2 (RFC 7540)** | IETF | RFC | Optional framing transport |
| **QUIC (RFC 9000)** | IETF | RFC | Optional low‑latency transport |
| **TLS 1.3 (RFC 8446)** | IETF | RFC | Mutual authentication & encryption |
| **Ed25519 / EdDSA** | DJB, NIST | *Public Domain* | Envelope signatures |
| **JWT (RFC 7519)** | IETF | RFC | Capability tokens (alternative) |
| **Macaroons** | Google | Apache 2.0 / Paper | Capability tokens (alternative) |
| **Model Context Protocol (MCP)** | Anthropic (OSS spec) | Apache 2.0 | Conceptual foundation; FEP envelopes extend MCP grammar |
| **Embodied Cognition (theory)** | Varela, Thompson, Clark et al. | Academic literature / Fair use | Core philosophical inspiration for “body” + “mind” separation and environment‑bound intelligence |
| **cBoringSSL / OpenSSL** | Google / OSF | Apache 2.0 / Apache 1.1 | TLS implementation in broker binaries |
| **Docker / OCI** | Docker Inc. / OCI | Apache 2.0 | Container packaging for brokers, routers, bodies |
| **BusyBox / Alpine Linux** | BusyBox | GPL 2.0 / MIT | Base images for fem‑broker & fem‑coder |
| **gVisor** | Google | Apache 2.0 | Optional sandbox for fem‑coder |
| **WASI Preview 1** | Bytecode Alliance | Apache 2.0 | Future WASM body templates |
| **Flow‑Based Programming (FBP)** | J. Paul Morrison | Various | Conceptual influence (tool flow) |
| **Provider Router Pattern** | Original author (uploaded docs) | CC‑BY‑SA 4.0 | Routing logic guide |
| **Security Tier Pattern** | Original author (uploaded docs) | CC‑BY‑SA 4.0 | Capability scope design |

If your project or standard is referenced but missing, open a PR against `docs/ATTRIBUTION.md`.

