package main

import (
	"encoding/pem"
	"flag"
	"io"
	"os"

	"git.wntrmute.dev/kyle/goutils/lib"
)

func main() {
	flag.Parse()
	if flag.NArg() != 1 {
		lib.Errx(lib.ExitFailure, "a single filename is required")
	}

	var in []byte
	var err error

	path := flag.Arg(0)
	if path == "-" {
		in, err = io.ReadAll(os.Stdin)
	} else {
		in, err = os.ReadFile(flag.Arg(0))
	}
	if err != nil {
		lib.Err(lib.ExitFailure, err, "couldn't read file")
	}

	p, _ := pem.Decode(in)
	if p == nil {
		lib.Errx(lib.ExitFailure, "%s isn't a PEM-encoded file", flag.Arg(0))
	}
	if _, err = os.Stdout.Write(p.Bytes); err != nil {
		lib.Err(lib.ExitFailure, err, "writing body")
	}
}
