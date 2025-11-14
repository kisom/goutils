tlsinfo: show TLS version, cipher, and peer certificates
---------------------------------------------------------

Description
  tlsinfo connects to a TLS server and prints the negotiated TLS version and
  cipher suite, followed by details for each certificate in the server’s
  presented chain (as provided by the server).

Usage
  tlsinfo <hostname:port>

Output
  The program prints the negotiated protocol and cipher, then one section per
  certificate in the order received from the server. Example fields:
    TLS Version: TLS 1.3
    Cipher Suite: TLS_AES_128_GCM_SHA256
    Certificate 1
        Subject: CN=example.com, O=Example Corp, C=US
        Issuer: CN=Example Root CA, O=Example Corp, C=US
        DNS Names: [example.com www.example.com]
        Not Before: 2025-01-01 00:00:00 +0000 UTC
        Not After:  2026-01-01 23:59:59 +0000 UTC

Examples
  # Inspect a public HTTPS endpoint
  tlsinfo example.com:443

Notes
  - Verification is intentionally disabled (InsecureSkipVerify=true). The tool
    does not validate the server certificate or hostname; it is for inspection
    only.
  - The SNI/ServerName is inferred from <hostname> when applicable.
  - You must specify a port (e.g., 443 for HTTPS).
  - The entire certificate chain is printed exactly as presented by the server.
