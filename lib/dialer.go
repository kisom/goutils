// Package lib contains reusable helpers. This file provides proxy-aware
// dialers for plain TCP and TLS connections using environment variables.
//
// Supported proxy environment variables (checked case-insensitively):
//   - SOCKS5_PROXY  (e.g., socks5://user:pass@host:1080)
//   - HTTPS_PROXY   (e.g., https://user:pass@host:443)
//   - HTTP_PROXY    (e.g., http://user:pass@host:3128)
//
// Precedence when multiple proxies are set (both for net and TLS dialers):
//  1. SOCKS5_PROXY
//  2. HTTPS_PROXY
//  3. HTTP_PROXY
//
// Both uppercase and lowercase variable names are honored.
package lib

import (
	"bufio"
	"context"
	"crypto/tls"
	"encoding/base64"
	"errors"
	"flag"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	xproxy "golang.org/x/net/proxy"

	"git.wntrmute.dev/kyle/goutils/dbg"
)

// StrictBaselineTLSConfig returns a secure TLS config.
// Many of the tools in this repo are designed to debug broken TLS systems
// and therefore explicitly support old or insecure TLS setups.
func StrictBaselineTLSConfig() *tls.Config {
	return &tls.Config{
		MinVersion:         tls.VersionTLS12,
		InsecureSkipVerify: false, // explicitly set
	}
}

func StrictTLSFlag(useStrict *bool) {
	flag.BoolVar(useStrict, "strict-tls", false, "Use strict TLS configuration (disables certificate verification)")
}

func BaselineTLSConfig(skipVerify bool, secure bool) (*tls.Config, error) {
	if secure && skipVerify {
		return nil, errors.New("cannot skip verification and use secure TLS")
	}

	if skipVerify {
		return &tls.Config{InsecureSkipVerify: true}, nil // #nosec G402 - intentional
	}

	if secure {
		return StrictBaselineTLSConfig(), nil
	}

	return &tls.Config{}, nil // #nosec G402 - intentional
}

var debug = dbg.NewFromEnv()

// DialerOpts controls creation of proxy-aware dialers.
//
// Timeout controls the maximum amount of time spent establishing the
// underlying TCP connection and any proxy handshake. If zero, a
// reasonable default (30s) is used.
//
// TLSConfig is used by the TLS dialer to configure the TLS handshake to
// the target endpoint. If TLSConfig.ServerName is empty, it will be set
// from the host portion of the address passed to DialContext.
type DialerOpts struct {
	Timeout   time.Duration
	TLSConfig *tls.Config
}

// ContextDialer matches the common DialContext signature used by net and tls dialers.
type ContextDialer interface {
	DialContext(ctx context.Context, network, address string) (net.Conn, error)
}

// DialTCP is a convenience helper that dials a TCP connection to address
// using a proxy-aware dialer derived from opts. It honors SOCKS5_PROXY,
// HTTPS_PROXY, and HTTP_PROXY environment variables.
func DialTCP(ctx context.Context, address string, opts DialerOpts) (net.Conn, error) {
	d, err := NewNetDialer(opts)
	if err != nil {
		return nil, err
	}
	return d.DialContext(ctx, "tcp", address)
}

// DialTLS is a convenience helper that dials a TLS-wrapped TCP connection to
// address using a proxy-aware dialer derived from opts. It returns a *tls.Conn.
// It honors SOCKS5_PROXY, HTTPS_PROXY, and HTTP_PROXY environment variables and
// uses opts.TLSConfig for the handshake (filling ServerName from address if empty).
func DialTLS(ctx context.Context, address string, opts DialerOpts) (*tls.Conn, error) {
	d, err := NewTLSDialer(opts)
	if err != nil {
		return nil, err
	}

	c, err := d.DialContext(ctx, "tcp", address)
	if err != nil {
		return nil, err
	}

	tlsConn, ok := c.(*tls.Conn)
	if !ok {
		_ = c.Close()
		return nil, fmt.Errorf("DialTLS: expected *tls.Conn, got %T", c)
	}
	return tlsConn, nil
}

