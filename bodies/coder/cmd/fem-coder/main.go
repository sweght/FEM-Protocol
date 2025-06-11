package main

import (
	"bytes"
	"crypto/ed25519"
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"time"

	"github.com/fep-fem/protocol"
)

type Agent struct {
	ID        string
	BrokerURL string
	PubKey    ed25519.PublicKey
	PrivKey   ed25519.PrivateKey
	client    *http.Client
	mcpServer *http.Server
	mcpPort   int
}

type ToolHandler func(params map[string]interface{}) (interface{}, error)

func main() {
	// Parse command line flags
	brokerURL := flag.String("broker", "https://localhost:4433", "Broker URL to connect to")
	agentID := flag.String("agent", "fem-coder-001", "Agent identifier")
	mcpPort := flag.Int("mcp-port", 8080, "Port for MCP server to listen on")
	flag.Parse()

	log.Printf("fem-coder starting - Agent ID: %s, Broker: %s, MCP Port: %d", *agentID, *brokerURL, *mcpPort)

	// Generate key pair for this agent
	pubKey, privKey, err := protocol.GenerateKeyPair()
	if err != nil {
		log.Fatalf("Failed to generate key pair: %v", err)
	}

	// Create agent
	agent := &Agent{
		ID:        *agentID,
		BrokerURL: *brokerURL,
		PubKey:    pubKey,
		PrivKey:   privKey,
		mcpPort:   *mcpPort,
		client: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true, // For demo with self-signed certs
				},
			},
			Timeout: 10 * time.Second,
		},
	}

	// Start MCP server
	if err := agent.initializeAndStartMCPServer(); err != nil {
		log.Fatalf("Failed to start MCP server: %v", err)
	}

	// Register with broker
	if err := agent.registerWithBroker(); err != nil {
		log.Fatalf("Failed to register with broker: %v", err)
	}

	log.Println("Registration successful. Agent is running with MCP endpoint.")

	// Keep the agent running (in a real implementation, this would listen for incoming messages)
	select {}
}

func (a *Agent) initializeAndStartMCPServer() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/mcp", a.handleMCPRequest)

	a.mcpServer = &http.Server{
		Addr:    fmt.Sprintf(":%d", a.mcpPort),
		Handler: mux,
	}

	log.Printf("Starting MCP server for agent %s on port %d", a.ID, a.mcpPort)
	go func() {
		if err := a.mcpServer.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("MCP server for agent %s failed: %v", a.ID, err)
		}
	}()
	return nil
}

