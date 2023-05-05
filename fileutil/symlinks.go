package fileutil

import (
	"path/filepath"
	"strings"
)

// ValidateSymlink checks to make sure a symlink exists in some top-level
// directory.
func ValidateSymlink(symlink, topLevel string) bool {
	target, err := filepath.EvalSymlinks(symlink)
	if err != nil {
		return false
	}
	return strings.HasPrefix(target, topLevel)
}
