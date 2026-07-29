// This file covers runtime package black-box behaviors, including wasm artifact
// contracts and dynamic-route dispatch paths.

package runtime_test

import (
	"context"
	"encoding/json"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	pluginv1 "lina-core/api/plugin/v1"
	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/service/datascope"
	"lina-core/internal/service/plugin/internal/catalog"
	"lina-core/internal/service/plugin/internal/lifecycle"
	"lina-core/internal/service/plugin/internal/plugintypes"
	"lina-core/internal/service/plugin/internal/runtime"
	"lina-core/internal/service/plugin/internal/testutil"
	"lina-core/pkg/plugin/pluginbridge/protocol"
	"lina-core/pkg/plugin/pluginhost"
	"lina-core/pkg/statusflag"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestBuildRuntimeWasmArtifactEmbedsBackendContracts verifies that hook and
// resource declarations are embedded into the generated runtime artifact.
func TestBuildRuntimeWasmArtifactEmbedsBackendContracts(t *testing.T) {
	services := testutil.NewServices()
	pluginDir := t.TempDir()

	testutil.WriteTestFile(
		t,
		filepath.Join(pluginDir, "plugin.yaml"),
		"id: plugin-dev-dynamic-contract\nname: Dynamic Contract\nversion: v0.2.0\ntype: dynamic\nscope_nature: tenant_aware\nsupports_multi_tenant: false\ndefault_install_mode: global\n",
	)
	testutil.WriteTestFile(
		t,
		filepath.Join(pluginDir, "hack", "config.yaml"),
		`wasm:
  hooks:
    - event: auth.login.succeeded
      action: sleep
      timeout: 50ms
      sleep: 10ms
  resources:
    - key: records
      type: table-list
      table: plugin_runtime_records
      fields:
        - name: id
          column: id
        - name: status
          column: status
      filters:
        - param: status
          column: status
          operator: eq
      orderBy:
        column: id
        direction: asc
      operations:
        - query
        - get
        - update
      keyField: id
      writableFields:
        - status
      access: both
      dataScope:
        userColumn: owner_user_id
`,
	)

	buildOut := testutil.BuildRuntimeArtifactWithHackTool(t, pluginDir)

	artifact, err := services.Runtime.ParseRuntimeWasmArtifactContent(buildOut.ArtifactPath, buildOut.Content)
	if err != nil {
		t.Fatalf("expected dynamic artifact parse to succeed, got error: %v", err)
	}
	if len(artifact.HookSpecs) != 1 {
		t.Fatalf("expected 1 embedded hook spec, got %d", len(artifact.HookSpecs))
	}
	if artifact.HookSpecs[0].Action != pluginhost.HookActionSleep {
		t.Fatalf("expected embedded hook action sleep, got %s", artifact.HookSpecs[0].Action)
	}
	if len(artifact.ResourceSpecs) != 1 {
		t.Fatalf("expected 1 embedded resource spec, got %d", len(artifact.ResourceSpecs))
	}
	if artifact.ResourceSpecs[0].DataScope == nil || artifact.ResourceSpecs[0].DataScope.UserColumn != "owner_user_id" {
		t.Fatalf("expected embedded resource data scope userColumn owner_user_id, got %#v", artifact.ResourceSpecs[0].DataScope)
	}
	if artifact.ResourceSpecs[0].KeyField != "id" || artifact.ResourceSpecs[0].Access != "both" {
		t.Fatalf("expected embedded resource governance fields, got %#v", artifact.ResourceSpecs[0])
	}
	if len(artifact.ResourceSpecs[0].WritableFields) != 1 || artifact.ResourceSpecs[0].WritableFields[0] != "status" {
		t.Fatalf("expected embedded writableFields to contain status, got %#v", artifact.ResourceSpecs[0].WritableFields)
	}
}

// TestRunBundledDynamicSampleBeforeInstallLifecycleAllowsRuntimeLog verifies
// the bundled dynamic sample can run its BeforeInstall callback, including the
// runtime.log.write host service used by the callback implementation.
func TestRunBundledDynamicSampleBeforeInstallLifecycleAllowsRuntimeLog(t *testing.T) {
	testutil.EnsureBundledRuntimeSampleArtifactForTests(t)

	services := testutil.NewServices()
	artifactPath := filepath.Join(testutil.TestDynamicStorageDir(), testutil.RuntimeArtifactFileName("linapro-demo-dynamic"))
	manifest, err := services.Catalog.LoadManifestFromArtifactPath(artifactPath)
	if err != nil {
		t.Fatalf("expected bundled dynamic manifest to load, got error: %v", err)
	}
	for _, contract := range manifest.LifecycleHandlers {
		if contract != nil && contract.Operation.String() == pluginhost.LifecycleHookBeforeInstall.String() {
			// CI runs this package with -race, so the real bundled sample gets a
			// wider test-only cold-start budget without changing production defaults.
			contract.TimeoutMs = int((2 * time.Minute) / time.Millisecond)
			break
		}
	}

	decision, err := services.Runtime.RunDynamicLifecyclePrecondition(context.Background(), manifest, runtime.DynamicLifecycleInput{
		PluginID:  manifest.ID,
		Operation: pluginhost.LifecycleHookBeforeInstall,
	})
	if err != nil {
		t.Fatalf("expected bundled BeforeInstall lifecycle to succeed, got error: %v decision=%#v", err, decision)
	}
	if decision == nil || !decision.OK {
		t.Fatalf("expected bundled BeforeInstall lifecycle to allow install, got %#v", decision)
	}
}

// TestLoadRuntimePluginManifestFromArtifactHydratesBackendContracts verifies
// that runtime manifest loading restores embedded backend contracts.
func TestLoadRuntimePluginManifestFromArtifactHydratesBackendContracts(t *testing.T) {
	services := testutil.NewServices()
	pluginDir := t.TempDir()

	testutil.WriteTestFile(
		t,
		filepath.Join(pluginDir, "plugin.yaml"),
		"id: plugin-dev-dynamic-active-contract\nname: Active Contract\nversion: v0.2.0\ntype: dynamic\nscope_nature: tenant_aware\nsupports_multi_tenant: false\ndefault_install_mode: global\n",
	)
	testutil.WriteTestFile(
		t,
		filepath.Join(pluginDir, "hack", "config.yaml"),
		`wasm:
  hooks:
    - event: auth.login.succeeded
      action: sleep
      timeout: 50ms
      sleep: 10ms
  resources:
    - key: records
      type: table-list
      table: plugin_runtime_records
      fields:
        - name: id
          column: id
        - name: status
          column: status
      orderBy:
        column: id
        direction: asc
      operations:
        - query
        - get
      keyField: id
      access: request
`,
	)

	buildOut := testutil.BuildRuntimeArtifactWithHackTool(t, pluginDir)
	if err := os.MkdirAll(filepath.Dir(buildOut.ArtifactPath), 0o755); err != nil {
		t.Fatalf("expected runtime artifact directory to be created, got error: %v", err)
	}
	if err := os.WriteFile(buildOut.ArtifactPath, buildOut.Content, 0o644); err != nil {
		t.Fatalf("expected runtime artifact to be written, got error: %v", err)
	}

	manifest, err := services.Catalog.LoadManifestFromArtifactPath(buildOut.ArtifactPath)
	if err != nil {
		t.Fatalf("expected runtime manifest load to succeed, got error: %v", err)
	}
	if len(manifest.Hooks) != 1 {
		t.Fatalf("expected runtime manifest to expose 1 hook, got %d", len(manifest.Hooks))
	}
	if len(manifest.BackendResources) != 1 {
		t.Fatalf("expected runtime manifest to expose 1 backend resource, got %d", len(manifest.BackendResources))
	}
	if _, ok := manifest.BackendResources["records"]; !ok {
		t.Fatalf("expected runtime manifest to expose resource key records, got %#v", manifest.BackendResources)
	}
	if manifest.BackendResources["records"].KeyField != "id" || len(manifest.BackendResources["records"].Operations) != 2 {
		t.Fatalf("expected runtime manifest to expose resource governance fields, got %#v", manifest.BackendResources["records"])
	}
}

// TestBundledDynamicSampleDeclaresBeforeAndAfterLifecycleCallbacks verifies
// the official dynamic sample registers the full canonical lifecycle callback
// set in its runtime artifact.
func TestBundledDynamicSampleDeclaresBeforeAndAfterLifecycleCallbacks(t *testing.T) {
	services := testutil.NewServices()
	repoRoot, err := testutil.FindRepoRoot(".")
	if err != nil {
		t.Fatalf("expected repo root to resolve, got error: %v", err)
	}
	pluginDir := filepath.Join(repoRoot, "apps", "lina-plugins", "linapro-demo-dynamic")
	if _, statErr := os.Stat(filepath.Join(pluginDir, "plugin.yaml")); statErr != nil {
		if os.IsNotExist(statErr) {
			t.Skip("official plugin workspace is not initialized")
		}
		t.Fatalf("expected linapro-demo-dynamic plugin.yaml to stat, got error: %v", statErr)
	}

	buildOut := testutil.BuildRuntimeArtifactWithHackTool(t, pluginDir)
	artifact, err := services.Runtime.ParseRuntimeWasmArtifactContent(buildOut.ArtifactPath, buildOut.Content)
	if err != nil {
		t.Fatalf("expected bundled dynamic sample artifact to parse, got error: %v", err)
	}

	expected := map[protocol.LifecycleOperation]struct{}{
		protocol.LifecycleOperationBeforeInstall:           {},
		protocol.LifecycleOperationAfterInstall:            {},
		protocol.LifecycleOperationBeforeUpgrade:           {},
		protocol.LifecycleOperationUpgrade:                 {},
		protocol.LifecycleOperationAfterUpgrade:            {},
		protocol.LifecycleOperationBeforeDisable:           {},
		protocol.LifecycleOperationAfterDisable:            {},
		protocol.LifecycleOperationBeforeUninstall:         {},
		protocol.LifecycleOperationUninstall:               {},
		protocol.LifecycleOperationAfterUninstall:          {},
		protocol.LifecycleOperationBeforeTenantDisable:     {},
		protocol.LifecycleOperationAfterTenantDisable:      {},
		protocol.LifecycleOperationBeforeTenantDelete:      {},
		protocol.LifecycleOperationAfterTenantDelete:       {},
		protocol.LifecycleOperationBeforeInstallModeChange: {},
		protocol.LifecycleOperationAfterInstallModeChange:  {},
	}
	if len(artifact.LifecycleContracts) != len(expected) {
		t.Fatalf("expected %d lifecycle contracts, got %d", len(expected), len(artifact.LifecycleContracts))
	}
	for _, contract := range artifact.LifecycleContracts {
		if contract == nil {
			t.Fatalf("expected lifecycle contract not to be nil")
		}
		if _, ok := expected[contract.Operation]; !ok {
			t.Fatalf("unexpected lifecycle operation %s", contract.Operation)
		}
		delete(expected, contract.Operation)
	}
	if len(expected) != 0 {
		t.Fatalf("expected all lifecycle operations to be declared, missing=%#v", expected)
	}
}

// TestMatchDynamicRoutePathSupportsParams verifies parameter placeholders are
// extracted from public route paths.
func TestMatchDynamicRoutePathSupportsParams(t *testing.T) {
	params, ok := runtime.MatchDynamicRoutePath("/records/{id}/detail", "/records/42/detail")
	if !ok {
		t.Fatal("expected dynamic path match to succeed")
	}
	if params["id"] != "42" {
		t.Fatalf("expected path param id=42, got %#v", params)
	}
}

// TestBuildMetadataMapsRouteGovernance verifies that matched route
// metadata is projected into a generic dynamic-route context.
func TestBuildMetadataMapsRouteGovernance(t *testing.T) {
	metadata := runtime.BuildMetadata(&runtime.DynamicRouteRuntimeState{
		Match: &runtime.DynamicRouteMatch{
			PluginID:   "linapro-demo-dynamic",
			PublicPath: "/x/linapro-demo-dynamic/api/v1/review",
			Route: &protocol.RouteContract{
				Method:  http.MethodGet,
				Tags:    []string{"plugin-review", "dynamic"},
				Summary: "Review summary",
				Meta: map[string]string{
					"x-route-purpose": "review",
				},
			},
		},
	})
	if metadata == nil {
		t.Fatal("expected dynamic route metadata to be built")
	}
	if metadata.PluginID != "linapro-demo-dynamic" {
		t.Fatalf("expected plugin id linapro-demo-dynamic, got %q", metadata.PluginID)
	}
	if metadata.Method != http.MethodGet {
		t.Fatalf("expected method GET, got %q", metadata.Method)
	}
	if metadata.PublicPath != "/x/linapro-demo-dynamic/api/v1/review" {
		t.Fatalf("expected public path to be preserved, got %q", metadata.PublicPath)
	}
	if len(metadata.Tags) != 2 || metadata.Tags[0] != "plugin-review" || metadata.Tags[1] != "dynamic" {
		t.Fatalf("expected route tags to be preserved, got %#v", metadata.Tags)
	}
	if metadata.Summary != "Review summary" {
		t.Fatalf("expected summary to be preserved, got %q", metadata.Summary)
	}
	if metadata.Meta["x-route-purpose"] != "review" {
		t.Fatalf("expected route metadata x-route-purpose review, got %#v", metadata.Meta)
	}
}

// TestDispatchDynamicRouteReturnsNotFoundWhenTenantPluginDisabled verifies
// tenant-scoped dynamic routes are hidden unless the current tenant enabled the
// plugin, even when the platform registry row is installed and enabled.
func TestDispatchDynamicRouteReturnsNotFoundWhenTenantPluginDisabled(t *testing.T) {
	var (
		services = testutil.NewServices()
		ctx      = datascope.WithTenantScope(context.Background(), 7001)
		pluginID = "plugin-dev-dynamic-route-tenant-disabled"
	)

	artifactPath := testutil.CreateTestRuntimeStorageArtifactWithFrontendAssetsAndBackendContracts(
		t,
		pluginID,
		"Tenant Disabled Route Plugin",
		"v1.0.0",
		testutil.DefaultTestRuntimeFrontendAssets(),
		nil,
		nil,
		[]*protocol.RouteContract{
			{
				Path:        "/api/v1/summary",
				Method:      http.MethodGet,
				Access:      protocol.AccessPublic,
				RequestType: "SummaryReq",
			},
		},
		&protocol.BridgeSpec{
			ABIVersion:     protocol.SupportedABIVersion,
			RuntimeKind:    protocol.RuntimeKindWasm,
			RouteExecution: true,
			RequestCodec:   protocol.CodecProtobuf,
			ResponseCodec:  protocol.CodecProtobuf,
			AllocExport:    "allocate",
			ExecuteExport:  "execute",
		},
	)
	testutil.CleanupPluginGovernanceRowsHard(t, context.Background(), pluginID)
	if _, err := dao.SysPluginState.Ctx(context.Background()).
		Where(do.SysPluginState{PluginId: pluginID}).
		Delete(); err != nil {
		t.Fatalf("cleanup dynamic route plugin state failed: %v", err)
	}
	t.Cleanup(func() {
		if _, err := dao.SysPluginState.Ctx(context.Background()).
			Where(do.SysPluginState{PluginId: pluginID}).
			Delete(); err != nil {
			t.Fatalf("cleanup dynamic route plugin state failed: %v", err)
		}
		testutil.CleanupPluginGovernanceRowsHard(t, context.Background(), pluginID)
	})

	manifest, err := services.Catalog.LoadManifestFromArtifactPath(artifactPath)
	if err != nil {
		t.Fatalf("load dynamic route manifest failed: %v", err)
	}
	manifest.ScopeNature = pluginv1.ScopeNatureTenantAware.String()
	manifest.DefaultInstallMode = pluginv1.InstallModeTenantScoped.String()
	if _, err = services.Store.SyncManifest(context.Background(), manifest); err != nil {
		t.Fatalf("sync dynamic route manifest failed: %v", err)
	}
	if err = services.Store.SetPluginInstalled(context.Background(), pluginID, statusflag.Installed.Int()); err != nil {
		t.Fatalf("set dynamic route plugin installed failed: %v", err)
	}
	if err = services.Store.SetPluginStatus(context.Background(), pluginID, statusflag.EnabledValue.Int()); err != nil {
		t.Fatalf("set dynamic route plugin enabled failed: %v", err)
	}
	if _, err = dao.SysPlugin.Ctx(context.Background()).
		Where(do.SysPlugin{PluginId: pluginID}).
		Data(do.SysPlugin{
			ScopeNature: pluginv1.ScopeNatureTenantAware.String(),
			InstallMode: pluginv1.InstallModeTenantScoped.String(),
		}).
		Update(); err != nil {
		t.Fatalf("set dynamic route plugin tenant governance failed: %v", err)
	}

	request := &ghttp.Request{}
	request.Request = httptest.NewRequest(http.MethodGet, pluginhost.PluginAPINamespacePrefix+"/"+pluginID+"/api/v1/summary", nil)
	response, err := services.Runtime.DispatchDynamicRoute(ctx, &runtime.DynamicRouteDispatchInput{Request: request})
	if err != nil {
		t.Fatalf("dispatch disabled tenant plugin route failed: %v", err)
	}
	if response == nil || response.StatusCode != http.StatusNotFound {
		t.Fatalf("expected disabled tenant plugin route to return 404, got %#v", response)
	}

	if err = services.Integration.SetTenantPluginEnabledState(ctx, pluginID, datascope.CurrentTenantID(ctx), true); err != nil {
		t.Fatalf("enable plugin for tenant failed: %v", err)
	}
	response, err = services.Runtime.DispatchDynamicRoute(ctx, &runtime.DynamicRouteDispatchInput{Request: request})
	if err == nil && response != nil && response.StatusCode == http.StatusNotFound {
		t.Fatalf("expected enabled tenant plugin route to pass routing, got %#v", response)
	}
	if err != nil && strings.Contains(err.Error(), "not found") {
		t.Fatalf("expected enabled tenant plugin route to pass routing, got error: %v", err)
	}
}

// TestDispatchDynamicRouteReturnsUpgradeRequiredWhenPendingUpgrade verifies
// dynamic business routes are blocked while a newer artifact awaits runtime upgrade.
func TestDispatchDynamicRouteReturnsUpgradeRequiredWhenPendingUpgrade(t *testing.T) {
	var (
		services   = testutil.NewServices()
		ctx        = context.Background()
		pluginID   = "plugin-dev-dynamic-route-pending-upgrade"
		oldVersion = "v0.1.0"
		newVersion = "v0.2.0"
	)

	artifactPath := testutil.CreateTestRuntimeStorageArtifactWithFrontendAssetsAndBackendContracts(
		t,
		pluginID,
		"Dynamic Route Pending Upgrade Plugin",
		oldVersion,
		testutil.DefaultTestRuntimeFrontendAssets(),
		nil,
		nil,
		[]*protocol.RouteContract{
			{
				Path:        "/api/v1/summary",
				Method:      http.MethodGet,
				Access:      protocol.AccessPublic,
				RequestType: "SummaryReq",
			},
		},
		&protocol.BridgeSpec{
			ABIVersion:     protocol.SupportedABIVersion,
			RuntimeKind:    protocol.RuntimeKindWasm,
			RouteExecution: true,
			RequestCodec:   protocol.CodecProtobuf,
			ResponseCodec:  protocol.CodecProtobuf,
		},
	)

	testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	t.Cleanup(func() {
		testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	})

	manifest, err := services.Catalog.LoadManifestFromArtifactPath(artifactPath)
	if err != nil {
		t.Fatalf("expected dynamic route manifest to load, got error: %v", err)
	}
	manifest.ScopeNature = pluginv1.ScopeNaturePlatformOnly.String()
	manifest.DefaultInstallMode = pluginv1.InstallModeGlobal.String()
	if _, err = services.Store.SyncManifest(ctx, manifest); err != nil {
		t.Fatalf("expected dynamic route manifest sync to succeed, got error: %v", err)
	}
	if err = services.Store.SetPluginInstalled(ctx, pluginID, statusflag.Installed.Int()); err != nil {
		t.Fatalf("expected dynamic route plugin install state to be set, got error: %v", err)
	}
	if err = services.Store.SetPluginStatus(ctx, pluginID, statusflag.EnabledValue.Int()); err != nil {
		t.Fatalf("expected dynamic route plugin enable state to be set, got error: %v", err)
	}

	testutil.CreateTestRuntimeStorageArtifactWithFrontendAssetsAndBackendContracts(
		t,
		pluginID,
		"Dynamic Route Pending Upgrade Plugin",
		newVersion,
		testutil.DefaultTestRuntimeFrontendAssets(),
		nil,
		nil,
		[]*protocol.RouteContract{
			{
				Path:        "/api/v1/summary",
				Method:      http.MethodGet,
				Access:      protocol.AccessPublic,
				RequestType: "SummaryReq",
			},
		},
		&protocol.BridgeSpec{
			ABIVersion:     protocol.SupportedABIVersion,
			RuntimeKind:    protocol.RuntimeKindWasm,
			RouteExecution: true,
			RequestCodec:   protocol.CodecProtobuf,
			ResponseCodec:  protocol.CodecProtobuf,
		},
	)
	newManifest, err := services.Catalog.LoadManifestFromArtifactPath(artifactPath)
	if err != nil {
		t.Fatalf("expected new dynamic route manifest to load, got error: %v", err)
	}
	if _, err = services.Store.SyncManifest(ctx, newManifest); err != nil {
		t.Fatalf("expected new dynamic route manifest sync to succeed, got error: %v", err)
	}

	request := &ghttp.Request{}
	request.Request = httptest.NewRequest(http.MethodGet, pluginhost.PluginAPINamespacePrefix+"/"+pluginID+"/api/v1/summary", nil)
	response, err := services.Runtime.DispatchDynamicRoute(ctx, &runtime.DynamicRouteDispatchInput{Request: request})
	if err != nil {
		t.Fatalf("expected pending-upgrade dynamic route to return bridge failure response, got error: %v", err)
	}
	if response == nil || response.StatusCode != http.StatusConflict {
		t.Fatalf("expected pending-upgrade dynamic route to return 409, got %#v", response)
	}
	if response.Failure == nil || response.Failure.Code != "PLUGIN_RUNTIME_UPGRADE_REQUIRED" {
		t.Fatalf("expected stable upgrade-required failure code, got %#v", response)
	}
}

// TestDispatchDynamicRouteBlocksFailedActiveRelease verifies rollback failure
// state uses the conservative exposure policy and does not serve a failed
// dynamic release even if the archived route manifest is still loadable.
func TestDispatchDynamicRouteBlocksFailedActiveRelease(t *testing.T) {
	var (
		services = testutil.NewServices()
		ctx      = context.Background()
		pluginID = "plugin-dev-dynamic-route-failed-active-release"
		version  = "v1.0.0"
	)

	artifactPath := testutil.CreateTestRuntimeStorageArtifactWithFrontendAssetsAndBackendContracts(
		t,
		pluginID,
		"Failed Active Release Route Plugin",
		version,
		testutil.DefaultTestRuntimeFrontendAssets(),
		nil,
		nil,
		[]*protocol.RouteContract{
			{
				Path:        "/api/v1/summary",
				Method:      http.MethodGet,
				Access:      protocol.AccessPublic,
				RequestType: "SummaryReq",
			},
		},
		&protocol.BridgeSpec{
			ABIVersion:     protocol.SupportedABIVersion,
			RuntimeKind:    protocol.RuntimeKindWasm,
			RouteExecution: true,
			RequestCodec:   protocol.CodecProtobuf,
			ResponseCodec:  protocol.CodecProtobuf,
		},
	)

	testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	t.Cleanup(func() {
		testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	})

	if _, err := services.Catalog.LoadManifestFromArtifactPath(artifactPath); err != nil {
		t.Fatalf("expected dynamic route manifest to load, got error: %v", err)
	}
	if _, err := services.Lifecycle.Install(ctx, pluginID, lifecycle.InstallOptions{}); err != nil {
		t.Fatalf("expected dynamic route plugin install to succeed, got error: %v", err)
	}
	if err := services.Store.SetPluginStatus(ctx, pluginID, statusflag.EnabledValue.Int()); err != nil {
		t.Fatalf("expected dynamic route plugin enable state to be set, got error: %v", err)
	}
	registry, err := services.Store.GetRegistry(ctx, pluginID)
	if err != nil {
		t.Fatalf("expected registry lookup to succeed, got error: %v", err)
	}
	if registry == nil || registry.ReleaseId <= 0 {
		t.Fatalf("expected installed plugin to point at active release, got %#v", registry)
	}
	if err = services.Store.UpdateReleaseState(ctx, registry.ReleaseId, plugintypes.ReleaseStatusFailed, ""); err != nil {
		t.Fatalf("expected release failure marker update to succeed, got error: %v", err)
	}

	request := &ghttp.Request{}
	request.Request = httptest.NewRequest(http.MethodGet, pluginhost.PluginAPINamespacePrefix+"/"+pluginID+"/api/v1/summary", nil)
	response, err := services.Runtime.DispatchDynamicRoute(ctx, &runtime.DynamicRouteDispatchInput{Request: request})
	if err != nil {
		t.Fatalf("expected failed-release dynamic route to return bridge failure response, got error: %v", err)
	}
	if response == nil || response.StatusCode != http.StatusConflict {
		t.Fatalf("expected failed-release dynamic route to return 409, got %#v", response)
	}
	if response.Failure == nil || response.Failure.Code != "PLUGIN_RUNTIME_UPGRADE_REQUIRED" {
		t.Fatalf("expected stable upgrade-required failure code, got %#v", response)
	}
}

// TestDispatchDynamicRouteAllowsPluginOwnedPathShapes verifies the runtime only
// forces the `/x/{pluginId}` prefix and preserves the following plugin-owned
// path content for contract matching and bridge metadata.
func TestDispatchDynamicRouteAllowsPluginOwnedPathShapes(t *testing.T) {
	var (
		services = testutil.NewServices()
		ctx      = context.Background()
		pluginID = "plugin-dev-dynamic-route-owned-paths"
	)

	artifactPath := testutil.CreateTestRuntimeStorageArtifactWithFrontendAssetsAndBackendContracts(
		t,
		pluginID,
		"Dynamic Route Owned Paths Plugin",
		"v1.0.0",
		testutil.DefaultTestRuntimeFrontendAssets(),
		nil,
		nil,
		[]*protocol.RouteContract{
			{
				Path:        "/api/v2/summary",
				Method:      http.MethodGet,
				Access:      protocol.AccessPublic,
				RequestType: "SummaryV2Req",
			},
			{
				Path:        "/interface/m1/summary",
				Method:      http.MethodGet,
				Access:      protocol.AccessPublic,
				RequestType: "InterfaceSummaryReq",
			},
			{
				Path:        "/graphql",
				Method:      http.MethodPost,
				Access:      protocol.AccessPublic,
				RequestType: "GraphQLReq",
			},
			{
				Path:        "/",
				Method:      http.MethodGet,
				Access:      protocol.AccessPublic,
				RequestType: "RootReq",
			},
		},
		&protocol.BridgeSpec{
			ABIVersion:     protocol.SupportedABIVersion,
			RuntimeKind:    protocol.RuntimeKindWasm,
			RouteExecution: false,
		},
	)
	testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	t.Cleanup(func() {
		testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	})

	manifest, err := services.Catalog.LoadManifestFromArtifactPath(artifactPath)
	if err != nil {
		t.Fatalf("load dynamic route manifest failed: %v", err)
	}
	manifest.ScopeNature = pluginv1.ScopeNaturePlatformOnly.String()
	manifest.DefaultInstallMode = pluginv1.InstallModeGlobal.String()
	if _, err = services.Store.SyncManifest(ctx, manifest); err != nil {
		t.Fatalf("sync dynamic route manifest failed: %v", err)
	}
	if err = services.Store.SetPluginInstalled(ctx, pluginID, statusflag.Installed.Int()); err != nil {
		t.Fatalf("set dynamic route plugin installed failed: %v", err)
	}
	if err = services.Store.SetPluginStatus(ctx, pluginID, statusflag.EnabledValue.Int()); err != nil {
		t.Fatalf("set dynamic route plugin enabled failed: %v", err)
	}
	services.Integration.SetPluginEnabledState(pluginID, true)

	tests := []struct {
		name       string
		method     string
		publicPath string
	}{
		{
			name:       "api v2",
			method:     http.MethodGet,
			publicPath: "/x/" + pluginID + "/api/v2/summary",
		},
		{
			name:       "interface",
			method:     http.MethodGet,
			publicPath: "/x/" + pluginID + "/interface/m1/summary",
		},
		{
			name:       "graphql",
			method:     http.MethodPost,
			publicPath: "/x/" + pluginID + "/graphql",
		},
		{
			name:       "root",
			method:     http.MethodGet,
			publicPath: "/x/" + pluginID,
		},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			request := &ghttp.Request{}
			request.Request = httptest.NewRequest(testCase.method, testCase.publicPath, nil)
			response, err := services.Runtime.DispatchDynamicRoute(
				ctx,
				&runtime.DynamicRouteDispatchInput{Request: request},
			)
			if err != nil {
				t.Fatalf("expected dynamic route dispatch to return bridge response, got error: %v", err)
			}
			if response == nil || response.StatusCode != http.StatusNotImplemented {
				t.Fatalf("expected matched placeholder route to return 501, got %#v", response)
			}
		})
	}
}

