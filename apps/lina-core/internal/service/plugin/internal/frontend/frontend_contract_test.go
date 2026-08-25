// This file covers runtime hosted-menu validation against embedded frontend assets.

package frontend_test

import (
	"context"
	"encoding/base64"
	pluginv1 "lina-core/api/plugin/v1"
	"path/filepath"
	"testing"

	"lina-core/internal/model/entity"
	"lina-core/internal/service/plugin/internal/catalog"
	"lina-core/internal/service/plugin/internal/testutil"
)

// TestValidateHostedMenuBindingsAcceptsHostedRuntimeModes verifies that iframe,
// new-window, and embedded mount menus can all bind valid hosted assets.
func TestValidateHostedMenuBindingsAcceptsHostedRuntimeModes(t *testing.T) {
	services := testutil.NewServices()
	service := services.Frontend

	resetBundleCache(t, service)

	pluginDir := testutil.CreateTestRuntimePluginDirWithFrontendAssets(
		t,
		"plugin-dev-dynamic-bindings",
		"Runtime Binding Plugin",
		"v0.3.0",
		[]*catalog.ArtifactFrontendAsset{
			{
				Path:          "frontend/pages/index.html",
				ContentBase64: base64.StdEncoding.EncodeToString([]byte("<html><body>hosted entry</body></html>")),
				ContentType:   "text/html; charset=utf-8",
			},
			{
				Path:          "frontend/pages/mount.js",
				ContentBase64: base64.StdEncoding.EncodeToString([]byte("export function mount() {}")),
				ContentType:   "application/javascript",
			},
		},
		nil,
		nil,
	)

	manifest := &catalog.Manifest{
		ID:           "plugin-dev-dynamic-bindings",
		Name:         "Runtime Binding Plugin",
		Version:      "v0.3.0",
		Type:         pluginv1.PluginTypeDynamic.String(),
		ManifestPath: filepath.Join(pluginDir, "plugin.yaml"),
		RootDir:      pluginDir,
	}
	if err := services.Catalog.ValidateManifest(manifest, manifest.ManifestPath); err != nil {
		t.Fatalf("expected dynamic manifest to be valid, got error: %v", err)
	}

	hostedBaseURL := service.BuildRuntimeFrontendPublicBaseURL(manifest.ID, manifest.Version)
	menus := []*entity.SysMenu{
		{
			MenuKey: "plugin:plugin-dev-dynamic-bindings:iframe-entry",
			Name:    "Hosted iframe entry",
			Path:    hostedBaseURL + "index.html",
			IsFrame: 0,
		},
		{
			MenuKey: "plugin:plugin-dev-dynamic-bindings:new-window-entry",
			Name:    "Hosted new window entry",
			Path:    hostedBaseURL + "index.html",
			IsFrame: 1,
		},
		{
			MenuKey: "plugin:plugin-dev-dynamic-bindings:embedded-entry",
			Name:    "Hosted embedded entry",
			Path:    hostedBaseURL + "mount.js",
			IsFrame: 0,
		},
	}

	if err := service.ValidateHostedMenuBindings(context.Background(), manifest, menus); err != nil {
		t.Fatalf("expected runtime hosted menu bindings to be valid, got error: %v", err)
	}
}

// TestValidateHostedMenuBindingsIgnoresWorkbenchEmbeddedQuery verifies leftover
// workbench query keys do not turn a hosted HTML entry into an embedded menu.
func TestValidateHostedMenuBindingsIgnoresWorkbenchEmbeddedQuery(t *testing.T) {
	services := testutil.NewServices()
	service := services.Frontend

	resetBundleCache(t, service)

	pluginDir := testutil.CreateTestRuntimePluginDirWithFrontendAssets(
		t,
		"plugin-dev-dynamic-broken-bindings",
		"Broken Runtime Binding Plugin",
		"v0.3.1",
		[]*catalog.ArtifactFrontendAsset{
			{
				Path:          "frontend/pages/index.html",
				ContentBase64: base64.StdEncoding.EncodeToString([]byte("<html><body>hosted entry</body></html>")),
				ContentType:   "text/html; charset=utf-8",
			},
		},
		nil,
		nil,
	)

	manifest := &catalog.Manifest{
		ID:           "plugin-dev-dynamic-broken-bindings",
		Name:         "Broken Runtime Binding Plugin",
		Version:      "v0.3.1",
		Type:         pluginv1.PluginTypeDynamic.String(),
		ManifestPath: filepath.Join(pluginDir, "plugin.yaml"),
		RootDir:      pluginDir,
	}
	if err := services.Catalog.ValidateManifest(manifest, manifest.ManifestPath); err != nil {
		t.Fatalf("expected dynamic manifest to be valid, got error: %v", err)
	}

	hostedBaseURL := service.BuildRuntimeFrontendPublicBaseURL(manifest.ID, manifest.Version)
	menus := []*entity.SysMenu{
		{
			MenuKey:    "plugin:plugin-dev-dynamic-broken-bindings:embedded-entry",
			Name:       "Broken embedded entry",
			Path:       hostedBaseURL + "index.html",
			QueryParam: `{"pluginAccessMode":"embedded-mount"}`,
			IsFrame:    0,
		},
	}

	if err := service.ValidateHostedMenuBindings(context.Background(), manifest, menus); err != nil {
		t.Fatalf("expected hosted HTML menu to stay valid without workbench query, got: %v", err)
	}
}
