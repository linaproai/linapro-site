// Package apicontractcheck enforces generic public API DTO contract rules for
// linactl lint.go. It walks API DTO trees under the repository without a
// per-resource type registry: new api/<resource>/v1 definitions are covered
// automatically.
//
// Rules:
//  1. Production Go files under API roots must not import internal entity packages.
//  2. Exported response-shaped structs must not expose sensitive JSON field names
//     (password, deletedAt, engine, hash; path only on file-related type names).
//  3. Response-shaped struct fields must not embed/reference entity package types
//     (defense in depth when imports are aliased).
package apicontractcheck

import (
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Options configures an API contract scan.
type Options struct {
	// ScopeDirs limits discovery to API roots under these module directories.
	// Empty scans the default repository API roots.
	ScopeDirs []string
}

// Finding is one contract violation (slash path, repo-relative when possible).
type Finding struct {
	Path   string
	Reason string
}

// Check scans API DTO roots and reports contract violations to out.
func Check(repoRoot string, out io.Writer, opts Options) error {
	root := filepath.Clean(strings.TrimSpace(repoRoot))
	if root == "" || root == "." {
		return errors.New("apicontractcheck: repository root is required")
	}

	apiRoots, err := discoverAPIRoots(root, opts.ScopeDirs)
	if err != nil {
		return err
	}
	if len(apiRoots) == 0 {
		if out != nil {
			fmt.Fprintln(out, "API contract check: scanned roots=0 findings=0")
		}
		return nil
	}

	var findings []Finding
	for _, apiRoot := range apiRoots {
		part, scanErr := scanAPIRoot(root, apiRoot)
		if scanErr != nil {
			return scanErr
		}
		findings = append(findings, part...)
	}
	sort.Slice(findings, func(i, j int) bool {
		if findings[i].Path == findings[j].Path {
			return findings[i].Reason < findings[j].Reason
		}
		return findings[i].Path < findings[j].Path
	})

	if out != nil {
		fmt.Fprintf(out, "API contract check: scanned roots=%d findings=%d\n", len(apiRoots), len(findings))
		for _, finding := range findings {
			fmt.Fprintf(out, "- %s: %s\n", finding.Path, finding.Reason)
		}
	}
	if len(findings) == 0 {
		return nil
	}
	return fmt.Errorf("apicontractcheck: %d API contract violation(s)", len(findings))
}

// discoverAPIRoots finds host and official-plugin API DTO trees.
func discoverAPIRoots(repoRoot string, scopeDirs []string) ([]string, error) {
	var candidates []string
	if len(scopeDirs) == 0 {
		candidates = append(candidates, filepath.Join(repoRoot, "apps", "lina-core", "api"))
		pluginsRoot := filepath.Join(repoRoot, "apps", "lina-plugins")
		entries, err := os.ReadDir(pluginsRoot)
		if err != nil && !os.IsNotExist(err) {
			return nil, fmt.Errorf("apicontractcheck: read plugins root: %w", err)
		}
		for _, entry := range entries {
			if !entry.IsDir() || strings.HasPrefix(entry.Name(), ".") {
				continue
			}
			candidates = append(candidates, filepath.Join(pluginsRoot, entry.Name(), "backend", "api"))
		}
	} else {
		for _, dir := range scopeDirs {
			dir = strings.TrimSpace(dir)
			if dir == "" {
				continue
			}
			if !filepath.IsAbs(dir) {
				dir = filepath.Join(repoRoot, filepath.FromSlash(dir))
			}
			abs, err := filepath.Abs(dir)
			if err != nil {
				return nil, fmt.Errorf("apicontractcheck: resolve scope %s: %w", dir, err)
			}
			// Module dir may be apps/lina-core or plugin backend.
			for _, rel := range []string{"api", filepath.Join("backend", "api")} {
				candidates = append(candidates, filepath.Join(abs, rel))
			}
			// If scope already points at .../api, include it.
			if filepath.Base(abs) == "api" {
				candidates = append(candidates, abs)
			}
		}
	}

	seen := make(map[string]struct{})
	var roots []string
	for _, candidate := range candidates {
		clean := filepath.Clean(candidate)
		info, err := os.Stat(clean)
		if err != nil || !info.IsDir() {
			continue
		}
		if _, ok := seen[clean]; ok {
			continue
		}
		// Only treat as an API DTO root when it contains production .go files.
		hasGo, err := dirHasProductionGo(clean)
		if err != nil {
			return nil, err
		}
		if !hasGo {
			continue
		}
		seen[clean] = struct{}{}
		roots = append(roots, clean)
	}
	sort.Strings(roots)
	return roots, nil
}

// dirHasProductionGo reports whether root or a descendant has non-test Go files.
func dirHasProductionGo(root string) (bool, error) {
	found := false
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			name := d.Name()
			if name == "testdata" || name == "vendor" || strings.HasPrefix(name, ".") {
				return fs.SkipDir
			}
			return nil
		}
		if strings.HasSuffix(d.Name(), ".go") && !strings.HasSuffix(d.Name(), "_test.go") {
			found = true
			return fs.SkipAll
		}
		return nil
	})
	if err != nil && !errors.Is(err, fs.SkipAll) {
		return false, err
	}
	return found, nil
}

