ca-signed: verify certificates against a CA
-------------------------------------------

Description
  ca-signed verifies whether one or more certificates are signed by a given
  Certificate Authority (CA). It prints a concise status per input certificate
  along with the certificate’s expiration date when validation succeeds.

Usage
  ca-signed CA.pem cert1.pem [cert2.pem ...]

  - CA.pem: A file containing one or more CA certificates in PEM, DER, or PKCS#7/PKCS#12 formats.
  - certN.pem: A file containing the end-entity (leaf) certificate to verify. If the file contains a chain,
               the first certificate is treated as the leaf and the remaining ones are used as intermediates.

Output format
  For each input certificate file, one line is printed:
    <filename>: OK (expires YYYY-MM-DD)
    <filename>: INVALID

Special self-test mode
  ca-signed selftest

  Runs a built-in test suite using embedded certificates. This mode requires no
  external files or network access. The program exits with code 0 if all tests
  pass, or a non-zero exit code if any test fails. Example output lines include
  whether validation succeeds and the leaf’s expiration when applicable.

Examples
  # Verify a server certificate against a root CA
  ca-signed isrg-root-x1.pem le-e7.pem

  # Run the embedded self-test suite
  ca-signed selftest

Notes
  - The tool attempts to parse certificates in PEM first, then falls back to
    DER/PKCS#7/PKCS#12 (with an empty password) where applicable.
  - Expiration is shown for the leaf certificate only.
  - In selftest mode, test certificates are compiled into the binary using go:embed.
