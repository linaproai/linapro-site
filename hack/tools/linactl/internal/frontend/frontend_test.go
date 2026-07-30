package frontend

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"linactl/internal/toolutil"
)

type capturedRun struct {
	dir  string
	name string
	args []string
}

func TestEnsureDepsInstallsPluginFrontendPackages(t *testing.T) {
	root := t.TempDir()
	writeTestFile(t, toolutil.ViteCommand(root), "")
	writeTestFile(t, filepath.Join(root, "apps", "lina-plugins", "alpha-plugin", "plugin.yaml"), "id: alpha-plugin\ntype: source\n")
	writeTestFile(t, filepath.Join(root, "apps", "lina-plugins", "alpha-plugin", "frontend", "package.json"), "{}\n")
	writeTestFile(t, filepath.Join(root, "apps", "lina-plugins", "beta-plugin", "plugin.yaml"), "id: beta-plugin\ntype: source\n")
	writeTestFile(t, filepath.Join(root, "apps", "lina-plugins", "beta-plugin", "frontend", "package.json"), "{}\n")
	if err := os.MkdirAll(filepath.Join(root, "apps", "lina-plugins", "beta-plugin", "frontend", "node_modules"), 0o755); err != nil {
		t.Fatalf("mkdir node_modules: %v", err)
	}
	writeTestFile(t, filepath.Join(root, "apps", "lina-plugins", "not-a-plugin", "frontend", "package.json"), "{}\n")
	writeTestFile(t, filepath.Join(root, "apps", "lina-plugins", "dynamic-plugin", "plugin.yaml"), "id: dynamic-plugin\ntype: dynamic\n")
	writeTestFile(t, filepath.Join(root, "apps", "lina-plugins", "dynamic-plugin", "frontend", "package.json"), "{}\n")

	var out bytes.Buffer
	var calls []capturedRun
	err := EnsureDeps(context.Background(), root, &out, func(_ context.Context, dir string, name string, args ...string) error {
		calls = append(calls, capturedRun{dir: dir, name: name, args: append([]string(nil), args...)})
		return nil
	})
	if err != nil {
		t.Fatalf("EnsureDeps returned error: %v", err)
	}

	expected := []capturedRun{
		{
			dir:  filepath.Join(root, "apps", "lina-plugins", "alpha-plugin", "frontend"),
			name: "pnpm",
			args: []string{"install", "--config.auto-install-peers=false"},
		},
	}
	if !reflect.DeepEqual(calls, expected) {
		t.Fatalf("unexpected commands:\ngot: %#v\nwant: %#v", calls, expected)
	}
	if !bytes.Contains(out.Bytes(), []byte("apps")) || !bytes.Contains(out.Bytes(), []byte("alpha-plugin")) {
		t.Fatalf("expected plugin install output to mention package path, got %q", out.String())
	}
}

func TestEnsureDepsInstallsHostBeforePluginFrontendPackages(t *testing.T) {
	root := t.TempDir()
	writeTestFile(t, filepath.Join(root, "apps", "lina-plugins", "alpha-plugin", "plugin.yaml"), "id: alpha-plugin\ntype: source\n")
	writeTestFile(t, filepath.Join(root, "apps", "lina-plugins", "alpha-plugin", "frontend", "package.json"), "{}\n")

	var calls []capturedRun
	err := EnsureDeps(context.Background(), root, bytes.NewBuffer(nil), func(_ context.Context, dir string, name string, args ...string) error {
		calls = append(calls, capturedRun{dir: dir, name: name, args: append([]string(nil), args...)})
		return nil
	})
	if err != nil {
		t.Fatalf("EnsureDeps returned error: %v", err)
	}

	expected := []capturedRun{
		{
			dir:  filepath.Join(root, "apps", "lina-vben"),
			name: "pnpm",
			args: []string{"install"},
		},
		{
			dir:  filepath.Join(root, "apps", "lina-plugins", "alpha-plugin", "frontend"),
			name: "pnpm",
			args: []string{"install", "--config.auto-install-peers=false"},
		},
	}
	if !reflect.DeepEqual(calls, expected) {
		t.Fatalf("unexpected commands:\ngot: %#v\nwant: %#v", calls, expected)
	}
}

func writeTestFile(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}
