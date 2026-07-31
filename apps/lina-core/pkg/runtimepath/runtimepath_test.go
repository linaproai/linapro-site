package runtimepath

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWorkspaceRootFromDiscoversMonorepo(t *testing.T) {
	t.Cleanup(ClearCache)
	repo := newTestRepo(t)
	coreDir := filepath.Join(repo, "apps", "lina-core")

	got := WorkspaceRootFrom(coreDir)
	want := filepath.Clean(repo)
	if filepath.Clean(got) != want {
		// Accept symlink-canonicalized forms (/var vs /private/var on macOS).
		if realPath(t, got) != realPath(t, want) {
			t.Fatalf("WorkspaceRootFrom(%q)=%q, want %q", coreDir, got, want)
		}
	}
}

func TestWorkspaceRootEnvOverride(t *testing.T) {
	t.Cleanup(ClearCache)
	override := t.TempDir()
	t.Setenv(EnvWorkspaceRoot, override)

	got := WorkspaceRootFrom(t.TempDir())
	if got != filepath.Clean(override) {
		t.Fatalf("env override: got %q want %q", got, override)
	}
}

func TestResolveRelativeAnchorsAtWorkspaceRoot(t *testing.T) {
	t.Cleanup(ClearCache)
	repo := newTestRepo(t)
	coreDir := filepath.Join(repo, "apps", "lina-core")

	got := ResolveFrom("temp/upload", coreDir)
	want := filepath.Join(filepath.Clean(repo), "temp", "upload")
	if filepath.Clean(got) != want && realPath(t, got) != realPath(t, want) {
		t.Fatalf("ResolveFrom: got %q want %q", got, want)
	}
}

func TestResolveAbsoluteUnchanged(t *testing.T) {
	t.Cleanup(ClearCache)
	abs := filepath.Join(t.TempDir(), "absolute-upload")
	got := ResolveFrom(abs, t.TempDir())
	if got != filepath.Clean(abs) {
		t.Fatalf("absolute resolve: got %q want %q", got, abs)
	}
}

func TestDataRootDefaultAndEnv(t *testing.T) {
	t.Cleanup(ClearCache)
	repo := newTestRepo(t)
	defaultRoot := DataRootFrom(repo)
	if defaultRoot != filepath.Join(filepath.Clean(repo), DefaultDataDirName) {
		t.Fatalf("default DataRoot: got %q", defaultRoot)
	}

	override := t.TempDir()
	t.Setenv(EnvDataRoot, override)
	if got := DataRootFrom(repo); got != filepath.Clean(override) {
		t.Fatalf("DataRoot env: got %q want %q", got, override)
	}
}

func TestResolveWithDefault(t *testing.T) {
	t.Cleanup(ClearCache)
	repo := newTestRepo(t)
	previousGetwd := getwd
	getwd = func() (string, error) {
		return filepath.Join(repo, "apps", "lina-core"), nil
	}
	t.Cleanup(func() { getwd = previousGetwd })

	got := ResolveWithDefault("", "temp/output")
	want := filepath.Join(filepath.Clean(repo), "temp", "output")
	if filepath.Clean(got) != want && realPath(t, got) != realPath(t, want) {
		t.Fatalf("ResolveWithDefault: got %q want %q", got, want)
	}
}

func TestWorkspaceRootFallsBackToStartDir(t *testing.T) {
	t.Cleanup(ClearCache)
	orphan := t.TempDir()
	got := WorkspaceRootFrom(orphan)
	if got != realPath(t, orphan) && got != filepath.Clean(orphan) {
		// Abs/EvalSymlinks differences across platforms are acceptable when no markers exist.
		abs, err := filepath.Abs(orphan)
		if err != nil {
			t.Fatal(err)
		}
		if filepath.Clean(got) != filepath.Clean(abs) {
			t.Fatalf("fallback root: got %q want %q", got, abs)
		}
	}
}

func newTestRepo(t *testing.T) string {
	t.Helper()
	repo := t.TempDir()
	for _, dir := range []string{
		filepath.Join(repo, "apps", "lina-core"),
		filepath.Join(repo, "apps", "lina-vben"),
	} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", dir, err)
		}
	}
	if err := os.WriteFile(filepath.Join(repo, "go.work"), []byte("go 1.26\n"), 0o644); err != nil {
		t.Fatalf("write go.work: %v", err)
	}
	if err := os.WriteFile(filepath.Join(repo, "apps", "lina-core", "go.mod"), []byte("module lina-core\n"), 0o644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}
	if err := os.WriteFile(filepath.Join(repo, "apps", "lina-vben", "package.json"), []byte("{}\n"), 0o644); err != nil {
		t.Fatalf("write package.json: %v", err)
	}
	return repo
}

func realPath(t *testing.T, pathValue string) string {
	t.Helper()
	resolved, err := filepath.EvalSymlinks(pathValue)
	if err != nil {
		t.Fatalf("EvalSymlinks(%q): %v", pathValue, err)
	}
	return filepath.Clean(resolved)
}
