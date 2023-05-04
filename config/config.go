// Package config implements a simple global configuration system that
// supports a file with key=value pairs and environment variables. Note
// that the config system is global.
//
// This package is intended to be used for small daemons: some configuration
// file is optionally populated at program start, then this is used to
// transparently look up configuration values from either that file or the
// environment.
package config

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"

	"git.wntrmute.dev/kyle/goutils/config/iniconf"
)

// NB: Rather than define a singleton type, everything is defined at
// the top-level

var (
	vars   = map[string]string{}
	prefix = ""
)

// SetEnvPrefix sets the prefix for all environment variables; it's
// assumed to not be needed for files.
func SetEnvPrefix(pfx string) {
	prefix = pfx
}

func addLine(line string) {
	if strings.HasPrefix(line, "#") || line == "" {
		return
	}

	lineParts := strings.SplitN(line, "=", 2)
	if len(lineParts) != 2 {
		log.Print("skipping line: ", line)
		return // silently ignore empty keys
	}

	lineParts[0] = strings.TrimSpace(lineParts[0])
	lineParts[1] = strings.TrimSpace(lineParts[1])
	vars[lineParts[0]] = lineParts[1]
}

// LoadFile scans the file at path for key=value pairs and adds them
// to the configuration.
func LoadFile(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		addLine(line)
	}

	if err = scanner.Err(); err != nil {
		return err
	}

	return nil
}

// LoadFileFor scans the ini file at path, loading the default section
// and overriding any keys found under section. If strict is true, the
// named section must exist (i.e. to catch typos in the section name).
func LoadFileFor(path, section string, strict bool) error {
	cmap, err := iniconf.ParseFile(path)
	if err != nil {
		return err
	}

	for key, value := range cmap[iniconf.DefaultSection] {
		vars[key] = value
	}

	smap, ok := cmap[section]
	if !ok {
		if strict {
			return fmt.Errorf("config: section '%s' wasn't found in the config file", section)
		}
		return nil
	}

	for key, value := range smap {
		vars[key] = value
	}

	return nil
}

// Get retrieves a value from either a configuration file or the
// environment. Note that values from a file will override environment
// variables.
func Get(key string) string {
	if v, ok := vars[key]; ok {
		return v
	}
	return os.Getenv(prefix + key)
}

// GetDefault retrieves a value from either a configuration file or
// the environment. Note that value from a file will override
// environment variables. If a value isn't found (e.g. Get returns an
// empty string), the default value will be used.
func GetDefault(key, def string) string {
	if v := Get(key); v != "" {
		return v
	}
	return def
}

// Require retrieves a value from either a configuration file or the
// environment. If the key isn't present, it will call log.Fatal, printing
// the missing key.
func Require(key string) string {
	if v, ok := vars[key]; ok {
		return v
	}

	v, ok := os.LookupEnv(prefix + key)
	if !ok {
		var envMessage string
		if prefix != "" {
			envMessage = " (note: looked for the key " + prefix + key
			envMessage += " in the local env)"
		}
		log.Fatalf("missing required configuration value %s%s", key, envMessage)
	}

	return v
}
