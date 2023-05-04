GOUTILS

### Note: this repo is archived and not being maintained here anymore.

This is a collection of small utility code I've written in Go; the `cmd/`
directory has a number of command-line utilities. Rather than keep all
of these in superfluous repositories of their own, I'm putting them here.

Contents:

    ahash/          Provides hashes from string algorithm specifiers.
    assert/         Error handling, assertion-style.
    cmd/
        atping/     Automated TCP ping, meant for putting in cronjobs.
        certchain/  Display the certificate chain from a
                    TLS connection.
        certdump/   Dump certificate information.
        certexpiry/ Print a list of certificate subjects and expiry times
                    or warn about certificates expiring within a certain
                    window.
        certverify/ Verify a TLS X.509 certificate, optionally printing
                    the time to expiry and checking for revocations.
        clustersh/  Run commands or transfer files across multiple
                    servers via SSH.
        cruntar/    Untar an archive with hard links, copying instead of
                    linking.
        csrpubdump/ Dump the public key from an X.509 certificate request.
	diskimg/    Write a disk image to a device.
	eig/	    EEPROM image generator.
        fragment/   Print a fragment of a file.
        jlp/        JSON linter/prettifier.
        kgz/        Custom gzip compressor / decompressor that handles 99%
                    of my use cases.
	parts/	    Simple parts database management for my collection of
		    electronic components.
        pem2bin/    Dump the binary body of a PEM-encoded block.
        pembody/    Print the body of a PEM certificate.
        pemit/      Dump data to a PEM file.
        readchain/  Print the common name for the certificates
                    in a bundle.
        renfnv/     Rename a file to base32-encoded 64-bit FNV-1a hash.
        rhash/      Compute the digest of remote files.
        showimp/    List the external (e.g. non-stdlib and outside the
                    current working directory) imports for a Go file.
        ski         Display the SKI for PEM-encoded TLS material.
	sprox/	    Simple TCP proxy.
        stealchain/ Dump the verified chain from a TLS 
                    connection to a server.
	stealchain- Dump the verified chain from a TLS 
	  server/   connection from a client.
        subjhash/   Print or match subject info from a certificate.
        tlskeypair/ Check whether a TLS certificate and key file match.
        utc/        Convert times to UTC.
        yamll/      A small YAML linter.
    config/	    A simple global configuration system where configuration
    		    data is pulled from a file or an environment variable
		    transparently.
    dbg/	    A debug printer.
    die/            Death of a program.
    fileutil/       Common file functions.
    lib/            Commonly-useful functions for writing Go programs.
    logging/        A logging library.
    mwc/            MultiwriteCloser implementation.
    rand/	    Utilities for working with math/rand.
    sbuf/           A byte buffer that can be wiped.
    seekbuf/	    A read-seekable byte buffer.
    syslog/	    Syslog-type logging.
    tee/            Emulate tee(1)'s functionality in io.Writers.
    testio/         Various I/O utilities useful during testing.
    testutil/       Various utility functions useful during testing.


Each program should have a small README in the directory with more
information.

All code here is licensed under the ISC license.
