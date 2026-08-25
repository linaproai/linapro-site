// This file verifies the plugin controller constructor depends on the
// composed plugin Service instead of a split management facet.

package plugin

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"

	pluginsvc "lina-core/internal/service/plugin"
)

// TestNewV1DependsOnPluginService locks the plugin controller to the composed
// plugin Service so management HTTP handlers cannot compile against a split
// management, startup, or runtime facet.
func TestNewV1DependsOnPluginService(t *testing.T) {
	want := reflect.TypeOf((*pluginsvc.Service)(nil)).Elem()
	got := reflect.TypeOf(NewV1).In(0)
	if got != want {
		t.Fatalf("NewV1 plugin dependency type = %s, want pluginsvc.Service", got)
	}
}

// TestHostControllersDoNotDependOnSplitServiceFacets scans host HTTP
// controller constructors and rejects dependencies on interfaces split out of
// a component Service. Registry and plugin capability SPI contracts are
// independent products, not Service facets.
func TestHostControllersDoNotDependOnSplitServiceFacets(t *testing.T) {
	forbiddenSuffixes := []string{
		"Management",
		"Startup",
		"Runtime",
		"Reader",
		"Writer",
		"Store",
		"Directory",
		"LockStore",
	}
	allowedExact := map[string]struct{}{
		"Registry": {},
		"Service":  {},
	}

	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve controller test file path")
	}
	parent := filepath.Dir(filepath.Dir(thisFile))
	if filepath.Base(parent) != "controller" {
		t.Fatalf("expected test to run beside controller packages, parent=%s", parent)
	}

	entries, err := os.ReadDir(parent)
	if err != nil {
		t.Fatalf("read controller directory: %v", err)
	}

	fset := token.NewFileSet()
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		newFile := filepath.Join(parent, entry.Name(), entry.Name()+"_new.go")
		src, readErr := os.ReadFile(newFile)
		if readErr != nil {
			if os.IsNotExist(readErr) {
				continue
			}
			t.Fatalf("read %s: %v", newFile, readErr)
		}
		file, parseErr := parser.ParseFile(fset, newFile, src, 0)
		if parseErr != nil {
			t.Fatalf("parse %s: %v", newFile, parseErr)
		}
		ast.Inspect(file, func(node ast.Node) bool {
			funcDecl, ok := node.(*ast.FuncDecl)
			if !ok || funcDecl.Name == nil || funcDecl.Name.Name != "NewV1" || funcDecl.Type == nil || funcDecl.Type.Params == nil {
				return true
			}
			for _, field := range funcDecl.Type.Params.List {
				typeName := selectorTypeName(field.Type)
				if typeName == "" {
					t.Fatalf("%s: NewV1 has unsupported parameter type", newFile)
				}
				if _, allowed := allowedExact[typeName]; allowed {
					continue
				}
				for _, suffix := range forbiddenSuffixes {
					if typeName == suffix || strings.HasSuffix(typeName, suffix) {
						t.Fatalf("%s: NewV1 depends on split service facet %s; use the component Service", newFile, typeName)
					}
				}
			}
			return true
		})
	}
}

// selectorTypeName returns the imported or local identifier of a constructor
// parameter type such as pluginsvc.Service.
func selectorTypeName(expr ast.Expr) string {
	switch typed := expr.(type) {
	case *ast.Ident:
		return typed.Name
	case *ast.SelectorExpr:
		return typed.Sel.Name
	case *ast.StarExpr:
		return selectorTypeName(typed.X)
	default:
		return ""
	}
}
