package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"log"
	"os"
)

// FEPEnvelope represents a basic FEP envelope structure
type FEPEnvelope struct {
	Type    string          `json:"type,omitempty"`
	Agent   string          `json:"agent,omitempty"`
	TS      int64           `json:"ts,omitempty"`
	Nonce   string          `json:"nonce,omitempty"`
	Sig     string          `json:"sig,omitempty"`
	Payload json.RawMessage `json:"payload,omitempty"`
}

// ToolCallPayload represents the payload for a toolCall envelope
type ToolCallPayload struct {
	Tool       string                 `json:"tool"`
	Parameters map[string]interface{} `json:"parameters"`
}

func main() {
	// Parse command line flags (for consistency, though coder reads from stdin)
	listenAddr := flag.String("listen", ":4433", "Address to listen on (unused for coder)")
	flag.Parse()

	log.Printf("fem-coder started (listen flag: %s, but reading from stdin)", *listenAddr)

	// Create scanner for stdin
	scanner := bufio.NewScanner(os.Stdin)
	writer := bufio.NewWriter(os.Stdout)

	// Process each line from stdin
	for scanner.Scan() {
		line := scanner.Bytes()

		// Try to parse as JSON envelope
		var envelope FEPEnvelope
		if err := json.Unmarshal(line, &envelope); err != nil {
			log.Printf("Invalid JSON received: %v", err)
			continue
		}

		// Process based on envelope type
		if envelope.Type == "toolCall" {
			handleToolCall(&envelope, writer)
		} else {
			// Echo other envelope types back
			if _, err := writer.Write(line); err != nil {
				log.Printf("Failed to write response: %v", err)
				continue
			}
			if err := writer.WriteByte('\n'); err != nil {
				log.Printf("Failed to write newline: %v", err)
				continue
			}
			if err := writer.Flush(); err != nil {
				log.Printf("Failed to flush: %v", err)
				continue
			}
			log.Printf("Echoed envelope type: %s", envelope.Type)
		}
	}

	if err := scanner.Err(); err != nil {
		log.Printf("Scanner error: %v", err)
	}
}

func handleToolCall(envelope *FEPEnvelope, writer *bufio.Writer) {
	// Parse the toolCall payload
	var payload ToolCallPayload
	if err := json.Unmarshal(envelope.Payload, &payload); err != nil {
		log.Printf("Failed to parse toolCall payload: %v", err)
		return
	}

	log.Printf("Processing toolCall for tool: %s", payload.Tool)

	// Create a simple toolResult response
	resultEnvelope := FEPEnvelope{
		Type:  "toolResult",
		Agent: "fem-coder",
		TS:    envelope.TS,
		Nonce: envelope.Nonce,
		Sig:   "placeholder-signature",
	}

	// Create result payload
	resultPayload := map[string]interface{}{
		"tool":   payload.Tool,
		"result": map[string]interface{}{"status": "processed", "message": "Tool call processed by fem-coder"},
	}

	payloadBytes, err := json.Marshal(resultPayload)
	if err != nil {
		log.Printf("Failed to marshal result payload: %v", err)
		return
	}
	resultEnvelope.Payload = payloadBytes

	// Write the result
	resultBytes, err := json.Marshal(resultEnvelope)
	if err != nil {
		log.Printf("Failed to marshal result envelope: %v", err)
		return
	}

	if _, err := writer.Write(resultBytes); err != nil {
		log.Printf("Failed to write result: %v", err)
		return
	}
	if err := writer.WriteByte('\n'); err != nil {
		log.Printf("Failed to write newline: %v", err)
		return
	}
	if err := writer.Flush(); err != nil {
		log.Printf("Failed to flush: %v", err)
		return
	}

	log.Printf("Sent toolResult for tool: %s", payload.Tool)
}