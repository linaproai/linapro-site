// This file implements marketplace catalog, publish, review, document, risk,
// and download HTTP handlers. Controllers stay thin and project service outputs
// to API DTOs without embedding marketplace domain rules.

package market

import (
	"context"
	"strings"

	"github.com/gogf/gf/v2/frame/g"

	"lina-core/pkg/logger"
	marketv1 "linapro-plugin-marketplace/backend/api/market/v1"
	marketplacesvc "linapro-plugin-marketplace/backend/internal/service/marketplace"
)

// PluginList returns the visibility-filtered marketplace catalog page.
func (c *ControllerV1) PluginList(ctx context.Context, req *marketv1.PluginListReq) (*marketv1.PluginListRes, error) {
	out, err := c.marketSvc.ListPlugins(ctx, marketplacesvc.ListPluginsInput{
		PageNum:     req.PageNum,
		PageSize:    req.PageSize,
		Keyword:     req.Keyword,
		PluginType:  req.PluginType,
		TagCode:     req.TagCode,
		Publisher:   req.Publisher,
		HostVersion: req.HostVersion,
		Visibility:  c.currentVisibilitySubject(ctx),
		Locale:      resolveDocumentLocale(ctx, ""),
	})
	if err != nil {
		return nil, err
	}
	return &marketv1.PluginListRes{List: out.List, Total: out.Total}, nil
}

// MyPluginList returns marketplace plugins owned by the current publisher user.
func (c *ControllerV1) MyPluginList(ctx context.Context, req *marketv1.MyPluginListReq) (*marketv1.MyPluginListRes, error) {
	userID, err := c.requireUserID(ctx)
	if err != nil {
		return nil, err
	}
	out, err := c.marketSvc.ListOwnedPlugins(ctx, marketplacesvc.ListOwnedPluginsInput{
		PageNum:     req.PageNum,
		PageSize:    req.PageSize,
		Keyword:     req.Keyword,
		PluginType:  req.PluginType,
		Status:      req.Status,
		OwnerUserID: userID,
		Locale:      resolveDocumentLocale(ctx, ""),
	})
	if err != nil {
		return nil, err
	}
	return &marketv1.MyPluginListRes{List: out.List, Total: out.Total}, nil
}

// ManagedPluginList returns all marketplace plugins for review operators.
func (c *ControllerV1) ManagedPluginList(
	ctx context.Context,
	req *marketv1.ManagedPluginListReq,
) (*marketv1.ManagedPluginListRes, error) {
	out, err := c.marketSvc.ListManagedPlugins(ctx, marketplacesvc.ListManagedPluginsInput{
		PageNum:    req.PageNum,
		PageSize:   req.PageSize,
		Keyword:    req.Keyword,
		PluginType: req.PluginType,
		Status:     req.Status,
		Publisher:  req.Publisher,
		Locale:     resolveDocumentLocale(ctx, ""),
	})
	if err != nil {
		return nil, err
	}
	return &marketv1.ManagedPluginListRes{List: out.List, Total: out.Total}, nil
}

// ReviewQueueList returns marketplace releases awaiting review.
func (c *ControllerV1) ReviewQueueList(
	ctx context.Context,
	req *marketv1.ReviewQueueListReq,
) (*marketv1.ReviewQueueListRes, error) {
	out, err := c.marketSvc.ListReviewQueue(ctx, marketplacesvc.ListReviewQueueInput{
		PageNum:      req.PageNum,
		PageSize:     req.PageSize,
		PluginID:     req.PluginId,
		ReviewStatus: req.ReviewStatus,
		Keyword:      req.Keyword,
		Locale:       resolveDocumentLocale(ctx, ""),
	})
	if err != nil {
		return nil, err
	}
	return &marketv1.ReviewQueueListRes{List: out.List, Total: out.Total}, nil
}

// PluginDetail returns one marketplace plugin detail projection.
func (c *ControllerV1) PluginDetail(ctx context.Context, req *marketv1.PluginDetailReq) (*marketv1.PluginDetailRes, error) {
	out, err := c.marketSvc.GetPluginDetail(ctx, marketplacesvc.GetPluginDetailInput{
		PluginID:   req.PluginId,
		Visibility: c.currentVisibilitySubject(ctx),
		Locale:     resolveDocumentLocale(ctx, ""),
	})
	if err != nil {
		return nil, err
	}
	return &marketv1.PluginDetailRes{Plugin: out.Plugin}, nil
}

