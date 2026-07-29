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

// TestReconcileProcessStatusFromRelease covers the re-registration failure mode
// where official-plugins monorepo plugins already have submitted/published
// releases and discovery reports zero new drafts.
func TestReconcileProcessStatusFromRelease(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		release *entity.PluginMarketplaceRelease
		want    marketv1.MarketplaceProcessStatus
		ok      bool
	}{
		{
			name:    "nil release cannot reconcile",
			release: nil,
			ok:      false,
		},
		{
			name: "mutable draft still needs verify path",
			release: &entity.PluginMarketplaceRelease{
				ReleaseStatus: marketv1.MarketplaceStatusDraft.String(),
				ReviewStatus:  marketv1.MarketplaceReviewStatusDraft.String(),
			},
			ok: false,
		},
		{
			name: "rejected draft still needs verify path",
			release: &entity.PluginMarketplaceRelease{
				ReleaseStatus: marketv1.MarketplaceStatusDraft.String(),
				ReviewStatus:  marketv1.MarketplaceReviewStatusRejected.String(),
			},
			ok: false,
		},
		{
			name: "submitted draft restores pending_review",
			release: &entity.PluginMarketplaceRelease{
				ReleaseStatus: marketv1.MarketplaceStatusDraft.String(),
				ReviewStatus:  marketv1.MarketplaceReviewStatusSubmitted.String(),
			},
			want: marketv1.MarketplaceProcessStatusPendingReview,
			ok:   true,
		},
		{
			name: "reviewing draft restores pending_review",
			release: &entity.PluginMarketplaceRelease{
				ReleaseStatus: marketv1.MarketplaceStatusDraft.String(),
				ReviewStatus:  marketv1.MarketplaceReviewStatusReviewing.String(),
			},
			want: marketv1.MarketplaceProcessStatusPendingReview,
			ok:   true,
		},
		{
			name: "published release restores completed",
			release: &entity.PluginMarketplaceRelease{
				ReleaseStatus: marketv1.MarketplaceStatusPublished.String(),
				ReviewStatus:  marketv1.MarketplaceReviewStatusApproved.String(),
			},
			want: marketv1.MarketplaceProcessStatusCompleted,
			ok:   true,
		},
		{
			name: "approved review restores completed even if status still draft",
			release: &entity.PluginMarketplaceRelease{
				ReleaseStatus: marketv1.MarketplaceStatusDraft.String(),
				ReviewStatus:  marketv1.MarketplaceReviewStatusApproved.String(),
			},
			want: marketv1.MarketplaceProcessStatusCompleted,
			ok:   true,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, ok := reconcileProcessStatusFromRelease(tc.release)
			if ok != tc.ok {
				t.Fatalf("ok=%v want %v (status=%s)", ok, tc.ok, got)
			}
			if ok && got != tc.want {
				t.Fatalf("status=%s want %s", got, tc.want)
			}
		})
	}
}
