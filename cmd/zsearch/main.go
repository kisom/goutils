// zsearch is a utility for searching zlib-compressed files for a
// search string. It was really designed for use with the Git object
// store, i.e. to aid in the recovery of files after Git does what Git
// do.
package main

import (
	"bufio"
	"bytes"
	"compress/zlib"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
)

const defaultDirectory = ".git/objects"

func errorf(format string, a ...interface{}) {
	fmt.Fprintf(os.Stderr, format, a...)
	if format[len(format)-1] != '\n' {
		fmt.Fprintf(os.Stderr, "\n")
	}
}

func isDir(path string) bool {
	fi, err := os.Stat(path)
	if err != nil {
		return false
	}

	return fi.IsDir()
}

func loadFile(path string) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	buf := new(bytes.Buffer)
	zread, err := zlib.NewReader(file)
	if err != nil {
		return nil, err
	}
	defer zread.Close()

	_, err = io.Copy(buf, zread)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func showFile(path string) {
	fileData, err := loadFile(path)
	if err != nil {
		errorf("%v", err)
		return
	}

	fmt.Printf("%s\n", fileData)
}

func searchFile(path string, search *regexp.Regexp) error {
	file, err := os.Open(path)
	if err != nil {
		errorf("%v", err)
		return err
	}
	defer file.Close()

	zread, err := zlib.NewReader(file)
	if err != nil {
		errorf("%v", err)
		return err
	}
	defer zread.Close()

	zbuf := bufio.NewReader(zread)
	if search.MatchReader(zbuf) {
		fileData, err := loadFile(path)
		if err != nil {
			errorf("%v", err)
			return err
		}
		fmt.Printf("%s:\n%s\n", path, fileData)
	}
	return nil
}

func buildWalker(searchExpr *regexp.Regexp) filepath.WalkFunc {
	return func(path string, info os.FileInfo, err error) error {
		if info.Mode().IsRegular() {
			return searchFile(path, searchExpr)
		}
		return nil
	}
}

func main() {
	flSearch := flag.String("s", "", "search string (should be an RE2 regular expression)")
	flag.Parse()

	if *flSearch == "" {
		for _, path := range flag.Args() {
			showFile(path)
		}
	} else {
		search, err := regexp.Compile(*flSearch)
		if err != nil {
			errorf("Bad regexp: %v", err)
			return
		}

		pathList := flag.Args()
		if len(pathList) == 0 {
			pathList = []string{defaultDirectory}
		}

		for _, path := range pathList {
			if isDir(path) {
				err := filepath.Walk(path, buildWalker(search))
				if err != nil {
					errorf("%v", err)
					return
				}
			} else {
				searchFile(path, search)
			}
		}
	}
}
