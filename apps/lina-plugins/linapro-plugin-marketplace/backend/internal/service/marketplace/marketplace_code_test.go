// This file verifies marketplace business error metadata and plugin-owned
// runtime translation coverage. The checks keep stable error codes, derived
// message keys, and English fallback text aligned with manifest/i18n resources.

package marketplace

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"lina-core/pkg/bizerr"
)

func TestMarketplaceErrorCodesHaveFallbackAndTranslations(t *testing.T) {
	enUS := readRuntimeI18nBundle(t, "../../../../manifest/i18n/en-US/plugin.json")
	zhCN := readRuntimeI18nBundle(t, "../../../../manifest/i18n/zh-CN/plugin.json")

	for _, code := range marketplaceErrorCodesForTest() {
		meta, ok := bizerr.Metadata(code)
		if !ok {
			t.Fatal("expected marketplace error code metadata")
		}
		if !strings.HasPrefix(meta.ErrorCode, "PLUGIN_MARKETPLACE_") {
			t.Fatalf("unexpected marketplace error code: %s", meta.ErrorCode)
		}
		if !strings.HasPrefix(meta.MessageKey, "error.plugin.marketplace.") {
			t.Fatalf("unexpected marketplace message key for %s: %s", meta.ErrorCode, meta.MessageKey)
		}
		if meta.Fallback == "" {
			t.Fatalf("expected English fallback for %s", meta.ErrorCode)
		}
		if got := enUS[meta.MessageKey]; got != meta.Fallback {
			t.Fatalf("expected en-US fallback for %s to be %q, got %q", meta.MessageKey, meta.Fallback, got)
		}
		if got := zhCN[meta.MessageKey]; got == "" {
			t.Fatalf("expected zh-CN translation for %s", meta.MessageKey)
		}
	}
}

func readRuntimeI18nBundle(t *testing.T, path string) map[string]string {
	t.Helper()

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read runtime i18n bundle %s: %v", path, err)
	}
	var bundle map[string]string
	if err = json.Unmarshal(content, &bundle); err != nil {
		t.Fatalf("decode runtime i18n bundle %s: %v", path, err)
	}
	return bundle
}

func marketplaceErrorCodesForTest() []*bizerr.Code {
	return []*bizerr.Code{
		CodeMarketplaceInvalidInput,
		CodeMarketplacePublisherAlreadyExists,
		CodeMarketplacePublisherNotFound,
		CodeMarketplacePublisherUnavailable,
		CodeMarketplacePluginNotFound,
		CodeMarketplacePluginIDOwned,
		CodeMarketplaceReleaseNotFound,
		CodeMarketplaceReleaseImmutable,
		CodeMarketplaceReleaseDraftExists,
		CodeMarketplacePackageInvalid,
		CodeMarketplacePackageStructureInvalid,
		CodeMarketplacePackageManifestMismatch,
		CodeMarketplacePackageScanFailed,
		CodeMarketplaceDocumentInvalid,
		CodeMarketplaceDocumentNotFound,
		CodeMarketplaceDownloadArtifactNotFound,
		CodeMarketplaceDownloadSessionNotFound,
		CodeMarketplaceDownloadSessionExpired,
		CodeMarketplaceDownloadSessionUnavailable,
		CodeMarketplaceReviewStateInvalid,
		CodeMarketplaceStatusInvalid,
		CodeMarketplaceStorageFailed,
	}
}
