// This file verifies local marketplace artifact store put/open path safety.

package marketplace

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"lina-core/pkg/runtimepath"
)

func TestLocalArtifactStorePutOpenAndRejectsTraversal(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	store, err := NewLocalArtifactStore(root)
	if err != nil {
		t.Fatalf("NewLocalArtifactStore: %v", err)
	}

	ctx := context.Background()
	key := "source/demo/v1.0.0/pkg.zip"
	content := []byte("marketplace-package")
	if err = store.Put(ctx, key, bytes.NewReader(content)); err != nil {
		t.Fatalf("Put: %v", err)
	}

	reader, err := store.Open(ctx, key)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer func() {
		_ = reader.Close()
	}()
	got, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(key)))
	if err != nil {
		t.Fatalf("read stored file: %v", err)
	}
	if !bytes.Equal(got, content) {
		t.Fatalf("stored content = %q, want %q", got, content)
	}

	if _, err = store.LocalPath(ctx, "../escape.zip"); err == nil {
		t.Fatal("expected path traversal to be rejected")
	}
	if _, err = store.Open(ctx, "missing/package.zip"); err == nil {
		t.Fatal("expected missing object to fail")
	}

	if err = store.Delete(ctx, key); err != nil {
		t.Fatalf("Delete existing: %v", err)
	}
	if _, err = store.Open(ctx, key); err == nil {
		t.Fatal("expected open after delete to fail")
	}
	if err = store.Delete(ctx, key); err != nil {
		t.Fatalf("Delete missing must be idempotent: %v", err)
	}
}

func TestResolveArtifactStoreRootUsesConfigAndDefault(t *testing.T) {
	t.Parallel()

	wantDefault := runtimepath.Resolve(defaultArtifactStoreRoot)
	if got := resolveArtifactStoreRoot(context.Background(), nil); got != wantDefault {
		t.Fatalf("nil config root = %q, want %q", got, wantDefault)
	}
	if got := resolveArtifactStoreRoot(context.Background(), stubPluginConfig{}); got != wantDefault {
		t.Fatalf("empty config root = %q, want %q", got, wantDefault)
	}

	configured := filepath.Join(t.TempDir(), "marketplace-artifacts")
	got := resolveArtifactStoreRoot(context.Background(), stubPluginConfig{
		values: map[string]string{
			configKeyStorageRoot: configured,
		},
	})
	if got != filepath.Clean(configured) {
		t.Fatalf("configured root = %q, want %q", got, configured)
	}
}

func TestNewLocalArtifactStoreEmptyRootUsesWorkspaceTempDefault(t *testing.T) {
	// Not parallel: mutates process CWD and workspace-root env.

	// Isolate CWD and force workspace root via env so relative defaults do not
	// write into the real repository tree during unit tests.
	originalWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd: %v", err)
	}
	workDir := t.TempDir()
	if err = os.Chdir(workDir); err != nil {
		t.Fatalf("Chdir: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(originalWD)
	})
	t.Setenv(runtimepath.EnvWorkspaceRoot, workDir)
	runtimepath.ClearCache()
	t.Cleanup(runtimepath.ClearCache)

	store, err := NewLocalArtifactStore("")
	if err != nil {
		t.Fatalf("NewLocalArtifactStore: %v", err)
	}
	root := store.Root()
	want := filepath.Join(workDir, filepath.FromSlash(defaultArtifactStoreRoot))
	// macOS may surface temp paths as /var vs /private/var; canonicalize both.
	absWant, err := filepath.EvalSymlinks(want)
	if err != nil {
		// directory was just created under store root parent; use Clean want
		absWant = filepath.Clean(want)
	}
	absRoot, err := filepath.EvalSymlinks(root)
	if err != nil {
		t.Fatalf("EvalSymlinks root: %v", err)
	}
	if absRoot != absWant && root != filepath.Clean(want) {
		t.Fatalf("default root = %q, want %q", root, want)
	}
	if !strings.HasSuffix(filepath.ToSlash(root), "temp/plugin-marketplace/artifacts") {
		t.Fatalf("default root = %q, want suffix temp/plugin-marketplace/artifacts", root)
	}
	if _, err = os.Stat(root); err != nil {
		t.Fatalf("default root should exist: %v", err)
	}
}

func TestNewUsesConfiguredArtifactRoot(t *testing.T) {
	t.Parallel()

	configured := filepath.Join(t.TempDir(), "from-config")
	svc, err := New(nil, stubPluginConfig{
		values: map[string]string{
			configKeyStorageRoot: configured,
		},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	impl, ok := svc.(*serviceImpl)
	if !ok || impl == nil {
		t.Fatal("expected serviceImpl")
	}
	store, ok := impl.artifacts.(*LocalArtifactStore)
	if !ok || store == nil {
		t.Fatal("expected LocalArtifactStore")
	}
	if store.Root() != filepath.Clean(configured) {
		t.Fatalf("store root = %q, want %q", store.Root(), configured)
	}
}

func TestResolveArtifactStoreRootRelativeAnchorsWorkspace(t *testing.T) {
	// Not parallel: mutates workspace-root env.

	workDir := t.TempDir()
	t.Setenv(runtimepath.EnvWorkspaceRoot, workDir)
	runtimepath.ClearCache()
	t.Cleanup(runtimepath.ClearCache)

	got := resolveArtifactStoreRoot(context.Background(), stubPluginConfig{
		values: map[string]string{
			configKeyStorageRoot: "temp/plugin-marketplace/artifacts",
		},
	})
	want := filepath.Join(workDir, "temp", "plugin-marketplace", "artifacts")
	if got != want {
		t.Fatalf("relative root = %q, want %q", got, want)
	}
}
