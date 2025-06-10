package broker

import (
	"crypto/ed25519"
	"encoding/json"
	"fmt"
	"net"
	"sync"
	"time"

	protocol "github.com/fep-fem/protocol"
	"github.com/sirupsen/logrus"
)

// Broker represents a FEM broker node
type Broker struct {
	id           string
	transport    *protocol.Transport
	agents       map[string]*AgentInfo
	routers      map[string]*RouterInfo
	capabilities *protocol.CapabilityManager
	logger       *logrus.Logger
	mu           sync.RWMutex
}

// AgentInfo stores information about registered agents
type AgentInfo struct {
	ID           string
	PublicKey    ed25519.PublicKey
	Capabilities []string
	Metadata     map[string]interface{}
	RegisteredAt time.Time
	LastSeen     time.Time
}

// RouterInfo stores information about connected routers
type RouterInfo struct {
	ID           string
	Endpoint     string
	PublicKey    ed25519.PublicKey
	Capabilities []string
	ConnectedAt  time.Time
}

// NewBroker creates a new broker instance
func NewBroker(id string, privateKey ed25519.PrivateKey) (*Broker, error) {
	transport, err := protocol.NewTransport(privateKey)
	if err != nil {
		return nil, err
	}

	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	broker := &Broker{
		id:           id,
		transport:    transport,
		agents:       make(map[string]*AgentInfo),
		routers:      make(map[string]*RouterInfo),
		capabilities: protocol.NewCapabilityManager([]byte("broker-secret")), // In production, use secure key
		logger:       logger,
	}

	// Register envelope handlers
	broker.setupHandlers()

	return broker, nil
}

// setupHandlers registers handlers for different envelope types
func (b *Broker) setupHandlers() {
	b.transport.RegisterHandler(protocol.EnvelopeRegisterAgent, b.handleRegisterAgent)
	b.transport.RegisterHandler(protocol.EnvelopeRegisterBroker, b.handleRegisterBroker)
	b.transport.RegisterHandler(protocol.EnvelopeEmitEvent, b.handleEmitEvent)
	b.transport.RegisterHandler(protocol.EnvelopeRenderInstruction, b.handleRenderInstruction)
	b.transport.RegisterHandler(protocol.EnvelopeToolCall, b.handleToolCall)
	b.transport.RegisterHandler(protocol.EnvelopeToolResult, b.handleToolResult)
	b.transport.RegisterHandler(protocol.EnvelopeRevoke, b.handleRevoke)
}

// Start starts the broker on the specified address
func (b *Broker) Start(address string) error {
	b.logger.WithField("address", address).Info("Starting FEM broker")
	return b.transport.Listen(address)
}

// handleRegisterAgent handles agent registration
func (b *Broker) handleRegisterAgent(envelope *protocol.Envelope, conn net.Conn) error {
	var body protocol.RegisterAgentBody
	if err := json.Unmarshal(envelope.Body, &body); err != nil {
		return err
	}

	// Decode public key
	pubKey, err := protocol.DecodePublicKey(body.PubKey)
	if err != nil {
		return err
	}

	// Verify signature
	if err := envelope.Verify(pubKey); err != nil {
		return err
	}

	// Register agent
	b.mu.Lock()
	b.agents[envelope.Agent] = &AgentInfo{
		ID:           envelope.Agent,
		PublicKey:    pubKey,
		Capabilities: body.Capabilities,
		Metadata:     body.Metadata,
		RegisteredAt: time.Now(),
		LastSeen:     time.Now(),
	}
	b.mu.Unlock()

	b.logger.WithFields(logrus.Fields{
		"agent":        envelope.Agent,
		"capabilities": body.Capabilities,
	}).Info("Agent registered")

	// Send acknowledgment
	ack := protocol.NewEnvelope(protocol.EnvelopeEmitEvent, b.id)
	ack.Body, _ = json.Marshal(protocol.EmitEventBody{
		Event: "agent.registered",
		Payload: map[string]interface{}{
			"agent": envelope.Agent,
		},
	})

	stream := protocol.NewStream(conn)
	return stream.WriteEnvelope(ack)
}

// handleRegisterBroker handles broker/router registration
func (b *Broker) handleRegisterBroker(envelope *protocol.Envelope, conn net.Conn) error {
	var body protocol.RegisterBrokerBody
	if err := json.Unmarshal(envelope.Body, &body); err != nil {
		return err
	}

	// Decode public key
	pubKey, err := protocol.DecodePublicKey(body.PubKey)
	if err != nil {
		return err
	}

	// Verify signature
	if err := envelope.Verify(pubKey); err != nil {
		return err
	}

	// Register router
	b.mu.Lock()
	b.routers[body.BrokerID] = &RouterInfo{
		ID:           body.BrokerID,
		Endpoint:     body.Endpoint,
		PublicKey:    pubKey,
		Capabilities: body.Capabilities,
		ConnectedAt:  time.Now(),
	}
	b.mu.Unlock()

	b.logger.WithFields(logrus.Fields{
		"router":   body.BrokerID,
		"endpoint": body.Endpoint,
	}).Info("Router registered")

	return nil
}

