package main

import (
	"flag"
	"os"

	"git.wntrmute.dev/kyle/goutils/die"
)

// size of a kilobit in bytes.
const kilobit = 128
const pageSize = 4096

func main() {
	size := flag.Int("s", 256*kilobit, "size of EEPROM image in kilobits")
	fill := flag.Uint("f", 0, "byte to fill image with")
	flag.Parse()

	if *fill > 256 {
		die.With("`fill` argument must be a byte value")
	}

	path := "eeprom.img"

	if flag.NArg() > 0 {
		path = flag.Arg(0)
	}

	fillByte := uint8(*fill & 0xff) // #nosec G115 clearing out of bounds bits

	buf := make([]byte, pageSize)
	for i := range pageSize {
		buf[i] = fillByte
	}

	pages := *size / pageSize
	last := *size % pageSize

	file, err := os.Create(path)
	die.If(err)
	defer file.Close()

	for range pages {
		_, err = file.Write(buf)
		die.If(err)
	}

	if last != 0 {
		_, err = file.Write(buf[:last])
		die.If(err)
	}
}
