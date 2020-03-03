// Package config implements a simple global configuration system that
// supports a file with key=value pairs and environment variables. Note
// that the config system is global.
package config

import (
	"bufio"
	"log"
	"os"
	"strings"
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
