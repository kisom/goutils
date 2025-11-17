package main

import (
	"crypto/x509"
	"encoding/hex"
	"flag"
	"fmt"
	"strings"

	"git.wntrmute.dev/kyle/goutils/certlib"
	"git.wntrmute.dev/kyle/goutils/die"
)

const (
	displayInt = iota + 1
	displayLHex
	displayUHex
)

func parseDisplayMode(mode string) int {
	mode = strings.ToLower(mode)
	switch mode {
	case "int":
		return displayInt
	case "hex":
		return displayLHex
	case "uhex":
		return displayUHex
	default:
		die.With("invalid display mode ", mode)
	}

	return displayInt
}

func serialString(cert *x509.Certificate, mode int) string {
	switch mode {
	case displayInt:
		return cert.SerialNumber.String()
	case displayLHex:
		return hex.EncodeToString(cert.SerialNumber.Bytes())
	case displayUHex:
		return strings.ToUpper(hex.EncodeToString(cert.SerialNumber.Bytes()))
	default:
		return cert.SerialNumber.String()
	}
}

func main() {
	displayAs := flag.String("d", "int", "display mode (int, hex, uhex)")
	flag.Parse()

	displayMode := parseDisplayMode(*displayAs)

	for _, arg := range flag.Args() {
		cert, err := certlib.LoadCertificate(arg)
		die.If(err)

		fmt.Printf("%s: %x\n", arg, serialString(cert, displayMode))
	}
}
