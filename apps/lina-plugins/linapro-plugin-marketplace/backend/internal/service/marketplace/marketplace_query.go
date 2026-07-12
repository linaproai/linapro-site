// This file implements marketplace catalog read paths, read-model refresh, and
// release risk projections. Catalog list queries use the plugin read-model table
// for bounded pagination, then batch-load publisher snapshots for the current
// page. Detail and release queries load only the bounded associated rows needed
// by the requested projection.

package marketplace

import (
	"archive/zip"
	"context"
	"encoding/json"
	"html"
	"strings"
	"time"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/encoding/gjson"

	"lina-core/pkg/bizerr"
	marketv1 "linapro-plugin-marketplace/backend/api/market/v1"
	"linapro-plugin-marketplace/backend/internal/dao"
	"linapro-plugin-marketplace/backend/internal/model/do"
	"linapro-plugin-marketplace/backend/internal/model/entity"
)

const (
	defaultMarketplacePageNum  = 1
	defaultMarketplacePageSize = 20
	maxMarketplacePageSize     = 100

	sourceDeliverySourceRebuildRequired = "source_rebuild_required"
	sourceDeliveryDynamicUploadRequired = "dynamic_upload_required"
)

// ListPlugins returns a read-model-backed paginated marketplace catalog page.
func (s *serviceImpl) ListPlugins(ctx context.Context, in ListPluginsInput) (*PluginListOutput, error) {
	pageNum, pageSize := normalizeMarketplacePage(in.PageNum, in.PageSize)
	model, empty, err := s.buildPluginListModel(ctx, in)
	if err != nil {
		return nil, err
	}
	if empty {
		return &PluginListOutput{List: []*marketv1.MarketplacePluginListItem{}, Total: 0}, nil
	}

	total, err := model.Clone().Count()
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeMarketplaceStorageFailed)
	}
	var rows []*entity.PluginMarketplacePluginReadModel
	cols := dao.PluginMarketplacePluginReadModel.Columns()
	err = model.Clone().
		Fields(
			cols.PluginRecordId,
			cols.PublisherId,
			cols.PublisherName,
			cols.PublisherVerified,
			cols.PluginId,
			cols.Name,
			cols.Summary,
			cols.PluginType,
			cols.MarketStatus,
			cols.Visibility,
			cols.LatestReleaseId,
			cols.LatestVersion,
			cols.MinHostVersion,
			cols.MaxHostVersion,
			cols.PrimaryTag,
			cols.TagCodes,
			cols.RiskCounts,
			cols.DownloadCount,
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

	publishers, err := s.batchPublishersByID(ctx, publisherIDsFromReadModels(rows))
	if err != nil {
		return nil, err
	}
	items := make([]*marketv1.MarketplacePluginListItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, pluginListItemFromReadModel(row, publishers[row.PublisherId]))
	}
	return &PluginListOutput{List: items, Total: total}, nil
}

// GetPluginDetail returns one marketplace detail projection.
func (s *serviceImpl) GetPluginDetail(ctx context.Context, in GetPluginDetailInput) (*PluginDetailOutput, error) {
	plugin, owned, err := s.resolveAccessiblePlugin(ctx, in.PluginID, in.Visibility)
	if err != nil {
		return nil, err
	}
	if plugin == nil {
		return nil, bizerr.NewCode(CodeMarketplacePluginNotFound)
	}

	var latestRelease *entity.PluginMarketplaceRelease
	if plugin.LatestReleaseId > 0 {
		managementAccess := marketplaceManagementReadAllowed(owned, in.Visibility)
		if managementAccess {
			latestRelease, err = s.getReleaseByID(ctx, plugin.LatestReleaseId)
		} else {
			latestRelease, err = s.getVisibleReleaseByID(
				ctx,
				plugin.LatestReleaseId,
				in.Visibility,
				marketplaceVisibilityPermissionView,
			)
		}
		if err != nil {
			return nil, err
		}
		if latestRelease == nil && !managementAccess {
			return nil, bizerr.NewCode(CodeMarketplacePluginNotFound)
		}
	}
	publishers, err := s.batchPublishersByID(ctx, []int{plugin.PublisherId})
	if err != nil {
		return nil, err
	}
	tags, err := s.tagsForPluginRecord(ctx, plugin.Id)
	if err != nil {
		return nil, err
	}
	artifacts, err := s.batchPrimaryArtifactsByRelease(ctx, []*entity.PluginMarketplaceRelease{latestRelease})
	if err != nil {
		return nil, err
	}
	detail := pluginDetailItemFromEntities(plugin, publishers[plugin.PublisherId], tags, latestRelease, artifacts)
	return &PluginDetailOutput{Plugin: detail}, nil
}

// ListReleases returns a paginated release page for one marketplace plugin.
func (s *serviceImpl) ListReleases(ctx context.Context, in ListReleasesInput) (*ReleaseListOutput, error) {
	plugin, owned, err := s.resolveAccessiblePlugin(ctx, in.PluginID, in.Visibility)
	if err != nil {
		return nil, err
	}
	if plugin == nil {
		return nil, bizerr.NewCode(CodeMarketplacePluginNotFound)
	}

	pageNum, pageSize := normalizeMarketplacePage(in.PageNum, in.PageSize)
	model := dao.PluginMarketplaceRelease.Ctx(ctx).
		Where(do.PluginMarketplaceRelease{PluginRecordId: plugin.Id})
	releaseCols := dao.PluginMarketplaceRelease.Columns()
	managementAccess := marketplaceManagementReadAllowed(owned, in.Visibility)
	if !managementAccess {
		model = s.applyMarketplaceVisibilityFilter(
			ctx,
			model,
			releaseCols.Visibility,
			releaseCols.PluginRecordId,
			in.Visibility,
			marketplaceVisibilityPermissionView,
		)
	}
	requestedStatus := normalizeKey(in.Status.String())
	statusFilter, statusAllowed := marketplaceReleaseStatusFilter(requestedStatus, managementAccess)
	if !statusAllowed {
		return &ReleaseListOutput{List: []*marketv1.MarketplaceReleaseItem{}, Total: 0}, nil
	}
	if statusFilter != "" {
		model = model.Where(do.PluginMarketplaceRelease{ReleaseStatus: statusFilter})
	}
	if normalizeKey(in.ReviewStatus.String()) != "" {
		model = model.Where(do.PluginMarketplaceRelease{ReviewStatus: in.ReviewStatus.String()})
	}

	total, err := model.Clone().Count()
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeMarketplaceStorageFailed)
	}
	var rows []*entity.PluginMarketplaceRelease
	cols := dao.PluginMarketplaceRelease.Columns()
	err = model.Clone().
		Fields(
			cols.Id,
			cols.PluginId,
			cols.ReleaseVersion,
			cols.PluginType,
			cols.ReleaseStatus,
			cols.ReviewStatus,
			cols.Visibility,
			cols.MinHostVersion,
			cols.MaxHostVersion,
			cols.ReviewMessage,
			cols.SubmittedAt,
			cols.ReviewedAt,
			cols.PublishedAt,
			cols.UpdatedAt,
		).
		OrderDesc(cols.PublishedAt).
		OrderDesc(cols.Id).
		Page(pageNum, pageSize).
		Scan(&rows)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeMarketplaceStorageFailed)
	}

	artifacts, err := s.batchPrimaryArtifactsByRelease(ctx, rows)
	if err != nil {
		return nil, err
	}
	items := make([]*marketv1.MarketplaceReleaseItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, releaseItemFromEntity(row, artifacts[row.Id]))
	}
	return &ReleaseListOutput{List: items, Total: total}, nil
}

