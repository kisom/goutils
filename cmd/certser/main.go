package main

import (
	"crypto/x509"
	"flag"
	"fmt"
	"strings"

	"git.wntrmute.dev/kyle/goutils/certlib"
	"git.wntrmute.dev/kyle/goutils/die"
	"git.wntrmute.dev/kyle/goutils/lib"
)

const displayInt lib.HexEncodeMode = iota

func parseDisplayMode(mode string) lib.HexEncodeMode {
	mode = strings.ToLower(mode)

	if mode == "int" {
		return displayInt
	}

	return lib.ParseHexEncodeMode(mode)
}

func serialString(cert *x509.Certificate, mode lib.HexEncodeMode) string {
	if mode == displayInt {
		return cert.SerialNumber.String()
	}

	return lib.HexEncode(cert.SerialNumber.Bytes(), mode)
}

func main() {
	displayAs := flag.String("d", "int", "display mode (int, hex, uhex)")
	showExpiry := flag.Bool("e", false, "show expiry date")
	flag.Parse()

	displayMode := parseDisplayMode(*displayAs)

	for _, arg := range flag.Args() {
		cert, err := certlib.LoadCertificate(arg)
		die.If(err)

		fmt.Printf("%s: %s", arg, serialString(cert, displayMode))
		if *showExpiry {
			fmt.Printf(" (%s)", cert.NotAfter.Format("2006-01-02"))
		}
		fmt.Println()
	}
}
