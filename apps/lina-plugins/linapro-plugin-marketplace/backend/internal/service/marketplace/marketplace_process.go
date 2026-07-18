// This file implements the async marketplace process pipeline:
// pending_verify -> pending_review -> completed (published).
// Scheduled jobs advance plugins without blocking the add-plugin request path.
// Git metadata discovery runs inside the pending_verify stage.

package marketplace

import (
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

const (
	// maxProcessBatch limits one cron run so a large backlog cannot starve other jobs.
	maxProcessBatch = 50
)

// ProcessMarketplacePipeline advances plugins waiting in the async verify path.
// It returns how many plugins were successfully transitioned in this run.
func (s *serviceImpl) ProcessMarketplacePipeline(ctx context.Context) (int, error) {
	plugins, err := s.listPluginsForProcessPipeline(ctx, maxProcessBatch)
	if err != nil {
		return 0, err
	}
	advanced := 0
	for _, plugin := range plugins {
		if plugin == nil {
			continue
		}
		changed, processErr := s.processOnePlugin(ctx, plugin)
		if processErr != nil {
			_ = s.markPluginProcessFailed(ctx, plugin, processErr)
			continue
		}
		if changed {
			advanced++
		}
	}
	return advanced, nil
}

// processOnePlugin advances one plugin by a single pipeline step when possible.
func (s *serviceImpl) processOnePlugin(
	ctx context.Context,
	plugin *entity.PluginMarketplacePlugin,
) (bool, error) {
	status := marketv1.MarketplaceProcessStatus(normalizeProcessStatus(plugin.ProcessStatus))
	switch status {
	case marketv1.MarketplaceProcessStatusPendingVerify:
		return s.processPendingVerify(ctx, plugin)
	default:
		return false, nil
	}
}

// processPendingVerify discovers Git metadata when needed, validates that a
// processable draft exists, then auto-submits it into the human review queue
// (pending_review). Upload packages already have package bytes on the platform.
func (s *serviceImpl) processPendingVerify(
	ctx context.Context,
	plugin *entity.PluginMarketplacePlugin,
) (bool, error) {
	release, err := s.findProcessableDraftRelease(ctx, plugin)
	if err != nil {
		return false, err
	}
	if release == nil {
		// Git plugins need async discovery before a draft is available.
		if normalizeSourceKind(plugin.SourceKind) == gitSourceKind {
			if _, discoverErr := s.DiscoverGitMetadata(ctx, DiscoverGitMetadataInput{PluginID: plugin.PluginId}); discoverErr != nil {
				return false, discoverErr
			}
			release, err = s.findProcessableDraftRelease(ctx, plugin)
			if err != nil {
				return false, err
			}
		}
	}
	if release == nil {
		return false, bizerr.NewCode(
			CodeMarketplacePackageScanFailed,
			bizerr.P("diagnostic", "no verifiable draft release is ready for review"),
		)
	}
	if !releaseHasVerificationPayload(release) {
		return false, bizerr.NewCode(
			CodeMarketplacePackageScanFailed,
			bizerr.P("diagnostic", "draft release is missing verification summaries"),
		)
	}

	publisher, err := s.getPublisherEntityByID(ctx, plugin.PublisherId)
	if err != nil {
		return false, err
	}
	if publisher == nil || publisher.OwnerUserId <= 0 {
		return false, bizerr.NewCode(CodeMarketplaceInvalidInput)
	}

	if canSubmitReview(release) {
		if _, submitErr := s.SubmitReleaseReview(ctx, SubmitReleaseReviewInput{
			PublisherKey: publisher.PublisherKey,
			OwnerUserID:  publisher.OwnerUserId,
			PluginID:     plugin.PluginId,
			Version:      release.ReleaseVersion,
			Message:      "auto-submitted after async verification",
		}); submitErr != nil {
			return false, submitErr
		}
	}

	if err = s.setPluginProcessStatus(ctx, plugin.Id, marketv1.MarketplaceProcessStatusPendingReview, ""); err != nil {
		return false, err
	}
	if _, updateErr := dao.PluginMarketplaceRelease.Ctx(ctx).
		Where(do.PluginMarketplaceRelease{Id: release.Id}).
		Data(do.PluginMarketplaceRelease{
			ProcessStatus: marketv1.MarketplaceProcessStatusPendingReview.String(),
		}).
		Update(); updateErr != nil {
		return false, bizerr.WrapCode(updateErr, CodeMarketplaceStorageFailed)
	}
	return true, nil
}

// listPluginsForProcessPipeline loads a bounded set of plugins waiting for async work.
func (s *serviceImpl) listPluginsForProcessPipeline(
	ctx context.Context,
	limit int,
) ([]*entity.PluginMarketplacePlugin, error) {
	if limit <= 0 {
		limit = maxProcessBatch
	}
	cols := dao.PluginMarketplacePlugin.Columns()
	var rows []*entity.PluginMarketplacePlugin
	// Include published plugins when a newer draft version re-enters the pipeline.
	err := dao.PluginMarketplacePlugin.Ctx(ctx).
		WhereIn(cols.ProcessStatus, []string{
			marketv1.MarketplaceProcessStatusPendingVerify.String(),
		}).
		WhereIn(cols.MarketStatus, []string{
			marketv1.MarketplaceStatusDraft.String(),
			marketv1.MarketplaceStatusPublished.String(),
		}).
		OrderAsc(cols.UpdatedAt).
		OrderAsc(cols.Id).
		Limit(limit).
		Scan(&rows)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeMarketplaceStorageFailed)
	}
	return rows, nil
}

