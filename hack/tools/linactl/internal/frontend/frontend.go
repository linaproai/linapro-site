// Package frontend provides dependency checks for linactl commands that run the
// Vite frontend. It owns the node_modules sentinel logic so command files only
// decide when frontend dependencies are required.
package frontend

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"linactl/internal/toolutil"
)

// CommandRunner runs one child command with caller-provided working directory.
type CommandRunner func(context.Context, string, string, ...string) error

// EnsureDeps checks that the frontend node_modules are installed.
// If the vite binary is missing, it runs pnpm install automatically.
func EnsureDeps(ctx context.Context, root string, out io.Writer, run CommandRunner) error {
	vite := toolutil.ViteCommand(root)
	if _, err := os.Stat(vite); err != nil {
		fmt.Fprintln(out, "Frontend dependencies not installed; running pnpm install...")
		if err = run(ctx, filepath.Join(root, "apps", "lina-vben"), "pnpm", "install"); err != nil {
			return err
		}
	}
	return EnsurePluginFrontendDeps(ctx, root, out, run)
}

// EnsurePluginFrontendDeps installs source-plugin frontend package deps when a
// plugin owns a frontend/package.json and its local node_modules is missing.
func EnsurePluginFrontendDeps(ctx context.Context, root string, out io.Writer, run CommandRunner) error {
	dirs, err := pluginFrontendPackageDirs(root)
	if err != nil {
		return err
	}
	for _, dir := range dirs {
		nodeModulesDir := filepath.Join(dir, "node_modules")
		info, statErr := os.Stat(nodeModulesDir)
		if statErr == nil {
			if !info.IsDir() {
				return fmt.Errorf("plugin frontend node_modules path is not a directory: %s", nodeModulesDir)
			}
			continue
		}
		if !errors.Is(statErr, os.ErrNotExist) {
			return statErr
		}

		relativeDir, relErr := filepath.Rel(root, dir)
		if relErr != nil {
			relativeDir = dir
		}
		fmt.Fprintf(out, "Plugin frontend dependencies not installed; running pnpm install: %s\n", relativeDir)
		if err = run(ctx, dir, "pnpm", "install", "--config.auto-install-peers=false"); err != nil {
			return err
		}
	}
	return nil
}

func pluginFrontendPackageDirs(root string) ([]string, error) {
	pluginRoot := filepath.Join(root, "apps", "lina-plugins")
	entries, err := os.ReadDir(pluginRoot)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}

	dirs := make([]string, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		pluginDir := filepath.Join(pluginRoot, entry.Name())
		pluginType, typeErr := readPluginType(filepath.Join(pluginDir, "plugin.yaml"))
		if typeErr != nil {
			return nil, typeErr
		}
		if pluginType != "source" {
			continue
		}
		frontendDir := filepath.Join(pluginDir, "frontend")
		if !fileExists(filepath.Join(frontendDir, "package.json")) {
			continue
		}
		dirs = append(dirs, frontendDir)
	}
	sort.Strings(dirs)
	return dirs, nil
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func readPluginType(path string) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", nil
		}
		return "", err
	}

	for _, line := range strings.Split(string(content), "\n") {
		if strings.TrimSpace(line) == "" || strings.HasPrefix(strings.TrimSpace(line), "#") {
			continue
		}
		if strings.HasPrefix(line, " ") || strings.HasPrefix(line, "\t") {
			continue
		}
		key, value, ok := strings.Cut(line, ":")
		if !ok || strings.TrimSpace(key) != "type" {
			continue
		}
		value, _, _ = strings.Cut(value, "#")
		return strings.Trim(strings.TrimSpace(value), `"'`), nil
	}
	return "", nil
}
