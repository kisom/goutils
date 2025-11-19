package main

import (
	"flag"
	"fmt"
	"os"

	"git.wntrmute.dev/kyle/goutils/certlib"
	"git.wntrmute.dev/kyle/goutils/die"
)

// functionality refactored into certlib

func main() {
	var keyFile, certFile string
	flag.StringVar(&keyFile, "k", "", "TLS private `key` file")
	flag.StringVar(&certFile, "c", "", "TLS `certificate` file")
	flag.Parse()

	cert, err := certlib.LoadCertificate(certFile)
	die.If(err)

	priv, err := certlib.LoadPrivateKey(keyFile)
	die.If(err)

	matched, reason := certlib.MatchKeys(cert, priv)
	if matched {
		fmt.Println("Match.")
		return
	}
	fmt.Printf("No match (%s).\n", reason)
	os.Exit(1)
}
