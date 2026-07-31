// This file verifies source ZIP scanner behavior for valid packages, unsafe
// archive paths, and plugin.yaml identity mismatches.

package marketplace

import (
	"archive/zip"
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"lina-core/pkg/bizerr"
)

// uploadOwnershipArtifactStore counts upload storage side effects.
type uploadOwnershipArtifactStore struct {
	putCalls     int
	putFileCalls int
}

// Put records an in-memory object write.
func (s *uploadOwnershipArtifactStore) Put(context.Context, string, io.Reader) error {
	s.putCalls++
	return nil
}

// PutFile records a local-file object write.
func (s *uploadOwnershipArtifactStore) PutFile(context.Context, string, string) error {
	s.putFileCalls++
	return nil
}

// Open is unused by upload ownership tests.
func (*uploadOwnershipArtifactStore) Open(context.Context, string) (io.ReadCloser, error) {
	return nil, nil
}

// LocalPath is unused by upload ownership tests.
func (*uploadOwnershipArtifactStore) LocalPath(context.Context, string) (string, error) {
	return "", nil
}

// Delete is unused by upload ownership tests.
func (*uploadOwnershipArtifactStore) Delete(context.Context, string) error {
	return nil
}

func TestUploadSourcePackageRequiresOwnerBeforeArtifactStorage(t *testing.T) {
	store := &uploadOwnershipArtifactStore{}
	service := &serviceImpl{artifacts: store}

	_, err := service.UploadSourcePackage(context.Background(), UploadSourcePackageInput{})
	if !bizerr.Is(err, CodeMarketplaceInvalidInput) {
		t.Fatalf("expected invalid owner error, got %v", err)
	}
	if store.putCalls != 0 || store.putFileCalls != 0 {
		t.Fatalf("unexpected artifact writes before ownership validation: %#v", store)
	}
}

func TestScanSourcePackageAcceptsSingleRootPackage(t *testing.T) {
	packagePath := writeSourcePackageZip(t, "linapro-demo-source/", map[string]string{
		"plugin.yaml": `id: linapro-demo-source
name: LinaPro Demo Source
version: v0.1.0
type: source
distribution: managed
dependencies:
  framework:
    version: ">=v0.1.0 <v1.0.0"
i18n:
  enabled: true
  default: en-US
  locales:
    - locale: en-US
      nativeName: English
`,
		"go.mod":                           "module linapro-demo-source\n",
		"plugin_embed.go":                  "package linapro_demo_source\n",
		"backend/plugin.go":                "package backend\n",
		"frontend/pages/index.vue":         "<template><div /></template>\n",
		"manifest/docs/en-US/index.md":     "# Demo\n",
		"manifest/i18n/en-US/plugin.json":  "{}\n",
		"manifest/sql/001-demo-source.sql": "CREATE TABLE IF NOT EXISTS demo_source (id INT);\n",
	})

	scan, err := scanSourcePackage(UploadSourcePackageInput{
		PluginID:    "linapro-demo-source",
		Version:     "v0.1.0",
		PackagePath: packagePath,
		FileName:    "linapro-demo-source.zip",
	})
	if err != nil {
		t.Fatalf("scanSourcePackage returned error: %v", err)
	}
	if scan.manifest.ID != "linapro-demo-source" {
		t.Fatalf("unexpected plugin id: %s", scan.manifest.ID)
	}
	if scan.minHostVersion != "v0.1.0" || scan.maxHostVersion != "v1.0.0" {
		t.Fatalf("unexpected host bounds: min=%s max=%s", scan.minHostVersion, scan.maxHostVersion)
	}
	if scan.packageSha256 == "" || scan.manifestSha256 == "" {
		t.Fatal("expected package and manifest checksums")
	}
	if !strings.Contains(scan.sqlSummary, "install_sql") {
		t.Fatalf("expected install SQL summary, got %s", scan.sqlSummary)
	}
	if !strings.Contains(scan.docsSummary, "manifest/docs/en-US/index.md") {
		t.Fatalf("expected docs summary, got %s", scan.docsSummary)
	}
	if !strings.Contains(scan.i18nSummary, "manifest/i18n/en-US/plugin.json") {
		t.Fatalf("expected i18n summary, got %s", scan.i18nSummary)
	}
}

func TestScanSourcePackageRejectsTraversalPath(t *testing.T) {
	packagePath := writeSourcePackageZip(t, "", map[string]string{
		"../plugin.yaml": "id: linapro-demo-source\n",
	})

	_, err := scanSourcePackage(UploadSourcePackageInput{
		PluginID:    "linapro-demo-source",
		Version:     "v0.1.0",
		PackagePath: packagePath,
		FileName:    "linapro-demo-source.zip",
	})
	if !bizerr.Is(err, CodeMarketplacePackageInvalid) {
		t.Fatalf("expected package invalid error, got %v", err)
	}
}

