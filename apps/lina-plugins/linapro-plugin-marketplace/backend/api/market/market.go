// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package market

import (
	"context"

	"linapro-plugin-marketplace/backend/api/market/v1"
)

type IMarketV1 interface {
	PluginList(ctx context.Context, req *v1.PluginListReq) (res *v1.PluginListRes, err error)
	MyPluginList(ctx context.Context, req *v1.MyPluginListReq) (res *v1.MyPluginListRes, err error)
	ManagedPluginList(ctx context.Context, req *v1.ManagedPluginListReq) (res *v1.ManagedPluginListRes, err error)
	ReviewQueueList(ctx context.Context, req *v1.ReviewQueueListReq) (res *v1.ReviewQueueListRes, err error)
	PluginDetail(ctx context.Context, req *v1.PluginDetailReq) (res *v1.PluginDetailRes, err error)
	MyPluginDetail(ctx context.Context, req *v1.MyPluginDetailReq) (res *v1.MyPluginDetailRes, err error)
	ManagedPluginDetail(ctx context.Context, req *v1.ManagedPluginDetailReq) (res *v1.ManagedPluginDetailRes, err error)
	ReleaseList(ctx context.Context, req *v1.ReleaseListReq) (res *v1.ReleaseListRes, err error)
	MyReleaseList(ctx context.Context, req *v1.MyReleaseListReq) (res *v1.MyReleaseListRes, err error)
	ManagedReleaseList(ctx context.Context, req *v1.ManagedReleaseListReq) (res *v1.ManagedReleaseListRes, err error)
	ReleaseDocs(ctx context.Context, req *v1.ReleaseDocsReq) (res *v1.ReleaseDocsRes, err error)
	MyReleaseDocs(ctx context.Context, req *v1.MyReleaseDocsReq) (res *v1.MyReleaseDocsRes, err error)
	ManagedReleaseDocs(ctx context.Context, req *v1.ManagedReleaseDocsReq) (res *v1.ManagedReleaseDocsRes, err error)
	ReleaseRisks(ctx context.Context, req *v1.ReleaseRisksReq) (res *v1.ReleaseRisksRes, err error)
	MyReleaseRisks(ctx context.Context, req *v1.MyReleaseRisksReq) (res *v1.MyReleaseRisksRes, err error)
	ManagedReleaseRisks(ctx context.Context, req *v1.ManagedReleaseRisksReq) (res *v1.ManagedReleaseRisksRes, err error)
	DownloadSessionCreate(ctx context.Context, req *v1.DownloadSessionCreateReq) (res *v1.DownloadSessionCreateRes, err error)
	DownloadSessionGet(ctx context.Context, req *v1.DownloadSessionGetReq) (res *v1.DownloadSessionGetRes, err error)
	DownloadSessionContent(ctx context.Context, req *v1.DownloadSessionContentReq) (res *v1.DownloadSessionContentRes, err error)
	PublisherList(ctx context.Context, req *v1.PublisherListReq) (res *v1.PublisherListRes, err error)
	PublisherCreate(ctx context.Context, req *v1.PublisherCreateReq) (res *v1.PublisherCreateRes, err error)
	PublisherUpdate(ctx context.Context, req *v1.PublisherUpdateReq) (res *v1.PublisherUpdateRes, err error)
	PluginCreate(ctx context.Context, req *v1.PluginCreateReq) (res *v1.PluginCreateRes, err error)
	ReleaseUpload(ctx context.Context, req *v1.ReleaseUploadReq) (res *v1.ReleaseUploadRes, err error)
	ReleaseSubmitReview(ctx context.Context, req *v1.ReleaseSubmitReviewReq) (res *v1.ReleaseSubmitReviewRes, err error)
	ReleaseReview(ctx context.Context, req *v1.ReleaseReviewReq) (res *v1.ReleaseReviewRes, err error)
	PluginStatusUpdate(ctx context.Context, req *v1.PluginStatusUpdateReq) (res *v1.PluginStatusUpdateRes, err error)
	GitSourceRegister(ctx context.Context, req *v1.GitSourceRegisterReq) (res *v1.GitSourceRegisterRes, err error)
	GitSourceSync(ctx context.Context, req *v1.GitSourceSyncReq) (res *v1.GitSourceSyncRes, err error)
	ReleaseDistribution(ctx context.Context, req *v1.ReleaseDistributionReq) (res *v1.ReleaseDistributionRes, err error)
	MyReleaseDistribution(ctx context.Context, req *v1.MyReleaseDistributionReq) (res *v1.MyReleaseDistributionRes, err error)
	PackageAdd(ctx context.Context, req *v1.PackageAddReq) (res *v1.PackageAddRes, err error)
	PluginPublish(ctx context.Context, req *v1.PluginPublishReq) (res *v1.PluginPublishRes, err error)
	PluginDelist(ctx context.Context, req *v1.PluginDelistReq) (res *v1.PluginDelistRes, err error)
}