// scanAPIRoot walks one API tree and collects findings.
func scanAPIRoot(repoRoot, apiRoot string) ([]Finding, error) {
	var findings []Finding
	err := filepath.WalkDir(apiRoot, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			name := d.Name()
			if name == "testdata" || name == "vendor" || strings.HasPrefix(name, ".") {
				return fs.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(d.Name(), ".go") || strings.HasSuffix(d.Name(), "_test.go") {
			return nil
		}
		fileFindings, err := scanFile(repoRoot, path)
		if err != nil {
			return err
		}
		findings = append(findings, fileFindings...)
		return nil
	})
	return findings, err
}

// scanFile parses one production API Go file for contract violations.
func scanFile(repoRoot, path string) ([]Finding, error) {
	fileSet := token.NewFileSet()
	file, err := parser.ParseFile(fileSet, path, nil, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("apicontractcheck: parse %s: %w", path, err)
	}

	rel := path
	if r, relErr := filepath.Rel(repoRoot, path); relErr == nil {
		rel = filepath.ToSlash(r)
	}

	var findings []Finding
	entityAliases := entityImportAliases(file)
	if len(entityAliases) > 0 {
		for _, alias := range entityAliases {
			_ = alias
		}
		findings = append(findings, Finding{
			Path:   rel,
			Reason: "must not import internal model/entity packages in public API DTOs",
		})
	}

	for _, decl := range file.Decls {
		gen, ok := decl.(*ast.GenDecl)
		if !ok || gen.Tok != token.TYPE {
			continue
		}
		for _, spec := range gen.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok || !typeSpec.Name.IsExported() {
				continue
			}
			structType, ok := typeSpec.Type.(*ast.StructType)
			if !ok {
				continue
			}
			typeName := typeSpec.Name.Name
			if !isResponseShapedType(typeName) {
				continue
			}
			findings = append(findings, scanResponseStruct(rel, typeName, structType, entityAliases)...)
		}
	}
	return findings, nil
}

// entityImportAliases returns import names that resolve to internal entity packages.
// The map key is the identifier used in code (alias or default last path segment).
func entityImportAliases(file *ast.File) map[string]struct{} {
	out := make(map[string]struct{})
	for _, imp := range file.Imports {
		path := strings.Trim(imp.Path.Value, `"`)
		if !isEntityImportPath(path) {
			continue
		}
		name := filepath.Base(path)
		if imp.Name != nil {
			name = imp.Name.Name
			if name == "." || name == "_" {
				// Dot-import is still a banned dependency; caller reports the import finding.
				name = filepath.Base(path)
			}
		}
		out[name] = struct{}{}
	}
	return out
}

// isEntityImportPath reports whether importPath is an internal entity package.
func isEntityImportPath(importPath string) bool {
	// Match .../internal/model/entity and nested paths under it.
	return strings.Contains(importPath, "/internal/model/entity") ||
		strings.HasSuffix(importPath, "internal/model/entity")
}

