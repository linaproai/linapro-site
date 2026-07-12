// This file verifies local marketplace artifact store put/open path safety.

package marketplace

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
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