func (a *Agent) handleMCPRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var reqBody struct {
		Method string `json:"method"`
		Params struct {
			Name      string                 `json:"name"`
			Arguments map[string]interface{} `json:"arguments"`
		} `json:"params"`
		ID int `json:"id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		http.Error(w, "Invalid JSON request", http.StatusBadRequest)
		return
	}

	if reqBody.Method != "tools/call" {
		http.Error(w, "Unsupported method", http.StatusBadRequest)
		return
	}

	handlers := map[string]ToolHandler{
		"code.execute": a.handleCodeOrShellExecution,
		"shell.run":    a.handleCodeOrShellExecution,
	}

	handler, exists := handlers[reqBody.Params.Name]
	if !exists {
		http.Error(w, fmt.Sprintf("Tool '%s' not found", reqBody.Params.Name), http.StatusNotFound)
		return
	}

	result, err := handler(reqBody.Params.Arguments)

	var responseBody map[string]interface{}
	if err != nil {
		responseBody = map[string]interface{}{
			"jsonrpc": "2.0",
			"error": map[string]interface{}{
				"code":    -32603,
				"message": err.Error(),
			},
			"id": reqBody.ID,
		}
	} else {
		responseBody = map[string]interface{}{
			"jsonrpc": "2.0",
			"result":  result,
			"id":      reqBody.ID,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(responseBody)
}

func (a *Agent) handleCodeOrShellExecution(params map[string]interface{}) (interface{}, error) {
	command, ok := params["code"].(string)
	if !ok {
		command, ok = params["command"].(string)
	}
	if !ok {
		return nil, fmt.Errorf("parameter 'code' or 'command' of type string is required")
	}

	if tool, p_ok := params["tool"].(string); p_ok && tool == "shell.run" {
		cmd := exec.Command("sh", "-c", command)
		output, err := cmd.CombinedOutput()
		if err != nil {
			return nil, fmt.Errorf("execution failed: %w, output: %s", err, string(output))
		}
		return map[string]interface{}{"output": string(output)}, nil
	}
	
	cmd := exec.Command("sh", "-c", command)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("execution failed: %w, output: %s", err, string(output))
	}
	return map[string]interface{}{"output": string(output)}, nil
}

func (a *Agent) registerWithBroker() error {
	mcpTools := []protocol.MCPTool{
		{Name: "code.execute", Description: "Executes a command and returns its output."},
		{Name: "shell.run", Description: "Runs a shell command."},
	}
	
	capabilities := make([]string, len(mcpTools))
	for i, tool := range mcpTools {
		capabilities[i] = tool.Name
	}

	bodyDef := &protocol.BodyDefinition{
		Name:         "default-coder-body",
		Environment:  "local-dev",
		Capabilities: capabilities,
		MCPTools:     mcpTools,
	}

	envelope := &protocol.RegisterAgentEnvelope{
		BaseEnvelope: protocol.BaseEnvelope{
			Type: protocol.EnvelopeRegisterAgent,
			CommonHeaders: protocol.CommonHeaders{
				Agent: a.ID,
				TS:    time.Now().UnixMilli(),
				Nonce: fmt.Sprintf("%d", time.Now().UnixNano()),
			},
		},
		Body: protocol.RegisterAgentBody{
			PubKey:          protocol.EncodePublicKey(a.PubKey),
			Capabilities:    capabilities,
			MCPEndpoint:     fmt.Sprintf("http://localhost:%d/mcp", a.mcpPort),
			BodyDefinition:  bodyDef,
			EnvironmentType: "local-dev",
		},
	}

	// Sign the envelope
	if err := envelope.Sign(a.PrivKey); err != nil {
		return fmt.Errorf("failed to sign envelope: %w", err)
	}

	// Marshal to JSON
	data, err := json.Marshal(envelope)
	if err != nil {
		return fmt.Errorf("failed to marshal envelope: %w", err)
	}

	// Send to broker
	resp, err := a.client.Post(a.BrokerURL+"/", "application/json", bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("failed to send registration: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("broker returned status %d", resp.StatusCode)
	}

	log.Printf("Registration successful - Agent %s registered with broker", a.ID)
	return nil
}

// executeCode handles code execution tool calls
func (a *Agent) executeCode(command string, args []string) (string, error) {
	log.Printf("Executing: %s %v", command, args)
	
	cmd := exec.Command(command, args...)
	output, err := cmd.CombinedOutput()
	
	if err != nil {
		return "", fmt.Errorf("execution failed: %w, output: %s", err, string(output))
	}
	
	return string(output), nil
}

// handleToolCall processes incoming tool call requests
func (a *Agent) handleToolCall(envelope *protocol.ToolCallEnvelope) (*protocol.ToolResultEnvelope, error) {
	toolName := envelope.Body.Tool
	params := envelope.Body.Parameters
	
	log.Printf("Handling tool call: %s", toolName)
	
	var result interface{}
	var execError string
	
	switch toolName {
	case "code.execute":
		// Extract command and args from parameters
		command, ok := params["command"].(string)
		if !ok {
			execError = "missing or invalid 'command' parameter"
		} else {
			argsSlice := []string{}
			if args, exists := params["args"]; exists {
				if argsList, ok := args.([]interface{}); ok {
					for _, arg := range argsList {
						if argStr, ok := arg.(string); ok {
							argsSlice = append(argsSlice, argStr)
						}
					}
				}
			}
			
			output, err := a.executeCode(command, argsSlice)
			if err != nil {
				execError = err.Error()
			} else {
				result = map[string]interface{}{
					"output": output,
					"status": "success",
				}
			}
		}
		
	case "shell.run":
		// Simple shell execution
		command, ok := params["command"].(string)
		if !ok {
			execError = "missing or invalid 'command' parameter"
		} else {
			output, err := a.executeCode("sh", []string{"-c", command})
			if err != nil {
				execError = err.Error()
			} else {
				result = map[string]interface{}{
					"output": output,
					"status": "success",
				}
			}
		}
		
	default:
		execError = fmt.Sprintf("unknown tool: %s", toolName)
	}
	
	// Create result envelope
	resultEnvelope := &protocol.ToolResultEnvelope{
		BaseEnvelope: protocol.BaseEnvelope{
			Type: protocol.EnvelopeToolResult,
			CommonHeaders: protocol.CommonHeaders{
				Agent: a.ID,
				TS:    time.Now().UnixMilli(),
				Nonce: fmt.Sprintf("%d", time.Now().UnixNano()),
			},
		},
		Body: protocol.ToolResultBody{
			RequestID: envelope.Nonce, // Use the original request nonce as ID
			Success:   execError == "",
			Result:    result,
			Error:     execError,
		},
	}
	
	return resultEnvelope, nil
}