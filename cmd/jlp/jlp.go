package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"

	"github.com/kisom/goutils/lib"
)

func prettify(file string) error {
	in, err := ioutil.ReadFile(file)
	if err != nil {
		lib.Warn(err, "ReadFile")
		return err
	}

	var buf = &bytes.Buffer{}
	err = json.Indent(buf, in, "", "    ")
	if err != nil {
		lib.Warn(err, "%s", file)
		return err
	}

	err = ioutil.WriteFile(file, buf.Bytes(), 0644)
	if err != nil {
		lib.Warn(err, "WriteFile")
	}

	return err
}

func compact(file string) error {
	in, err := ioutil.ReadFile(file)
	if err != nil {
		lib.Warn(err, "ReadFile")
		return err
	}

	var buf = &bytes.Buffer{}
	err = json.Compact(buf, in)
	if err != nil {
		lib.Warn(err, "%s", file)
		return err
	}

	err = ioutil.WriteFile(file, buf.Bytes(), 0644)
	if err != nil {
		lib.Warn(err, "WriteFile")
	}

	return err
}

func usage() {
	progname := lib.ProgName()
	fmt.Printf(`Usage: %s [-h] files...
	%s is used to lint and prettify (or compact) JSON files. The
	files will be updated in-place.

	Flags:
	-c	Compact files.
	-h	Print this help message.
`, progname, progname)

}

func init() {
	flag.Usage = usage
}

func main() {
	var shouldCompact bool
	flag.BoolVar(&shouldCompact, "c", false, "Compact files instead of prettifying.")
	flag.Parse()

	action := prettify
	if shouldCompact {
		action = compact
	}

	var errCount int
	for _, fileName := range flag.Args() {
		err := action(fileName)
		if err != nil {
			errCount++
		}
	}

	if errCount > 0 {
		lib.Errx(lib.ExitFailure, "Not all files succeeded.")
	}
}