// ListReleaseRisks returns paginated scanner risk findings for one release.
func (s *serviceImpl) ListReleaseRisks(ctx context.Context, in ListReleaseRisksInput) (*RiskListOutput, error) {
	release, err := s.requireVisibleRelease(
		ctx,
		in.PluginID,
		in.Version,
		in.Visibility,
		marketplaceVisibilityPermissionView,
	)
	if err != nil {
		return nil, err
	}

	pageNum, pageSize := normalizeMarketplacePage(in.PageNum, in.PageSize)
	model := dao.PluginMarketplaceRisk.Ctx(ctx).Where(do.PluginMarketplaceRisk{ReleaseId: release.Id})
	if normalizeKey(in.Type.String()) != "" {
		model = model.Where(do.PluginMarketplaceRisk{RiskType: in.Type.String()})
	}
	if normalizeKey(in.Severity.String()) != "" {
		model = model.Where(do.PluginMarketplaceRisk{Severity: in.Severity.String()})
	}

	total, err := model.Clone().Count()
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeMarketplaceStorageFailed)
	}
	var rows []*entity.PluginMarketplaceRisk
	cols := dao.PluginMarketplaceRisk.Columns()
	err = model.Clone().
		Fields(cols.RiskType, cols.Severity, cols.Source, cols.Summary, cols.Payload, cols.CreatedAt).
		OrderDesc(cols.Id).
		Page(pageNum, pageSize).
		Scan(&rows)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeMarketplaceStorageFailed)
	}
	items := make([]*marketv1.MarketplaceRiskItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, riskItemFromEntity(row))
	}
	return &RiskListOutput{List: items, Total: total}, nil
}

// GetReleaseDocument returns one marketplace document projection with rendered body.
func (s *serviceImpl) GetReleaseDocument(ctx context.Context, in GetReleaseDocumentInput) (*DocumentOutput, error) {
	record, err := s.ResolveReleaseDocumentIndex(ctx, in)
	if err != nil {
		return nil, err
	}
	item := documentItemFromRecord(record)
	if item == nil {
		return &DocumentOutput{}, nil
	}
	content, err := s.loadDocumentRenderedContent(ctx, record)
	if err != nil {
		return nil, err
	}
	if content != "" {
		item.Content = content
	}
	return &DocumentOutput{Document: item}, nil
}

// rebuildPluginReadModel refreshes the catalog projection for one marketplace plugin.
func (s *serviceImpl) rebuildPluginReadModel(ctx context.Context, pluginRecordID int) error {
	var plugin *entity.PluginMarketplacePlugin
	if err := dao.PluginMarketplacePlugin.Ctx(ctx).
		Where(do.PluginMarketplacePlugin{Id: pluginRecordID}).
		Scan(&plugin); err != nil {
		return bizerr.WrapCode(err, CodeMarketplaceStorageFailed)
	}
	if plugin == nil {
		return bizerr.NewCode(CodeMarketplacePluginNotFound)
	}
	if plugin.LatestReleaseId <= 0 {
		return s.deletePluginReadModel(ctx, plugin.PluginId)
	}

	release, err := s.getReleaseByID(ctx, plugin.LatestReleaseId)
	if err != nil {
		return err
	}
	if release == nil {
		return s.deletePluginReadModel(ctx, plugin.PluginId)
	}
	publisher, err := s.getPublisherByID(ctx, plugin.PublisherId)
	if err != nil {
		return err
	}
	tagCodes, err := s.tagCodesForPluginRecord(ctx, plugin.Id)
	if err != nil {
		return err
	}
	tagCodesJSON, err := packageJSONString(tagCodes)
	if err != nil {
		return err
	}

	data := do.PluginMarketplacePluginReadModel{
		PluginRecordId:    plugin.Id,
		PublisherId:       plugin.PublisherId,
		PublisherName:     publisher.Name,
		PublisherVerified: publisher.Verified,
		PluginId:          plugin.PluginId,
		Name:              plugin.Name,
		Summary:           plugin.Summary,
		PluginType:        plugin.PluginType,
		MarketStatus:      plugin.MarketStatus,
		Visibility:        marketplaceEffectiveVisibility(plugin.Visibility, release.Visibility),
		LatestReleaseId:   release.Id,
		LatestVersion:     release.ReleaseVersion,
		MinHostVersion:    release.MinHostVersion,
		MaxHostVersion:    release.MaxHostVersion,
		PrimaryTag:        firstString(tagCodes),
		TagCodes:          tagCodesJSON,
		RiskCounts:        defaultJSONString(release.RiskSummary, defaultObjectSummary),
		DownloadCount:     plugin.DownloadCount,
		PublishedAt:       cloneTime(release.PublishedAt),
		SearchText:        marketplaceSearchText(plugin, publisher, release, tagCodes),
	}

	existing, err := s.getReadModelByPluginID(ctx, plugin.PluginId)
	if err != nil {
		return err
	}
	if existing == nil {
		if _, err = dao.PluginMarketplacePluginReadModel.Ctx(ctx).Data(data).Insert(); err != nil {
			return bizerr.WrapCode(err, CodeMarketplaceStorageFailed)
		}
		return nil
	}
	if _, err = dao.PluginMarketplacePluginReadModel.Ctx(ctx).
		Where(do.PluginMarketplacePluginReadModel{Id: existing.Id}).
		Data(data).
		Update(); err != nil {
		return bizerr.WrapCode(err, CodeMarketplaceStorageFailed)
	}
	return nil
}

