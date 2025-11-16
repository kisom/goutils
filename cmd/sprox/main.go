package main

import (
	"flag"
	"io"
	"net"

	"git.wntrmute.dev/kyle/goutils/die"
	"git.wntrmute.dev/kyle/goutils/lib"
)

func proxy(conn net.Conn, inside string) error {
    proxyConn, err := net.Dial("tcp", inside)
    if err != nil {
        return err
    }

	defer proxyConn.Close()
	defer conn.Close()

    go func() {
        _, _ = io.Copy(conn, proxyConn)
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
            _, _ = lib.Warn(err, "accept failed")
            continue
        }

        go func() {
            if err := proxy(conn, "127.0.0.1:"+inside); err != nil {
                _, _ = lib.Warn(err, "proxy error")
            }
        }()
    }
}
