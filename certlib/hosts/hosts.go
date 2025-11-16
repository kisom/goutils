package hosts

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"
)

const defaultHTTPSPort = 443

type Target struct {
	Host string
	Port int
}

func (t *Target) String() string {
	return fmt.Sprintf("%s:%d", t.Host, t.Port)
}

func parseURL(host string) (string, int, error) {
	url, err := url.Parse(host)
	if err != nil {
		return "", 0, fmt.Errorf("certlib/hosts: invalid host: %s", host)
	}

	if strings.ToLower(url.Scheme) != "https" {
		return "", 0, errors.New("certlib/hosts: only https scheme supported")
	}

	if url.Port() == "" {
		return url.Hostname(), defaultHTTPSPort, nil
	}

	portInt, err2 := strconv.ParseInt(url.Port(), 10, 16)
	if err2 != nil {
		return "", 0, fmt.Errorf("certlib/hosts: invalid port: %s", url.Port())
	}

	return url.Hostname(), int(portInt), nil
}

func parseHostPort(host string) (string, int, error) {
	host, sport, err := net.SplitHostPort(host)
	if err == nil {
		portInt, err2 := strconv.ParseInt(sport, 10, 16)
		if err2 != nil {
			return "", 0, fmt.Errorf("certlib/hosts: invalid port: %s", sport)
		}

		return host, int(portInt), nil
	}

	return host, defaultHTTPSPort, nil
}

func ParseHost(host string) (*Target, error) {
	host, port, err := parseURL(host)
	if err == nil {
		return &Target{Host: host, Port: port}, nil
	}

	host, port, err = parseHostPort(host)
	if err == nil {
		return &Target{Host: host, Port: port}, nil
	}

	return nil, fmt.Errorf("certlib/hosts: invalid host: %s", host)
}

func ParseHosts(hosts ...string) ([]*Target, error) {
	targets := make([]*Target, 0, len(hosts))
	for _, host := range hosts {
		target, err := ParseHost(host)
		if err != nil {
			return nil, err
		}
		targets = append(targets, target)
	}

	return targets, nil
}