// findProcessableDraftRelease returns the newest mutable draft that can enter review.
func (s *serviceImpl) findProcessableDraftRelease(
	ctx context.Context,
	plugin *entity.PluginMarketplacePlugin,
) (*entity.PluginMarketplaceRelease, error) {
	if plugin == nil {
		return nil, nil
	}
	cols := dao.PluginMarketplaceRelease.Columns()
	var release *entity.PluginMarketplaceRelease
	err := dao.PluginMarketplaceRelease.Ctx(ctx).
		Where(do.PluginMarketplaceRelease{
			PluginId:      plugin.PluginId,
			ReleaseStatus: marketv1.MarketplaceStatusDraft.String(),
		}).
		WhereIn(cols.ReviewStatus, []string{
			marketv1.MarketplaceReviewStatusDraft.String(),
			marketv1.MarketplaceReviewStatusRejected.String(),
		}).
		OrderDesc(cols.UpdatedAt).
		OrderDesc(cols.Id).
		Limit(1).
		Scan(&release)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeMarketplaceStorageFailed)
	}
	return release, nil
}

// releaseHasVerificationPayload reports whether a draft carries scan/manifest data.
func releaseHasVerificationPayload(release *entity.PluginMarketplaceRelease) bool {
	if release == nil {
		return false
	}
	manifest := strings.TrimSpace(release.ManifestSnapshot)
	if manifest == "" || manifest == "{}" || manifest == "null" {
		return false
	}
	return true
}

// setPluginProcessStatus updates plugin process_status and optional diagnostic message.
func (s *serviceImpl) setPluginProcessStatus(
	ctx context.Context,
	pluginRecordID int,
	status marketv1.MarketplaceProcessStatus,
	message string,
) error {
	if pluginRecordID <= 0 {
		return bizerr.NewCode(CodeMarketplaceInvalidInput)
	}
	data := do.PluginMarketplacePlugin{
		ProcessStatus: status.String(),
	}
	// Reuse last_sync_message for pipeline diagnostics without adding a new column.
	if strings.TrimSpace(message) != "" {
		data.LastSyncMessage = strings.TrimSpace(message)
	}
	if _, err := dao.PluginMarketplacePlugin.Ctx(ctx).
		Where(do.PluginMarketplacePlugin{Id: pluginRecordID}).
		Data(data).
		Update(); err != nil {
		return bizerr.WrapCode(err, CodeMarketplaceStorageFailed)
	}
	return nil
}

