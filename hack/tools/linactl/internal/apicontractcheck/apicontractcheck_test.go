// This file verifies generic API contract scanning used by linactl lint.go.

package apicontractcheck

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestCheckDetectsEntityImportAndForbiddenResponseFields covers both generic rules.
func TestCheckDetectsEntityImportAndForbiddenResponseFields(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	apiDir := filepath.Join(root, "apps", "lina-core", "api", "demo", "v1")
	if err := os.MkdirAll(apiDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	mustWrite(t, filepath.Join(apiDir, "demo.go"), `package v1

import (
	entity "lina-core/internal/model/entity"
)

type DemoItem struct {
	Password string `+"`"+`json:"password"`+"`"+`
	Name     string `+"`"+`json:"name"`+"`"+`
	Ref      entity.SysUser
}

type DemoReq struct {
	Password string `+"`"+`json:"password"`+"`"+`
}
`)

	var buf bytes.Buffer
	err := Check(root, &buf, Options{})
	if err == nil {
		t.Fatalf("expected findings, output=%s", buf.String())
	}
	out := buf.String()
	if !strings.Contains(out, "must not import internal model/entity") {
		t.Fatalf("expected entity import finding, got %s", out)
	}
	if !strings.Contains(out, `DemoItem exposes forbidden JSON field "password"`) {
		t.Fatalf("expected password finding on response item, got %s", out)
	}
	if strings.Contains(out, "DemoReq") {
		t.Fatalf("request DTOs must not be scanned for password, got %s", out)
	}
}

// TestCheckAllowsMenuPathAndRejectsFilePath keeps path heuristic scoped.
func TestCheckAllowsMenuPathAndRejectsFilePath(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	apiDir := filepath.Join(root, "apps", "lina-core", "api", "mixed", "v1")
	if err := os.MkdirAll(apiDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	mustWrite(t, filepath.Join(apiDir, "mixed.go"), `package v1

type MenuItem struct {
	Path string `+"`"+`json:"path"`+"`"+`
}

type FileItem struct {
	Path string `+"`"+`json:"path"`+"`"+`
}
`)

	var buf bytes.Buffer
	err := Check(root, &buf, Options{})
	if err == nil {
		t.Fatalf("expected file path finding, output=%s", buf.String())
	}
	out := buf.String()
	if !strings.Contains(out, `FileItem exposes forbidden JSON field "path"`) {
		t.Fatalf("expected FileItem path finding, got %s", out)
	}
	if strings.Contains(out, "MenuItem") {
		t.Fatalf("MenuItem path should be allowed, got %s", out)
	}
}

// TestCheckPassesCleanTree verifies a valid API tree has zero findings.
func TestCheckPassesCleanTree(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	apiDir := filepath.Join(root, "apps", "lina-core", "api", "clean", "v1")
	if err := os.MkdirAll(apiDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	mustWrite(t, filepath.Join(apiDir, "clean.go"), `package v1

type CleanItem struct {
	Id   int    `+"`"+`json:"id"`+"`"+`
	Name string `+"`"+`json:"name"`+"`"+`
}

type CleanReq struct {
	Password string `+"`"+`json:"password"`+"`"+`
}
`)

	var buf bytes.Buffer
	if err := Check(root, &buf, Options{}); err != nil {
		t.Fatalf("expected clean tree to pass: %v\n%s", err, buf.String())
	}
}

// TestRepositoryHostAPIPasses ensures live host api/ currently satisfies rules.
func TestRepositoryHostAPIPasses(t *testing.T) {
	t.Parallel()

	root := findRepoRoot(t)
	var buf bytes.Buffer
	if err := Check(root, &buf, Options{}); err != nil {
		t.Fatalf("host API contract check failed: %v\n%s", err, buf.String())
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
