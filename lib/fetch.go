package lib

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"net"
	"os"

	"git.wntrmute.dev/kyle/goutils/certlib"
	"git.wntrmute.dev/kyle/goutils/certlib/hosts"
	"git.wntrmute.dev/kyle/goutils/fileutil"
)

// Note: Previously this package exposed a FetcherOpts type. It has been
// refactored to use *tls.Config directly for configuring TLS behavior.

// Fetcher is an interface for fetching certificates from a remote source. It
// currently supports fetching from a server or a file.
type Fetcher interface {
	Get() (*x509.Certificate, error)
	GetChain() ([]*x509.Certificate, error)
	String() string
}

type ServerFetcher struct {
	host     string
	port     int
	insecure bool
	roots    *x509.CertPool
}

// WithRoots sets the roots for the ServerFetcher.
func WithRoots(roots *x509.CertPool) func(*ServerFetcher) {
	return func(sf *ServerFetcher) {
		sf.roots = roots
	}
}

// WithSkipVerify sets the insecure flag for the ServerFetcher.
func WithSkipVerify() func(*ServerFetcher) {
	return func(sf *ServerFetcher) {
		sf.insecure = true
	}
}

// ParseServer parses a server string into a ServerFetcher. It can be a URL or a
// a host:port pair.
func ParseServer(host string) (*ServerFetcher, error) {
	target, err := hosts.ParseHost(host)
	if err != nil {
		return nil, fmt.Errorf("failed to parse server: %w", err)
	}

	return &ServerFetcher{
		host: target.Host,
		port: target.Port,
	}, nil
}

func (sf *ServerFetcher) String() string {
	return fmt.Sprintf("tls://%s", net.JoinHostPort(sf.host, Itoa(sf.port, -1)))
}

func (sf *ServerFetcher) GetChain() ([]*x509.Certificate, error) {
	opts := DialerOpts{
		TLSConfig: &tls.Config{
			InsecureSkipVerify: sf.insecure, // #nosec G402 - no shit sherlock
			RootCAs:            sf.roots,
		},
	}

	conn, err := DialTLS(context.Background(), net.JoinHostPort(sf.host, Itoa(sf.port, -1)), opts)
	if err != nil {
		return nil, fmt.Errorf("failed to dial server: %w", err)
	}
	defer conn.Close()

	state := conn.ConnectionState()
	return state.PeerCertificates, nil
}

func (sf *ServerFetcher) Get() (*x509.Certificate, error) {
	certs, err := sf.GetChain()
	if err != nil {
		return nil, err
	}

	return certs[0], nil
}

type FileFetcher struct {
	path string
}

func NewFileFetcher(path string) *FileFetcher {
	return &FileFetcher{
		path: path,
	}
}

func (ff *FileFetcher) String() string {
	return ff.path
}

func (ff *FileFetcher) GetChain() ([]*x509.Certificate, error) {
	if ff.path == "-" {
		certData, err := io.ReadAll(os.Stdin)
		if err != nil {
			return nil, fmt.Errorf("failed to read from stdin: %w", err)
		}

		return certlib.ParseCertificatesPEM(certData)
	}

	certs, err := certlib.LoadCertificates(ff.path)
	if err != nil {
		return nil, fmt.Errorf("failed to load chain: %w", err)
	}

	return certs, nil
}

func (ff *FileFetcher) Get() (*x509.Certificate, error) {
	certs, err := ff.GetChain()
	if err != nil {
		return nil, err
	}

	return certs[0], nil
}

// GetCertificateChain fetches a certificate chain from a remote source.
// If cfg is non-nil and spec refers to a TLS server, the provided TLS
// configuration will be used to control verification behavior (e.g.,
// InsecureSkipVerify, RootCAs).
func GetCertificateChain(spec string, cfg *tls.Config) ([]*x509.Certificate, error) {
	if fileutil.FileDoesExist(spec) {
		return NewFileFetcher(spec).GetChain()
	}

	fetcher, err := ParseServer(spec)
	if err != nil {
		return nil, err
	}

	if cfg != nil {
		fetcher.insecure = cfg.InsecureSkipVerify
		fetcher.roots = cfg.RootCAs
	}

	return fetcher.GetChain()
}

// GetCertificate fetches the first certificate from a certificate chain.
func GetCertificate(spec string, cfg *tls.Config) (*x509.Certificate, error) {
	certs, err := GetCertificateChain(spec, cfg)
	if err != nil {
		return nil, err
	}

	return certs[0], nil
}
