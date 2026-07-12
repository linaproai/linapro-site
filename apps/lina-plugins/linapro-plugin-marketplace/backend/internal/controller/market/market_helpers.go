// This file contains controller helpers for request identity, visibility
// snapshots, and DTO projection used by marketplace HTTP handlers.

package market

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"

	"lina-core/pkg/bizerr"
	"lina-core/pkg/logger"
	marketv1 "linapro-plugin-marketplace/backend/api/market/v1"
	marketplacesvc "linapro-plugin-marketplace/backend/internal/service/marketplace"
)

// currentVisibilitySubject builds the marketplace visibility filter subject.
func (c *ControllerV1) currentVisibilitySubject(ctx context.Context) marketplacesvc.VisibilitySubject {
	current := c.currentContext(ctx)
	return marketplacesvc.VisibilitySubject{
		UserID:   int64(current.UserID),
		TenantID: int64(current.TenantID),
	}
}

// currentPublisherVisibilitySubject marks reads already protected by the publish permission.
func (c *ControllerV1) currentPublisherVisibilitySubject(ctx context.Context) marketplacesvc.VisibilitySubject {
	subject := c.currentVisibilitySubject(ctx)
	subject.CanPublish = true
	return subject
}

// currentReviewerVisibilitySubject marks reads already protected by the review permission.
func (c *ControllerV1) currentReviewerVisibilitySubject(ctx context.Context) marketplacesvc.VisibilitySubject {
	subject := c.currentVisibilitySubject(ctx)
	subject.CanReview = true
	return subject
}

// currentContext returns the plugin-visible business context snapshot.
func (c *ControllerV1) currentContext(ctx context.Context) bizctxSnapshot {
	if c != nil && c.bizCtx != nil {
		current := c.bizCtx.Current(ctx)
		return bizctxSnapshot{
			UserID:   current.UserID,
			TenantID: current.TenantID,
		}
	}
	return bizctxSnapshot{}
}

type bizctxSnapshot struct {
	UserID   int
	TenantID int
}

// requireUserID returns the authenticated user id or an invalid-input business error.
func (c *ControllerV1) requireUserID(ctx context.Context) (int64, error) {
	userID := int64(c.currentContext(ctx).UserID)
	if userID <= 0 {
		return 0, bizerr.NewCode(marketplacesvc.CodeMarketplaceInvalidInput)
	}
	return userID, nil
}

// saveUploadToTempFile writes the multipart file field to a temp path for package scanning.
func saveUploadToTempFile(ctx context.Context, fieldName string) (localPath string, fileName string, contentType string, cleanup func(), err error) {
	request := g.RequestFromCtx(ctx)
	if request == nil {
		return "", "", "", nil, bizerr.NewCode(marketplacesvc.CodeMarketplaceInvalidInput)
	}
	upload := request.GetUploadFile(fieldName)
	if upload == nil {
		return "", "", "", nil, bizerr.NewCode(marketplacesvc.CodeMarketplaceInvalidInput)
	}
	tempDir, err := os.MkdirTemp("", "marketplace-upload-*")
	if err != nil {
		return "", "", "", nil, bizerr.WrapCode(err, marketplacesvc.CodeMarketplaceStorageFailed)
	}
	cleanup = func() {
		if cleanupErr := os.RemoveAll(tempDir); cleanupErr != nil {
			logger.Error(ctx, "remove marketplace upload temp directory:", cleanupErr)
		}
	}
	savedName, err := upload.Save(tempDir, true)
	if err != nil {
		cleanup()
		return "", "", "", nil, bizerr.WrapCode(err, marketplacesvc.CodeMarketplaceStorageFailed)
	}
	localPath = filepath.Join(tempDir, savedName)
	fileName = strings.TrimSpace(upload.Filename)
	if fileName == "" {
		fileName = savedName
	}
	contentType = strings.TrimSpace(upload.Header.Get("Content-Type"))
	return localPath, fileName, contentType, cleanup, nil
}

