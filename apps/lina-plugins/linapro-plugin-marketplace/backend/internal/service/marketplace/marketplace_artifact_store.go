// This file implements the marketplace artifact store used to persist uploaded
// packages and stream controlled downloads. The default implementation keeps
// objects on the local filesystem under a plugin-owned root directory so the
// marketplace plugin can run without an external object-storage dependency.

package marketplace

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/pkg/bizerr"
	"lina-core/pkg/plugin/capability/plugincap"
)

// ArtifactStore persists marketplace package bytes keyed by storage key.
type ArtifactStore interface {
	// Put stores one object, replacing any existing object with the same key.
	Put(ctx context.Context, key string, body io.Reader) error
	// PutFile copies one local file into the store under the supplied key.
	PutFile(ctx context.Context, key string, localPath string) error
	// Open returns a reader for one stored object.
	Open(ctx context.Context, key string) (io.ReadCloser, error)
	// LocalPath returns a filesystem path suitable for zip.OpenReader when the
	// store is local. Remote-backed implementations may return an error.
	LocalPath(ctx context.Context, key string) (string, error)
}

// LocalArtifactStore stores marketplace artifacts under one root directory.
type LocalArtifactStore struct {
	root string
}

// defaultArtifactStoreRoot is the relative filesystem root used when no
// storage.root config or constructor argument is provided. It stays under the
// process working directory's temp tree so development artifacts are not written
// into tracked source paths such as apps/lina-core/data/.
const defaultArtifactStoreRoot = "temp/plugin-marketplace/artifacts"

// configKeyStorageRoot is the plugin-scoped config key for the local artifact
// root used by package uploads, Git docs snapshots, and download streaming.
const configKeyStorageRoot = "storage.root"

// NewLocalArtifactStore creates a local-disk artifact store rooted at rootDir.
// Empty rootDir defaults to temp/plugin-marketplace/artifacts under the process
// working directory (typically apps/lina-core/temp when the host starts there).
func NewLocalArtifactStore(rootDir string) (*LocalArtifactStore, error) {
	root := strings.TrimSpace(rootDir)
	if root == "" {
		root = defaultArtifactStoreRoot
	}
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return nil, gerror.Wrapf(err, "resolve marketplace artifact root %q", root)
	}
	if err = os.MkdirAll(absRoot, 0o755); err != nil {
		return nil, gerror.Wrapf(err, "create marketplace artifact root %q", absRoot)
	}
	return &LocalArtifactStore{root: absRoot}, nil
}

// resolveArtifactStoreRoot returns the configured local artifact root, or the
// default temp path when storage.root is unset or blank.
func resolveArtifactStoreRoot(ctx context.Context, pluginConfig plugincap.ConfigService) string {
	if pluginConfig != nil {
		configured, err := pluginConfig.String(ctx, configKeyStorageRoot, "")
		if err == nil {
			if root := strings.TrimSpace(configured); root != "" {
				return root
			}
		}
	}
	return defaultArtifactStoreRoot
}

// Root returns the absolute filesystem root used by this store. Tests and
// diagnostics use it to assert configured storage locations.
func (s *LocalArtifactStore) Root() string {
	if s == nil {
		return ""
	}
	return s.root
}

// Put stores object bytes for one storage key.
func (s *LocalArtifactStore) Put(_ context.Context, key string, body io.Reader) error {
	if body == nil {
		return bizerr.NewCode(CodeMarketplaceInvalidInput)
	}
	target, err := s.resolvePath(key)
	if err != nil {
		return err
	}
	if err = os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return bizerr.WrapCode(err, CodeMarketplaceStorageFailed)
	}
	tempPath := target + ".tmp"
	file, err := os.OpenFile(tempPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
	if err != nil {
		return bizerr.WrapCode(err, CodeMarketplaceStorageFailed)
	}
	if _, copyErr := io.Copy(file, body); copyErr != nil {
		_ = file.Close()
		_ = os.Remove(tempPath)
		return bizerr.WrapCode(copyErr, CodeMarketplaceStorageFailed)
	}
	if closeErr := file.Close(); closeErr != nil {
		_ = os.Remove(tempPath)
		return bizerr.WrapCode(closeErr, CodeMarketplaceStorageFailed)
	}
	if err = os.Rename(tempPath, target); err != nil {
		_ = os.Remove(tempPath)
		return bizerr.WrapCode(err, CodeMarketplaceStorageFailed)
	}
	return nil
}

// PutFile copies a local file into the store under key.
func (s *LocalArtifactStore) PutFile(ctx context.Context, key string, localPath string) error {
	file, err := os.Open(strings.TrimSpace(localPath))
	if err != nil {
		return bizerr.WrapCode(err, CodeMarketplaceStorageFailed)
	}
	defer func() {
		_ = file.Close()
	}()
	return s.Put(ctx, key, file)
}

// Open opens one stored object for reading.
func (s *LocalArtifactStore) Open(_ context.Context, key string) (io.ReadCloser, error) {
	target, err := s.resolvePath(key)
	if err != nil {
		return nil, err
	}
	file, err := os.Open(target)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, bizerr.NewCode(CodeMarketplaceDownloadArtifactNotFound)
		}
		return nil, bizerr.WrapCode(err, CodeMarketplaceStorageFailed)
	}
	return file, nil
}

// LocalPath returns the absolute filesystem path for one storage key.
func (s *LocalArtifactStore) LocalPath(_ context.Context, key string) (string, error) {
	return s.resolvePath(key)
}

// resolvePath maps a storage key to a path under the store root and rejects
// traversal outside that root.
func (s *LocalArtifactStore) resolvePath(key string) (string, error) {
	if s == nil || strings.TrimSpace(s.root) == "" {
		return "", bizerr.NewCode(CodeMarketplaceStorageFailed)
	}
	normalized := strings.TrimSpace(key)
	normalized = strings.ReplaceAll(normalized, "\\", "/")
	normalized = strings.TrimPrefix(normalized, "/")
	if normalized == "" {
		return "", bizerr.NewCode(CodeMarketplaceInvalidInput)
	}
	if strings.Contains(normalized, "..") {
		return "", bizerr.NewCode(CodeMarketplaceInvalidInput)
	}
	target := filepath.Join(s.root, filepath.FromSlash(normalized))
	rel, err := filepath.Rel(s.root, target)
	if err != nil || strings.HasPrefix(rel, "..") || filepath.IsAbs(rel) {
		return "", bizerr.NewCode(CodeMarketplaceInvalidInput)
	}
	return target, nil
}
