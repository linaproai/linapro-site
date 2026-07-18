// This file implements publisher-owned marketplace lifecycle operations used by
// the My Plugins workbench: package-driven add, publish-for-review, and delist.
// Public catalog visibility remains gated by market status and review approval.

package marketplace

import (
	"archive/zip"
	"context"
	"strings"
	"time"

	"github.com/gogf/gf/v2/database/gdb"

	"lina-core/pkg/bizerr"
	marketv1 "linapro-plugin-marketplace/backend/api/market/v1"
	"linapro-plugin-marketplace/backend/internal/dao"
	"linapro-plugin-marketplace/backend/internal/model/do"
	"linapro-plugin-marketplace/backend/internal/model/entity"
)

// PackageAddResult carries the draft plugin and release created from one package.
type PackageAddResult struct {
	Plugin  *PluginRecord
	Release *ReleaseRecord
}

// PackageAddInput carries one My Plugins package add request.
type PackageAddInput struct {
	PublisherKey string
	OwnerUserID  int64
	PackagePath  string
	FileName     string
	ContentType  string
	ReplaceDraft bool
}

// RequestPluginPublishInput carries one owner publish-for-review request.
type RequestPluginPublishInput struct {
	OwnerUserID int64
	PluginID    string
	Version     string
	Message     string
}

// OwnerDelistPluginInput carries one owner delist request.
type OwnerDelistPluginInput struct {
	OwnerUserID int64
	PluginID    string
	Message     string
}

// AddPluginPackage unpacks one uploaded package, creates or updates the owned
// draft plugin identity from plugin.yaml, and stores a draft release without
// submitting marketplace review.
func (s *serviceImpl) AddPluginPackage(
	ctx context.Context,
	in PackageAddInput,
) (*PackageAddResult, error) {
	if in.OwnerUserID <= 0 {
		return nil, bizerr.NewCode(CodeMarketplaceInvalidInput)
	}
	pluginType, err := detectPackagePluginType(in.PackagePath, in.FileName)
	if err != nil {
		return nil, err
	}

	var result *PackageAddResult
	switch pluginType {
	case marketv1.MarketplacePluginTypeDynamic:
		uploadResult, uploadErr := s.UploadDynamicPackage(ctx, UploadDynamicPackageInput{
			PublisherKey: in.PublisherKey,
			OwnerUserID:  in.OwnerUserID,
			PackagePath:  in.PackagePath,
			FileName:     in.FileName,
			ContentType:  in.ContentType,
			Visibility:   marketv1.MarketplaceVisibilityPrivate,
			ReplaceDraft: in.ReplaceDraft,
			AutoCreate:   true,
		})
		if uploadErr != nil {
			return nil, uploadErr
		}
		plugin, getErr := s.getPluginRecordByPluginID(ctx, uploadResult.Release.PluginID)
		if getErr != nil {
			return nil, getErr
		}
		result = &PackageAddResult{Plugin: plugin, Release: uploadResult.Release}
	default:
		uploadResult, uploadErr := s.UploadSourcePackage(ctx, UploadSourcePackageInput{
			PublisherKey: in.PublisherKey,
			OwnerUserID:  in.OwnerUserID,
			PackagePath:  in.PackagePath,
			FileName:     in.FileName,
			ContentType:  in.ContentType,
			Visibility:   marketv1.MarketplaceVisibilityPrivate,
			ReplaceDraft: in.ReplaceDraft,
			AutoCreate:   true,
		})
		if uploadErr != nil {
			return nil, uploadErr
		}
		plugin, getErr := s.getPluginRecordByPluginID(ctx, uploadResult.Release.PluginID)
		if getErr != nil {
			return nil, getErr
		}
		result = &PackageAddResult{Plugin: plugin, Release: uploadResult.Release}
	}

	// Package bytes are already on the platform; enter pending_verify for async
	// validation finalization and automatic review submission.
	if result != nil && result.Plugin != nil {
		pluginEntity, getErr := s.getPluginByID(ctx, result.Plugin.PluginID)
		if getErr != nil {
			return nil, getErr
		}
		var releaseEntity *entity.PluginMarketplaceRelease
		if result.Release != nil {
			releaseEntity, getErr = s.getReleaseByPluginVersion(ctx, result.Release.PluginID, result.Release.Version)
			if getErr != nil {
				return nil, getErr
			}
		}
		if applyErr := s.applyProcessStatusAfterAdd(
			ctx,
			pluginEntity,
			releaseEntity,
			marketv1.MarketplaceProcessStatusPendingVerify,
		); applyErr != nil {
			return nil, applyErr
		}
		refreshed, getErr := s.getPluginRecordByPluginID(ctx, result.Plugin.PluginID)
		if getErr != nil {
			return nil, getErr
		}
		result.Plugin = refreshed
		if result.Release != nil {
			refreshedRelease, releaseErr := s.getReleaseRecordByID(ctx, result.Release.ID)
			if releaseErr != nil {
				return nil, releaseErr
			}
			result.Release = refreshedRelease
		}
	}
	return result, nil
}

