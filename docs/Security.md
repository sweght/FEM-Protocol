# Security Guide: Secure Hosted Embodiment

The FEM Protocol implements comprehensive security for **Secure Hosted Embodiment** scenarios with cryptographic guarantees, embodiment session isolation, and capability-based authorization.

## Table of Contents
- [Security Model](#security-model)
- [Embodiment Session Security](#embodiment-session-security)
- [Cryptographic Security](#cryptographic-security)
- [Transport Security](#transport-security)
- [Host Security](#host-security)
- [Guest Security](#guest-security)
- [Production Security](#production-security)
- [Threat Model](#threat-model)
- [Security Best Practices](#security-best-practices)

## Security Model

The FEM Protocol security is built on four pillars for hosted embodiment:

1. **Cryptographic Identity** - All agents have Ed25519 identities with message signing
2. **Secure Delegated Control** - Hosts delegate specific control to guests within boundaries
3. **Session Isolation** - Each embodiment session is cryptographically isolated
4. **Audit Trail** - Complete logging of all actions within embodiment sessions

### Security Guarantees

✅ **Message Authenticity** - Every envelope is cryptographically signed  
✅ **Message Integrity** - Tampering is cryptographically detectable  
✅ **Replay Protection** - Nonces and timestamps prevent replay attacks  
✅ **Transport Encryption** - TLS 1.3+ encrypts all network communication  
✅ **Embodiment Isolation** - Sessions are isolated with unique tokens  
✅ **Permission Enforcement** - Every guest action validated against session permissions  
✅ **Resource Limiting** - CPU, memory, and disk usage bounded per session  
✅ **Audit Logging** - Complete trail of all embodiment activities  

## Embodiment Session Security

### Session Token Security

**Session Token Properties**:
- **256-bit Entropy**: Cryptographically random, unpredictable
- **Unique Per Session**: No token reuse across embodiment sessions
- **Time-Bounded**: Automatic expiration with session termination
- **Permission-Linked**: Token binds to specific guest permissions

**Token Lifecycle**:
```
Host grants embodiment → Generate session token → Guest receives token →
Guest includes token in all tool calls → Host validates token → 
Session expires → Token becomes invalid
```

### Permission Model

**Secure Delegated Control** ensures hosts retain ultimate control:

```go
type SessionPermissions struct {
    // File system boundaries
    AllowedPaths     []string `json:"allowedPaths"`
    DeniedPaths      []string `json:"deniedPaths"`
    
    // Command restrictions
    AllowedCommands  []string `json:"allowedCommands"`
    DeniedCommands   []string `json:"deniedCommands"`
    
    // Resource limits
    MaxCPUPercent    int      `json:"maxCpuPercent"`
    MaxMemoryMB      int      `json:"maxMemoryMb"`
    MaxDiskWriteMB   int      `json:"maxDiskWriteMb"`
    
    // Time constraints
    SessionTimeout   duration `json:"sessionTimeout"`
    MaxActionsPerHour int     `json:"maxActionsPerHour"`
}
```

**Permission Validation**:
Every guest action is validated:
1. Session token verification
2. Action within allowed permissions
3. Resource limits not exceeded
4. Time constraints respected

### Session Monitoring

**Real-Time Monitoring**:
- **Resource Usage**: Continuous CPU, memory, disk tracking
- **Action Auditing**: Every tool call logged with parameters and results
- **Violation Detection**: Immediate detection of policy breaches
- **Health Checking**: Session validity and host availability

**Violation Response**:
```go
type ViolationResponse struct {
    Warning     bool   // Warn guest but continue session
    Suspend     bool   // Temporarily suspend session
    Terminate   bool   // Immediately end session
    Report      bool   // Report to broker for reputation tracking
}
```

## Cryptographic Security

### Ed25519 Digital Signatures

**Algorithm Properties**:
- **Key Size**: 32 bytes private key, 32 bytes public key
- **Signature Size**: 64 bytes
- **Security Level**: ~128-bit security equivalent
- **Performance**: ~70,000 signatures/second, ~25,000 verifications/second

**Signing Process for Embodiment**:
1. **Agent Identity**: Each agent has unique Ed25519 keypair
2. **Message Signing**: All envelopes signed before transmission
3. **Session Requests**: Embodiment requests cryptographically authenticated
4. **Tool Calls**: Every tool call within session includes signature

### Session Token Generation

**Secure Random Generation**:
```go
func generateSessionToken() string {
    // 256 bits of cryptographic randomness
    entropy := make([]byte, 32)
    _, err := rand.Read(entropy)
    if err != nil {
        panic("Failed to generate secure random bytes")
    }
    
    // Base64 encode for transport
    return base64.URLEncoding.EncodeToString(entropy)
}
```

### Signature Verification

**Multi-Layer Verification**:
1. **Broker Verification**: Verifies agent identity during registration
2. **Host Verification**: Verifies guest identity during embodiment request
3. **Session Verification**: Verifies guest signatures for each tool call

## Transport Security

### TLS Requirements

**Minimum Standards**:
- **TLS Version**: 1.3 or higher required
- **Certificate Validation**: Full chain validation in production
- **Cipher Suites**: Only secure, modern ciphers allowed
- **Perfect Forward Secrecy**: Required for all connections

**Embodiment-Specific Transport**:
```
Guest → TLS 1.3+ → Host MCP Endpoint
• Session token in headers
• Tool calls over encrypted channel
• Response data encrypted
• Session monitoring over secure channel
```

### Certificate Management

**Development**: Self-signed certificates for local testing
**Production**: Valid certificates from trusted CA required

```bash
# Development certificate generation
openssl req -x509 -newkey rsa:4096 -keyout host.key -out host.crt \
  -days 365 -nodes -subj "/CN=localhost"

# Production with Let's Encrypt
certbot certonly --standalone -d your-host.example.com
```

## Host Security

### Body Definition Security

**Principle of Least Privilege**:
```go
// Good: Restrictive body definition
func createSecureDevBody() BodyDefinition {
    return BodyDefinition{
        BodyID: "secure-dev-v1",
        MCPTools: []MCPToolDef{
            {Name: "file.read", Constraints: "/workspace/*"},
            {Name: "shell.execute", Constraints: "safe-commands-only"},
        },
        SecurityPolicy: SecurityPolicy{
            AllowedPaths: []string{"/workspace/*", "/tmp/guest/*"},
            DeniedCommands: []string{"rm -rf", "sudo", "curl"},
            ResourceLimits: ResourceLimits{
                MaxCPUPercent: 25,
                MaxMemoryMB: 512,
            },
        },
    }
}
```

### Input Validation

**Comprehensive Validation**:
```go
func validateGuestInput(input string, policy SecurityPolicy) error {
    // Length validation
    if len(input) > policy.MaxInputLength {
        return fmt.Errorf("input too long: %d > %d", len(input), policy.MaxInputLength)
    }
    
    // Path traversal protection
    if strings.Contains(input, "..") {
        return fmt.Errorf("path traversal attempt detected")
    }
    
    // Command injection protection
    dangerousPatterns := []string{";", "&&", "||", "|", "`", "$("}
    for _, pattern := range dangerousPatterns {
        if strings.Contains(input, pattern) {
            return fmt.Errorf("potentially dangerous pattern: %s", pattern)
        }
    }
    
    return nil
}
```

### Resource Isolation

**Process Sandboxing**:
```go
type ResourceLimiter struct {
    MaxCPUPercent int
    MaxMemoryMB   int
    MaxProcesses  int
}

func (rl *ResourceLimiter) ExecuteWithLimits(cmd string) (*Result, error) {
    // Create isolated process group
    cmd := exec.Command("sh", "-c", cmd)
    cmd.SysProcAttr = &syscall.SysProcAttr{
        Setpgid: true,
    }
    
    // Apply resource limits
    cmd.SysProcAttr.Credential = &syscall.Credential{
        Uid: guestUID,
        Gid: guestGID,
    }
    
    // Execute with timeout and monitoring
    return rl.executeWithMonitoring(cmd)
}
```

## Guest Security

### Session Management

**Secure Session Handling**:
```go
type EmbodimentClient struct {
    sessionToken   string
    hostEndpoint   string
    permissions    []string
    expiryTime     time.Time
    mcpClient      *MCPClient
    violated       bool
}

func (ec *EmbodimentClient) ValidateSession() error {
    if ec.violated {
        return fmt.Errorf("session marked as violated")
    }
    
    if time.Now().After(ec.expiryTime) {
        return fmt.Errorf("session expired")
    }
    
    if ec.sessionToken == "" {
        return fmt.Errorf("no valid session token")
    }
    
    return nil
}
```

### Error Handling

**Information Disclosure Prevention**:
```go
func (ec *EmbodimentClient) HandleToolError(err error) error {
    // Log full error for debugging (if enabled)
    if debugMode {
        log.Printf("Tool error: %v", err)
    }
    
    // Return sanitized error to prevent information disclosure
    switch {
    case strings.Contains(err.Error(), "permission denied"):
        return fmt.Errorf("insufficient permissions")
    case strings.Contains(err.Error(), "resource limit"):
        return fmt.Errorf("resource limit exceeded")
    case strings.Contains(err.Error(), "session"):
        return fmt.Errorf("session error")
    default:
        return fmt.Errorf("operation failed")
    }
}
```

## Production Security

### Security Configuration

**Production Security Checklist**:
```yaml
# production-security.yml
fem_protocol:
  security:
    # Cryptographic settings
    require_ed25519_signatures: true
    min_signature_age: 30s
    max_signature_age: 300s
    
    # Session security
    max_session_duration: 3600s
    session_token_entropy: 256
    require_session_renewal: true
    
    # Host security
    enable_resource_limiting: true
    enable_audit_logging: true
    max_concurrent_guests: 5
    
    # Network security
    require_tls: true
    min_tls_version: "1.3"
    validate_certificates: true
```

### Monitoring and Alerting

**Security Monitoring**:
```go
type SecurityMonitor struct {
    violationThreshold  int
    alertManager       AlertManager
    auditLogger        AuditLogger
}

func (sm *SecurityMonitor) MonitorSession(session *EmbodimentSession) {
    // Track violation patterns
    if session.ViolationCount > sm.violationThreshold {
        sm.alertManager.SendAlert(AlertCriticalViolations, session)
    }
    
    // Monitor resource abuse
    if session.ResourceUsage.ExceedsLimits() {
        sm.alertManager.SendAlert(AlertResourceAbuse, session)
    }
    
    // Track unusual patterns
    if sm.detectAnomalousActivity(session) {
        sm.alertManager.SendAlert(AlertAnomalousActivity, session)
    }
}
```

### Audit Logging

**Comprehensive Audit Trail**:
```go
type AuditEntry struct {
    Timestamp     time.Time `json:"timestamp"`
    SessionToken  string    `json:"sessionToken"`
    GuestID       string    `json:"guestId"`
    HostID        string    `json:"hostId"`
    Action        string    `json:"action"`
    Parameters    string    `json:"parameters"`
    Result        string    `json:"result"`
    Success       bool      `json:"success"`
    ViolationType string    `json:"violationType,omitempty"`
}

func (al *AuditLogger) LogAction(session *EmbodimentSession, action string, params map[string]interface{}, result interface{}, success bool) {
    entry := AuditEntry{
        Timestamp:    time.Now(),
        SessionToken: session.SessionToken,
        GuestID:      session.GuestID,
        HostID:       session.HostID,
        Action:       action,
        Parameters:   encodeParameters(params),
        Result:       encodeResult(result),
        Success:      success,
    }
    
    al.writeAuditEntry(entry)
}
```

## Threat Model

### Threats and Mitigations

**1. Malicious Guest Agent**
- **Threat**: Guest attempts to exceed granted permissions
- **Mitigation**: Real-time permission validation, resource limiting, session termination

**2. Compromised Host**
- **Threat**: Host provides malicious body definitions
- **Mitigation**: Guest validation of body definitions, reputation tracking, broker oversight

**3. Session Hijacking**
- **Threat**: Attacker attempts to use stolen session tokens
- **Mitigation**: Cryptographic session binding, IP validation, short token lifetimes

**4. Resource Exhaustion**
- **Threat**: Guest consumes excessive host resources
- **Mitigation**: Strict resource limits, real-time monitoring, automatic termination

**5. Information Disclosure**
- **Threat**: Guest accesses unauthorized data
- **Mitigation**: Path restrictions, file system isolation, permission validation

**6. Command Injection**
- **Threat**: Guest injects malicious commands
- **Mitigation**: Input sanitization, command whitelisting, sandboxed execution

### Trust Boundaries

```
┌─────────────────────────────────────────────────────────────┐
│                    Host Environment                         │
│  ┌─────────────────────────────────────────────────────┐   │
│  │              Guest Session                          │   │
│  │  ┌─────────────────────────────────────────────┐   │   │
│  │  │         Sandboxed Execution                 │   │   │
│  │  │  • Resource limits enforced                 │   │   │
│  │  │  • Path restrictions active                 │   │   │
│  │  │  • Command filtering enabled                │   │   │
│  │  └─────────────────────────────────────────────┘   │   │
│  │  Session Token Required for All Actions            │   │
│  └─────────────────────────────────────────────────────┘   │
│  Host retains ultimate control                             │
└─────────────────────────────────────────────────────────────┘
```

## Security Best Practices

### For Host Developers

1. **Minimize Attack Surface**: Only expose necessary tools in body definitions
2. **Validate All Inputs**: Never trust guest-provided data
3. **Implement Resource Limits**: Prevent resource exhaustion attacks
4. **Monitor Sessions**: Track guest behavior for anomalies
5. **Audit Everything**: Log all guest actions for security analysis

### For Guest Developers

1. **Validate Sessions**: Always check session validity before tool calls
2. **Handle Errors Gracefully**: Don't expose sensitive information in error messages
3. **Respect Boundaries**: Stay within granted permissions
4. **Implement Timeouts**: Handle session expiration gracefully
5. **Verify Host Identity**: Ensure connecting to legitimate hosts

### For Network Operators

1. **Use TLS Everywhere**: Encrypt all FEM Protocol communications
2. **Monitor Traffic**: Watch for unusual patterns or attacks
3. **Update Regularly**: Keep all components updated with security patches
4. **Implement Rate Limiting**: Prevent abuse of broker services
5. **Backup Audit Logs**: Ensure audit trails are preserved and protected

The FEM Protocol's security model enables powerful collaboration while maintaining strong security boundaries, ensuring that **Secure Hosted Embodiment** remains both functional and safe.