// MyPluginDetail returns one detail projection within publisher ownership.
func (c *ControllerV1) MyPluginDetail(ctx context.Context, req *marketv1.MyPluginDetailReq) (*marketv1.MyPluginDetailRes, error) {
	out, err := c.marketSvc.GetPluginDetail(ctx, marketplacesvc.GetPluginDetailInput{
		PluginID:   req.PluginId,
		Visibility: c.currentPublisherVisibilitySubject(ctx),
		Locale:     resolveDocumentLocale(ctx, ""),
	})
	if err != nil {
		return nil, err
	}
	return &marketv1.MyPluginDetailRes{Plugin: out.Plugin}, nil
}

// ManagedPluginDetail returns one detail projection for a review operator.
func (c *ControllerV1) ManagedPluginDetail(ctx context.Context, req *marketv1.ManagedPluginDetailReq) (*marketv1.ManagedPluginDetailRes, error) {
	out, err := c.marketSvc.GetPluginDetail(ctx, marketplacesvc.GetPluginDetailInput{
		PluginID:   req.PluginId,
		Visibility: c.currentReviewerVisibilitySubject(ctx),
		Locale:     resolveDocumentLocale(ctx, ""),
	})
	if err != nil {
		return nil, err
	}
	return &marketv1.ManagedPluginDetailRes{Plugin: out.Plugin}, nil
}

// ReleaseList returns paginated releases for one marketplace plugin.
func (c *ControllerV1) ReleaseList(ctx context.Context, req *marketv1.ReleaseListReq) (*marketv1.ReleaseListRes, error) {
	out, err := c.marketSvc.ListReleases(ctx, marketplacesvc.ListReleasesInput{
		PluginID:     req.PluginId,
		PageNum:      req.PageNum,
		PageSize:     req.PageSize,
		Status:       req.Status,
		ReviewStatus: req.ReviewStatus,
		Visibility:   c.currentVisibilitySubject(ctx),
	})
	if err != nil {
		return nil, err
	}
	return &marketv1.ReleaseListRes{List: out.List, Total: out.Total}, nil
}

// MyReleaseList returns releases within publisher ownership.
func (c *ControllerV1) MyReleaseList(ctx context.Context, req *marketv1.MyReleaseListReq) (*marketv1.MyReleaseListRes, error) {
	out, err := c.marketSvc.ListReleases(ctx, marketplacesvc.ListReleasesInput{
		PluginID:     req.PluginId,
		PageNum:      req.PageNum,
		PageSize:     req.PageSize,
		Status:       req.Status,
		ReviewStatus: req.ReviewStatus,
		Visibility:   c.currentPublisherVisibilitySubject(ctx),
	})
	if err != nil {
		return nil, err
	}
	return &marketv1.MyReleaseListRes{List: out.List, Total: out.Total}, nil
}

// ManagedReleaseList returns releases for a review operator.
func (c *ControllerV1) ManagedReleaseList(ctx context.Context, req *marketv1.ManagedReleaseListReq) (*marketv1.ManagedReleaseListRes, error) {
	out, err := c.marketSvc.ListReleases(ctx, marketplacesvc.ListReleasesInput{
		PluginID:     req.PluginId,
		PageNum:      req.PageNum,
		PageSize:     req.PageSize,
		Status:       req.Status,
		ReviewStatus: req.ReviewStatus,
		Visibility:   c.currentReviewerVisibilitySubject(ctx),
	})
	if err != nil {
		return nil, err
	}
	return &marketv1.ManagedReleaseListRes{List: out.List, Total: out.Total}, nil
}

// ReleaseDocs returns one version-scoped marketplace document.
func (c *ControllerV1) ReleaseDocs(ctx context.Context, req *marketv1.ReleaseDocsReq) (*marketv1.ReleaseDocsRes, error) {
	out, err := c.marketSvc.GetReleaseDocument(ctx, marketplacesvc.GetReleaseDocumentInput{
		PluginID:   req.PluginId,
		Version:    req.Version,
		Locale:     resolveDocumentLocale(ctx, req.Locale),
		Path:       req.Path,
		Visibility: c.currentVisibilitySubject(ctx),
	})
	if err != nil {
		return nil, err
	}
	return &marketv1.ReleaseDocsRes{Document: out.Document, Documents: out.Documents}, nil
}

