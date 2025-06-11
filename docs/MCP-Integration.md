# MCP Integration Guide: Federating Your Tools

This comprehensive guide shows how to transform your existing MCP tools into federated, discoverable capabilities using FEM, and how to build new agents that leverage the full power of MCP federation.

## Table of Contents
- [Quick Start: 3-Line Federation](#quick-start-3-line-federation)
- [Understanding MCP-FEM Integration](#understanding-mcp-fem-integration)
- [Migration Scenarios](#migration-scenarios)
- [Building New Federated Agents](#building-new-federated-agents)
- [Environment-Specific Embodiment](#environment-specific-embodiment)
- [Cross-Organization Federation](#cross-organization-federation)
- [Security and Access Control](#security-and-access-control)
- [Troubleshooting](#troubleshooting)

## Quick Start: 3-Line Federation

Transform any existing MCP server into a federated tool in under 30 seconds:

### Before: Isolated MCP Server
```python
# your_existing_tool.py
from mcp import MCPServer

server = MCPServer("My Tool", port=8080)

@server.tool("analyze.data")
async def analyze_data(data: str) -> dict:
    return {"result": f"Analyzed: {data}"}

if __name__ == "__main__":
    server.run()  # Only accessible via direct connection
```

### After: Federated MCP Tool
```python
# your_federated_tool.py
from mcp import MCPServer
from fem_sdk import FEMAgent  # ← Line 1: Import FEM

server = MCPServer("My Tool", port=8080)

@server.tool("analyze.data")
async def analyze_data(data: str) -> dict:
    return {"result": f"Analyzed: {data}"}

# ← Line 2: Create FEM agent
agent = FEMAgent("data-analyzer")
agent.expose_mcp_server(server)

async def main():
    await server.start()
    await agent.connect("https://broker.fem.network")  # ← Line 3: Connect to federation
    print("Tool now discoverable across the entire FEM network!")

if __name__ == "__main__":
    import asyncio
    asyncio.run(main())
```

**What You've Gained**:
- ✅ **Global Discovery**: Your tool is now discoverable by any agent in the FEM network
- ✅ **Zero Configuration**: Other agents can find and use your tool without hardcoding endpoints
- ✅ **Load Balancing**: Multiple instances automatically distributed
- ✅ **Security**: Cryptographic signatures on every tool call
- ✅ **Federation**: Cross-organization tool sharing

## Understanding MCP-FEM Integration

### The Architecture
FEM creates a three-layer architecture that maintains full MCP compatibility:

```
┌─────────────────────────────────────────────────────────────┐
│                    Application Layer                        │
│              (Your LLMs and AI Applications)                │
└─────────────────────┬───────────────────────────────────────┘
                      │ Standard MCP Protocol
┌─────────────────────▼───────────────────────────────────────┐
│                   MCP Tool Layer                            │
│     (Your existing MCP servers and clients work unchanged)  │
└─────────────────────┬───────────────────────────────────────┘
                      │ FEM Federation
┌─────────────────────▼───────────────────────────────────────┐
│                 FEP Federation Layer                        │
│        (Discovery, routing, security, embodiment)           │
└─────────────────────────────────────────────────────────────┘
```

### Key Principles

1. **MCP Compatibility**: All existing MCP tools work without modification
2. **Federation Layer**: FEM adds discovery and cross-organization capabilities
3. **Environment Awareness**: Tools adapt to deployment context
4. **Security Integration**: Cryptographic security without breaking MCP standard

## Migration Scenarios

### Scenario 1: Individual Tool Migration

**Situation**: You have a single MCP tool you want to make discoverable.

**Solution**: Add FEM federation wrapper:

```python
# Step 1: Keep your existing MCP implementation
@mcp_server.tool("file.compress")
async def compress_file(file_path: str, format: str = "gzip") -> dict:
    # Your existing implementation
    compressed_path = compress(file_path, format)
    return {"compressed_file": compressed_path, "original_size": orig_size, "compressed_size": new_size}

# Step 2: Add FEM federation
fem_agent = FEMAgent("file-compressor")
fem_agent.expose_mcp_server(mcp_server)
await fem_agent.connect(os.getenv("FEM_BROKER", "https://localhost:8443"))
```

### Scenario 2: Tool Suite Migration

**Situation**: You have multiple related MCP tools that should be grouped together.

**Solution**: Create a unified agent that exposes multiple tools:

```python
class DataProcessingSuite(FEMAgent):
    def __init__(self):
        super().__init__("data-processing-suite")
        self.mcp_server = MCPServer("Data Processing", port=8080)
        
        # Register multiple related tools
        self.register_tools()
        
    def register_tools(self):
        @self.mcp_server.tool("data.clean")
        async def clean_data(data: str) -> str:
            return cleaned_data
            
        @self.mcp_server.tool("data.validate") 
        async def validate_data(data: str, schema: dict) -> dict:
            return validation_result
            
        @self.mcp_server.tool("data.transform")
        async def transform_data(data: str, rules: list) -> str:
            return transformed_data
            
    async def start(self):
        self.expose_mcp_server(self.mcp_server)
        await self.connect("https://broker.fem.network")
```

### Scenario 3: Organization-Wide Migration

**Situation**: Your organization has dozens of MCP tools that need federation.

**Solution**: Systematic migration with centralized discovery:

```python
# migration_manager.py
class OrganizationMCPMigration:
    def __init__(self, broker_url: str):
        self.broker_url = broker_url
        self.agents = {}
        
    async def migrate_tool(self, tool_config: dict):
        """Migrate a single MCP tool to FEM federation"""
        agent = FEMAgent(tool_config["agent_id"])
        
        # Load existing MCP server
        server = await self.load_mcp_server(tool_config["mcp_path"])
        agent.expose_mcp_server(server)
        
        # Apply organization security policies
        agent.set_security_policy(tool_config["security_policy"])
        
        # Connect to federation
        await agent.connect(self.broker_url)
        self.agents[tool_config["agent_id"]] = agent
        
    async def migrate_from_config(self, config_file: str):
        """Migrate all tools defined in configuration"""
        config = load_yaml(config_file)
        for tool in config["tools"]:
            await self.migrate_tool(tool)
            
# organization_tools.yaml
tools:
  - agent_id: "finance-calculator"
    mcp_path: "./finance/calculator_server.py"
    security_policy: "department:finance"
  - agent_id: "data-validator"  
    mcp_path: "./data/validator_server.py"
    security_policy: "department:engineering,data"
```

## Building New Federated Agents

### Agent Design Patterns

#### Pattern 1: Simple Tool Provider
```python
class CalculatorAgent(FEMAgent):
    """Simple agent that provides calculation tools"""
    
    def __init__(self):
        super().__init__("calculator")
        self.mcp_server = MCPServer("Calculator", port=8080)
        self.setup_tools()
        
    def setup_tools(self):
        @self.mcp_server.tool("math.add")
        async def add(a: float, b: float) -> float:
            return a + b
            
        @self.mcp_server.tool("math.factorial")
        async def factorial(n: int) -> int:
            if n <= 1:
                return 1
            return n * await self.factorial(n - 1)
```

#### Pattern 2: Tool Consumer and Provider
```python
class DataPipelineAgent(FEMAgent):
    """Agent that uses other agents' tools to build data pipelines"""
    
    def __init__(self):
        super().__init__("data-pipeline")
        self.mcp_server = MCPServer("Data Pipeline", port=8081)
        self.mcp_client = MCPClient()
        self.setup_tools()
        
    def setup_tools(self):
        @self.mcp_server.tool("pipeline.process")
        async def process_data(input_data: str, pipeline_config: dict) -> dict:
            results = {}
            
            # Use remote tools via MCP federation
            for step in pipeline_config["steps"]:
                tool_agent = await self.discover_tool(step["tool"])
                step_result = await self.mcp_client.call_tool(
                    tool_agent.mcp_endpoint,
                    step["tool"],
                    step["parameters"]
                )
                results[step["name"]] = step_result
                
            return {"pipeline_results": results}
```

#### Pattern 3: Environment-Adaptive Agent
```python
class UniversalFileAgent(FEMAgent):
    """Agent that adapts file operations to environment"""
    
    def __init__(self):
        super().__init__("universal-file")
        self.mcp_server = MCPServer("Universal File Agent", port=8082)
        
    async def embody(self, environment: str):
        """Adapt tools based on environment"""
        if environment == "local":
            await self.embody_local()
        elif environment == "cloud":
            await self.embody_cloud()
        elif environment == "mobile":
            await self.embody_mobile()
            
    async def embody_local(self):
        @self.mcp_server.tool("file.read")
        async def read_file(path: str) -> str:
            with open(path, 'r') as f:
                return f.read()
                
        @self.mcp_server.tool("file.write")
        async def write_file(path: str, content: str) -> bool:
            with open(path, 'w') as f:
                f.write(content)
            return True
            
    async def embody_cloud(self):
        @self.mcp_server.tool("file.read")
        async def read_s3(path: str) -> str:
            bucket, key = path.split('/', 1)
            return await s3_client.get_object(bucket, key)
            
        @self.mcp_server.tool("file.write")
        async def write_s3(path: str, content: str) -> bool:
            bucket, key = path.split('/', 1)
            await s3_client.put_object(bucket, key, content)
            return True
```

## Environment-Specific Embodiment

### Body Definition Templates

Create reusable body definitions for different environments:

```json
{
  "bodyTemplates": {
    "local-development": {
      "description": "Local development environment with full filesystem access",
      "mcpTools": [
        {"name": "file.read", "implementation": "filesystem"},
        {"name": "file.write", "implementation": "filesystem"},
        {"name": "shell.execute", "implementation": "subprocess"},
        {"name": "git.status", "implementation": "git-cli"}
      ],
      "securityPolicy": {
        "allowedPaths": ["/tmp", "/home/user/workspace"],
        "networkAccess": "full",
        "shellAccess": true
      }
    },
    "cloud-production": {
      "description": "Cloud production environment with service integrations",
      "mcpTools": [
        {"name": "file.read", "implementation": "s3"},
        {"name": "file.write", "implementation": "s3"},
        {"name": "db.query", "implementation": "rds"},
        {"name": "message.send", "implementation": "sqs"}
      ],
      "securityPolicy": {
        "allowedServices": ["s3", "rds", "sqs"],
        "networkAccess": "restricted",
        "shellAccess": false
      }
    },
    "edge-device": {
      "description": "Resource-constrained edge device",
      "mcpTools": [
        {"name": "sensor.read", "implementation": "gpio"},
        {"name": "data.compress", "implementation": "lightweight"},
        {"name": "cache.store", "implementation": "local-storage"}
      ],
      "securityPolicy": {
        "maxMemory": "256MB",
        "maxConcurrency": 2,
        "networkAccess": "minimal"
      }
    }
  }
}
```

### Dynamic Embodiment Implementation

```python
class AdaptiveAgent(FEMAgent):
    def __init__(self, agent_id: str):
        super().__init__(agent_id)
        self.body_templates = load_body_templates()
        
    async def auto_embody(self):
        """Automatically detect environment and embody appropriately"""
        env = await self.detect_environment()
        template = self.select_body_template(env)
        await self.embody_from_template(template)
        
    async def detect_environment(self) -> dict:
        """Detect current deployment environment"""
        env = {"type": "unknown", "capabilities": []}
        
        # Detect cloud providers
        if os.getenv("AWS_REGION"):
            env["type"] = "aws-cloud"
            env["capabilities"].extend(["s3", "rds", "lambda"])
            
        elif os.getenv("GOOGLE_CLOUD_PROJECT"):
            env["type"] = "gcp-cloud"
            env["capabilities"].extend(["gcs", "bigquery", "functions"])
            
        # Detect container environments
        elif os.path.exists("/.dockerenv"):
            env["type"] = "container"
            env["capabilities"].extend(["filesystem", "network"])
            
        # Default to local
        else:
            env["type"] = "local"
            env["capabilities"].extend(["filesystem", "shell", "network"])
            
        return env
        
    def select_body_template(self, env: dict) -> dict:
        """Select appropriate body template for environment"""
        if env["type"].endswith("-cloud"):
            return self.body_templates["cloud-production"]
        elif env["type"] == "container":
            return self.body_templates["container-production"]
        else:
            return self.body_templates["local-development"]
```

## Cross-Organization Federation

### Setting Up Secure Inter-Organization Tool Sharing

#### Organization A: Tool Provider
```python
# org_a_setup.py
class OrgAToolProvider:
    def __init__(self):
        self.broker = FEMBroker("org-a-broker", port=8443)
        self.agents = []
        
    async def setup_public_tools(self):
        """Setup tools that can be shared with partners"""
        
        # Data validation service (public)
        validator = DataValidatorAgent("data-validator")
        validator.set_access_policy("public")  # Anyone can use
        await validator.connect(self.broker.url)
        
        # ML inference service (partner-only)
        ml_agent = MLInferenceAgent("ml-inference")
        ml_agent.set_access_policy("partners:org-b,org-c")  # Only specific orgs
        await ml_agent.connect(self.broker.url)
        
        # Internal analysis tools (private)
        analyzer = InternalAnalyzer("internal-analyzer")
        analyzer.set_access_policy("internal-only")  # Org A only
        await analyzer.connect(self.broker.url)
        
    async def setup_federation(self):
        """Connect to partner organizations"""
        await self.broker.federate_with("https://org-b-broker.example.com:8443")
        await self.broker.federate_with("https://org-c-broker.example.com:8443")
```

#### Organization B: Tool Consumer
```python
# org_b_setup.py
class OrgBToolConsumer:
    def __init__(self):
        self.broker = FEMBroker("org-b-broker", port=8444)
        
    async def create_workflow_agent(self):
        """Create agent that uses cross-org tools"""
        workflow = WorkflowAgent("cross-org-workflow")
        
        @workflow.mcp_server.tool("analysis.comprehensive")
        async def comprehensive_analysis(data: str) -> dict:
            # Discover and use Org A's validation tool
            validators = await workflow.discover_tools("data.validate", source="org-a")
            validation = await workflow.call_remote_tool(validators[0], {"data": data})
            
            # Use Org A's ML inference (if authorized)
            ml_tools = await workflow.discover_tools("ml.infer", source="org-a")
            if ml_tools:
                inference = await workflow.call_remote_tool(ml_tools[0], {"data": data})
            else:
                inference = {"error": "ML inference not authorized"}
                
            return {
                "validation": validation,
                "ml_inference": inference,
                "source_org": "org-a"
            }
            
        await workflow.connect(self.broker.url)
```

### Federation Security Patterns

#### Role-Based Cross-Organization Access
```python
class FederationSecurityManager:
    def __init__(self, org_id: str):
        self.org_id = org_id
        self.partner_policies = {}
        
    def setup_partner_policy(self, partner_org: str, policy: dict):
        """Define what tools can be shared with specific partners"""
        self.partner_policies[partner_org] = {
            "allowed_tools": policy.get("allowed_tools", []),
            "denied_tools": policy.get("denied_tools", []),
            "rate_limits": policy.get("rate_limits", {}),
            "data_classification": policy.get("data_classification", "public")
        }
        
    async def authorize_cross_org_call(self, calling_org: str, tool: str) -> bool:
        """Check if cross-org tool call is authorized"""
        if calling_org not in self.partner_policies:
            return False
            
        policy = self.partner_policies[calling_org]
        
        # Check tool allowlist
        if policy["allowed_tools"] and tool not in policy["allowed_tools"]:
            return False
            
        # Check tool denylist
        if tool in policy.get("denied_tools", []):
            return False
            
        # Check rate limits
        if not await self.check_rate_limits(calling_org, tool):
            return False
            
        return True

# Example usage
security = FederationSecurityManager("org-a")
security.setup_partner_policy("org-b", {
    "allowed_tools": ["data.validate", "file.compress"],
    "denied_tools": ["internal.*", "admin.*"],
    "rate_limits": {"requests_per_hour": 1000},
    "data_classification": "partner-shared"
})
```

## Security and Access Control

### Capability-Based Security for MCP Tools

```python
class SecureMCPAgent(FEMAgent):
    def __init__(self, agent_id: str):
        super().__init__(agent_id)
        self.capability_checker = CapabilityChecker()
        
    def register_secure_tool(self, tool_name: str, required_caps: list, handler):
        """Register MCP tool with capability requirements"""
        
        @self.mcp_server.tool(tool_name)
        async def secure_handler(*args, **kwargs):
            # Get calling agent info from FEP context
            caller = self.get_calling_agent()
            
            # Check if caller has required capabilities
            if not self.capability_checker.has_capabilities(caller, required_caps):
                raise PermissionError(f"Agent {caller.id} lacks required capabilities: {required_caps}")
                
            # Log access for audit trail
            self.audit_log.log_tool_access(caller.id, tool_name, args, kwargs)
            
            # Execute tool
            return await handler(*args, **kwargs)

# Example: File agent with capability-based access
file_agent = SecureMCPAgent("secure-file-agent")

file_agent.register_secure_tool(
    "file.read",
    required_caps=["file.read"],
    handler=read_file_handler
)

file_agent.register_secure_tool(
    "file.delete", 
    required_caps=["file.write", "file.delete"],
    handler=delete_file_handler
)
```

### Audit Logging and Compliance

```python
class MCPAuditLogger:
    def __init__(self, agent_id: str):
        self.agent_id = agent_id
        self.log_handler = configure_audit_logging()
        
    def log_tool_access(self, caller_id: str, tool: str, args: tuple, kwargs: dict):
        """Log all MCP tool access for compliance"""
        audit_entry = {
            "timestamp": datetime.utcnow().isoformat(),
            "agent_id": self.agent_id,
            "caller_id": caller_id,
            "tool": tool,
            "parameters": self.sanitize_parameters(args, kwargs),
            "source_ip": self.get_caller_ip(caller_id),
            "session_id": self.get_session_id(caller_id)
        }
        
        self.log_handler.info(json.dumps(audit_entry))
        
    def sanitize_parameters(self, args: tuple, kwargs: dict) -> dict:
        """Remove sensitive data from parameters for logging"""
        sanitized = {}
        
        for key, value in kwargs.items():
            if key.lower() in ['password', 'token', 'secret', 'api_key']:
                sanitized[key] = "***REDACTED***"
            else:
                sanitized[key] = str(value)[:100]  # Truncate long values
                
        return sanitized
```

## Troubleshooting

### Common Issues and Solutions

#### Issue 1: MCP Tool Discovery Fails
**Symptoms**: Agents can't find each other's tools
**Causes**: 
- Agent not properly registered with broker
- Network connectivity issues
- Capability name mismatches

**Solutions**:
```python
# Debug tool discovery
async def debug_tool_discovery(agent: FEMAgent):
    # Check broker connection
    status = await agent.check_broker_connection()
    print(f"Broker connection: {status}")
    
    # List registered capabilities
    caps = await agent.get_registered_capabilities()
    print(f"Registered capabilities: {caps}")
    
    # Test tool discovery
    available_tools = await agent.discover_tools("*")
    print(f"Available tools: {len(available_tools)}")
    for tool in available_tools:
        print(f"  - {tool.name} from {tool.agent_id}")
```

#### Issue 2: Cross-Organization Access Denied
**Symptoms**: "Permission denied" when calling federated tools
**Causes**:
- Missing federation setup
- Incorrect security policies
- Network firewall issues

**Solutions**:
```python
# Debug federation access
async def debug_federation_access(agent: FEMAgent, target_org: str):
    # Check federation status
    fed_status = await agent.broker.get_federation_status()
    print(f"Federated brokers: {fed_status}")
    
    # Test cross-org tool discovery
    remote_tools = await agent.discover_tools("*", source=target_org)
    print(f"Remote tools from {target_org}: {len(remote_tools)}")
    
    # Check security policies
    policies = await agent.get_security_policies(target_org)
    print(f"Security policies for {target_org}: {policies}")
```

#### Issue 3: Environment Embodiment Errors
**Symptoms**: Agent can't adapt to environment
**Causes**:
- Missing environment detection
- Invalid body template
- Resource constraints

**Solutions**:
```python
# Debug embodiment process
async def debug_embodiment(agent: FEMAgent):
    # Check environment detection
    env = await agent.detect_environment()
    print(f"Detected environment: {env}")
    
    # Validate body template
    template = agent.get_body_template(env["type"])
    validation = agent.validate_body_template(template)
    print(f"Template validation: {validation}")
    
    # Check resource constraints
    resources = await agent.check_available_resources()
    print(f"Available resources: {resources}")
```

### Performance Optimization

#### Tool Call Caching
```python
class CachedMCPClient:
    def __init__(self):
        self.cache = TTLCache(maxsize=1000, ttl=300)  # 5-minute cache
        
    async def call_tool(self, endpoint: str, tool: str, params: dict) -> dict:
        # Create cache key from tool and parameters
        cache_key = self.create_cache_key(tool, params)
        
        # Check cache first
        if cache_key in self.cache:
            return self.cache[cache_key]
            
        # Call remote tool
        result = await self.mcp_client.call_tool(endpoint, tool, params)
        
        # Cache result if cacheable
        if self.is_cacheable(tool, result):
            self.cache[cache_key] = result
            
        return result
```

#### Connection Pooling
```python
class FederatedMCPClient:
    def __init__(self):
        self.connection_pools = {}
        
    async def get_connection(self, endpoint: str) -> MCPConnection:
        """Get pooled connection to MCP endpoint"""
        if endpoint not in self.connection_pools:
            self.connection_pools[endpoint] = ConnectionPool(
                endpoint=endpoint,
                min_size=1,
                max_size=10,
                idle_timeout=300
            )
            
        return await self.connection_pools[endpoint].acquire()
```

This completes the comprehensive MCP Integration guide. It provides practical, actionable guidance for migrating existing MCP tools to FEM federation and building new federated agents that leverage the full power of the combined MCP+FEP architecture.