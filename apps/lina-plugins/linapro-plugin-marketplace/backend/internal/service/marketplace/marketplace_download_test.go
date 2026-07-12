// This file verifies download-session projections and normalization helpers
// without requiring a live database. Database-backed session creation, event
// insertion, and statistics refresh are covered by the later integration task.

package marketplace

import (
	"strings"
	"testing"
	"time"

	marketv1 "linapro-plugin-marketplace/backend/api/market/v1"
	"linapro-plugin-marketplace/backend/internal/model/entity"
)

func TestNormalizeDownloadSessionTTLBounds(t *testing.T) {
	if got := normalizeDownloadSessionTTL(0); got != defaultDownloadSessionTTL {
		t.Fatalf("expected default ttl, got %s", got)
	}
	if got := normalizeDownloadSessionTTL(maxDownloadSessionTTL + time.Minute); got != maxDownloadSessionTTL {
		t.Fatalf("expected max ttl, got %s", got)
	}
	if got := normalizeDownloadSessionTTL(time.Minute); got != time.Minute {
		t.Fatalf("expected caller ttl, got %s", got)
	}
}

func TestDownloadSessionItemFromEntityBuildsControlledURL(t *testing.T) {
	now := time.UnixMilli(1767247200000)
	expiresAt := now.Add(time.Minute)
	item := downloadSessionItemFromEntityAt(&entity.PluginMarketplaceDownloadSession{
		SessionId:         "mpdl_abc",
		PluginId:          "linapro-demo-source",
		ReleaseVersion:    "v0.1.0",
		ArtifactType:      marketv1.MarketplaceArtifactTypeSourceZip.String(),
		ArtifactSizeBytes: 1024,
		Sha256:            strings.Repeat("a", 64),
		Status:            marketv1.MarketplaceDownloadSessionStatusActive.String(),
		ExpiresAt:         &expiresAt,
		CreatedAt:         &now,
	}, now)

	if item.DownloadUrl != "/x/linapro-plugin-marketplace/api/v1/market/download-sessions/mpdl_abc/content" {
		t.Fatalf("unexpected download url: %s", item.DownloadUrl)
	}
	if item.ExpiresAt == nil || *item.ExpiresAt != expiresAt.UnixMilli() {
		t.Fatalf("unexpected expiresAt: %#v", item.ExpiresAt)
	}
	if item.Status != marketv1.MarketplaceDownloadSessionStatusActive {
		t.Fatalf("unexpected status: %s", item.Status)
	}
}

func TestDownloadSessionItemFromEntityOmitsURLWhenConsumed(t *testing.T) {
	now := time.UnixMilli(1767247200000)
	expiresAt := now.Add(time.Minute)
	item := downloadSessionItemFromEntityAt(&entity.PluginMarketplaceDownloadSession{
		SessionId:      "mpdl_abc",
		Status:         marketv1.MarketplaceDownloadSessionStatusConsumed.String(),
		ExpiresAt:      &expiresAt,
		ReleaseVersion: "v0.1.0",
	}, now)

	if item.DownloadUrl != "" {
		t.Fatalf("expected consumed session to omit download url, got %s", item.DownloadUrl)
	}
}

func TestNormalizeDownloadEventTypeDefaultsToStarted(t *testing.T) {
	if got := normalizeDownloadEventType(""); got != DownloadEventTypeStarted {
		t.Fatalf("expected blank event to default to started, got %s", got)
	}
	if got := normalizeDownloadEventType(DownloadEventTypeCompleted); got != DownloadEventTypeCompleted {
		t.Fatalf("expected completed event to be preserved, got %s", got)
	}
}

func TestDownloadSessionActiveAndNotExpired(t *testing.T) {
	now := time.UnixMilli(1767247200000)
	expiresAt := now.Add(time.Minute)
	if !downloadSessionActiveAndNotExpired(&entity.PluginMarketplaceDownloadSession{
		Status:    marketv1.MarketplaceDownloadSessionStatusActive.String(),
		ExpiresAt: &expiresAt,
	}, now) {
		t.Fatal("expected active future session to be usable")
	}
	if downloadSessionActiveAndNotExpired(&entity.PluginMarketplaceDownloadSession{
		Status:    marketv1.MarketplaceDownloadSessionStatusConsumed.String(),
		ExpiresAt: &expiresAt,
	}, now) {
		t.Fatal("expected consumed session to be unusable for event recording")
	}

	expiredAt := now.Add(-time.Second)
	if downloadSessionActiveAndNotExpired(&entity.PluginMarketplaceDownloadSession{
		Status:    marketv1.MarketplaceDownloadSessionStatusActive.String(),
		ExpiresAt: &expiredAt,
	}, now) {
		t.Fatal("expected expired active session to be unusable")
	}
}

func TestDownloadSessionAllowsMetadata(t *testing.T) {
	if !downloadSessionAllowsMetadata(marketv1.MarketplaceDownloadSessionStatusActive) {
		t.Fatal("expected active session metadata to be readable")
	}
	if !downloadSessionAllowsMetadata(marketv1.MarketplaceDownloadSessionStatusConsumed) {
		t.Fatal("expected consumed session metadata to be readable")
	}
	if downloadSessionAllowsMetadata(marketv1.MarketplaceDownloadSessionStatusRevoked) {
		t.Fatal("expected revoked session metadata to be unavailable")
	}
	if downloadSessionAllowsMetadata(marketv1.MarketplaceDownloadSessionStatus("unknown")) {
		t.Fatal("expected unknown session status to be unavailable")
	}
}

func TestDownloadAuthorizationSnapshotJSON(t *testing.T) {
	content, err := downloadAuthorizationSnapshotJSON(VisibilitySubject{
		UserID:   1001,
		TenantID: 2001,
	}, &ArtifactRecord{
		ArtifactType: marketv1.MarketplaceArtifactTypeDynamicZip,
	}, time.UnixMilli(1767247200000))
	if err != nil {
		t.Fatalf("downloadAuthorizationSnapshotJSON returned error: %v", err)
	}
	if !strings.Contains(content, `"userId":1001`) || !strings.Contains(content, `"permission":"download"`) {
		t.Fatalf("unexpected authorization snapshot: %s", content)
	}
}

func TestDownloadVisibilitySubjectAddsRequesterScope(t *testing.T) {
	subject := downloadVisibilitySubject(VisibilitySubject{}, 1001)
	if subject.UserID != 1001 {
		t.Fatalf("expected requester user scope to be added, got %#v", subject)
	}

	subject = downloadVisibilitySubject(VisibilitySubject{UserID: 2002}, 1001)
	if subject.UserID != 2002 {
		t.Fatalf("expected explicit user scope to be preserved, got %#v", subject)
	}
}

func TestDownloadStatisticsSnapshotDataUsesSameTotal(t *testing.T) {
	pluginData := downloadStatisticsPluginData(23)
	readModelData := downloadStatisticsReadModelData(23)
	if pluginData.DownloadCount != 23 {
		t.Fatalf("unexpected plugin download snapshot: %#v", pluginData.DownloadCount)
	}
	if readModelData.DownloadCount != 23 {
		t.Fatalf("unexpected read-model download snapshot: %#v", readModelData.DownloadCount)
	}
}

func TestValidDownloadArtifactType(t *testing.T) {
	if !validDownloadArtifactType(marketv1.MarketplaceArtifactTypePluginWasm) {
		t.Fatal("expected plugin_wasm to be a valid download artifact type")
	}
	if validDownloadArtifactType(marketv1.MarketplaceArtifactType("unknown")) {
		t.Fatal("expected unknown artifact type to be invalid")
	}
}
