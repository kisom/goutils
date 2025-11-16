package main

import (
	"context"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"flag"
	"fmt"
	"net"
	"os"

	"git.wntrmute.dev/kyle/goutils/die"
)

func main() {
	cfg := &tls.Config{} // #nosec G402

	var sysRoot, listenAddr, certFile, keyFile string
	var verify bool
	flag.StringVar(&sysRoot, "ca", "", "provide an alternate CA bundle")
	flag.StringVar(&listenAddr, "listen", ":443", "address to listen on")
	flag.StringVar(&certFile, "cert", "", "server certificate to present to clients")
	flag.StringVar(&keyFile, "key", "", "key for server certificate")
	flag.BoolVar(&verify, "verify", false, "verify client certificates")
	flag.Parse()

	if verify {
		cfg.ClientAuth = tls.RequireAndVerifyClientCert
	} else {
		cfg.ClientAuth = tls.RequestClientCert
	}
	if certFile == "" {
		fmt.Println("[!] missing required flag -cert")
		os.Exit(1)
	}
	if keyFile == "" {
		fmt.Println("[!] missing required flag -key")
		os.Exit(1)
	}
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		fmt.Printf("[!] could not load server key pair: %v", err)
		os.Exit(1)
	}
	cfg.Certificates = append(cfg.Certificates, cert)
	if sysRoot != "" {
		var pemList []byte
		pemList, err = os.ReadFile(sysRoot)
		die.If(err)

		roots := x509.NewCertPool()
		if !roots.AppendCertsFromPEM(pemList) {
			fmt.Printf("[!] no valid roots found")
			roots = nil
		}

		cfg.RootCAs = roots
	}

	lc := &net.ListenConfig{}
	l, err := lc.Listen(context.Background(), "tcp", listenAddr)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	for {
		var conn net.Conn
		conn, err = l.Accept()
		if err != nil {
			fmt.Println(err.Error())
			continue
		}
		handleConn(conn, cfg)
	}
}

// handleConn performs a TLS handshake, extracts the peer chain, and writes it to a file.
func handleConn(conn net.Conn, cfg *tls.Config) {
	defer conn.Close()
	raddr := conn.RemoteAddr()
	tconn := tls.Server(conn, cfg)
	if err := tconn.HandshakeContext(context.Background()); err != nil {
		fmt.Printf("[+] %v: failed to complete handshake: %v\n", raddr, err)
		return
	}
	cs := tconn.ConnectionState()
	if len(cs.PeerCertificates) == 0 {
		fmt.Printf("[+] %v: no chain presented\n", raddr)
		return
	}

	var chain []byte
	for _, cert := range cs.PeerCertificates {
		p := &pem.Block{Type: "CERTIFICATE", Bytes: cert.Raw}
		chain = append(chain, pem.EncodeToMemory(p)...)
	}

	var nonce [16]byte
	if _, err := rand.Read(nonce[:]); err != nil {
		fmt.Printf("[+] %v: failed to generate filename nonce: %v\n", raddr, err)
		return
	}
	fname := fmt.Sprintf("%v-%v.pem", raddr, hex.EncodeToString(nonce[:]))
	if err := os.WriteFile(fname, chain, 0o644); err != nil {
		fmt.Printf("[+] %v: failed to write %v: %v\n", raddr, fname, err)
		return
	}
	fmt.Printf("%v: [+] wrote %v.\n", raddr, fname)
}
