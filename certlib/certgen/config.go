package certgen

import (
	"crypto"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"fmt"
	"math/big"
	"net"
	"strings"
	"time"

	"git.wntrmute.dev/kyle/goutils/lib"
)

type KeySpec struct {
	Algorithm string `yaml:"algorithm"`
	Size      int    `yaml:"size"`
}

func (ks KeySpec) String() string {
	if strings.ToLower(ks.Algorithm) == nameEd25519 {
		return nameEd25519
	}

	return fmt.Sprintf("%s-%d", ks.Algorithm, ks.Size)
}

func (ks KeySpec) Generate() (crypto.PublicKey, crypto.PrivateKey, error) {
	switch strings.ToLower(ks.Algorithm) {
	case "rsa":
		return GenerateKey(x509.RSA, ks.Size)
	case "ecdsa":
		return GenerateKey(x509.ECDSA, ks.Size)
	case nameEd25519:
		return GenerateKey(x509.Ed25519, 0)
	default:
		return nil, nil, fmt.Errorf("unknown key algorithm: %s", ks.Algorithm)
	}
}

func (ks KeySpec) SigningAlgorithm() (x509.SignatureAlgorithm, error) {
	switch strings.ToLower(ks.Algorithm) {
	case "rsa":
		return x509.SHA512WithRSAPSS, nil
	case "ecdsa":
		return x509.ECDSAWithSHA512, nil
	case nameEd25519:
		return x509.PureEd25519, nil
	default:
		return 0, fmt.Errorf("unknown key algorithm: %s", ks.Algorithm)
	}
}

type Subject struct {
	CommonName         string   `yaml:"common_name"`
	Country            string   `yaml:"country"`
	Locality           string   `yaml:"locality"`
	Province           string   `yaml:"province"`
	Organization       string   `yaml:"organization"`
	OrganizationalUnit string   `yaml:"organizational_unit"`
	Email              []string `yaml:"email"`
	DNSNames           []string `yaml:"dns"`
	IPAddresses        []string `yaml:"ips"`
}

type CertificateRequest struct {
	KeySpec KeySpec `yaml:"key"`
	Subject Subject `yaml:"subject"`
	Profile Profile `yaml:"profile"`
}

func (cs CertificateRequest) Request(priv crypto.PrivateKey) (*x509.CertificateRequest, error) {
	subject := pkix.Name{}
	subject.CommonName = cs.Subject.CommonName
	subject.Country = []string{cs.Subject.Country}
	subject.Locality = []string{cs.Subject.Locality}
	subject.Province = []string{cs.Subject.Province}
	subject.Organization = []string{cs.Subject.Organization}
	subject.OrganizationalUnit = []string{cs.Subject.OrganizationalUnit}

	ipAddresses := make([]net.IP, 0, len(cs.Subject.IPAddresses))
	for i, ip := range cs.Subject.IPAddresses {
		ipAddresses = append(ipAddresses, net.ParseIP(ip))
		if ipAddresses[i] == nil {
			return nil, fmt.Errorf("invalid IP address: %s", ip)
		}
	}

	req := &x509.CertificateRequest{
		PublicKeyAlgorithm: 0,
		PublicKey:          getPublic(priv),
		Subject:            subject,
		EmailAddresses:     cs.Subject.Email,
		DNSNames:           cs.Subject.DNSNames,
		IPAddresses:        ipAddresses,
	}

	reqBytes, err := x509.CreateCertificateRequest(rand.Reader, req, priv)
	if err != nil {
		return nil, fmt.Errorf("failed to create certificate request: %w", err)
	}

	req, err = x509.ParseCertificateRequest(reqBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse certificate request: %w", err)
	}

	return req, nil
}

func (cs CertificateRequest) Generate() (crypto.PrivateKey, *x509.CertificateRequest, error) {
	_, priv, err := cs.KeySpec.Generate()
	if err != nil {
		return nil, nil, err
	}

	req, err := cs.Request(priv)
	if err != nil {
		return nil, nil, err
	}

	return priv, req, nil
}

type Profile struct {
	IsCA         bool     `yaml:"is_ca"`
	PathLen      int      `yaml:"path_len"`
	KeyUse       []string `yaml:"key_uses"`
	ExtKeyUsages []string `yaml:"ext_key_usages"`
	Expiry       string   `yaml:"expiry"`
}

func (p Profile) templateFromRequest(req *x509.CertificateRequest) (*x509.Certificate, error) {
	serial, err := SerialNumber()
	if err != nil {
		return nil, fmt.Errorf("failed to generate serial number: %w", err)
	}

	expiry, err := lib.ParseDuration(p.Expiry)
	if err != nil {
		return nil, fmt.Errorf("parsing expiry: %w", err)
	}

	certTemplate := &x509.Certificate{
		SignatureAlgorithm:    req.SignatureAlgorithm,
		PublicKeyAlgorithm:    req.PublicKeyAlgorithm,
		PublicKey:             req.PublicKey,
		SerialNumber:          serial,
		Subject:               req.Subject,
		NotBefore:             time.Now().Add(-1 * time.Hour),
		NotAfter:              time.Now().Add(expiry),
		BasicConstraintsValid: true,
		IsCA:                  p.IsCA,
		MaxPathLen:            p.PathLen,
		DNSNames:              req.DNSNames,
		IPAddresses:           req.IPAddresses,
	}

	for _, sku := range p.KeyUse {
		ku, ok := keyUsageStrings[sku]
		if !ok {
			return nil, fmt.Errorf("invalid key usage: %s", p.KeyUse)
		}

		certTemplate.KeyUsage |= ku
	}

	for _, extKeyUsage := range p.ExtKeyUsages {
		eku, ok := extKeyUsageStrings[extKeyUsage]
		if !ok {
			return nil, fmt.Errorf("invalid extended key usage: %s", extKeyUsage)
		}
		certTemplate.ExtKeyUsage = append(certTemplate.ExtKeyUsage, eku)
	}

	return certTemplate, nil
}

func (p Profile) SignRequest(
	parent *x509.Certificate,
	req *x509.CertificateRequest,
	priv crypto.PrivateKey,
) (*x509.Certificate, error) {
	tpl, err := p.templateFromRequest(req)
	if err != nil {
		return nil, fmt.Errorf("failed to create certificate template: %w", err)
	}

	certBytes, err := x509.CreateCertificate(rand.Reader, tpl, parent, req.PublicKey, priv)
	if err != nil {
		return nil, fmt.Errorf("failed to create certificate: %w", err)
	}

	cert, err := x509.ParseCertificate(certBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse certificate: %w", err)
	}

	return cert, nil
}

func (p Profile) SelfSign(req *x509.CertificateRequest, priv crypto.PrivateKey) (*x509.Certificate, error) {
	certTemplate, err := p.templateFromRequest(req)
	if err != nil {
		return nil, fmt.Errorf("failed to create certificate template: %w", err)
	}

	return p.SignRequest(certTemplate, req, priv)
}

func SerialNumber() (*big.Int, error) {
	serialNumberBytes := make([]byte, 20)
	_, err := rand.Read(serialNumberBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to generate serial number: %w", err)
	}
	return new(big.Int).SetBytes(serialNumberBytes), nil
}

// GenerateSelfSigned generates a self-signed certificate using the given certificate request.
func GenerateSelfSigned(creq *CertificateRequest) (*x509.Certificate, crypto.PrivateKey, error) {
	priv, req, err := creq.Generate()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate certificate request: %w", err)
	}

	cert, err := creq.Profile.SelfSign(req, priv)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to self-sign certificate: %w", err)
	}

	return cert, priv, nil
}