// MyReleaseDocs returns documentation within publisher ownership.
func (c *ControllerV1) MyReleaseDocs(ctx context.Context, req *marketv1.MyReleaseDocsReq) (*marketv1.MyReleaseDocsRes, error) {
	out, err := c.marketSvc.GetReleaseDocument(ctx, marketplacesvc.GetReleaseDocumentInput{
		PluginID:   req.PluginId,
		Version:    req.Version,
		Locale:     resolveDocumentLocale(ctx, req.Locale),
		Path:       req.Path,
		Visibility: c.currentPublisherVisibilitySubject(ctx),
	})
	if err != nil {
		return nil, err
	}
	return &marketv1.MyReleaseDocsRes{Document: out.Document, Documents: out.Documents}, nil
}

// ManagedReleaseDocs returns documentation for a review operator.
func (c *ControllerV1) ManagedReleaseDocs(ctx context.Context, req *marketv1.ManagedReleaseDocsReq) (*marketv1.ManagedReleaseDocsRes, error) {
	out, err := c.marketSvc.GetReleaseDocument(ctx, marketplacesvc.GetReleaseDocumentInput{
		PluginID:   req.PluginId,
		Version:    req.Version,
		Locale:     resolveDocumentLocale(ctx, req.Locale),
		Path:       req.Path,
		Visibility: c.currentReviewerVisibilitySubject(ctx),
	})
	if err != nil {
		return nil, err
	}
	return &marketv1.ManagedReleaseDocsRes{Document: out.Document, Documents: out.Documents}, nil
}

// ReleaseRisks returns paginated scanner risk findings for one release.
func (c *ControllerV1) ReleaseRisks(ctx context.Context, req *marketv1.ReleaseRisksReq) (*marketv1.ReleaseRisksRes, error) {
	out, err := c.marketSvc.ListReleaseRisks(ctx, marketplacesvc.ListReleaseRisksInput{
		PluginID:   req.PluginId,
		Version:    req.Version,
		PageNum:    req.PageNum,
		PageSize:   req.PageSize,
		Type:       req.Type,
		Severity:   req.Severity,
		Visibility: c.currentVisibilitySubject(ctx),
	})
	if err != nil {
		return nil, err
	}
	return &marketv1.ReleaseRisksRes{List: out.List, Total: out.Total}, nil
}

// MyReleaseRisks returns risk findings within publisher ownership.
func (c *ControllerV1) MyReleaseRisks(ctx context.Context, req *marketv1.MyReleaseRisksReq) (*marketv1.MyReleaseRisksRes, error) {
	out, err := c.marketSvc.ListReleaseRisks(ctx, marketplacesvc.ListReleaseRisksInput{
		PluginID:   req.PluginId,
		Version:    req.Version,
		PageNum:    req.PageNum,
		PageSize:   req.PageSize,
		Type:       req.Type,
		Severity:   req.Severity,
		Visibility: c.currentPublisherVisibilitySubject(ctx),
	})
	if err != nil {
		return nil, err
	}
	return &marketv1.MyReleaseRisksRes{List: out.List, Total: out.Total}, nil
}

// ManagedReleaseRisks returns risk findings for a review operator.
func (c *ControllerV1) ManagedReleaseRisks(ctx context.Context, req *marketv1.ManagedReleaseRisksReq) (*marketv1.ManagedReleaseRisksRes, error) {
	out, err := c.marketSvc.ListReleaseRisks(ctx, marketplacesvc.ListReleaseRisksInput{
		PluginID:   req.PluginId,
		Version:    req.Version,
		PageNum:    req.PageNum,
		PageSize:   req.PageSize,
		Type:       req.Type,
		Severity:   req.Severity,
		Visibility: c.currentReviewerVisibilitySubject(ctx),
	})
	if err != nil {
		return nil, err
	}
	return &marketv1.ManagedReleaseRisksRes{List: out.List, Total: out.Total}, nil
}

// DownloadSessionCreate creates one short-lived download session.
func (c *ControllerV1) DownloadSessionCreate(
	ctx context.Context,
	req *marketv1.DownloadSessionCreateReq,
) (*marketv1.DownloadSessionCreateRes, error) {
	userID, err := c.requireUserID(ctx)
	if err != nil {
		return nil, err
	}
	out, err := c.marketSvc.CreateDownloadSession(ctx, marketplacesvc.CreateDownloadSessionInput{
		PluginID:        req.PluginId,
		Version:         req.Version,
		ArtifactType:    req.ArtifactType,
		RequesterUserID: userID,
		Visibility:      c.currentVisibilitySubject(ctx),
	})
	if err != nil {
		return nil, err
	}
	return &marketv1.DownloadSessionCreateRes{Session: out.Session}, nil
}

