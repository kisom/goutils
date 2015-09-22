GOUTILS

This is a collection of small utility code I've written in Go; the `cmd/`
directory has a number of command-line utilities. Rather than keep all
of these in superfluous repositories of their own, I'm putting them here.

Contents:

	die/			Death of a program.
	cmd/
		certchain/	Display the certificate chain from a
				TLS connection.
                certdump/       Dump certificate information.
		csrpubdump/	Dump the public key from an X.509
				certificate request.
                fragment/       Print a fragment of a file.
		readchain/	Print the common name for the certificates
				in a bundle.
		stealchain/	Dump the verified chain from a TLS
				connection.
	
Each program should have a small README in the directory with more information.

All code here is licensed under the MIT license.
