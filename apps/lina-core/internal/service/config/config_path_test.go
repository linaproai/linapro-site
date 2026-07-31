// This file verifies repository-root anchoring for runtime filesystem paths.

package config

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"lina-core/pkg/runtimepath"
)

// TestResolveRuntimePathFromWorkingDirAnchorsAtRepositoryRoot verifies relative
// runtime paths resolve from the LinaPro repository root instead of the caller
// process working directory.
func TestResolveRuntimePathFromWorkingDirAnchorsAtRepositoryRoot(t *testing.T) {
	repoRoot := newTestRepoRoot(t)
	backendWorkingDir := filepath.Join(repoRoot, "apps", "lina-core")

	resolvedPath := runtimepath.ResolveFrom("temp/output", backendWorkingDir)
	expectedPath := filepath.Join(repoRoot, "temp", "output")
	if sameFilesystemPath(t, resolvedPath, expectedPath) {
		return
	}
	t.Fatalf("expected path %q, got %q", expectedPath, resolvedPath)
}

// TestConfigRuntimePathGettersAnchorRelativePathsAtRepositoryRoot verifies upload
// and dynamic plugin storage paths share the same root anchoring behavior.
func TestConfigRuntimePathGettersAnchorRelativePathsAtRepositoryRoot(t *testing.T) {
	setTestConfigContent(t, `
upload:
  path: temp/upload
plugin:
  dynamic:
    storagePath: temp/output
`)
	var (
		repoRoot          = newTestRepoRoot(t)
		backendWorkingDir = filepath.Join(repoRoot, "apps", "lina-core")
		expectedRepoRoot  = realTestPath(t, repoRoot)
	)
	withWorkingDir(t, backendWorkingDir)
	setPluginDynamicStoragePathOverride("")
	t.Cleanup(func() {
		setPluginDynamicStoragePathOverride("")
	})

	svc := New()
	uploadPath := svc.GetUploadPath(context.Background())
	if !sameFilesystemPath(t, uploadPath, filepath.Join(expectedRepoRoot, "temp", "upload")) {
		t.Fatalf("expected upload path under repo temp, got %q", uploadPath)
	}
	pluginPath := svc.GetPluginDynamicStoragePath(context.Background())
	if !sameFilesystemPath(t, pluginPath, filepath.Join(expectedRepoRoot, "temp", "output")) {
		t.Fatalf("expected plugin storage path under repo temp, got %q", pluginPath)
	}
}

// newTestRepoRoot creates a minimal LinaPro-like repository layout for path
// resolution tests without depending on the real checkout location.
func newTestRepoRoot(t *testing.T) string {
	t.Helper()

	repoRoot := t.TempDir()
	for _, dir := range []string{
		filepath.Join(repoRoot, "apps", "lina-core"),
		filepath.Join(repoRoot, "apps", "lina-vben"),
	} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatalf("create test repo dir %s: %v", dir, err)
		}
	}
	if err := os.WriteFile(filepath.Join(repoRoot, "go.work"), []byte("go 1.26\n"), 0o644); err != nil {
		t.Fatalf("write test go.work: %v", err)
	}
	return repoRoot
}

// realTestPath resolves symlinks in an existing test path before string
// comparisons on platforms where os.Getwd canonicalizes temporary directories.
func realTestPath(t *testing.T, targetPath string) string {
	t.Helper()

	realPath, err := filepath.EvalSymlinks(targetPath)
	if err != nil {
		t.Fatalf("resolve real test path %s: %v", targetPath, err)
	}
	return realPath
}

// sameFilesystemPath compares paths after Clean and optional EvalSymlinks of
// existing ancestors so macOS /var vs /private/var does not fail tests for
// paths that do not exist yet.
func sameFilesystemPath(t *testing.T, left string, right string) bool {
	t.Helper()
	if filepath.Clean(left) == filepath.Clean(right) {
		return true
	}
	return canonicalizePath(left) == canonicalizePath(right)
}

func canonicalizePath(pathValue string) string {
	cleaned := filepath.Clean(pathValue)
	current := cleaned
	for {
		if real, err := filepath.EvalSymlinks(current); err == nil {
			suffix, err := filepath.Rel(current, cleaned)
			if err != nil || strings.HasPrefix(suffix, "..") {
				return cleaned
			}
			if suffix == "." {
				return real
			}
			return filepath.Join(real, suffix)
		}
		parent := filepath.Dir(current)
		if parent == current {
			return cleaned
		}
		current = parent
	}
}

// withWorkingDir changes the process working directory for one test and
// restores it during cleanup.
func withWorkingDir(t *testing.T, workingDir string) {
	t.Helper()

	originalWorkingDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("get original working directory: %v", err)
	}
	if err = os.Chdir(workingDir); err != nil {
		t.Fatalf("change working directory to %s: %v", workingDir, err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(originalWorkingDir); err != nil {
			t.Fatalf("restore working directory: %v", err)
		}
	})
}
