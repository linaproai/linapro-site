// This file verifies local marketplace artifact store put/open path safety.

package marketplace

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
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
}

func TestResolveArtifactStoreRootUsesConfigAndDefault(t *testing.T) {
	t.Parallel()

	if got := resolveArtifactStoreRoot(context.Background(), nil); got != defaultArtifactStoreRoot {
		t.Fatalf("nil config root = %q, want %q", got, defaultArtifactStoreRoot)
	}
	if got := resolveArtifactStoreRoot(context.Background(), stubPluginConfig{}); got != defaultArtifactStoreRoot {
		t.Fatalf("empty config root = %q, want %q", got, defaultArtifactStoreRoot)
	}

	configured := filepath.Join(t.TempDir(), "marketplace-artifacts")
	got := resolveArtifactStoreRoot(context.Background(), stubPluginConfig{
		values: map[string]string{
			configKeyStorageRoot: configured,
		},
	})
	if got != configured {
		t.Fatalf("configured root = %q, want %q", got, configured)
	}
}

func TestNewLocalArtifactStoreEmptyRootUsesTempDefault(t *testing.T) {
	t.Parallel()

	// Use an isolated process CWD so the default relative path does not write
	// into the repository tree during unit tests.
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

	store, err := NewLocalArtifactStore("")
	if err != nil {
		t.Fatalf("NewLocalArtifactStore: %v", err)
	}
	root := store.Root()
	wantSuffix := filepath.FromSlash(defaultArtifactStoreRoot)
	if !strings.HasSuffix(root, wantSuffix) {
		t.Fatalf("default root = %q, want suffix %q", root, wantSuffix)
	}
	// macOS may surface temp paths as /var vs /private/var; canonicalize both.
	absWorkDir, err := filepath.EvalSymlinks(workDir)
	if err != nil {
		t.Fatalf("EvalSymlinks workdir: %v", err)
	}
	absRoot, err := filepath.EvalSymlinks(root)
	if err != nil {
		t.Fatalf("EvalSymlinks root: %v", err)
	}
	rel, err := filepath.Rel(absWorkDir, absRoot)
	if err != nil || strings.HasPrefix(rel, "..") {
		t.Fatalf("default root %q should stay under test workdir %q (rel=%q err=%v)", absRoot, absWorkDir, rel, err)
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
	abs, absErr := filepath.Abs(configured)
	if absErr != nil {
		t.Fatalf("Abs: %v", absErr)
	}
	if store.Root() != abs {
		t.Fatalf("store root = %q, want %q", store.Root(), abs)
	}
}
