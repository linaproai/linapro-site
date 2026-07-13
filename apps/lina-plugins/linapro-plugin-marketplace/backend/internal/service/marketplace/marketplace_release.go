// This file implements marketplace release draft mutability and review state
// transitions.

package marketplace

import (
	"context"
	"time"

	"github.com/gogf/gf/v2/database/gdb"

	"lina-core/pkg/bizerr"
	marketv1 "linapro-plugin-marketplace/backend/api/market/v1"
	"linapro-plugin-marketplace/backend/internal/dao"
	"linapro-plugin-marketplace/backend/internal/model/do"
	"linapro-plugin-marketplace/backend/internal/model/entity"
)

// SaveReleaseDraft creates or replaces a mutable marketplace release draft.
func (s *serviceImpl) SaveReleaseDraft(ctx context.Context, in SaveReleaseDraftInput) (*ReleaseRecord, error) {
	publisher, err := s.requirePublisherOwnedByUser(ctx, in.PublisherKey, in.OwnerUserID)
	if err != nil {
		return nil, err
	}
	plugin, err := s.requireOwnedPlugin(ctx, publisher, in.PluginID)
	if err != nil {
		return nil, err
	}
	if normalizeKey(in.Version) == "" {
		return nil, bizerr.NewCode(CodeMarketplaceInvalidInput)
	}

	existing, err := s.getReleaseByPluginVersion(ctx, plugin.PluginId, in.Version)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		if immutableRelease(existing) {
			return nil, bizerr.NewCode(CodeMarketplaceReleaseImmutable)
		}
		if !in.ReplaceDraft {
			return nil, bizerr.NewCode(CodeMarketplaceReleaseDraftExists)
		}
		return s.replaceReleaseDraft(ctx, existing.Id, plugin, in)
	}

	id, err := dao.PluginMarketplaceRelease.Ctx(ctx).Data(s.releaseDraftData(plugin, in)).InsertAndGetId()
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeMarketplaceStorageFailed)
	}
	return s.getReleaseRecordByID(ctx, intID(id))
}

// SubmitReleaseReview moves one mutable release to submitted review state.
func (s *serviceImpl) SubmitReleaseReview(ctx context.Context, in SubmitReleaseReviewInput) (*ReleaseRecord, error) {
	plugin, err := s.requirePluginForPublisher(ctx, in.PublisherKey, in.PluginID, in.OwnerUserID)
	if err != nil {
		return nil, err
	}
	release, err := s.requireRelease(ctx, plugin.PluginId, in.Version)
	if err != nil {
		return nil, err
	}
	if !canSubmitReview(release) {
		return nil, bizerr.NewCode(CodeMarketplaceReviewStateInvalid)
	}

	now := time.Now()
	if _, err = dao.PluginMarketplaceRelease.Ctx(ctx).
		Where(do.PluginMarketplaceRelease{Id: release.Id}).
		Data(do.PluginMarketplaceRelease{
			ReviewStatus:  marketv1.MarketplaceReviewStatusSubmitted.String(),
			ReviewMessage: normalizeKey(in.Message),
			SubmittedAt:   &now,
		}).
		Update(); err != nil {
		return nil, bizerr.WrapCode(err, CodeMarketplaceStorageFailed)
	}
	return s.getReleaseRecordByID(ctx, release.Id)
}

// ReviewRelease applies an approval or rejection decision to one submitted release.
func (s *serviceImpl) ReviewRelease(ctx context.Context, in ReviewReleaseInput) (*ReleaseRecord, error) {
	if !validReviewDecision(in.ReviewStatus) {
		return nil, bizerr.NewCode(CodeMarketplaceReviewStateInvalid)
	}
	plugin, err := s.getPluginByID(ctx, in.PluginID)
	if err != nil {
		return nil, err
	}
	if plugin == nil {
		return nil, bizerr.NewCode(CodeMarketplacePluginNotFound)
	}
	release, err := s.requireRelease(ctx, plugin.PluginId, in.Version)
	if err != nil {
		return nil, err
	}
	if !canReviewRelease(release) {
		return nil, bizerr.NewCode(CodeMarketplaceReviewStateInvalid)
	}

	if in.ReviewStatus == marketv1.MarketplaceReviewStatusRejected {
		return s.rejectRelease(ctx, release, in.Message)
	}
	return s.approveRelease(ctx, plugin, release, in.Message)
}