// NewNetDialer returns a ContextDialer that dials TCP connections using
// proxies discovered from the environment (SOCKS5_PROXY, HTTPS_PROXY, HTTP_PROXY).
// The returned dialer supports context cancellation for direct and HTTP(S)
// proxies and applies the configured timeout to connection/proxy handshake.
func NewNetDialer(opts DialerOpts) (ContextDialer, error) {
	if opts.Timeout <= 0 {
		opts.Timeout = 30 * time.Second
	}

	if u := getProxyURLFromEnv("SOCKS5_PROXY"); u != nil {
		debug.Printf("using SOCKS5 proxy %q\n", u)
		return newSOCKS5Dialer(u, opts)
	}

	if u := getProxyURLFromEnv("HTTPS_PROXY"); u != nil {
		// Respect the proxy URL scheme. Zscaler may set HTTPS_PROXY to an HTTP proxy
		// running locally; in that case we must NOT TLS-wrap the proxy connection.
		debug.Printf("using HTTPS proxy %q\n", u)
		return &httpProxyDialer{
			proxyURL: u,
			timeout:  opts.Timeout,
			secure:   strings.EqualFold(u.Scheme, "https"),
			config:   opts.TLSConfig,
		}, nil
	}

	if u := getProxyURLFromEnv("HTTP_PROXY"); u != nil {
		debug.Printf("using HTTP proxy %q\n", u)
		return &httpProxyDialer{
			proxyURL: u,
			timeout:  opts.Timeout,
			// Only TLS-wrap the proxy connection if the URL scheme is https.
			secure: strings.EqualFold(u.Scheme, "https"),
			config: opts.TLSConfig,
		}, nil
	}

	// Direct dialer
	return &net.Dialer{Timeout: opts.Timeout}, nil
}

// NewTLSDialer returns a ContextDialer that establishes a TLS connection to
// the destination, while honoring SOCKS5_PROXY/HTTPS_PROXY/HTTP_PROXY.
//
// The returned dialer performs proxy negotiation (if any), then completes a
// TLS handshake to the target using opts.TLSConfig.
func NewTLSDialer(opts DialerOpts) (ContextDialer, error) {
	if opts.Timeout <= 0 {
		opts.Timeout = 30 * time.Second
	}

	// Prefer SOCKS5 if present.
	if u := getProxyURLFromEnv("SOCKS5_PROXY"); u != nil {
		debug.Printf("using SOCKS5 proxy %q\n", u)
		base, err := newSOCKS5Dialer(u, opts)
		if err != nil {
			return nil, err
		}
		return &tlsWrappingDialer{base: base, tcfg: opts.TLSConfig, timeout: opts.Timeout}, nil
	}

	// For TLS, prefer HTTPS proxy over HTTP if both set.
	if u := getProxyURLFromEnv("HTTPS_PROXY"); u != nil {
		debug.Printf("using HTTPS proxy %q\n", u)
		base := &httpProxyDialer{
			proxyURL: u,
			timeout:  opts.Timeout,
			secure:   strings.EqualFold(u.Scheme, "https"),
			config:   opts.TLSConfig,
		}
		return &tlsWrappingDialer{base: base, tcfg: opts.TLSConfig, timeout: opts.Timeout}, nil
	}

	if u := getProxyURLFromEnv("HTTP_PROXY"); u != nil {
		debug.Printf("using HTTP proxy %q\n", u)
		base := &httpProxyDialer{
			proxyURL: u,
			timeout:  opts.Timeout,
			secure:   strings.EqualFold(u.Scheme, "https"),
			config:   opts.TLSConfig,
		}
		return &tlsWrappingDialer{base: base, tcfg: opts.TLSConfig, timeout: opts.Timeout}, nil
	}

	// Direct TLS
	base := &net.Dialer{Timeout: opts.Timeout}
	return &tlsWrappingDialer{base: base, tcfg: opts.TLSConfig, timeout: opts.Timeout}, nil
}

// ---- Implementation helpers ----

func getProxyURLFromEnv(name string) *url.URL {
	// check both upper/lowercase
	v := os.Getenv(name)
	if v == "" {
		v = os.Getenv(strings.ToLower(name))
	}
	if v == "" {
		return nil
	}
	// If scheme omitted, infer from env var name.
	if !strings.Contains(v, "://") {
		switch strings.ToUpper(name) {
		case "SOCKS5_PROXY":
			v = "socks5://" + v
		case "HTTPS_PROXY":
			v = "https://" + v
		default:
			v = "http://" + v
		}
	}
	u, err := url.Parse(v)
	if err != nil {
		return nil
	}
	return u
}

