package main

import (
	"crypto/x509"
	"flag"
	"fmt"
	"os"

	"git.wntrmute.dev/kyle/goutils/certlib"
	"git.wntrmute.dev/kyle/goutils/certlib/verify"
	"git.wntrmute.dev/kyle/goutils/die"
	"git.wntrmute.dev/kyle/goutils/lib"
)

type appConfig struct {
	caFile, intFile             string
	forceIntermediateBundle     bool
	revexp, skipVerify, verbose bool
	strictTLS                   bool
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
	lib.StrictTLSFlag(&cfg.strictTLS)
	flag.Parse()

	if flag.NArg() == 0 {
		die.With("usage: certverify targets...")
	}

	return cfg
}

func main() {
	var (
		roots, ints *x509.CertPool
		err         error
		failed      bool
	)

	cfg := parseFlags()

	opts := &verify.Opts{
		CheckRevocation:    cfg.revexp,
		ForceIntermediates: cfg.forceIntermediateBundle,
		Verbose:            cfg.verbose,
	}

	if cfg.caFile != "" {
		if cfg.verbose {
			fmt.Printf("loading CA certificates from %s\n", cfg.caFile)
		}

		roots, err = certlib.LoadPEMCertPool(cfg.caFile)
		die.If(err)
	}

	if cfg.intFile != "" {
		if cfg.verbose {
			fmt.Printf("loading intermediate certificates from %s\n", cfg.intFile)
		}

		ints, err = certlib.LoadPEMCertPool(cfg.intFile)
		die.If(err)
	}

	opts.Config, err = lib.BaselineTLSConfig(cfg.skipVerify, cfg.strictTLS)
	die.If(err)

	opts.Config.RootCAs = roots
	opts.Intermediates = ints

	for _, arg := range flag.Args() {
		_, err = verify.Chain(os.Stdout, arg, opts)
		if err != nil {
			lib.Warn(err, "while verifying %s", arg)
			failed = true
		} else {
			fmt.Printf("%s: OK\n", arg)
		}
	}

	if failed {
		os.Exit(1)
	}
}
