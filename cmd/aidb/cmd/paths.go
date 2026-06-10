package cmd

import (
	"path/filepath"
	"strings"
)

// resolveArg turns a CLI path argument into an absolute path
func resolveArg(cwd, arg string) string {
	if filepath.IsAbs(arg) {
		return arg
	}
	return filepath.Join(cwd, arg)
}

// withinDir reports whether child is inside parent (path-aware, unlike a string-prefix check)
func withinDir(parent, child string) bool {
	rel, err := filepath.Rel(parent, child)
	if err != nil {
		return false
	}
	return rel != ".." && !strings.HasPrefix(rel, "../")
}
