package main

import (
	_ "embed"
	"flag"
	"fmt"
	"os"

	"git.wntrmute.dev/kyle/goutils/certlib/bundler"
)

var (
	configFile string
	outputDir  string
)

//go:embed README.txt
var readmeContent string

func usage() {
	fmt.Fprint(os.Stderr, readmeContent)
}

func main() {
	flag.Usage = usage
	flag.StringVar(&configFile, "c", "bundle.yaml", "path to YAML configuration file")
	flag.StringVar(&outputDir, "o", "pkg", "output directory for archives")
	flag.Parse()

	if configFile == "" {
		fmt.Fprintf(os.Stderr, "Error: configuration file required (-c flag)\n")
		os.Exit(1)
	}

	if err := bundler.Run(configFile, outputDir); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Certificate bundling completed successfully")
}
