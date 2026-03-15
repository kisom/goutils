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
