// This file implements publisher-owned and reviewer-managed marketplace list
// queries. Owned lists filter publishers by owner_user_id in the database;
// managed and review-queue lists batch-load related projections for the page.

package marketplace

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"

	"lina-core/pkg/bizerr"
	marketv1 "linapro-plugin-marketplace/backend/api/market/v1"
	"linapro-plugin-marketplace/backend/internal/dao"
	"linapro-plugin-marketplace/backend/internal/model/do"
	"linapro-plugin-marketplace/backend/internal/model/entity"
)

// ListOwnedPlugins returns plugins owned by publishers bound to one user.
func (s *serviceImpl) ListOwnedPlugins(ctx context.Context, in ListOwnedPluginsInput) (*PluginListOutput, error) {
	if in.OwnerUserID <= 0 {
		return nil, bizerr.NewCode(CodeMarketplaceInvalidInput)
	}
	return s.listPluginsFromIdentityTable(ctx, pluginIdentityListFilter{
		PageNum:     in.PageNum,
		PageSize:    in.PageSize,
		Keyword:     in.Keyword,
		PluginType:  in.PluginType,
		Status:      in.Status,
		OwnerUserID: in.OwnerUserID,
		Locale:      in.Locale,
	})
}

// ListManagedPlugins returns plugins across all publishers for reviewers.
func (s *serviceImpl) ListManagedPlugins(ctx context.Context, in ListManagedPluginsInput) (*PluginListOutput, error) {
	var publisherIDs []int
	if publisherKey := normalizeKey(in.Publisher); publisherKey != "" {
		publisher, err := s.getPublisherByKey(ctx, publisherKey)
		if err != nil {
			return nil, err
		}
		if publisher == nil {
			return &PluginListOutput{List: []*marketv1.MarketplacePluginListItem{}, Total: 0}, nil
		}
		publisherIDs = []int{publisher.Id}
	}
	return s.listPluginsFromIdentityTable(ctx, pluginIdentityListFilter{
		PageNum:            in.PageNum,
		PageSize:           in.PageSize,
		Keyword:            in.Keyword,
		PluginType:         in.PluginType,
		Status:             in.Status,
		PublisherIDs:       publisherIDs,
		MatchPublisherName: true,
		Locale:             in.Locale,
	})
}

// ListReviewQueue returns releases awaiting marketplace review.
func (s *serviceImpl) ListReviewQueue(ctx context.Context, in ListReviewQueueInput) (*ReviewQueueOutput, error) {
	pageNum, pageSize := normalizeMarketplacePage(in.PageNum, in.PageSize)
	cols := dao.PluginMarketplaceRelease.Columns()
	model := dao.PluginMarketplaceRelease.Ctx(ctx)

	if pluginID := normalizeKey(in.PluginID); pluginID != "" {
		model = model.Where(do.PluginMarketplaceRelease{PluginId: pluginID})
	}
	model = model.WhereIn(cols.ReviewStatus, reviewQueueStatuses(in.ReviewStatus))
	if keyword := normalizeKey(in.Keyword); keyword != "" {
		like := "%" + keyword + "%"
		model = model.Where(
			model.Builder().
				WhereLike(cols.PluginId, like).
				WhereOrLike(cols.ReleaseVersion, like),
		)
	}

	total, err := model.Clone().Count()
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeMarketplaceStorageFailed)
	}
	var rows []*entity.PluginMarketplaceRelease
	err = model.Clone().
		Fields(
			cols.Id,
			cols.PluginRecordId,
			cols.PublisherId,
			cols.PluginId,
			cols.ReleaseVersion,
			cols.PluginType,
			cols.ReleaseStatus,
			cols.ReviewStatus,
			cols.Visibility,
			cols.ReviewMessage,
			cols.SubmittedAt,
			cols.UpdatedAt,
		).
		OrderDesc(cols.SubmittedAt).
		OrderDesc(cols.UpdatedAt).
		OrderDesc(cols.Id).
		Page(pageNum, pageSize).
		Scan(&rows)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeMarketplaceStorageFailed)
	}

	pluginNames, err := s.batchPluginNamesByRecordID(ctx, pluginRecordIDsFromReleases(rows))
	if err != nil {
		return nil, err
	}
	displayByRelease, err := s.batchDisplayI18nByReleaseIDs(ctx, releaseIDsFromEntities(rows))
	if err != nil {
		return nil, err
	}
	publishers, err := s.batchPublishersByID(ctx, publisherIDsFromReleases(rows))
	if err != nil {
		return nil, err
	}
	artifacts, err := s.batchPrimaryArtifactsByRelease(ctx, rows)
	if err != nil {
		return nil, err
	}

	items := make([]*marketv1.MarketplaceReviewQueueItem, 0, len(rows))
	for _, row := range rows {
		if row == nil {
			continue
		}
		items = append(items, &marketv1.MarketplaceReviewQueueItem{
			PluginId:      row.PluginId,
			PluginName:    reviewQueuePluginName(row, pluginNames, displayByRelease, in.Locale),
			Version:       row.ReleaseVersion,
			PluginType:    marketv1.MarketplacePluginType(row.PluginType),
			ReleaseStatus: marketv1.MarketplaceStatus(row.ReleaseStatus),
			ReviewStatus:  marketv1.MarketplaceReviewStatus(row.ReviewStatus),
			Visibility:    marketv1.MarketplaceVisibility(row.Visibility),
			Publisher:     publisherItemFromEntity(publishers[row.PublisherId], "", false),
			Artifact:      artifactItemFromRecord(artifacts[row.Id]),
			ReviewMessage: row.ReviewMessage,
			SubmittedAt:   unixMillisPtr(row.SubmittedAt),
			UpdatedAt:     unixMillisPtr(row.UpdatedAt),
		})
	}
	return &ReviewQueueOutput{List: items, Total: total}, nil
}

