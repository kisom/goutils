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

	"git.wntrmute.dev/kyle/goutils/lib"
)

const defaultDirectory = ".git/objects"

// maxDecompressedSize limits how many bytes we will decompress from a zlib
// stream to mitigate decompression bombs (gosec G110).
// Increase this if you expect larger objects.
const maxDecompressedSize int64 = 64 << 30 // 64 GiB

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

	// Protect against decompression bombs by limiting how much we read.
	lr := io.LimitReader(zread, maxDecompressedSize+1)
	if _, err = buf.ReadFrom(lr); err != nil {
		return nil, err
	}
	if int64(buf.Len()) > maxDecompressedSize {
		return nil, fmt.Errorf("decompressed size exceeds limit (%d bytes)", maxDecompressedSize)
	}
	return buf.Bytes(), nil
}

func showFile(path string) {
	fileData, err := loadFile(path)
	if err != nil {
		lib.Warn(err, "failed to load %s", path)
		return
	}

	fmt.Printf("%s\n", fileData)
}

func searchFile(path string, search *regexp.Regexp) error {
	file, err := os.Open(path)
	if err != nil {
		lib.Warn(err, "failed to open %s", path)
		return err
	}
	defer file.Close()

	zread, err := zlib.NewReader(file)
	if err != nil {
		lib.Warn(err, "failed to decompress %s", path)
		return err
	}
	defer zread.Close()

	// Limit how much we scan to avoid DoS via huge decompression.
	lr := io.LimitReader(zread, maxDecompressedSize+1)
	zbuf := bufio.NewReader(lr)
	if !search.MatchReader(zbuf) {
		return nil
	}

	fileData, err := loadFile(path)
	if err != nil {
		lib.Warn(err, "failed to load %s", path)
		return err
	}
	fmt.Printf("%s:\n%s\n", path, fileData)
	return nil
}

func buildWalker(searchExpr *regexp.Regexp) filepath.WalkFunc {
	return func(path string, info os.FileInfo, _ error) error {
		if !info.Mode().IsRegular() {
			return nil
		}
		return searchFile(path, searchExpr)
	}
}

// runSearch compiles the search expression and processes the provided paths.
// It returns an error for fatal conditions; per-file errors are logged.
func runSearch(expr string) error {
	search, err := regexp.Compile(expr)
	if err != nil {
		return fmt.Errorf("invalid regexp: %w", err)
	}

	pathList := flag.Args()
	if len(pathList) == 0 {
		pathList = []string{defaultDirectory}
	}

	for _, path := range pathList {
		if isDir(path) {
			if err2 := filepath.Walk(path, buildWalker(search)); err2 != nil {
				return err2
			}
			continue
		}
		if err2 := searchFile(path, search); err2 != nil {
			// Non-fatal: keep going, but report it.
			lib.Warn(err2, "non-fatal error while searching files")
		}
	}
	return nil
}

func main() {
	flSearch := flag.String("s", "", "search string (should be an RE2 regular expression)")
	flag.Parse()

	if *flSearch == "" {
		for _, path := range flag.Args() {
			showFile(path)
		}
		return
	}

	if err := runSearch(*flSearch); err != nil {
		lib.Err(lib.ExitFailure, err, "failed to run search")
	}
}
