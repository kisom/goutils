package verify

import (
	"crypto/x509"
	"fmt"
	"time"

	"git.wntrmute.dev/kyle/goutils/certlib/dump"
)

const DefaultLeeway = 2160 * time.Hour // three months

type CertCheck struct {
	Cert   *x509.Certificate
	leeway time.Duration
}

func NewCertCheck(cert *x509.Certificate, leeway time.Duration) *CertCheck {
	return &CertCheck{
		Cert:   cert,
		leeway: leeway,
	}
}

func (c CertCheck) Expiry() time.Duration {
	return time.Until(c.Cert.NotAfter)
}

func (c CertCheck) IsExpiring(leeway time.Duration) bool {
	return c.Expiry() < leeway
}

// Err returns nil if the certificate is not expiring within the leeway period.
func (c CertCheck) Err() error {
	if !c.IsExpiring(c.leeway) {
		return nil
	}

	return fmt.Errorf("%s expires in %s", dump.DisplayName(c.Cert.Subject), c.Expiry())
}

func (c CertCheck) Name() string {
	return fmt.Sprintf("%s/SN=%s", dump.DisplayName(c.Cert.Subject),
		c.Cert.SerialNumber)
}

func (c CertCheck) String() string {
	return fmt.Sprintf("%s expires on %s (in %s)\n", c.Name(), c.Cert.NotAfter, c.Expiry())
}
