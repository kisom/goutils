package main

import (
	"crypto/sha256"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"path/filepath"

	"github.com/kisom/goutils/lib"
)

func fetch(remote string) ([]byte, error) {
	resp, err := http.Get(remote)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}

func main() {
	flag.Parse()

	for _, remote := range flag.Args() {
		u, err := url.Parse(remote)
		if err != nil {
			lib.Warn(err, "parsing %s", remote)
			continue
		}

		name := filepath.Base(u.Path)
		if name == "" {
			lib.Warnx("source URL doesn't appear to name a file")
			continue
		}

		body, err := fetch(remote)
		if err != nil {
			lib.Warn(err, "fetching %s", remote)
			continue
		}

		h := sha256.Sum256(body)
		fmt.Printf("%s: sha256=%x\n", name, h)
	}
}