// RequestPluginPublish submits an owned plugin release for marketplace review.
// Delisted plugins re-enter the review queue and stay invisible until approved.
func (s *serviceImpl) RequestPluginPublish(
	ctx context.Context,
	in RequestPluginPublishInput,
) (*ReleaseRecord, error) {
	plugin, err := s.requirePluginForPublisher(ctx, "", in.PluginID, in.OwnerUserID)
	if err != nil {
		return nil, err
	}
	release, err := s.findPublishCandidateRelease(ctx, plugin, in.Version)
	if err != nil {
		return nil, err
	}
	if release == nil {
		return nil, bizerr.NewCode(CodeMarketplaceReleaseNotFound)
	}

	if marketv1.MarketplaceStatus(release.ReleaseStatus) == marketv1.MarketplaceStatusDelisted ||
		marketv1.MarketplaceStatus(plugin.MarketStatus) == marketv1.MarketplaceStatusDelisted {
		return s.submitDelistedReleaseForReview(ctx, plugin, release, in.Message)
	}
	return s.SubmitReleaseReview(ctx, SubmitReleaseReviewInput{
		OwnerUserID: in.OwnerUserID,
		PluginID:    plugin.PluginId,
		Version:     release.ReleaseVersion,
		Message:     in.Message,
	})
}

// OwnerDelistPlugin withdraws one owned published plugin from the public catalog.
func (s *serviceImpl) OwnerDelistPlugin(
	ctx context.Context,
	in OwnerDelistPluginInput,
) (*PluginRecord, error) {
	plugin, err := s.requirePluginForPublisher(ctx, "", in.PluginID, in.OwnerUserID)
	if err != nil {
		return nil, err
	}
	if marketv1.MarketplaceStatus(plugin.MarketStatus) != marketv1.MarketplaceStatusPublished {
		return nil, bizerr.NewCode(CodeMarketplaceStatusInvalid)
	}
	if plugin.LatestReleaseId <= 0 {
		return nil, bizerr.NewCode(CodeMarketplaceStatusInvalid)
	}

	if err = dao.PluginMarketplacePlugin.Transaction(ctx, func(ctx context.Context, _ gdb.TX) error {
		if _, updateErr := dao.PluginMarketplacePlugin.Ctx(ctx).
			Where(do.PluginMarketplacePlugin{Id: plugin.Id}).
			Data(do.PluginMarketplacePlugin{
				MarketStatus: marketv1.MarketplaceStatusDelisted.String(),
				Visibility:   marketv1.MarketplaceVisibilityPrivate.String(),
			}).
			Update(); updateErr != nil {
			return bizerr.WrapCode(updateErr, CodeMarketplaceStorageFailed)
		}
		if _, updateErr := dao.PluginMarketplaceRelease.Ctx(ctx).
			Where(do.PluginMarketplaceRelease{Id: plugin.LatestReleaseId}).
			Data(do.PluginMarketplaceRelease{
				ReleaseStatus: marketv1.MarketplaceStatusDelisted.String(),
				ReviewMessage: normalizeKey(in.Message),
			}).
			Update(); updateErr != nil {
			return bizerr.WrapCode(updateErr, CodeMarketplaceStorageFailed)
		}
		return s.rebuildPluginReadModel(ctx, plugin.Id)
	}); err != nil {
		return nil, err
	}
	return s.getPluginRecordByID(ctx, plugin.Id)
}

