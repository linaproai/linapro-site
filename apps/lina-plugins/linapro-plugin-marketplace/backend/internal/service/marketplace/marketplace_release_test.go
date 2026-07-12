// This file verifies release mutability helpers that guard marketplace version
// overwrite behavior before draft replacement reaches persistence.

package marketplace

import (
	"context"
	"testing"

	"lina-core/pkg/bizerr"
	marketv1 "linapro-plugin-marketplace/backend/api/market/v1"
	"linapro-plugin-marketplace/backend/internal/model/entity"
)

func TestReleaseWritesRequireOwnerUserID(t *testing.T) {
	service := &serviceImpl{}
	ctx := context.Background()

	if _, err := service.SaveReleaseDraft(ctx, SaveReleaseDraftInput{
		PublisherKey: "publisher-a",
		PluginID:     "plugin-a",
		Version:      "v1.0.0",
	}); !bizerr.Is(err, CodeMarketplaceInvalidInput) {
		t.Fatalf("expected release draft owner error, got %v", err)
	}
	if _, err := service.SubmitReleaseReview(ctx, SubmitReleaseReviewInput{
		PluginID: "plugin-a",
		Version:  "v1.0.0",
	}); !bizerr.Is(err, CodeMarketplaceInvalidInput) {
		t.Fatalf("expected review submission owner error, got %v", err)
	}
}

func TestImmutableReleaseBlocksPublishedAndReviewStates(t *testing.T) {
	cases := []struct {
		name    string
		release *entity.PluginMarketplaceRelease
		want    bool
	}{
		{
			name:    "nil release is mutable",
			release: nil,
			want:    false,
		},
		{
			name: "draft review stays mutable",
			release: &entity.PluginMarketplaceRelease{
				ReleaseStatus: marketv1.MarketplaceStatusDraft.String(),
				ReviewStatus:  marketv1.MarketplaceReviewStatusDraft.String(),
			},
			want: false,
		},
		{
			name: "rejected review stays mutable",
			release: &entity.PluginMarketplaceRelease{
				ReleaseStatus: marketv1.MarketplaceStatusDraft.String(),
				ReviewStatus:  marketv1.MarketplaceReviewStatusRejected.String(),
			},
			want: false,
		},
		{
			name: "submitted review is immutable",
			release: &entity.PluginMarketplaceRelease{
				ReleaseStatus: marketv1.MarketplaceStatusDraft.String(),
				ReviewStatus:  marketv1.MarketplaceReviewStatusSubmitted.String(),
			},
			want: true,
		},
		{
			name: "reviewing release is immutable",
			release: &entity.PluginMarketplaceRelease{
				ReleaseStatus: marketv1.MarketplaceStatusDraft.String(),
				ReviewStatus:  marketv1.MarketplaceReviewStatusReviewing.String(),
			},
			want: true,
		},
		{
			name: "approved release is immutable",
			release: &entity.PluginMarketplaceRelease{
				ReleaseStatus: marketv1.MarketplaceStatusDraft.String(),
				ReviewStatus:  marketv1.MarketplaceReviewStatusApproved.String(),
			},
			want: true,
		},
		{
			name: "published release is immutable",
			release: &entity.PluginMarketplaceRelease{
				ReleaseStatus: marketv1.MarketplaceStatusPublished.String(),
				ReviewStatus:  marketv1.MarketplaceReviewStatusApproved.String(),
			},
			want: true,
		},
		{
			name: "delisted release is immutable",
			release: &entity.PluginMarketplaceRelease{
				ReleaseStatus: marketv1.MarketplaceStatusDelisted.String(),
				ReviewStatus:  marketv1.MarketplaceReviewStatusApproved.String(),
			},
			want: true,
		},
		{
			name: "deprecated release is immutable",
			release: &entity.PluginMarketplaceRelease{
				ReleaseStatus: marketv1.MarketplaceStatusDeprecated.String(),
				ReviewStatus:  marketv1.MarketplaceReviewStatusApproved.String(),
			},
			want: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := immutableRelease(tc.release); got != tc.want {
				t.Fatalf("immutableRelease() = %v, want %v", got, tc.want)
			}
		})
	}
}