// replaceReleaseRisks replaces scanner risk finding rows for one mutable release.
func (s *serviceImpl) replaceReleaseRisks(
	ctx context.Context,
	release *ReleaseRecord,
	diagnostics []*PackageDiagnostic,
) error {
	if release == nil {
		return bizerr.NewCode(CodeMarketplaceInvalidInput)
	}

	var existingRows []*entity.PluginMarketplaceRisk
	if err := dao.PluginMarketplaceRisk.Ctx(ctx).
		Where(do.PluginMarketplaceRisk{ReleaseId: release.ID}).
		Scan(&existingRows); err != nil {
		return bizerr.WrapCode(err, CodeMarketplaceStorageFailed)
	}
	return dao.PluginMarketplaceRisk.Transaction(ctx, func(ctx context.Context, _ gdb.TX) error {
		for _, row := range existingRows {
			if row == nil {
				continue
			}
			if _, err := dao.PluginMarketplaceRisk.Ctx(ctx).
				Where(do.PluginMarketplaceRisk{Id: row.Id}).
				Delete(); err != nil {
				return bizerr.WrapCode(err, CodeMarketplaceStorageFailed)
			}
		}
		for _, diagnostic := range diagnostics {
			if diagnostic == nil {
				continue
			}
			payload, err := packageJSONString(packageDiagnosticRiskPayload{Code: diagnostic.Code})
			if err != nil {
				return err
			}
			if _, err = dao.PluginMarketplaceRisk.Ctx(ctx).Data(do.PluginMarketplaceRisk{
				ReleaseId:      release.ID,
				PluginId:       release.PluginID,
				ReleaseVersion: release.Version,
				RiskType:       diagnosticRiskType(diagnostic).String(),
				Severity:       diagnostic.Severity.String(),
				Source:         normalizeKey(diagnostic.Source),
				Summary:        normalizeKey(diagnostic.Message),
				Payload:        payload,
			}).Insert(); err != nil {
				return bizerr.WrapCode(err, CodeMarketplaceStorageFailed)
			}
		}
		return nil
	})
}

// buildPluginListModel applies bounded database-side filters to the read-model query.
func (s *serviceImpl) buildPluginListModel(
	ctx context.Context,
	in ListPluginsInput,
) (*gdb.Model, bool, error) {
	cols := dao.PluginMarketplacePluginReadModel.Columns()
	model := dao.PluginMarketplacePluginReadModel.Ctx(ctx).
		Where(pluginListReadModelBaseCriteria())
	model = s.applyMarketplaceVisibilityFilter(
		ctx,
		model,
		cols.Visibility,
		cols.PluginRecordId,
		in.Visibility,
		marketplaceVisibilityPermissionView,
	)

	if normalizeKey(in.PluginType.String()) != "" {
		model = model.Where(do.PluginMarketplacePluginReadModel{PluginType: in.PluginType.String()})
	}
	if keyword := normalizeKey(in.Keyword); keyword != "" {
		model = model.WhereLike(cols.SearchText, "%"+keyword+"%")
	}
	if hostVersion := normalizeKey(in.HostVersion); hostVersion != "" {
		model = applyMarketplaceHostVersionFilter(model, cols.MinHostVersion, cols.MaxHostVersion, hostVersion)
	}
	if publisherKey := normalizeKey(in.Publisher); publisherKey != "" {
		publisher, err := s.getPublisherByKey(ctx, publisherKey)
		if err != nil {
			return nil, false, err
		}
		if publisher == nil {
			return nil, true, nil
		}
		model = model.Where(do.PluginMarketplacePluginReadModel{PublisherId: publisher.Id})
	}
	if tagCode := normalizeKey(in.TagCode); tagCode != "" {
		ids, err := s.pluginRecordIDsByTag(ctx, tagCode)
		if err != nil {
			return nil, false, err
		}
		if len(ids) == 0 {
			return nil, true, nil
		}
		model = model.WhereIn(cols.PluginRecordId, ids)
	}
	return model, false, nil
}

// pluginListReadModelBaseCriteria limits catalog reads to published read-model rows.
func pluginListReadModelBaseCriteria() do.PluginMarketplacePluginReadModel {
	return do.PluginMarketplacePluginReadModel{MarketStatus: marketv1.MarketplaceStatusPublished.String()}
}

// resolveAccessiblePlugin returns a published visible plugin or an unpublished
// identity row that belongs to the caller or is available to a reviewer.
// owned is true only when the caller owns the publisher profile.
func (s *serviceImpl) resolveAccessiblePlugin(
	ctx context.Context,
	pluginID string,
	subject VisibilitySubject,
) (*entity.PluginMarketplacePlugin, bool, error) {
	plugin, err := s.getPublishedPluginByID(ctx, pluginID, subject)
	if err != nil {
		return nil, false, err
	}
	if plugin != nil {
		owned, ownerErr := s.pluginOwnedByUser(ctx, plugin, subject.UserID)
		if ownerErr != nil {
			return nil, false, ownerErr
		}
		if !marketplacePublishedReadAllowed(owned, subject) {
			return nil, false, nil
		}
		return plugin, owned, nil
	}
	if subject.UserID <= 0 {
		return nil, false, nil
	}
	plugin, err = s.getPluginByID(ctx, pluginID)
	if err != nil {
		return nil, false, err
	}
	if plugin == nil {
		return nil, false, nil
	}
	owned, err := s.pluginOwnedByUser(ctx, plugin, subject.UserID)
	if err != nil {
		return nil, false, err
	}
	if !marketplaceManagementReadAllowed(owned, subject) {
		return nil, false, nil
	}
	return plugin, owned, nil
}

// pluginOwnedByUser verifies publisher ownership without exposing publisher data.
func (s *serviceImpl) pluginOwnedByUser(
	ctx context.Context,
	plugin *entity.PluginMarketplacePlugin,
	userID int64,
) (bool, error) {
	if plugin == nil || plugin.PublisherId <= 0 || userID <= 0 {
		return false, nil
	}
	count, err := dao.PluginMarketplacePublisher.Ctx(ctx).
		Where(do.PluginMarketplacePublisher{Id: plugin.PublisherId, OwnerUserId: userID}).
		Count()
	if err != nil {
		return false, bizerr.WrapCode(err, CodeMarketplaceStorageFailed)
	}
	return count > 0, nil
}

