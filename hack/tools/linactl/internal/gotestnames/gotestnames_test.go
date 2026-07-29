// This file verifies strict 1:1 Go unit-test file naming and allowlist
// filtering used by linactl lint.go.

package gotestnames

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestCheckRequiresExactProductionSibling verifies X_test.go needs X.go.
func TestCheckRequiresExactProductionSibling(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	pkg := filepath.Join(root, "apps", "lina-core", "internal", "service", "demo")
	if err := os.MkdirAll(pkg, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	mustWrite(t, filepath.Join(pkg, "demo.go"), "package demo\n")
	mustWrite(t, filepath.Join(pkg, "demo_test.go"), "package demo\n")
	mustWrite(t, filepath.Join(pkg, "demo_batch_delete_test.go"), "package demo\n")
	mustWrite(t, filepath.Join(pkg, "demo_dbdriver_test.go"), "package demo\n")

	allowlistDir := filepath.Join(root, "hack", "tools", "linactl", "internal", "gotestnames")
	if err := os.MkdirAll(allowlistDir, 0o755); err != nil {
		t.Fatalf("mkdir allowlist: %v", err)
	}
	allowlistPath := filepath.Join(allowlistDir, "allowlist.json")
	mustWrite(t, allowlistPath, `{"paths":[]}`+"\n")

	var buf bytes.Buffer
	err := Check(root, &buf, Options{})
	if err == nil {
		t.Fatalf("expected unpaired error, output=%s", buf.String())
	}
	if !strings.Contains(buf.String(), "demo_batch_delete_test.go") {
		t.Fatalf("expected aspect split rejected, got %s", buf.String())
	}
	if !strings.Contains(buf.String(), "demo_dbdriver_test.go") {
		t.Fatalf("expected dbdriver exception removed, got %s", buf.String())
	}
	if strings.Contains(buf.String(), "demo_test.go") {
		t.Fatalf("exact pair should pass, output=%s", buf.String())
	}

	mustWrite(t, allowlistPath, `{"paths":[
		"apps/lina-core/internal/service/demo/demo_batch_delete_test.go",
		"apps/lina-core/internal/service/demo/demo_dbdriver_test.go"
	]}`+"\n")
	buf.Reset()
	if err = Check(root, &buf, Options{}); err != nil {
		t.Fatalf("allowlisted orphans should pass: %v\n%s", err, buf.String())
	}
}

// TestCheckScopeDirsLimitsScan ensures dir-scoped lint only checks that tree.
func TestCheckScopeDirsLimitsScan(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	badPkg := filepath.Join(root, "apps", "lina-core", "internal", "service", "bad")
	goodPkg := filepath.Join(root, "apps", "lina-core", "internal", "service", "good")
	for _, pkg := range []string{badPkg, goodPkg} {
		if err := os.MkdirAll(pkg, 0o755); err != nil {
			t.Fatalf("mkdir: %v", err)
		}
	}
	mustWrite(t, filepath.Join(badPkg, "bad.go"), "package bad\n")
	mustWrite(t, filepath.Join(badPkg, "orphan_test.go"), "package bad\n")
	mustWrite(t, filepath.Join(goodPkg, "good.go"), "package good\n")
	mustWrite(t, filepath.Join(goodPkg, "good_test.go"), "package good\n")

	var buf bytes.Buffer
	if err := Check(root, &buf, Options{ScopeDirs: []string{goodPkg}}); err != nil {
		t.Fatalf("scoped good package should pass: %v\n%s", err, buf.String())
	}

	buf.Reset()
	if err := Check(root, &buf, Options{ScopeDirs: []string{badPkg}}); err == nil {
		t.Fatalf("scoped bad package should fail, output=%s", buf.String())
	}
}

// TestRepositoryScanPasses ensures the live repository has no unpaired
// *_test.go outside the frozen allowlist.
func TestRepositoryScanPasses(t *testing.T) {
	t.Parallel()

	root := findRepoRoot(t)
	var buf bytes.Buffer
	if err := Check(root, &buf, Options{}); err != nil {
		t.Fatalf("repository gotestnames check failed: %v\n%s", err, buf.String())
	}
}

func findRepoRoot(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	dir := wd
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.work")); err == nil {
			return dir
		}
		if _, err := os.Stat(filepath.Join(dir, "hack", "tools", "linactl")); err == nil {
			if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
				return dir
			}
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatalf("could not locate repository root from %s", wd)
		}
		dir = parent
	}
}

func mustWrite(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}