// setLatestMutableReleaseProcessStatus updates process_status on the newest mutable draft.
func (s *serviceImpl) setLatestMutableReleaseProcessStatus(
	ctx context.Context,
	pluginID string,
	status marketv1.MarketplaceProcessStatus,
) error {
	release, err := s.findProcessableDraftRelease(ctx, &entity.PluginMarketplacePlugin{PluginId: pluginID})
	if err != nil || release == nil {
		return err
	}
	if _, updateErr := dao.PluginMarketplaceRelease.Ctx(ctx).
		Where(do.PluginMarketplaceRelease{Id: release.Id}).
		Data(do.PluginMarketplaceRelease{ProcessStatus: status.String()}).
		Update(); updateErr != nil {
		return bizerr.WrapCode(updateErr, CodeMarketplaceStorageFailed)
	}
	return nil
}

// markPluginProcessFailed records a failed pipeline step without removing the plugin
// from the owner's My Plugins list.
func (s *serviceImpl) markPluginProcessFailed(
	ctx context.Context,
	plugin *entity.PluginMarketplacePlugin,
	processErr error,
) error {
	if plugin == nil {
		return nil
	}
	message := "process pipeline failed"
	if processErr != nil {
		message = gitPublicErrorMessage(processErr)
		if strings.TrimSpace(message) == "" {
			message = processErr.Error()
		}
	}
	if len(message) > 1024 {
		message = message[:1024]
	}
	now := time.Now()
	if _, err := dao.PluginMarketplacePlugin.Ctx(ctx).
		Where(do.PluginMarketplacePlugin{Id: plugin.Id}).
		Data(do.PluginMarketplacePlugin{
			ProcessStatus:   marketv1.MarketplaceProcessStatusFailed.String(),
			LastSyncStatus:  gitSyncStatusFailed,
			LastSyncMessage: message,
			LastSyncAt:      &now,
		}).
		Update(); err != nil {
		return bizerr.WrapCode(err, CodeMarketplaceStorageFailed)
	}
	_ = s.setLatestMutableReleaseProcessStatus(ctx, plugin.PluginId, marketv1.MarketplaceProcessStatusFailed)
	return nil
}

// getPublisherEntityByID loads one publisher by primary key.
func (s *serviceImpl) getPublisherEntityByID(
	ctx context.Context,
	publisherID int,
) (*entity.PluginMarketplacePublisher, error) {
	if publisherID <= 0 {
		return nil, nil
	}
	var row *entity.PluginMarketplacePublisher
	if err := dao.PluginMarketplacePublisher.Ctx(ctx).
		Where(do.PluginMarketplacePublisher{Id: publisherID}).
		Scan(&row); err != nil {
		return nil, bizerr.WrapCode(err, CodeMarketplaceStorageFailed)
	}
	return row, nil
}

// ensureProcessStatusOnDraft is used when creating drafts so add-plugin results
// immediately show pending_verify in My Plugins for both Git and upload sources.
func ensureProcessStatusOnDraft(
	_ string,
) marketv1.MarketplaceProcessStatus {
	return marketv1.MarketplaceProcessStatusPendingVerify
}

// applyProcessStatusAfterAdd writes process_status for a newly added plugin/release pair.
func (s *serviceImpl) applyProcessStatusAfterAdd(
	ctx context.Context,
	plugin *entity.PluginMarketplacePlugin,
	release *entity.PluginMarketplaceRelease,
	status marketv1.MarketplaceProcessStatus,
) error {
	if plugin == nil {
		return bizerr.NewCode(CodeMarketplaceInvalidInput)
	}
	return dao.PluginMarketplacePlugin.Transaction(ctx, func(ctx context.Context, _ gdb.TX) error {
		if err := s.setPluginProcessStatus(ctx, plugin.Id, status, ""); err != nil {
			return err
		}
		if release != nil {
			if _, err := dao.PluginMarketplaceRelease.Ctx(ctx).
				Where(do.PluginMarketplaceRelease{Id: release.Id}).
				Data(do.PluginMarketplaceRelease{ProcessStatus: status.String()}).
				Update(); err != nil {
				return bizerr.WrapCode(err, CodeMarketplaceStorageFailed)
			}
		}
		return nil
	})
}