// marketplaceManagementReadAllowed centralizes unpublished workbench access.
func marketplaceManagementReadAllowed(owned bool, subject VisibilitySubject) bool {
	return (owned && subject.CanPublish) || subject.CanReview
}

// marketplacePublishedReadAllowed keeps My endpoints owner-only while public
// and reviewer endpoints retain their intended published visibility behavior.
func marketplacePublishedReadAllowed(owned bool, subject VisibilitySubject) bool {
	return !subject.CanPublish || subject.CanReview || owned
}

// marketplaceReleaseStatusFilter prevents public filters from widening the
// release query beyond published rows.
func marketplaceReleaseStatusFilter(requestedStatus string, managementAccess bool) (string, bool) {
	if managementAccess {
		return requestedStatus, true
	}
	if requestedStatus != "" && requestedStatus != marketv1.MarketplaceStatusPublished.String() {
		return "", false
	}
	return marketv1.MarketplaceStatusPublished.String(), true
}

// marketplaceReleaseManagementReadAllowed keeps unpublished inspection on
// view paths; reviewer or publisher state never expands download permission.
func marketplaceReleaseManagementReadAllowed(
	owned bool,
	subject VisibilitySubject,
	permission marketplaceVisibilityPermission,
) bool {
	return permission == marketplaceVisibilityPermissionView && marketplaceManagementReadAllowed(owned, subject)
}

// applyMarketplaceHostVersionFilter keeps compatibility filtering in the database query.
func applyMarketplaceHostVersionFilter(model *gdb.Model, minColumn string, maxColumn string, hostVersion string) *gdb.Model {
	return model.
		Where("("+minColumn+" = '' OR "+minColumn+" <= ?)", hostVersion).
		Where("("+maxColumn+" = '' OR "+maxColumn+" >= ?)", hostVersion)
}

// getPublishedPluginByID loads one visible published marketplace plugin by ID.
func (s *serviceImpl) getPublishedPluginByID(
	ctx context.Context,
	pluginID string,
	subject VisibilitySubject,
) (*entity.PluginMarketplacePlugin, error) {
	normalizedID := normalizeKey(pluginID)
	if normalizedID == "" {
		return nil, nil
	}
	var plugin *entity.PluginMarketplacePlugin
	cols := dao.PluginMarketplacePlugin.Columns()
	model := dao.PluginMarketplacePlugin.Ctx(ctx).
		Where(do.PluginMarketplacePlugin{
			PluginId:     normalizedID,
			MarketStatus: marketv1.MarketplaceStatusPublished.String(),
		})
	model = s.applyMarketplaceVisibilityFilter(
		ctx,
		model,
		cols.Visibility,
		cols.Id,
		subject,
		marketplaceVisibilityPermissionView,
	)
	if err := model.Scan(&plugin); err != nil {
		return nil, bizerr.WrapCode(err, CodeMarketplaceStorageFailed)
	}
	return plugin, nil
}

// requireVisibleRelease loads one accessible release after plugin and release filtering.
// Owners and reviewers may read unpublished releases for workbench inspection.
func (s *serviceImpl) requireVisibleRelease(
	ctx context.Context,
	pluginID string,
	version string,
	subject VisibilitySubject,
	permission marketplaceVisibilityPermission,
) (*entity.PluginMarketplaceRelease, error) {
	plugin, owned, err := s.resolveAccessiblePlugin(ctx, pluginID, subject)
	if err != nil {
		return nil, err
	}
	if plugin != nil {
		if marketplaceReleaseManagementReadAllowed(owned, subject, permission) {
			release, err := s.getReleaseByPluginVersion(ctx, plugin.PluginId, version)
			if err != nil {
				return nil, err
			}
			if release == nil {
				return nil, bizerr.NewCode(CodeMarketplaceReleaseNotFound)
			}
			return release, nil
		}
		release, err := s.getVisibleReleaseByPluginVersion(ctx, plugin.PluginId, version, subject, permission)
		if err != nil {
			return nil, err
		}
		if release != nil {
			return release, nil
		}
	}

	return nil, bizerr.NewCode(CodeMarketplaceReleaseNotFound)
}

// getVisibleReleaseByPluginVersion loads one published release through a visibility predicate.
func (s *serviceImpl) getVisibleReleaseByPluginVersion(
	ctx context.Context,
	pluginID string,
	version string,
	subject VisibilitySubject,
	permission marketplaceVisibilityPermission,
) (*entity.PluginMarketplaceRelease, error) {
	normalizedPluginID := normalizeKey(pluginID)
	normalizedVersion := normalizeKey(version)
	if normalizedPluginID == "" || normalizedVersion == "" {
		return nil, nil
	}

	var release *entity.PluginMarketplaceRelease
	cols := dao.PluginMarketplaceRelease.Columns()
	model := dao.PluginMarketplaceRelease.Ctx(ctx).
		Where(do.PluginMarketplaceRelease{
			PluginId:       normalizedPluginID,
			ReleaseVersion: normalizedVersion,
			ReleaseStatus:  marketv1.MarketplaceStatusPublished.String(),
		})
	model = s.applyMarketplaceVisibilityFilter(
		ctx,
		model,
		cols.Visibility,
		cols.PluginRecordId,
		subject,
		permission,
	)
	if err := model.Scan(&release); err != nil {
		return nil, bizerr.WrapCode(err, CodeMarketplaceStorageFailed)
	}
	return release, nil
}

// getVisibleReleaseByID loads one published release by ID through a visibility predicate.
func (s *serviceImpl) getVisibleReleaseByID(
	ctx context.Context,
	id int,
	subject VisibilitySubject,
	permission marketplaceVisibilityPermission,
) (*entity.PluginMarketplaceRelease, error) {
	if id <= 0 {
		return nil, nil
	}

	var release *entity.PluginMarketplaceRelease
	cols := dao.PluginMarketplaceRelease.Columns()
	model := dao.PluginMarketplaceRelease.Ctx(ctx).
		Where(do.PluginMarketplaceRelease{
			Id:            id,
			ReleaseStatus: marketv1.MarketplaceStatusPublished.String(),
		})
	model = s.applyMarketplaceVisibilityFilter(
		ctx,
		model,
		cols.Visibility,
		cols.PluginRecordId,
		subject,
		permission,
	)
	if err := model.Scan(&release); err != nil {
		return nil, bizerr.WrapCode(err, CodeMarketplaceStorageFailed)
	}
	return release, nil
}