// UpdatePluginStatus updates lifecycle state for a plugin and its latest release.
func (s *serviceImpl) UpdatePluginStatus(ctx context.Context, in UpdatePluginStatusInput) (*PluginRecord, error) {
	if !validPluginStatusUpdate(in.Status) {
		return nil, bizerr.NewCode(CodeMarketplaceStatusInvalid)
	}
	plugin, err := s.getPluginByID(ctx, in.PluginID)
	if err != nil {
		return nil, err
	}
	if plugin == nil {
		return nil, bizerr.NewCode(CodeMarketplacePluginNotFound)
	}
	if plugin.LatestReleaseId <= 0 {
		return nil, bizerr.NewCode(CodeMarketplaceStatusInvalid)
	}

	if err = dao.PluginMarketplacePlugin.Transaction(ctx, func(ctx context.Context, _ gdb.TX) error {
		if _, updateErr := dao.PluginMarketplacePlugin.Ctx(ctx).
			Where(do.PluginMarketplacePlugin{Id: plugin.Id}).
			Data(do.PluginMarketplacePlugin{MarketStatus: in.Status.String()}).
			Update(); updateErr != nil {
			return bizerr.WrapCode(updateErr, CodeMarketplaceStorageFailed)
		}
		if _, updateErr := dao.PluginMarketplaceRelease.Ctx(ctx).
			Where(do.PluginMarketplaceRelease{Id: plugin.LatestReleaseId}).
			Data(do.PluginMarketplaceRelease{
				ReleaseStatus: in.Status.String(),
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

// replaceReleaseDraft updates an existing mutable release draft with fresh scan summaries.
func (s *serviceImpl) replaceReleaseDraft(
	ctx context.Context,
	releaseID int,
	plugin *entity.PluginMarketplacePlugin,
	in SaveReleaseDraftInput,
) (*ReleaseRecord, error) {
	if err := dao.PluginMarketplaceRelease.Transaction(ctx, func(ctx context.Context, _ gdb.TX) error {
		if _, updateErr := dao.PluginMarketplaceRelease.Ctx(ctx).
			Where(do.PluginMarketplaceRelease{Id: releaseID}).
			Data(s.releaseDraftData(plugin, in)).
			Update(); updateErr != nil {
			return bizerr.WrapCode(updateErr, CodeMarketplaceStorageFailed)
		}

		// Generated DO structs omit nil pointer fields; explicit column nils
		// reset stale review timestamps when a rejected release becomes draft again.
		cols := dao.PluginMarketplaceRelease.Columns()
		if _, updateErr := dao.PluginMarketplaceRelease.Ctx(ctx).
			Where(do.PluginMarketplaceRelease{Id: releaseID}).
			Data(cols.SubmittedAt, nil, cols.ReviewedAt, nil, cols.PublishedAt, nil).
			Update(); updateErr != nil {
			return bizerr.WrapCode(updateErr, CodeMarketplaceStorageFailed)
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return s.getReleaseRecordByID(ctx, releaseID)
}

// rejectRelease records a reviewer rejection and keeps the release mutable.
func (s *serviceImpl) rejectRelease(
	ctx context.Context,
	release *entity.PluginMarketplaceRelease,
	message string,
) (*ReleaseRecord, error) {
	now := time.Now()
	if _, err := dao.PluginMarketplaceRelease.Ctx(ctx).
		Where(do.PluginMarketplaceRelease{Id: release.Id}).
		Data(do.PluginMarketplaceRelease{
			ReleaseStatus: marketv1.MarketplaceStatusDraft.String(),
			ReviewStatus:  marketv1.MarketplaceReviewStatusRejected.String(),
			ReviewMessage: normalizeKey(message),
			ReviewedAt:    &now,
		}).
		Update(); err != nil {
		return nil, bizerr.WrapCode(err, CodeMarketplaceStorageFailed)
	}
	return s.getReleaseRecordByID(ctx, release.Id)
}

// approveRelease publishes a submitted release and updates the plugin latest-version anchor.
func (s *serviceImpl) approveRelease(
	ctx context.Context,
	plugin *entity.PluginMarketplacePlugin,
	release *entity.PluginMarketplaceRelease,
	message string,
) (*ReleaseRecord, error) {
	now := time.Now()
	if err := dao.PluginMarketplaceRelease.Transaction(ctx, func(ctx context.Context, _ gdb.TX) error {
		if _, updateErr := dao.PluginMarketplaceRelease.Ctx(ctx).
			Where(do.PluginMarketplaceRelease{Id: release.Id}).
			Data(do.PluginMarketplaceRelease{
				ReleaseStatus: marketv1.MarketplaceStatusPublished.String(),
				ReviewStatus:  marketv1.MarketplaceReviewStatusApproved.String(),
				ReviewMessage: normalizeKey(message),
				ReviewedAt:    &now,
				PublishedAt:   &now,
			}).
			Update(); updateErr != nil {
			return bizerr.WrapCode(updateErr, CodeMarketplaceStorageFailed)
		}

		pluginData := do.PluginMarketplacePlugin{
			MarketStatus:    marketv1.MarketplaceStatusPublished.String(),
			LatestReleaseId: release.Id,
			LatestVersion:   release.ReleaseVersion,
		}
		if plugin.PublishedAt == nil {
			pluginData.PublishedAt = &now
		}
		if _, updateErr := dao.PluginMarketplacePlugin.Ctx(ctx).
			Where(do.PluginMarketplacePlugin{Id: plugin.Id}).
			Data(pluginData).
			Update(); updateErr != nil {
			return bizerr.WrapCode(updateErr, CodeMarketplaceStorageFailed)
		}
		return s.rebuildPluginReadModel(ctx, plugin.Id)
	}); err != nil {
		return nil, err
	}
	return s.getReleaseRecordByID(ctx, release.Id)
}

// releaseDraftData builds the DO payload for inserting or replacing a draft release.
func (s *serviceImpl) releaseDraftData(
	plugin *entity.PluginMarketplacePlugin,
	in SaveReleaseDraftInput,
) do.PluginMarketplaceRelease {
	pluginType := normalizePluginType(in.PluginType)
	return do.PluginMarketplaceRelease{
		PluginRecordId:     plugin.Id,
		PublisherId:        plugin.PublisherId,
		PluginId:           plugin.PluginId,
		ReleaseVersion:     normalizeKey(in.Version),
		SourceRef:          normalizeKey(in.SourceRef),
		PluginType:         pluginType.String(),
		ReleaseStatus:      marketv1.MarketplaceStatusDraft.String(),
		ReviewStatus:       marketv1.MarketplaceReviewStatusDraft.String(),
		Visibility:         normalizeVisibility(in.Visibility).String(),
		MinHostVersion:     normalizeKey(in.MinHostVersion),
		MaxHostVersion:     normalizeKey(in.MaxHostVersion),
		ManifestSnapshot:   defaultJSONString(in.ManifestSnapshot, defaultManifestSnapshot),
		DependencySummary:  defaultJSONString(in.DependencySummary, defaultCollectionSummary),
		HostServiceSummary: defaultJSONString(in.HostServiceSummary, defaultCollectionSummary),
		RouteSummary:       defaultJSONString(in.RouteSummary, defaultCollectionSummary),
		SqlSummary:         defaultJSONString(in.SQLSummary, defaultCollectionSummary),
		I18NSummary:        defaultJSONString(in.I18NSummary, defaultCollectionSummary),
		DocsSummary:        defaultJSONString(in.DocsSummary, defaultCollectionSummary),
		RiskSummary:        defaultJSONString(in.RiskSummary, defaultObjectSummary),
		ReviewMessage:      normalizeKey(in.ReviewMessage),
	}
}

// requireRelease loads a release or returns a not-found business error.
func (s *serviceImpl) requireRelease(
	ctx context.Context,
	pluginID string,
	version string,
) (*entity.PluginMarketplaceRelease, error) {
	release, err := s.getReleaseByPluginVersion(ctx, pluginID, version)
	if err != nil {
		return nil, err
	}
	if release == nil {
		return nil, bizerr.NewCode(CodeMarketplaceReleaseNotFound)
	}
	return release, nil
}

// getReleaseByPluginVersion loads one release by stable plugin ID and version.
func (s *serviceImpl) getReleaseByPluginVersion(
	ctx context.Context,
	pluginID string,
	version string,
) (*entity.PluginMarketplaceRelease, error) {
	normalizedPluginID := normalizeKey(pluginID)
	normalizedVersion := normalizeKey(version)
	if normalizedPluginID == "" || normalizedVersion == "" {
		return nil, nil
	}

	var release *entity.PluginMarketplaceRelease
	if err := dao.PluginMarketplaceRelease.Ctx(ctx).
		Where(do.PluginMarketplaceRelease{
			PluginId:       normalizedPluginID,
			ReleaseVersion: normalizedVersion,
		}).
		Scan(&release); err != nil {
		return nil, bizerr.WrapCode(err, CodeMarketplaceStorageFailed)
	}
	return release, nil
}

// getReleaseRecordByID loads one release by primary key and projects it to a service record.
func (s *serviceImpl) getReleaseRecordByID(ctx context.Context, id int) (*ReleaseRecord, error) {
	var release *entity.PluginMarketplaceRelease
	if err := dao.PluginMarketplaceRelease.Ctx(ctx).
		Where(do.PluginMarketplaceRelease{Id: id}).
		Scan(&release); err != nil {
		return nil, bizerr.WrapCode(err, CodeMarketplaceStorageFailed)
	}
	if release == nil {
		return nil, bizerr.NewCode(CodeMarketplaceReleaseNotFound)
	}
	return releaseRecordFromEntity(release), nil
}