// isResponseShapedType uses naming conventions so new DTOs need no registry.
// Request types (Req) are excluded; password etc. remain valid on write DTOs.
func isResponseShapedType(typeName string) bool {
	if typeName == "" || strings.HasSuffix(typeName, "Req") {
		return false
	}
	// Common response / projection suffixes used across host and plugins.
	for _, suffix := range []string{
		"Res",
		"Item",
		"Items",
		"Option",
		"Options",
		"Row",
		"Rows",
		"View",
		"Summary",
		"Detail",
		"Info",
		"Payload",
	} {
		if strings.HasSuffix(typeName, suffix) {
			return true
		}
	}
	return false
}

// scanResponseStruct checks json tags and entity-typed fields on one struct.
func scanResponseStruct(relPath, typeName string, structType *ast.StructType, entityAliases map[string]struct{}) []Finding {
	if structType.Fields == nil {
		return nil
	}
	var findings []Finding
	for _, field := range structType.Fields.List {
		// Embedded fields: recurse into local struct literals only; named embeds checked via type expr.
		if len(field.Names) == 0 {
			if st, ok := field.Type.(*ast.StructType); ok {
				findings = append(findings, scanResponseStruct(relPath, typeName, st, entityAliases)...)
			}
			if fieldTypeRefersEntity(field.Type, entityAliases) {
				findings = append(findings, Finding{
					Path:   relPath,
					Reason: fmt.Sprintf("%s embeds/references internal entity type", typeName),
				})
			}
			continue
		}
		jsonName := jsonFieldName(field)
		if jsonName != "" && jsonName != "-" && isForbiddenResponseJSON(typeName, jsonName) {
			findings = append(findings, Finding{
				Path:   relPath,
				Reason: fmt.Sprintf("%s exposes forbidden JSON field %q", typeName, jsonName),
			})
		}
		if fieldTypeRefersEntity(field.Type, entityAliases) {
			findings = append(findings, Finding{
				Path:   relPath,
				Reason: fmt.Sprintf("%s field references internal entity type", typeName),
			})
		}
	}
	return findings
}

// jsonFieldName extracts the primary json tag name from a field.
func jsonFieldName(field *ast.Field) string {
	if field.Tag == nil {
		return ""
	}
	tag := field.Tag.Value
	// tag is a raw string literal including backticks.
	tag = strings.Trim(tag, "`")
	// Parse loosely: look for json:"name,opts"
	const prefix = `json:"`
	idx := strings.Index(tag, prefix)
	if idx < 0 {
		return ""
	}
	rest := tag[idx+len(prefix):]
	end := strings.IndexByte(rest, '"')
	if end < 0 {
		return ""
	}
	value := rest[:end]
	name, _, _ := strings.Cut(value, ",")
	return name
}

// isForbiddenResponseJSON reports sensitive response field names.
// `path` is only forbidden on file-related types (menu/file routing path is valid).
func isForbiddenResponseJSON(typeName, jsonName string) bool {
	switch jsonName {
	case "password", "deletedAt", "engine", "hash":
		return true
	case "path":
		return strings.Contains(strings.ToLower(typeName), "file")
	default:
		return false
	}
}

// fieldTypeRefersEntity walks a type expression for entity package selectors.
func fieldTypeRefersEntity(expr ast.Expr, entityAliases map[string]struct{}) bool {
	if expr == nil || len(entityAliases) == 0 {
		return false
	}
	switch t := expr.(type) {
	case *ast.StarExpr:
		return fieldTypeRefersEntity(t.X, entityAliases)
	case *ast.ArrayType:
		return fieldTypeRefersEntity(t.Elt, entityAliases)
	case *ast.MapType:
		return fieldTypeRefersEntity(t.Key, entityAliases) || fieldTypeRefersEntity(t.Value, entityAliases)
	case *ast.SelectorExpr:
		if id, ok := t.X.(*ast.Ident); ok {
			if _, banned := entityAliases[id.Name]; banned {
				return true
			}
		}
		return false
	case *ast.Ident:
		return false
	case *ast.StructType:
		if t.Fields == nil {
			return false
		}
		for _, field := range t.Fields.List {
			if fieldTypeRefersEntity(field.Type, entityAliases) {
				return true
			}
		}
		return false
	default:
		return false
	}
}