// reviewQueuePluginName selects the requested release locale and preserves the
// identity-table name as the final fallback.
func reviewQueuePluginName(
	row *entity.PluginMarketplaceRelease,
	pluginNames map[int]string,
	displayByRelease map[int][]*entity.PluginMarketplaceDisplayI18n,
	locale string,
) string {
	if row == nil {
		return ""
	}
	name, _ := pickDisplayNameSummary(
		displayByRelease[row.Id],
		locale,
		"",
		pluginNames[row.PluginRecordId],
		"",
	)
	return name
}

type pluginIdentityListFilter struct {
	PageNum    int
	PageSize   int
	Keyword    string
	PluginType marketv1.MarketplacePluginType
	// Status accepts marketplace lifecycle values or process pipeline values.
	Status             string
	OwnerUserID        int64
	PublisherIDs       []int
	MatchPublisherName bool
	Locale             string
}

func (s *serviceImpl) listPluginsFromIdentityTable(
	ctx context.Context,
	in pluginIdentityListFilter,
) (*PluginListOutput, error) {
	pageNum, pageSize := normalizeMarketplacePage(in.PageNum, in.PageSize)
	cols := dao.PluginMarketplacePlugin.Columns()
	model := dao.PluginMarketplacePlugin.Ctx(ctx)
	model = applyOwnerPublisherFilter(
		model,
		dao.PluginMarketplacePublisher.Ctx(ctx),
		in.OwnerUserID,
	)
	if len(in.PublisherIDs) > 0 {
		model = model.WhereIn(cols.PublisherId, in.PublisherIDs)
	}
	if pluginType := normalizeKey(in.PluginType.String()); pluginType != "" {
		model = model.Where(do.PluginMarketplacePlugin{PluginType: pluginType})
	}
	if status := normalizeKey(in.Status); status != "" {
		model = applyPluginListStatusFilter(model, status)
	}
	if keyword := normalizeKey(in.Keyword); keyword != "" {
		like := "%" + keyword + "%"
		builder := model.Builder().
			WhereLike(cols.PluginId, like).
			WhereOrLike(cols.Name, like).
			WhereOrLike(cols.Summary, like)
		if in.MatchPublisherName {
			publisherCols := dao.PluginMarketplacePublisher.Columns()
			publisherModel := dao.PluginMarketplacePublisher.Ctx(ctx).
				Fields(publisherCols.Id).
				Where(
					dao.PluginMarketplacePublisher.Ctx(ctx).Builder().
						WhereLike(publisherCols.PublisherKey, like).
						WhereOrLike(publisherCols.Name, like),
				)
			builder = builder.WhereOrIn(cols.PublisherId, publisherModel)
		}
		model = model.Where(builder)
	}

	total, err := model.Clone().Count()
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeMarketplaceStorageFailed)
	}
	var rows []*entity.PluginMarketplacePlugin
	err = model.Clone().
		Fields(
			cols.Id,
			cols.PublisherId,
			cols.PluginId,
			cols.Name,
			cols.Summary,
			cols.PluginType,
			cols.MarketStatus,
			cols.ProcessStatus,
			cols.Visibility,
			cols.LatestReleaseId,
			cols.LatestVersion,
			cols.DownloadCount,
			cols.SourceKind,
			cols.RepoUrl,
			cols.RepoProvider,
			cols.LastSyncStatus,
			cols.LastSyncMessage,
			cols.LastSyncAt,
			cols.PublishedAt,
			cols.UpdatedAt,
		).
		OrderDesc(cols.UpdatedAt).
		OrderDesc(cols.Id).
		Page(pageNum, pageSize).
		Scan(&rows)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeMarketplaceStorageFailed)
	}

	publishers, err := s.batchPublishersByID(ctx, publisherIDsFromPlugins(rows))
	if err != nil {
		return nil, err
	}
	releases, err := s.batchReleasesByID(ctx, latestReleaseIDsFromPlugins(rows))
	if err != nil {
		return nil, err
	}
	draftReviews, err := s.batchNewestDraftReviewByPluginID(ctx, pluginIDsFromPlugins(rows))
	if err != nil {
		return nil, err
	}
	tagCodesByPlugin, err := s.batchTagCodesForPluginRecords(ctx, pluginRecordIDsFromPlugins(rows))
	if err != nil {
		return nil, err
	}
	displayByRelease, err := s.batchDisplayI18nByReleaseIDs(ctx, latestReleaseIDsFromPlugins(rows))
	if err != nil {
		return nil, err
	}

	items := make([]*marketv1.MarketplacePluginListItem, 0, len(rows))
	for _, row := range rows {
		if row == nil {
			continue
		}
		release := releases[row.LatestReleaseId]
		tagCodes := tagCodesByPlugin[row.Id]
		displayName, displaySummary := pickDisplayNameSummary(
			displayByRelease[row.LatestReleaseId],
			in.Locale,
			"",
			row.Name,
			row.Summary,
		)
		item := &marketv1.MarketplacePluginListItem{
			PluginId:        row.PluginId,
			Name:            displayName,
			Summary:         displaySummary,
			Publisher:       publisherItemFromEntity(publishers[row.PublisherId], "", false),
			PluginType:      marketv1.MarketplacePluginType(row.PluginType),
			MarketStatus:    marketv1.MarketplaceStatus(row.MarketStatus),
			ProcessStatus:   marketv1.MarketplaceProcessStatus(normalizeProcessStatus(row.ProcessStatus)),
			Visibility:      marketv1.MarketplaceVisibility(row.Visibility),
			LatestVersion:   row.LatestVersion,
			TagCodes:        tagCodes,
			PrimaryTag:      firstString(tagCodes),
			DownloadCount:   row.DownloadCount,
			SourceKind:      marketv1.MarketplaceSourceKind(normalizeSourceKind(row.SourceKind)),
			RepoUrl:         row.RepoUrl,
			RepoProvider:    marketv1.MarketplaceRepoProvider(row.RepoProvider),
			LastSyncStatus:  row.LastSyncStatus,
			LastSyncMessage: row.LastSyncMessage,
			LastSyncAt:      unixMillisPtr(row.LastSyncAt),
			PublishedAt:     unixMillisPtr(row.PublishedAt),
			UpdatedAt:       unixMillisPtr(row.UpdatedAt),
			RiskCounts:      marketv1.MarketplaceRiskCounts{},
		}
		if release != nil {
			item.LatestReviewStatus = marketv1.MarketplaceReviewStatus(release.ReviewStatus)
			item.MinHostVersion = release.MinHostVersion
			item.MaxHostVersion = release.MaxHostVersion
			item.RiskCounts = decodeRiskCounts(release.RiskSummary)
			if item.LatestVersion == "" {
				item.LatestVersion = release.ReleaseVersion
			}
		}
		// Prefer the newest mutable draft/rejected review state for the workbench
		// so owners can publish new versions without overwriting the published latest.
		if draft := draftReviews[row.PluginId]; draft != nil {
			item.LatestReviewStatus = marketv1.MarketplaceReviewStatus(draft.ReviewStatus)
			item.LatestVersion = draft.ReleaseVersion
		}
		items = append(items, item)
	}
	return &PluginListOutput{List: items, Total: total}, nil
}