// NewHTTPClient returns an *http.Client that is proxy-aware.
//
// Behavior:
//   - If SOCKS5_PROXY is set, the client routes all TCP connections through the
//     SOCKS5 proxy using a custom DialContext, and disables HTTP(S) proxying in
//     the transport (per our precedence SOCKS5 > HTTPS > HTTP).
//   - Otherwise, it uses http.ProxyFromEnvironment which supports HTTP_PROXY,
//     HTTPS_PROXY, and NO_PROXY/no_proxy.
//   - Connection and TLS handshake timeouts are derived from opts.Timeout.
//   - For HTTPS targets, opts.TLSConfig is applied to the transport.
func NewHTTPClient(opts DialerOpts) (*http.Client, error) {
	if opts.Timeout <= 0 {
		opts.Timeout = 30 * time.Second
	}

	// Base transport configuration
	tr := &http.Transport{
		TLSClientConfig:     opts.TLSConfig,
		TLSHandshakeTimeout: opts.Timeout,
		// Leave other fields as Go defaults for compatibility.
	}

	// If SOCKS5 is configured, use our dialer and disable HTTP proxying to
	// avoid double-proxying. Otherwise, rely on ProxyFromEnvironment for
	// HTTP(S) proxies and still set a connect timeout via net.Dialer.
	if u := getProxyURLFromEnv("SOCKS5_PROXY"); u != nil {
		d, err := newSOCKS5Dialer(u, opts)
		if err != nil {
			return nil, err
		}
		tr.Proxy = nil
		tr.DialContext = func(ctx context.Context, network, address string) (net.Conn, error) {
			return d.DialContext(ctx, network, address)
		}
	} else {
		tr.Proxy = http.ProxyFromEnvironment
		// Use a standard net.Dialer to ensure we apply a connect timeout.
		nd := &net.Dialer{Timeout: opts.Timeout}
		tr.DialContext = nd.DialContext
	}

	// Construct client; we don't set Client.Timeout here to avoid affecting
	// streaming responses. Callers can set it if they want an overall deadline.
	return &http.Client{Transport: tr}, nil
}

// httpProxyDialer implements CONNECT tunneling over HTTP or HTTPS proxy.
type httpProxyDialer struct {
	proxyURL *url.URL
	timeout  time.Duration
	secure   bool // true for HTTPS proxy
	config   *tls.Config
}

// proxyAddress returns host:port for the proxy, applying defaults by scheme when missing.
func (d *httpProxyDialer) proxyAddress() string {
	proxyAddr := d.proxyURL.Host
	if !strings.Contains(proxyAddr, ":") {
		if d.secure {
			proxyAddr += ":443"
		} else {
			proxyAddr += ":80"
		}
	}
	return proxyAddr
}

// tlsWrapProxyConn performs a TLS handshake to the proxy when d.secure is true.
// It clones the provided tls.Config (if any), ensures ServerName and a safe
// minimum TLS version.
func (d *httpProxyDialer) tlsWrapProxyConn(ctx context.Context, conn net.Conn) (net.Conn, error) {
	host := d.proxyURL.Hostname()
	// Clone provided config (if any) to avoid mutating caller's config.
	cfg := &tls.Config{} // #nosec G402 - intentional
	if d.config != nil {
		cfg = d.config.Clone()
	}

	if cfg.ServerName == "" {
		cfg.ServerName = host
	}

	tlsConn := tls.Client(conn, cfg)
	if err := tlsConn.HandshakeContext(ctx); err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("tls handshake with https proxy failed: %w", err)
	}
	return tlsConn, nil
}

// readConnectResponse reads and validates the proxy's response to a CONNECT
// request. It returns nil on a 200 status and an error otherwise.
func readConnectResponse(br *bufio.Reader) error {
	statusLine, err := br.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read CONNECT response: %w", err)
	}

	if !strings.HasPrefix(statusLine, "HTTP/") {
		return fmt.Errorf("invalid proxy response: %q", strings.TrimSpace(statusLine))
	}

	if !strings.Contains(statusLine, " 200 ") && !strings.HasSuffix(strings.TrimSpace(statusLine), " 200") {
		// Drain headers for context
		_ = drainHeaders(br)
		return fmt.Errorf("proxy CONNECT failed: %s", strings.TrimSpace(statusLine))
	}

	return drainHeaders(br)
}

func (d *httpProxyDialer) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	if !strings.HasPrefix(network, "tcp") {
		return nil, fmt.Errorf("http proxy dialer only supports TCP, got %q", network)
	}

	// Dial to proxy
	var nd = &net.Dialer{Timeout: d.timeout}
	conn, err := nd.DialContext(ctx, "tcp", d.proxyAddress())
	if err != nil {
		return nil, err
	}

	// Deadline covering CONNECT and (for TLS wrapper) will be handled by caller too.
	if d.timeout > 0 {
		_ = conn.SetDeadline(time.Now().Add(d.timeout))
	}

	// If HTTPS proxy, wrap with TLS to the proxy itself.
	if d.secure {
		c, werr := d.tlsWrapProxyConn(ctx, conn)
		if werr != nil {
			return nil, werr
		}
		conn = c
	}

	req := buildConnectRequest(d.proxyURL, address)
	if _, err = conn.Write([]byte(req)); err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("failed to write CONNECT request: %w", err)
	}

	// Read proxy response until end of headers
	br := bufio.NewReader(conn)
	if err = readConnectResponse(br); err != nil {
		_ = conn.Close()
		return nil, err
	}

	// Clear deadline for caller to manage further I/O.
	_ = conn.SetDeadline(time.Time{})
	return conn, nil
}

