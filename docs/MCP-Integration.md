# MCP Integration Guide: Hosted Embodiment

This comprehensive guide shows how to integrate the Model Context Protocol (MCP) with FEM Protocol's **Secure Hosted Embodiment** paradigm. Learn how to transform your MCP tools into embodiment-ready "bodies" and build agents that leverage delegated control through hosted embodiment.

## Table of Contents
- [Quick Start: MCP to Hosted Embodiment](#quick-start-mcp-to-hosted-embodiment)
- [Understanding MCP-FEM Integration](#understanding-mcp-fem-integration)
- [Host Body Development](#host-body-development)
- [Guest Embodiment Patterns](#guest-embodiment-patterns)
- [Session-Aware MCP Tools](#session-aware-mcp-tools)
- [Cross-Device Embodiment](#cross-device-embodiment)
- [Security and Permission Management](#security-and-permission-management)
- [Troubleshooting Embodiment](#troubleshooting-embodiment)

## Quick Start: MCP to Hosted Embodiment

Transform any existing MCP server into a host body for guest embodiment:

### Before: Standard MCP Server
```python
# your_existing_tool.py
from mcp import MCPServer

server = MCPServer("Developer Tools", port=8080)

@server.tool("shell.execute")
async def execute_shell(command: str) -> dict:
    result = subprocess.run(command, shell=True, capture_output=True, text=True)
    return {"stdout": result.stdout, "stderr": result.stderr, "returncode": result.returncode}

if __name__ == "__main__":
    server.run()  # Only direct MCP connections
```

### After: Embodiment-Ready Host
```python
# your_host_agent.py
from mcp import MCPServer
from fem_protocol import HostAgent, BodyDefinition, SecurityPolicy

# Create host agent
host = HostAgent("laptop-dev-host")

# Define body with security policies
body_def = BodyDefinition(
    body_id="developer-workstation-v1",
    description="Secure development environment with shell access",
    environment_type="local-development",
    mcp_tools=[
        {
            "name": "shell.execute",
            "description": "Execute shell commands in sandboxed environment",
            "input_schema": {
                "type": "object",
                "properties": {
                    "command": {"type": "string"},
                    "workdir": {"type": "string", "default": "/home/alice/projects"}
                }
            }
        }
    ],
    security_policy=SecurityPolicy(
        allowed_paths=["/home/alice/projects/*"],
        denied_commands=["rm -rf", "sudo", "curl"],
        max_session_duration=3600,  # 1 hour sessions
        max_concurrent_guests=2
    )
)

# Register body and start hosting
async def main():
    await host.define_body(body_def)
    await host.start_mcp_server(port=8080)
    await host.register_with_broker("https://fem-broker:8443")
    print("Host ready for guest embodiment!")

if __name__ == "__main__":
    import asyncio
    asyncio.run(main())
```

**What You've Gained**:
- ✅ **Secure Delegated Control**: Guests can control your environment within defined boundaries
- ✅ **Session Management**: Time-bounded, isolated sessions with audit logging
- ✅ **Permission Enforcement**: Fine-grained control over what guests can access
- ✅ **Cross-Device Access**: Your tools accessible from any device with a guest agent
- ✅ **Automatic Discovery**: Guest agents can find and request embodiment in your host

## Understanding MCP-FEM Integration

### The Hosted Embodiment Architecture
FEM Protocol creates a hosted embodiment layer that uses MCP as the tool interface:

```
┌─────────────────────────────────────────────────────────────┐
│                    Guest Agent                              │
│               (Guest "Mind" with Goals)                     │
└─────────────────────┬───────────────────────────────────────┘
                      │ FEM Protocol (Embodiment Requests)
┌─────────────────────▼───────────────────────────────────────┐
│                   Host Agent                                │
│              (Security & Session Management)                │
└─────────────────────┬───────────────────────────────────────┘
                      │ Session-Scoped MCP Protocol
┌─────────────────────▼───────────────────────────────────────┐
│                   MCP Tool Layer                            │
│              (Your existing MCP tools)                      │
└─────────────────────┬───────────────────────────────────────┘
                      │ Environment Integration
┌─────────────────────▼───────────────────────────────────────┐
│                Host Environment                             │
│        (File system, shell, applications, etc.)            │
└─────────────────────────────────────────────────────────────┘
```

### Key Principles

1. **MCP as Body Interface**: MCP tools define what guests can control in host environments
2. **Session-Scoped Access**: Each embodiment session gets isolated MCP endpoint with specific permissions
3. **Security by Design**: All guest actions validated against host-defined security policies
4. **Environment Awareness**: Bodies adapt MCP tool implementations to deployment contexts

## Host Body Development

### Creating Your First Host Body

**Scenario**: You have MCP tools and want to offer them for guest embodiment.

**Solution**: Wrap your MCP server in a host body with security policies:

```python
# Step 1: Keep your existing MCP implementation
@mcp_server.tool("file.compress")
async def compress_file(file_path: str, format: str = "gzip") -> dict:
    # Validate path against session permissions first
    session = get_current_embodiment_session()
    if not session.validate_path_access(file_path):
        raise PermissionError(f"Path not allowed: {file_path}")
    
    # Your existing implementation
    compressed_path = compress(file_path, format)
    return {"compressed_file": compressed_path, "original_size": orig_size, "compressed_size": new_size}

# Step 2: Create host agent with embodiment support
host_agent = HostAgent("file-host")
body_def = BodyDefinition(
    body_id="file-operations-v1",
    description="File compression and manipulation tools",
    mcp_tools=[{"name": "file.compress", "description": "Compress files safely"}],
    security_policy=SecurityPolicy(
        allowed_paths=["/tmp/uploads/*", "/home/user/documents/*"],
        denied_paths=["/etc/*", "/root/*"],
        max_file_size_mb=100,
        max_session_duration=1800
    )
)

# Step 3: Start hosting embodiment sessions
await host_agent.define_body(body_def)
await host_agent.start_embodiment_hosting()
```

### Host Body Patterns

#### Pattern 1: Single-Purpose Body

Create a body that offers related tools as a cohesive embodiment experience:

```python
class DataProcessingHost(HostAgent):
    def __init__(self):
        super().__init__("data-processing-host")
        self.mcp_server = MCPServer("Data Processing", port=8080)
        self.register_tools()
        
    def register_tools(self):
        @self.mcp_server.tool("data.clean")
        async def clean_data(data: str) -> str:
            session = self.get_current_session()
            session.log_action("data.clean", {"data_size": len(data)})
            return cleaned_data
            
        @self.mcp_server.tool("data.validate") 
        async def validate_data(data: str, schema: dict) -> dict:
            session = self.get_current_session()
            if not session.can_access_schema(schema):
                raise PermissionError("Schema access denied")
            return validation_result
            
        @self.mcp_server.tool("data.transform")
        async def transform_data(data: str, rules: list) -> str:
            session = self.get_current_session()
            session.track_resource_usage("memory", estimate_memory_usage(data))
            return transformed_data
            
    async def start(self):
        # Define comprehensive data processing body
        body_def = BodyDefinition(
            body_id="data-processing-suite-v1",
            description="Complete data processing pipeline with validation",
            mcp_tools=[
                {"name": "data.clean", "description": "Clean and normalize data"},
                {"name": "data.validate", "description": "Validate against schema"},
                {"name": "data.transform", "description": "Apply transformation rules"}
            ],
            security_policy=SecurityPolicy(
                max_data_size_mb=50,
                allowed_schemas=["user-defined", "json-schema"],
                max_session_duration=2400,  # 40 minutes for data processing
                max_concurrent_guests=3
            )
        )
        
        await self.define_body(body_def)
        await self.start_embodiment_hosting()
```

#### Pattern 2: Multi-Environment Body

Create environment-aware bodies that adapt tool behavior:

```python
class AdaptiveFileHost(HostAgent):
    def __init__(self):
        super().__init__("adaptive-file-host")
        self.environment = self.detect_environment()
        
    def detect_environment(self) -> str:
        if os.getenv("AWS_REGION"):
            return "cloud-aws"
        elif os.path.exists("/.dockerenv"):
            return "container"
        else:
            return "local"
            
    async def create_environment_body(self):
        if self.environment == "cloud-aws":
            return await self.create_s3_body()
        elif self.environment == "container":
            return await self.create_container_body()
        else:
            return await self.create_local_body()
            
    async def create_local_body(self) -> BodyDefinition:
        @self.mcp_server.tool("file.read")
        async def read_local_file(path: str) -> str:
            session = self.get_current_session()
            if not session.validate_path_access(path):
                raise PermissionError(f"Path access denied: {path}")
            with open(path, 'r') as f:
                return f.read()
                
        return BodyDefinition(
            body_id="local-file-ops-v1",
            description="Local filesystem operations",
            environment_type="local-development",
            mcp_tools=[{"name": "file.read", "description": "Read local files"}],
            security_policy=SecurityPolicy(
                allowed_paths=["/home/user/*", "/tmp/*"],
                denied_paths=["/etc/*", "/root/*", "/sys/*"]
            )
        )
        
    async def create_s3_body(self) -> BodyDefinition:
        @self.mcp_server.tool("file.read")
        async def read_s3_file(bucket: str, key: str) -> str:
            session = self.get_current_session()
            if not session.can_access_bucket(bucket):
                raise PermissionError(f"Bucket access denied: {bucket}")
            return await s3_client.get_object(bucket, key)
            
        return BodyDefinition(
            body_id="s3-file-ops-v1",
            description="S3 object operations", 
            environment_type="cloud-aws",
            mcp_tools=[{"name": "file.read", "description": "Read S3 objects"}],
            security_policy=SecurityPolicy(
                allowed_buckets=["user-data", "temp-storage"],
                denied_buckets=["system-config", "secrets"]
            )
        )
```

## Guest Embodiment Patterns

### Guest Agent Design

#### Pattern 1: Goal-Oriented Guest
```python
class DevelopmentGuest(GuestAgent):
    """Guest agent that embodies in development environments"""
    
    def __init__(self):
        super().__init__("dev-assistant")
        self.preferences = DiscoveryPreferences(
            required_capabilities=["shell.execute", "file.read", "file.write"],
            preferred_environments=["local-development"],
            max_session_duration=3600  # 1 hour
        )
        
    async def execute_development_task(self, task_description: str):
        # Discover suitable development environment
        bodies = await self.discover_bodies()
        dev_bodies = [b for b in bodies if "development" in b.environment_type]
        
        if not dev_bodies:
            raise RuntimeError("No development environments available")
            
        # Request embodiment
        best_body = self.evaluate_bodies(dev_bodies)[0]
        session = await self.request_embodiment(
            best_body, 
            goals=[task_description],
            duration=1800  # 30 minutes
        )
        
        # Execute development workflow
        try:
            await self.run_development_workflow(session, task_description)
        finally:
            await self.graceful_session_exit(session)
            
    async def run_development_workflow(self, session: EmbodimentSession, task: str):
        # Check project status
        git_status = await self.call_tool(session, "shell.execute", {
            "command": "git status",
            "workdir": "/home/alice/projects/my-app"
        })
        
        # Read configuration
        config = await self.call_tool(session, "file.read", {
            "path": "/home/alice/projects/my-app/package.json"
        })
        
        # Make informed decisions based on project state
        if "clean" in git_status["stdout"]:
            await self.call_tool(session, "shell.execute", {
                "command": "npm run build",
                "workdir": "/home/alice/projects/my-app"
            })
```

#### Pattern 2: Multi-Host Guest
```python
class WorkflowGuest(GuestAgent):
    """Guest agent that orchestrates workflows across multiple hosts"""
    
    def __init__(self):
        super().__init__("workflow-orchestrator")
        self.active_sessions = {}
        
    async def execute_data_pipeline(self, pipeline_config: dict):
        """Execute pipeline across different specialized hosts"""
        results = {}
        
        # Step 1: Data validation (requires data processing host)
        validation_bodies = await self.discover_bodies_with_capability("data.validate")
        val_session = await self.request_embodiment(
            validation_bodies[0],
            goals=["Validate input data"],
            duration=300  # 5 minutes
        )
        
        try:
            validation_result = await self.call_tool(val_session, "data.validate", {
                "data": pipeline_config["input_data"],
                "schema": pipeline_config["validation_schema"]
            })
            results["validation"] = validation_result
            
            # Step 2: File processing (requires file operations host)
            if validation_result["valid"]:
                file_bodies = await self.discover_bodies_with_capability("file.compress")
                file_session = await self.request_embodiment(
                    file_bodies[0],
                    goals=["Process and store results"], 
                    duration=600  # 10 minutes
                )
                
                try:
                    compression_result = await self.call_tool(file_session, "file.compress", {
                        "file_path": "/tmp/validated_data.json",
                        "format": "gzip"
                    })
                    results["compression"] = compression_result
                finally:
                    await self.graceful_session_exit(file_session)
        finally:
            await self.graceful_session_exit(val_session)
            
        return results
```

#### Pattern 3: Context-Aware Guest
```python
class ContextAwareGuest(GuestAgent):
    """Guest that adapts behavior based on host environment and capabilities"""
    
    def __init__(self):
        super().__init__("context-aware-assistant")
        self.execution_context = {}
        
    async def adaptive_file_operation(self, operation_type: str, **params):
        """Adapt file operations based on available host capabilities"""
        
        # Discover available file operations
        file_bodies = await self.discover_bodies_with_capability("file.*")
        
        # Select best body based on environment and capabilities
        best_body = None
        for body in file_bodies:
            if body.environment_type == "cloud-aws" and operation_type == "bulk":
                best_body = body  # Prefer cloud for bulk operations
                break
            elif body.environment_type == "local-development" and operation_type == "edit":
                best_body = body  # Prefer local for editing
                break
        
        if not best_body:
            best_body = file_bodies[0]  # Fallback to first available
            
        # Request embodiment with specific context
        session = await self.request_embodiment(
            best_body,
            goals=[f"Perform {operation_type} file operation"],
            context={"operation_type": operation_type, "params": params}
        )
        
        # Execute appropriate operation based on body capabilities
        try:
            if "file.batch_process" in body.capabilities and operation_type == "bulk":
                result = await self.call_tool(session, "file.batch_process", params)
            else:
                result = await self.call_tool(session, "file.read", params)
                
            # Update context for future operations
            self.execution_context["last_operation"] = {
                "type": operation_type,
                "host": best_body.host_id,
                "success": True
            }
            
            return result
        except Exception as e:
            self.execution_context["last_operation"] = {
                "type": operation_type,
                "host": best_body.host_id,
                "success": False,
                "error": str(e)
            }
            raise
        finally:
            await self.graceful_session_exit(session)
```

## Session-Aware MCP Tools

### Implementing Session Context

MCP tools need to be aware of embodiment sessions for proper security and auditing:

```python
class SessionAwareMCPServer:
    def __init__(self, name: str, port: int):
        self.mcp_server = MCPServer(name, port)
        self.active_sessions = {}
        self.session_context = contextvars.ContextVar('embodiment_session')
        
    def session_aware_tool(self, tool_name: str):
        """Decorator for session-aware MCP tools"""
        def decorator(func):
            @self.mcp_server.tool(tool_name)
            async def wrapper(*args, **kwargs):
                # Get current embodiment session
                session = self.session_context.get(None)
                if not session:
                    raise RuntimeError("Tool called outside embodiment session")
                    
                # Validate session and permissions
                if not session.is_valid():
                    raise PermissionError("Invalid or expired session")
                    
                if not session.can_use_tool(tool_name):
                    raise PermissionError(f"Tool {tool_name} not permitted in this session")
                    
                # Log tool usage
                session.log_tool_call(tool_name, args, kwargs)
                
                # Execute tool with session context
                try:
                    result = await func(session, *args, **kwargs)
                    session.log_tool_result(tool_name, result, success=True)
                    return result
                except Exception as e:
                    session.log_tool_result(tool_name, None, success=False, error=str(e))
                    raise
                    
            return wrapper
        return decorator
        
    def set_session_context(self, session_token: str):
        """Set the current embodiment session context"""
        session = self.active_sessions.get(session_token)
        if session:
            self.session_context.set(session)

# Usage example
server = SessionAwareMCPServer("Secure Dev Tools", 8080)

@server.session_aware_tool("shell.execute")
async def execute_shell(session: EmbodimentSession, command: str, workdir: str = None) -> dict:
    """Execute shell command with session security"""
    
    # Validate command against session policy
    if not session.security_policy.is_command_allowed(command):
        raise PermissionError(f"Command not allowed: {command}")
        
    # Validate working directory
    if workdir and not session.security_policy.is_path_allowed(workdir):
        raise PermissionError(f"Working directory not allowed: {workdir}")
        
    # Execute with resource limits
    result = subprocess.run(
        command,
        shell=True,
        cwd=workdir,
        capture_output=True,
        text=True,
        timeout=session.security_policy.max_command_timeout
    )
    
    return {
        "stdout": result.stdout,
        "stderr": result.stderr,
        "returncode": result.returncode,
        "command": command,
        "workdir": workdir
    }
```

### MCP Tool Permission Validation

```python
class PermissionValidator:
    def __init__(self, security_policy: SecurityPolicy):
        self.policy = security_policy
        
    def validate_file_access(self, session: EmbodimentSession, path: str, operation: str) -> bool:
        """Validate file access permissions"""
        
        # Check allowed paths
        allowed = False
        for allowed_pattern in self.policy.allowed_paths:
            if fnmatch.fnmatch(path, allowed_pattern):
                allowed = True
                break
                
        if not allowed:
            return False
            
        # Check denied paths (takes precedence)
        for denied_pattern in self.policy.denied_paths:
            if fnmatch.fnmatch(path, denied_pattern):
                return False
                
        # Check operation-specific permissions
        if operation == "write" and not session.has_permission("file.write"):
            return False
        elif operation == "read" and not session.has_permission("file.read"):
            return False
            
        return True
        
    def validate_shell_command(self, session: EmbodimentSession, command: str) -> bool:
        """Validate shell command permissions"""
        
        # Extract base command
        base_command = command.split()[0] if command.split() else ""
        
        # Check denied commands first
        for denied_cmd in self.policy.denied_commands:
            if denied_cmd in command:
                return False
                
        # Check allowed commands (if specified)
        if self.policy.allowed_commands:
            return base_command in self.policy.allowed_commands
            
        # Check session permissions
        return session.has_permission("shell.execute")
        
    def validate_resource_usage(self, session: EmbodimentSession, resource_type: str, amount: float) -> bool:
        """Validate resource usage against limits"""
        
        current_usage = session.get_resource_usage(resource_type)
        limit = getattr(self.policy.resource_limits, f"max_{resource_type}", None)
        
        if limit is None:
            return True  # No limit specified
            
        return current_usage + amount <= limit

# Integration with MCP tools
@server.session_aware_tool("file.read")
async def read_file_secure(session: EmbodimentSession, path: str) -> str:
    """Secure file reading with permission validation"""
    
    validator = PermissionValidator(session.security_policy)
    
    # Validate file access
    if not validator.validate_file_access(session, path, "read"):
        raise PermissionError(f"File read access denied: {path}")
        
    # Check file size against resource limits
    file_size = os.path.getsize(path) / (1024 * 1024)  # MB
    if not validator.validate_resource_usage(session, "memory_mb", file_size):
        raise ResourceError(f"File too large, would exceed memory limit: {path}")
        
    # Audit log the access
    session.audit_log.log_file_access("read", path, success=True)
    
    try:
        with open(path, 'r', encoding='utf-8') as f:
            content = f.read()
        return content
    except Exception as e:
        session.audit_log.log_file_access("read", path, success=False, error=str(e))
        raise
```

## Cross-Device Embodiment

### Mobile-to-Desktop Development Workflow

Enable seamless cross-device embodiment for development workflows:

#### Desktop Host Setup
```python
# desktop_dev_host.py
class DesktopDevelopmentHost(HostAgent):
    def __init__(self):
        super().__init__("laptop-dev-host-alice")
        self.setup_development_body()
        
    def setup_development_body(self):
        """Create body optimized for mobile guest access"""
        
        @self.mcp_server.tool("project.status")
        async def get_project_status(session: EmbodimentSession, project_path: str) -> dict:
            """Get comprehensive project status for mobile display"""
            
            # Validate project path
            if not session.security_policy.is_path_allowed(project_path):
                raise PermissionError(f"Project path not allowed: {project_path}")
                
            # Gather project information
            git_status = await self.execute_git_command("status --porcelain", project_path)
            package_info = await self.read_package_info(project_path)
            
            return {
                "project_path": project_path,
                "git_status": git_status,
                "package_info": package_info,
                "last_modified": self.get_last_modified_time(project_path),
                "build_status": await self.check_build_status(project_path)
            }
            
        @self.mcp_server.tool("dev.server")
        async def manage_dev_server(session: EmbodimentSession, action: str, project_path: str, port: int = 3000) -> dict:
            """Start/stop development server for mobile testing"""
            
            if action == "start":
                # Start dev server and return connection info
                process = await self.start_dev_server(project_path, port)
                return {
                    "status": "started",
                    "pid": process.pid,
                    "url": f"http://{self.get_local_ip()}:{port}",
                    "mobile_accessible": True
                }
            elif action == "stop":
                await self.stop_dev_server(project_path)
                return {"status": "stopped"}
                
        # Define mobile-optimized body
        body_def = BodyDefinition(
            body_id="mobile-dev-access-v1",
            description="Mobile-optimized development environment access",
            environment_type="local-development",
            mcp_tools=[
                {"name": "project.status", "description": "Get project status optimized for mobile"},
                {"name": "dev.server", "description": "Manage development servers"},
                {"name": "file.read", "description": "Read project files"},
                {"name": "shell.execute", "description": "Execute safe development commands"}
            ],
            security_policy=SecurityPolicy(
                allowed_paths=["/home/alice/projects/*"],
                allowed_commands=["git", "npm", "yarn", "node", "python"],
                denied_commands=["rm -rf", "sudo", "curl", "wget"],
                max_session_duration=3600,  # 1 hour
                max_concurrent_guests=2,
                mobile_optimized=True
            )
        )
        
        await self.define_body(body_def)
```

#### Mobile Guest Implementation
```python
# mobile_dev_guest.py
class MobileDevelopmentGuest(GuestAgent):
    def __init__(self):
        super().__init__("phone-dev-assistant-bob")
        self.preferences = DiscoveryPreferences(
            required_capabilities=["project.status", "dev.server", "file.read"],
            preferred_environments=["local-development"],
            max_session_duration=1800,  # 30 minutes
            mobile_optimized=True
        )
        
    async def check_project_from_mobile(self, project_name: str) -> dict:
        """Check project status from mobile device"""
        
        # Discover desktop development hosts
        dev_hosts = await self.discover_bodies()
        mobile_compatible = [h for h in dev_hosts if h.security_policy.mobile_optimized]
        
        if not mobile_compatible:
            raise RuntimeError("No mobile-compatible development hosts found")
            
        # Request embodiment with mobile context
        session = await self.request_embodiment(
            mobile_compatible[0],
            goals=[f"Check status of {project_name} project"],
            context={"device_type": "mobile", "project": project_name}
        )
        
        try:
            # Get project status optimized for mobile display
            project_path = f"/home/alice/projects/{project_name}"
            status = await self.call_tool(session, "project.status", {
                "project_path": project_path
            })
            
            # Start dev server if needed for mobile testing
            if status["build_status"]["needs_rebuild"]:
                await self.call_tool(session, "shell.execute", {
                    "command": "npm run build",
                    "workdir": project_path
                })
                
            # Start dev server for mobile access
            server_info = await self.call_tool(session, "dev.server", {
                "action": "start",
                "project_path": project_path,
                "port": 3000
            })
            
            return {
                "project_status": status,
                "dev_server": server_info,
                "mobile_url": server_info["url"],
                "session_id": session.session_token
            }
            
        finally:
            # Keep session alive for continued mobile access
            await self.extend_session_if_needed(session)
            
    async def mobile_friendly_file_browse(self, session: EmbodimentSession, directory: str) -> dict:
        """Browse files with mobile-optimized output"""
        
        files_result = await self.call_tool(session, "shell.execute", {
            "command": f"find {directory} -type f -name '*.py' -o -name '*.js' -o -name '*.json' | head -20",
            "workdir": directory
        })
        
        # Format for mobile display
        files = files_result["stdout"].strip().split("\n") if files_result["stdout"] else []
        
        return {
            "directory": directory,
            "files": [{
                "path": f,
                "name": os.path.basename(f),
                "type": f.split(".")[-1] if "." in f else "unknown"
            } for f in files[:10]],  # Limit for mobile
            "total_found": len(files),
            "truncated": len(files) > 10
        }
```

### Cross-Device Session Management

#### Persistent Session Coordination
```python
class CrossDeviceSessionManager:
    def __init__(self, guest_id: str):
        self.guest_id = guest_id
        self.active_sessions = {}
        self.device_contexts = {}
        
    async def coordinate_cross_device_session(self, primary_device: str, secondary_device: str):
        """Coordinate embodiment session across multiple devices"""
        
        # Get primary session (e.g., from laptop)
        primary_session = self.active_sessions.get(primary_device)
        if not primary_session:
            raise RuntimeError(f"No active session on primary device: {primary_device}")
            
        # Request secondary access (e.g., from phone)
        secondary_session = await self.request_secondary_embodiment(
            primary_session.host_id,
            primary_session.body_id,
            context={
                "primary_session": primary_session.session_token,
                "device_type": secondary_device,
                "access_type": "secondary"
            }
        )
        
        # Synchronize session states
        await self.synchronize_session_contexts(primary_session, secondary_session)
        
        return {
            "primary_session": primary_session.session_token,
            "secondary_session": secondary_session.session_token,
            "synchronized": True
        }
        
    async def handle_device_handoff(self, from_device: str, to_device: str, context: dict):
        """Handle seamless handoff between devices"""
        
        from_session = self.active_sessions[from_device]
        
        # Save current state
        session_state = {
            "current_directory": context.get("current_directory"),
            "active_processes": context.get("active_processes", []),
            "project_context": context.get("project_context"),
            "temporary_files": context.get("temporary_files", [])
        }
        
        # Request new session on target device
        to_session = await self.request_embodiment_with_state(
            from_session.host_id,
            from_session.body_id,
            inherited_state=session_state,
            device_type=to_device
        )
        
        # Transfer context
        await self.transfer_session_context(from_session, to_session)
        
        # Gracefully close original session
        await self.graceful_session_close(from_session, preserve_state=True)
        
        return to_session
        
    async def synchronize_session_contexts(self, primary: EmbodimentSession, secondary: EmbodimentSession):
        """Keep session contexts synchronized across devices"""
        
        # Share read-only context
        shared_context = {
            "project_status": primary.context.get("project_status"),
            "recent_files": primary.context.get("recent_files", []),
            "active_servers": primary.context.get("active_servers", {})
        }
        
        secondary.context.update(shared_context)
        
        # Set up context synchronization
        primary.add_context_sync_target(secondary.session_token)
        secondary.add_context_sync_source(primary.session_token)
```

## Security and Permission Management

### Session-Based Security for MCP Tools

```python
class SecureEmbodimentHost(HostAgent):
    def __init__(self, agent_id: str):
        super().__init__(agent_id)
        self.permission_manager = EmbodimentPermissionManager()
        
    def register_secure_mcp_tool(self, tool_name: str, permission_requirements: list, handler):
        """Register MCP tool with embodiment session security"""
        
        @self.mcp_server.tool(tool_name)
        async def secure_handler(*args, **kwargs):
            # Get current embodiment session
            session = self.get_current_embodiment_session()
            if not session:
                raise RuntimeError(f"Tool {tool_name} called outside embodiment session")
                
            # Validate session
            if not session.is_valid():
                raise PermissionError("Invalid or expired embodiment session")
                
            # Check session permissions
            if not self.permission_manager.validate_tool_access(session, tool_name, permission_requirements):
                raise PermissionError(f"Insufficient permissions for {tool_name}")
                
            # Validate tool parameters against session policy
            validation_result = self.permission_manager.validate_parameters(session, tool_name, args, kwargs)
            if not validation_result.valid:
                raise PermissionError(f"Parameter validation failed: {validation_result.reason}")
                
            # Log access for audit trail
            session.audit_log.log_tool_access(tool_name, args, kwargs)
            
            # Execute tool with session context
            try:
                result = await handler(session, *args, **kwargs)
                session.audit_log.log_tool_result(tool_name, result, success=True)
                return result
            except Exception as e:
                session.audit_log.log_tool_result(tool_name, None, success=False, error=str(e))
                session.increment_violation_count()
                raise

# Example: Secure file operations host
file_host = SecureEmbodimentHost("secure-file-host")

file_host.register_secure_mcp_tool(
    "file.read",
    permission_requirements=["file.read"],
    handler=secure_file_read_handler
)

file_host.register_secure_mcp_tool(
    "file.write", 
    permission_requirements=["file.write"],
    handler=secure_file_write_handler
)

file_host.register_secure_mcp_tool(
    "shell.execute",
    permission_requirements=["shell.execute"],
    handler=secure_shell_execute_handler
)
```

### Embodiment Session Audit Logging

```python
class EmbodimentAuditLogger:
    def __init__(self, host_id: str):
        self.host_id = host_id
        self.log_handler = configure_embodiment_audit_logging()
        
    def log_session_start(self, session: EmbodimentSession):
        """Log embodiment session initiation"""
        audit_entry = {
            "event_type": "session_start",
            "timestamp": datetime.utcnow().isoformat(),
            "host_id": self.host_id,
            "session_token": session.session_token,
            "guest_id": session.guest_id,
            "body_id": session.body_id,
            "granted_permissions": session.permissions,
            "session_duration": session.max_duration,
            "security_policy": session.security_policy.to_dict(),
            "guest_device_info": session.guest_context.get("device_info", {})
        }
        
        self.log_handler.info(json.dumps(audit_entry))
        
    def log_tool_access(self, session: EmbodimentSession, tool: str, args: tuple, kwargs: dict):
        """Log MCP tool access within embodiment session"""
        audit_entry = {
            "event_type": "tool_access",
            "timestamp": datetime.utcnow().isoformat(),
            "host_id": self.host_id,
            "session_token": session.session_token,
            "guest_id": session.guest_id,
            "tool": tool,
            "parameters": self.sanitize_parameters(args, kwargs),
            "session_time_elapsed": (datetime.utcnow() - session.start_time).total_seconds(),
            "violation_count": session.violation_count,
            "resource_usage": session.get_current_resource_usage()
        }
        
        self.log_handler.info(json.dumps(audit_entry))
        
    def sanitize_parameters(self, args: tuple, kwargs: dict) -> dict:
        """Remove sensitive data from parameters for logging"""
        sanitized = {"args": [], "kwargs": {}}
        
        # Sanitize positional args
        for arg in args:
            if isinstance(arg, str) and len(arg) > 500:
                sanitized["args"].append(arg[:100] + "...truncated")
            else:
                sanitized["args"].append(str(arg))
                
        # Sanitize keyword args
        for key, value in kwargs.items():
            if key.lower() in ['password', 'token', 'secret', 'api_key', 'private_key']:
                sanitized["kwargs"][key] = "***REDACTED***"
            elif isinstance(value, str) and len(value) > 500:
                sanitized["kwargs"][key] = value[:100] + "...truncated"
            else:
                sanitized["kwargs"][key] = str(value)
                
        return sanitized
```

## Troubleshooting Embodiment

### Common Embodiment Issues and Solutions

#### Issue 1: Embodiment Request Denied
**Symptoms**: Guest agents receive "embodiment denied" responses
**Causes**: 
- Insufficient trust level
- Security policy violations
- Host resource constraints
- Session limit exceeded

**Solutions**:
```python
# Debug embodiment request
async def debug_embodiment_request(guest: GuestAgent, host_id: str, body_id: str):
    # Check guest trust level
    guest_profile = await guest.get_profile()
    print(f"Guest trust level: {guest_profile.trust_level}")
    
    # Check host body requirements
    body_info = await guest.get_body_info(host_id, body_id)
    print(f"Required trust level: {body_info.security_policy.trust_level_required}")
    
    # Check host availability
    availability = await guest.check_host_availability(host_id)
    print(f"Host availability: {availability}")
    
    # Check security policy compatibility
    compatibility = await guest.check_policy_compatibility(body_info.security_policy)
    print(f"Policy compatibility: {compatibility}")
```

#### Issue 2: Session Permission Violations
**Symptoms**: Tool calls fail with permission errors during active sessions
**Causes**:
- Path access violations
- Command restrictions
- Resource limit exceeded
- Session token expired

**Solutions**:
```python
# Debug session permissions
async def debug_session_permissions(session: EmbodimentSession, tool_name: str, parameters: dict):
    # Check session validity
    print(f"Session valid: {session.is_valid()}")
    print(f"Session expires: {session.expiry_time}")
    print(f"Current time: {datetime.utcnow()}")
    
    # Check tool permissions
    has_permission = session.has_permission(tool_name)
    print(f"Has permission for {tool_name}: {has_permission}")
    
    # Check parameter validation
    validator = PermissionValidator(session.security_policy)
    
    if tool_name.startswith("file."):
        path = parameters.get("path")
        if path:
            file_access = validator.validate_file_access(session, path, tool_name.split(".")[1])
            print(f"File access allowed for {path}: {file_access}")
            
    elif tool_name == "shell.execute":
        command = parameters.get("command")
        if command:
            command_allowed = validator.validate_shell_command(session, command)
            print(f"Command allowed '{command}': {command_allowed}")
    
    # Check resource usage
    resource_usage = session.get_current_resource_usage()
    print(f"Current resource usage: {resource_usage}")
    print(f"Resource limits: {session.security_policy.resource_limits}")
```

#### Issue 3: Cross-Device Session Sync Issues
**Symptoms**: Session state not synchronized between devices
**Causes**:
- Network connectivity issues
- Session token mismatch
- Context synchronization failures
- Device-specific limitations

**Solutions**:
```python
# Debug cross-device session sync
async def debug_cross_device_sync(session_manager: CrossDeviceSessionManager, device1: str, device2: str):
    # Check active sessions on both devices
    device1_sessions = session_manager.get_active_sessions(device1)
    device2_sessions = session_manager.get_active_sessions(device2)
    
    print(f"Active sessions on {device1}: {len(device1_sessions)}")
    print(f"Active sessions on {device2}: {len(device2_sessions)}")
    
    # Check session synchronization status
    for session1 in device1_sessions:
        sync_targets = session1.get_sync_targets()
        print(f"Session {session1.session_token} sync targets: {sync_targets}")
        
        for target_token in sync_targets:
            sync_status = await session_manager.check_sync_status(session1.session_token, target_token)
            print(f"Sync status with {target_token}: {sync_status}")
    
    # Test network connectivity between devices
    connectivity = await session_manager.test_device_connectivity(device1, device2)
    print(f"Device connectivity: {connectivity}")
    
    # Check context synchronization
    context_sync = await session_manager.test_context_sync(device1_sessions[0], device2_sessions[0])
    print(f"Context sync test: {context_sync}")
```

### Performance Optimization for Embodiment

#### Session-Aware Tool Call Optimization
```python
class OptimizedEmbodimentClient:
    def __init__(self):
        self.session_cache = TTLCache(maxsize=100, ttl=1800)  # 30-minute session cache
        self.tool_result_cache = TTLCache(maxsize=1000, ttl=300)  # 5-minute result cache
        
    async def call_tool_optimized(self, session: EmbodimentSession, tool: str, params: dict) -> dict:
        # Check if tool result can be cached
        if self.is_cacheable_tool(tool, params):
            cache_key = self.create_session_cache_key(session.session_token, tool, params)
            
            if cache_key in self.tool_result_cache:
                # Update session activity without actual tool call
                session.update_last_activity()
                return self.tool_result_cache[cache_key]
        
        # Execute tool call
        result = await self.execute_tool_in_session(session, tool, params)
        
        # Cache result if appropriate
        if self.is_cacheable_tool(tool, params):
            self.tool_result_cache[cache_key] = result
            
        return result
        
    def is_cacheable_tool(self, tool: str, params: dict) -> bool:
        """Determine if tool results can be safely cached"""
        # Don't cache write operations or time-sensitive tools
        non_cacheable = ["file.write", "shell.execute", "time.now", "random.*"]
        return not any(tool.startswith(pattern.rstrip("*")) for pattern in non_cacheable)
```

#### Session Connection Pooling
```python
class EmbodimentConnectionManager:
    def __init__(self):
        self.host_connections = {}
        self.session_connections = {}
        
    async def get_session_connection(self, session: EmbodimentSession) -> MCPConnection:
        """Get optimized connection for embodiment session"""
        
        session_key = f"{session.host_id}:{session.session_token}"
        
        if session_key not in self.session_connections:
            # Create session-specific connection with session token
            connection = await self.create_session_connection(
                session.mcp_endpoint,
                session.session_token
            )
            self.session_connections[session_key] = connection
            
        return self.session_connections[session_key]
        
    async def cleanup_expired_sessions(self):
        """Clean up connections for expired sessions"""
        current_time = datetime.utcnow()
        expired_sessions = []
        
        for session_key, connection in self.session_connections.items():
            if connection.session_expired(current_time):
                expired_sessions.append(session_key)
                await connection.close()
                
        for session_key in expired_sessions:
            del self.session_connections[session_key]
```

This comprehensive MCP Integration guide demonstrates how to transform existing MCP tools into **Secure Hosted Embodiment** systems, enabling powerful cross-device collaboration while maintaining strong security boundaries and session management.