// DownloadSessionGet returns one requester-owned download session.
func (c *ControllerV1) DownloadSessionGet(
	ctx context.Context,
	req *marketv1.DownloadSessionGetReq,
) (*marketv1.DownloadSessionGetRes, error) {
	userID, err := c.requireUserID(ctx)
	if err != nil {
		return nil, err
	}
	out, err := c.marketSvc.GetDownloadSession(ctx, marketplacesvc.GetDownloadSessionInput{
		SessionID:       req.SessionId,
		RequesterUserID: userID,
		Visibility:      c.currentVisibilitySubject(ctx),
	})
	if err != nil {
		return nil, err
	}
	return &marketv1.DownloadSessionGetRes{Session: out.Session}, nil
}

// DownloadSessionContent streams artifact bytes for one download session.
func (c *ControllerV1) DownloadSessionContent(
	ctx context.Context,
	req *marketv1.DownloadSessionContentReq,
) (*marketv1.DownloadSessionContentRes, error) {
	userID, err := c.requireUserID(ctx)
	if err != nil {
		return nil, err
	}
	out, err := c.marketSvc.OpenDownloadContent(ctx, marketplacesvc.OpenDownloadContentInput{
		SessionID:       req.SessionId,
		RequesterUserID: userID,
		Visibility:      c.currentVisibilitySubject(ctx),
	})
	if err != nil {
		return nil, err
	}
	request := g.RequestFromCtx(ctx)
	sizeBytes := int64(0)
	if out.Session != nil {
		sizeBytes = out.Session.SizeBytes
	}
	if err = writeDownloadStream(ctx, request, out.FileName, sizeBytes, out.Body); err != nil {
		recordErr := c.marketSvc.RecordDownloadEvent(ctx, marketplacesvc.RecordDownloadEventInput{
			SessionID:       req.SessionId,
			EventType:       marketplacesvc.DownloadEventTypeFailed,
			RequesterUserID: userID,
			Visibility:      c.currentVisibilitySubject(ctx),
		})
		if recordErr != nil {
			logger.Error(ctx, "record marketplace download failure event:", recordErr)
		}
		return nil, err
	}
	recordErr := c.marketSvc.RecordDownloadEvent(ctx, marketplacesvc.RecordDownloadEventInput{
		SessionID:       req.SessionId,
		EventType:       marketplacesvc.DownloadEventTypeCompleted,
		RequesterUserID: userID,
		Visibility:      c.currentVisibilitySubject(ctx),
	})
	if recordErr != nil {
		logger.Error(ctx, "record marketplace download completion event:", recordErr)
	}
	return &marketv1.DownloadSessionContentRes{}, nil
}

// PublisherList returns publisher profiles available to the current operator.
func (c *ControllerV1) PublisherList(
	ctx context.Context,
	req *marketv1.PublisherListReq,
) (*marketv1.PublisherListRes, error) {
	userID, err := c.requireUserID(ctx)
	if err != nil {
		return nil, err
	}
	out, err := c.marketSvc.ListPublishers(ctx, marketplacesvc.ListPublishersInput{
		PageNum:     req.PageNum,
		PageSize:    req.PageSize,
		Keyword:     req.Keyword,
		OwnerUserID: userID,
	})
	if err != nil {
		return nil, err
	}
	return &marketv1.PublisherListRes{List: out.List, Total: out.Total}, nil
}

// PublisherCreate creates one marketplace publisher profile.
func (c *ControllerV1) PublisherCreate(
	ctx context.Context,
	req *marketv1.PublisherCreateReq,
) (*marketv1.PublisherCreateRes, error) {
	userID, err := c.requireUserID(ctx)
	if err != nil {
		return nil, err
	}
	record, err := c.marketSvc.CreatePublisher(ctx, marketplacesvc.CreatePublisherInput{
		PublisherKey: req.PublisherKey,
		Name:         req.Name,
		Summary:      req.Summary,
		Homepage:     req.Homepage,
		ContactEmail: req.ContactEmail,
		OwnerUserID:  userID,
	})
	if err != nil {
		return nil, err
	}
	return &marketv1.PublisherCreateRes{Publisher: publisherItemFromRecord(record)}, nil
}

