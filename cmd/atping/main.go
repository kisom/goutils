package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"time"
)

const (
	defaultServer = "google.com"
	defaultPort   = "80"
)

var verbose bool

func connect(addr string, dport string, six bool, timeout time.Duration) error {
	_, _, err := net.SplitHostPort(addr)
	if err != nil {
		addr = net.JoinHostPort(addr, dport)
	}

	proto := "tcp"
	if six {
		proto += "6"
	}

	if verbose {
		fmt.Printf("connecting to %s/%s... ", addr, proto)
		os.Stdout.Sync()
	}

	conn, err := net.DialTimeout(proto, addr, timeout)
	if err != nil {
		if verbose {
			fmt.Println("failed.")
		}
		return err
	}

	if verbose {
		fmt.Println("OK")
	}
	conn.Close()
	return nil
}

func main() {
	var (
		port    string
		timeout time.Duration
		six     bool
	)

	flag.BoolVar(&six, "6", false, "require IPv6")
	flag.StringVar(&port, "p", defaultPort, "`port` to connect to instead of "+defaultPort)
	flag.DurationVar(&timeout, "t", 3*time.Second, "`timeout`")
	flag.BoolVar(&verbose, "v", false, "verbose mode: print server and protocol when connecting")
	flag.Parse()

	var servers []string
	if flag.NArg() == 0 {
		servers = []string{defaultServer}
	} else {
		servers = flag.Args()
	}

	for _, server := range servers {
		err := connect(server, port, six, timeout)
		if err != nil {
			os.Exit(1)
		}
	}
}
