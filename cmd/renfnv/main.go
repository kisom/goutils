package main

import (
	"encoding/base32"
	"encoding/binary"
	"flag"
	"fmt"
	"hash/fnv"
	"io"
	"os"
	"path/filepath"
	"strings"

	"git.wntrmute.dev/kyle/goutils/fileutil"
	"git.wntrmute.dev/kyle/goutils/lib"
)

func hashName(path, encodedHash string) string {
	basename := filepath.Base(path)
	location := filepath.Dir(path)
	ext := filepath.Ext(basename)
	return filepath.Join(location, encodedHash+ext)
}

func newName(path string) (string, error) {
	h := fnv.New32a()

	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	_, err = io.Copy(h, f)
	if err != nil {
		return "", err
	}

	var buf [8]byte
	binary.BigEndian.PutUint32(buf[:], h.Sum32())
	encodedHash := base32.StdEncoding.EncodeToString(h.Sum(nil))
	encodedHash = strings.TrimRight(encodedHash, "=")
	return hashName(path, encodedHash), nil
}

func move(dst, src string, force bool) error {
	if fileutil.FileDoesExist(dst) && !force {
		return fmt.Errorf("%s exists (pass the -f flag to overwrite)", dst)
	}
	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}

	var retErr error
	defer func(e *error) {
		dstFile.Close()
		if *e != nil {
			os.Remove(dst)
		}
	}(&retErr)

	srcFile, err := os.Open(src)
	if err != nil {
		retErr = err
		return err
	}
	defer srcFile.Close()

	if _, err = io.Copy(dstFile, srcFile); err != nil {
		retErr = err
		return err
	}

	os.Remove(src)
	return nil
}

func usage(w io.Writer) {
	fmt.Fprintf(w, `Usage: renfnv [-fhlnv] files...

renfnv renames files to the base32-encoded 32-bit FNV-1a hash of their
contents, preserving the dirname and extension.

Options:
	-f	force overwriting of files when there is a collision.
	-h	print this help message.
	-l	list changed files.
	-n	Perform a dry run: don't actually move files.
	-v	Print all files as they are processed. If both -v and -l
		are specified, it will behave as if only -v was specified.
`)
}

func init() {
	flag.Usage = func() { usage(os.Stdout) }
}

type options struct {
	dryRun, force, printChanged, verbose bool
}

func processOne(file string, opt options) error {
	renamed, err := newName(file)
	if err != nil {
		_, _ = lib.Warn(err, "failed to get new file name")
		return err
	}
	if opt.verbose && !opt.printChanged {
		fmt.Fprintln(os.Stdout, file)
	}
	if renamed == file {
		return nil
	}
	if !opt.dryRun {
		if err = move(renamed, file, opt.force); err != nil {
			_, _ = lib.Warn(err, "failed to rename file from %s to %s", file, renamed)
			return err
		}
	}
	if opt.printChanged && !opt.verbose {
		fmt.Fprintln(os.Stdout, file, "->", renamed)
	}
	return nil
}

func run(dryRun, force, printChanged, verbose bool, files []string) {
	if verbose && printChanged {
		printChanged = false
	}
	opt := options{dryRun: dryRun, force: force, printChanged: printChanged, verbose: verbose}
	for _, file := range files {
		_ = processOne(file, opt)
	}
}

func main() {
	var dryRun, force, printChanged, verbose bool
	flag.BoolVar(&force, "f", false, "force overwriting of files if there is a collision")
	flag.BoolVar(&printChanged, "l", false, "list changed files")
	flag.BoolVar(&dryRun, "n", false, "dry run --- don't perform moves")
	flag.BoolVar(&verbose, "v", false, "list all processed files")

	flag.Parse()
	run(dryRun, force, printChanged, verbose, flag.Args())
}
