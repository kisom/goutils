package main

import (
	"compress/flate"
	"compress/gzip"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const gzipExt = ".gz"

func compress(path, target string, level int) error {
	sourceFile, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("opening file for read: %w", err)
	}
	defer sourceFile.Close()

	destFile, err := os.Create(target)
	if err != nil {
		return fmt.Errorf("opening file for write: %w", err)
	}
	defer destFile.Close()

	gzipCompressor, err := gzip.NewWriterLevel(destFile, level)
	if err != nil {
		return fmt.Errorf("invalid compression level: %w", err)
	}
	defer gzipCompressor.Close()

	_, err = io.Copy(gzipCompressor, sourceFile)
	if err != nil {
		return fmt.Errorf("compressing file: %w", err)
	}

	return nil
}

func uncompress(path, target string, unrestrict bool) error {
	sourceFile, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("opening file for read: %w", err)
	}
	defer sourceFile.Close()

	fi, err := sourceFile.Stat()
	if err != nil {
		return fmt.Errorf("reading file stats: %w", err)
	}

	maxDecompressionSize := fi.Size() * 32

	gzipUncompressor, err := gzip.NewReader(sourceFile)
	if err != nil {
		return fmt.Errorf("reading gzip headers: %w", err)
	}
	defer gzipUncompressor.Close()

	var reader io.Reader = &io.LimitedReader{
		R: gzipUncompressor,
		N: maxDecompressionSize,
	}

	if unrestrict {
		reader = gzipUncompressor
	}

	destFile, err := os.Create(target)
	if err != nil {
		return fmt.Errorf("opening file for write: %w", err)
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, reader)
	if err != nil {
		return fmt.Errorf("uncompressing file: %w", err)
	}

	return nil
}

func usage(w io.Writer) {
	fmt.Fprintf(w, `Usage: %s [-l] source [target]

kgz is like gzip, but supports compressing and decompressing to a different
directory than the source file is in.

Flags:
	-l level	Compression level (0-9). Only meaninful when
			compressing a file.
`, os.Args[0])
}

func init() {
	flag.Usage = func() { usage(os.Stderr) }
}

func isDir(path string) bool {
	file, err := os.Open(path)
	if err == nil {
		defer file.Close()
		stat, err2 := file.Stat()
		if err2 != nil {
			return false
		}

		if stat.IsDir() {
			return true
		}
	}

	return false
}

func pathForUncompressing(source, dest string) (string, error) {
	if !isDir(dest) {
		return dest, nil
	}

	source = filepath.Base(source)
	if !strings.HasSuffix(source, gzipExt) {
		return "", fmt.Errorf("%s is a not gzip-compressed file", source)
	}
	outFile := source[:len(source)-len(gzipExt)]
	outFile = filepath.Join(dest, outFile)
	return outFile, nil
}

func pathForCompressing(source, dest string) (string, error) {
	if !isDir(dest) {
		return dest, nil
	}

	source = filepath.Base(source)
	if strings.HasSuffix(source, gzipExt) {
		return "", fmt.Errorf("%s is a gzip-compressed file", source)
	}

	dest = filepath.Join(dest, source+gzipExt)
	return dest, nil
}

func main() {
	var level int
	var path string
	var target = "."
	var err error
	var unrestrict bool

	flag.IntVar(&level, "l", flate.DefaultCompression, "compression level")
	flag.BoolVar(&unrestrict, "u", false, "do not restrict decompression")
	flag.Parse()

	if flag.NArg() < 1 || flag.NArg() > 2 {
		usage(os.Stderr)
		os.Exit(1)
	}

	path = flag.Arg(0)
	if flag.NArg() == 2 {
		target = flag.Arg(1)
	}

	if strings.HasSuffix(path, gzipExt) {
		target, err = pathForUncompressing(path, target)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err)
			os.Exit(1)
		}

		err = uncompress(path, target, unrestrict)
		if err != nil {
			os.Remove(target)
			fmt.Fprintf(os.Stderr, "%s\n", err)
			os.Exit(1)
		}
		return
	}

	target, err = pathForCompressing(path, target)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}

	err = compress(path, target, level)
	if err != nil {
		os.Remove(target)
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}
