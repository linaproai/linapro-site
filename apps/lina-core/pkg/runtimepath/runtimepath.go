// Package runtimepath resolves host and plugin local filesystem paths against a
// stable workspace root instead of the process working directory.
//
// Relative configured paths such as temp/upload are joined to WorkspaceRoot so
// make dev (WorkDir=apps/lina-core) and monorepo tooling (temp under repo root)
// share one data layout. Absolute paths are cleaned and returned unchanged.
package runtimepath

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
)

const (
	// EnvWorkspaceRoot overrides workspace root discovery when set to an
	// absolute path.
	EnvWorkspaceRoot = "LINAPRO_WORKSPACE_ROOT"
	// EnvDataRoot overrides the default data root ({WorkspaceRoot}/temp) when
	// set to an absolute path.
	EnvDataRoot = "LINAPRO_DATA_ROOT"

	// DefaultDataDirName is the default data directory under WorkspaceRoot.
	DefaultDataDirName = "temp"

	// repositoryRootSearchLimit bounds upward workspace-root discovery.
	repositoryRootSearchLimit = 12
)

var (
	// lookupEnv is swapped in tests.
	lookupEnv = os.LookupEnv
	// getwd is swapped in tests.
	getwd = os.Getwd
	// stat is swapped in tests for existence checks used by root markers.
	stat = os.Stat

	workspaceRootCache sync.Map // map[string]string keyed by startDir abs path
)

// WorkspaceRoot returns the monorepo or installation workspace root resolved
// from the process working directory.
func WorkspaceRoot() string {
	workingDir, err := getwd()
	if err != nil {
		return ""
	}
	return WorkspaceRootFrom(workingDir)
}

// WorkspaceRootFrom resolves the workspace root starting from startDir.
// Priority: LINAPRO_WORKSPACE_ROOT env, upward marker discovery, then startDir.
func WorkspaceRootFrom(startDir string) string {
	if root, ok := envAbsolutePath(EnvWorkspaceRoot); ok {
		return root
	}

	baseDir := cleanPath(startDir)
	if baseDir == "" {
		baseDir = "."
	}
	if abs, err := filepath.Abs(baseDir); err == nil {
		baseDir = abs
	}
	baseDir = filepath.Clean(baseDir)

	if cached, ok := workspaceRootCache.Load(baseDir); ok {
		if root, ok := cached.(string); ok && root != "" {
			return root
		}
	}

	root, err := findWorkspaceRoot(baseDir)
	if err != nil {
		root = baseDir
	}
	workspaceRootCache.Store(baseDir, root)
	return root
}

// DataRoot returns the writable data root. Priority: LINAPRO_DATA_ROOT env,
// otherwise Join(WorkspaceRoot(), "temp").
func DataRoot() string {
	return DataRootFrom(WorkspaceRoot())
}

// DataRootFrom returns the data root for an explicit workspace root.
func DataRootFrom(workspaceRoot string) string {
	if root, ok := envAbsolutePath(EnvDataRoot); ok {
		return root
	}
	base := cleanPath(workspaceRoot)
	if base == "" {
		base = WorkspaceRoot()
	}
	if base == "" {
		return DefaultDataDirName
	}
	return filepath.Clean(filepath.Join(base, DefaultDataDirName))
}

// Resolve resolves a configured filesystem path against the process workspace
// root. Empty input returns empty; absolute paths are cleaned; relative paths
// join WorkspaceRoot (phase-1 semantics compatible with temp/* config values).
func Resolve(configuredPath string) string {
	workingDir, err := getwd()
	if err != nil {
		return cleanPath(configuredPath)
	}
	return ResolveFrom(configuredPath, workingDir)
}

// ResolveFrom resolves configuredPath using workspace discovery from startDir.
func ResolveFrom(configuredPath string, startDir string) string {
	cleanedPath := cleanPath(configuredPath)
	if cleanedPath == "" {
		return ""
	}
	if filepath.IsAbs(cleanedPath) {
		return cleanedPath
	}
	root := WorkspaceRootFrom(startDir)
	if root == "" {
		root = cleanPath(startDir)
	}
	if root == "" {
		return cleanedPath
	}
	return filepath.Clean(filepath.Join(root, cleanedPath))
}

// ResolveWithDefault resolves configuredPath, falling back to defaultPath when
// the configured value is blank.
func ResolveWithDefault(configuredPath string, defaultPath string) string {
	cleaned := cleanPath(configuredPath)
	if cleaned == "" {
		cleaned = cleanPath(defaultPath)
	}
	return Resolve(cleaned)
}

// ClearCache drops cached workspace-root lookups. Intended for tests.
func ClearCache() {
	workspaceRootCache.Range(func(key, _ any) bool {
		workspaceRootCache.Delete(key)
		return true
	})
}

func envAbsolutePath(key string) (string, bool) {
	value, ok := lookupEnv(key)
	if !ok {
		return "", false
	}
	cleaned := cleanPath(value)
	if cleaned == "" || !filepath.IsAbs(cleaned) {
		return "", false
	}
	return cleaned, true
}

func cleanPath(pathValue string) string {
	trimmed := strings.TrimSpace(pathValue)
	if trimmed == "" {
		return ""
	}
	return filepath.Clean(trimmed)
}

func findWorkspaceRoot(startDir string) (string, error) {
	current := filepath.Clean(startDir)
	for depth := 0; depth < repositoryRootSearchLimit; depth++ {
		if isWorkspaceRoot(current) {
			return current, nil
		}
		parent := filepath.Dir(current)
		if parent == current {
			break
		}
		current = parent
	}
	return "", os.ErrNotExist
}

func isWorkspaceRoot(dir string) bool {
	if pathExists(filepath.Join(dir, "go.work")) && pathExists(filepath.Join(dir, "apps", "lina-core")) {
		return true
	}
	return pathExists(filepath.Join(dir, "apps", "lina-core", "go.mod")) &&
		pathExists(filepath.Join(dir, "apps", "lina-vben", "package.json"))
}

func pathExists(pathValue string) bool {
	_, err := stat(pathValue)
	return err == nil
}
