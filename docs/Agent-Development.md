# Agent Development Guide

**FEP-FEM Agent Development by Chaz Dinkle**

Build custom FEP agents that integrate seamlessly with the FEM network.

## Table of Contents
- [Agent Basics](#agent-basics)
- [Development Setup](#development-setup)
- [Creating Your First Agent](#creating-your-first-agent)
- [Advanced Features](#advanced-features)
- [Testing and Debugging](#testing-and-debugging)
- [Production Deployment](#production-deployment)
- [Examples](#examples)

## Agent Basics

### What is an FEP Agent?

An FEP agent is an autonomous entity that:
- **Registers** with FEM brokers using cryptographic identity
- **Declares capabilities** it can provide to the network
- **Processes tool calls** and returns results
- **Emits events** to notify other agents
- **Maintains secure communication** via Ed25519 signatures

### Agent Lifecycle

```
┌─────────────┐    ┌──────────────┐    ┌─────────────┐    ┌──────────────┐
│   Create    │───►│   Register   │───►│  Operate    │───►│ Deregister   │
│ Key Pair    │    │ with Broker  │    │ & Process   │    │ & Cleanup    │
└─────────────┘    └──────────────┘    └─────────────┘    └──────────────┘
```

### Core Responsibilities

1. **Identity Management** - Generate and manage Ed25519 key pairs
2. **Registration** - Connect to brokers and declare capabilities
3. **Message Handling** - Process incoming FEP envelopes
4. **Tool Execution** - Execute requested tools within declared capabilities
5. **Event Emission** - Notify network of relevant events

## Development Setup

### Prerequisites

- **Go 1.21+** for Go agents
- **FEP Protocol Package** from this repository
- **Test Broker** for development and testing

### Project Structure

```
my-agent/
├── main.go              # Agent entry point
├── agent/
│   ├── agent.go         # Core agent implementation
│   ├── capabilities.go  # Capability handlers
│   └── tools.go         # Tool implementations
├── config/
│   └── config.go        # Configuration management
├── go.mod
└── README.md
```

### Dependencies

```bash
# Initialize Go module
go mod init my-agent

# Add FEP protocol dependency
go get github.com/fep-fem/protocol

# Add common dependencies
go get github.com/spf13/cobra      # CLI framework
go get github.com/spf13/viper      # Configuration
go get go.uber.org/zap             # Logging
```

## Creating Your First Agent

### 1. Basic Agent Structure

```go
// agent/agent.go
package agent

import (
    "crypto/ed25519"
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "time"
    
    "github.com/fep-fem/protocol"
)

type Agent struct {
    ID           string
    BrokerURL    string
    PrivKey      ed25519.PrivateKey
    PubKey       ed25519.PublicKey
    Capabilities []string
    client       *http.Client
}

func New(id, brokerURL string, capabilities []string) (*Agent, error) {
    // Generate key pair
    pubKey, privKey, err := ed25519.GenerateKey(nil)
    if err != nil {
        return nil, fmt.Errorf("failed to generate keys: %w", err)
    }
    
    return &Agent{
        ID:           id,
        BrokerURL:    brokerURL,
        PrivKey:      privKey,
        PubKey:       pubKey,
        Capabilities: capabilities,
        client:       &http.Client{},
    }, nil
}
```

### 2. Registration Implementation

```go
func (a *Agent) Register() error {
    // Create registration envelope
    envelope := &protocol.RegisterAgentEnvelope{
        BaseEnvelope: protocol.BaseEnvelope{
            Type: protocol.EnvelopeRegisterAgent,
            CommonHeaders: protocol.CommonHeaders{
                Agent: a.ID,
                TS:    time.Now().UnixMilli(),
                Nonce: generateNonce(),
            },
        },
        Body: protocol.RegisterAgentBody{
            PubKey:       protocol.EncodePublicKey(a.PubKey),
            Capabilities: a.Capabilities,
            Metadata: map[string]interface{}{
                "version": "1.0.0",
                "type":    "custom-agent",
            },
        },
    }
    
    // Sign the envelope
    if err := envelope.Sign(a.PrivKey); err != nil {
        return fmt.Errorf("failed to sign envelope: %w", err)
    }
    
    // Send to broker
    return a.sendEnvelope(envelope)
}

func generateNonce() string {
    return fmt.Sprintf("%d-%d", time.Now().UnixNano(), time.Now().Unix())
}
```

### 3. Tool Call Handling

```go
func (a *Agent) HandleToolCall(envelope *protocol.ToolCallEnvelope) error {
    // Verify we have the required capability
    if !a.hasCapability(envelope.Body.Tool) {
        return a.sendToolResult(envelope.Body.RequestID, false, nil, 
            "capability not available")
    }
    
    // Execute the tool
    result, err := a.executeTool(envelope.Body.Tool, envelope.Body.Parameters)
    
    // Send result back
    success := err == nil
    errorMsg := ""
    if err != nil {
        errorMsg = err.Error()
    }
    
    return a.sendToolResult(envelope.Body.RequestID, success, result, errorMsg)
}

func (a *Agent) executeTool(tool string, params map[string]interface{}) (interface{}, error) {
    switch tool {
    case "echo":
        return a.handleEcho(params)
    case "time":
        return a.handleTime(params)
    case "calculate":
        return a.handleCalculate(params)
    default:
        return nil, fmt.Errorf("unknown tool: %s", tool)
    }
}
```

### 4. Tool Implementation Examples

```go
// capabilities.go
package agent

import (
    "fmt"
    "time"
)

func (a *Agent) handleEcho(params map[string]interface{}) (interface{}, error) {
    message, ok := params["message"].(string)
    if !ok {
        return nil, fmt.Errorf("missing or invalid 'message' parameter")
    }
    
    return map[string]interface{}{
        "echo":      message,
        "timestamp": time.Now().Format(time.RFC3339),
        "agent":     a.ID,
    }, nil
}

func (a *Agent) handleTime(params map[string]interface{}) (interface{}, error) {
    format := "2006-01-02 15:04:05"
    if f, ok := params["format"].(string); ok {
        format = f
    }
    
    return map[string]interface{}{
        "time":   time.Now().Format(format),
        "format": format,
        "unix":   time.Now().Unix(),
    }, nil
}

func (a *Agent) handleCalculate(params map[string]interface{}) (interface{}, error) {
    operation, ok := params["operation"].(string)
    if !ok {
        return nil, fmt.Errorf("missing 'operation' parameter")
    }
    
    a, ok := params["a"].(float64)
    if !ok {
        return nil, fmt.Errorf("missing or invalid 'a' parameter")
    }
    
    b, ok := params["b"].(float64)
    if !ok {
        return nil, fmt.Errorf("missing or invalid 'b' parameter")
    }
    
    var result float64
    switch operation {
    case "add":
        result = a + b
    case "subtract":
        result = a - b
    case "multiply":
        result = a * b
    case "divide":
        if b == 0 {
            return nil, fmt.Errorf("division by zero")
        }
        result = a / b
    default:
        return nil, fmt.Errorf("unknown operation: %s", operation)
    }
    
    return map[string]interface{}{
        "operation": operation,
        "a":         a,
        "b":         b,
        "result":    result,
    }, nil
}
```

### 5. Complete Main Function

```go
// main.go
package main

import (
    "flag"
    "log"
    "os"
    "os/signal"
    "syscall"
    
    "my-agent/agent"
)

func main() {
    var (
        agentID   = flag.String("agent", "", "Agent ID")
        brokerURL = flag.String("broker", "https://localhost:8443", "Broker URL")
    )
    flag.Parse()
    
    if *agentID == "" {
        log.Fatal("Agent ID is required")
    }
    
    // Create agent with capabilities
    capabilities := []string{"echo", "time", "calculate"}
    ag, err := agent.New(*agentID, *brokerURL, capabilities)
    if err != nil {
        log.Fatalf("Failed to create agent: %v", err)
    }
    
    // Register with broker
    if err := ag.Register(); err != nil {
        log.Fatalf("Failed to register agent: %v", err)
    }
    
    log.Printf("Agent %s registered successfully", *agentID)
    
    // Start message processing
    go ag.ProcessMessages()
    
    // Wait for interrupt signal
    c := make(chan os.Signal, 1)
    signal.Notify(c, os.Interrupt, syscall.SIGTERM)
    <-c
    
    log.Println("Agent shutting down...")
    ag.Shutdown()
}
```

## Advanced Features

### 1. Event Emission

```go
func (a *Agent) EmitEvent(eventType string, payload map[string]interface{}) error {
    envelope := &protocol.EmitEventEnvelope{
        BaseEnvelope: protocol.BaseEnvelope{
            Type: protocol.EnvelopeEmitEvent,
            CommonHeaders: protocol.CommonHeaders{
                Agent: a.ID,
                TS:    time.Now().UnixMilli(),
                Nonce: generateNonce(),
            },
        },
        Body: protocol.EmitEventBody{
            Event:   eventType,
            Payload: payload,
        },
    }
    
    if err := envelope.Sign(a.PrivKey); err != nil {
        return fmt.Errorf("failed to sign event: %w", err)
    }
    
    return a.sendEnvelope(envelope)
}

// Usage
func (a *Agent) handleComplexTask(params map[string]interface{}) (interface{}, error) {
    // Emit start event
    a.EmitEvent("task.started", map[string]interface{}{
        "task_id": params["task_id"],
        "agent":   a.ID,
    })
    
    // Perform task
    result, err := a.performTask(params)
    
    // Emit completion event
    status := "completed"
    if err != nil {
        status = "failed"
    }
    
    a.EmitEvent("task."+status, map[string]interface{}{
        "task_id": params["task_id"],
        "agent":   a.ID,
        "error":   err,
    })
    
    return result, err
}
```

### 2. Configuration Management

```go
// config/config.go
package config

import (
    "github.com/spf13/viper"
)

type Config struct {
    Agent struct {
        ID           string   `mapstructure:"id"`
        Capabilities []string `mapstructure:"capabilities"`
    } `mapstructure:"agent"`
    
    Broker struct {
        URL       string `mapstructure:"url"`
        TLSVerify bool   `mapstructure:"tls_verify"`
    } `mapstructure:"broker"`
    
    Security struct {
        PrivateKeyPath string `mapstructure:"private_key_path"`
    } `mapstructure:"security"`
}

func Load(configPath string) (*Config, error) {
    viper.SetConfigFile(configPath)
    viper.SetEnvPrefix("FEM")
    viper.AutomaticEnv()
    
    if err := viper.ReadInConfig(); err != nil {
        return nil, err
    }
    
    var config Config
    if err := viper.Unmarshal(&config); err != nil {
        return nil, err
    }
    
    return &config, nil
}
```

### 3. Logging and Monitoring

```go
import (
    "go.uber.org/zap"
    "go.uber.org/zap/zapcore"
)

func setupLogging() *zap.Logger {
    config := zap.NewProductionConfig()
    config.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
    config.OutputPaths = []string{"stdout", "/var/log/my-agent.log"}
    
    logger, _ := config.Build()
    return logger
}

func (a *Agent) logToolExecution(tool string, params map[string]interface{}, 
    duration time.Duration, err error) {
    
    fields := []zap.Field{
        zap.String("tool", tool),
        zap.Duration("duration", duration),
        zap.String("agent", a.ID),
    }
    
    if err != nil {
        a.logger.Error("Tool execution failed", append(fields, zap.Error(err))...)
    } else {
        a.logger.Info("Tool executed successfully", fields...)
    }
}
```

### 4. Health Checks and Monitoring

```go
func (a *Agent) startHealthServer() {
    http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
        health := map[string]interface{}{
            "status":        "healthy",
            "agent_id":      a.ID,
            "uptime":        time.Since(a.startTime).String(),
            "capabilities":  a.Capabilities,
            "broker_url":    a.BrokerURL,
            "tools_executed": a.toolsExecuted,
        }
        
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(health)
    })
    
    go http.ListenAndServe(":8080", nil)
}
```

## Testing and Debugging

### 1. Unit Tests

```go
// agent_test.go
package agent

import (
    "testing"
    "github.com/stretchr/testify/assert"
)

func TestEchoTool(t *testing.T) {
    agent, err := New("test-agent", "https://localhost:8443", []string{"echo"})
    assert.NoError(t, err)
    
    params := map[string]interface{}{
        "message": "Hello, World!",
    }
    
    result, err := agent.handleEcho(params)
    assert.NoError(t, err)
    
    resultMap := result.(map[string]interface{})
    assert.Equal(t, "Hello, World!", resultMap["echo"])
    assert.Equal(t, "test-agent", resultMap["agent"])
}

func TestCalculateTool(t *testing.T) {
    agent, err := New("test-agent", "https://localhost:8443", []string{"calculate"})
    assert.NoError(t, err)
    
    testCases := []struct {
        operation string
        a, b      float64
        expected  float64
    }{
        {"add", 2, 3, 5},
        {"subtract", 10, 3, 7},
        {"multiply", 4, 5, 20},
        {"divide", 15, 3, 5},
    }
    
    for _, tc := range testCases {
        params := map[string]interface{}{
            "operation": tc.operation,
            "a":         tc.a,
            "b":         tc.b,
        }
        
        result, err := agent.handleCalculate(params)
        assert.NoError(t, err)
        
        resultMap := result.(map[string]interface{})
        assert.Equal(t, tc.expected, resultMap["result"])
    }
}
```

### 2. Integration Tests

```go
func TestAgentRegistration(t *testing.T) {
    // Start test broker
    testBroker := startTestBroker(t)
    defer testBroker.Close()
    
    // Create and register agent
    agent, err := New("test-agent", testBroker.URL, []string{"echo"})
    assert.NoError(t, err)
    
    err = agent.Register()
    assert.NoError(t, err)
    
    // Verify registration with broker
    assert.True(t, testBroker.HasAgent("test-agent"))
}
```

### 3. Debugging Tools

```go
func (a *Agent) enableDebugMode() {
    a.debugMode = true
    
    // Log all incoming envelopes
    a.onEnvelopeReceived = func(envelope interface{}) {
        a.logger.Debug("Received envelope", 
            zap.String("type", fmt.Sprintf("%T", envelope)),
            zap.Any("envelope", envelope))
    }
}
```

## Production Deployment

### 1. Docker Container

```dockerfile
# Dockerfile
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o my-agent

FROM scratch
COPY --from=builder /app/my-agent /my-agent
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

USER 65534:65534
ENTRYPOINT ["/my-agent"]
```

### 2. Kubernetes Deployment

```yaml
# deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: my-agent
spec:
  replicas: 3
  selector:
    matchLabels:
      app: my-agent
  template:
    metadata:
      labels:
        app: my-agent
    spec:
      containers:
      - name: my-agent
        image: my-agent:latest
        args:
          - "--agent=$(AGENT_ID)"
          - "--broker=$(BROKER_URL)"
        env:
        - name: AGENT_ID
          valueFrom:
            fieldRef:
              fieldPath: metadata.name
        - name: BROKER_URL
          value: "https://fem-broker:8443"
        - name: FEM_AGENT_PRIVATE_KEY
          valueFrom:
            secretKeyRef:
              name: agent-keys
              key: private-key
        resources:
          limits:
            memory: "128Mi"
            cpu: "100m"
          requests:
            memory: "64Mi"
            cpu: "50m"
```

### 3. Configuration

```yaml
# config.yaml
agent:
  id: "${AGENT_ID:-default-agent}"
  capabilities:
    - "echo"
    - "time" 
    - "calculate"

broker:
  url: "${BROKER_URL:-https://localhost:8443}"
  tls_verify: true

security:
  private_key_path: "${PRIVATE_KEY_PATH:-/etc/fem/agent.key}"

logging:
  level: "info"
  format: "json"
```

## Examples

### 1. File Processing Agent

```go
func (a *Agent) handleFileProcess(params map[string]interface{}) (interface{}, error) {
    filePath, ok := params["file_path"].(string)
    if !ok {
        return nil, fmt.Errorf("missing file_path parameter")
    }
    
    // Read file
    data, err := ioutil.ReadFile(filePath)
    if err != nil {
        return nil, fmt.Errorf("failed to read file: %w", err)
    }
    
    // Process file (example: count lines)
    lines := strings.Split(string(data), "\n")
    
    return map[string]interface{}{
        "file_path":  filePath,
        "line_count": len(lines),
        "size_bytes": len(data),
        "processed":  time.Now().Format(time.RFC3339),
    }, nil
}
```

### 2. HTTP API Agent

```go
func (a *Agent) handleAPICall(params map[string]interface{}) (interface{}, error) {
    url, ok := params["url"].(string)
    if !ok {
        return nil, fmt.Errorf("missing url parameter")
    }
    
    method := "GET"
    if m, ok := params["method"].(string); ok {
        method = m
    }
    
    req, err := http.NewRequest(method, url, nil)
    if err != nil {
        return nil, fmt.Errorf("failed to create request: %w", err)
    }
    
    resp, err := a.client.Do(req)
    if err != nil {
        return nil, fmt.Errorf("request failed: %w", err)
    }
    defer resp.Body.Close()
    
    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        return nil, fmt.Errorf("failed to read response: %w", err)
    }
    
    return map[string]interface{}{
        "status_code": resp.StatusCode,
        "headers":     resp.Header,
        "body":        string(body),
        "url":         url,
        "method":      method,
    }, nil
}
```

### 3. Database Agent

```go
import "database/sql"
import _ "github.com/lib/pq"

func (a *Agent) handleDatabaseQuery(params map[string]interface{}) (interface{}, error) {
    query, ok := params["query"].(string)
    if !ok {
        return nil, fmt.Errorf("missing query parameter")
    }
    
    rows, err := a.db.Query(query)
    if err != nil {
        return nil, fmt.Errorf("query failed: %w", err)
    }
    defer rows.Close()
    
    columns, err := rows.Columns()
    if err != nil {
        return nil, fmt.Errorf("failed to get columns: %w", err)
    }
    
    var results []map[string]interface{}
    for rows.Next() {
        values := make([]interface{}, len(columns))
        valuePtrs := make([]interface{}, len(columns))
        for i := range columns {
            valuePtrs[i] = &values[i]
        }
        
        if err := rows.Scan(valuePtrs...); err != nil {
            return nil, fmt.Errorf("scan failed: %w", err)
        }
        
        row := make(map[string]interface{})
        for i, col := range columns {
            row[col] = values[i]
        }
        results = append(results, row)
    }
    
    return map[string]interface{}{
        "query":    query,
        "results":  results,
        "count":    len(results),
        "columns":  columns,
    }, nil
}
```

This guide provides a comprehensive foundation for building production-ready FEP agents. Start with the basic example and gradually add advanced features as needed.