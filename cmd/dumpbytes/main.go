package main

import (
	"flag"
	"fmt"
	"git.wntrmute.dev/kyle/goutils/die"
	"io"
	"os"
)

func usage(w io.Writer, exc int) {
	fmt.Fprintln(w, `usage: dumpbytes <file>`)
	os.Exit(exc)
}

func printBytes(buf []byte) {
	fmt.Printf("\t")
	for i := 0; i < len(buf); i++ {
		fmt.Printf("0x%02x, ", buf[i])
	}
	fmt.Println()
}

func dumpFile(path string, indentLevel int) error {
	indent := ""
	for i := 0; i < indentLevel; i++ {
		indent += "\t"
	}

	file, err := os.Open(path)
	if err != nil {
		return err
	}

	defer file.Close()

	fmt.Printf("%svar buffer = []byte{\n", indent)
	for {
		buf := make([]byte, 8)
		n, err := file.Read(buf)
		if err == io.EOF {
			if n > 0 {
				fmt.Printf("%s", indent)
				printBytes(buf[:n])
			}
			break
		}

		if err != nil {
			return err
		}

		fmt.Printf("%s", indent)
		printBytes(buf[:n])
	}

	fmt.Printf("%s}\n", indent)
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