func applyOwnerPublisherFilter(model, publisherModel *gdb.Model, ownerUserID int64) *gdb.Model {
	if ownerUserID <= 0 {
		return model
	}
	publisherCols := dao.PluginMarketplacePublisher.Columns()
	publisherModel = publisherModel.
		Fields(publisherCols.Id).
		Where(publisherCols.OwnerUserId, ownerUserID)
	return model.WhereIn(dao.PluginMarketplacePlugin.Columns().PublisherId, publisherModel)
}

func reviewQueueStatuses(requested marketv1.MarketplaceReviewStatus) []string {
	if status := normalizeKey(requested.String()); status != "" {
		return []string{status}
	}
	return []string{
		marketv1.MarketplaceReviewStatusSubmitted.String(),
		marketv1.MarketplaceReviewStatusReviewing.String(),
	}
}

// classifyPluginListStatusFilter maps a publisher-facing status option to the
// identity-table column it filters. Process pipeline values filter
// process_status; marketplace lifecycle values filter market_status.
func classifyPluginListStatusFilter(status string) (column string, value string, ok bool) {
	status = normalizeKey(status)
	switch status {
	case marketv1.MarketplaceProcessStatusPendingVerify.String(),
		marketv1.MarketplaceProcessStatusPendingReview.String(),
		marketv1.MarketplaceProcessStatusCompleted.String(),
		marketv1.MarketplaceProcessStatusFailed.String():
		return "process_status", status, true
	case marketv1.MarketplaceStatusDraft.String(),
		marketv1.MarketplaceStatusPublished.String(),
		marketv1.MarketplaceStatusDelisted.String(),
		marketv1.MarketplaceStatusDeprecated.String():
		return "market_status", status, true
	default:
		return "", "", false
	}
}

