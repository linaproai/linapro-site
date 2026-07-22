// This file contains normalization, validation, and projection helpers shared
// by marketplace service operations.

package marketplace

import (
	"strings"
	"time"

	"github.com/gogf/gf/v2/util/gconv"

	marketv1 "linapro-plugin-marketplace/backend/api/market/v1"
	"linapro-plugin-marketplace/backend/internal/model/entity"
)

const (
	defaultManifestSnapshot  = "{}"
	defaultObjectSummary     = "{}"
	defaultCollectionSummary = "[]"
)

// normalizeKey trims stable identity keys before storage and lookup.
func normalizeKey(value string) string {
	return strings.TrimSpace(value)
}

// normalizeVisibility returns public visibility when callers pass the zero value.
func normalizeVisibility(value marketv1.MarketplaceVisibility) marketv1.MarketplaceVisibility {
	if strings.TrimSpace(value.String()) == "" {
		return marketv1.MarketplaceVisibilityPublic
	}
	return value
}

// normalizePluginType returns source type when callers pass the zero value.
func normalizePluginType(value marketv1.MarketplacePluginType) marketv1.MarketplacePluginType {
	if strings.TrimSpace(value.String()) == "" {
		return marketv1.MarketplacePluginTypeSource
	}
	return value
}

// normalizeProcessStatus returns pending_verify when callers pass a blank status.
func normalizeProcessStatus(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return marketv1.MarketplaceProcessStatusPendingVerify.String()
	}
	return trimmed
}

// defaultJSONString returns fallback when value is blank after trimming.
func defaultJSONString(value string, fallback string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return fallback
	}
	return trimmed
}

// validReviewDecision reports whether status is a supported reviewer decision.
func validReviewDecision(status marketv1.MarketplaceReviewStatus) bool {
	return status == marketv1.MarketplaceReviewStatusApproved ||
		status == marketv1.MarketplaceReviewStatusRejected
}

// immutableRelease reports whether a release can no longer be changed by draft upload.
func immutableRelease(release *entity.PluginMarketplaceRelease) bool {
	if release == nil {
		return false
	}
	switch marketv1.MarketplaceStatus(release.ReleaseStatus) {
	case marketv1.MarketplaceStatusPublished, marketv1.MarketplaceStatusDelisted, marketv1.MarketplaceStatusDeprecated:
		return true
	}
	switch marketv1.MarketplaceReviewStatus(release.ReviewStatus) {
	case marketv1.MarketplaceReviewStatusSubmitted, marketv1.MarketplaceReviewStatusReviewing, marketv1.MarketplaceReviewStatusApproved:
		return true
	default:
		return false
	}
}

// canSubmitReview reports whether a release is mutable enough to enter review.
func canSubmitReview(release *entity.PluginMarketplaceRelease) bool {
	if release == nil || marketv1.MarketplaceStatus(release.ReleaseStatus) != marketv1.MarketplaceStatusDraft {
		return false
	}
	reviewStatus := marketv1.MarketplaceReviewStatus(release.ReviewStatus)
	return reviewStatus == marketv1.MarketplaceReviewStatusDraft ||
		reviewStatus == marketv1.MarketplaceReviewStatusRejected
}

// canReviewRelease reports whether a release can receive a reviewer decision.
func canReviewRelease(release *entity.PluginMarketplaceRelease) bool {
	if release == nil || marketv1.MarketplaceStatus(release.ReleaseStatus) != marketv1.MarketplaceStatusDraft {
		return false
	}
	reviewStatus := marketv1.MarketplaceReviewStatus(release.ReviewStatus)
	return reviewStatus == marketv1.MarketplaceReviewStatusSubmitted ||
		reviewStatus == marketv1.MarketplaceReviewStatusReviewing
}

// validPluginStatusUpdate reports whether a reviewer-driven status target is supported.
func validPluginStatusUpdate(status marketv1.MarketplaceStatus) bool {
	switch status {
	case marketv1.MarketplaceStatusPublished, marketv1.MarketplaceStatusDelisted, marketv1.MarketplaceStatusDeprecated:
		return true
	default:
		return false
	}
}

// publisherRecordFromEntity maps a generated publisher entity to a service record.
func publisherRecordFromEntity(row *entity.PluginMarketplacePublisher) *PublisherRecord {
	if row == nil {
		return nil
	}
	return &PublisherRecord{
		ID:           row.Id,
		PublisherKey: row.PublisherKey,
		Name:         row.Name,
		Summary:      row.Summary,
		OwnerUserID:  row.OwnerUserId,
		OwnerOrgID:   row.OwnerOrgId,
		Verified:     row.Verified,
		Status:       PublisherStatus(row.Status),
		Homepage:     row.Homepage,
		ContactEmail: row.ContactEmail,
	}
}

