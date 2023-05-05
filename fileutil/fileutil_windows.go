//go:build windows
// +build windows

// Package fileutil contains common file functions.
package fileutil

import (
	"errors"
	"os"
)

// FileDoesExist returns true if the file exists.
func FileDoesExist(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// DirectoryDoesExist returns true if the file exists.
func DirectoryDoesExist(path string) bool {
	fi, err := os.Stat(path)
	if err != nil {
		return false
	}

	return fi.Mode().IsDir()
}

const (
	// AccessExists checks whether the file exists. This is invalid outside of
	// Unix systems.
	AccessExists = 0

	// AccessRead checks whether the user has read permissions on
	// the file. This is invalid outside of Unix systems.
	AccessRead = 0

	// AccessWrite checks whether the user has write permissions
	// on the file. This is invalid outside of Unix systems.
	AccessWrite = 0

	// AccessExec checks whether the user has executable
	// permissions on the file. This is invalid outside of Unix systems.
	AccessExec = 0
)

// Access is a Unix-only call, and has no meaning on Windows.
func Access(path string, mode int) error {
	return errors.New("fileutil: Access is meaningless on Windows")
}
