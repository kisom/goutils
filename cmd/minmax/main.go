package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
)

var kinds = map[string]int{
	"sym": 0,
	"tf":  1,
	"fn":  2,
	"sp":  3,
}

func dieIf(err error) {
	if err == nil {
		return
	}
	fmt.Fprintf(os.Stderr, "[!] %s\n", err)
	os.Exit(1)
}

func usage() {
	fmt.Fprintf(os.Stderr, "usage: minmax type min max\n")
	fmt.Fprintf(os.Stderr, "    type is one of fn, sp, sym, tf\n")
	os.Exit(1)
}

func main() {
	flag.Parse()

	if flag.NArg() != 3 {
		usage()
	}

	kind, ok := kinds[flag.Arg(0)]
	if !ok {
		usage()
	}

	minVal, err := strconv.Atoi(flag.Arg(1))
	dieIf(err)

	maxVal, err := strconv.Atoi(flag.Arg(2))
	dieIf(err)

	code := kind << 6
	code += (minVal << 3)
	code += maxVal
	fmt.Fprintf(os.Stdout, "%0o\n", code)
}