// handleEmitEvent handles event emissions
func (b *Broker) handleEmitEvent(envelope *protocol.Envelope, conn net.Conn) error {
	var body protocol.EmitEventBody
	if err := json.Unmarshal(envelope.Body, &body); err != nil {
		return err
	}

	b.logger.WithFields(logrus.Fields{
		"agent": envelope.Agent,
		"event": body.Event,
	}).Debug("Event received")

	// Broadcast event to all connected agents
	b.broadcastEvent(envelope)

	return nil
}

// handleToolCall handles tool call requests
func (b *Broker) handleToolCall(envelope *protocol.Envelope, conn net.Conn) error {
	var body protocol.ToolCallBody
	if err := json.Unmarshal(envelope.Body, &body); err != nil {
		return err
	}

	b.logger.WithFields(logrus.Fields{
		"agent":     envelope.Agent,
		"tool":      body.Tool,
		"requestId": body.RequestID,
	}).Debug("Tool call received")

	// Route to appropriate body/executor
	// In a full implementation, this would route to registered bodies
	
	// For now, echo back an error
	result := protocol.NewEnvelope(protocol.EnvelopeToolResult, b.id)
	result.Body, _ = json.Marshal(protocol.ToolResultBody{
		RequestID: body.RequestID,
		Success:   false,
		Error:     "No body registered for tool: " + body.Tool,
	})

	stream := protocol.NewStream(conn)
	return stream.WriteEnvelope(result)
}

// handleToolResult handles tool execution results
func (b *Broker) handleToolResult(envelope *protocol.Envelope, conn net.Conn) error {
	var body protocol.ToolResultBody
	if err := json.Unmarshal(envelope.Body, &body); err != nil {
		return err
	}

	b.logger.WithFields(logrus.Fields{
		"agent":     envelope.Agent,
		"requestId": body.RequestID,
		"success":   body.Success,
	}).Debug("Tool result received")

	// Route result back to requesting agent
	// In a full implementation, this would maintain request tracking

	return nil
}

// handleRenderInstruction handles rendering instructions
func (b *Broker) handleRenderInstruction(envelope *protocol.Envelope, conn net.Conn) error {
	var body protocol.RenderInstructionBody
	if err := json.Unmarshal(envelope.Body, &body); err != nil {
		return err
	}

	b.logger.WithFields(logrus.Fields{
		"agent":       envelope.Agent,
		"instruction": body.Instruction,
	}).Debug("Render instruction received")

	// Route to UI/rendering subsystem
	// In a full implementation, this would forward to UI handlers

	return nil
}

// handleRevoke handles revocation requests
func (b *Broker) handleRevoke(envelope *protocol.Envelope, conn net.Conn) error {
	var body protocol.RevokeBody
	if err := json.Unmarshal(envelope.Body, &body); err != nil {
		return err
	}

	b.logger.WithFields(logrus.Fields{
		"target": body.Target,
		"reason": body.Reason,
	}).Info("Revocation received")

	// Remove agent or router
	b.mu.Lock()
	delete(b.agents, body.Target)
	delete(b.routers, body.Target)
	b.mu.Unlock()

	// Propagate revocation
	b.broadcastRevocation(body.Target, body.Reason)

	return nil
}

// broadcastEvent broadcasts an event to all connected agents
func (b *Broker) broadcastEvent(envelope *protocol.Envelope) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	// In a full implementation, this would maintain active connections
	// and forward the envelope to all connected agents
}

// broadcastRevocation broadcasts a revocation to all nodes
func (b *Broker) broadcastRevocation(target, reason string) {
	revoke := protocol.NewEnvelope(protocol.EnvelopeRevoke, b.id)
	revoke.Body, _ = json.Marshal(protocol.RevokeBody{
		Target: target,
		Reason: reason,
	})

	// In a full implementation, this would forward to all connected nodes
}

// GetAgents returns a list of registered agents
func (b *Broker) GetAgents() []AgentInfo {
	b.mu.RLock()
	defer b.mu.RUnlock()

	agents := make([]AgentInfo, 0, len(b.agents))
	for _, agent := range b.agents {
		agents = append(agents, *agent)
	}
	return agents
}

// GetRouters returns a list of connected routers
func (b *Broker) GetRouters() []RouterInfo {
	b.mu.RLock()
	defer b.mu.RUnlock()

	routers := make([]RouterInfo, 0, len(b.routers))
	for _, router := range b.routers {
		routers = append(routers, *router)
	}
	return routers
}