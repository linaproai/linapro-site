// Package backend tests verify the marketplace source-plugin static
// registration and embedded asset binding remain discoverable by the host.
package backend

import (
	"io/fs"
	"testing"

	"lina-core/pkg/plugin/pluginhost"
	marketplace "linapro-plugin-marketplace"
)

// TestSourcePluginRegistration verifies the compile-time source-plugin
// declaration uses the same plugin ID as the manifest and exposes embedded
// plugin resources to the host scanner.
func TestSourcePluginRegistration(t *testing.T) {
	t.Parallel()

	definition, ok := pluginhost.GetSourcePlugin(marketplace.PluginID)
	if !ok {
		t.Fatalf("expected source plugin %s to be registered", marketplace.PluginID)
	}
	if definition.ID() != marketplace.PluginID {
		t.Fatalf("expected registered plugin ID %q, got %q", marketplace.PluginID, definition.ID())
	}
	if _, err := fs.Stat(definition.GetEmbeddedFiles(), "plugin.yaml"); err != nil {
		t.Fatalf("expected embedded plugin.yaml to be readable: %v", err)
	}
	if len(definition.GetRouteRegistrars()) == 0 {
		t.Fatal("expected marketplace HTTP route registrar to be registered")
	}
}

func TestRegisterRejectsNilDeclarations(t *testing.T) {
	t.Parallel()

	if err := Register(nil); err == nil {
		t.Fatal("expected nil declarations to be rejected")
	}
}
