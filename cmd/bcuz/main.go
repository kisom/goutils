package main

import (
	"archive/zip"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

var keepArchive bool

func removedir(dir string, existed bool) {
	if !existed {
		os.RemoveAll(dir)
	}
}

func unpackFile(path string) error {
	var dir string
	var existed bool

	fmt.Printf("[+] processing %s:\n", path)

	base := filepath.Base(path[:len(path)-4])
	pieces := strings.SplitN(base, "-", 2)
	if len(pieces) == 2 {
		artist := strings.TrimSpace(pieces[0])
		album := strings.TrimSpace(pieces[1])
		dir = filepath.Join(artist, album)
	} else {
		dir = base
	}

	_, err := os.Stat(dir)
	if err == nil {
		existed = true
	}

	fmt.Printf("\tunpack directory: %s\n", dir)
	err = os.MkdirAll(dir, 0755)
	if err != nil {
		return err
	}

	r, err := zip.OpenReader(path)
	if err != nil {
		removedir(dir, existed)
		return err
	}
	defer r.Close()

	var rc io.ReadCloser
	for _, f := range r.File {
		fmt.Printf("\tunpacking %s\n", f.FileHeader.Name)
		rc, err = f.Open()
		if err != nil {
			rc.Close()
			removedir(dir, existed)
			return err
		}

		out, err := os.Create(filepath.Join(dir, f.FileHeader.Name))
		if err != nil {
			rc.Close()
			removedir(dir, existed)
			return err
		}

		_, err = io.Copy(out, rc)
		if err != nil {
			rc.Close()
			removedir(dir, existed)
			return err
		}

		out.Close()
		rc.Close()
	}

	if !keepArchive {
		return os.Remove(path)
	}
	return nil
}

func main() {
	flag.BoolVar(&keepArchive, "k", false, "don't remove the archive file after unpacking")
	flag.Parse()

	for _, path := range flag.Args() {
		err := unpackFile(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "[!] failed to process %s: %s\n", path, err)
		}
	}
}
