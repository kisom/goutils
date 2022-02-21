package config

import (
	"os"
	"path/filepath"
)

// canUseXDGConfigDir checks whether the XDG config directory exists
// and is accessible by the current user. If it is present, it will
// be returned. Note that if the directory does not exist, it is
// presumed unusable.
func canUseXDGConfigDir() (string, bool) {
	xdgDir := os.Getenv("XDG_CONFIG_DIR")
	if xdgDir == "" {
		userDir := os.Getenv("HOME")
		if userDir == "" {
			return "", false
		}

		xdgDir = filepath.Join(userDir, ".config")
	}

	fi, err := os.Stat(xdgDir)
	if err != nil {
		return "", false
	}

	if !fi.IsDir() {
		return "", false
	}

	return xdgDir, true
}

// DefaultConfigPath returns a sensible default configuration file path.
func DefaultConfigPath(dir, base string) string {
	dirPath, ok := canUseXDGConfigDir()
	if !ok {
		dirPath = "/etc"
	}

	return filepath.Join(dirPath, dir, base)
}
