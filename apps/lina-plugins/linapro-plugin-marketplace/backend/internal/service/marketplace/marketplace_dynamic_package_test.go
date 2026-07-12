// This file verifies dynamic ZIP scanner behavior for valid runtime packages,
// forbidden source directories, and root-versus-embedded manifest mismatches.

package marketplace

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"

	"lina-core/pkg/bizerr"
	"lina-core/pkg/plugin/pluginbridge/contract"
	"lina-core/pkg/plugin/pluginbridge/protocol"
)

func TestUploadDynamicPackageRequiresOwnerBeforeArtifactStorage(t *testing.T) {
	store := &uploadOwnershipArtifactStore{}
	service := &serviceImpl{artifacts: store}

	_, err := service.UploadDynamicPackage(context.Background(), UploadDynamicPackageInput{})
	if !bizerr.Is(err, CodeMarketplaceInvalidInput) {
		t.Fatalf("expected invalid owner error, got %v", err)
	}
	if store.putCalls != 0 || store.putFileCalls != 0 {
		t.Fatalf("unexpected artifact writes before ownership validation: %#v", store)
	}
}

func TestScanDynamicPackageAcceptsRuntimePackage(t *testing.T) {
	packagePath := writeSourcePackageZip(t, "linapro-demo-dynamic/", map[string]string{
		"plugin.yaml":            dynamicPackageRootManifest("v0.1.0"),
		"plugin.wasm":            string(buildDynamicTestWasm(t, "v0.1.0")),
		"README.md":              "# Dynamic Demo\n",
		"manifest/docs/index.md": "# Runtime Docs\n",
	})

	scan, err := scanDynamicPackage(UploadDynamicPackageInput{
		PluginID:    "linapro-demo-dynamic",
		Version:     "v0.1.0",
		PackagePath: packagePath,
		FileName:    "linapro-demo-dynamic.zip",
	})
	if err != nil {
		t.Fatalf("scanDynamicPackage returned error: %v", err)
	}
	if scan.manifest.ID != "linapro-demo-dynamic" {
		t.Fatalf("unexpected plugin id: %s", scan.manifest.ID)
	}
	if scan.wasmSha256 == "" || scan.manifestSha256 == "" {
		t.Fatal("expected wasm and manifest checksums")
	}
	if !strings.Contains(scan.hostServiceSummary, protocol.HostServiceRuntime) {
		t.Fatalf("expected host service summary, got %s", scan.hostServiceSummary)
	}
	if !strings.Contains(scan.routeSummary, "/health") {
		t.Fatalf("expected route summary, got %s", scan.routeSummary)
	}
	if !strings.Contains(scan.sqlSummary, "install_sql") {
		t.Fatalf("expected SQL summary, got %s", scan.sqlSummary)
	}
	if !strings.Contains(scan.docsSummary, "manifest/docs/index.md") {
		t.Fatalf("expected docs summary, got %s", scan.docsSummary)
	}
}

func TestScanDynamicPackageRejectsSourceDevelopmentEntries(t *testing.T) {
	cases := []struct {
		entryName string
		content   string
	}{
		{entryName: "backend/plugin.go", content: "package backend\n"},
		{entryName: "frontend/pages/index.vue", content: "<template><div /></template>\n"},
		{entryName: "hack/config.yaml", content: "build: {}\n"},
		{entryName: "main.go", content: "package main\n"},
	}

	for _, tc := range cases {
		t.Run(tc.entryName, func(t *testing.T) {
			packagePath := writeSourcePackageZip(t, "", map[string]string{
				"plugin.yaml": dynamicPackageRootManifest("v0.1.0"),
				"plugin.wasm": string(buildDynamicTestWasm(t, "v0.1.0")),
				tc.entryName:  tc.content,
			})

			_, err := scanDynamicPackage(UploadDynamicPackageInput{
				PluginID:    "linapro-demo-dynamic",
				Version:     "v0.1.0",
				PackagePath: packagePath,
				FileName:    "linapro-demo-dynamic.zip",
			})
			if !bizerr.Is(err, CodeMarketplacePackageStructureInvalid) {
				t.Fatalf("expected package structure invalid error, got %v", err)
			}
		})
	}
}

func TestScanDynamicPackageRejectsEmbeddedManifestMismatch(t *testing.T) {
	packagePath := writeSourcePackageZip(t, "", map[string]string{
		"plugin.yaml": dynamicPackageRootManifest("v0.1.0"),
		"plugin.wasm": string(buildDynamicTestWasm(t, "v0.2.0")),
	})

	_, err := scanDynamicPackage(UploadDynamicPackageInput{
		PluginID:    "linapro-demo-dynamic",
		Version:     "v0.1.0",
		PackagePath: packagePath,
		FileName:    "linapro-demo-dynamic.zip",
	})
	if !bizerr.Is(err, CodeMarketplacePackageManifestMismatch) {
		t.Fatalf("expected package manifest mismatch error, got %v", err)
	}
}