// PublisherUpdate updates one marketplace publisher profile owned by the current operator.
func (c *ControllerV1) PublisherUpdate(
	ctx context.Context,
	req *marketv1.PublisherUpdateReq,
) (*marketv1.PublisherUpdateRes, error) {
	userID, err := c.requireUserID(ctx)
	if err != nil {
		return nil, err
	}
	currentKey := req.PathPublisherKey
	if currentKey == "" {
		currentKey = req.PublisherKey
	}
	record, err := c.marketSvc.UpdatePublisher(ctx, marketplacesvc.UpdatePublisherInput{
		CurrentPublisherKey: currentKey,
		PublisherKey:        req.PublisherKey,
		Name:                req.Name,
		Summary:             req.Summary,
		Homepage:            req.Homepage,
		ContactEmail:        req.ContactEmail,
		OwnerUserID:         userID,
	})
	if err != nil {
		return nil, err
	}
	return &marketv1.PublisherUpdateRes{Publisher: publisherItemFromRecord(record)}, nil
}

// PluginCreate creates or updates one marketplace plugin identity draft.
func (c *ControllerV1) PluginCreate(
	ctx context.Context,
	req *marketv1.PluginCreateReq,
) (*marketv1.PluginCreateRes, error) {
	userID, err := c.requireUserID(ctx)
	if err != nil {
		return nil, err
	}
	record, err := c.marketSvc.SavePluginDraft(ctx, marketplacesvc.SavePluginDraftInput{
		PublisherKey: req.PublisherKey,
		OwnerUserID:  userID,
		PluginID:     req.PluginId,
		Name:         req.Name,
		Summary:      req.Summary,
		Description:  req.Description,
		PluginType:   req.PluginType,
		Visibility:   req.Visibility,
		Icon:         req.Icon,
		Homepage:     req.Homepage,
		Repository:   req.Repository,
		License:      req.License,
	})
	if err != nil {
		return nil, err
	}
	return &marketv1.PluginCreateRes{Plugin: pluginDetailFromRecord(record)}, nil
}

// ReleaseUpload uploads one marketplace package draft.
func (c *ControllerV1) ReleaseUpload(
	ctx context.Context,
	req *marketv1.ReleaseUploadReq,
) (*marketv1.ReleaseUploadRes, error) {
	userID, err := c.requireUserID(ctx)
	if err != nil {
		return nil, err
	}
	publisherKey := ""
	if request := g.RequestFromCtx(ctx); request != nil {
		publisherKey = strings.TrimSpace(request.GetForm("publisherKey").String())
	}
	localPath, fileName, contentType, cleanup, err := saveUploadToTempFile(ctx, "file")
	if err != nil {
		return nil, err
	}
	defer cleanup()

	var (
		release  *marketplacesvc.ReleaseRecord
		artifact *marketplacesvc.ArtifactRecord
	)
	switch req.PluginType {
	case marketv1.MarketplacePluginTypeDynamic:
		result, uploadErr := c.marketSvc.UploadDynamicPackage(ctx, marketplacesvc.UploadDynamicPackageInput{
			PublisherKey:   publisherKey,
			OwnerUserID:    userID,
			PluginID:       req.PluginId,
			Version:        req.Version,
			PackagePath:    localPath,
			FileName:       fileName,
			ContentType:    contentType,
			Visibility:     req.Visibility,
			MinHostVersion: req.MinHostVersion,
			MaxHostVersion: req.MaxHostVersion,
			ReplaceDraft:   req.ReplaceDraft,
		})
		if uploadErr != nil {
			return nil, uploadErr
		}
		release = result.Release
		artifact = result.PackageArtifact
	default:
		result, uploadErr := c.marketSvc.UploadSourcePackage(ctx, marketplacesvc.UploadSourcePackageInput{
			PublisherKey:   publisherKey,
			OwnerUserID:    userID,
			PluginID:       req.PluginId,
			Version:        req.Version,
			PackagePath:    localPath,
			FileName:       fileName,
			ContentType:    contentType,
			Visibility:     req.Visibility,
			MinHostVersion: req.MinHostVersion,
			MaxHostVersion: req.MaxHostVersion,
			ReplaceDraft:   req.ReplaceDraft,
		})
		if uploadErr != nil {
			return nil, uploadErr
		}
		release = result.Release
		artifact = result.Artifact
	}
	return &marketv1.ReleaseUploadRes{Release: releaseItemFromRecords(release, artifact)}, nil
}

