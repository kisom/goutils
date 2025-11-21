package fetch

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
	"git.wntrmute.dev/kyle/goutils/lib"
	"git.wntrmute.dev/kyle/goutils/lib/dialer"
)

// Note: Previously this package exposed a FetcherOpts type. It has been
// refactored to use *tls.Config directly for configuring TLS behavior.

// Fetcher is an interface for fetching certificates from a remote source. It
// currently supports fetching from a server or a file.
type Fetcher interface {
	// Get retrieves the leaf certificate from the source.
	Get() (*x509.Certificate, error)

	// GetChain retrieves the entire chain from the Fetcher.
	GetChain() ([]*x509.Certificate, error)

	// String returns a string representation of the Fetcher.
	String() string
}

func NewFetcher(spec string, tcfg *tls.Config) (Fetcher, error) {
	if fileutil.FileDoesExist(spec) || spec == "-" {
		return NewFileFetcher(spec), nil
	}

	fetcher, err := ParseServer(spec, tcfg)
	if err != nil {
		return nil, err
	}

	fetcher.config = tcfg

	return fetcher, nil
}

// ServerFetcher retrieves certificates from a TLS connection.
type ServerFetcher struct {
	host   string
	port   int
	config *tls.Config
}

// ParseServer parses a server string into a ServerFetcher. It can be a URL or a
// a host:port pair.
func ParseServer(host string, cfg *tls.Config) (*ServerFetcher, error) {
	target, err := hosts.ParseHost(host)
	if err != nil {
		return nil, fmt.Errorf("failed to parse server: %w", err)
	}

	return &ServerFetcher{
		host:   target.Host,
		port:   target.Port,
		config: cfg,
	}, nil
}

func (sf *ServerFetcher) String() string {
	return fmt.Sprintf("tls://%s", net.JoinHostPort(sf.host, lib.Itoa(sf.port, -1)))
}

func (sf *ServerFetcher) GetChain() ([]*x509.Certificate, error) {
	opts := dialer.Opts{
		TLSConfig: sf.config,
	}

	conn, err := dialer.DialTLS(context.Background(), net.JoinHostPort(sf.host, lib.Itoa(sf.port, -1)), opts)
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

// FileFetcher retrieves certificates from files on disk.
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
	fetcher, err := NewFetcher(spec, cfg)
	if err != nil {
		return nil, err
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
