// Package gotestnames enforces 1:1 Go unit-test file naming: each *_test.go
// must have a sibling production file with the same stem (X.go ↔ X_test.go).
//
// linactl lint.go runs Check before golangci-lint. Historical unpaired files
// may be listed in allowlist.json; new paths must not be added without fixing
// the name or merging into the paired test file.
package gotestnames

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const (
	// DefaultAllowlistRel is the repository-relative path of the frozen orphan list.
	DefaultAllowlistRel = "hack/tools/linactl/internal/gotestnames/allowlist.json"
)

// Options configures a repository scan.
type Options struct {
	// AllowlistPath overrides the default allowlist file. Empty uses DefaultAllowlistRel under RepoRoot.
	AllowlistPath string
	// ScopeDirs limits the scan to packages under these absolute or repo-relative
	// directories (typically Go module roots from lint.go). Empty scans the full
	// default roots under the repository.
	ScopeDirs []string
}

// Finding is one unpaired *_test.go path (slash-separated, repo-relative).
type Finding struct {
	Path   string
	Reason string
}

// allowlistFile is the JSON shape of allowlist.json.
type allowlistFile struct {
	// Paths are slash-separated repository-relative *_test.go paths.
	Paths []string `json:"paths"`
}

// Check scans Go package directories for unit-test files that do not pair 1:1
// with a production source. It writes a short summary to out and returns a
// non-nil error when findings remain after allowlist filtering.
func Check(repoRoot string, out io.Writer, opts Options) error {
	root := filepath.Clean(strings.TrimSpace(repoRoot))
	if root == "" || root == "." {
		return errors.New("gotestnames: repository root is required")
	}

	allowlist, err := loadAllowlist(root, opts.AllowlistPath)
	if err != nil {
		return err
	}

	scopeAbs, err := normalizeScopeDirs(root, opts.ScopeDirs)
	if err != nil {
		return err
	}

	findings, err := scanRepository(root, scopeAbs)
	if err != nil {
		return err
	}

	var blocked []Finding
	for _, finding := range findings {
		if _, ok := allowlist[finding.Path]; ok {
			continue
		}
		blocked = append(blocked, finding)
	}
	sort.Slice(blocked, func(i, j int) bool { return blocked[i].Path < blocked[j].Path })

	if out != nil {
		fmt.Fprintf(
			out,
			"Go test name check: scanned findings=%d allowlisted=%d blocked=%d\n",
			len(findings),
			len(findings)-len(blocked),
			len(blocked),
		)
		for _, finding := range blocked {
			fmt.Fprintf(out, "- %s: %s\n", finding.Path, finding.Reason)
		}
	}
	if len(blocked) == 0 {
		return nil
	}
	return fmt.Errorf(
		"gotestnames: %d unpaired *_test.go file(s); require 1:1 name with production X.go → X_test.go (merge or rename; allowlist is frozen for historical only)",
		len(blocked),
	)
}

// loadAllowlist reads the JSON allowlist into a set of slash paths.
func loadAllowlist(repoRoot, override string) (map[string]struct{}, error) {
	path := strings.TrimSpace(override)
	if path == "" {
		path = filepath.Join(repoRoot, filepath.FromSlash(DefaultAllowlistRel))
	} else if !filepath.IsAbs(path) {
		path = filepath.Join(repoRoot, filepath.FromSlash(path))
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return map[string]struct{}{}, nil
		}
		return nil, fmt.Errorf("gotestnames: read allowlist %s: %w", path, err)
	}

	var file allowlistFile
	if err = json.Unmarshal(data, &file); err != nil {
		return nil, fmt.Errorf("gotestnames: parse allowlist %s: %w", path, err)
	}

	set := make(map[string]struct{}, len(file.Paths))
	for _, p := range file.Paths {
		p = strings.TrimSpace(filepath.ToSlash(p))
		if p == "" || strings.HasPrefix(p, "#") {
			continue
		}
		set[p] = struct{}{}
	}
	return set, nil
}

// normalizeScopeDirs converts scope entries to absolute cleaned paths.
func normalizeScopeDirs(repoRoot string, scopeDirs []string) ([]string, error) {
	if len(scopeDirs) == 0 {
		return nil, nil
	}
	out := make([]string, 0, len(scopeDirs))
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
			return nil, fmt.Errorf("gotestnames: resolve scope dir %s: %w", dir, err)
		}
		out = append(out, filepath.Clean(abs))
	}
	return out, nil
}

// scanRepository walks default roots (or scope) and collects unpaired tests.
func scanRepository(repoRoot string, scopeAbs []string) ([]Finding, error) {
	var roots []string
	if len(scopeAbs) > 0 {
		roots = append(roots, scopeAbs...)
	} else {
		for _, rel := range defaultScanRoots() {
			roots = append(roots, filepath.Join(repoRoot, filepath.FromSlash(rel)))
		}
	}

	var findings []Finding
	seenDirs := make(map[string]struct{})
	for _, root := range roots {
		info, err := os.Stat(root)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, fmt.Errorf("gotestnames: stat %s: %w", root, err)
		}
		if !info.IsDir() {
			continue
		}
		err = filepath.WalkDir(root, func(path string, d fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if d.IsDir() {
				name := d.Name()
				if shouldSkipDir(name) {
					return fs.SkipDir
				}
				return nil
			}
			if !strings.HasSuffix(d.Name(), ".go") {
				return nil
			}
			dir := filepath.Dir(path)
			if _, ok := seenDirs[dir]; ok {
				return nil
			}
			seenDirs[dir] = struct{}{}
			pkgFindings, err := scanPackageDir(repoRoot, dir)
			if err != nil {
				return err
			}
			findings = append(findings, pkgFindings...)
			return nil
		})
		if err != nil {
			return nil, err
		}
	}

	sort.Slice(findings, func(i, j int) bool { return findings[i].Path < findings[j].Path })
	return findings, nil
}

// defaultScanRoots lists repository-relative trees that own maintained Go code.
func defaultScanRoots() []string {
	return []string{
		"apps/lina-core",
		"apps/lina-plugins",
		"hack/tools",
	}
}

// shouldSkipDir reports directories that must not be walked for Go tests.
func shouldSkipDir(name string) bool {
	switch name {
	case ".git", "node_modules", "vendor", "dist", "temp", "testdata", "bin":
		return true
	default:
		return strings.HasPrefix(name, ".")
	}
}

// scanPackageDir evaluates every *_test.go in one directory.
func scanPackageDir(repoRoot, dir string) ([]Finding, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("gotestnames: read dir %s: %w", dir, err)
	}

	names := make(map[string]struct{}, len(entries))
	var tests []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasSuffix(name, ".go") {
			continue
		}
		names[name] = struct{}{}
		if strings.HasSuffix(name, "_test.go") {
			tests = append(tests, name)
		}
	}
	if len(tests) == 0 {
		return nil, nil
	}

	var findings []Finding
	for _, testName := range tests {
		stem := strings.TrimSuffix(testName, "_test.go")
		if _, ok := names[stem+".go"]; ok {
			continue
		}
		rel, err := filepath.Rel(repoRoot, filepath.Join(dir, testName))
		if err != nil {
			return nil, fmt.Errorf("gotestnames: rel path for %s: %w", testName, err)
		}
		findings = append(findings, Finding{
			Path:   filepath.ToSlash(rel),
			Reason: fmt.Sprintf("missing production sibling %s.go (require 1:1 X.go ↔ X_test.go)", stem),
		})
	}
	return findings, nil
}
