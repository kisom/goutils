package main

// Prompt:
// The current main.go should accept a list of paths to search. In each
// of those paths, without recursing, it should find all files ending in
// C/C++ source extensions and print them one per line.

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

var extensions = []string{
	".c", ".cpp", ".cc", ".cxx",
	".h", ".hpp", ".hh", ".hxx",
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <path> [path...]\n", os.Args[0])
		os.Exit(1)
	}

	for _, path := range os.Args[1:] {
		entries, err := os.ReadDir(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading %s: %v\n", path, err)
			continue
		}

		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}

			name := entry.Name()
			ext := filepath.Ext(name)
			if slices.Contains(extensions, strings.ToLower(ext)) {
				fmt.Println(filepath.Join(path, name))
			}
		}
	}
}
