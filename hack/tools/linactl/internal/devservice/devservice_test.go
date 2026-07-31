package devservice

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestServicesInjectsWorkspaceAndDataRootEnv(t *testing.T) {
	root := t.TempDir()
	services := Services(root, 9120, 3000)
	if len(services) < 1 {
		t.Fatal("expected backend service")
	}
	backend := services[0]
	if backend.WorkDir != filepath.Join(root, "apps", "lina-core") {
		t.Fatalf("backend WorkDir = %q", backend.WorkDir)
	}
	var hasWorkspace, hasData bool
	for _, env := range backend.Env {
		if env == "LINAPRO_WORKSPACE_ROOT="+root {
			hasWorkspace = true
		}
		if env == "LINAPRO_DATA_ROOT="+filepath.Join(root, "temp") {
			hasData = true
		}
	}
	if !hasWorkspace || !hasData {
		t.Fatalf("missing workspace/data env in %v", backend.Env)
	}
	// Frontend must not require workspace env for this contract.
	if len(services) > 1 {
		for _, env := range services[1].Env {
			if strings.HasPrefix(env, "LINAPRO_WORKSPACE_ROOT=") {
				t.Fatalf("frontend should not require workspace root env, got %q", env)
			}
		}
	}
}
