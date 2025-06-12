# Changelog

All notable changes to the FEP-FEM project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.3.0] - 2025-06-11

### Added - MCP Federation Complete 🚀
- **Complete MCP Integration**: Agents now expose tools via standard MCP protocol
- **Tool Discovery System**: Broker-level MCP tool registry with pattern matching
- **Federation Capabilities**: Cross-agent tool discovery and calling
- **Embodiment Framework**: Environment-specific tool adaptation with body definitions
- **New Protocol Envelopes**: `discoverTools`, `toolsDiscovered`, `embodimentUpdate`
- **MCP Server Integration**: Agents expose tools via HTTP JSON-RPC endpoints
- **Demo Implementation**: Working `demo-mcp-federation.sh` script
- **Integration Tests**: Complete test suite validating full federation loop

### Enhanced
- **Protocol Extension**: Extended from 7 to 10 envelope types
- **Agent Capabilities**: fem-coder now includes MCP server with `--mcp-port` flag
- **Broker Functionality**: Full MCP tool registry with advanced discovery
- **Documentation**: Streamlined roadmap and updated quick start guides

### Technical Implementation
- **Thread-Safe Registry**: Concurrent MCP tool indexing and discovery
- **Pattern Matching**: Wildcard tool discovery (e.g., "math.*", "file.*")
- **Environment Adaptation**: Dynamic tool sets based on deployment context
- **Cryptographic Security**: All MCP federation secured with Ed25519 signatures

### Breaking Changes
- **Agent Registration**: Now includes optional MCP endpoint and body definition
- **Protocol Compatibility**: Maintains backward compatibility with v0.1.x agents

### Documentation
- **Implementation Roadmap**: All phases A-I marked complete with streamlined format
- **Quick Start Guide**: Updated with working demo and manual testing examples
- **API Examples**: Real curl commands that work with current implementation

## [0.1.3] - 2025-06-11

### Added
- Cross-platform automated releases via GitHub Actions
- SHA256 checksums for all release binaries
- Comprehensive release notes with documentation links

### Fixed
- GitHub Actions workflow permissions for release creation
- Router build paths in CI/CD pipeline

### Changed
- Improved release automation with proper binary packaging

## [0.1.2] - 2025-06-11

### Fixed
- Router build configuration in GitHub Actions
- Build paths for components with cmd subdirectories

## [0.1.1] - 2025-06-11

### Fixed
- GitHub Actions workflow YAML syntax errors
- Duplicate trigger configuration

## [0.1.0] - 2025-06-11

### Added
- Complete FEP protocol implementation with all 7 envelope types
- Ed25519 cryptographic signing and verification
- JWT-based capability management system
- fem-broker: HTTP-based message broker with TLS support
- fem-router: Mesh networking component for federation
- fem-coder: Sandboxed code execution agent
- Comprehensive test suite with 24+ unit tests
- Cross-platform build system with Makefile
- Docker configurations for all components
- GitHub Actions CI/CD pipeline
- Complete protocol specification documentation
- Working demo and test scripts

### Security
- End-to-end message signing with Ed25519
- Capability-based authorization system
- TLS 1.3 transport encryption
- Replay attack protection with nonces

### Technical
- Go implementation of core protocol
- Module-based architecture with clean interfaces
- Comprehensive error handling and validation
- Self-signed certificate generation for development

## Release Workflow

This project uses automated releases triggered by version tags:

1. **Create and push a version tag**: `git tag v0.x.y && git push origin v0.x.y`
2. **Automated build**: GitHub Actions builds cross-platform binaries
3. **Release creation**: Automatic release with binaries and checksums
4. **Update changelog**: Add release notes to this file

## Version History Summary

- **v0.3.0**: 🚀 Complete MCP federation implementation
- **v0.2.0**: Previous release
- **v0.1.3**: Production automation and releases
- **v0.1.2-v0.1.1**: Build system fixes  
- **v0.1.0**: Initial complete implementation

[Unreleased]: https://github.com/chazmaniandinkle/FEP-FEM/compare/v0.3.0...HEAD
[0.3.0]: https://github.com/chazmaniandinkle/FEP-FEM/compare/v0.2.0...v0.3.0
[0.2.0]: https://github.com/chazmaniandinkle/FEP-FEM/compare/v0.1.3...v0.2.0
[0.1.3]: https://github.com/chazmaniandinkle/FEP-FEM/compare/v0.1.2...v0.1.3
[0.1.2]: https://github.com/chazmaniandinkle/FEP-FEM/compare/v0.1.1...v0.1.2
[0.1.1]: https://github.com/chazmaniandinkle/FEP-FEM/compare/v0.1.0...v0.1.1
[0.1.0]: https://github.com/chazmaniandinkle/FEP-FEM/releases/tag/v0.1.0