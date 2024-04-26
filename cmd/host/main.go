package main

import (
	"flag"
	"fmt"
	"log"
	"net"
)

func lookupHost(host string) error {
	cname, err := net.LookupCNAME(host)
	if err != nil {
		return err
	}

	if cname != host {
		fmt.Printf("%s is a CNAME for %s\n", host, cname)
		host = cname
	}

	addrs, err := net.LookupHost(host)
	if err != nil {
		return err
	}

	for _, addr := range addrs {
		fmt.Printf("\t%s\n", addr)
	}

	return nil
}

func main() {
	flag.Parse()

	for _, arg := range flag.Args() {
		if err := lookupHost(arg); err != nil {
			log.Println("%s: %s", arg, err)
		}
	}
}
