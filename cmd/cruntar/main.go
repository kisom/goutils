package main

import (
	"archive/tar"
	"compress/bzip2"
	"compress/gzip"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"git.wntrmute.dev/kyle/goutils/die"
	"git.wntrmute.dev/kyle/goutils/fileutil"
)

var (
	preserveOwners bool
	preserveMode   bool
	verbose        bool
)

func setupFile(hdr *tar.Header, file *os.File) error {
	if preserveMode {
		if verbose {
			fmt.Printf("\tchmod %0#o\n", hdr.Mode)
		}
		err := file.Chmod(os.FileMode(hdr.Mode))
		if err != nil {
			return err
		}
	}

	if preserveOwners {
		fmt.Printf("\tchown %d:%d\n", hdr.Uid, hdr.Gid)
		err := file.Chown(hdr.Uid, hdr.Gid)
		if err != nil {
			return err
		}
	}

	return nil
}

func linkTarget(target, top string) string {
	if filepath.IsAbs(target) {
		return target
	}

	return filepath.Clean(filepath.Join(target, top))
}

func processFile(tfr *tar.Reader, hdr *tar.Header, top string) error {
	if verbose {
		fmt.Println(hdr.Name)
	}
	filePath := filepath.Clean(filepath.Join(top, hdr.Name))
	switch hdr.Typeflag {
	case tar.TypeReg:
		file, err := os.Create(filePath)
		if err != nil {
			return err
		}

		_, err = io.Copy(file, tfr)
		if err != nil {
			return err
		}

		err = setupFile(hdr, file)
		if err != nil {
			return err
		}
	case tar.TypeLink:
		file, err := os.Create(filePath)
		if err != nil {
			return err
		}

		source, err := os.Open(hdr.Linkname)
		if err != nil {
			return err
		}

		_, err = io.Copy(file, source)
		if err != nil {
			return err
		}

		err = setupFile(hdr, file)
		if err != nil {
			return err
		}
	case tar.TypeSymlink:
		if !fileutil.ValidateSymlink(hdr.Linkname, top) {
			return fmt.Errorf("symlink %s is outside the top-level %s",
				hdr.Linkname, top)
		}
		path := linkTarget(hdr.Linkname, top)
		if ok, err := filepath.Match(top+"/*", filepath.Clean(path)); !ok {
			return fmt.Errorf("symlink %s isn't in %s", hdr.Linkname, top)
		} else if err != nil {
			return err
		}

		err := os.Symlink(linkTarget(hdr.Linkname, top), filePath)
		if err != nil {
			return err
		}
	case tar.TypeDir:
		err := os.MkdirAll(filePath, os.FileMode(hdr.Mode))
		if err != nil {
			return err
		}
	}

	return nil
}

var compression = map[string]bool{
	"gzip":  false,
	"bzip2": false,
}

type bzipCloser struct {
	r io.Reader
}

func (brc *bzipCloser) Read(p []byte) (int, error) {
	return brc.r.Read(p)
}

func (brc *bzipCloser) Close() error {
	return nil
}

func newBzipCloser(r io.ReadCloser) (io.ReadCloser, error) {
	br := bzip2.NewReader(r)
	return &bzipCloser{r: br}, nil
}

var compressFuncs = map[string]func(io.ReadCloser) (io.ReadCloser, error){
	"gzip":  func(r io.ReadCloser) (io.ReadCloser, error) { return gzip.NewReader(r) },
	"bzip2": newBzipCloser,
}

func verifyCompression() bool {
	var compressed bool
	for _, v := range compression {
		if compressed && v {
			return false
		}
		compressed = compressed || v
	}
	return true
}

func getReader(r io.ReadCloser) (io.ReadCloser, error) {
	for c, v := range compression {
		if v {
			return compressFuncs[c](r)
		}
	}

	return r, nil
}

func openArchive(path string) (io.ReadCloser, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	r, err := getReader(file)
	if err != nil {
		return nil, err
	}

	return r, nil
}

var compressFlags struct {
	z bool
	j bool
}

func parseCompressFlags() error {
	if compressFlags.z {
		compression["gzip"] = true
	}

	if compressFlags.j {
		compression["bzip2"] = true
	}

	if !verifyCompression() {
		return errors.New("multiple compression formats specified")
	}

	return nil
}

func usage(w io.Writer) {
	fmt.Fprintf(w, `ChromeOS untar

This is a tool that is intended to support untarring on SquashFS file
systems. In particular, every time it encounters a hard link, it
will just create a copy of the file.

Usage: cruntar [-jmvpz] archive [dest]

Flags:
	-a	Shortcut for -m -p: preserve owners and file mode.
	-j	The archive is compressed with bzip2.
	-m	Preserve file modes.
	-p	Preserve ownership.
	-v	Print the name of each file as it is being processed.
	-z	The archive is compressed with gzip.
`)
}

func init() {
	flag.Usage = func() { usage(os.Stderr) }
}

func main() {
	var archive, help bool
	flag.BoolVar(&archive, "a", false, "Shortcut for -m -p: preserve owners and file mode.")
	flag.BoolVar(&help, "h", false, "print a help message")
	flag.BoolVar(&compressFlags.j, "j", false, "bzip2 compression")
	flag.BoolVar(&preserveMode, "m", false, "preserve file modes")
	flag.BoolVar(&preserveOwners, "p", false, "preserve ownership")
	flag.BoolVar(&verbose, "v", false, "verbose mode")
	flag.BoolVar(&compressFlags.z, "z", false, "gzip compression")
	flag.Parse()

	if help {
		usage(os.Stdout)
		os.Exit(0)
	}

	if archive {
		preserveMode = true
		preserveOwners = true
	}

	err := parseCompressFlags()
	die.If(err)

	if flag.NArg() == 0 {
		return
	}

	top := "./"
	if flag.NArg() > 1 {
		top = flag.Arg(1)
	}

	r, err := openArchive(flag.Arg(0))
	die.If(err)

	tfr := tar.NewReader(r)
	for {
		hdr, err := tfr.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		die.If(err)

		err = processFile(tfr, hdr, top)
		die.If(err)
	}

	r.Close()
}
