// This file covers publisher-owned package add, publish-for-review, and delist
// lifecycle guards that do not require a live database.

package marketplace

import (
	"context"
	"testing"

	marketv1 "linapro-plugin-marketplace/backend/api/market/v1"
	"linapro-plugin-marketplace/backend/internal/model/entity"
)

func TestRequestPluginPublishRequiresOwner(t *testing.T) {
	t.Parallel()
	service := &serviceImpl{}
	_, err := service.RequestPluginPublish(context.Background(), RequestPluginPublishInput{
		PluginID: "demo",
	})
	if err == nil {
		t.Fatal("expected missing owner user to fail")
	}
}

func TestOwnerDelistRequiresOwner(t *testing.T) {
	t.Parallel()
	service := &serviceImpl{}
	_, err := service.OwnerDelistPlugin(context.Background(), OwnerDelistPluginInput{
		PluginID: "demo",
	})
	if err == nil {
		t.Fatal("expected missing owner user to fail")
	}
}

func TestAddPluginPackageRequiresOwner(t *testing.T) {
	t.Parallel()
	service := &serviceImpl{}
	_, err := service.AddPluginPackage(context.Background(), PackageAddInput{
		PackagePath: "/tmp/missing.zip",
		FileName:    "missing.zip",
	})
	if err == nil {
		t.Fatal("expected missing owner user to fail")
	}
}

func TestCanSubmitReviewAllowsDraftAndRejectedOnly(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name    string
		release *entity.PluginMarketplaceRelease
		want    bool
	}{
		{
			name: "draft draft",
			release: &entity.PluginMarketplaceRelease{
				ReleaseStatus: marketv1.MarketplaceStatusDraft.String(),
				ReviewStatus:  marketv1.MarketplaceReviewStatusDraft.String(),
			},
			want: true,
		},
		{
			name: "draft rejected",
			release: &entity.PluginMarketplaceRelease{
				ReleaseStatus: marketv1.MarketplaceStatusDraft.String(),
				ReviewStatus:  marketv1.MarketplaceReviewStatusRejected.String(),
			},
			want: true,
		},
		{
			name: "draft submitted",
			release: &entity.PluginMarketplaceRelease{
				ReleaseStatus: marketv1.MarketplaceStatusDraft.String(),
				ReviewStatus:  marketv1.MarketplaceReviewStatusSubmitted.String(),
			},
			want: false,
		},
		{
			name: "published",
			release: &entity.PluginMarketplaceRelease{
				ReleaseStatus: marketv1.MarketplaceStatusPublished.String(),
				ReviewStatus:  marketv1.MarketplaceReviewStatusApproved.String(),
			},
			want: false,
		},
		{
			name: "delisted",
			release: &entity.PluginMarketplaceRelease{
				ReleaseStatus: marketv1.MarketplaceStatusDelisted.String(),
				ReviewStatus:  marketv1.MarketplaceReviewStatusApproved.String(),
			},
			want: false,
		},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := canSubmitReview(tc.release); got != tc.want {
				t.Fatalf("canSubmitReview()=%v want %v", got, tc.want)
			}
		})
	}
}

func TestDetectPackagePluginTypeRequiresReadableArchive(t *testing.T) {
	t.Parallel()
	_, err := detectPackagePluginType("/tmp/does-not-exist-marketplace.zip", "does-not-exist-marketplace.zip")
	if err == nil {
		t.Fatal("expected missing package detection to fail")
	}
}
