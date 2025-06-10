package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"encoding/pem"
	"flag"
	"io"
	"log"
	"math/big"
	"net"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/fep-fem/protocol"
)

// Broker represents the FEM broker server
type Broker struct {
	agents    map[string]*Agent
	mu        sync.RWMutex
	tlsConfig *tls.Config
}

// Agent represents a registered agent
type Agent struct {
	ID           string
	Capabilities []string
	Endpoint     string
	RegisteredAt time.Time
}

func main() {
	var listen string
	flag.StringVar(&listen, "listen", ":4433", "Address to listen on")
	flag.Parse()

	broker := NewBroker()

	// Generate self-signed certificate
	cert, err := generateSelfSignedCert()
	if err != nil {
		log.Fatalf("Failed to generate certificate: %v", err)
	}

	broker.tlsConfig = &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS13,
	}

	// Create HTTPS server
	server := &http.Server{
		Addr:      listen,
		Handler:   broker,
		TLSConfig: broker.tlsConfig,
	}

	log.Printf("FEM Broker starting on %s", listen)
	log.Fatal(server.ListenAndServeTLS("", ""))
}

// NewBroker creates a new broker instance
func NewBroker() *Broker {
	return &Broker{
		agents: make(map[string]*Agent),
	}
}

// ServeHTTP implements the http.Handler interface
func (b *Broker) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Read body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Parse envelope
	envelope, err := protocol.ParseEnvelope(body)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid envelope: %v", err), http.StatusBadRequest)
		return
	}

	// Log the received envelope
	log.Printf("Received %s envelope from %s", envelope.Type, envelope.Agent)

	// Process based on envelope type
	switch envelope.Type {
	case protocol.EnvelopeRegisterAgent:
		b.handleRegisterAgent(w, &envelope)
	case protocol.EnvelopeRegisterBroker:
		b.handleRegisterBroker(w, &envelope)
	case protocol.EnvelopeEmitEvent:
		b.handleEmitEvent(w, &envelope)
	case protocol.EnvelopeRenderInstruction:
		b.handleRenderInstruction(w, &envelope)
	case protocol.EnvelopeToolCall:
		b.handleToolCall(w, &envelope)
	case protocol.EnvelopeToolResult:
		b.handleToolResult(w, &envelope)
	case protocol.EnvelopeRevoke:
		b.handleRevoke(w, &envelope)
	default:
		http.Error(w, "Unknown envelope type", http.StatusBadRequest)
		return
	}
}

// handleRegisterAgent processes agent registration
func (b *Broker) handleRegisterAgent(w http.ResponseWriter, env *protocol.Envelope) {
	var body struct {
		Capabilities []string `json:"capabilities"`
		Endpoint     string   `json:"endpoint"`
	}

	if err := json.Unmarshal(env.Body, &body); err != nil {
		http.Error(w, "Invalid body", http.StatusBadRequest)
		return
	}

	b.mu.Lock()
	b.agents[env.Headers.Agent] = &Agent{
		ID:           env.Headers.Agent,
		Capabilities: body.Capabilities,
		Endpoint:     body.Endpoint,
		RegisteredAt: time.Now(),
	}
	b.mu.Unlock()

	log.Printf("Registered agent %s with capabilities %v", env.Headers.Agent, body.Capabilities)

	response := map[string]interface{}{
		"status": "registered",
		"agent":  env.Headers.Agent,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleRegisterBroker processes broker registration
func (b *Broker) handleRegisterBroker(w http.ResponseWriter, env *protocol.Envelope) {
	var body struct {
		Endpoint   string                 `json:"endpoint"`
		Embodiment map[string]interface{} `json:"embodiment,omitempty"`
	}

	if err := json.Unmarshal(env.Body, &body); err != nil {
		http.Error(w, "Invalid body", http.StatusBadRequest)
		return
	}

	log.Printf("Broker registration from %s at %s", env.Headers.Agent, body.Endpoint)

	response := map[string]interface{}{
		"status": "registered",
		"broker": env.Headers.Agent,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleEmitEvent processes event emissions
func (b *Broker) handleEmitEvent(w http.ResponseWriter, env *protocol.Envelope) {
	var body struct {
		EventType string                 `json:"eventType"`
		Data      map[string]interface{} `json:"data"`
	}

	if err := json.Unmarshal(env.Body, &body); err != nil {
		http.Error(w, "Invalid body", http.StatusBadRequest)
		return
	}

	log.Printf("Event %s from %s: %v", body.EventType, env.Headers.Agent, body.Data)

	// In a real implementation, this would fan out to subscribers
	response := map[string]interface{}{
		"status": "emitted",
		"event":  body.EventType,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleRenderInstruction processes render instructions
func (b *Broker) handleRenderInstruction(w http.ResponseWriter, env *protocol.Envelope) {
	var body struct {
		Instruction string                 `json:"instruction"`
		Context     map[string]interface{} `json:"context,omitempty"`
	}

	if err := json.Unmarshal(env.Body, &body); err != nil {
		http.Error(w, "Invalid body", http.StatusBadRequest)
		return
	}

	log.Printf("Render instruction from %s: %s", env.Headers.Agent, body.Instruction)

	response := map[string]interface{}{
		"status": "rendered",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleToolCall processes tool calls
func (b *Broker) handleToolCall(w http.ResponseWriter, env *protocol.Envelope) {
	var body struct {
		Tool       string                 `json:"tool"`
		Parameters map[string]interface{} `json:"parameters"`
	}

	if err := json.Unmarshal(env.Body, &body); err != nil {
		http.Error(w, "Invalid body", http.StatusBadRequest)
		return
	}

	log.Printf("Tool call %s from %s", body.Tool, env.Headers.Agent)

	// In a real implementation, this would route to the appropriate tool handler
	response := map[string]interface{}{
		"status": "processing",
		"tool":   body.Tool,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleToolResult processes tool results
func (b *Broker) handleToolResult(w http.ResponseWriter, env *protocol.Envelope) {
	var body struct {
		Tool   string      `json:"tool"`
		Result interface{} `json:"result"`
		Error  string      `json:"error,omitempty"`
	}

	if err := json.Unmarshal(env.Body, &body); err != nil {
		http.Error(w, "Invalid body", http.StatusBadRequest)
		return
	}

	log.Printf("Tool result for %s from %s", body.Tool, env.Headers.Agent)

	response := map[string]interface{}{
		"status": "received",
		"tool":   body.Tool,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleRevoke processes revocation
func (b *Broker) handleRevoke(w http.ResponseWriter, env *protocol.Envelope) {
	var body struct {
		Target string `json:"target"`
		Reason string `json:"reason"`
	}

	if err := json.Unmarshal(env.Body, &body); err != nil {
		http.Error(w, "Invalid body", http.StatusBadRequest)
		return
	}

	b.mu.Lock()
	delete(b.agents, body.Target)
	b.mu.Unlock()

	log.Printf("Revoked %s for reason: %s", body.Target, body.Reason)

	response := map[string]interface{}{
		"status": "revoked",
		"target": body.Target,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// generateSelfSignedCert generates a self-signed certificate for TLS
func generateSelfSignedCert() (tls.Certificate, error) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return tls.Certificate{}, err
	}

	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"FEM Broker"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IPAddresses:           []net.IP{net.IPv4(127, 0, 0, 1)},
		DNSNames:              []string{"localhost"},
	}

	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		return tls.Certificate{}, err
	}

	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)})
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})

	return tls.X509KeyPair(certPEM, keyPEM)
}

func init() {
	// Set up logging
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	log.SetOutput(os.Stdout)
}