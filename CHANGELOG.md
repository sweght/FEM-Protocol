# Changelog

All notable changes to the FEP-FEM project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Comprehensive protocol documentation and guides
- Changelog for tracking project evolution

### Changed
- Updated README to reflect production-ready status
- Removed outdated references to integration testing phase

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

- **v0.1.3**: Production automation and releases
- **v0.1.2-v0.1.1**: Build system fixes  
- **v0.1.0**: Initial complete implementation

[Unreleased]: https://github.com/chazmaniandinkle/FEP-FEM/compare/v0.1.3...HEAD
[0.1.3]: https://github.com/chazmaniandinkle/FEP-FEM/compare/v0.1.2...v0.1.3
[0.1.2]: https://github.com/chazmaniandinkle/FEP-FEM/compare/v0.1.1...v0.1.2
[0.1.1]: https://github.com/chazmaniandinkle/FEP-FEM/compare/v0.1.0...v0.1.1
[0.1.0]: https://github.com/chazmaniandinkle/FEP-FEM/releases/tag/v0.1.0