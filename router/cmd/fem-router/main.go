package main

import (
	"bufio"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"encoding/pem"
	"flag"
	"log"
	"math/big"
	"net"
	"time"
)

func main() {
	// Parse command line flags
	listenAddr := flag.String("listen", ":4433", "Address to listen on")
	flag.Parse()

	// Generate self-signed certificate
	cert, err := generateSelfSignedCert()
	if err != nil {
		log.Fatalf("Failed to generate certificate: %v", err)
	}

	// Create TLS configuration
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
	}

	// Start TLS listener
	listener, err := tls.Listen("tcp", *listenAddr, tlsConfig)
	if err != nil {
		log.Fatalf("Failed to listen on %s: %v", *listenAddr, err)
	}
	defer listener.Close()

	log.Printf("fem-router listening on %s", *listenAddr)

	// Accept connections
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Failed to accept connection: %v", err)
			continue
		}

		// Handle each connection in a goroutine
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	scanner := bufio.NewScanner(conn)
	writer := bufio.NewWriter(conn)

	for scanner.Scan() {
		line := scanner.Bytes()

		// Try to parse as JSON to validate
		var jsonData map[string]interface{}
		if err := json.Unmarshal(line, &jsonData); err != nil {
			log.Printf("Invalid JSON received: %v", err)
			continue
		}

		// Echo the line back
		if _, err := writer.Write(line); err != nil {
			log.Printf("Failed to write response: %v", err)
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

		log.Printf("Echoed JSON: %s", string(line))
	}

	if err := scanner.Err(); err != nil {
		log.Printf("Scanner error: %v", err)
	}
}

func generateSelfSignedCert() (tls.Certificate, error) {
	// Generate RSA key
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return tls.Certificate{}, err
	}

	// Create certificate template
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"FEM Router"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IPAddresses:           []net.IP{net.IPv4(127, 0, 0, 1)},
		DNSNames:              []string{"localhost"},
	}

	// Generate certificate
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		return tls.Certificate{}, err
	}

	// Encode certificate and key
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)})

	// Create TLS certificate
	return tls.X509KeyPair(certPEM, keyPEM)
}