func buildConnectRequest(proxyURL *url.URL, target string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "CONNECT %s HTTP/1.1\r\n", target)
	fmt.Fprintf(&b, "Host: %s\r\n", target)
	b.WriteString("Proxy-Connection: Keep-Alive\r\n")
	b.WriteString("User-Agent: goutils-dialer/1\r\n")

	if proxyURL.User != nil {
		user := proxyURL.User.Username()
		pass, _ := proxyURL.User.Password()
		auth := base64.StdEncoding.EncodeToString([]byte(user + ":" + pass))
		fmt.Fprintf(&b, "Proxy-Authorization: Basic %s\r\n", auth)
	}
	b.WriteString("\r\n")
	return b.String()
}

func drainHeaders(br *bufio.Reader) error {
	for {
		line, err := br.ReadString('\n')
		if err != nil {
			return fmt.Errorf("reading proxy headers: %w", err)
		}
		if line == "\r\n" || line == "\n" {
			return nil
		}
	}
}

// newSOCKS5Dialer builds a context-aware wrapper over the x/net/proxy dialer.
func newSOCKS5Dialer(u *url.URL, opts DialerOpts) (ContextDialer, error) {
	var auth *xproxy.Auth
	if u.User != nil {
		user := u.User.Username()
		pass, _ := u.User.Password()
		auth = &xproxy.Auth{User: user, Password: pass}
	}
	forward := &net.Dialer{Timeout: opts.Timeout}
	d, err := xproxy.SOCKS5("tcp", hostPortWithDefault(u, "1080"), auth, forward)
	if err != nil {
		return nil, err
	}
	return &socks5ContextDialer{d: d, timeout: opts.Timeout}, nil
}

type socks5ContextDialer struct {
	d       xproxy.Dialer // lacks context; we wrap it
	timeout time.Duration
}

func (s *socks5ContextDialer) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	if !strings.HasPrefix(network, "tcp") {
		return nil, errors.New("socks5 dialer only supports TCP")
	}
	// Best-effort context support: run the non-context dial in a goroutine
	// and respect ctx cancellation/timeout.
	type result struct {
		c   net.Conn
		err error
	}
	ch := make(chan result, 1)
	go func() {
		c, err := s.d.Dial("tcp", address)
		ch <- result{c: c, err: err}
	}()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case r := <-ch:
		return r.c, r.err
	}
}

// tlsWrappingDialer performs a TLS handshake over an existing base dialer.
type tlsWrappingDialer struct {
	base    ContextDialer
	tcfg    *tls.Config
	timeout time.Duration
}

func (t *tlsWrappingDialer) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	if !strings.HasPrefix(network, "tcp") {
		return nil, fmt.Errorf("tls dialer only supports TCP, got %q", network)
	}
	raw, err := t.base.DialContext(ctx, network, address)
	if err != nil {
		return nil, err
	}

	// Apply deadline for handshake.
	if t.timeout > 0 {
		_ = raw.SetDeadline(time.Now().Add(t.timeout))
	}

	var h string
	host := address

	if h, _, err = net.SplitHostPort(address); err == nil {
		host = h
	}
	var cfg *tls.Config
	if t.tcfg != nil {
		// Clone to avoid copying internal locks and to prevent mutating caller's config.
		c := t.tcfg.Clone()
		if c.ServerName == "" {
			c.ServerName = host
		}
		cfg = c
	} else {
		cfg = &tls.Config{ServerName: host, MinVersion: tls.VersionTLS12}
	}

	tlsConn := tls.Client(raw, cfg)
	if err = tlsConn.HandshakeContext(ctx); err != nil {
		_ = raw.Close()
		return nil, err
	}

	// Clear deadline after successful handshake
	_ = tlsConn.SetDeadline(time.Time{})
	return tlsConn, nil
}

func hostPortWithDefault(u *url.URL, defPort string) string {
	host := u.Host
	if !strings.Contains(host, ":") {
		host += ":" + defPort
	}
	return host
}
