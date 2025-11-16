package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"git.wntrmute.dev/kyle/goutils/die"
)

func usage(w io.Writer, exc int) {
	fmt.Fprintln(w, `usage: dumpbytes -n tabs <file>`)
	os.Exit(exc)
}

func printBytes(buf []byte) {
	fmt.Printf("\t")
	for i := range buf {
		fmt.Printf("0x%02x, ", buf[i])
	}
	fmt.Println()
}

func dumpFile(path string, indentLevel int) error {
	var indent strings.Builder
	for range indentLevel {
		indent.WriteByte('\t')
	}

	file, err := os.Open(path)
	if err != nil {
		return err
	}

	defer file.Close()

	fmt.Printf("%svar buffer = []byte{\n", indent.String())
	var n int
	for {
		buf := make([]byte, 8)
		n, err = file.Read(buf)
		if errors.Is(err, io.EOF) {
			if n > 0 {
				fmt.Printf("%s", indent.String())
				printBytes(buf[:n])
			}
			break
		}

		if err != nil {
			return err
		}

		fmt.Printf("%s", indent.String())
		printBytes(buf[:n])
	}

	fmt.Printf("%s}\n", indent.String())
	return nil
}

func main() {
	indent := 0
	flag.Usage = func() { usage(os.Stderr, 0) }
	flag.IntVar(&indent, "n", 0, "indent level")
	flag.Parse()

	for _, file := range flag.Args() {
		err := dumpFile(file, indent)
		die.If(err)
	}
}