// resolvePublisherKeyForPackageUpload resolves ownership for package upload and
// optionally creates a private draft plugin identity from plugin.yaml metadata.
func (s *serviceImpl) resolvePublisherKeyForPackageUpload(
	ctx context.Context,
	publisherKey string,
	ownerUserID int64,
	manifest *sourcePackageManifest,
	pluginType marketv1.MarketplacePluginType,
	visibility marketv1.MarketplaceVisibility,
	autoCreate bool,
) (string, error) {
	if manifest == nil || normalizeKey(manifest.ID) == "" {
		return "", bizerr.NewCode(CodeMarketplaceInvalidInput)
	}
	existing, err := s.getPluginByID(ctx, manifest.ID)
	if err != nil {
		return "", err
	}
	if existing != nil {
		return s.resolvePublisherKeyForPlugin(ctx, publisherKey, manifest.ID, ownerUserID)
	}
	if !autoCreate {
		return s.resolvePublisherKeyForPlugin(ctx, publisherKey, manifest.ID, ownerUserID)
	}

	key := normalizeKey(publisherKey)
	if key == "" {
		// Fall back to the operator's single bound publisher when only one exists.
		list, listErr := s.ListPublishers(ctx, ListPublishersInput{
			OwnerUserID: ownerUserID,
			PageNum:     1,
			PageSize:    2,
		})
		if listErr != nil {
			return "", listErr
		}
		if list == nil || len(list.List) != 1 || list.List[0] == nil {
			return "", bizerr.NewCode(CodeMarketplaceInvalidInput)
		}
		key = list.List[0].PublisherKey
	}

	summary := normalizeKey(manifest.Description)
	if summary == "" {
		summary = normalizeKey(manifest.Name)
	}
	if summary == "" {
		summary = manifest.ID
	}
	if len(summary) > 512 {
		summary = summary[:512]
	}
	// Added plugins stay private until an explicit publish action is approved.
	_ = visibility
	if _, err = s.SavePluginDraft(ctx, SavePluginDraftInput{
		PublisherKey: key,
		OwnerUserID:  ownerUserID,
		PluginID:     manifest.ID,
		Name:         firstNonEmpty(manifest.Name, manifest.ID),
		Summary:      summary,
		Description:  normalizeKey(manifest.Description),
		PluginType:   pluginType,
		Visibility:   marketv1.MarketplaceVisibilityPrivate,
		Homepage:     normalizeKey(manifest.Homepage),
		License:      normalizeKey(manifest.License),
	}); err != nil {
		return "", err
	}
	return key, nil
}

// detectPackagePluginType inspects an uploaded archive and returns source or dynamic.
func detectPackagePluginType(
	packagePath string,
	fileName string,
) (marketv1.MarketplacePluginType, error) {
	scanPath, cleanup, err := materializeZipPackagePath(packagePath, fileName)
	if err != nil {
		return "", err
	}
	defer cleanup()

	zipReader, err := zip.OpenReader(scanPath)
	if err != nil {
		return "", packageDiagnosticError(CodeMarketplacePackageInvalid, "package must be a valid ZIP or tar.gz container")
	}
	defer zipReader.Close()

	fileIndex, err := indexSourceZipFiles(zipReader.File)
	if err != nil {
		return "", err
	}
	rootPrefix, err := detectSourcePackageRoot(fileIndex)
	if err != nil {
		// Dynamic packages also have a root; fall through with empty prefix scan.
		rootPrefix = ""
	}
	if _, ok := fileIndex[rootPrefix+dynamicPackageWasmPath]; ok {
		return marketv1.MarketplacePluginTypeDynamic, nil
	}
	// Probe nested roots when root detection failed for dynamic packages.
	for pathName := range fileIndex {
		if strings.HasSuffix(pathName, "/"+dynamicPackageWasmPath) || pathName == dynamicPackageWasmPath {
			return marketv1.MarketplacePluginTypeDynamic, nil
		}
	}
	return marketv1.MarketplacePluginTypeSource, nil
}

// touchPluginLatestDraft anchors draft plugin list projections on the newest
// mutable release so My Plugins can show version and review state before publish.
func (s *serviceImpl) touchPluginLatestDraft(
	ctx context.Context,
	plugin *entity.PluginMarketplacePlugin,
	release *ReleaseRecord,
) error {
	if plugin == nil || release == nil {
		return nil
	}
	if marketv1.MarketplaceStatus(plugin.MarketStatus) != marketv1.MarketplaceStatusDraft {
		return nil
	}
	if _, err := dao.PluginMarketplacePlugin.Ctx(ctx).
		Where(do.PluginMarketplacePlugin{Id: plugin.Id}).
		Data(do.PluginMarketplacePlugin{
			LatestReleaseId: release.ID,
			LatestVersion:   release.Version,
		}).
		Update(); err != nil {
		return bizerr.WrapCode(err, CodeMarketplaceStorageFailed)
	}
	return nil
}

