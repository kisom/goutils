package certgen

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewTestCA(t *testing.T) {
	ca, err := NewTestCA()
	if err != nil {
		t.Fatalf("NewTestCA: %v", err)
	}

	cert := ca.Certificate()
	if !cert.IsCA {
		t.Fatal("expected CA certificate")
	}
	if cert.Subject.CommonName != "Test CA" {
		t.Fatalf("got CN %q, want %q", cert.Subject.CommonName, "Test CA")
	}
}

func TestCertificatePEMRoundtrip(t *testing.T) {
	ca, err := NewTestCA()
	if err != nil {
		t.Fatalf("NewTestCA: %v", err)
	}

	pemBytes := ca.CertificatePEM()
	block, _ := pem.Decode(pemBytes)
	if block == nil {
		t.Fatal("failed to decode PEM")
	}
	if block.Type != "CERTIFICATE" {
		t.Fatalf("got PEM type %q, want CERTIFICATE", block.Type)
	}

	parsed, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		t.Fatalf("parse certificate: %v", err)
	}
	if !parsed.Equal(ca.Certificate()) {
		t.Fatal("parsed certificate does not match original")
	}
}

func TestCertPool(t *testing.T) {
	ca, err := NewTestCA()
	if err != nil {
		t.Fatalf("NewTestCA: %v", err)
	}

	pool := ca.CertPool()

	// Verify the CA cert validates against its own pool.
	chains, err := ca.Certificate().Verify(x509.VerifyOptions{
		Roots: pool,
	})
	if err != nil {
		t.Fatalf("verify CA cert against its own pool: %v", err)
	}
	if len(chains) == 0 {
		t.Fatal("expected at least one chain")
	}
}

func TestIssue(t *testing.T) {
	ca, err := NewTestCA()
	if err != nil {
		t.Fatalf("NewTestCA: %v", err)
	}

	dnsNames := []string{"example.test", "www.example.test"}
	ips := []net.IP{net.IPv4(10, 0, 0, 1)}

	key, cert, err := ca.Issue(dnsNames, ips)
	if err != nil {
		t.Fatalf("Issue: %v", err)
	}
	if key == nil {
		t.Fatal("expected non-nil key")
	}
	if cert.IsCA {
		t.Fatal("leaf cert should not be CA")
	}
	if cert.Subject.CommonName != "example.test" {
		t.Fatalf("got CN %q, want %q", cert.Subject.CommonName, "example.test")
	}

	// Verify the leaf cert chains to the CA.
	_, err = cert.Verify(x509.VerifyOptions{
		Roots:     ca.CertPool(),
		DNSName:   "example.test",
		KeyUsages: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	})
	if err != nil {
		t.Fatalf("verify leaf cert: %v", err)
	}
}

func TestIssueServerTLS(t *testing.T) {
	ca, err := NewTestCA()
	if err != nil {
		t.Fatalf("NewTestCA: %v", err)
	}

	key, cert, err := ca.IssueServer()
	if err != nil {
		t.Fatalf("IssueServer: %v", err)
	}

	// Start a TLS server with the issued cert.
	serverCfg := ca.ServerTLSConfig(key, cert)
	srv := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	srv.TLS = &serverCfg
	srv.StartTLS()
	defer srv.Close()

	// Create a client that verifies the server cert against the CA.
	clientCfg := ca.TLSConfig()
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &clientCfg,
		},
	}

	resp, err := client.Get(srv.URL)
	if err != nil {
		t.Fatalf("GET: %v", err)
	}
	resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("got status %d, want 200", resp.StatusCode)
	}
}

func TestTLSConfigReturnsByValue(t *testing.T) {
	ca, err := NewTestCA()
	if err != nil {
		t.Fatalf("NewTestCA: %v", err)
	}

	cfg1 := ca.TLSConfig()
	cfg2 := ca.TLSConfig()

	// Modifying one should not affect the other.
	cfg1.ServerName = "modified"
	if cfg2.ServerName == "modified" {
		t.Fatal("TLSConfig should return independent values")
	}
}

func TestTLSConfigEnforcesTLS13(t *testing.T) {
	ca, err := NewTestCA()
	if err != nil {
		t.Fatalf("NewTestCA: %v", err)
	}

	cfg := ca.TLSConfig()
	if cfg.MinVersion != tls.VersionTLS13 {
		t.Fatalf("got MinVersion %d, want TLS 1.3 (%d)", cfg.MinVersion, tls.VersionTLS13)
	}
}

func TestMustTestCA(t *testing.T) {
	// Should not panic.
	ca := MustTestCA()
	if ca.Certificate() == nil {
		t.Fatal("expected non-nil certificate")
	}
}

func TestPrivateKeyPEM(t *testing.T) {
	ca, err := NewTestCA()
	if err != nil {
		t.Fatalf("NewTestCA: %v", err)
	}

	key, _, err := ca.IssueServer()
	if err != nil {
		t.Fatalf("IssueServer: %v", err)
	}

	pemBytes, err := PrivateKeyPEM(key)
	if err != nil {
		t.Fatalf("PrivateKeyPEM: %v", err)
	}

	block, _ := pem.Decode(pemBytes)
	if block == nil {
		t.Fatal("failed to decode PEM")
	}
	if block.Type != "PRIVATE KEY" {
		t.Fatalf("got PEM type %q, want PRIVATE KEY", block.Type)
	}
}

func TestUntrustedCAFails(t *testing.T) {
	ca1 := MustTestCA()
	ca2 := MustTestCA()

	// Issue a cert from ca1, try to verify against ca2's pool.
	_, cert, err := ca1.IssueServer()
	if err != nil {
		t.Fatalf("IssueServer: %v", err)
	}

	_, err = cert.Verify(x509.VerifyOptions{
		Roots:   ca2.CertPool(),
		DNSName: "localhost",
	})
	if err == nil {
		t.Fatal("expected verification to fail with wrong CA")
	}
}