func TestScanSourcePackageRejectsMissingSourceStructure(t *testing.T) {
	cases := []struct {
		name  string
		files map[string]string
	}{
		{
			name: "missing backend entry",
			files: map[string]string{
				"plugin.yaml":                  sourcePackageTestManifest("v0.1.0"),
				"go.mod":                       "module linapro-demo-source\n",
				"plugin_embed.go":              "package linapro_demo_source\n",
				"frontend/pages/index.vue":     "<template><div /></template>\n",
				"manifest/docs/en-US/index.md": "# Demo\n",
			},
		},
		{
			name: "missing frontend directory",
			files: map[string]string{
				"plugin.yaml":                  sourcePackageTestManifest("v0.1.0"),
				"go.mod":                       "module linapro-demo-source\n",
				"plugin_embed.go":              "package linapro_demo_source\n",
				"backend/plugin.go":            "package backend\n",
				"manifest/docs/en-US/index.md": "# Demo\n",
			},
		},
		{
			name: "missing plugin embed",
			files: map[string]string{
				"plugin.yaml":                  sourcePackageTestManifest("v0.1.0"),
				"go.mod":                       "module linapro-demo-source\n",
				"backend/plugin.go":            "package backend\n",
				"frontend/pages/index.vue":     "<template><div /></template>\n",
				"manifest/docs/en-US/index.md": "# Demo\n",
			},
		},
		{
			name: "missing docs markdown",
			files: map[string]string{
				"plugin.yaml":                   sourcePackageTestManifest("v0.1.0"),
				"go.mod":                        "module linapro-demo-source\n",
				"plugin_embed.go":               "package linapro_demo_source\n",
				"backend/plugin.go":             "package backend\n",
				"frontend/pages/index.vue":      "<template><div /></template>\n",
				"manifest/docs/assets/logo.txt": "asset\n",
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			packagePath := writeSourcePackageZip(t, "", tc.files)
			_, err := scanSourcePackage(UploadSourcePackageInput{
				PluginID:    "linapro-demo-source",
				Version:     "v0.1.0",
				PackagePath: packagePath,
				FileName:    "linapro-demo-source.zip",
			})
			if !bizerr.Is(err, CodeMarketplacePackageStructureInvalid) {
				t.Fatalf("expected package structure invalid error, got %v", err)
			}
		})
	}
}

func TestScanSourcePackageRejectsManifestMismatch(t *testing.T) {
	packagePath := writeSourcePackageZip(t, "", map[string]string{
		"plugin.yaml": `id: linapro-demo-source
name: LinaPro Demo Source
version: v0.2.0
type: source
`,
		"go.mod":                       "module linapro-demo-source\n",
		"plugin_embed.go":              "package linapro_demo_source\n",
		"backend/plugin.go":            "package backend\n",
		"frontend/pages/index.vue":     "<template><div /></template>\n",
		"manifest/docs/en-US/index.md": "# Demo\n",
	})

	_, err := scanSourcePackage(UploadSourcePackageInput{
		PluginID:    "linapro-demo-source",
		Version:     "v0.1.0",
		PackagePath: packagePath,
		FileName:    "linapro-demo-source.zip",
	})
	if !bizerr.Is(err, CodeMarketplacePackageManifestMismatch) {
		t.Fatalf("expected package manifest mismatch error, got %v", err)
	}
}

func sourcePackageTestManifest(version string) string {
	return `id: linapro-demo-source
name: LinaPro Demo Source
version: ` + version + `
type: source
distribution: managed
`
}

func writeSourcePackageZip(t *testing.T, rootPrefix string, files map[string]string) string {
	t.Helper()

	zipPath := filepath.Join(t.TempDir(), "source-package.zip")
	file, err := os.Create(zipPath)
	if err != nil {
		t.Fatalf("create zip file: %v", err)
	}
	writer := zip.NewWriter(file)
	for name, content := range files {
		entry, createErr := writer.Create(rootPrefix + name)
		if createErr != nil {
			t.Fatalf("create zip entry %s: %v", name, createErr)
		}
		if _, writeErr := entry.Write([]byte(content)); writeErr != nil {
			t.Fatalf("write zip entry %s: %v", name, writeErr)
		}
	}
	if err = writer.Close(); err != nil {
		t.Fatalf("close zip writer: %v", err)
	}
	if err = file.Close(); err != nil {
		t.Fatalf("close zip file: %v", err)
	}
	return zipPath
}