func dynamicPackageRootManifest(version string) string {
	return `id: linapro-demo-dynamic
name: LinaPro Demo Dynamic
version: ` + version + `
type: dynamic
distribution: managed
scope_nature: tenant_aware
supports_multi_tenant: true
default_install_mode: tenant_scoped
dependencies:
  framework:
    version: ">=v0.1.0 <v1.0.0"
hostServices:
  - service: runtime
    methods:
      - info.now
`
}

func buildDynamicTestWasm(t *testing.T, version string) []byte {
	t.Helper()

	supportsMultiTenant := true
	manifest := &sourcePackageManifest{
		ID:                  "linapro-demo-dynamic",
		Name:                "LinaPro Demo Dynamic",
		Version:             version,
		Type:                "dynamic",
		Distribution:        "managed",
		ScopeNature:         "tenant_aware",
		SupportsMultiTenant: &supportsMultiTenant,
		DefaultInstallMode:  "tenant_scoped",
		Dependencies: &sourcePackageDependencySpec{
			Framework: &sourcePackageFrameworkDependency{Version: ">=v0.1.0 <v1.0.0"},
		},
	}
	runtimeMetadata := &protocol.RuntimeArtifactMetadata{
		RuntimeKind:           contract.RuntimeKindWasm,
		ABIVersion:            contract.SupportedABIVersion,
		I18NAssetCount:        1,
		APIDocI18NAssetCount:  1,
		SQLAssetCount:         1,
		ManifestResourceCount: 1,
		RouteCount:            1,
	}
	hostServices := []*protocol.HostServiceSpec{{
		Service: protocol.HostServiceRuntime,
		Methods: []string{protocol.HostServiceMethodRuntimeInfoNow},
	}}
	routes := []*contract.RouteContract{{
		Method:      "GET",
		Path:        "/health",
		Access:      contract.AccessLogin,
		Permission:  "linapro-demo-dynamic:health:view",
		RequestType: "HealthReq",
	}}
	bridgeSpec := &contract.BridgeSpec{
		ABIVersion:     contract.SupportedABIVersion,
		RuntimeKind:    contract.RuntimeKindWasm,
		RouteExecution: true,
		RequestCodec:   contract.CodecProtobuf,
		ResponseCodec:  contract.CodecProtobuf,
	}
	installSQL := []*dynamicPackageSQLAsset{{
		Key:     "001-demo-dynamic.sql",
		Content: "CREATE TABLE IF NOT EXISTS plugin_linapro_demo_dynamic_records (id INT);",
	}}
	runtimeI18N := []*dynamicPackageLocaleAsset{{Locale: "en-US", Content: `{"demo":"Demo"}`}}
	apiDocI18N := []*dynamicPackageLocaleAsset{{Locale: "zh-CN", Content: `{"demo":"演示"}`}}
	resources := []*dynamicPackageManifestResource{{
		Path:          "manifest/docs/en-US/index.md",
		ContentBase64: base64.StdEncoding.EncodeToString([]byte("# Embedded Docs\n")),
	}}

	wasm := []byte{0x00, 0x61, 0x73, 0x6d, 0x01, 0x00, 0x00, 0x00}
	wasm = appendDynamicTestSection(t, wasm, protocol.WasmSectionManifest, manifest)
	wasm = appendDynamicTestSection(t, wasm, protocol.WasmSectionRuntime, runtimeMetadata)
	wasm = appendDynamicTestSection(t, wasm, protocol.WasmSectionBackendHostServices, hostServices)
	wasm = appendDynamicTestSection(t, wasm, protocol.WasmSectionBackendRoutes, routes)
	wasm = appendDynamicTestSection(t, wasm, protocol.WasmSectionBackendBridge, bridgeSpec)
	wasm = appendDynamicTestSection(t, wasm, protocol.WasmSectionInstallSQL, installSQL)
	wasm = appendDynamicTestSection(t, wasm, protocol.WasmSectionI18NAssets, runtimeI18N)
	wasm = appendDynamicTestSection(t, wasm, protocol.WasmSectionAPIDocI18NAssets, apiDocI18N)
	wasm = appendDynamicTestSection(t, wasm, protocol.WasmSectionManifestResources, resources)
	return wasm
}

func appendDynamicTestSection(t *testing.T, wasm []byte, name string, value interface{}) []byte {
	t.Helper()

	payload, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("marshal wasm section %s: %v", name, err)
	}
	sectionPayload := append([]byte{}, encodeDynamicTestULEB128(uint32(len(name)))...)
	sectionPayload = append(sectionPayload, []byte(name)...)
	sectionPayload = append(sectionPayload, payload...)

	next := append([]byte{}, wasm...)
	next = append(next, 0x00)
	next = append(next, encodeDynamicTestULEB128(uint32(len(sectionPayload)))...)
	next = append(next, sectionPayload...)
	return next
}

func encodeDynamicTestULEB128(value uint32) []byte {
	result := make([]byte, 0, 5)
	for {
		current := byte(value & 0x7f)
		value >>= 7
		if value != 0 {
			current |= 0x80
		}
		result = append(result, current)
		if value == 0 {
			return result
		}
	}
}
