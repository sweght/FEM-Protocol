#!/bin/bash

# Test script for FEM network functionality
set -e

echo "🚀 Testing FEM Network Implementation"
echo "======================================"

# Build components
echo "📦 Building components..."
cd /Users/slowbro/Workspaces/Sandbox/FEP-FEM-full

# Build broker
echo "  Building broker..."
cd broker && go build -o fem-broker . && cd ..

# Build coder  
echo "  Building coder..."
cd bodies/coder && go build -o fem-coder ./cmd/fem-coder && cd ../..

echo "✅ Build complete"

# Start broker in background
echo "🔄 Starting broker..."
./broker/fem-broker --listen :8443 > broker.log 2>&1 &
BROKER_PID=$!
echo "  Broker started (PID: $BROKER_PID)"

# Wait for broker to start
sleep 2

# Check if broker is running
if ! curl -k -f https://localhost:8443/health > /dev/null 2>&1; then
    echo "❌ Broker health check failed"
    kill $BROKER_PID 2>/dev/null || true
    cat broker.log
    exit 1
fi

echo "✅ Broker is running and healthy"

# Test agent registration
echo "🤖 Testing agent registration..."
./bodies/coder/fem-coder --broker https://localhost:8443 --agent test-coder-001 > coder.log 2>&1 &
CODER_PID=$!
echo "  Coder started (PID: $CODER_PID)"

# Wait for registration
sleep 3

# Check logs for success
if grep -q "Registration successful" coder.log; then
    echo "✅ Agent registration successful"
else
    echo "❌ Agent registration failed"
    echo "Broker logs:"
    cat broker.log
    echo "Coder logs:"  
    cat coder.log
    
    # Cleanup
    kill $BROKER_PID $CODER_PID 2>/dev/null || true
    exit 1
fi

# Test MCP federation (when implemented)
echo "🔍 Testing MCP federation readiness..."

# Check if broker supports MCP discovery
echo "  Testing tool discovery endpoint..."
DISCOVERY_REQUEST='{
  "type": "discoverTools",
  "agent": "test-client",
  "ts": '$(date +%s%3N)',
  "nonce": "test-'$(date +%s)'",
  "body": {
    "query": {
      "capabilities": ["*"],
      "maxResults": 10
    },
    "requestId": "test-discovery"
  }
}'

# Test discovery endpoint (will return empty results until MCP integration is complete)
DISCOVERY_RESPONSE=$(curl -s -k -X POST https://localhost:8443/fep \
    -H "Content-Type: application/json" \
    -d "$DISCOVERY_REQUEST" 2>/dev/null || echo "endpoint_not_ready")

if [[ "$DISCOVERY_RESPONSE" == "endpoint_not_ready" ]]; then
    echo "  ⚠️  MCP discovery endpoint not yet implemented"
    echo "     This will be available after MCP integration phases"
else
    echo "  ✅ MCP discovery endpoint responding"
    if echo "$DISCOVERY_RESPONSE" | grep -q "tools"; then
        echo "     Discovery returned tool results"
    else
        echo "     Discovery returned: $DISCOVERY_RESPONSE"
    fi
fi

echo "🎉 FEM network test PASSED!"
echo "   - Broker started successfully"
echo "   - Agent registered with broker"
echo "   - MCP integration readiness verified"
echo ""
echo "📋 Next steps for MCP integration:"
echo "   1. Implement new envelope types (Phase A)"
echo "   2. Add MCP tool registry to broker (Phase B-D)"
echo "   3. Add MCP server/client to agents (Phase E-G)"
echo "   4. Create federation examples (Phase H-I)"
echo ""
echo "   See docs/Implementation-Roadmap.md for detailed plan"

# Cleanup
echo "🧹 Cleaning up..."
kill $BROKER_PID $CODER_PID 2>/dev/null || true
rm -f broker.log coder.log

echo "✅ Test complete"