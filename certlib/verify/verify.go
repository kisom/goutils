package verify

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io"

	"git.wntrmute.dev/kyle/goutils/certlib/revoke"
	"git.wntrmute.dev/kyle/goutils/lib"
)

func bundleIntermediates(w io.Writer, chain []*x509.Certificate, pool *x509.CertPool, verbose bool) *x509.CertPool {
	for _, intermediate := range chain[1:] {
		if verbose {
			fmt.Fprintf(w, "[+] adding intermediate with SKI %x\n", intermediate.SubjectKeyId)
		}
		pool.AddCert(intermediate)
	}

	return pool
}

type Opts struct {
	Verbose            bool
	Config             *tls.Config
	Intermediates      *x509.CertPool
	ForceIntermediates bool
	CheckRevocation    bool
	KeyUsages          []x509.ExtKeyUsage
}

type verifyResult struct {
	chain []*x509.Certificate
	roots *x509.CertPool
	ints  *x509.CertPool
}

func prepareVerification(w io.Writer, target string, opts *Opts) (*verifyResult, error) {
	var (
		roots, ints *x509.CertPool
		err         error
	)

	if opts == nil {
		opts = &Opts{
			Config:             lib.StrictBaselineTLSConfig(),
			ForceIntermediates: false,
		}
	}

	if opts.Config.RootCAs == nil {
		roots, err = x509.SystemCertPool()
		if err != nil {
			return nil, fmt.Errorf("couldn't load system cert pool: %w", err)
		}

		opts.Config.RootCAs = roots
	}

	if opts.Intermediates == nil {
		ints = x509.NewCertPool()
	} else {
		ints = opts.Intermediates.Clone()
	}

	roots = opts.Config.RootCAs.Clone()

	chain, err := lib.GetCertificateChain(target, opts.Config)
	if err != nil {
		return nil, fmt.Errorf("fetching certificate chain: %w", err)
	}

	if opts.Verbose {
		fmt.Fprintf(w, "[+] %s has %d certificates\n", target, len(chain))
	}

	if len(chain) > 1 && opts.ForceIntermediates {
		ints = bundleIntermediates(w, chain, ints, opts.Verbose)
	}

	return &verifyResult{
		chain: chain,
		roots: roots,
		ints:  ints,
	}, nil
}

// Chain fetches the certificate chain for a target and verifies it.
func Chain(w io.Writer, target string, opts *Opts) ([]*x509.Certificate, error) {
	result, err := prepareVerification(w, target, opts)
	if err != nil {
		return nil, fmt.Errorf("certificate verification failed: %w", err)
	}

	chains, err := CertWith(result.chain[0], result.roots, result.ints, opts.CheckRevocation, opts.KeyUsages...)
	if err != nil {
		return nil, fmt.Errorf("certificate verification failed: %w", err)
	}

	return chains, nil
}

// CertWith verifies a certificate against a set of roots and intermediates.
func CertWith(
	cert *x509.Certificate,
	roots, ints *x509.CertPool,
	checkRevocation bool,
	keyUses ...x509.ExtKeyUsage,
) ([]*x509.Certificate, error) {
	if len(keyUses) == 0 {
		keyUses = []x509.ExtKeyUsage{x509.ExtKeyUsageAny}
	}

	opts := x509.VerifyOptions{
		Intermediates: ints,
		Roots:         roots,
		KeyUsages:     keyUses,
	}

	chains, err := cert.Verify(opts)
	if err != nil {
		return nil, err
	}

	if checkRevocation {
		revoked, ok := revoke.VerifyCertificate(cert)
		if !ok {
			return nil, errors.New("failed to check certificate revocation status")
		}

		if revoked {
			return nil, errors.New("certificate is revoked")
		}
	}

	if len(chains) == 0 {
		return nil, errors.New("no valid certificate chain found")
	}

	return chains[0], nil
}
