//go:build !linux

package config

import (
	"os/user"
	"path/filepath"
)

// DefaultConfigPath returns a sensible default configuration file path.
func DefaultConfigPath(dir, base string) string {
	user, err := user.Current()
	if err != nil || user.HomeDir == "" {
		return filepath.Join(dir, base)
	}

	return filepath.Join(user.HomeDir, dir, base)
}
