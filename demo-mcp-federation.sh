#!/bin/bash
# A script to demonstrate a full, working FEM MCP Federation loop.

set -e
trap 'kill $(jobs -p)' EXIT

echo "🚀 Starting FEM MCP Federation Demo"

# 1. Build components
echo "📦 Building broker and coder..."
make build

# 2. Start the broker in the background
echo "🔄 Starting FEM Broker on https://localhost:8443..."
./bin/fem-broker --listen :8443 > broker.log 2>&1 &
sleep 2 # Wait for broker to initialize

# 3. Start two different agents with MCP servers on different ports
echo "🤖 Starting Agent-1 (calculator) on MCP port 8080..."
./bin/fem-coder --agent calculator-001 --broker https://localhost:8443 --mcp-port 8080 > agent1.log 2>&1 &
sleep 2

echo "🤖 Starting Agent-2 (executor) on MCP port 8081..."
./bin/fem-coder --agent executor-001 --broker https://localhost:8443 --mcp-port 8081 > agent2.log 2>&1 &
sleep 2

echo "✅ Network is up. Broker and two agents are running."
echo "----------------------------------------------------"

# 4. Use curl to act as a third agent discovering tools via the broker
echo "🔍 Discovering all 'code.execute' tools via the broker..."

DISCOVERY_REQUEST='{
  "type": "discoverTools",
  "agent": "test-client",
  "ts": '$(date +%s%3N)',
  "nonce": "discovery-nonce-'$(date +%s)'",
  "body": {
    "query": { "capabilities": ["code.execute"] },
    "requestId": "discovery-req-1"
  }
}'

# Note: The tool client doesn't need a signature for this demo
# In a production system, all envelopes would be signed.
curl -k -s -X POST https://localhost:8443/ \
    -H "Content-Type: application/json" \
    -d "$DISCOVERY_REQUEST" | jq .

echo "----------------------------------------------------"
echo "📞 Calling the 'code.execute' tool directly on Agent-2's MCP endpoint..."

TOOL_CALL_REQUEST='{
    "jsonrpc": "2.0",
    "method": "tools/call",
    "params": {
        "name": "code.execute",
        "arguments": {
            "command": "echo Hello from a federated tool call!"
        }
    },
    "id": 1
}'

curl -s -X POST http://localhost:8081/mcp \
    -H "Content-Type: application/json" \
    -d "$TOOL_CALL_REQUEST" | jq .

echo "----------------------------------------------------"
echo "🎉 Demo Complete! The full federation loop is working."

# The trap will kill background jobs on script exit.