// writeDownloadStream writes artifact bytes to the HTTP response as an attachment.
func writeDownloadStream(
	ctx context.Context,
	request *ghttp.Request,
	fileName string,
	sizeBytes int64,
	body io.ReadCloser,
) (err error) {
	if request == nil || body == nil {
		return bizerr.NewCode(marketplacesvc.CodeMarketplaceInvalidInput)
	}
	defer func() {
		closeErr := body.Close()
		if err == nil && closeErr != nil {
			err = bizerr.WrapCode(closeErr, marketplacesvc.CodeMarketplaceStorageFailed)
		}
	}()
	if strings.TrimSpace(fileName) == "" {
		fileName = "package.bin"
	}
	request.Response.Header().Set("Content-Type", "application/octet-stream")
	request.Response.Header().Set(
		"Content-Disposition",
		`attachment; filename="`+sanitizeDownloadFileName(fileName)+`"`,
	)
	if sizeBytes > 0 {
		request.Response.Header().Set("Content-Length", strconv.FormatInt(sizeBytes, 10))
	}
	if _, err = io.Copy(request.Response.RawWriter(), body); err != nil {
		return bizerr.WrapCode(err, marketplacesvc.CodeMarketplaceStorageFailed)
	}
	request.ExitAll()
	return nil
}

// sanitizeDownloadFileName keeps Content-Disposition filenames header-safe.
func sanitizeDownloadFileName(fileName string) string {
	replacer := strings.NewReplacer("\\", "_", "\"", "_", "\r", "_", "\n", "_")
	return replacer.Replace(fileName)
}

// publisherItemFromRecord maps a service publisher record to the public DTO.
func publisherItemFromRecord(record *marketplacesvc.PublisherRecord) *marketv1.MarketplacePublisherItem {
	if record == nil {
		return nil
	}
	return &marketv1.MarketplacePublisherItem{
		PublisherKey: record.PublisherKey,
		Name:         record.Name,
		Summary:      record.Summary,
		Verified:     record.Verified,
		Homepage:     record.Homepage,
	}
}

// pluginDetailFromRecord maps a draft plugin identity to a detail DTO shell.
func pluginDetailFromRecord(record *marketplacesvc.PluginRecord) *marketv1.MarketplacePluginDetailItem {
	if record == nil {
		return nil
	}
	return &marketv1.MarketplacePluginDetailItem{
		PluginId:      record.PluginID,
		Name:          record.Name,
		Summary:       record.Summary,
		Description:   record.Description,
		PluginType:    record.PluginType,
		MarketStatus:  record.MarketStatus,
		Visibility:    record.Visibility,
		LatestVersion: record.LatestVersion,
		Icon:          record.Icon,
		Homepage:      record.Homepage,
		Repository:    record.Repository,
		License:       record.License,
		DownloadCount: record.DownloadCount,
		PublishedAt:   unixMillisPtr(record.PublishedAt),
		UpdatedAt:     unixMillisPtr(record.UpdatedAt),
	}
}

// releaseItemFromRecords maps a draft release and primary artifact to the public DTO.
func releaseItemFromRecords(
	release *marketplacesvc.ReleaseRecord,
	artifact *marketplacesvc.ArtifactRecord,
) *marketv1.MarketplaceReleaseItem {
	if release == nil {
		return nil
	}
	var artifactItem *marketv1.MarketplaceArtifactItem
	if artifact != nil {
		artifactItem = &marketv1.MarketplaceArtifactItem{
			ArtifactType:   artifact.ArtifactType,
			FileName:       artifact.FileName,
			ContentType:    artifact.ContentType,
			SizeBytes:      artifact.SizeBytes,
			Sha256:         artifact.Sha256,
			ManifestSha256: artifact.ManifestSha256,
			WasmSha256:     artifact.WasmSha256,
		}
	}
	return &marketv1.MarketplaceReleaseItem{
		PluginId:       release.PluginID,
		Version:        release.Version,
		PluginType:     release.PluginType,
		ReleaseStatus:  release.ReleaseStatus,
		ReviewStatus:   release.ReviewStatus,
		Visibility:     release.Visibility,
		MinHostVersion: release.MinHostVersion,
		MaxHostVersion: release.MaxHostVersion,
		ReviewMessage:  release.ReviewMessage,
		Artifact:       artifactItem,
		SubmittedAt:    unixMillisPtr(release.SubmittedAt),
		ReviewedAt:     unixMillisPtr(release.ReviewedAt),
		PublishedAt:    unixMillisPtr(release.PublishedAt),
		UpdatedAt:      unixMillisPtr(release.UpdatedAt),
	}
}

// unixMillisPtr converts a time pointer to Unix milliseconds.
func unixMillisPtr(value *time.Time) *int64 {
	if value == nil {
		return nil
	}
	millis := value.UnixMilli()
	return &millis
}
