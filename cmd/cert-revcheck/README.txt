cert-revcheck: check certificate expiry and revocation
-----------------------------------------------------

Description
  cert-revcheck accepts a list of certificate files (PEM or DER) or
  site addresses (host[:port]) and checks whether the leaf certificate
  is expired or revoked. Revocation checks use CRL and OCSP via the
  certlib/revoke package.

Usage
  cert-revcheck [options] <target> [<target>...]

Options
  -hardfail       treat revocation check failures as fatal (default: false)
  -timeout dur    HTTP/OCSP/CRL timeout for network operations (default: 10s)
  -v              verbose output

Targets
  - File paths to certificates in PEM or DER format.
  - Site addresses in the form host or host:port. If no port is
    provided, 443 is assumed.

Examples
  # Check a PEM file
  cert-revcheck ./server.pem

  # Check a DER (single) certificate
  cert-revcheck ./server.der

  # Check a live site (leaf certificate)
  cert-revcheck example.com:443

Notes
  - For sites, only the leaf certificate is checked.
  - When -hardfail is set, network issues during OCSP/CRL fetch will
    cause the check to fail (treated as revoked).