// applyPluginListStatusFilter applies either marketplace lifecycle status or
// process pipeline status filters so publisher UI options match list tags.
func applyPluginListStatusFilter(model *gdb.Model, status string) *gdb.Model {
	column, value, ok := classifyPluginListStatusFilter(status)
	if !ok {
		return model
	}
	switch column {
	case "process_status":
		return model.Where(do.PluginMarketplacePlugin{ProcessStatus: value})
	case "market_status":
		return model.Where(do.PluginMarketplacePlugin{MarketStatus: value})
	default:
		return model
	}
}

func (s *serviceImpl) batchReleasesByID(ctx context.Context, ids []int) (map[int]*entity.PluginMarketplaceRelease, error) {
	out := map[int]*entity.PluginMarketplaceRelease{}
	if len(ids) == 0 {
		return out, nil
	}
	var rows []*entity.PluginMarketplaceRelease
	if err := dao.PluginMarketplaceRelease.Ctx(ctx).
		WhereIn(dao.PluginMarketplaceRelease.Columns().Id, ids).
		Scan(&rows); err != nil {
		return nil, bizerr.WrapCode(err, CodeMarketplaceStorageFailed)
	}
	for _, row := range rows {
		if row != nil {
			out[row.Id] = row
		}
	}
	return out, nil
}

func (s *serviceImpl) batchPluginNamesByRecordID(ctx context.Context, ids []int) (map[int]string, error) {
	out := map[int]string{}
	if len(ids) == 0 {
		return out, nil
	}
	var rows []*entity.PluginMarketplacePlugin
	cols := dao.PluginMarketplacePlugin.Columns()
	if err := dao.PluginMarketplacePlugin.Ctx(ctx).
		Fields(cols.Id, cols.Name).
		WhereIn(cols.Id, ids).
		Scan(&rows); err != nil {
		return nil, bizerr.WrapCode(err, CodeMarketplaceStorageFailed)
	}
	for _, row := range rows {
		if row != nil {
			out[row.Id] = row.Name
		}
	}
	return out, nil
}

// batchTagCodesForPluginRecords loads all page-level plugin tag relations in one query.
func (s *serviceImpl) batchTagCodesForPluginRecords(ctx context.Context, pluginRecordIDs []int) (map[int][]string, error) {
	ids := uniqueInts(pluginRecordIDs)
	if len(ids) == 0 {
		return map[int][]string{}, nil
	}
	var rows []*entity.PluginMarketplacePluginTag
	cols := dao.PluginMarketplacePluginTag.Columns()
	if err := dao.PluginMarketplacePluginTag.Ctx(ctx).
		Fields(cols.PluginRecordId, cols.TagCode).
		WhereIn(cols.PluginRecordId, ids).
		OrderAsc(cols.PluginRecordId).
		OrderAsc(cols.Id).
		Scan(&rows); err != nil {
		return nil, bizerr.WrapCode(err, CodeMarketplaceStorageFailed)
	}
	return tagCodesByPluginRecord(rows), nil
}

// tagCodesByPluginRecord groups one batch query result without changing tag order.
func tagCodesByPluginRecord(rows []*entity.PluginMarketplacePluginTag) map[int][]string {
	out := make(map[int][]string)
	for _, row := range rows {
		if row == nil || row.PluginRecordId <= 0 || normalizeKey(row.TagCode) == "" {
			continue
		}
		out[row.PluginRecordId] = append(out[row.PluginRecordId], row.TagCode)
	}
	for pluginRecordID, codes := range out {
		out[pluginRecordID] = uniqueStrings(codes)
	}
	return out
}

