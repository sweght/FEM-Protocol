# Deployment Guide: Hosted Embodiment

Production deployment strategies for FEM Protocol networks enabling **Secure Hosted Embodiment** across various environments and scales.

## Table of Contents
- [Embodiment Deployment Overview](#embodiment-deployment-overview)
- [Host Agent Deployment](#host-agent-deployment)
- [Guest Agent Deployment](#guest-agent-deployment)
- [Cross-Device Embodiment Setup](#cross-device-embodiment-setup)
- [Container-Based Embodiment](#container-based-embodiment)
- [Cloud Embodiment Infrastructure](#cloud-embodiment-infrastructure)
- [Monitoring Embodiment Sessions](#monitoring-embodiment-sessions)
- [Scaling Embodiment Networks](#scaling-embodiment-networks)
- [Troubleshooting Embodiment](#troubleshooting-embodiment)

## Embodiment Deployment Overview

FEM Protocol supports multiple deployment patterns optimized for **Secure Hosted Embodiment** scenarios. Deployments must consider host body capabilities, guest discovery mechanisms, session management, and cross-device accessibility.

### Embodiment Architecture Patterns

**Personal Cross-Device Embodiment**
```
┌─────────────────────────────────────────────────────────────────────────────┐
│                      FEM Broker                                     │
│               (Embodiment Coordination)                              │
└──────────────────────────────┬───────────────────────────────────────────────────┘
                              │ Embodiment Discovery & Session Management
        ┌──────────────────────────────┼─────────────────────────────────────────────┐
        │                              │                                  │
┌───────▼─────────────────────────┐   ┌───────▼─────────────────────────┐   ┌───────▼─────────────────────────┐
│        Laptop Host Agent        │   │       Desktop Host Agent      │   │        Phone Guest Agent      │
│   • Development Body           │   │   • Gaming Body               │   │   • Discovers Host Bodies      │
│   • Terminal & File Access     │   │   • GPU & Rendering Tools     │   │   • Requests Embodiment        │
│   • MCP Server: :8080          │   │   • MCP Server: :8081          │   │   • Exercises Delegated Control│
└─────────────────────────────────┘   └─────────────────────────────────┘   └─────────────────────────────────┘

**Collaborative Virtual Presence**
```
                    ┌─────────────────────────────────────────────────────────┐
                    │                    FEM Broker                             │
                    └────────────────────────────┬──────────────────────────────┘
                                            │
┌─────────────────────────────────────────────┼─────────────────────────────────────────────┐
│           Live2D Host Application              │                Guest AI Agent                  │
│   • Avatar Body Definition                  │   • Requests Live2D Puppet Body           │
│   • Expression & Animation Controls        │   • Controls Avatar via MCP Tools         │
│   • Security: Avatar State Only           │   • Secure Session with Time Limits       │
└─────────────────────────────────────────────┘─────────────────────────────────────────────┘

**Enterprise Application Embodiment**
```
    ┌─────────────────────────────────────────────────────────────────────────────┐
    │                    Enterprise FEM Broker                              │
    │            (Multi-Tenant Embodiment Management)                        │
    └──────────────────────────────────────┬───────────────────────────────────────┘
                                  │
  ┌──────────────────────────────────┼───────────────────────────────────────┐
  │            Department A                   │           Department B                     │
  │   Host: ERP System Body                  │   Host: Analytics Platform Body           │
  │   Guest: Mobile Workflow Agent           │   Guest: Report Generation Agent          │
  └──────────────────────────────────┘───────────────────────────────────────┘
```

### Embodiment Deployment Considerations

- **Session Management** - Concurrent embodiment limits, session lifecycles, token management
- **Security Boundaries** - Permission enforcement, audit logging, resource isolation
- **Cross-Device Access** - Network discovery, mobile optimization, session handoff
- **Body Compatibility** - Environment adaptation, tool schema validation
- **Performance Isolation** - Resource limits per session, priority management
- **Availability Requirements** - Host redundancy, session failover, embodiment discovery
- **Compliance & Privacy** - Data boundaries, session audit trails, regulation adherence
- **Network Topology** - Latency optimization, bandwidth management, firewall traversal

## Single Node Deployment

### Development Setup

```bash
# Simple embodiment development deployment
./fem-broker --listen :8443 &
./fem-host-agent --broker https://localhost:8443 --agent laptop-host-alice &
```

### Cross-Device Embodiment Development Setup

```bash
# Start broker with embodiment coordination
./fem-broker --listen :8443 --embodiment-enabled &

# Start host agents offering different bodies
./fem-host-agent --broker https://localhost:8443 --agent laptop-host-alice --mcp-port 8080 \
  --body developer-workstation-v1 &

./fem-host-agent --broker https://localhost:8443 --agent desktop-host-alice --mcp-port 8081 \
  --body gaming-workstation-v1 &

# Test body discovery
curl -k -X POST https://localhost:8443/fem \
  -H "Content-Type: application/json" \
  -d '{
    "type": "discoverBodies",
    "agent": "phone-guest-bob",
    "ts": '$(date +%s%3N)',
    "nonce": "discover-'$(date +%s)'",
    "body": {
      "query": {"capabilities": ["shell.*", "file.read"], "environmentType": "local-development"},
      "guestProfile": {"guestId": "phone-guest-bob", "intendedUse": "mobile-development"},
      "requestId": "discover-dev-bodies"
    }
  }'

# Test embodiment request
curl -k -X POST https://localhost:8443/fem \
  -H "Content-Type: application/json" \
  -d '{
    "type": "requestEmbodiment",
    "agent": "phone-guest-bob",
    "ts": '$(date +%s%3N)',
    "nonce": "embody-'$(date +%s)'",
    "body": {
      "hostAgentId": "laptop-host-alice",
      "bodyId": "developer-workstation-v1",
      "requestedDuration": 1800,
      "intendedActions": ["Check project status", "Run development server"],
      "requestId": "embody-dev-001"
    }
  }'
```

### Production Single Node

#### 1. System Requirements

```bash
# Minimum requirements for embodiment hosting
CPU: 2 cores
Memory: 4GB RAM
Storage: 20GB SSD
Network: 1Gbps

# Recommended for multi-guest embodiment
CPU: 8 cores
Memory: 16GB RAM
Storage: 100GB NVMe SSD
Network: 10Gbps

# Additional considerations for embodiment:
# - Separate memory allocation per embodiment session
# - Storage isolation for guest workspaces
# - Network bandwidth for real-time tool execution
```

#### 2. Installation

```bash
# Download release
wget https://github.com/chazmaniandinkle/FEP-FEM/releases/latest/download/fem-v0.1.3-linux-amd64.tar.gz
tar -xzf fem-v0.1.3-linux-amd64.tar.gz
sudo mv fem-* /usr/local/bin/

# Create service user
sudo useradd --system --shell /bin/false fem-broker
sudo mkdir -p /etc/fem /var/lib/fem /var/log/fem
sudo chown fem-broker:fem-broker /var/lib/fem /var/log/fem
```

#### 3. TLS Certificate Setup

```bash
# Generate production certificate
sudo openssl req -new -x509 -days 365 -nodes \
  -out /etc/fem/broker.crt \
  -keyout /etc/fem/broker.key \
  -subj "/CN=broker.example.com" \
  -addext "subjectAltName=DNS:broker.example.com,DNS:localhost,IP:10.0.1.100"

sudo chown fem-broker:fem-broker /etc/fem/broker.*
sudo chmod 600 /etc/fem/broker.key
```

#### 4. Systemd Service

```ini
# /etc/systemd/system/fem-broker.service
[Unit]
Description=FEM Broker - Embodiment Coordination
After=network.target
Wants=network-online.target

[Service]
Type=simple
User=fem-broker
Group=fem-broker
ExecStart=/usr/local/bin/fem-broker \
  --listen :8443 \
  --cert /etc/fem/broker.crt \
  --key /etc/fem/broker.key \
  --log-level info \
  --embodiment-enabled \
  --session-timeout 3600 \
  --max-concurrent-sessions 100
Restart=always
RestartSec=5s
LimitNOFILE=1048576
NoNewPrivileges=yes
ProtectSystem=strict
ProtectHome=yes
PrivateTmp=yes

[Install]
WantedBy=multi-user.target
```

```bash
# Enable and start service
sudo systemctl daemon-reload
sudo systemctl enable fem-broker
sudo systemctl start fem-broker
sudo systemctl status fem-broker
```

#### 5. Firewall Configuration

```bash
# UFW firewall rules
sudo ufw allow 8443/tcp comment "FEM Broker"
sudo ufw enable

# Or iptables
sudo iptables -A INPUT -p tcp --dport 8443 -j ACCEPT
sudo iptables-save > /etc/iptables/rules.v4
```

### Load Balancer Setup

```nginx
# /etc/nginx/conf.d/fem-broker.conf
upstream fem-brokers {
    server 10.0.1.100:8443;
    # Add more brokers for HA
    # server 10.0.1.101:8443;
    # server 10.0.1.102:8443;
}

server {
    listen 443 ssl http2;
    server_name fem.example.com;
    
    ssl_certificate /etc/ssl/certs/fem.example.com.crt;
    ssl_certificate_key /etc/ssl/private/fem.example.com.key;
    
    location / {
        proxy_pass https://fem-brokers;
        proxy_ssl_verify off;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

## Multi-Node Embodiment Networks

### Hub-and-Spoke Embodiment Topology

#### Central Embodiment Hub

```bash
# Hub broker for embodiment coordination
./fem-broker \
  --listen :8443 \
  --cert /etc/ssl/certs/hub-broker.crt \
  --key /etc/ssl/private/hub-broker.key \
  --embodiment-mode hub \
  --federation-listen :8444 \
  --cross-region-embodiment true
```

#### Regional Embodiment Brokers

```bash
# Regional broker for local embodiment
./fem-broker \
  --listen :8443 \
  --cert /etc/ssl/certs/regional-broker.crt \
  --key /etc/ssl/private/regional-broker.key \
  --embodiment-mode spoke \
  --hub-broker https://hub.example.com:8444 \
  --local-embodiment-priority true
```

### Mesh Embodiment Topology

#### Broker A Configuration

```yaml
# broker-a.yaml
listen: ":8443"
cert: "/etc/ssl/certs/broker-a.crt"
key: "/etc/ssl/private/broker-a.key"
embodiment:
  mode: "mesh"
  peers:
    - "https://broker-b.example.com:8443"
    - "https://broker-c.example.com:8443"
  discovery:
    enabled: true
    interval: "30s"
    cross_broker_sessions: true
```

#### Service Discovery

```bash
# DNS-based discovery
dig SRV _fem._tcp.brokers.example.com

# Returns:
# broker-a.example.com. 0 0 8443
# broker-b.example.com. 0 0 8443  
# broker-c.example.com. 0 0 8443
```

### Cross-Organization Embodiment Federation

#### Organization A (Hospital) Configuration

```yaml
# hospital-a-broker.yaml
listen: ":8443"
cert: "/etc/ssl/certs/hospital-a.crt"
key: "/etc/ssl/private/hospital-a.key"
embodiment:
  cross_org_enabled: true
  body_sharing:
    external_access: "restricted"
    allowed_body_types:
      - "data-analysis-secure"
      - "statistical-compute"
    denied_body_types:
      - "patient-data-access"
federation:
  mode: "cross-org"
  peers:
    - name: "hospital-b"
      endpoint: "https://hospital-b.medical.network:8443"
      trust_policy: "verified"
      embodiment_access: "mutual"
security:
  embodiment_policies:
    - pattern: "patient.*"
      access: "internal-only"
    - pattern: "analysis.*"
      access: "federated"
    - pattern: "secure-compute.*"
      access: "cross-org"
```

#### Organization B (Hospital) Configuration

```yaml
# hospital-b-broker.yaml
listen: ":8443"
cert: "/etc/ssl/certs/hospital-b.crt"
key: "/etc/ssl/private/hospital-b.key"
embodiment:
  cross_org_enabled: true
  body_sharing:
    external_access: "restricted"
    allowed_body_types:
      - "ml-training-secure"
      - "model-inference"
    denied_body_types:
      - "patient-data-access"
federation:
  mode: "cross-org"
  peers:
    - name: "hospital-a"
      endpoint: "https://hospital-a.medical.network:8443"
      trust_policy: "verified"
      embodiment_access: "mutual"
security:
  embodiment_policies:
    - pattern: "patient.*"
      access: "internal-only"
    - pattern: "ml.*"
      access: "federated"
    - pattern: "secure-compute.*"
      access: "cross-org"
```

#### Cross-Org Embodiment Example

```bash
# Hospital A: Host offering secure data analysis body
./fem-host-agent \
  --broker https://hospital-a.medical.network:8443 \
  --agent hospital-a-host \
  --body data-analysis-secure \
  --mcp-port 8080 \
  --embodiment-policy /etc/fem/hospital-embodiment.json

# Hospital B: Guest seeking analysis capabilities
./fem-guest-agent \
  --broker https://hospital-b.medical.network:8443 \
  --agent hospital-b-researcher \
  --target-capabilities "data.aggregate,analysis.statistical" \
  --cross-org-enabled
```

#### Embodiment Policy Example

```json
{
  "name": "hospital-embodiment-policy",
  "environments": {
    "secure-clinical": {
      "description": "Secure clinical environment with patient data access",
      "constraints": {
        "data_locality": "required",
        "encryption": "required",
        "audit_logging": "required"
      },
      "allowed_tools": [
        "data.aggregate",
        "analysis.statistical"
      ],
      "denied_tools": [
        "data.export",
        "data.raw"
      ]
    },
    "research-network": {
      "description": "Research collaboration environment",
      "constraints": {
        "data_anonymization": "required",
        "result_sharing": "allowed"
      },
      "allowed_tools": [
        "ml.training",
        "model.inference",
        "analysis.aggregate"
      ]
    }
  }
}
```

## Container Orchestration

### Docker Compose

```yaml
# docker-compose.yml
version: '3.8'

services:
  fem-broker:
    image: fem-broker:latest
    ports:
      - "8443:8443"
    environment:
      - FEM_LISTEN=:8443
      - FEM_TLS_CERT=/certs/broker.crt
      - FEM_TLS_KEY=/certs/broker.key
      - FEM_EMBODIMENT_ENABLED=true
      - FEM_SESSION_TIMEOUT=3600s
      - FEM_MAX_CONCURRENT_SESSIONS=100
    volumes:
      - ./certs:/certs:ro
      - broker-data:/var/lib/fem
    networks:
      - fem-network
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "curl", "-f", "-k", "https://localhost:8443/health"]
      interval: 30s
      timeout: 10s
      retries: 3

  fem-host-dev:
    image: fem-host-agent:latest
    environment:
      - FEM_BROKER_URL=https://fem-broker:8443
      - FEM_AGENT_ID=dev-host-001
      - FEM_AGENT_TYPE=host
      - FEM_MCP_PORT=8080
      - FEM_BODY_ID=container-dev-v1
      - FEM_BODY_TOOLS=shell.execute,file.read,file.write,code.execute
      - FEM_ENVIRONMENT_TYPE=container
    depends_on:
      - fem-broker
    networks:
      - fem-network
    restart: unless-stopped
    deploy:
      replicas: 2

  fem-host-data:
    image: fem-host-agent:latest
    environment:
      - FEM_BROKER_URL=https://fem-broker:8443
      - FEM_AGENT_ID=data-host-001
      - FEM_AGENT_TYPE=host
      - FEM_MCP_PORT=8081
      - FEM_BODY_ID=data-processing-v1
      - FEM_BODY_TOOLS=data.analyze,data.transform,file.read
      - FEM_ENVIRONMENT_TYPE=container
    depends_on:
      - fem-broker
    networks:
      - fem-network
    restart: unless-stopped
    deploy:
      replicas: 2

volumes:
  broker-data:

networks:
  fem-network:
    driver: bridge
```

### Kubernetes Deployment

#### Namespace and RBAC

```yaml
# namespace.yaml
apiVersion: v1
kind: Namespace
metadata:
  name: fem-system
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: fem-broker
  namespace: fem-system
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: fem-broker
rules:
- apiGroups: [""]
  resources: ["pods", "services", "endpoints"]
  verbs: ["get", "list", "watch"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: fem-broker
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: fem-broker
subjects:
- kind: ServiceAccount
  name: fem-broker
  namespace: fem-system
```

#### Broker Deployment

```yaml
# broker-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: fem-broker
  namespace: fem-system
spec:
  replicas: 3
  selector:
    matchLabels:
      app: fem-broker
  template:
    metadata:
      labels:
        app: fem-broker
    spec:
      serviceAccountName: fem-broker
      containers:
      - name: fem-broker
        image: fem-broker:v0.1.3
        ports:
        - containerPort: 8443
          name: https
        env:
        - name: FEM_LISTEN
          value: ":8443"
        - name: FEM_TLS_CERT
          value: "/etc/tls/tls.crt"
        - name: FEM_TLS_KEY
          value: "/etc/tls/tls.key"
        volumeMounts:
        - name: tls
          mountPath: /etc/tls
          readOnly: true
        resources:
          requests:
            memory: "256Mi"
            cpu: "100m"
          limits:
            memory: "512Mi"
            cpu: "500m"
        livenessProbe:
          httpGet:
            path: /health
            port: 8443
            scheme: HTTPS
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /health
            port: 8443
            scheme: HTTPS
          initialDelaySeconds: 5
          periodSeconds: 5
      volumes:
      - name: tls
        secret:
          secretName: fem-broker-tls
---
apiVersion: v1
kind: Service
metadata:
  name: fem-broker
  namespace: fem-system
spec:
  selector:
    app: fem-broker
  ports:
  - port: 8443
    targetPort: 8443
    name: https
  type: LoadBalancer
```

#### Agent DaemonSet

```yaml
# agent-daemonset.yaml
apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: fem-coder
  namespace: fem-system
spec:
  selector:
    matchLabels:
      app: fem-coder
  template:
    metadata:
      labels:
        app: fem-coder
    spec:
      containers:
      - name: fem-coder
        image: fem-coder:v0.1.3
        env:
        - name: FEM_BROKER_URL
          value: "https://fem-broker:8443"
        - name: FEM_AGENT_ID
          valueFrom:
            fieldRef:
              fieldPath: spec.nodeName
        - name: FEM_AGENT_CAPABILITIES
          value: "code.execute,shell.run,file.read"
        resources:
          requests:
            memory: "128Mi"
            cpu: "100m"
          limits:
            memory: "512Mi"
            cpu: "500m"
        securityContext:
          runAsNonRoot: true
          runAsUser: 65534
          allowPrivilegeEscalation: false
          readOnlyRootFilesystem: true
```

### Helm Chart

```yaml
# Chart.yaml
apiVersion: v2
name: fem
description: Federated Embodied Mesh
type: application
version: 0.1.3
appVersion: "0.1.3"

# values.yaml
broker:
  replicaCount: 3
  image:
    repository: fem-broker
    tag: v0.1.3
    pullPolicy: IfNotPresent
  
  service:
    type: LoadBalancer
    port: 8443
  
  tls:
    enabled: true
    secretName: fem-broker-tls
  
  resources:
    requests:
      memory: 256Mi
      cpu: 100m
    limits:
      memory: 512Mi
      cpu: 500m

agents:
  coder:
    enabled: true
    replicaCount: 5
    image:
      repository: fem-coder
      tag: v0.1.3
    capabilities:
      - "code.execute"
      - "shell.run"
      - "file.read"
```

## Cloud Deployments

### AWS Deployment

#### EC2 with Auto Scaling

```yaml
# cloudformation-template.yaml
AWSTemplateFormatVersion: '2010-09-09'
Resources:
  FEMBrokerLaunchTemplate:
    Type: AWS::EC2::LaunchTemplate
    Properties:
      LaunchTemplateName: fem-broker-template
      LaunchTemplateData:
        ImageId: ami-0abcdef1234567890  # Amazon Linux 2
        InstanceType: t3.medium
        SecurityGroupIds:
          - !Ref FEMSecurityGroup
        UserData:
          Fn::Base64: !Sub |
            #!/bin/bash
            yum update -y
            wget https://github.com/chazmaniandinkle/FEP-FEM/releases/latest/download/fem-v0.1.3-linux-amd64.tar.gz
            tar -xzf fem-v0.1.3-linux-amd64.tar.gz
            mv fem-broker /usr/local/bin/
            
            # Start broker
            /usr/local/bin/fem-broker --listen :8443

  FEMAutoScalingGroup:
    Type: AWS::AutoScaling::AutoScalingGroup
    Properties:
      LaunchTemplate:
        LaunchTemplateId: !Ref FEMBrokerLaunchTemplate
        Version: !GetAtt FEMBrokerLaunchTemplate.LatestVersionNumber
      MinSize: 1
      MaxSize: 5
      DesiredCapacity: 3
      VPCZoneIdentifier:
        - subnet-12345678
        - subnet-87654321
      TargetGroupARNs:
        - !Ref FEMTargetGroup

  FEMLoadBalancer:
    Type: AWS::ElasticLoadBalancingV2::LoadBalancer
    Properties:
      Type: application
      Scheme: internet-facing
      Subnets:
        - subnet-12345678
        - subnet-87654321
      SecurityGroups:
        - !Ref FEMSecurityGroup
```

#### EKS Deployment

```bash
# Create EKS cluster
eksctl create cluster \
  --name fem-cluster \
  --region us-west-2 \
  --nodegroup-name fem-nodes \
  --node-type t3.medium \
  --nodes 3 \
  --nodes-min 1 \
  --nodes-max 10

# Deploy FEM
kubectl apply -f k8s/
```

### Google Cloud Platform

```yaml
# gke-cluster.yaml
apiVersion: container.v1
kind: Cluster
metadata:
  name: fem-cluster
spec:
  location: us-central1
  initialNodeCount: 3
  nodeConfig:
    machineType: e2-standard-2
    oauthScopes:
    - "https://www.googleapis.com/auth/cloud-platform"
```

### Azure Deployment

```yaml
# aks-cluster.yaml
apiVersion: v1
kind: Service
metadata:
  name: fem-broker-service
spec:
  type: LoadBalancer
  selector:
    app: fem-broker
  ports:
  - port: 8443
    targetPort: 8443
```

## Monitoring and Operations

### Prometheus Monitoring

```yaml
# prometheus-config.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: prometheus-config
data:
  prometheus.yml: |
    global:
      scrape_interval: 15s
    
    scrape_configs:
    - job_name: 'fem-broker'
      static_configs:
      - targets: ['fem-broker:8443']
      scheme: https
      tls_config:
        insecure_skip_verify: true
      metrics_path: /metrics
    
    - job_name: 'fem-agents'
      kubernetes_sd_configs:
      - role: pod
        namespaces:
          names:
          - fem-system
      relabel_configs:
      - source_labels: [__meta_kubernetes_pod_label_app]
        action: keep
        regex: fem-coder
```

### Grafana Dashboard

```json
{
  "dashboard": {
    "title": "FEM Network Dashboard",
    "panels": [
      {
        "title": "Active Agents",
        "type": "stat",
        "targets": [
          {
            "expr": "fem_broker_registered_agents_total"
          }
        ]
      },
      {
        "title": "Message Throughput",
        "type": "graph",
        "targets": [
          {
            "expr": "rate(fem_broker_messages_processed_total[5m])"
          }
        ]
      },
      {
        "title": "Tool Execution Latency",
        "type": "heatmap",
        "targets": [
          {
            "expr": "histogram_quantile(0.95, rate(fem_agent_tool_execution_duration_seconds_bucket[5m]))"
          }
        ]
      }
    ]
  }
}
```

### Log Aggregation

```yaml
# fluentd-config.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: fluentd-config
data:
  fluent.conf: |
    <source>
      @type tail
      path /var/log/fem/*.log
      pos_file /var/log/fluentd-fem.log.pos
      tag fem.*
      format json
    </source>
    
    <match fem.**>
      @type elasticsearch
      host elasticsearch.logging.svc.cluster.local
      port 9200
      index_name fem-logs
    </match>
```

### Health Checks

```bash
# Health check script
#!/bin/bash
BROKER_URL="https://fem-broker:8443"

# Check broker health
curl -k -f "$BROKER_URL/health" || exit 1

# Check agent count
AGENT_COUNT=$(curl -k -s "$BROKER_URL/metrics" | grep fem_broker_registered_agents_total | awk '{print $2}')
if [ "$AGENT_COUNT" -lt 1 ]; then
  echo "No agents registered"
  exit 1
fi

echo "Health check passed: $AGENT_COUNT agents registered"
```

## Scaling Strategies

### Horizontal Scaling

#### Broker Scaling

```bash
# Scale broker replicas
kubectl scale deployment fem-broker --replicas=5

# Auto-scaling based on CPU
kubectl autoscale deployment fem-broker --cpu-percent=70 --min=3 --max=10
```

#### Agent Scaling

```yaml
# HorizontalPodAutoscaler for agents
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: fem-coder-hpa
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: fem-coder
  minReplicas: 5
  maxReplicas: 100
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
  - type: Pods
    pods:
      metric:
        name: fem_agent_queue_length
      target:
        type: AverageValue
        averageValue: "5"
```

### Geographic Distribution

```yaml
# Multi-region deployment
regions:
  us-west-2:
    brokers: 3
    agents: 10
  us-east-1:
    brokers: 3
    agents: 10
  eu-west-1:
    brokers: 2
    agents: 5

federation:
  topology: mesh
  cross_region: true
  latency_threshold: 100ms
```

### Performance Tuning

```yaml
# Broker performance configuration
broker:
  max_connections: 10000
  message_buffer_size: 1000
  worker_pool_size: 100
  tls_config:
    max_version: "1.3"
    cipher_suites:
      - "TLS_AES_256_GCM_SHA384"
      - "TLS_CHACHA20_POLY1305_SHA256"

agent:
  max_concurrent_tools: 10
  tool_timeout: 30s
  heartbeat_interval: 10s
```

## Troubleshooting Embodiment

### Common Embodiment Issues

#### 1. Embodiment Request Failures

```bash
# Check broker connectivity
curl -k -v https://fem-broker:8443/health

# Test body discovery
curl -k -X POST https://fem-broker:8443/fem \
  -H "Content-Type: application/json" \
  -d '{"type": "discoverBodies", "agent": "test-guest", "body": {"query": {"capabilities": ["*"]}}}'

# Check host agent availability
kubectl logs -f deployment/fem-host-dev
```

#### 2. Session Management Issues

```bash
# Check active embodiment sessions
curl -k -s https://fem-broker:8443/metrics | grep embodiment_sessions

# Monitor session timeouts
kubectl logs -f deployment/fem-broker | grep "session.*expired"

# Verify session tokens
./fem-debug verify-session --token session-abc123
```

#### 3. Permission and Security Issues

```bash
# Check security policy violations
kubectl logs -f deployment/fem-host-dev | grep "permission.*denied"

# Monitor audit logs
tail -f /var/log/fem/embodiment-audit.log

# Test tool execution permissions
curl -k -X POST https://host:8080/mcp/sessions/token123/tools/call \
  -H "Content-Type: application/json" \
  -d '{"tool": "shell.execute", "parameters": {"command": "pwd"}}'
```

### Debug Tools

```bash
# FEM embodiment diagnostic script
#!/bin/bash
echo "=== FEM Embodiment Diagnostics ==="

echo "1. Broker Health:"
curl -k -s https://fem-broker:8443/health | jq .

echo "2. Active Embodiment Sessions:"
curl -k -s https://fem-broker:8443/metrics | grep embodiment_sessions

echo "3. Host Agent Count:"
kubectl get pods -l app=fem-host-agent --no-headers | wc -l

echo "4. Network Connectivity:"
kubectl exec -it deployment/fem-host-dev -- curl -k https://fem-broker:8443/health

echo "5. Resource Usage:"
kubectl top pods -n fem-system

echo "6. Recent Embodiment Events:"
kubectl get events -n fem-system --sort-by='.lastTimestamp' | grep -i embodiment | tail -10
```

### Log Analysis

```bash
# Analyze broker logs for embodiment errors
kubectl logs deployment/fem-broker | grep -i "embodiment.*error"

# Check host agent body registration patterns  
kubectl logs deployment/fem-host-dev | grep "body.*registered"

# Monitor embodiment session creation rates
kubectl logs deployment/fem-broker | grep "embodiment.*granted" | tail -100

# Track tool execution patterns
kubectl logs deployment/fem-host-dev | grep "tool.*executed" | tail -50
```

This deployment guide provides comprehensive strategies for deploying FEM Protocol networks that enable **Secure Hosted Embodiment** from development to large-scale production environments.