package certgen

import (
	"slices"
	"testing"
)

func TestIsFQDN(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"example.com", true},
		{"sub.example.com", true},
		{"example.com.", true}, // trailing dot
		{"localhost", false},   // no dot
		{"", false},
		{"foo bar.com", false}, // space
		{"-bad.com", false},    // leading hyphen
		{"bad-.com", false},    // trailing hyphen
		{"a..b.com", false},    // empty label
	}

	for _, tt := range tests {
		got := isFQDN(tt.input)
		if got != tt.want {
			t.Errorf("isFQDN(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestRequestAddsFQDNAsDNSSAN(t *testing.T) {
	creq := &CertificateRequest{
		KeySpec: KeySpec{Algorithm: "ecdsa", Size: 256},
		Subject: Subject{
			CommonName:   "example.com",
			Organization: "Test Org",
		},
		Profile: Profile{
			Expiry: "1h",
		},
	}

	_, req, err := creq.Generate()
	if err != nil {
		t.Fatalf("Generate() error: %v", err)
	}

	if !slices.Contains(req.DNSNames, "example.com") {
		t.Errorf("expected DNS SAN to contain %q, got %v", "example.com", req.DNSNames)
	}
}

func TestRequestFQDNNotDuplicated(t *testing.T) {
	creq := &CertificateRequest{
		KeySpec: KeySpec{Algorithm: "ecdsa", Size: 256},
		Subject: Subject{
			CommonName:   "example.com",
			Organization: "Test Org",
			DNSNames:     []string{"example.com", "www.example.com"},
		},
		Profile: Profile{
			Expiry: "1h",
		},
	}

	_, req, err := creq.Generate()
	if err != nil {
		t.Fatalf("Generate() error: %v", err)
	}

	count := 0
	for _, name := range req.DNSNames {
		if name == "example.com" {
			count++
		}
	}

	if count != 1 {
		t.Errorf("expected exactly 1 occurrence of %q in DNS SANs, got %d: %v", "example.com", count, req.DNSNames)
	}
}

func TestProfileAIAFieldsInCertificate(t *testing.T) {
	caKey := KeySpec{Algorithm: "ecdsa", Size: 256}
	_, caPriv, err := caKey.Generate()
	if err != nil {
		t.Fatalf("generate CA key: %v", err)
	}

	caProfile := Profile{
		IsCA:    true,
		PathLen: 1,
		KeyUse:  []string{"cert sign", "crl sign"},
		Expiry:  "8760h",
	}

	caReq := &CertificateRequest{
		KeySpec: caKey,
		Subject: Subject{CommonName: "Test CA", Organization: "Test"},
		Profile: caProfile,
	}
	caCSR, err := caReq.Request(caPriv)
	if err != nil {
		t.Fatalf("generate CA CSR: %v", err)
	}
	caCert, err := caProfile.SelfSign(caCSR, caPriv)
	if err != nil {
		t.Fatalf("self-sign CA: %v", err)
	}

	leafProfile := Profile{
		KeyUse:                []string{"digital signature"},
		ExtKeyUsages:          []string{"server auth"},
		Expiry:                "24h",
		OCSPServer:            []string{"https://ocsp.example.com"},
		IssuingCertificateURL: []string{"https://pki.example.com/ca.pem"},
	}

	leafReq := &CertificateRequest{
		KeySpec: KeySpec{Algorithm: "ecdsa", Size: 256},
		Subject: Subject{CommonName: "leaf.example.com", Organization: "Test"},
		Profile: leafProfile,
	}
	_, leafCSR, err := leafReq.Generate()
	if err != nil {
		t.Fatalf("generate leaf CSR: %v", err)
	}

	leafCert, err := leafProfile.SignRequest(caCert, leafCSR, caPriv)
	if err != nil {
		t.Fatalf("sign leaf: %v", err)
	}

	if len(leafCert.OCSPServer) != 1 || leafCert.OCSPServer[0] != "https://ocsp.example.com" {
		t.Errorf("OCSPServer = %v, want [https://ocsp.example.com]", leafCert.OCSPServer)
	}
	if len(leafCert.IssuingCertificateURL) != 1 || leafCert.IssuingCertificateURL[0] != "https://pki.example.com/ca.pem" {
		t.Errorf("IssuingCertificateURL = %v, want [https://pki.example.com/ca.pem]", leafCert.IssuingCertificateURL)
	}
}

func TestProfileWithoutAIAOmitsExtension(t *testing.T) {
	profile := Profile{
		KeyUse:       []string{"digital signature"},
		ExtKeyUsages: []string{"server auth"},
		Expiry:       "24h",
	}

	creq := &CertificateRequest{
		KeySpec: KeySpec{Algorithm: "ecdsa", Size: 256},
		Subject: Subject{CommonName: "noaia.example.com", Organization: "Test"},
		Profile: profile,
	}
	cert, _, err := GenerateSelfSigned(creq)
	if err != nil {
		t.Fatalf("generate: %v", err)
	}

	if len(cert.OCSPServer) != 0 {
		t.Errorf("OCSPServer = %v, want empty", cert.OCSPServer)
	}
	if len(cert.IssuingCertificateURL) != 0 {
		t.Errorf("IssuingCertificateURL = %v, want empty", cert.IssuingCertificateURL)
	}
}

func TestRequestNonFQDNCommonNameNotAdded(t *testing.T) {
	creq := &CertificateRequest{
		KeySpec: KeySpec{Algorithm: "ecdsa", Size: 256},
		Subject: Subject{
			CommonName:   "localhost",
			Organization: "Test Org",
		},
		Profile: Profile{
			Expiry: "1h",
		},
	}

	_, req, err := creq.Generate()
	if err != nil {
		t.Fatalf("Generate() error: %v", err)
	}

	if slices.Contains(req.DNSNames, "localhost") {
		t.Errorf("expected DNS SANs to not contain %q, got %v", "localhost", req.DNSNames)
	}
}
