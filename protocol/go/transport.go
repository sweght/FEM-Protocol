package protocol

import (
	"bufio"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net"
	"sync"
	"time"
)

// Transport handles FEP protocol communication
type Transport struct {
	privateKey ed25519.PrivateKey
	publicKey  ed25519.PublicKey
	tlsConfig  *tls.Config
	handlers   map[EnvelopeType]EnvelopeHandler
	mu         sync.RWMutex
}

// EnvelopeHandler processes incoming envelopes
type EnvelopeHandler func(envelope *Envelope, conn net.Conn) error

// NewTransport creates a new FEP transport
func NewTransport(privateKey ed25519.PrivateKey) (*Transport, error) {
	if privateKey == nil {
		// Generate new key pair
		publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
		if err != nil {
			return nil, err
		}
		return &Transport{
			privateKey: privateKey,
			publicKey:  publicKey,
			handlers:   make(map[EnvelopeType]EnvelopeHandler),
		}, nil
	}

	return &Transport{
		privateKey: privateKey,
		publicKey:  privateKey.Public().(ed25519.PublicKey),
		handlers:   make(map[EnvelopeType]EnvelopeHandler),
	}, nil
}

// GenerateSelfSignedCert generates a self-signed certificate for TLS
func (t *Transport) GenerateSelfSignedCert() error {
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"FEM Node"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, t.publicKey, t.privateKey)
	if err != nil {
		return err
	}

	cert, err := x509.ParseCertificate(certDER)
	if err != nil {
		return err
	}

	t.tlsConfig = &tls.Config{
		Certificates: []tls.Certificate{
			{
				Certificate: [][]byte{certDER},
				PrivateKey:  t.privateKey,
				Leaf:        cert,
			},
		},
		MinVersion: tls.VersionTLS13,
	}

	return nil
}

// Listen starts listening for FEP connections
func (t *Transport) Listen(address string) error {
	if t.tlsConfig == nil {
		if err := t.GenerateSelfSignedCert(); err != nil {
			return err
		}
	}

	listener, err := tls.Listen("tcp", address, t.tlsConfig)
	if err != nil {
		return err
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}
		go t.handleConnection(conn)
	}
}

// handleConnection handles an incoming connection
func (t *Transport) handleConnection(conn net.Conn) {
	defer conn.Close()

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		var envelope Envelope
		if err := json.Unmarshal(scanner.Bytes(), &envelope); err != nil {
			continue
		}

		// Handle envelope
		t.mu.RLock()
		handler, exists := t.handlers[envelope.Type]
		t.mu.RUnlock()

		if exists {
			if err := handler(&envelope, conn); err != nil {
				// Log error
				continue
			}
		}
	}
}

// RegisterHandler registers a handler for an envelope type
func (t *Transport) RegisterHandler(envType EnvelopeType, handler EnvelopeHandler) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.handlers[envType] = handler
}

// Send sends an envelope to a remote endpoint
func (t *Transport) Send(endpoint string, envelope *Envelope) error {
	// Sign the envelope
	if err := envelope.Sign(t.privateKey); err != nil {
		return err
	}

	// Connect to endpoint
	conn, err := tls.Dial("tcp", endpoint, &tls.Config{
		InsecureSkipVerify: true, // In production, verify certificates
		MinVersion:         tls.VersionTLS13,
	})
	if err != nil {
		return err
	}
	defer conn.Close()

	// Send envelope
	data, err := json.Marshal(envelope)
	if err != nil {
		return err
	}

	_, err = conn.Write(append(data, '\n'))
	return err
}

// Client represents a FEP client connection
type Client struct {
	transport *Transport
	endpoint  string
	conn      net.Conn
	mu        sync.Mutex
}

// NewClient creates a new FEP client
func NewClient(endpoint string, privateKey ed25519.PrivateKey) (*Client, error) {
	transport, err := NewTransport(privateKey)
	if err != nil {
		return nil, err
	}

	return &Client{
		transport: transport,
		endpoint:  endpoint,
	}, nil
}

// Connect establishes a connection to the server
func (c *Client) Connect() error {
	conn, err := tls.Dial("tcp", c.endpoint, &tls.Config{
		InsecureSkipVerify: true,
		MinVersion:         tls.VersionTLS13,
	})
	if err != nil {
		return err
	}

	c.conn = conn
	return nil
}

// SendEnvelope sends an envelope to the server
func (c *Client) SendEnvelope(envelope *Envelope) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn == nil {
		return fmt.Errorf("not connected")
	}

	// Sign the envelope
	if err := envelope.Sign(c.transport.privateKey); err != nil {
		return err
	}

	// Send envelope
	data, err := json.Marshal(envelope)
	if err != nil {
		return err
	}

	_, err = c.conn.Write(append(data, '\n'))
	return err
}

// ReadEnvelope reads an envelope from the server
func (c *Client) ReadEnvelope() (*Envelope, error) {
	if c.conn == nil {
		return nil, fmt.Errorf("not connected")
	}

	reader := bufio.NewReader(c.conn)
	line, err := reader.ReadBytes('\n')
	if err != nil {
		return nil, err
	}

	var envelope Envelope
	if err := json.Unmarshal(line, &envelope); err != nil {
		return nil, err
	}

	return &envelope, nil
}

// Close closes the client connection
func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// Stream represents a bidirectional FEP stream
type Stream struct {
	reader *bufio.Reader
	writer io.Writer
	mu     sync.Mutex
}

// NewStream creates a new FEP stream
func NewStream(conn net.Conn) *Stream {
	return &Stream{
		reader: bufio.NewReader(conn),
		writer: conn,
	}
}

// ReadEnvelope reads an envelope from the stream
func (s *Stream) ReadEnvelope() (*Envelope, error) {
	line, err := s.reader.ReadBytes('\n')
	if err != nil {
		return nil, err
	}

	var envelope Envelope
	if err := json.Unmarshal(line, &envelope); err != nil {
		return nil, err
	}

	return &envelope, nil
}

// WriteEnvelope writes an envelope to the stream
func (s *Stream) WriteEnvelope(envelope *Envelope) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := json.Marshal(envelope)
	if err != nil {
		return err
	}

	_, err = s.writer.Write(append(data, '\n'))
	return err
}