// getReleaseByID loads one release by primary key.
func (s *serviceImpl) getReleaseByID(ctx context.Context, id int) (*entity.PluginMarketplaceRelease, error) {
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

// getReadModelByPluginID loads one read-model row by plugin ID.
func (s *serviceImpl) getReadModelByPluginID(
	ctx context.Context,
	pluginID string,
) (*entity.PluginMarketplacePluginReadModel, error) {
	var row *entity.PluginMarketplacePluginReadModel
	if err := dao.PluginMarketplacePluginReadModel.Ctx(ctx).
		Where(do.PluginMarketplacePluginReadModel{PluginId: normalizeKey(pluginID)}).
		Scan(&row); err != nil {
		return nil, bizerr.WrapCode(err, CodeMarketplaceStorageFailed)
	}
	return row, nil
}

// deletePluginReadModel removes a stale read-model row for plugins without a publish anchor.
func (s *serviceImpl) deletePluginReadModel(ctx context.Context, pluginID string) error {
	if normalizeKey(pluginID) == "" {
		return nil
	}
	if _, err := dao.PluginMarketplacePluginReadModel.Ctx(ctx).
		Where(do.PluginMarketplacePluginReadModel{PluginId: normalizeKey(pluginID)}).
		Delete(); err != nil {
		return bizerr.WrapCode(err, CodeMarketplaceStorageFailed)
	}
	return nil
}

// pluginRecordIDsByTag resolves a tag filter to plugin record IDs in one query.
func (s *serviceImpl) pluginRecordIDsByTag(ctx context.Context, tagCode string) ([]int, error) {
	var rows []*entity.PluginMarketplacePluginTag
	if err := dao.PluginMarketplacePluginTag.Ctx(ctx).
		Fields(dao.PluginMarketplacePluginTag.Columns().PluginRecordId).
		Where(do.PluginMarketplacePluginTag{TagCode: normalizeKey(tagCode)}).
		Scan(&rows); err != nil {
		return nil, bizerr.WrapCode(err, CodeMarketplaceStorageFailed)
	}
	ids := make([]int, 0, len(rows))
	for _, row := range rows {
		if row == nil {
			continue
		}
		ids = append(ids, row.PluginRecordId)
	}
	return uniqueInts(ids), nil
}

// tagCodesForPluginRecord returns the stable tag code list for one plugin.
func (s *serviceImpl) tagCodesForPluginRecord(ctx context.Context, pluginRecordID int) ([]string, error) {
	var rows []*entity.PluginMarketplacePluginTag
	cols := dao.PluginMarketplacePluginTag.Columns()
	if err := dao.PluginMarketplacePluginTag.Ctx(ctx).
		Fields(cols.TagCode).
		Where(do.PluginMarketplacePluginTag{PluginRecordId: pluginRecordID}).
		OrderAsc(cols.Id).
		Scan(&rows); err != nil {
		return nil, bizerr.WrapCode(err, CodeMarketplaceStorageFailed)
	}
	codes := make([]string, 0, len(rows))
	for _, row := range rows {
		if row == nil || normalizeKey(row.TagCode) == "" {
			continue
		}
		codes = append(codes, row.TagCode)
	}
	return uniqueStrings(codes), nil
}

// tagsForPluginRecord batch-loads tag display metadata for one plugin.
func (s *serviceImpl) tagsForPluginRecord(ctx context.Context, pluginRecordID int) ([]*marketv1.MarketplaceTagItem, error) {
	codes, err := s.tagCodesForPluginRecord(ctx, pluginRecordID)
	if err != nil {
		return nil, err
	}
	if len(codes) == 0 {
		return []*marketv1.MarketplaceTagItem{}, nil
	}
	var rows []*entity.PluginMarketplaceTag
	tagCols := dao.PluginMarketplaceTag.Columns()
	if err = dao.PluginMarketplaceTag.Ctx(ctx).
		Fields(tagCols.TagCode, tagCols.Name, tagCols.TagType).
		WhereIn(tagCols.TagCode, codes).
		OrderAsc(tagCols.Sort).
		OrderAsc(tagCols.Id).
		Scan(&rows); err != nil {
		return nil, bizerr.WrapCode(err, CodeMarketplaceStorageFailed)
	}
	byCode := make(map[string]*entity.PluginMarketplaceTag, len(rows))
	for _, row := range rows {
		if row == nil {
			continue
		}
		byCode[row.TagCode] = row
	}
	items := make([]*marketv1.MarketplaceTagItem, 0, len(codes))
	for _, code := range codes {
		row := byCode[code]
		if row == nil {
			items = append(items, &marketv1.MarketplaceTagItem{Code: code, Name: code})
			continue
		}
		items = append(items, &marketv1.MarketplaceTagItem{
			Code: row.TagCode,
			Name: row.Name,
			Type: row.TagType,
		})
	}
	return items, nil
}

// batchPublishersByID loads publisher rows for the current page in one query.
func (s *serviceImpl) batchPublishersByID(
	ctx context.Context,
	ids []int,
) (map[int]*entity.PluginMarketplacePublisher, error) {
	unique := uniqueInts(ids)
	if len(unique) == 0 {
		return map[int]*entity.PluginMarketplacePublisher{}, nil
	}
	var rows []*entity.PluginMarketplacePublisher
	cols := dao.PluginMarketplacePublisher.Columns()
	if err := dao.PluginMarketplacePublisher.Ctx(ctx).
		Fields(cols.Id, cols.PublisherKey, cols.Name, cols.Summary, cols.Verified, cols.Homepage).
		WhereIn(cols.Id, unique).
		Scan(&rows); err != nil {
		return nil, bizerr.WrapCode(err, CodeMarketplaceStorageFailed)
	}
	byID := make(map[int]*entity.PluginMarketplacePublisher, len(rows))
	for _, row := range rows {
		if row == nil {
			continue
		}
		byID[row.Id] = row
	}
	return byID, nil
}

// batchPrimaryArtifactsByRelease loads artifacts for the current release page in one query.
func (s *serviceImpl) batchPrimaryArtifactsByRelease(
	ctx context.Context,
	releases []*entity.PluginMarketplaceRelease,
) (map[int]*ArtifactRecord, error) {
	releaseIDs := releaseIDsFromEntities(releases)
	if len(releaseIDs) == 0 {
		return map[int]*ArtifactRecord{}, nil
	}
	var rows []*entity.PluginMarketplaceArtifact
	cols := dao.PluginMarketplaceArtifact.Columns()
	if err := dao.PluginMarketplaceArtifact.Ctx(ctx).
		Fields(
			cols.Id,
			cols.ReleaseId,
			cols.PluginId,
			cols.ReleaseVersion,
			cols.ArtifactType,
			cols.StorageKey,
			cols.FileName,
			cols.ContentType,
			cols.SizeBytes,
			cols.Sha256,
			cols.ManifestSha256,
			cols.WasmSha256,
			cols.UpdatedAt,
		).
		WhereIn(cols.ReleaseId, releaseIDs).
		OrderAsc(cols.ReleaseId).
		OrderAsc(cols.ArtifactType).
		Scan(&rows); err != nil {
		return nil, bizerr.WrapCode(err, CodeMarketplaceStorageFailed)
	}

	releaseTypes := make(map[int]marketv1.MarketplacePluginType, len(releases))
	for _, release := range releases {
		if release == nil {
			continue
		}
		releaseTypes[release.Id] = marketv1.MarketplacePluginType(release.PluginType)
	}
	byRelease := make(map[int]*entity.PluginMarketplaceArtifact, len(releaseIDs))
	for _, row := range rows {
		if row == nil {
			continue
		}
		current := byRelease[row.ReleaseId]
		if current == nil || artifactPriority(row, releaseTypes[row.ReleaseId]) < artifactPriority(current, releaseTypes[row.ReleaseId]) {
			byRelease[row.ReleaseId] = row
		}
	}
	out := make(map[int]*ArtifactRecord, len(byRelease))
	for releaseID, row := range byRelease {
		out[releaseID] = artifactRecordFromEntity(row)
	}
	return out, nil
}

// pluginListItemFromReadModel projects one read-model row to API DTO.
func pluginListItemFromReadModel(
	row *entity.PluginMarketplacePluginReadModel,
	publisher *entity.PluginMarketplacePublisher,
) *marketv1.MarketplacePluginListItem {
	if row == nil {
		return nil
	}
	return &marketv1.MarketplacePluginListItem{
		PluginId:       row.PluginId,
		Name:           row.Name,
		Summary:        row.Summary,
		Publisher:      publisherItemFromEntity(publisher, row.PublisherName, row.PublisherVerified),
		PluginType:     marketv1.MarketplacePluginType(row.PluginType),
		MarketStatus:   marketv1.MarketplaceStatus(row.MarketStatus),
		Visibility:     marketv1.MarketplaceVisibility(row.Visibility),
		LatestVersion:  row.LatestVersion,
		MinHostVersion: row.MinHostVersion,
		MaxHostVersion: row.MaxHostVersion,
		PrimaryTag:     row.PrimaryTag,
		TagCodes:       decodeStringArrayJSON(row.TagCodes),
		RiskCounts:     decodeRiskCounts(row.RiskCounts),
		DownloadCount:  row.DownloadCount,
		PublishedAt:    unixMillisPtr(row.PublishedAt),
		UpdatedAt:      unixMillisPtr(row.UpdatedAt),
	}
}

// pluginDetailItemFromEntities projects plugin, release, tag, publisher, and artifact rows.
func pluginDetailItemFromEntities(
	plugin *entity.PluginMarketplacePlugin,
	publisher *entity.PluginMarketplacePublisher,
	tags []*marketv1.MarketplaceTagItem,
	latestRelease *entity.PluginMarketplaceRelease,
	artifacts map[int]*ArtifactRecord,
) *marketv1.MarketplacePluginDetailItem {
	if plugin == nil {
		return nil
	}
	riskCounts := marketv1.MarketplaceRiskCounts{}
	var latestItem *marketv1.MarketplaceReleaseItem
	if latestRelease != nil {
		riskCounts = decodeRiskCounts(latestRelease.RiskSummary)
		latestItem = releaseItemFromEntity(latestRelease, artifacts[latestRelease.Id])
	}
	return &marketv1.MarketplacePluginDetailItem{
		PluginId:       plugin.PluginId,
		Name:           plugin.Name,
		Summary:        plugin.Summary,
		Description:    plugin.Description,
		Publisher:      publisherItemFromEntity(publisher, "", false),
		PluginType:     marketv1.MarketplacePluginType(plugin.PluginType),
		MarketStatus:   marketv1.MarketplaceStatus(plugin.MarketStatus),
		Visibility:     marketv1.MarketplaceVisibility(plugin.Visibility),
		LatestVersion:  plugin.LatestVersion,
		Icon:           plugin.Icon,
		Homepage:       plugin.Homepage,
		Repository:     plugin.Repository,
		License:        plugin.License,
		Tags:           tags,
		LatestRelease:  latestItem,
		RiskCounts:     riskCounts,
		DownloadCount:  plugin.DownloadCount,
		SourceDelivery: sourceDeliveryForPluginType(marketv1.MarketplacePluginType(plugin.PluginType)),
		PublishedAt:    unixMillisPtr(plugin.PublishedAt),
		UpdatedAt:      unixMillisPtr(plugin.UpdatedAt),
	}
}

// releaseItemFromEntity projects one release row and artifact summary to API DTO.
func releaseItemFromEntity(
	row *entity.PluginMarketplaceRelease,
	artifact *ArtifactRecord,
) *marketv1.MarketplaceReleaseItem {
	if row == nil {
		return nil
	}
	return &marketv1.MarketplaceReleaseItem{
		PluginId:       row.PluginId,
		Version:        row.ReleaseVersion,
		PluginType:     marketv1.MarketplacePluginType(row.PluginType),
		ReleaseStatus:  marketv1.MarketplaceStatus(row.ReleaseStatus),
		ReviewStatus:   marketv1.MarketplaceReviewStatus(row.ReviewStatus),
		Visibility:     marketv1.MarketplaceVisibility(row.Visibility),
		MinHostVersion: row.MinHostVersion,
		MaxHostVersion: row.MaxHostVersion,
		ReviewMessage:  row.ReviewMessage,
		Artifact:       artifactItemFromRecord(artifact),
		SubmittedAt:    unixMillisPtr(row.SubmittedAt),
		ReviewedAt:     unixMillisPtr(row.ReviewedAt),
		PublishedAt:    unixMillisPtr(row.PublishedAt),
		UpdatedAt:      unixMillisPtr(row.UpdatedAt),
	}
}

// artifactItemFromRecord projects one artifact service record to API DTO.
func artifactItemFromRecord(record *ArtifactRecord) *marketv1.MarketplaceArtifactItem {
	if record == nil {
		return nil
	}
	return &marketv1.MarketplaceArtifactItem{
		ArtifactType:   record.ArtifactType,
		FileName:       record.FileName,
		ContentType:    record.ContentType,
		SizeBytes:      record.SizeBytes,
		Sha256:         record.Sha256,
		ManifestSha256: record.ManifestSha256,
		WasmSha256:     record.WasmSha256,
	}
}

// riskItemFromEntity projects one risk row to API DTO.
func riskItemFromEntity(row *entity.PluginMarketplaceRisk) *marketv1.MarketplaceRiskItem {
	if row == nil {
		return nil
	}
	return &marketv1.MarketplaceRiskItem{
		Type:      marketv1.MarketplaceRiskType(row.RiskType),
		Severity:  marketv1.MarketplaceRiskSeverity(row.Severity),
		Source:    row.Source,
		Summary:   row.Summary,
		Payload:   gjson.New(row.Payload).Map(),
		CreatedAt: unixMillisPtr(row.CreatedAt),
	}
}

// documentItemFromRecord projects indexed document metadata to API DTO.
func documentItemFromRecord(record *DocumentRecord) *marketv1.MarketplaceDocumentItem {
	if record == nil {
		return nil
	}
	content := ""
	if record.SearchText != "" {
		// Fallback only: real reads replace this with package-rendered HTML.
		content = "<p>" + html.EscapeString(record.SearchText) + "</p>"
	}
	return &marketv1.MarketplaceDocumentItem{
		PluginId:       record.PluginID,
		Version:        record.Version,
		Locale:         record.RequestedLocale,
		ResolvedLocale: record.ResolvedLocale,
		Path:           record.Path,
		SourceKind:     record.SourceKind,
		Title:          record.Title,
		Summary:        record.Summary,
		Content:        content,
		ContentHash:    record.ContentHash,
		FallbackUsed:   record.FallbackUsed,
		UpdatedAt:      unixMillisPtr(record.UpdatedAt),
	}
}

// loadDocumentRenderedContent reopens the release package and renders the selected document.
func (s *serviceImpl) loadDocumentRenderedContent(
	ctx context.Context,
	record *DocumentRecord,
) (rendered string, err error) {
	if record == nil || s.artifacts == nil {
		return "", nil
	}
	artifact, err := s.selectPackageArtifactForRelease(ctx, record.ReleaseID)
	if err != nil || artifact == nil || normalizeKey(artifact.StorageKey) == "" {
		return "", err
	}
	localPath, err := s.artifacts.LocalPath(ctx, artifact.StorageKey)
	if err != nil {
		return "", err
	}
	zipReader, err := zip.OpenReader(localPath)
	if err != nil {
		return "", bizerr.WrapCode(err, CodeMarketplaceStorageFailed)
	}
	defer func() {
		if closeErr := zipReader.Close(); closeErr != nil && err == nil {
			err = bizerr.WrapCode(closeErr, CodeMarketplaceStorageFailed)
		}
	}()
	fileIndex, err := indexSourceZipFiles(zipReader.File)
	if err != nil {
		return "", err
	}
	rootPrefix := detectPackageRootPrefix(fileIndex)
	relativePath := documentRelativePackagePath(record, rootPrefix)
	file := fileIndex[relativePath]
	if file == nil {
		// Try without root prefix when package is single-rooted but index keys include root.
		file = fileIndex[strings.TrimPrefix(relativePath, rootPrefix)]
	}
	if file == nil {
		return "", nil
	}
	raw, err := readZipFile(file)
	if err != nil {
		return "", bizerr.WrapCode(err, CodeMarketplacePackageScanFailed)
	}
	item, err := indexMarketplaceDocument(record.ResolvedLocale, record.Path, record.SourceKind, string(raw))
	if err != nil {
		return "", err
	}
	return item.RenderedContent, nil
}

// selectPackageArtifactForRelease returns the primary package artifact for one release.
func (s *serviceImpl) selectPackageArtifactForRelease(
	ctx context.Context,
	releaseID int,
) (*ArtifactRecord, error) {
	if releaseID <= 0 {
		return nil, nil
	}
	for _, artifactType := range []string{
		marketv1.MarketplaceArtifactTypeSourceZip.String(),
		marketv1.MarketplaceArtifactTypeDynamicZip.String(),
	} {
		row, err := s.getArtifactByReleaseType(ctx, releaseID, artifactType)
		if err != nil {
			return nil, err
		}
		if row != nil {
			return artifactRecordFromEntity(row), nil
		}
	}
	return nil, nil
}

// documentRelativePackagePath rebuilds the package-relative path for one indexed document.
func documentRelativePackagePath(record *DocumentRecord, rootPrefix string) string {
	if record == nil {
		return ""
	}
	switch record.SourceKind {
	case documentSourceKindReadme:
		return rootPrefix + record.Path
	default:
		if looksLikeLocale(record.ResolvedLocale) && record.ResolvedLocale != defaultDocumentLocale {
			return rootPrefix + marketplaceDocsPrefix + record.ResolvedLocale + "/" + record.Path
		}
		return rootPrefix + marketplaceDocsPrefix + record.Path
	}
}

// detectPackageRootPrefix returns the single plugin root prefix when present.
func detectPackageRootPrefix(files map[string]*zip.File) string {
	if _, ok := files[sourcePackageManifestPath]; ok {
		return ""
	}
	root := ""
	for filePath := range files {
		segments := strings.Split(filePath, "/")
		if len(segments) < 2 {
			return ""
		}
		if root == "" {
			root = segments[0]
			continue
		}
		if root != segments[0] {
			return ""
		}
	}
	if root == "" {
		return ""
	}
	return root + "/"
}

// publisherItemFromEntity projects a publisher row with read-model fallback fields.
func publisherItemFromEntity(
	row *entity.PluginMarketplacePublisher,
	fallbackName string,
	fallbackVerified bool,
) *marketv1.MarketplacePublisherItem {
	if row == nil {
		return &marketv1.MarketplacePublisherItem{
			Name:     fallbackName,
			Verified: fallbackVerified,
		}
	}
	return &marketv1.MarketplacePublisherItem{
		PublisherKey: row.PublisherKey,
		Name:         row.Name,
		Summary:      row.Summary,
		Verified:     row.Verified,
		Homepage:     row.Homepage,
	}
}

// sourceDeliveryForPluginType returns the marketplace delivery guidance token.
func sourceDeliveryForPluginType(pluginType marketv1.MarketplacePluginType) string {
	if pluginType == marketv1.MarketplacePluginTypeDynamic {
		return sourceDeliveryDynamicUploadRequired
	}
	return sourceDeliverySourceRebuildRequired
}

// normalizeMarketplacePage applies default and maximum page bounds.
func normalizeMarketplacePage(pageNum int, pageSize int) (int, int) {
	if pageNum < defaultMarketplacePageNum {
		pageNum = defaultMarketplacePageNum
	}
	if pageSize <= 0 {
		pageSize = defaultMarketplacePageSize
	}
	if pageSize > maxMarketplacePageSize {
		pageSize = maxMarketplacePageSize
	}
	return pageNum, pageSize
}

// decodeStringArrayJSON converts a JSON array snapshot into stable strings.
func decodeStringArrayJSON(value string) []string {
	var items []string
	if err := json.Unmarshal([]byte(defaultJSONString(value, defaultCollectionSummary)), &items); err != nil {
		return []string{}
	}
	return items
}

// decodeRiskCounts converts a risk-count JSON snapshot into API DTO counts.
func decodeRiskCounts(value string) marketv1.MarketplaceRiskCounts {
	counts := marketv1.MarketplaceRiskCounts{}
	if err := json.Unmarshal([]byte(defaultJSONString(value, defaultObjectSummary)), &counts); err != nil {
		return marketv1.MarketplaceRiskCounts{}
	}
	return counts
}

// unixMillisPtr converts an internal time pointer to an API Unix millisecond pointer.
func unixMillisPtr(value *time.Time) *int64 {
	if value == nil {
		return nil
	}
	millis := value.UnixMilli()
	return &millis
}

// releaseIDsFromEntities returns unique release IDs from entity rows.
func releaseIDsFromEntities(rows []*entity.PluginMarketplaceRelease) []int {
	ids := make([]int, 0, len(rows))
	for _, row := range rows {
		if row == nil || row.Id <= 0 {
			continue
		}
		ids = append(ids, row.Id)
	}
	return uniqueInts(ids)
}

// publisherIDsFromReadModels returns unique publisher IDs from read-model rows.
func publisherIDsFromReadModels(rows []*entity.PluginMarketplacePluginReadModel) []int {
	ids := make([]int, 0, len(rows))
	for _, row := range rows {
		if row == nil || row.PublisherId <= 0 {
			continue
		}
		ids = append(ids, row.PublisherId)
	}
	return uniqueInts(ids)
}

// artifactPriority chooses the primary artifact for release display and downloads.
func artifactPriority(row *entity.PluginMarketplaceArtifact, pluginType marketv1.MarketplacePluginType) int {
	if row == nil {
		return 100
	}
	artifactType := marketv1.MarketplaceArtifactType(row.ArtifactType)
	if pluginType == marketv1.MarketplacePluginTypeDynamic {
		switch artifactType {
		case marketv1.MarketplaceArtifactTypeDynamicZip:
			return 1
		case marketv1.MarketplaceArtifactTypePluginWasm:
			return 2
		default:
			return 10
		}
	}
	if artifactType == marketv1.MarketplaceArtifactTypeSourceZip {
		return 1
	}
	return 10
}

// marketplaceSearchText builds the DB-side fuzzy-search projection.
func marketplaceSearchText(
	plugin *entity.PluginMarketplacePlugin,
	publisher *PublisherRecord,
	release *entity.PluginMarketplaceRelease,
	tagCodes []string,
) string {
	parts := make([]string, 0, 8+len(tagCodes))
	if plugin != nil {
		parts = append(parts, plugin.PluginId, plugin.Name, plugin.Summary, plugin.Description)
	}
	if publisher != nil {
		parts = append(parts, publisher.PublisherKey, publisher.Name, publisher.Summary)
	}
	if release != nil {
		parts = append(parts, release.ReleaseVersion, release.MinHostVersion, release.MaxHostVersion)
	}
	parts = append(parts, tagCodes...)
	return strings.Join(nonEmptyStrings(parts), " ")
}

// firstString returns the first non-empty string in values.
func firstString(values []string) string {
	for _, value := range values {
		if normalizeKey(value) != "" {
			return value
		}
	}
	return ""
}

// nonEmptyStrings filters blank values.
func nonEmptyStrings(values []string) []string {
	out := make([]string, 0, len(values))
	for _, value := range values {
		if normalizeKey(value) == "" {
			continue
		}
		out = append(out, normalizeKey(value))
	}
	return out
}

// uniqueInts returns unique positive IDs in first-seen order.
func uniqueInts(values []int) []int {
	seen := make(map[int]struct{}, len(values))
	out := make([]int, 0, len(values))
	for _, value := range values {
		if value <= 0 {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	return out
}

// uniqueStrings returns unique non-empty strings in first-seen order.
func uniqueStrings(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	out := make([]string, 0, len(values))
	for _, value := range values {
		normalized := normalizeKey(value)
		if normalized == "" {
			continue
		}
		if _, ok := seen[normalized]; ok {
			continue
		}
		seen[normalized] = struct{}{}
		out = append(out, normalized)
	}
	return out
}

// packageDiagnosticRiskPayload stores a stable scanner code in risk rows.
type packageDiagnosticRiskPayload struct {
	Code string `json:"code"`
}

// diagnosticRiskType classifies package diagnostics into review risk types.
func diagnosticRiskType(diagnostic *PackageDiagnostic) marketv1.MarketplaceRiskType {
	if diagnostic == nil {
		return marketv1.MarketplaceRiskTypeDocs
	}
	code := normalizeKey(diagnostic.Code)
	source := normalizeKey(diagnostic.Source)
	switch {
	case strings.Contains(code, "host_services") || strings.Contains(source, "hostServices"):
		return marketv1.MarketplaceRiskTypeHostService
	case strings.Contains(code, "routes"):
		return marketv1.MarketplaceRiskTypeDynamicRoute
	case strings.Contains(code, "mock_sql"):
		return marketv1.MarketplaceRiskTypeMockSQL
	case strings.Contains(code, "sql"):
		return marketv1.MarketplaceRiskTypeInstallSQL
	case strings.Contains(code, "dependency"):
		return marketv1.MarketplaceRiskTypeDependency
	case strings.Contains(code, "multi_tenant"):
		return marketv1.MarketplaceRiskTypeMultiTenant
	default:
		return marketv1.MarketplaceRiskTypeDocs
	}
}
