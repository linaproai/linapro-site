// This file resolves filesystem paths from static config into stable runtime
// paths for host-local storage and development artifacts.

package config

import (
	"path/filepath"
	"strings"

	"lina-core/pkg/runtimepath"
)

// resolveRuntimePathWithDefault resolves a configured filesystem path, falling
// back to defaultPath when the configured value is blank.
func resolveRuntimePathWithDefault(configuredPath string, defaultPath string) string {
	return runtimepath.ResolveWithDefault(configuredPath, defaultPath)
}

// cleanConfigPath trims whitespace and normalizes path separators without
// changing whether the path is absolute or relative. Kept for host config
// helpers that only need cleaning before optional override checks.
func cleanConfigPath(configuredPath string) string {
	trimmedPath := strings.TrimSpace(configuredPath)
	if trimmedPath == "" {
		return ""
	}
	return filepath.Clean(trimmedPath)
}
