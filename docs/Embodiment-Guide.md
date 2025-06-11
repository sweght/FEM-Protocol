# Embodiment Guide: Adaptive Agents Across Environments

This guide explores the powerful concept of embodiment in FEM—how agents can adapt their capabilities (MCP tools) to their deployment environment, creating truly portable and context-aware AI systems.

## Table of Contents
- [Understanding Embodiment](#understanding-embodiment)
- [Embodiment Patterns](#embodiment-patterns)
- [Environment Detection](#environment-detection)
- [Body Definition Templates](#body-definition-templates)
- [Multi-Environment Deployment](#multi-environment-deployment)
- [Real-World Examples](#real-world-examples)
- [Advanced Patterns](#advanced-patterns)
- [Best Practices](#best-practices)

## Understanding Embodiment

### The Core Concept

Embodiment in FEM means that an agent (mind) adapts its capabilities (body) based on its operational context (environment). This enables a single agent codebase to work optimally across radically different deployment scenarios.

```
Mind (Agent Logic) + Environment (Context) → Body (MCP Tools)
```

### Why Embodiment Matters

Traditional approaches require separate implementations for each environment:
- Desktop app with file system tools
- Web app with DOM manipulation tools  
- Cloud service with API integration tools
- Mobile app with sensor access tools

With FEM embodiment, **one agent mind can work in all environments** by adapting its tool collection.

### Embodiment vs. Configuration

**Traditional Configuration**:
```python
# Different configs for different environments
if environment == "production":
    db_host = "prod-db.example.com"
else:
    db_host = "localhost"
```

**FEM Embodiment**:
```python
# Different capabilities for different environments
if environment == "cloud":
    self.register_tool("db.query", cloud_db_handler)
elif environment == "local":
    self.register_tool("db.query", sqlite_handler)
elif environment == "mobile":
    self.register_tool("db.query", local_storage_handler)
```

The difference: configuration changes parameters, embodiment changes **what the agent can do**.

## Embodiment Patterns

### Pattern 1: Static Embodiment

**Definition**: Agent's body is determined at deployment time and remains fixed.

**Use Cases**: 
- Production services with known environments
- Specialized agents with single-purpose deployment
- Security-sensitive environments requiring locked capabilities

**Implementation**:
```python
class StaticWebAgent(FEMAgent):
    def __init__(self, environment_type: str):
        super().__init__("web-agent")
        self.environment_type = environment_type
        self.mcp_server = MCPServer()
        
        # Body determined at initialization
        if environment_type == "browser":
            self.embody_browser()
        elif environment_type == "node":
            self.embody_node()
        else:
            raise ValueError(f"Unsupported environment: {environment_type}")
            
    def embody_browser(self):
        """Browser-specific tools only"""
        @self.mcp_server.tool("dom.query")
        async def query_dom(selector: str) -> list:
            return document.querySelectorAll(selector)
            
        @self.mcp_server.tool("storage.local.set")
        async def set_local_storage(key: str, value: str) -> bool:
            localStorage.setItem(key, value)
            return True
            
    def embody_node(self):
        """Node.js-specific tools only"""
        @self.mcp_server.tool("fs.read")
        async def read_file(path: str) -> str:
            with open(path, 'r') as f:
                return f.read()
                
        @self.mcp_server.tool("process.spawn")
        async def spawn_process(command: str, args: list) -> dict:
            result = subprocess.run([command] + args, capture_output=True)
            return {"stdout": result.stdout, "stderr": result.stderr}

# Deployment
browser_agent = StaticWebAgent("browser")  # Fixed browser capabilities
node_agent = StaticWebAgent("node")        # Fixed Node.js capabilities
```

### Pattern 2: Dynamic Embodiment

**Definition**: Agent adapts its body at runtime based on environment detection.

**Use Cases**:
- Multi-cloud deployments
- Edge-to-cloud migration
- Development-to-production promotion
- Hybrid applications

**Implementation**:
```python
class DynamicDataAgent(FEMAgent):
    def __init__(self):
        super().__init__("adaptive-data-agent")
        self.mcp_server = MCPServer()
        self.current_body = None
        
    async def start(self):
        """Auto-detect environment and embody accordingly"""
        environment = await self.detect_environment()
        await self.embody_for_environment(environment)
        await self.connect_to_broker()
        
    async def detect_environment(self) -> dict:
        """Detect current operational environment"""
        env = {"type": "unknown", "features": [], "constraints": {}}
        
        # Cloud provider detection
        if os.getenv("AWS_REGION"):
            env["type"] = "aws"
            env["features"].extend(["s3", "dynamodb", "lambda"])
        elif os.getenv("GOOGLE_CLOUD_PROJECT"):
            env["type"] "gcp"
            env["features"].extend(["gcs", "bigquery", "functions"])
        elif os.getenv("AZURE_SUBSCRIPTION_ID"):
            env["type"] = "azure"
            env["features"].extend(["blob", "cosmos", "functions"])
            
        # Container detection
        elif os.path.exists("/.dockerenv"):
            env["type"] = "container"
            env["features"].extend(["networked", "isolated"])
            
        # Local development
        else:
            env["type"] = "local"
            env["features"].extend(["filesystem", "shell", "network"])
            
        # Resource constraints
        env["constraints"]["memory"] = self.get_available_memory()
        env["constraints"]["cpu_cores"] = os.cpu_count()
        env["constraints"]["storage"] = self.get_available_storage()
        
        return env
        
    async def embody_for_environment(self, env: dict):
        """Adapt tools based on environment"""
        if env["type"] == "aws":
            await self.embody_aws(env)
        elif env["type"] == "gcp":
            await self.embody_gcp(env)
        elif env["type"] == "local":
            await self.embody_local(env)
        else:
            await self.embody_generic(env)
            
    async def embody_aws(self, env: dict):
        """AWS-specific tool implementations"""
        @self.mcp_server.tool("data.read")
        async def read_from_s3(bucket: str, key: str) -> str:
            return await s3_client.get_object(bucket, key)
            
        @self.mcp_server.tool("data.write")
        async def write_to_s3(bucket: str, key: str, data: str) -> bool:
            await s3_client.put_object(bucket, key, data)
            return True
            
        @self.mcp_server.tool("data.query")
        async def query_dynamodb(table: str, query: dict) -> list:
            return await dynamodb_client.query(table, query)
            
        self.current_body = "aws"
        
    async def embody_local(self, env: dict):
        """Local filesystem implementations"""
        @self.mcp_server.tool("data.read")
        async def read_local_file(path: str) -> str:
            with open(path, 'r') as f:
                return f.read()
                
        @self.mcp_server.tool("data.write")
        async def write_local_file(path: str, data: str) -> bool:
            with open(path, 'w') as f:
                f.write(data)
            return True
            
        @self.mcp_server.tool("data.query")
        async def query_sqlite(db_path: str, query: str) -> list:
            conn = sqlite3.connect(db_path)
            return conn.execute(query).fetchall()
            
        self.current_body = "local"
```

### Pattern 3: Progressive Embodiment

**Definition**: Agent gradually gains capabilities as it proves trustworthiness or meets conditions.

**Use Cases**:
- Security-conscious environments
- Capability escalation based on performance
- Trial/premium service tiers
- Trust-building systems

**Implementation**:
```python
class ProgressiveAgent(FEMAgent):
    def __init__(self):
        super().__init__("progressive-agent")
        self.trust_level = "untrusted"
        self.performance_score = 0
        self.mcp_server = MCPServer()
        
    async def start(self):
        """Start with minimal capabilities"""
        await self.embody_minimal()
        await self.connect_to_broker()
        
        # Monitor for progression opportunities
        self.start_progression_monitor()
        
    async def embody_minimal(self):
        """Minimal capabilities for untrusted agents"""
        @self.mcp_server.tool("data.read")
        async def read_public_data(path: str) -> str:
            # Only read from public directory
            if not path.startswith("/public/"):
                raise PermissionError("Access denied")
            return await safe_read_file(path)
            
        self.trust_level = "untrusted"
        self.announce_capabilities(["data.read.public"])
        
    async def progress_to_trusted(self):
        """Gain additional capabilities after proving trustworthiness"""
        @self.mcp_server.tool("data.write")
        async def write_data(path: str, data: str) -> bool:
            # Can now write to user directories
            if not self.validate_write_permissions(path):
                raise PermissionError("Write access denied")
            return await safe_write_file(path, data)
            
        @self.mcp_server.tool("process.spawn")
        async def spawn_limited_process(command: str) -> dict:
            # Can spawn processes from allowlist
            if command not in TRUSTED_COMMANDS:
                raise PermissionError("Command not allowed")
            return await safe_spawn(command)
            
        self.trust_level = "trusted"
        self.announce_capabilities(["data.read.public", "data.write.user", "process.spawn.limited"])
        
    async def progress_to_admin(self):
        """Gain administrative capabilities"""
        @self.mcp_server.tool("system.manage")
        async def manage_system(action: str, params: dict) -> dict:
            # Full system management capabilities
            return await admin_action(action, params)
            
        self.trust_level = "admin"
        self.announce_capabilities(["data.*", "process.*", "system.*"])
        
    def start_progression_monitor(self):
        """Monitor performance and trigger progressions"""
        async def monitor():
            while True:
                await asyncio.sleep(300)  # Check every 5 minutes
                
                # Check trust criteria
                if self.trust_level == "untrusted" and self.meets_trusted_criteria():
                    await self.progress_to_trusted()
                elif self.trust_level == "trusted" and self.meets_admin_criteria():
                    await self.progress_to_admin()
                    
        asyncio.create_task(monitor())
```

### Pattern 4: Multi-Body Embodiment

**Definition**: Single agent maintains multiple bodies simultaneously across different environments.

**Use Cases**:
- Cross-environment synchronization
- Distributed data processing
- Multi-region deployments
- Hybrid cloud architectures

**Implementation**:
```python
class MultiBodyAgent(FEMAgent):
    def __init__(self):
        super().__init__("multi-body-agent")
        self.bodies = {}  # Environment -> Body mapping
        self.coordination_queue = asyncio.Queue()
        
    async def add_body(self, environment: str, config: dict):
        """Add a new body in specified environment"""
        body = {
            "environment": environment,
            "mcp_server": MCPServer(f"body-{environment}"),
            "capabilities": [],
            "config": config
        }
        
        # Register environment-specific tools
        await self.setup_body_tools(body, environment)
        
        # Connect to environment-specific broker
        body["broker_connection"] = await self.connect_to_broker(config["broker_url"])
        
        self.bodies[environment] = body
        
    async def setup_body_tools(self, body: dict, environment: str):
        """Setup tools specific to each environment"""
        mcp_server = body["mcp_server"]
        
        if environment == "local":
            @mcp_server.tool("file.local.read")
            async def read_local(path: str) -> str:
                return await self.read_local_file(path)
                
        elif environment == "cloud":
            @mcp_server.tool("file.cloud.read")
            async def read_cloud(bucket: str, key: str) -> str:
                return await self.read_cloud_file(bucket, key)
                
        elif environment == "mobile":
            @mcp_server.tool("sensor.read")
            async def read_sensor(sensor_type: str) -> dict:
                return await self.read_mobile_sensor(sensor_type)
                
        # Add coordination tool to all bodies
        @mcp_server.tool("body.coordinate")
        async def coordinate_with_bodies(task: dict) -> dict:
            return await self.coordinate_task(task)
            
    async def coordinate_task(self, task: dict) -> dict:
        """Coordinate task execution across multiple bodies"""
        results = {}
        
        # Determine which bodies should handle which parts
        task_assignment = await self.plan_task_distribution(task)
        
        # Execute on each relevant body
        for env, subtask in task_assignment.items():
            if env in self.bodies:
                result = await self.execute_on_body(env, subtask)
                results[env] = result
                
        # Aggregate and return combined results
        return await self.aggregate_results(results)
        
    async def execute_on_body(self, environment: str, task: dict) -> dict:
        """Execute task on specific body"""
        body = self.bodies[environment]
        
        # Use the body's MCP tools to execute task
        if task["type"] == "file_operation":
            if environment == "local":
                return await body["mcp_server"].call_tool("file.local.read", task["params"])
            elif environment == "cloud":
                return await body["mcp_server"].call_tool("file.cloud.read", task["params"])
                
        elif task["type"] == "data_processing":
            # Process data using environment-appropriate tools
            return await self.process_data_on_body(body, task["data"])

# Usage example
agent = MultiBodyAgent()
await agent.add_body("local", {"broker_url": "https://local-broker:8443"})
await agent.add_body("cloud", {"broker_url": "https://cloud-broker:8443"})
await agent.add_body("mobile", {"broker_url": "https://mobile-broker:8443"})

# Now the agent can coordinate tasks across all three environments
task = {
    "type": "data_sync",
    "source": "local",
    "destinations": ["cloud", "mobile"],
    "data_path": "/home/user/important_data.json"
}
result = await agent.coordinate_task(task)
```

## Environment Detection

### Comprehensive Environment Detection

```python
class EnvironmentDetector:
    """Comprehensive environment detection and classification"""
    
    @staticmethod
    async def detect() -> dict:
        """Detect and classify current environment"""
        env = {
            "type": "unknown",
            "subtype": None,
            "cloud_provider": None,
            "container_runtime": None,
            "os_family": platform.system().lower(),
            "architecture": platform.machine(),
            "capabilities": [],
            "constraints": {},
            "services": [],
            "network": {},
            "security_context": {}
        }
        
        # Cloud provider detection
        env["cloud_provider"] = await EnvironmentDetector.detect_cloud_provider()
        if env["cloud_provider"]:
            env["type"] = "cloud"
            env["subtype"] = env["cloud_provider"]
            
        # Container detection
        env["container_runtime"] = await EnvironmentDetector.detect_container()
        if env["container_runtime"]:
            if env["type"] == "unknown":
                env["type"] = "container"
            env["subtype"] = env["container_runtime"]
            
        # Edge/IoT detection
        if await EnvironmentDetector.is_edge_device():
            env["type"] = "edge"
            env["subtype"] = await EnvironmentDetector.detect_edge_type()
            
        # Mobile detection
        if await EnvironmentDetector.is_mobile():
            env["type"] = "mobile"
            env["subtype"] = await EnvironmentDetector.detect_mobile_platform()
            
        # Browser detection
        if await EnvironmentDetector.is_browser():
            env["type"] = "browser"
            env["subtype"] = await EnvironmentDetector.detect_browser()
            
        # Default to local if nothing else detected
        if env["type"] == "unknown":
            env["type"] = "local"
            env["subtype"] = "development"
            
        # Detect available capabilities
        env["capabilities"] = await EnvironmentDetector.detect_capabilities(env)
        
        # Detect resource constraints
        env["constraints"] = await EnvironmentDetector.detect_constraints(env)
        
        # Detect available services
        env["services"] = await EnvironmentDetector.detect_services(env)
        
        # Detect network configuration
        env["network"] = await EnvironmentDetector.detect_network(env)
        
        # Detect security context
        env["security_context"] = await EnvironmentDetector.detect_security_context(env)
        
        return env
        
    @staticmethod
    async def detect_cloud_provider() -> str:
        """Detect cloud provider"""
        # AWS
        if os.getenv("AWS_REGION") or os.getenv("AWS_DEFAULT_REGION"):
            return "aws"
            
        # GCP
        if os.getenv("GOOGLE_CLOUD_PROJECT") or os.getenv("GCLOUD_PROJECT"):
            return "gcp"
            
        # Azure
        if os.getenv("AZURE_SUBSCRIPTION_ID") or os.getenv("AZURE_CLIENT_ID"):
            return "azure"
            
        # Try metadata endpoints
        try:
            # AWS metadata
            async with aiohttp.ClientSession(timeout=aiohttp.ClientTimeout(total=2)) as session:
                async with session.get("http://169.254.169.254/latest/meta-data/instance-id") as resp:
                    if resp.status == 200:
                        return "aws"
        except:
            pass
            
        try:
            # GCP metadata
            async with aiohttp.ClientSession(timeout=aiohttp.ClientTimeout(total=2)) as session:
                async with session.get("http://metadata.google.internal/computeMetadata/v1/instance/id",
                                     headers={"Metadata-Flavor": "Google"}) as resp:
                    if resp.status == 200:
                        return "gcp"
        except:
            pass
            
        return None
        
    @staticmethod
    async def detect_container() -> str:
        """Detect container runtime"""
        # Docker
        if os.path.exists("/.dockerenv"):
            return "docker"
            
        # Kubernetes
        if os.getenv("KUBERNETES_SERVICE_HOST"):
            return "kubernetes"
            
        # Check cgroup
        try:
            with open("/proc/1/cgroup", "r") as f:
                cgroup_content = f.read()
                if "docker" in cgroup_content:
                    return "docker"
                elif "kubepods" in cgroup_content:
                    return "kubernetes"
        except:
            pass
            
        return None
        
    @staticmethod
    async def detect_capabilities(env: dict) -> list:
        """Detect available capabilities based on environment"""
        capabilities = []
        
        if env["type"] == "local":
            capabilities.extend([
                "filesystem.full", "network.unrestricted", "shell.access",
                "process.spawn", "hardware.direct"
            ])
            
        elif env["type"] == "cloud":
            if env["subtype"] == "aws":
                capabilities.extend([
                    "s3.access", "dynamodb.access", "lambda.invoke",
                    "ses.send", "sqs.access", "sns.publish"
                ])
            elif env["subtype"] == "gcp":
                capabilities.extend([
                    "gcs.access", "bigquery.query", "functions.invoke",
                    "pubsub.publish", "firestore.access"
                ])
                
        elif env["type"] == "container":
            capabilities.extend([
                "filesystem.limited", "network.restricted",
                "process.isolated"
            ])
            
        elif env["type"] == "browser":
            capabilities.extend([
                "dom.access", "storage.local", "storage.session",
                "websocket.connect", "fetch.api", "webrtc.connect"
            ])
            
        elif env["type"] == "mobile":
            capabilities.extend([
                "sensors.access", "camera.access", "location.access",
                "storage.secure", "push.notifications"
            ])
            
        return capabilities
```

## Body Definition Templates

### Template-Based Embodiment

```yaml
# body_templates.yaml
templates:
  local-development:
    description: "Local development environment with full access"
    environment_types: ["local", "development"]
    mcp_tools:
      - name: "file.read"
        implementation: "filesystem"
        config:
          allowed_paths: ["/tmp", "/home/user/workspace", "/var/log"]
          max_file_size: "100MB"
      - name: "file.write"
        implementation: "filesystem"
        config:
          allowed_paths: ["/tmp", "/home/user/workspace"]
          backup_on_overwrite: true
      - name: "shell.execute"
        implementation: "subprocess"
        config:
          allowed_commands: ["ls", "cat", "grep", "find", "git"]
          timeout: 30
      - name: "process.monitor"
        implementation: "psutil"
    security_policy:
      trust_level: "high"
      sandbox: false
      network_access: "unrestricted"
    resource_limits:
      max_memory: "2GB"
      max_cpu_percent: 80
      max_file_handles: 1000
      
  cloud-production:
    description: "Cloud production environment with service integrations"
    environment_types: ["aws", "gcp", "azure"]
    mcp_tools:
      - name: "file.read"
        implementation: "cloud_storage"
        config:
          service_mapping:
            aws: "s3"
            gcp: "gcs"
            azure: "blob"
      - name: "file.write"
        implementation: "cloud_storage"
        config:
          versioning: true
          encryption: "AES256"
      - name: "database.query"
        implementation: "cloud_database"
        config:
          service_mapping:
            aws: "dynamodb"
            gcp: "firestore"
            azure: "cosmos"
      - name: "message.send"
        implementation: "cloud_messaging"
    security_policy:
      trust_level: "medium"
      sandbox: true
      network_access: "service_endpoints_only"
    resource_limits:
      max_memory: "1GB"
      max_cpu_percent: 50
      max_connections: 100
      
  edge-device:
    description: "Resource-constrained edge device"
    environment_types: ["edge", "iot", "embedded"]
    mcp_tools:
      - name: "sensor.read"
        implementation: "gpio"
        config:
          polling_interval: 1000
          cache_duration: 5000
      - name: "data.compress"
        implementation: "lightweight_compression"
        config:
          algorithm: "lz4"  # Fast compression for limited CPU
      - name: "cache.store"
        implementation: "memory_cache"
        config:
          max_size: "50MB"
          eviction_policy: "LRU"
      - name: "network.transmit"
        implementation: "batch_uploader"
        config:
          batch_size: 100
          retry_attempts: 3
    security_policy:
      trust_level: "low"
      sandbox: true
      network_access: "minimal"
    resource_limits:
      max_memory: "128MB"
      max_cpu_percent: 30
      max_storage: "1GB"
      
  browser-extension:
    description: "Browser extension environment"
    environment_types: ["browser", "extension"]
    mcp_tools:
      - name: "dom.query"
        implementation: "browser_api"
        config:
          allowed_origins: ["https://*.example.com"]
      - name: "storage.local.set"
        implementation: "browser_storage"
        config:
          quota_limit: "10MB"
      - name: "tab.manage"
        implementation: "browser_tabs"
        config:
          max_tabs: 10
      - name: "network.fetch"
        implementation: "browser_fetch"
        config:
          cors_enabled: true
    security_policy:
      trust_level: "low"
      sandbox: true
      network_access: "cors_restricted"
    resource_limits:
      max_memory: "100MB"
      max_storage: "50MB"
```

### Template Loading and Application

```python
class BodyTemplateManager:
    def __init__(self, template_file: str = "body_templates.yaml"):
        self.templates = self.load_templates(template_file)
        
    def load_templates(self, file_path: str) -> dict:
        """Load body templates from YAML file"""
        with open(file_path, 'r') as f:
            return yaml.safe_load(f)
            
    def get_template_for_environment(self, environment: dict) -> dict:
        """Select best template for detected environment"""
        env_type = environment["type"]
        env_subtype = environment.get("subtype")
        
        # Try exact match first
        template_key = f"{env_type}-{env_subtype}" if env_subtype else env_type
        if template_key in self.templates["templates"]:
            return self.templates["templates"][template_key]
            
        # Try environment type match
        for template_name, template in self.templates["templates"].items():
            if env_type in template.get("environment_types", []):
                return template
                
        # Default template
        return self.templates["templates"]["local-development"]
        
    async def apply_template(self, agent: FEMAgent, template: dict, environment: dict):
        """Apply body template to agent"""
        # Register MCP tools based on template
        for tool_def in template["mcp_tools"]:
            await self.register_tool_from_template(agent, tool_def, environment)
            
        # Apply security policy
        await self.apply_security_policy(agent, template["security_policy"])
        
        # Set resource limits
        await self.apply_resource_limits(agent, template["resource_limits"])
        
    async def register_tool_from_template(self, agent: FEMAgent, tool_def: dict, environment: dict):
        """Register a single tool based on template definition"""
        tool_name = tool_def["name"]
        implementation = tool_def["implementation"]
        config = tool_def.get("config", {})
        
        # Create tool handler based on implementation type
        handler = await self.create_tool_handler(implementation, config, environment)
        
        # Register with agent's MCP server
        agent.mcp_server.register_tool(tool_name, handler)

# Usage
template_manager = BodyTemplateManager()
environment = await EnvironmentDetector.detect()
template = template_manager.get_template_for_environment(environment)
await template_manager.apply_template(agent, template, environment)
```

## Multi-Environment Deployment

### Deployment Orchestration

```python
class MultiEnvironmentDeployment:
    """Orchestrate agent deployment across multiple environments"""
    
    def __init__(self, agent_definition: dict):
        self.agent_definition = agent_definition
        self.deployments = {}
        
    async def deploy_to_environments(self, environments: list):
        """Deploy agent to multiple environments simultaneously"""
        deployment_tasks = []
        
        for env_config in environments:
            task = asyncio.create_task(
                self.deploy_to_environment(env_config)
            )
            deployment_tasks.append((env_config["name"], task))
            
        # Wait for all deployments
        for env_name, task in deployment_tasks:
            try:
                deployment = await task
                self.deployments[env_name] = deployment
                print(f"✅ Deployed to {env_name}")
            except Exception as e:
                print(f"❌ Failed to deploy to {env_name}: {e}")
                
    async def deploy_to_environment(self, env_config: dict) -> dict:
        """Deploy agent to a single environment"""
        env_name = env_config["name"]
        
        # Create agent instance for this environment
        agent = FEMAgent(f"{self.agent_definition['name']}-{env_name}")
        
        # Detect or use provided environment specification
        if "environment" in env_config:
            environment = env_config["environment"]
        else:
            # Connect to environment and detect
            environment = await self.detect_remote_environment(env_config["connection"])
            
        # Apply appropriate embodiment
        template_manager = BodyTemplateManager()
        template = template_manager.get_template_for_environment(environment)
        await template_manager.apply_template(agent, template, environment)
        
        # Connect to environment-specific broker
        await agent.connect(env_config["broker_url"])
        
        return {
            "agent": agent,
            "environment": environment,
            "template": template,
            "status": "deployed"
        }
        
    async def coordinate_cross_environment_task(self, task: dict) -> dict:
        """Execute task that spans multiple environments"""
        results = {}
        
        # Determine which environments should handle which parts
        task_plan = await self.plan_cross_environment_task(task)
        
        # Execute on each environment
        for env_name, subtask in task_plan.items():
            if env_name in self.deployments:
                agent = self.deployments[env_name]["agent"]
                result = await self.execute_subtask(agent, subtask)
                results[env_name] = result
                
        return await self.aggregate_cross_environment_results(results)

# Example deployment configuration
deployment_config = {
    "agent_definition": {
        "name": "data-processor",
        "capabilities": ["data.process", "data.transform", "data.analyze"]
    },
    "environments": [
        {
            "name": "local-dev",
            "broker_url": "https://dev-broker.local:8443",
            "environment": {"type": "local", "subtype": "development"}
        },
        {
            "name": "aws-prod",
            "broker_url": "https://prod-broker.aws.example.com:8443",
            "environment": {"type": "cloud", "subtype": "aws"}
        },
        {
            "name": "edge-sensors",
            "broker_url": "https://edge-broker.iot.example.com:8443",
            "environment": {"type": "edge", "subtype": "iot"}
        }
    ]
}

# Deploy
deployment = MultiEnvironmentDeployment(deployment_config["agent_definition"])
await deployment.deploy_to_environments(deployment_config["environments"])
```

## Real-World Examples

### Example 1: Universal Content Management Agent

**Scenario**: Content management system that works across web, mobile, and desktop platforms.

```python
class UniversalContentAgent(FEMAgent):
    def __init__(self):
        super().__init__("universal-content")
        self.mcp_server = MCPServer()
        
    async def embody_for_platform(self, platform: str):
        """Embody differently for each platform"""
        
        if platform == "web":
            await self.embody_web()
        elif platform == "mobile":
            await self.embody_mobile()
        elif platform == "desktop":
            await self.embody_desktop()
            
    async def embody_web(self):
        """Web platform: DOM manipulation and browser APIs"""
        @self.mcp_server.tool("content.render")
        async def render_content(content: dict) -> str:
            # Render content as HTML
            return self.generate_html(content)
            
        @self.mcp_server.tool("content.interact")
        async def handle_interaction(event: dict) -> dict:
            # Handle browser events
            return self.process_dom_event(event)
            
        @self.mcp_server.tool("content.store")
        async def store_content(content: dict) -> bool:
            # Store in browser local storage
            return await self.store_in_browser(content)
            
    async def embody_mobile(self):
        """Mobile platform: touch interfaces and sensors"""
        @self.mcp_server.tool("content.render")
        async def render_content(content: dict) -> dict:
            # Render content for mobile UI
            return self.generate_mobile_ui(content)
            
        @self.mcp_server.tool("content.interact")
        async def handle_interaction(gesture: dict) -> dict:
            # Handle touch gestures
            return self.process_touch_gesture(gesture)
            
        @self.mcp_server.tool("content.capture")
        async def capture_media() -> dict:
            # Use mobile camera/microphone
            return await self.capture_from_device()
            
    async def embody_desktop(self):
        """Desktop platform: file system and native APIs"""
        @self.mcp_server.tool("content.render")
        async def render_content(content: dict) -> str:
            # Render content as native desktop UI
            return self.generate_native_ui(content)
            
        @self.mcp_server.tool("content.interact")
        async def handle_interaction(action: dict) -> dict:
            # Handle desktop UI events
            return self.process_desktop_action(action)
            
        @self.mcp_server.tool("content.export")
        async def export_content(content: dict, format: str) -> str:
            # Export to various desktop file formats
            return await self.export_to_file(content, format)
```

### Example 2: Adaptive Data Processing Pipeline

**Scenario**: Data processing that adapts to available compute resources and data sources.

```python
class AdaptiveDataPipeline(FEMAgent):
    def __init__(self):
        super().__init__("adaptive-pipeline")
        self.mcp_server = MCPServer()
        
    async def embody_for_resources(self, resources: dict):
        """Adapt processing based on available resources"""
        
        if resources["type"] == "high-performance":
            await self.embody_hpc()
        elif resources["type"] == "cloud-elastic":
            await self.embody_cloud()
        elif resources["type"] == "edge-constrained":
            await self.embody_edge()
            
    async def embody_hpc(self):
        """High-performance computing: parallel processing"""
        @self.mcp_server.tool("data.process")
        async def process_data(data: list) -> list:
            # Use multiprocessing for CPU-intensive work
            with multiprocessing.Pool() as pool:
                return pool.map(self.intensive_processing, data)
                
        @self.mcp_server.tool("data.analyze")
        async def analyze_data(data: list) -> dict:
            # Use GPU acceleration if available
            return await self.gpu_accelerated_analysis(data)
            
    async def embody_cloud(self):
        """Cloud elastic: auto-scaling and managed services"""
        @self.mcp_server.tool("data.process")
        async def process_data(data: list) -> list:
            # Use cloud functions for auto-scaling
            return await self.invoke_cloud_functions(data)
            
        @self.mcp_server.tool("data.analyze")
        async def analyze_data(data: list) -> dict:
            # Use managed ML services
            return await self.cloud_ml_analysis(data)
            
    async def embody_edge(self):
        """Edge constrained: lightweight processing"""
        @self.mcp_server.tool("data.process")
        async def process_data(data: list) -> list:
            # Process in small batches to manage memory
            results = []
            for batch in self.create_batches(data, size=100):
                batch_result = await self.lightweight_processing(batch)
                results.extend(batch_result)
            return results
            
        @self.mcp_server.tool("data.analyze")
        async def analyze_data(data: list) -> dict:
            # Use simplified algorithms for edge
            return await self.edge_optimized_analysis(data)
```

## Advanced Patterns

### Context-Aware Embodiment

```python
class ContextAwareAgent(FEMAgent):
    def __init__(self):
        super().__init__("context-aware")
        self.context_monitor = ContextMonitor()
        self.embodiment_history = []
        
    async def start(self):
        """Start with context monitoring"""
        await self.initial_embodiment()
        self.start_context_monitoring()
        
    async def initial_embodiment(self):
        """Initial embodiment based on startup context"""
        context = await self.context_monitor.get_current_context()
        await self.embody_for_context(context)
        
    def start_context_monitoring(self):
        """Monitor context changes and re-embody as needed"""
        async def monitor():
            while True:
                await asyncio.sleep(30)  # Check every 30 seconds
                
                current_context = await self.context_monitor.get_current_context()
                
                if self.should_re_embody(current_context):
                    await self.re_embody(current_context)
                    
        asyncio.create_task(monitor())
        
    def should_re_embody(self, new_context: dict) -> bool:
        """Determine if re-embodiment is needed"""
        if not self.embodiment_history:
            return True
            
        last_context = self.embodiment_history[-1]["context"]
        
        # Check for significant context changes
        if new_context["network_quality"] != last_context["network_quality"]:
            return True
        if new_context["battery_level"] < 20 and last_context["battery_level"] >= 20:
            return True
        if new_context["user_activity"] != last_context["user_activity"]:
            return True
            
        return False
        
    async def re_embody(self, context: dict):
        """Perform re-embodiment for new context"""
        # Store current embodiment
        current_embodiment = {
            "timestamp": datetime.now(),
            "context": context,
            "tools": list(self.mcp_server.get_tools().keys())
        }
        self.embodiment_history.append(current_embodiment)
        
        # Clear current tools
        self.mcp_server.clear_tools()
        
        # Embody for new context
        await self.embody_for_context(context)
        
        # Announce capability changes
        await self.announce_capability_update()
        
    async def embody_for_context(self, context: dict):
        """Embody based on comprehensive context"""
        # Network-aware tools
        if context["network_quality"] == "high":
            await self.add_network_intensive_tools()
        else:
            await self.add_offline_capable_tools()
            
        # Battery-aware tools
        if context["battery_level"] < 20:
            await self.add_low_power_tools()
        else:
            await self.add_full_power_tools()
            
        # Activity-aware tools
        if context["user_activity"] == "focused_work":
            await self.add_productivity_tools()
        elif context["user_activity"] == "casual_browsing":
            await self.add_entertainment_tools()
```

## Best Practices

### 1. Embodiment Design Principles

**Principle of Least Surprise**: Agent behavior should be predictable within each environment.

```python
# Good: Consistent interface, environment-appropriate implementation
@mcp_server.tool("file.read")
async def read_file(path: str) -> str:
    if self.environment == "local":
        return filesystem_read(path)
    elif self.environment == "cloud":
        return cloud_storage_read(path)
    # Same tool name, same interface, different implementation

# Bad: Different interfaces for different environments
@mcp_server.tool("file.read_local" if env == "local" else "s3.get_object")
```

**Graceful Degradation**: Maintain core functionality even in constrained environments.

```python
@mcp_server.tool("data.analyze")
async def analyze_data(data: list) -> dict:
    if self.has_capability("gpu.acceleration"):
        return await self.gpu_analysis(data)
    elif self.has_capability("multicore.processing"):
        return await self.parallel_analysis(data)
    else:
        return await self.basic_analysis(data)  # Always available
```

### 2. Performance Optimization

**Lazy Tool Registration**: Only register tools when they're actually available.

```python
async def embody_cloud(self):
    """Only register cloud tools if cloud services are accessible"""
    if await self.check_s3_access():
        @self.mcp_server.tool("storage.read")
        async def read_from_s3(bucket: str, key: str) -> str:
            return await s3_client.get_object(bucket, key)
            
    if await self.check_lambda_access():
        @self.mcp_server.tool("compute.function")
        async def invoke_lambda(function_name: str, payload: dict) -> dict:
            return await lambda_client.invoke(function_name, payload)
```

**Resource-Aware Embodiment**: Consider resource constraints when choosing tools.

```python
def select_processing_tool(self, data_size: int):
    """Select processing approach based on data size and available resources"""
    available_memory = psutil.virtual_memory().available
    
    if data_size > available_memory * 0.8:
        # Use streaming processing for large datasets
        return self.streaming_processor
    elif data_size > 1000000:
        # Use parallel processing for medium datasets
        return self.parallel_processor
    else:
        # Use simple processing for small datasets
        return self.simple_processor
```

### 3. Security Considerations

**Environment-Based Security Policies**: Apply stricter security in untrusted environments.

```python
def apply_security_policy(self, environment: dict):
    """Apply security policy based on environment trust level"""
    trust_level = self.assess_environment_trust(environment)
    
    if trust_level == "untrusted":
        self.enable_strict_sandboxing()
        self.limit_network_access()
        self.enable_input_validation()
    elif trust_level == "semi_trusted":
        self.enable_moderate_sandboxing()
        self.limit_privileged_operations()
    # No additional restrictions for trusted environments
```

**Capability Verification**: Verify agent capabilities match environment.

```python
async def verify_embodiment(self):
    """Verify that claimed capabilities are actually available"""
    for tool_name in self.mcp_server.get_tools():
        try:
            # Test each tool with minimal parameters
            await self.test_tool_availability(tool_name)
        except Exception as e:
            # Remove tools that aren't actually functional
            self.mcp_server.unregister_tool(tool_name)
            self.log_capability_verification_failure(tool_name, e)
```

### 4. Monitoring and Observability

**Embodiment Telemetry**: Track embodiment decisions and performance.

```python
class EmbodimentTelemetry:
    def __init__(self, agent_id: str):
        self.agent_id = agent_id
        self.metrics = MetricsCollector()
        
    def record_embodiment(self, environment: dict, tools: list, duration: float):
        """Record embodiment metrics"""
        self.metrics.record({
            "agent_id": self.agent_id,
            "environment_type": environment["type"],
            "environment_subtype": environment.get("subtype"),
            "tool_count": len(tools),
            "embodiment_duration": duration,
            "timestamp": datetime.utcnow()
        })
        
    def record_tool_performance(self, tool_name: str, execution_time: float, success: bool):
        """Record individual tool performance"""
        self.metrics.record({
            "agent_id": self.agent_id,
            "tool_name": tool_name,
            "execution_time": execution_time,
            "success": success,
            "timestamp": datetime.utcnow()
        })
```

This comprehensive embodiment guide provides the foundation for building truly adaptive agents that can thrive in any environment while maintaining consistent interfaces and reliable functionality.