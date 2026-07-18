package marketplace

import (
	"testing"

	marketv1 "linapro-plugin-marketplace/backend/api/market/v1"
	"linapro-plugin-marketplace/backend/internal/model/entity"
)

func TestEnsureProcessStatusOnDraft(t *testing.T) {
	t.Parallel()

	if got := ensureProcessStatusOnDraft(gitSourceKind); got != marketv1.MarketplaceProcessStatusPendingVerify {
		t.Fatalf("git source should start at pending_verify, got %s", got)
	}
	if got := ensureProcessStatusOnDraft(uploadSourceKind); got != marketv1.MarketplaceProcessStatusPendingVerify {
		t.Fatalf("upload source should start at pending_verify, got %s", got)
	}
	if got := ensureProcessStatusOnDraft(""); got != marketv1.MarketplaceProcessStatusPendingVerify {
		t.Fatalf("blank source should default to pending_verify, got %s", got)
	}
}

func TestNormalizeProcessStatus(t *testing.T) {
	t.Parallel()

	if got := normalizeProcessStatus(""); got != marketv1.MarketplaceProcessStatusPendingVerify.String() {
		t.Fatalf("blank process status should default to pending_verify, got %s", got)
	}
	if got := normalizeProcessStatus(" pending_review "); got != "pending_review" {
		t.Fatalf("unexpected normalize result: %s", got)
	}
}

func TestReleaseHasVerificationPayload(t *testing.T) {
	t.Parallel()

	if releaseHasVerificationPayload(nil) {
		t.Fatal("nil release should not be verifiable")
	}
	if releaseHasVerificationPayload(&entity.PluginMarketplaceRelease{ManifestSnapshot: "{}"}) {
		t.Fatal("empty manifest object should not be verifiable")
	}
	if releaseHasVerificationPayload(&entity.PluginMarketplaceRelease{ManifestSnapshot: ""}) {
		t.Fatal("blank manifest should not be verifiable")
	}
	if !releaseHasVerificationPayload(&entity.PluginMarketplaceRelease{
		ManifestSnapshot: `{"id":"demo","version":"v0.1.0"}`,
	}) {
		t.Fatal("non-empty manifest should be verifiable")
	}
}
