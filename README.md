GOUTILS

This is a collection of small utility code I've written in Go; the `cmd/`
directory has a number of command-line utilities. Rather than keep all
of these in superfluous repositories of their own, I'm putting them here.

Contents:

    die/            Death of a program.
    cmd/
        certchain/  Display the certificate chain from a
                    TLS connection.
        certdump/   Dump certificate information.
	certverify/ Verify a TLS X.509 certificate.
	clustersh/  Run commands or transfer files across multiple
                    servers via SSH.
        csrpubdump/ Dump the public key from an X.509
                    certificate request.
        fragment/   Print a fragment of a file.
        jlp/        JSON linter/prettifier.
	pem2bin/    Dump the binary body of a PEM-encoded block.
        pembody/    Print the body of a PEM certificate.
        showimp/    List the external (e.g. non-stdlib and outside the
                    current working directory) imports for a Go file.
        readchain/  Print the common name for the certificates
                    in a bundle.
        stealchain/ Dump the verified chain from a TLS
                    connection.
        tlskeypair/ Check whether a TLS certificate and key file match.
    lib/            Commonly-useful functions for writing Go programs.
    logging/        A logging library.
    mwc/            MultiwriteCloser implementation.
    sbuf/           A byte buffer that can be wiped.

    
Each program should have a small README in the directory with more information.

All code here is licensed under the MIT license.