// ReleaseSubmitReview submits one draft release for review.
func (c *ControllerV1) ReleaseSubmitReview(
	ctx context.Context,
	req *marketv1.ReleaseSubmitReviewReq,
) (*marketv1.ReleaseSubmitReviewRes, error) {
	userID, err := c.requireUserID(ctx)
	if err != nil {
		return nil, err
	}
	release, err := c.marketSvc.SubmitReleaseReview(ctx, marketplacesvc.SubmitReleaseReviewInput{
		OwnerUserID: userID,
		PluginID:    req.PluginId,
		Version:     req.Version,
		Message:     req.Message,
	})
	if err != nil {
		return nil, err
	}
	return &marketv1.ReleaseSubmitReviewRes{Release: releaseItemFromRecords(release, nil)}, nil
}

// ReleaseReview approves or rejects one submitted release.
func (c *ControllerV1) ReleaseReview(
	ctx context.Context,
	req *marketv1.ReleaseReviewReq,
) (*marketv1.ReleaseReviewRes, error) {
	if _, err := c.requireUserID(ctx); err != nil {
		return nil, err
	}
	release, err := c.marketSvc.ReviewRelease(ctx, marketplacesvc.ReviewReleaseInput{
		PluginID:     req.PluginId,
		Version:      req.Version,
		ReviewStatus: req.ReviewStatus,
		Message:      req.Message,
	})
	if err != nil {
		return nil, err
	}
	return &marketv1.ReleaseReviewRes{Release: releaseItemFromRecords(release, nil)}, nil
}

// PluginStatusUpdate updates marketplace lifecycle status for one plugin.
func (c *ControllerV1) PluginStatusUpdate(
	ctx context.Context,
	req *marketv1.PluginStatusUpdateReq,
) (*marketv1.PluginStatusUpdateRes, error) {
	if _, err := c.requireUserID(ctx); err != nil {
		return nil, err
	}
	record, err := c.marketSvc.UpdatePluginStatus(ctx, marketplacesvc.UpdatePluginStatusInput{
		PluginID: req.PluginId,
		Status:   req.Status,
		Message:  req.Message,
	})
	if err != nil {
		return nil, err
	}
	return &marketv1.PluginStatusUpdateRes{Plugin: pluginDetailFromRecord(record)}, nil
}

// PackageAdd uploads one plugin package and creates a private draft identity.
func (c *ControllerV1) PackageAdd(
	ctx context.Context,
	req *marketv1.PackageAddReq,
) (*marketv1.PackageAddRes, error) {
	userID, err := c.requireUserID(ctx)
	if err != nil {
		return nil, err
	}
	publisherKey := ""
	if request := g.RequestFromCtx(ctx); request != nil {
		publisherKey = strings.TrimSpace(request.GetForm("publisherKey").String())
		if req.PublisherKey == "" {
			req.PublisherKey = publisherKey
		}
		if request.GetForm("replaceDraft").String() != "" {
			req.ReplaceDraft = request.GetForm("replaceDraft").Bool()
		}
	}
	if publisherKey == "" {
		publisherKey = strings.TrimSpace(req.PublisherKey)
	}
	localPath, fileName, contentType, cleanup, err := saveUploadToTempFile(ctx, "file")
	if err != nil {
		return nil, err
	}
	defer cleanup()

	result, err := c.marketSvc.AddPluginPackage(ctx, marketplacesvc.PackageAddInput{
		PublisherKey: publisherKey,
		OwnerUserID:  userID,
		PackagePath:  localPath,
		FileName:     fileName,
		ContentType:  contentType,
		ReplaceDraft: req.ReplaceDraft,
	})
	if err != nil {
		return nil, err
	}
	return &marketv1.PackageAddRes{
		Plugin:  pluginDetailFromRecord(result.Plugin),
		Release: releaseItemFromRecords(result.Release, nil),
	}, nil
}

// PluginPublish submits one owned plugin for marketplace review.
func (c *ControllerV1) PluginPublish(
	ctx context.Context,
	req *marketv1.PluginPublishReq,
) (*marketv1.PluginPublishRes, error) {
	userID, err := c.requireUserID(ctx)
	if err != nil {
		return nil, err
	}
	release, err := c.marketSvc.RequestPluginPublish(ctx, marketplacesvc.RequestPluginPublishInput{
		OwnerUserID: userID,
		PluginID:    req.PluginId,
		Version:     req.Version,
		Message:     req.Message,
	})
	if err != nil {
		return nil, err
	}
	return &marketv1.PluginPublishRes{Release: releaseItemFromRecords(release, nil)}, nil
}