// findPublishCandidateRelease selects the release that should enter review.
func (s *serviceImpl) findPublishCandidateRelease(
	ctx context.Context,
	plugin *entity.PluginMarketplacePlugin,
	version string,
) (*entity.PluginMarketplaceRelease, error) {
	if plugin == nil {
		return nil, bizerr.NewCode(CodeMarketplacePluginNotFound)
	}
	if requested := normalizeKey(version); requested != "" {
		return s.requireRelease(ctx, plugin.PluginId, requested)
	}

	// Prefer the newest mutable draft/rejected release.
	var draft *entity.PluginMarketplaceRelease
	if err := dao.PluginMarketplaceRelease.Ctx(ctx).
		Where(do.PluginMarketplaceRelease{
			PluginId:      plugin.PluginId,
			ReleaseStatus: marketv1.MarketplaceStatusDraft.String(),
		}).
		WhereIn(dao.PluginMarketplaceRelease.Columns().ReviewStatus, []string{
			marketv1.MarketplaceReviewStatusDraft.String(),
			marketv1.MarketplaceReviewStatusRejected.String(),
		}).
		OrderDesc(dao.PluginMarketplaceRelease.Columns().UpdatedAt).
		OrderDesc(dao.PluginMarketplaceRelease.Columns().Id).
		Limit(1).
		Scan(&draft); err != nil {
		return nil, bizerr.WrapCode(err, CodeMarketplaceStorageFailed)
	}
	if draft != nil {
		return draft, nil
	}

	// Delisted plugins re-list through the latest delisted release.
	if marketv1.MarketplaceStatus(plugin.MarketStatus) == marketv1.MarketplaceStatusDelisted &&
		plugin.LatestReleaseId > 0 {
		return s.getReleaseEntityByID(ctx, plugin.LatestReleaseId)
	}
	return nil, nil
}

// submitDelistedReleaseForReview moves a delisted release back into review.
func (s *serviceImpl) submitDelistedReleaseForReview(
	ctx context.Context,
	plugin *entity.PluginMarketplacePlugin,
	release *entity.PluginMarketplaceRelease,
	message string,
) (*ReleaseRecord, error) {
	if plugin == nil || release == nil {
		return nil, bizerr.NewCode(CodeMarketplaceInvalidInput)
	}
	now := time.Now()
	if err := dao.PluginMarketplaceRelease.Transaction(ctx, func(ctx context.Context, _ gdb.TX) error {
		if _, updateErr := dao.PluginMarketplaceRelease.Ctx(ctx).
			Where(do.PluginMarketplaceRelease{Id: release.Id}).
			Data(do.PluginMarketplaceRelease{
				ReleaseStatus: marketv1.MarketplaceStatusDraft.String(),
				ReviewStatus:  marketv1.MarketplaceReviewStatusSubmitted.String(),
				ReviewMessage: normalizeKey(message),
				SubmittedAt:   &now,
			}).
			Update(); updateErr != nil {
			return bizerr.WrapCode(updateErr, CodeMarketplaceStorageFailed)
		}
		// Clear prior publish/review timestamps for the relist cycle.
		cols := dao.PluginMarketplaceRelease.Columns()
		if _, updateErr := dao.PluginMarketplaceRelease.Ctx(ctx).
			Where(do.PluginMarketplaceRelease{Id: release.Id}).
			Data(cols.ReviewedAt, nil, cols.PublishedAt, nil).
			Update(); updateErr != nil {
			return bizerr.WrapCode(updateErr, CodeMarketplaceStorageFailed)
		}
		if _, updateErr := dao.PluginMarketplacePlugin.Ctx(ctx).
			Where(do.PluginMarketplacePlugin{Id: plugin.Id}).
			Data(do.PluginMarketplacePlugin{
				MarketStatus: marketv1.MarketplaceStatusDraft.String(),
			}).
			Update(); updateErr != nil {
			return bizerr.WrapCode(updateErr, CodeMarketplaceStorageFailed)
		}
		return s.rebuildPluginReadModel(ctx, plugin.Id)
	}); err != nil {
		return nil, err
	}
	return s.getReleaseRecordByID(ctx, release.Id)
}

// getPluginRecordByPluginID loads a plugin record by stable plugin ID.
func (s *serviceImpl) getPluginRecordByPluginID(ctx context.Context, pluginID string) (*PluginRecord, error) {
	plugin, err := s.getPluginByID(ctx, pluginID)
	if err != nil {
		return nil, err
	}
	if plugin == nil {
		return nil, bizerr.NewCode(CodeMarketplacePluginNotFound)
	}
	return pluginRecordFromEntity(plugin), nil
}

// getReleaseEntityByID loads one release entity by primary key.
func (s *serviceImpl) getReleaseEntityByID(
	ctx context.Context,
	id int,
) (*entity.PluginMarketplaceRelease, error) {
	if id <= 0 {
		return nil, nil
	}
	var release *entity.PluginMarketplaceRelease
	if err := dao.PluginMarketplaceRelease.Ctx(ctx).
		Where(do.PluginMarketplaceRelease{Id: id}).
		Scan(&release); err != nil {
		return nil, bizerr.WrapCode(err, CodeMarketplaceStorageFailed)
	}
	return release, nil
}
