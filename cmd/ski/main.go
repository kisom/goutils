package main

import (

	// #nosec G505

	"flag"
	"fmt"
	"io"
	"os"

	"git.wntrmute.dev/kyle/goutils/certlib/ski"
	"git.wntrmute.dev/kyle/goutils/die"
	"git.wntrmute.dev/kyle/goutils/lib"
)

func usage(w io.Writer) {
	fmt.Fprintf(w, `ski: print subject key info for PEM-encoded files

Usage:
	ski [-hm] files...

Flags:
	-d  Hex encoding mode.
	-h	Print this help message.
	-m	All SKIs should match; as soon as an SKI mismatch is found,
		it is reported.
`)
}

func init() {
	flag.Usage = func() { usage(os.Stderr) }
}

func main() {
	var help, shouldMatch bool
	var displayModeString string
	flag.StringVar(&displayModeString, "d", "lower", "hex encoding mode")
	flag.BoolVar(&help, "h", false, "print a help message and exit")
	flag.BoolVar(&shouldMatch, "m", false, "all SKIs should match")
	flag.Parse()

	displayMode := lib.ParseHexEncodeMode(displayModeString)

	if help {
		usage(os.Stdout)
		os.Exit(0)
	}

	var matchSKI string
	for _, path := range flag.Args() {
		keyInfo, err := ski.ParsePEM(path)
		die.If(err)

		keySKI, err := keyInfo.SKI(displayMode)
		die.If(err)

		if matchSKI == "" {
			matchSKI = keySKI
		}

		if shouldMatch && matchSKI != keySKI {
			_, _ = lib.Warnx("%s: SKI mismatch (%s != %s)",
				path, matchSKI, keySKI)
		}
		fmt.Printf("%s  %s (%s %s)\n", path, keySKI, keyInfo.KeyType, keyInfo.FileType)
	}
}
