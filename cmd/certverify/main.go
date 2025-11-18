package main

import (
	"crypto/x509"
	"flag"
	"fmt"
	"os"
	"time"

	"git.wntrmute.dev/kyle/goutils/certlib"
	"git.wntrmute.dev/kyle/goutils/certlib/revoke"
	"git.wntrmute.dev/kyle/goutils/lib"
)

func printRevocation(cert *x509.Certificate) {
	remaining := time.Until(cert.NotAfter)
	fmt.Printf("certificate expires in %s.\n", lib.Duration(remaining))

	revoked, ok := revoke.VerifyCertificate(cert)
	if !ok {
		fmt.Fprintf(os.Stderr, "[!] the revocation check failed (failed to determine whether certificate\nwas revoked)")
		return
	}

	if revoked {
		fmt.Fprintf(os.Stderr, "[!] the certificate has been revoked\n")
		return
	}
}

type appConfig struct {
	caFile, intFile             string
	forceIntermediateBundle     bool
	revexp, skipVerify, verbose bool
}

func parseFlags() appConfig {
	var cfg appConfig
	flag.StringVar(&cfg.caFile, "ca", "", "CA certificate `bundle`")
	flag.StringVar(&cfg.intFile, "i", "", "intermediate `bundle`")
	flag.BoolVar(&cfg.forceIntermediateBundle, "f", false,
		"force the use of the intermediate bundle, ignoring any intermediates bundled with certificate")
	flag.BoolVar(&cfg.skipVerify, "k", false, "skip CA verification")
	flag.BoolVar(&cfg.revexp, "r", false, "print revocation and expiry information")
	flag.BoolVar(&cfg.verbose, "v", false, "verbose")
	flag.Parse()
	return cfg
}

func loadRoots(caFile string, verbose bool) (*x509.CertPool, error) {
	if caFile == "" {
		return x509.SystemCertPool()
	}

	if verbose {
		fmt.Println("[+] loading root certificates from", caFile)
	}
	return certlib.LoadPEMCertPool(caFile)
}

func loadIntermediates(intFile string, verbose bool) (*x509.CertPool, error) {
	if intFile == "" {
		return x509.NewCertPool(), nil
	}
	if verbose {
		fmt.Println("[+] loading intermediate certificates from", intFile)
	}
	// Note: use intFile here (previously used caFile mistakenly)
	return certlib.LoadPEMCertPool(intFile)
}

func addBundledIntermediates(chain []*x509.Certificate, pool *x509.CertPool, verbose bool) {
	for _, intermediate := range chain[1:] {
		if verbose {
			fmt.Printf("[+] adding intermediate with SKI %x\n", intermediate.SubjectKeyId)
		}
		pool.AddCert(intermediate)
	}
}

func verifyCert(cert *x509.Certificate, roots, ints *x509.CertPool) error {
	opts := x509.VerifyOptions{
		Intermediates: ints,
		Roots:         roots,
		KeyUsages:     []x509.ExtKeyUsage{x509.ExtKeyUsageAny},
	}
	_, err := cert.Verify(opts)
	return err
}

func run(cfg appConfig) error {
	roots, err := loadRoots(cfg.caFile, cfg.verbose)
	if err != nil {
		return err
	}

	ints, err := loadIntermediates(cfg.intFile, cfg.verbose)
	if err != nil {
		return err
	}

	if flag.NArg() != 1 {
		fmt.Fprintf(os.Stderr, "Usage: %s [-ca bundle] [-i bundle] cert", lib.ProgName())
	}

	combinedPool, err := certlib.LoadFullCertPool(cfg.caFile, cfg.intFile)
	if err != nil {
		return fmt.Errorf("failed to build combined pool: %w", err)
	}

	opts := &certlib.FetcherOpts{
		Roots:      combinedPool,
		SkipVerify: cfg.skipVerify,
	}

	chain, err := certlib.GetCertificateChain(flag.Arg(0), opts)
	if err != nil {
		return err
	}
	if cfg.verbose {
		fmt.Printf("[+] %s has %d certificates\n", flag.Arg(0), len(chain))
	}

	cert := chain[0]
	if len(chain) > 1 && !cfg.forceIntermediateBundle {
		addBundledIntermediates(chain, ints, cfg.verbose)
	}

	if err = verifyCert(cert, roots, ints); err != nil {
		return fmt.Errorf("certificate verification failed: %w", err)
	}

	if cfg.verbose {
		fmt.Println("OK")
	}

	if cfg.revexp {
		printRevocation(cert)
	}
	return nil
}

func main() {
	cfg := parseFlags()
	if err := run(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
