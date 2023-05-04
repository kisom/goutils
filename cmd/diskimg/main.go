package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"

	"git.wntrmute.dev/kyle/goutils/ahash"
	"git.wntrmute.dev/kyle/goutils/dbg"
	"git.wntrmute.dev/kyle/goutils/die"
)

const defaultHashAlgorithm = "sha256"

var (
	hAlgo      string
	debug = dbg.New()
)


func openImage(imageFile string) (image *os.File, hash []byte, err error) {
	image, err = os.Open(imageFile)
	if err != nil {
		return
	}

	hash, err = ahash.SumReader(hAlgo, image)
	if err != nil {
		return
	}

	_, err = image.Seek(0, 0)
	if err != nil {
		return
	}

	debug.Printf("%s  %x\n", imageFile, hash)
	return
}

func openDevice(devicePath string) (device *os.File, err error) {
	fi, err := os.Stat(devicePath)
	if err != nil {
		return
	}

	device, err = os.OpenFile(devicePath, os.O_RDWR|os.O_SYNC, fi.Mode())
	if err != nil {
		return
	}

	return
}

func main() {
	flag.StringVar(&hAlgo, "a", defaultHashAlgorithm, "default hash algorithm")
	flag.BoolVar(&debug.Enabled, "v", false, "enable debug logging")
	flag.Parse()

	if hAlgo == "list" {
		fmt.Println("Supported hashing algorithms:")
		for _, algo := range ahash.SecureHashList() {
			fmt.Printf("\t- %s\n", algo)
		}
		os.Exit(2)
	}

	if flag.NArg() != 2 {
		die.With("usage: diskimg image device")
	}

	imageFile := flag.Arg(0)
	devicePath := flag.Arg(1)

	debug.Printf("opening image %s for read\n", imageFile)
	image, hash, err := openImage(imageFile)
	if image != nil {
		defer image.Close()
	}
	die.If(err)

	debug.Printf("opening device %s for rw\n", devicePath)
	device, err := openDevice(devicePath)
	if device != nil {
		defer device.Close()
	}
	die.If(err)

	debug.Printf("writing %s -> %s\n", imageFile, devicePath)
	n, err := io.Copy(device, image)
	die.If(err)
	debug.Printf("wrote %d bytes to %s\n", n, devicePath)

	debug.Printf("syncing %s\n", devicePath)
	err = device.Sync()
	die.If(err)

	debug.Println("verifying the image was written successfully")
	_, err = device.Seek(0, 0)
	die.If(err)

	deviceHash, err := ahash.SumLimitedReader(hAlgo, device, n)
	die.If(err)

	if !bytes.Equal(deviceHash, hash) {
		fmt.Fprintln(os.Stderr, "Hash mismatch:")
		fmt.Fprintf(os.Stderr, "\t%s: %s\n", imageFile, hash)
		fmt.Fprintf(os.Stderr, "\t%s: %s\n", devicePath, deviceHash)
		os.Exit(1)
	}

	debug.Println("OK")
	os.Exit(0)
}
