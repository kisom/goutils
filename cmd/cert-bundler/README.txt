cert-bundler: create certificate chain archives
------------------------------------------------

Description
  cert-bundler creates archives of certificate chains from a YAML configuration
  file. It validates certificates, checks expiration dates, and generates
  archives in multiple formats (zip, tar.gz) with optional manifest files
  containing SHA256 checksums.

Usage
  cert-bundler [options]

  Options:
    -c <file>    Path to YAML configuration file (default: bundle.yaml)
    -o <dir>     Output directory for archives (default: pkg)

YAML Configuration Format

  The configuration file uses the following structure:

  config:
    hashes: <filename>
    expiry: <duration>
  chains:
    <group_name>:
      certs:
        - root: <path>
          intermediates:
            - <path>
            - <path>
        - root: <path>
          intermediates:
            - <path>
      outputs:
        include_single: <bool>
        include_individual: <bool>
        manifest: <bool>
        encoding: <encoding>
        formats:
          - <format>
          - <format>

Configuration Fields

  config:
    hashes: (optional) Name of the file to write SHA256 checksums of all
            generated archives. If omitted, no hash file is created.
    expiry: (optional) Expiration warning threshold. Certificates expiring
            within this period will trigger a warning. Supports formats like
            "1y" (year), "6m" (month), "30d" (day). Default: 1y

  chains:
    Each key under "chains" defines a named certificate group. All certificates
    in a group are bundled together into archives with the group name.

    certs:
      List of certificate chains. Each chain has:
        root: Path to root CA certificate (PEM or DER format)
        intermediates: List of paths to intermediate certificates

      All intermediates are validated against their root CA. An error is
      reported if signature verification fails.

    outputs:
      Defines output formats and content for the group's archives:

        include_single: (bool) If true, all certificates in the group are
                        concatenated into a single file named "bundle.pem"
                        (or "bundle.crt" for DER encoding).

        include_individual: (bool) If true, each certificate is included as
                            a separate file in the archive, named after the
                            original file (e.g., "int/cca2.pem" becomes
                            "cca2.pem").

        manifest: (bool) If true, a MANIFEST file is included containing
                  SHA256 checksums of all files in the archive.

        encoding: Specifies certificate encoding in the archive:
                  - "pem": PEM format with .pem extension (default)
                  - "der": DER format with .crt extension
                  - "both": Both PEM and DER versions are included

        formats: List of archive formats to generate:
                 - "zip": Creates a .zip archive
                 - "tgz": Creates a .tar.gz archive

Output Files

  For each group and format combination, an archive is created:
    <group_name>.zip or <group_name>.tar.gz

  If config.hashes is specified, a hash file is created in the output directory
  containing SHA256 checksums of all generated archives.

Example Configuration

  config:
    hashes: bundle.sha256
    expiry: 1y
  chains:
    core_certs:
      certs:
        - root: roots/core-ca.pem
          intermediates:
            - int/cca1.pem
            - int/cca2.pem
            - int/cca3.pem
        - root: roots/ssh-ca.pem
          intermediates:
            - ssh/ssh_dmz1.pem
            - ssh/ssh_internal.pem
      outputs:
        include_single: true
        include_individual: true
        manifest: true
        encoding: pem
        formats:
          - zip
          - tgz

  This configuration:
    - Creates core_certs.zip and core_certs.tar.gz in the output directory
    - Each archive contains bundle.pem (all certificates concatenated)
    - Each archive contains individual certificates (core-ca.pem, cca1.pem, etc.)
    - Each archive includes a MANIFEST file with SHA256 checksums
    - Creates bundle.sha256 with checksums of the two archives
    - Warns if any certificate expires within 1 year

Examples

  # Create bundles using default configuration (bundle.yaml -> pkg/)
  cert-bundler

  # Use custom configuration and output directory
  cert-bundler -c myconfig.yaml -o output

  # Create bundles from testdata configuration
  cert-bundler -c testdata/bundle.yaml -o testdata/pkg

Notes
  - Certificate paths in the YAML are relative to the current working directory
  - All intermediates must be properly signed by their specified root CA
  - Certificates are checked for expiration; warnings are printed to stderr
  - Expired certificates do not prevent archive creation but generate warnings
  - Both PEM and DER certificate formats are supported as input
  - Archive filenames use the group name, not individual chain names
  - If both include_single and include_individual are true, archives contain both