// pluginRecordFromEntity maps a generated plugin entity to a service record.
func pluginRecordFromEntity(row *entity.PluginMarketplacePlugin) *PluginRecord {
	if row == nil {
		return nil
	}
	sourceKind := marketv1.MarketplaceSourceKind(normalizeSourceKind(row.SourceKind))
	return &PluginRecord{
		ID:              row.Id,
		PublisherID:     row.PublisherId,
		PluginID:        row.PluginId,
		Name:            row.Name,
		Summary:         row.Summary,
		Description:     row.Description,
		PluginType:      marketv1.MarketplacePluginType(row.PluginType),
		MarketStatus:    marketv1.MarketplaceStatus(row.MarketStatus),
		ProcessStatus:   marketv1.MarketplaceProcessStatus(normalizeProcessStatus(row.ProcessStatus)),
		Visibility:      marketv1.MarketplaceVisibility(row.Visibility),
		LatestReleaseID: row.LatestReleaseId,
		LatestVersion:   row.LatestVersion,
		Icon:            row.Icon,
		Homepage:        row.Homepage,
		Repository:      row.Repository,
		License:         row.License,
		DownloadCount:   row.DownloadCount,
		SourceKind:      sourceKind,
		RepoURL:         row.RepoUrl,
		RepoProvider:    marketv1.MarketplaceRepoProvider(row.RepoProvider),
		RepoPath:        row.RepoPath,
		CredentialRef:   row.CredentialRef,
		LastSyncAt:      cloneTime(row.LastSyncAt),
		LastSyncStatus:  row.LastSyncStatus,
		LastSyncMessage: row.LastSyncMessage,
		PublishedAt:     cloneTime(row.PublishedAt),
		UpdatedAt:       cloneTime(row.UpdatedAt),
	}
}

// releaseRecordFromEntity maps a generated release entity to a service record.
func releaseRecordFromEntity(row *entity.PluginMarketplaceRelease) *ReleaseRecord {
	if row == nil {
		return nil
	}
	return &ReleaseRecord{
		ID:             row.Id,
		PluginRecordID: row.PluginRecordId,
		PublisherID:    row.PublisherId,
		PluginID:       row.PluginId,
		Version:        row.ReleaseVersion,
		SourceRef:      row.SourceRef,
		SourceCommit:   row.SourceCommit,
		PluginType:     marketv1.MarketplacePluginType(row.PluginType),
		ReleaseStatus:  marketv1.MarketplaceStatus(row.ReleaseStatus),
		ReviewStatus:   marketv1.MarketplaceReviewStatus(row.ReviewStatus),
		ProcessStatus:  marketv1.MarketplaceProcessStatus(normalizeProcessStatus(row.ProcessStatus)),
		Visibility:     marketv1.MarketplaceVisibility(row.Visibility),
		MinHostVersion: row.MinHostVersion,
		MaxHostVersion: row.MaxHostVersion,
		ReviewMessage:  row.ReviewMessage,
		SubmittedAt:    cloneTime(row.SubmittedAt),
		ReviewedAt:     cloneTime(row.ReviewedAt),
		PublishedAt:    cloneTime(row.PublishedAt),
		UpdatedAt:      cloneTime(row.UpdatedAt),
	}
}

// artifactRecordFromEntity maps a generated artifact entity to a service record.
func artifactRecordFromEntity(row *entity.PluginMarketplaceArtifact) *ArtifactRecord {
	if row == nil {
		return nil
	}
	return &ArtifactRecord{
		ID:             row.Id,
		ReleaseID:      row.ReleaseId,
		PluginID:       row.PluginId,
		Version:        row.ReleaseVersion,
		ArtifactType:   marketv1.MarketplaceArtifactType(row.ArtifactType),
		StorageKey:     row.StorageKey,
		FileName:       row.FileName,
		ContentType:    row.ContentType,
		SizeBytes:      row.SizeBytes,
		Sha256:         row.Sha256,
		ManifestSha256: row.ManifestSha256,
		WasmSha256:     row.WasmSha256,
		UpdatedAt:      cloneTime(row.UpdatedAt),
	}
}

// cloneTime avoids returning pointers owned by generated entity structs.
func cloneTime(value *time.Time) *time.Time {
	if value == nil {
		return nil
	}
	cloned := *value
	return &cloned
}

// intID converts an InsertAndGetId result to the generated entity ID type.
func intID(id int64) int {
	return gconv.Int(id)
}
