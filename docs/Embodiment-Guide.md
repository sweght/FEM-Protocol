# Environment-Aware Embodiment Guide

This guide covers how agents adapt their capabilities based on deployment environments within the FEM Protocol framework. This complements the [Hosted Embodiment Guide](Hosted-Embodiment-Guide.md) which covers host/guest relationships.

## Table of Contents
- [Environment Embodiment Overview](#environment-embodiment-overview)
- [Environment Detection](#environment-detection)
- [Body Adaptation Patterns](#body-adaptation-patterns)
- [Multi-Environment Deployment](#multi-environment-deployment)
- [Environment-Specific Examples](#environment-specific-examples)
- [Best Practices](#best-practices)

## Environment Embodiment Overview

### Core Concept

Environment embodiment enables agents to adapt their tool collections based on their deployment context. A single agent "mind" can operate effectively across different environments by embodying appropriate "bodies" for each context.

```
Agent Mind + Deployment Environment → Environment-Specific Body
```

This differs from hosted embodiment where external guests inhabit host-offered bodies. Here, the same agent adapts itself to different operational contexts.

### Why Environment Embodiment Matters

**Traditional Approach**: Separate implementations for each environment
- Desktop version with file system APIs
- Web version with DOM manipulation
- Cloud version with service integrations
- Mobile version with device sensors

**FEM Protocol Approach**: One agent mind, multiple environment bodies
- Same core logic across all deployments
- Environment-specific tool collections
- Automatic adaptation to deployment context
- Seamless migration between environments

## Environment Detection

### Automatic Environment Detection

```go
type EnvironmentDetector struct {
    detectors []EnvironmentCheck
}

type EnvironmentCheck struct {
    Name     string
    Check    func() bool
    Priority int
}

func (ed *EnvironmentDetector) DetectEnvironment() string {
    checks := []EnvironmentCheck{
        {"cloud-aws", ed.isAWS, 10},
        {"cloud-gcp", ed.isGCP, 10}, 
        {"cloud-azure", ed.isAzure, 10},
        {"container", ed.isContainer, 8},
        {"browser", ed.isBrowser, 7},
        {"mobile", ed.isMobile, 6},
        {"local-development", ed.isLocalDev, 5},
        {"local-production", ed.isLocalProd, 4},
    }
    
    for _, check := range checks {
        if check.Check() {
            return check.Name
        }
    }
    
    return "unknown"
}

func (ed *EnvironmentDetector) isAWS() bool {
    // Check for AWS metadata service
    resp, err := http.Get("http://169.254.169.254/latest/meta-data/")
    return err == nil && resp.StatusCode == 200
}

func (ed *EnvironmentDetector) isBrowser() bool {
    // Check for browser-specific globals
    return js.Global().Get("window").Truthy()
}

func (ed *EnvironmentDetector) isContainer() bool {
    // Check for container indicators
    _, err := os.Stat("/.dockerenv")
    return err == nil
}
```

### Manual Environment Configuration

```yaml
# fem-agent.yml
agent:
  environment:
    type: "cloud-aws"
    region: "us-west-2"
    stage: "production"
    
  embodiment:
    auto_detect: false
    fallback_environment: "local-development"
```

## Body Adaptation Patterns

### Pattern 1: Environment-Specific Tool Sets

**File Operations Across Environments**:

```go
type FileAgent struct {
    FEMAgent
    environment string
}

func (fa *FileAgent) EmbodyForEnvironment(env string) error {
    switch env {
    case "local-development", "local-production":
        return fa.embodyLocal()
    case "cloud-aws":
        return fa.embodyAWS()
    case "cloud-gcp":
        return fa.embodyGCP()
    case "browser":
        return fa.embodyBrowser()
    case "mobile":
        return fa.embodyMobile()
    default:
        return fmt.Errorf("unsupported environment: %s", env)
    }
}

func (fa *FileAgent) embodyLocal() error {
    fa.RegisterTool("file.read", &MCPTool{
        Name: "file.read",
        Handler: fa.readLocalFile,
        InputSchema: localFileSchema,
    })
    
    fa.RegisterTool("file.write", &MCPTool{
        Name: "file.write", 
        Handler: fa.writeLocalFile,
        InputSchema: localFileSchema,
    })
    
    return nil
}

func (fa *FileAgent) embodyAWS() error {
    fa.RegisterTool("file.read", &MCPTool{
        Name: "file.read",
        Handler: fa.readFromS3,
        InputSchema: s3FileSchema,
    })
    
    fa.RegisterTool("file.write", &MCPTool{
        Name: "file.write",
        Handler: fa.writeToS3, 
        InputSchema: s3FileSchema,
    })
    
    // AWS-specific tools
    fa.RegisterTool("file.list_buckets", &MCPTool{
        Name: "file.list_buckets",
        Handler: fa.listS3Buckets,
    })
    
    return nil
}

func (fa *FileAgent) embodyBrowser() error {
    fa.RegisterTool("file.read", &MCPTool{
        Name: "file.read",
        Handler: fa.readFromIndexedDB,
        InputSchema: indexedDBSchema,
    })
    
    fa.RegisterTool("file.download", &MCPTool{
        Name: "file.download",
        Handler: fa.downloadFile,
        InputSchema: downloadSchema,
    })
    
    // Browser-specific tools
    fa.RegisterTool("file.upload", &MCPTool{
        Name: "file.upload",
        Handler: fa.uploadViaForm,
    })
    
    return nil
}
```

### Pattern 2: Capability Scaling

**Computation Capabilities Based on Environment**:

```go
type ComputeAgent struct {
    FEMAgent
    environment string
}

func (ca *ComputeAgent) EmbodyForEnvironment(env string) error {
    baseTools := []MCPTool{
        {Name: "compute.basic", Handler: ca.basicComputation},
    }
    
    switch env {
    case "local-development":
        // Limited capabilities for development
        ca.RegisterTools(baseTools)
        
    case "local-production":
        // Full local capabilities
        ca.RegisterTools(append(baseTools, MCPTool{
            Name: "compute.heavy", Handler: ca.heavyComputation,
        }))
        
    case "cloud-aws":
        // Distributed computing capabilities
        ca.RegisterTools(append(baseTools, 
            MCPTool{Name: "compute.lambda", Handler: ca.lambdaExecution},
            MCPTool{Name: "compute.batch", Handler: ca.batchExecution},
            MCPTool{Name: "compute.parallel", Handler: ca.parallelExecution},
        ))
        
    case "mobile":
        // Energy-efficient capabilities only
        ca.RegisterTools([]MCPTool{
            {Name: "compute.light", Handler: ca.lightComputation},
        })
    }
    
    return nil
}
```

### Pattern 3: Security Context Adaptation

**Security Policies Based on Environment**:

```go
type SecurityPolicy struct {
    AllowedOperations []string
    ResourceLimits    ResourceLimits
    AuditLevel        string
}

func (agent *FEMAgent) getSecurityPolicy(env string) SecurityPolicy {
    switch env {
    case "local-development":
        return SecurityPolicy{
            AllowedOperations: []string{"*"}, // Full access for dev
            ResourceLimits: ResourceLimits{
                MaxCPU: 100,
                MaxMemory: 1024,
            },
            AuditLevel: "minimal",
        }
        
    case "cloud-production":
        return SecurityPolicy{
            AllowedOperations: []string{
                "file.read:/var/app/*",
                "network.http:external",
                "compute.process",
            },
            ResourceLimits: ResourceLimits{
                MaxCPU: 50,
                MaxMemory: 512,
            },
            AuditLevel: "comprehensive",
        }
        
    case "mobile":
        return SecurityPolicy{
            AllowedOperations: []string{
                "file.read:/app/data/*",
                "network.http:wifi_only",
            },
            ResourceLimits: ResourceLimits{
                MaxCPU: 25,
                MaxMemory: 128,
            },
            AuditLevel: "privacy_focused",
        }
    }
}
```

## Multi-Environment Deployment

### Environment Migration

**Seamless Environment Transitions**:

```go
type MigratableAgent struct {
    FEMAgent
    currentEnvironment string
    migrationHandlers  map[string]func() error
}

func (ma *MigratableAgent) MigrateToEnvironment(newEnv string) error {
    log.Printf("Migrating from %s to %s", ma.currentEnvironment, newEnv)
    
    // 1. Prepare for migration
    if err := ma.prepareForMigration(newEnv); err != nil {
        return fmt.Errorf("migration preparation failed: %w", err)
    }
    
    // 2. Save current state
    state, err := ma.serializeState()
    if err != nil {
        return fmt.Errorf("state serialization failed: %w", err)
    }
    
    // 3. Clear current embodiment
    ma.clearCurrentBody()
    
    // 4. Embody for new environment
    if err := ma.EmbodyForEnvironment(newEnv); err != nil {
        return fmt.Errorf("embodiment failed: %w", err)
    }
    
    // 5. Restore state in new environment
    if err := ma.restoreState(state, newEnv); err != nil {
        return fmt.Errorf("state restoration failed: %w", err)
    }
    
    ma.currentEnvironment = newEnv
    log.Printf("Migration to %s completed", newEnv)
    
    return nil
}
```

### Cross-Environment Tool Mapping

**Consistent Interface, Different Implementations**:

```go
type ToolMapper struct {
    mappings map[string]map[string]string
}

func NewToolMapper() *ToolMapper {
    return &ToolMapper{
        mappings: map[string]map[string]string{
            "file.read": {
                "local":     "filesystem_read",
                "cloud-aws": "s3_get_object", 
                "browser":   "indexeddb_get",
                "mobile":    "secure_storage_read",
            },
            "network.request": {
                "local":     "http_client",
                "cloud-aws": "lambda_request",
                "browser":   "fetch_api",
                "mobile":    "network_manager",
            },
        },
    }
}

func (tm *ToolMapper) GetImplementation(tool string, environment string) string {
    if envMap, exists := tm.mappings[tool]; exists {
        if impl, exists := envMap[environment]; exists {
            return impl
        }
    }
    return tool // Fallback to original tool name
}
```

## Environment-Specific Examples

### Example 1: Data Processing Agent

**Local Development Environment**:
```go
func (dpa *DataProcessingAgent) embodyLocalDev() error {
    // Small-scale processing with local files
    dpa.RegisterTool("data.process", &MCPTool{
        Handler: func(data []byte) ([]byte, error) {
            // Single-threaded processing
            return dpa.processInMemory(data)
        },
        ResourceLimits: ResourceLimits{MaxMemory: 100},
    })
    
    dpa.RegisterTool("data.persist", &MCPTool{
        Handler: func(data []byte, filename string) error {
            return os.WriteFile(filepath.Join("/tmp", filename), data, 0644)
        },
    })
    
    return nil
}
```

**Cloud Production Environment**:
```go
func (dpa *DataProcessingAgent) embodyCloudProd() error {
    // Distributed processing with cloud services
    dpa.RegisterTool("data.process", &MCPTool{
        Handler: func(data []byte) ([]byte, error) {
            // Offload to cloud compute
            return dpa.processWithLambda(data)
        },
        ResourceLimits: ResourceLimits{MaxMemory: 1000},
    })
    
    dpa.RegisterTool("data.persist", &MCPTool{
        Handler: func(data []byte, key string) error {
            return dpa.persistToS3(data, key)
        },
    })
    
    // Cloud-specific tools
    dpa.RegisterTool("data.distribute", &MCPTool{
        Handler: dpa.distributeProcessing,
    })
    
    return nil
}
```

### Example 2: Communication Agent

**Browser Environment**:
```json
{
  "bodyId": "comm-agent-browser",
  "environmentType": "browser",
  "mcpTools": [
    {
      "name": "comm.websocket",
      "description": "WebSocket communication",
      "implementation": "browser_websocket"
    },
    {
      "name": "comm.post_message", 
      "description": "Cross-frame communication",
      "implementation": "window_post_message"
    }
  ]
}
```

**Server Environment**:
```json
{
  "bodyId": "comm-agent-server",
  "environmentType": "server",
  "mcpTools": [
    {
      "name": "comm.websocket",
      "description": "WebSocket server",
      "implementation": "server_websocket"
    },
    {
      "name": "comm.message_queue",
      "description": "Async message queue",
      "implementation": "redis_queue"
    },
    {
      "name": "comm.email",
      "description": "Email notifications", 
      "implementation": "smtp_client"
    }
  ]
}
```

## Best Practices

### 1. Environment Detection Strategy

**Layered Detection**:
```go
func DetectEnvironment() string {
    // 1. Explicit configuration (highest priority)
    if env := os.Getenv("FEM_ENVIRONMENT"); env != "" {
        return env
    }
    
    // 2. Platform detection
    if isCloudEnvironment() {
        return detectCloudProvider()
    }
    
    // 3. Runtime detection
    if isBrowserContext() {
        return "browser"
    }
    
    // 4. Default fallback
    return "local-development"
}
```

### 2. Graceful Degradation

**Fallback Capabilities**:
```go
func (agent *FEMAgent) RegisterToolWithFallback(name string, primary MCPTool, fallback MCPTool) {
    if agent.environmentSupports(primary) {
        agent.RegisterTool(name, primary)
    } else {
        agent.RegisterTool(name, fallback)
        log.Printf("Using fallback implementation for %s", name)
    }
}
```

### 3. Environment Validation

**Pre-Embodiment Checks**:
```go
func (agent *FEMAgent) ValidateEnvironment(env string) error {
    requirements := agent.getEnvironmentRequirements(env)
    
    for _, req := range requirements {
        if !req.Check() {
            return fmt.Errorf("environment requirement not met: %s", req.Description)
        }
    }
    
    return nil
}
```

### 4. State Management Across Environments

**Environment-Agnostic State**:
```go
type AgentState struct {
    CoreData     map[string]interface{} `json:"core_data"`
    Environment  string                 `json:"environment"`
    Capabilities []string               `json:"capabilities"`
    Metadata     map[string]string      `json:"metadata"`
}

func (agent *FEMAgent) SerializeState() (*AgentState, error) {
    return &AgentState{
        CoreData:     agent.extractCoreData(),
        Environment:  agent.currentEnvironment,
        Capabilities: agent.getActiveCapabilities(),
        Metadata:     agent.getMetadata(),
    }, nil
}

func (agent *FEMAgent) RestoreState(state *AgentState, newEnvironment string) error {
    // Restore core data
    agent.loadCoreData(state.CoreData)
    
    // Map capabilities to new environment
    mappedCapabilities := agent.mapCapabilities(state.Capabilities, newEnvironment)
    
    // Restore compatible capabilities
    for _, capability := range mappedCapabilities {
        if agent.supportsCapability(capability, newEnvironment) {
            agent.enableCapability(capability)
        }
    }
    
    return nil
}
```

### 5. Resource Management

**Environment-Aware Resource Allocation**:
```go
func (agent *FEMAgent) getAllocatedResources(env string) ResourceAllocation {
    switch env {
    case "mobile":
        return ResourceAllocation{
            MaxMemoryMB: 64,
            MaxCPUPercent: 20,
            MaxNetworkKbps: 100,
        }
    case "cloud-production":
        return ResourceAllocation{
            MaxMemoryMB: 2048,
            MaxCPUPercent: 80,
            MaxNetworkKbps: 10000,
        }
    default:
        return ResourceAllocation{
            MaxMemoryMB: 512,
            MaxCPUPercent: 50,
            MaxNetworkKbps: 1000,
        }
    }
}
```

Environment-aware embodiment enables agents to operate efficiently across diverse deployment contexts while maintaining consistent behavior and interfaces. This creates truly portable AI agents that adapt automatically to their operational environment.