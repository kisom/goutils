// Package fileutil contains common file functions.
package fileutil

import (
	"os"

	"golang.org/x/sys/unix"
)

// FileExistsP returns true if the file exists.
func FileExistsP(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// DirectoryExistsP returns true if the file exists.
func DirectoryExistsP(path string) bool {
	fi, err := os.Stat(path)
	if err != nil {
		return false
	}

	return fi.Mode().IsDir()
}

const (
	// AccessExists checks whether the file exists.
	AccessExists = unix.F_OK

	// AccessRead checks whether the user has read permissions on
	// the file.
	AccessRead = unix.R_OK

	// AccessWrite checks whether the user has write permissions
	// on the file.
	AccessWrite = unix.W_OK

	// AccessExec checks whether the user has executable
	// permissions on the file.
	AccessExec = unix.X_OK
)

// Access returns a boolean indicating whether the mode being checked
// for is valid.
func Access(path string, mode int) error {
	return unix.Access(path, uint32(mode))
}
