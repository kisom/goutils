package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"git.wntrmute.dev/kyle/goutils/die"
)

var reUUID = regexp.MustCompile(`^\w{8}-\w{4}-\w{4}-\w{4}-\w{12}_(.+)$`)

func renamePath(path string, dryRun bool) error {
	dir, base := filepath.Split(path)

	base = reUUID.ReplaceAllString(base, "$1")
	newPath := filepath.Join(dir, base)

	if dryRun {
		fmt.Println(path, "->", newPath)
		return nil
	}

	err := os.Rename(path, newPath)
	if err != nil {
		return fmt.Errorf("renaming %s to %s failed: %v", path, newPath, err)
	}

	return nil
}

func test() bool {
	const testFilePath = "48793683-8568-47c2-9e2d-eecab3c4b639_Whispers of Chernobog.pdf"
	const expected = "Whispers of Chernobog.pdf"

	actual := reUUID.ReplaceAllString(testFilePath, "$1")
	return actual == expected
}

func main() {
	var err error

	if !test() {
		die.With("test failed")
	}

	dryRun := false
	flag.BoolVar(&dryRun, "n", dryRun, "don't rename files, just print what would be done")
	flag.Parse()

	paths := flag.Args()
	if len(paths) == 0 {
		paths, err = filepath.Glob("*")
		die.If(err)
	}

	for _, file := range paths {
		err = renamePath(file, dryRun)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: %v\n", file, err)
		}
	}
}
