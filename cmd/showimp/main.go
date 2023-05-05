// showimp is a utility for displaying the imports in a package.
package main

import (
	"flag"
	"fmt"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"git.wntrmute.dev/kyle/goutils/dbg"
	"git.wntrmute.dev/kyle/goutils/die"
)

var (
	gopath  string
	project string
)

var (
	debug        = dbg.New()
	fset         = &token.FileSet{}
	imports      = map[string]bool{}
	sourceRegexp = regexp.MustCompile(`^[^.].*\.go$`)
	stdLibRegexp = regexp.MustCompile(`^\w+(/\w+)*$`)
)

func init() {
	gopath = os.Getenv("GOPATH")
	if gopath == "" {
		fmt.Fprintf(os.Stderr, "GOPATH isn't set, can't proceed.")
		os.Exit(1)
	}
	gopath += "/src/"

	wd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to establish working directory: %v", err)
		os.Exit(1)
	}

	if !strings.HasPrefix(wd, gopath) {
		fmt.Fprintf(os.Stderr, "Can't determine my location in the GOPATH.\n")
		fmt.Fprintf(os.Stderr, "Working directory is %s\n", wd)
		fmt.Fprintf(os.Stderr, "Go source path is %s\n", gopath)
		os.Exit(1)
	}

	project = wd[len(gopath):]
}

func walkFile(path string, info os.FileInfo, err error) error {
	if ignores[path] {
		return filepath.SkipDir
	}

	if !sourceRegexp.MatchString(path) {
		return nil
	}

	debug.Println(path)

	f, err := parser.ParseFile(fset, path, nil, parser.ImportsOnly)
	if err != nil {
		return err
	}

	for _, importSpec := range f.Imports {
		importPath := strings.Trim(importSpec.Path.Value, `"`)
		if stdLibRegexp.MatchString(importPath) {
			debug.Println("standard lib:", importPath)
			continue
		} else if strings.HasPrefix(importPath, project) {
			debug.Println("internal import:", importPath)
			continue
		} else if strings.HasPrefix(importPath, "golang.org/") {
			debug.Println("extended lib:", importPath)
			continue
		}
		debug.Println("import:", importPath)
		imports[importPath] = true
	}

	return nil
}

var ignores = map[string]bool{}

func main() {
	var ignoreLine string
	var noVendor bool
	flag.StringVar(&ignoreLine, "i", "", "comma-separated list of directories to ignore")
	flag.BoolVar(&noVendor, "nv", false, "ignore the vendor directory")
	flag.BoolVar(&debug.Enabled, "v", false, "log debugging information")
	flag.Parse()

	if noVendor {
		ignores["vendor"] = true
	}

	for _, word := range strings.Split(ignoreLine, ",") {
		ignores[strings.TrimSpace(word)] = true
	}

	err := filepath.Walk(".", walkFile)
	die.If(err)

	fmt.Println("External imports:")
	importList := make([]string, 0, len(imports))
	for imp := range imports {
		importList = append(importList, imp)
	}
	sort.Strings(importList)

	for _, imp := range importList {
		fmt.Println("\t", imp)
	}
}
