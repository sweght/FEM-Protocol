# Security Guide

FEP-FEM implements defense-in-depth security with cryptographic guarantees, sandboxed execution, and capability-based authorization.

## Table of Contents
- [Security Model](#security-model)
- [Cryptographic Security](#cryptographic-security)
- [Transport Security](#transport-security)
- [Agent Sandbox Security](#agent-sandbox-security)
- [Capability Security](#capability-security)
- [Production Security](#production-security)
- [Threat Model](#threat-model)
- [Security Best Practices](#security-best-practices)

## Security Model

FEP-FEM security is built on three pillars:

1. **Cryptographic Integrity** - All messages are signed and verified
2. **Capability-Based Authorization** - Fine-grained permission model
3. **Sandboxed Execution** - Isolated agent runtime environments

### Security Guarantees

✅ **Message Authenticity** - Every envelope is cryptographically signed  
✅ **Message Integrity** - Tampering is cryptographically detectable  
✅ **Replay Protection** - Nonces and timestamps prevent replay attacks  
✅ **Transport Encryption** - TLS 1.3+ encrypts all network communication  
✅ **Capability Enforcement** - Agents can only perform authorized actions  
✅ **Execution Isolation** - Agent code runs in sandboxed environments  

## Cryptographic Security

### Ed25519 Digital Signatures

FEP uses **Ed25519** (EdDSA) for all message signing:

```
Security Level: ~128-bit
Key Size: 32 bytes (256 bits)
Signature Size: 64 bytes
Performance: ~60,000 signatures/second
```

**Why Ed25519?**
- Immune to timing attacks
- Small key and signature sizes
- Fast verification
- Mathematically robust (Curve25519)

### Key Management

#### Agent Key Pairs

```go
// Generate new key pair
pubKey, privKey, err := ed25519.GenerateKey(rand.Reader)

// Store private key securely
privKeyBytes := privKey.Seed()
// Never transmit or log private keys!

// Share public key for verification
pubKeyB64 := base64.StdEncoding.EncodeToString(pubKey)
```

#### Key Storage Best Practices

**Development:**
```bash
# Store in local file (development only)
echo "base64-private-key" > ~/.fem/agent.key
chmod 600 ~/.fem/agent.key
```

**Production:**
```bash
# Use environment variables
export FEM_AGENT_PRIVATE_KEY="base64-encoded-key"

# Or key management service
export FEM_KMS_KEY_ID="arn:aws:kms:region:account:key/key-id"
```

### Signature Process

#### 1. Envelope Creation
```go
envelope := &RegisterAgentEnvelope{
    BaseEnvelope: BaseEnvelope{
        Type:  "registerAgent",
        Agent: "my-agent",
        TS:    time.Now().UnixMilli(),
        Nonce: generateNonce(),
    },
    Body: RegisterAgentBody{
        PubKey: base64.StdEncoding.EncodeToString(pubKey),
        Capabilities: []string{"code.execute"},
    },
}
```

#### 2. Signing
```go
// Remove any existing signature
envelope.Sig = ""

// Serialize to canonical JSON
data, err := json.Marshal(envelope)
if err != nil {
    return err
}

// Generate signature
signature := ed25519.Sign(privKey, data)
envelope.Sig = base64.StdEncoding.EncodeToString(signature)
```

#### 3. Verification
```go
// Extract signature
sig, err := base64.StdEncoding.DecodeString(envelope.Sig)
if err != nil {
    return fmt.Errorf("invalid signature encoding")
}

// Temporarily remove signature for verification
originalSig := envelope.Sig
envelope.Sig = ""
defer func() { envelope.Sig = originalSig }()

// Serialize and verify
data, err := json.Marshal(envelope)
if err != nil {
    return err
}

if !ed25519.Verify(pubKey, data, sig) {
    return fmt.Errorf("signature verification failed")
}
```

### Replay Protection

#### Nonce Generation
```go
func generateNonce() string {
    // Cryptographically secure random number
    randomBytes := make([]byte, 16)
    rand.Read(randomBytes)
    
    // Combine with timestamp for uniqueness
    timestamp := time.Now().UnixNano()
    return fmt.Sprintf("%d-%x", timestamp, randomBytes)
}
```

#### Timestamp Validation
```go
func validateTimestamp(ts int64) error {
    now := time.Now().UnixMilli()
    maxAge := 5 * 60 * 1000 // 5 minutes in milliseconds
    
    if ts > now+maxAge {
        return fmt.Errorf("envelope from future")
    }
    
    if ts < now-maxAge {
        return fmt.Errorf("envelope too old")
    }
    
    return nil
}
```

## Transport Security

### TLS Configuration

#### Broker TLS Setup
```go
// Production configuration
tlsConfig := &tls.Config{
    MinVersion: tls.VersionTLS13,
    CipherSuites: []uint16{
        tls.TLS_AES_256_GCM_SHA384,
        tls.TLS_CHACHA20_POLY1305_SHA256,
        tls.TLS_AES_128_GCM_SHA256,
    },
    CurvePreferences: []tls.CurveID{
        tls.X25519,
        tls.CurveP384,
        tls.CurveP256,
    },
}
```

#### Certificate Management

**Development:**
```go
// Auto-generated self-signed certificate
cert, err := generateSelfSignedCert()
tlsConfig.Certificates = []tls.Certificate{cert}
```

**Production:**
```bash
# Use proper CA-signed certificates
./fem-broker \
  --listen :8443 \
  --cert /etc/ssl/certs/fem-broker.crt \
  --key /etc/ssl/private/fem-broker.key
```

#### Client Certificate Verification (Optional)
```go
tlsConfig.ClientAuth = tls.RequireAndVerifyClientCert
tlsConfig.ClientCAs = caCertPool
```

### Network Security

#### Firewall Configuration
```bash
# Allow only necessary ports
ufw allow 8443/tcp  # Broker HTTPS
ufw deny 8080/tcp   # Block HTTP
ufw enable
```

#### Network Isolation
```yaml
# Docker network isolation
version: '3'
services:
  fem-broker:
    networks:
      - fem-internal
    ports:
      - "8443:8443"
  
  fem-agent:
    networks:
      - fem-internal
    # No external ports

networks:
  fem-internal:
    driver: bridge
    internal: true
```

## Agent Sandbox Security

### Process Isolation

#### Container-Based Sandboxing
```dockerfile
# Minimal runtime environment
FROM scratch
COPY fem-coder /fem-coder
USER 65534:65534  # nobody:nobody
ENTRYPOINT ["/fem-coder"]
```

#### Resource Limits
```go
// CPU and memory limits
cmd := exec.Command("python", script)
cmd.SysProcAttr = &syscall.SysProcAttr{
    Setpgid: true,
}

// Set resource limits
rlimit := syscall.Rlimit{
    Cur: 100 * 1024 * 1024, // 100MB memory
    Max: 100 * 1024 * 1024,
}
syscall.Setrlimit(syscall.RLIMIT_AS, &rlimit)
```

### File System Security

#### Restricted File Access
```go
type SafeFileSystem struct {
    allowedPaths []string
    readOnly     bool
}

func (fs *SafeFileSystem) ReadFile(path string) ([]byte, error) {
    if !fs.isAllowed(path) {
        return nil, fmt.Errorf("access denied: %s", path)
    }
    
    // Additional path traversal protection
    cleanPath := filepath.Clean(path)
    if strings.Contains(cleanPath, "..") {
        return nil, fmt.Errorf("path traversal attempt")
    }
    
    return ioutil.ReadFile(cleanPath)
}
```

#### Temporary Directories
```go
// Create isolated temp directory
tempDir, err := ioutil.TempDir("", "fem-agent-*")
defer os.RemoveAll(tempDir)

// Restrict to temp directory
os.Chdir(tempDir)
```

### Network Restrictions

#### Outbound Connection Filtering
```go
// Block network access by default
func restrictNetwork() error {
    // Use iptables or similar to block agent network access
    cmd := exec.Command("iptables", "-A", "OUTPUT", 
        "-m", "owner", "--uid-owner", "fem-agent",
        "-j", "REJECT")
    return cmd.Run()
}
```

## Capability Security

### Fine-Grained Permissions

#### Capability Hierarchy
```
admin.*                 # Administrative operations
├── admin.agent.revoke  # Revoke agent access
├── admin.broker.config # Modify broker config
└── admin.system.shut   # Shutdown system

code.*                  # Code execution
├── code.execute        # General code execution
├── code.python         # Python-specific execution
└── code.javascript     # JavaScript execution

file.*                  # File operations  
├── file.read           # Read files
├── file.read.logs      # Read log files only
├── file.write          # Write files
└── file.write.temp     # Write to temp only

shell.*                 # Shell operations
├── shell.run           # Execute shell commands
└── shell.read          # Read-only shell access
```

#### Capability Validation
```go
func hasCapability(agent *Agent, required string) bool {
    for _, cap := range agent.Capabilities {
        if cap == required {
            return true
        }
        
        // Check wildcard permissions
        if strings.HasSuffix(cap, "*") {
            prefix := strings.TrimSuffix(cap, "*")
            if strings.HasPrefix(required, prefix) {
                return true
            }
        }
    }
    return false
}
```

### JWT-Based Capabilities (Advanced)

#### Capability Token Creation
```go
token := jwt.NewWithClaims(jwt.SigningMethodEdDSA, jwt.MapClaims{
    "agent": "my-agent",
    "capabilities": []string{"code.execute", "file.read"},
    "expires": time.Now().Add(time.Hour).Unix(),
    "issuer": "fem-broker-001",
})

tokenString, err := token.SignedString(brokerPrivateKey)
```

#### Token Validation
```go
token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
    return brokerPublicKey, nil
})

if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
    capabilities := claims["capabilities"].([]string)
    // Use capabilities for authorization
}
```

## Production Security

### Deployment Security

#### Environment Variables
```bash
# Never log these
export FEM_AGENT_PRIVATE_KEY="$(cat /secure/path/agent.key)"
export FEM_BROKER_TLS_CERT="/etc/ssl/certs/broker.crt"
export FEM_BROKER_TLS_KEY="/etc/ssl/private/broker.key"

# Restrict file permissions
chmod 600 /secure/path/agent.key
chown fem-agent:fem-agent /secure/path/agent.key
```

#### Service Configuration
```ini
# systemd service security
[Service]
User=fem-broker
Group=fem-broker
NoNewPrivileges=yes
ProtectSystem=strict
ProtectHome=yes
PrivateTmp=yes
DynamicUser=yes
```

### Monitoring and Logging

#### Security Event Logging
```go
// Log security events (but never secrets!)
log.Info("Agent registration attempt", 
    "agent", envelope.Agent,
    "ip", r.RemoteAddr,
    "capabilities", body.Capabilities)

// Log signature failures
log.Warn("Signature verification failed",
    "agent", envelope.Agent,
    "ip", r.RemoteAddr,
    "envelope_type", envelope.Type)
```

#### Rate Limiting
```go
type RateLimiter struct {
    requests map[string][]time.Time
    limit    int
    window   time.Duration
}

func (rl *RateLimiter) Allow(clientIP string) bool {
    now := time.Now()
    
    // Clean old requests
    rl.requests[clientIP] = filterRecent(rl.requests[clientIP], 
        now.Add(-rl.window))
    
    if len(rl.requests[clientIP]) >= rl.limit {
        return false
    }
    
    rl.requests[clientIP] = append(rl.requests[clientIP], now)
    return true
}
```

## Threat Model

### Threats Addressed

✅ **Message Tampering** - Ed25519 signatures detect any modification  
✅ **Replay Attacks** - Nonces and timestamps prevent reuse  
✅ **Man-in-the-Middle** - TLS 1.3 ensures transport security  
✅ **Unauthorized Access** - Capability system enforces permissions  
✅ **Code Injection** - Sandboxed execution isolates agent code  
✅ **Resource Exhaustion** - Resource limits prevent DoS  

### Threats Requiring Additional Mitigation

⚠️ **Compromised Agent Keys** - Rotate keys regularly, use short-lived tokens  
⚠️ **Broker Compromise** - Use distributed brokers, monitoring  
⚠️ **Side-Channel Attacks** - Use constant-time implementations  
⚠️ **Social Engineering** - Train operators, use multi-party controls  

### Attack Scenarios

#### 1. Malicious Agent Registration
**Attack**: Attacker tries to register with elevated capabilities  
**Mitigation**: Capability approval workflow, key verification

#### 2. Message Forgery
**Attack**: Attacker attempts to forge agent messages  
**Mitigation**: Ed25519 signatures make forgery computationally infeasible

#### 3. Replay Attack
**Attack**: Attacker captures and replays valid messages  
**Mitigation**: Nonces and timestamp windows prevent replays

#### 4. Sandbox Escape
**Attack**: Agent attempts to escape execution sandbox  
**Mitigation**: Container isolation, resource limits, restricted syscalls

## Security Best Practices

### For Operators

1. **Use Strong TLS Certificates**
   ```bash
   # Generate strong certificates
   openssl ecparam -genkey -name secp384r1 -out server.key
   openssl req -new -x509 -sha256 -key server.key -out server.crt -days 365
   ```

2. **Rotate Keys Regularly**
   ```bash
   # Automated key rotation
   */0 0 * * 0 /usr/local/bin/rotate-fem-keys.sh
   ```

3. **Monitor Security Events**
   ```bash
   # Set up log monitoring
   tail -f /var/log/fem/security.log | grep "SECURITY"
   ```

4. **Network Segmentation**
   ```yaml
   # Separate FEM network
   networks:
     fem-dmz:
       driver: bridge
       ipam:
         config:
           - subnet: 172.20.0.0/16
   ```

### For Developers

1. **Never Log Private Keys**
   ```go
   // WRONG
   log.Debug("Agent key", "key", privKey)
   
   // CORRECT  
   log.Debug("Agent registered", "agent", agentID)
   ```

2. **Validate All Inputs**
   ```go
   if len(envelope.Agent) == 0 || len(envelope.Agent) > 255 {
       return fmt.Errorf("invalid agent ID length")
   }
   ```

3. **Use Secure Random**
   ```go
   // Use crypto/rand, never math/rand for security
   nonce := make([]byte, 32)
   crypto/rand.Read(nonce)
   ```

4. **Fail Securely**
   ```go
   // Fail closed, not open
   if err := verifySignature(envelope); err != nil {
       return fmt.Errorf("access denied")
   }
   ```

### For Agent Developers

1. **Minimize Capabilities**
   ```go
   // Request only what you need
   capabilities := []string{"code.execute", "file.read.logs"}
   ```

2. **Validate Tool Parameters**
   ```go
   func executeCode(params map[string]interface{}) error {
       code, ok := params["code"].(string)
       if !ok || len(code) > maxCodeLength {
           return fmt.Errorf("invalid code parameter")
       }
   }
   ```

3. **Handle Errors Securely**
   ```go
   // Don't leak sensitive information in errors
   if err := sensitiveOperation(); err != nil {
       log.Error("Operation failed", "error", err)
       return fmt.Errorf("operation failed")
   }
   ```

This security model provides strong protection for federated AI agent networks while maintaining usability and performance.