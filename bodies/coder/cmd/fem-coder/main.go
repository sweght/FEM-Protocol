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
	ID       string
	BrokerURL string
	PubKey   ed25519.PublicKey
	PrivKey  ed25519.PrivateKey
	client   *http.Client
}

func main() {
	// Parse command line flags
	brokerURL := flag.String("broker", "https://localhost:4433", "Broker URL to connect to")
	agentID := flag.String("agent", "fem-coder-001", "Agent identifier")
	flag.Parse()

	log.Printf("fem-coder starting - Agent ID: %s, Broker: %s", *agentID, *brokerURL)

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
		client: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true, // For demo with self-signed certs
				},
			},
			Timeout: 10 * time.Second,
		},
	}

	// Register with broker
	if err := agent.registerWithBroker(); err != nil {
		log.Fatalf("Failed to register with broker: %v", err)
	}

	log.Println("Successfully registered with broker. Waiting for tool calls...")

	// Keep the agent running (in a real implementation, this would listen for incoming messages)
	select {}
}

func (a *Agent) registerWithBroker() error {
	// Create registration envelope
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
			Capabilities: []string{"code.execute", "shell.run", "file.read", "file.write"},
			PubKey:       protocol.EncodePublicKey(a.PubKey),
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
	resp, err := a.client.Post(a.BrokerURL+"/fep", "application/json", bytes.NewReader(data))
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