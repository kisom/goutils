package main

import (
    "crypto/x509"
    "encoding/pem"
    "flag"
    "fmt"
    "os"
)

func main() {
	flag.Parse()

	for _, fileName := range flag.Args() {
  data, err := os.ReadFile(fileName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "[!] %s: %v\n", fileName, err)
			continue
		}

  fmt.Fprintf(os.Stdout, "[+] %s:\n", fileName)
		rest := data[:]
		for {
			var p *pem.Block
			p, rest = pem.Decode(rest)
			if p == nil {
				break
			}

			cert, err := x509.ParseCertificate(p.Bytes)
			if err != nil {
				fmt.Fprintf(os.Stderr, "[!] %s: %v\n", fileName, err)
				break
			}

   fmt.Fprintf(os.Stdout, "\t%+v\n", cert.Subject.CommonName)
		}
	}
}