func publisherIDsFromPlugins(rows []*entity.PluginMarketplacePlugin) []int {
	seen := map[int]struct{}{}
	ids := make([]int, 0, len(rows))
	for _, row := range rows {
		if row == nil || row.PublisherId <= 0 {
			continue
		}
		if _, ok := seen[row.PublisherId]; ok {
			continue
		}
		seen[row.PublisherId] = struct{}{}
		ids = append(ids, row.PublisherId)
	}
	return ids
}

func latestReleaseIDsFromPlugins(rows []*entity.PluginMarketplacePlugin) []int {
	seen := map[int]struct{}{}
	ids := make([]int, 0, len(rows))
	for _, row := range rows {
		if row == nil || row.LatestReleaseId <= 0 {
			continue
		}
		if _, ok := seen[row.LatestReleaseId]; ok {
			continue
		}
		seen[row.LatestReleaseId] = struct{}{}
		ids = append(ids, row.LatestReleaseId)
	}
	return ids
}

func pluginIDsFromPlugins(rows []*entity.PluginMarketplacePlugin) []string {
	seen := map[string]struct{}{}
	ids := make([]string, 0, len(rows))
	for _, row := range rows {
		if row == nil {
			continue
		}
		pluginID := normalizeKey(row.PluginId)
		if pluginID == "" {
			continue
		}
		if _, ok := seen[pluginID]; ok {
			continue
		}
		seen[pluginID] = struct{}{}
		ids = append(ids, pluginID)
	}
	return ids
}

// batchNewestDraftReviewByPluginID loads the newest draft/rejected release for
// each plugin so My Plugins can surface publishable workbench state.
func (s *serviceImpl) batchNewestDraftReviewByPluginID(
	ctx context.Context,
	pluginIDs []string,
) (map[string]*entity.PluginMarketplaceRelease, error) {
	out := make(map[string]*entity.PluginMarketplaceRelease, len(pluginIDs))
	if len(pluginIDs) == 0 {
		return out, nil
	}
	var rows []*entity.PluginMarketplaceRelease
	cols := dao.PluginMarketplaceRelease.Columns()
	if err := dao.PluginMarketplaceRelease.Ctx(ctx).
		Fields(
			cols.Id,
			cols.PluginId,
			cols.ReleaseVersion,
			cols.ReleaseStatus,
			cols.ReviewStatus,
			cols.UpdatedAt,
		).
		WhereIn(cols.PluginId, pluginIDs).
		Where(cols.ReleaseStatus, marketv1.MarketplaceStatusDraft.String()).
		WhereIn(cols.ReviewStatus, []string{
			marketv1.MarketplaceReviewStatusDraft.String(),
			marketv1.MarketplaceReviewStatusRejected.String(),
		}).
		OrderDesc(cols.UpdatedAt).
		OrderDesc(cols.Id).
		Scan(&rows); err != nil {
		return nil, bizerr.WrapCode(err, CodeMarketplaceStorageFailed)
	}
	for _, row := range rows {
		if row == nil {
			continue
		}
		pluginID := normalizeKey(row.PluginId)
		if pluginID == "" {
			continue
		}
		if _, exists := out[pluginID]; exists {
			continue
		}
		out[pluginID] = row
	}
	return out, nil
}

func pluginRecordIDsFromPlugins(rows []*entity.PluginMarketplacePlugin) []int {
	ids := make([]int, 0, len(rows))
	for _, row := range rows {
		if row != nil && row.Id > 0 {
			ids = append(ids, row.Id)
		}
	}
	return ids
}

func pluginRecordIDsFromReleases(rows []*entity.PluginMarketplaceRelease) []int {
	seen := map[int]struct{}{}
	ids := make([]int, 0, len(rows))
	for _, row := range rows {
		if row == nil || row.PluginRecordId <= 0 {
			continue
		}
		if _, ok := seen[row.PluginRecordId]; ok {
			continue
		}
		seen[row.PluginRecordId] = struct{}{}
		ids = append(ids, row.PluginRecordId)
	}
	return ids
}

func publisherIDsFromReleases(rows []*entity.PluginMarketplaceRelease) []int {
	seen := map[int]struct{}{}
	ids := make([]int, 0, len(rows))
	for _, row := range rows {
		if row == nil || row.PublisherId <= 0 {
			continue
		}
		if _, ok := seen[row.PublisherId]; ok {
			continue
		}
		seen[row.PublisherId] = struct{}{}
		ids = append(ids, row.PublisherId)
	}
	return ids
}