// PluginDelist withdraws one owned published plugin from the public catalog.
func (c *ControllerV1) PluginDelist(
	ctx context.Context,
	req *marketv1.PluginDelistReq,
) (*marketv1.PluginDelistRes, error) {
	userID, err := c.requireUserID(ctx)
	if err != nil {
		return nil, err
	}
	record, err := c.marketSvc.OwnerDelistPlugin(ctx, marketplacesvc.OwnerDelistPluginInput{
		OwnerUserID: userID,
		PluginID:    req.PluginId,
		Message:     req.Message,
	})
	if err != nil {
		return nil, err
	}
	return &marketv1.PluginDelistRes{Plugin: pluginDetailFromRecord(record)}, nil
}

// GitSourceRegister registers one Git-backed repository and discovers plugins.
func (c *ControllerV1) GitSourceRegister(
	ctx context.Context,
	req *marketv1.GitSourceRegisterReq,
) (*marketv1.GitSourceRegisterRes, error) {
	userID, err := c.requireUserID(ctx)
	if err != nil {
		return nil, err
	}
	result, err := c.marketSvc.RegisterGitSource(ctx, marketplacesvc.RegisterGitSourceInput{
		PublisherKey: req.PublisherKey,
		OwnerUserID:  userID,
		RepoURL:      req.RepoUrl,
		AccessToken:  req.AccessToken,
		Visibility:   marketv1.MarketplaceVisibilityPrivate,
		Homepage:     req.Homepage,
		License:      req.License,
	})
	if err != nil {
		return nil, err
	}
	plugins := make([]*marketv1.MarketplacePluginDetailItem, 0)
	if result != nil {
		for _, record := range result.Plugins {
			if item := pluginDetailFromRecord(record); item != nil {
				plugins = append(plugins, item)
			}
		}
	}
	var primary *marketv1.MarketplacePluginDetailItem
	if len(plugins) > 0 {
		primary = plugins[0]
	}
	return &marketv1.GitSourceRegisterRes{Plugin: primary, Plugins: plugins}, nil
}

// GitSourceSync refreshes metadata for one owned Git marketplace plugin.
func (c *ControllerV1) GitSourceSync(
	ctx context.Context,
	req *marketv1.GitSourceSyncReq,
) (*marketv1.GitSourceSyncRes, error) {
	userID, err := c.requireUserID(ctx)
	if err != nil {
		return nil, err
	}
	result, err := c.marketSvc.DiscoverGitMetadata(ctx, marketplacesvc.DiscoverGitMetadataInput{
		PluginID:    req.PluginId,
		OwnerUserID: userID,
	})
	if err != nil {
		return nil, err
	}
	return &marketv1.GitSourceSyncRes{
		Plugin: pluginDetailFromRecord(result.Plugin),
		Synced: result.Synced,
	}, nil
}

// ReleaseDistribution returns CLI install metadata for one visible release.
func (c *ControllerV1) ReleaseDistribution(
	ctx context.Context,
	req *marketv1.ReleaseDistributionReq,
) (*marketv1.ReleaseDistributionRes, error) {
	item, err := c.marketSvc.GetDistribution(ctx, marketplacesvc.GetDistributionInput{
		PluginID:   req.PluginId,
		Version:    req.Version,
		Visibility: c.currentVisibilitySubject(ctx),
	})
	if err != nil {
		return nil, err
	}
	return &marketv1.ReleaseDistributionRes{Distribution: item}, nil
}

// MyReleaseDistribution returns CLI install metadata for one owned release.
func (c *ControllerV1) MyReleaseDistribution(
	ctx context.Context,
	req *marketv1.MyReleaseDistributionReq,
) (*marketv1.MyReleaseDistributionRes, error) {
	item, err := c.marketSvc.GetDistribution(ctx, marketplacesvc.GetDistributionInput{
		PluginID:   req.PluginId,
		Version:    req.Version,
		Visibility: c.currentPublisherVisibilitySubject(ctx),
	})
	if err != nil {
		return nil, err
	}
	return &marketv1.MyReleaseDistributionRes{Distribution: item}, nil
}