// TestExecuteDynamicWasmBridgeReturnsGuestResponse verifies that a bundled
// runtime plugin route executes and returns the guest response unchanged.
func TestExecuteDynamicWasmBridgeReturnsGuestResponse(t *testing.T) {
	testutil.EnsureBundledRuntimeSampleArtifactForTests(t)

	services := testutil.NewServices()
	manifest, err := loadBundledDynamicSampleManifest(t, services)
	if err != nil {
		t.Fatalf("expected bundled runtime artifact to load, got error: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), dynamicWasmBridgeTestTimeout)
	defer cancel()
	response, err := services.Runtime.ExecuteDynamicRoute(ctx, manifest, &protocol.BridgeRequestEnvelopeV1{
		PluginID: "linapro-demo-dynamic",
		Route: &protocol.RouteMatchSnapshotV1{
			InternalPath: "/api/v1/backend-summary",
			PublicPath:   "/x/linapro-demo-dynamic/api/v1/backend-summary",
			Access:       protocol.AccessLogin,
			Permission:   "linapro-demo-dynamic:backend:view",
			RequestType:  "BackendSummaryReq",
		},
		Identity: &protocol.IdentitySnapshotV1{
			UserID:       1,
			Username:     "admin",
			DataScope:    1,
			IsSuperAdmin: true,
		},
		Request: &protocol.HTTPRequestSnapshotV1{
			Method: http.MethodGet,
		},
	})
	if err != nil {
		t.Fatalf("expected dynamic wasm execution to succeed, got error: %v", err)
	}
	if response == nil || response.StatusCode != http.StatusOK {
		t.Fatalf("expected guest bridge response 200, got %#v", response)
	}
	if string(response.Body) == "" {
		t.Fatal("expected guest bridge response body to be non-empty")
	}
	if got := response.Headers["X-Lina-Plugin-Bridge"]; len(got) != 1 || got[0] != "linapro-demo-dynamic" {
		t.Fatalf("expected guest bridge header to be preserved, got %#v", response.Headers)
	}
	if got := response.Headers["X-Lina-Plugin-Middleware"]; len(got) != 1 || got[0] != "backend-summary" {
		t.Fatalf("expected guest-local middleware header to be preserved, got %#v", response.Headers)
	}

	payload := map[string]interface{}{}
	if err = json.Unmarshal(response.Body, &payload); err != nil {
		t.Fatalf("expected guest response body to be valid json, got error: %v", err)
	}
	if payload["pluginId"] != "linapro-demo-dynamic" {
		t.Fatalf("expected guest payload pluginId to be preserved, got %#v", payload)
	}
	if payload["authenticated"] != true {
		t.Fatalf("expected guest payload authenticated=true, got %#v", payload)
	}
}

// TestExecuteDynamicWasmBridgeHostCallDemoUsesStructuredHostServices verifies
// that structured host-service declarations are available inside guest code.
func TestExecuteDynamicWasmBridgeHostCallDemoUsesStructuredHostServices(t *testing.T) {
	testutil.EnsureBundledRuntimeSampleArtifactForTests(t)
	ensureDynamicDemoRecordTable(t)

	services := testutil.NewServices()
	manifest, err := loadBundledDynamicSampleManifest(t, services)
	if err != nil {
		t.Fatalf("expected bundled runtime artifact to load, got error: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), dynamicWasmBridgeTestTimeout)
	defer cancel()
	response, err := services.Runtime.ExecuteDynamicRoute(ctx, manifest, &protocol.BridgeRequestEnvelopeV1{
		PluginID:  "linapro-demo-dynamic",
		RequestID: "req-host-call-demo",
		Route: &protocol.RouteMatchSnapshotV1{
			InternalPath: "/api/v1/host-call-demo",
			PublicPath:   "/x/linapro-demo-dynamic/api/v1/host-call-demo",
			Access:       protocol.AccessLogin,
			Permission:   "linapro-demo-dynamic:backend:view",
			RequestType:  "HostCallDemoReq",
			QueryValues: map[string][]string{
				"skipNetwork": {"1"},
			},
		},
		Identity: &protocol.IdentitySnapshotV1{
			UserID:       1,
			Username:     "admin",
			DataScope:    1,
			IsSuperAdmin: true,
		},
		Request: &protocol.HTTPRequestSnapshotV1{
			Method: http.MethodGet,
		},
	})
	if err != nil {
		t.Fatalf("expected host call demo execution to succeed, got error: %v", err)
	}
	if response == nil || response.StatusCode != http.StatusOK {
		t.Fatalf("expected host call demo response 200, got %#v", response)
	}

	payload := map[string]interface{}{}
	if err = json.Unmarshal(response.Body, &payload); err != nil {
		t.Fatalf("expected host call demo body to be valid json, got error: %v", err)
	}
	if payload["pluginId"] != "linapro-demo-dynamic" {
		t.Fatalf("expected pluginId to be preserved, got %#v", payload)
	}
	if payload["visitCount"] == nil {
		t.Fatalf("expected visitCount to be returned, got %#v", payload)
	}

	runtimePayload, ok := payload["runtime"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected runtime payload object, got %#v full payload=%#v body=%s", payload["runtime"], payload, string(response.Body))
	}
	if runtimePayload["uuid"] == "" || runtimePayload["node"] == "" {
		t.Fatalf("expected runtime payload to include uuid and node, got %#v", runtimePayload)
	}

	storagePayload, ok := payload["storage"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected storage payload object, got %#v", payload["storage"])
	}
	if storagePayload["pathPrefix"] != "host-call-demo/" {
		t.Fatalf("expected storage pathPrefix host-call-demo/, got %#v", storagePayload)
	}
	if storagePayload["stored"] != true || storagePayload["deleted"] != true {
		t.Fatalf("expected storage payload to confirm store/delete lifecycle, got %#v", storagePayload)
	}

	dataPayload, ok := payload["data"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected data payload object, got %#v", payload["data"])
	}
	if dataPayload["table"] != "plugin_linapro_demo_dynamic_record" {
		t.Fatalf("expected data table plugin_linapro_demo_dynamic_record, got %#v", dataPayload)
	}
	if dataPayload["updated"] != true || dataPayload["deleted"] != true {
		t.Fatalf("expected data payload to confirm update/delete lifecycle, got %#v", dataPayload)
	}

	networkPayload, ok := payload["network"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected network payload object, got %#v", payload["network"])
	}
	if networkPayload["url"] != "https://example.com" {
		t.Fatalf("expected network url https://example.com, got %#v", networkPayload)
	}
	if networkPayload["skipped"] != true {
		t.Fatalf("expected network payload skipped=true during offline-safe test run, got %#v", networkPayload)
	}

	orgPayload, ok := payload["org"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected org payload object, got %#v", payload["org"])
	}
	if _, ok = orgPayload["available"].(bool); !ok {
		t.Fatalf("expected org payload to include availability, got %#v", orgPayload)
	}
	if orgPayload["assignmentCount"] == nil || orgPayload["currentUserDeptCount"] == nil {
		t.Fatalf("expected org payload to include current-user organization projections, got %#v", orgPayload)
	}

	tenantPayload, ok := payload["tenant"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected tenant payload object, got %#v", payload["tenant"])
	}
	if tenantPayload["visible"] != true {
		t.Fatalf("expected tenant visibility check to pass, got %#v", tenantPayload)
	}
	if tenantPayload["currentTenantId"] == nil || tenantPayload["userTenantCount"] == nil {
		t.Fatalf("expected tenant payload to include current tenant and user tenant count, got %#v", tenantPayload)
	}
}

// TestExecuteDynamicWasmBridgeCreatesDemoRecord verifies the bundled dynamic
// sample can execute the create demo-record route with data and storage host
// services, matching the E2E CRUD path.
func TestExecuteDynamicWasmBridgeCreatesDemoRecord(t *testing.T) {
	testutil.EnsureBundledRuntimeSampleArtifactForTests(t)
	ensureDynamicDemoRecordTable(t)

	services := testutil.NewServices()
	manifest, err := loadBundledDynamicSampleManifest(t, services)
	if err != nil {
		t.Fatalf("expected bundled runtime artifact to load, got error: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), dynamicWasmBridgeTestTimeout)
	defer cancel()
	response, err := services.Runtime.ExecuteDynamicRoute(ctx, manifest, &protocol.BridgeRequestEnvelopeV1{
		PluginID:  "linapro-demo-dynamic",
		RequestID: "req-demo-record-create",
		Route: &protocol.RouteMatchSnapshotV1{
			Method:       http.MethodPost,
			InternalPath: "/api/v1/demo-records",
			PublicPath:   "/x/linapro-demo-dynamic/api/v1/demo-records",
			Access:       protocol.AccessLogin,
			Permission:   "linapro-demo-dynamic:record:create",
			RequestType:  "CreateDemoRecordReq",
			RoutePath:    "/api/v1/demo-records",
		},
		Identity: &protocol.IdentitySnapshotV1{
			UserID:       1,
			Username:     "admin",
			DataScope:    1,
			IsSuperAdmin: true,
		},
		Request: &protocol.HTTPRequestSnapshotV1{
			Method:      http.MethodPost,
			ContentType: "application/json",
			Body: []byte(`{
				"title":"Dynamic route create test",
				"content":"Created through the bundled dynamic WASM bridge",
				"attachmentName":"linapro-demo-dynamic-note.txt",
				"attachmentContentBase64":"bGluYXByby1kZW1vLWR5bmFtaWMgYXR0YWNobWVudCBmaXh0dXJl",
				"attachmentContentType":"text/plain"
			}`),
		},
	})
	if err != nil {
		t.Fatalf("expected demo-record create route to succeed, got error: %v", err)
	}
	if response == nil || response.StatusCode != http.StatusOK {
		t.Fatalf("expected demo-record create response 200, got %#v body=%s", response, responseBodyForTest(response))
	}

	payload := map[string]interface{}{}
	if err = json.Unmarshal(response.Body, &payload); err != nil {
		t.Fatalf("expected demo-record create body to be valid json, got error: %v body=%s", err, string(response.Body))
	}
	if payload["title"] != "Dynamic route create test" || payload["hasAttachment"] != true {
		t.Fatalf("expected created demo-record payload with attachment, got %#v", payload)
	}
}

// loadBundledDynamicSampleManifest loads the bundled demo runtime artifact from test storage.
func loadBundledDynamicSampleManifest(t *testing.T, services *testutil.Services) (*catalog.Manifest, error) {
	t.Helper()

	artifactPath := filepath.Join(testutil.TestDynamicStorageDir(), testutil.RuntimeArtifactFileName("linapro-demo-dynamic"))
	return services.Catalog.LoadManifestFromArtifactPath(artifactPath)
}

// ensureDynamicDemoRecordTable provisions the dynamic sample table needed by
// bundled route tests that bypass the full plugin install lifecycle.
func ensureDynamicDemoRecordTable(t *testing.T) {
	t.Helper()
	ctx := context.Background()
	_, err := g.DB().Exec(ctx, `
CREATE TABLE IF NOT EXISTS plugin_linapro_demo_dynamic_record (
    "id"              VARCHAR(64) PRIMARY KEY,
    "tenant_id"       INT NOT NULL DEFAULT 0,
    "title"           VARCHAR(128) NOT NULL DEFAULT '',
    "content"         VARCHAR(1000) NOT NULL DEFAULT '',
    "attachment_name" VARCHAR(255) NOT NULL DEFAULT '',
    "attachment_path" VARCHAR(500) NOT NULL DEFAULT '',
    "created_at"      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at"      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
)`)
	if err != nil {
		t.Fatalf("failed to create dynamic demo record table: %v", err)
	}
	t.Cleanup(func() {
		if _, cleanupErr := g.DB().Exec(ctx, `DELETE FROM plugin_linapro_demo_dynamic_record WHERE "title" = ?`, "Dynamic route create test"); cleanupErr != nil {
			t.Fatalf("failed to cleanup dynamic demo record table: %v", cleanupErr)
		}
	})
}

// responseBodyForTest returns response body bytes without forcing every failure
// assertion to nil-check the response first.
func responseBodyForTest(response *protocol.BridgeResponseEnvelopeV1) []byte {
	if response == nil {
		return nil
	}
	return response.Body
}

const dynamicWasmBridgeTestTimeout = 2 * time.Minute
