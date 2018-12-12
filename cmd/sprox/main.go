package main

import (
	"flag"
	"io"
	"log"
	"net"

	"github.com/kisom/goutils/die"
)

func proxy(conn net.Conn, inside string) error {
	proxyConn, err := net.Dial("tcp", inside)
	if err != nil {
		return err
	}

	defer proxyConn.Close()
	defer conn.Close()

	go func() {
		io.Copy(conn, proxyConn)
	}()
	_, err = io.Copy(proxyConn, conn)
	return err
}

func main() {
	var outside, inside string
	flag.StringVar(&outside, "f", "8080", "outside port")
	flag.StringVar(&inside, "p", "4000", "inside port")
	flag.Parse()

	l, err := net.Listen("tcp", "0.0.0.0:"+outside)
	die.If(err)

	for {
		conn, err := l.Accept()
		if err != nil {
			log.Println(err)
			continue
		}

		go proxy(conn, "127.0.0.1:"+inside)
	}
}
