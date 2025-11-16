package main

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"

	"git.wntrmute.dev/kyle/goutils/ahash"
	"git.wntrmute.dev/kyle/goutils/die"
	"git.wntrmute.dev/kyle/goutils/lib"
)

func usage(w io.Writer) {
	fmt.Fprintf(w, `Usage: %s [-a algo] [-h] [-l set] urls...
Compute the hash over each URL.

Flags:
	-a algo		Specify the hash algorithm to use; the default is sha256.
	-h		Print this help message.
	-l set		List the hash functions under set. Set can be one of all,
			secure to list only cryptographic hash functions, or
			insecure to list only non-cryptographic hash functions.
	
`, lib.ProgName())
}

func init() {
	flag.Usage = func() { usage(os.Stderr) }
}

func main() {
	var algo, list string
	var help bool
	flag.StringVar(&algo, "a", "sha256", "hash algorithm to use")
	flag.BoolVar(&help, "h", false, "print a help message")
	flag.StringVar(&list, "l", "", "list known hash algorithms (one of all, secure, insecure)")
	flag.Parse()

	if help {
		usage(os.Stdout)
	}

	if list != "" {
		var hashes []string
		switch list {
		case "all":
			hashes = ahash.HashList()
		case "secure":
			hashes = ahash.SecureHashList()
		case "insecure":
			hashes = ahash.InsecureHashList()
		default:
			die.With("list option must be one of all, secure, or insecure.")
		}

		for _, algo := range hashes {
			fmt.Printf("- %s\n", algo)
		}
		os.Exit(1)
	}

	for _, remote := range flag.Args() {
		u, err := url.Parse(remote)
		if err != nil {
			_, _ = lib.Warn(err, "parsing %s", remote)
			continue
		}

		name := filepath.Base(u.Path)
		if name == "" {
			_, _ = lib.Warnx("source URL doesn't appear to name a file")
			continue
		}

		resp, err := http.Get(remote)
		if err != nil {
			_, _ = lib.Warn(err, "fetching %s", remote)
			continue
		}

		sum, err := ahash.SumReader(algo, resp.Body)
		resp.Body.Close()
		if err != nil {
			lib.Err(lib.ExitFailure, err, "while hashing data")
		}
		fmt.Printf("%s: %s=%x\n", name, algo, sum)
	}
}
