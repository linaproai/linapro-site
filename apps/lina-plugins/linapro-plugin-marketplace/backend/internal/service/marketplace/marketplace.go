// Package marketplace implements the plugin marketplace domain service. The
// service owns publisher profiles, plugin ID ownership, release mutability,
// review state transitions, package scanning, read-model refresh,
// marketplace visibility filtering, and controlled download sessions.
package marketplace

import (
	"context"

	marketv1 "linapro-plugin-marketplace/backend/api/market/v1"
)

// Service is the marketplace domain boundary used by controllers and later
// package scanning services.
type Service interface {
	// CreatePublisher creates a publisher profile for the current publishing
	// owner. Input zero values are rejected for required identity fields; the
	// returned record is the persisted publisher projection. It returns
	// CodeMarketplacePublisherAlreadyExists when the publisher key is already
	// owned and wraps storage failures with CodeMarketplaceStorageFailed.
	CreatePublisher(ctx context.Context, in CreatePublisherInput) (*PublisherRecord, error)

	// ListPublishers returns publisher profiles available to the current operator.
	// When OwnerUserID is positive, only publishers owned by that user are returned.
	ListPublishers(ctx context.Context, in ListPublishersInput) (*PublisherListOutput, error)

	// SavePluginDraft creates or updates a marketplace plugin identity owned by
	// the input publisher and requires OwnerUserID to own that publisher.
	// Existing plugin IDs owned by another publisher are rejected with
	// CodeMarketplacePluginIDOwned. The returned record is the
	// persisted plugin projection after insert or update. It does not mutate
	// tags, read models, releases, or visibility grants.
	SavePluginDraft(ctx context.Context, in SavePluginDraftInput) (*PluginRecord, error)

	// SaveReleaseDraft creates or replaces a mutable draft release after package
	// scanning has produced summaries and verifies OwnerUserID before mutation.
	// Published, delisted, deprecated, submitted, reviewing, and approved
	// releases are immutable and rejected.
	// The returned record is the persisted draft release projection.
	SaveReleaseDraft(ctx context.Context, in SaveReleaseDraftInput) (*ReleaseRecord, error)

	// UploadSourcePackage validates one source-plugin ZIP package, parses the
	// root plugin.yaml metadata, builds SQL, i18n, docs, dependency, and risk
	// summaries, verifies OwnerUserID before any artifact storage, saves or
	// replaces a mutable release draft, and records source artifact checksum
	// metadata. It returns package validation business errors for malformed
	// packages and storage errors for persistence failures.
	UploadSourcePackage(ctx context.Context, in UploadSourcePackageInput) (*SourcePackageUploadResult, error)

	// UploadDynamicPackage validates one dynamic-runtime ZIP package, parses the
	// root plugin.yaml and embedded plugin.wasm sections, verifies manifest,
	// ABI, hostServices, route, SQL, i18n, and resource consistency, verifies
	// OwnerUserID before any artifact storage, saves or replaces a mutable release
	// draft, and records dynamic ZIP plus extracted plugin.wasm checksums.
	UploadDynamicPackage(ctx context.Context, in UploadDynamicPackageInput) (*DynamicPackageUploadResult, error)

	// ResolveReleaseDocumentIndex returns the indexed marketplace document
	// metadata selected by locale fallback rules. It reads bounded index rows
	// for one release and does not load package content; later document-content
	// endpoints combine this metadata with artifact storage reads.
	ResolveReleaseDocumentIndex(ctx context.Context, in GetReleaseDocumentInput) (*DocumentRecord, error)

	// ListPlugins returns a paginated catalog page from the marketplace read
	// model. It applies fixed database-side filters, reads the current page
	// projection, and batch-loads publisher snapshots without reading package
	// artifacts or Markdown content for each row.
	ListPlugins(ctx context.Context, in ListPluginsInput) (*PluginListOutput, error)

	// ListOwnedPlugins returns marketplace plugins owned by publishers bound to
	// OwnerUserID. Draft and unpublished plugins are included for publisher
	// workbench management.
	ListOwnedPlugins(ctx context.Context, in ListOwnedPluginsInput) (*PluginListOutput, error)

	// ListManagedPlugins returns marketplace plugins across all publishers for
	// review operators, including draft and delisted statuses.
	ListManagedPlugins(ctx context.Context, in ListManagedPluginsInput) (*PluginListOutput, error)

	// ListReviewQueue returns releases awaiting marketplace review with plugin
	// and publisher projections batch-loaded for the current page.
	ListReviewQueue(ctx context.Context, in ListReviewQueueInput) (*ReviewQueueOutput, error)

	// GetPluginDetail returns one marketplace plugin detail projection with
	// latest release, tags, publisher, and risk summary. Invisible and missing
	// plugins are reported with CodeMarketplacePluginNotFound so private plugin
	// existence is not disclosed.
	GetPluginDetail(ctx context.Context, in GetPluginDetailInput) (*PluginDetailOutput, error)

	// ListReleases returns a paginated release page for one marketplace plugin.
	// It performs release filtering and pagination in the database, then
	// batch-loads primary artifact summaries for the current page.
	ListReleases(ctx context.Context, in ListReleasesInput) (*ReleaseListOutput, error)

	// ListReleaseRisks returns paginated scanner risk findings for one visible
	// release. Filters are pushed into the risk table query and the response
	// only reads risk rows for the requested release.
	ListReleaseRisks(ctx context.Context, in ListReleaseRisksInput) (*RiskListOutput, error)

	// GetReleaseDocument returns the selected document projection for one
	// release. The current implementation uses indexed safe metadata and search
	// text; durable package-content reads are introduced when storage-backed
	// download/content access is wired.
	GetReleaseDocument(ctx context.Context, in GetReleaseDocumentInput) (*DocumentOutput, error)

	// CreateDownloadSession creates a short-lived authorization session for one
	// visible downloadable release artifact. It binds the requester, artifact
	// checksum, expiration, and authorization snapshot before writing a created
	// download event.
	CreateDownloadSession(ctx context.Context, in CreateDownloadSessionInput) (*DownloadSessionOutput, error)

	// GetDownloadSession returns one requester-owned download session after
	// validating visibility, requester ownership, status, and expiration. It
	// rejects expired, revoked, missing, or unauthorized sessions without
	// returning artifact content.
	GetDownloadSession(ctx context.Context, in GetDownloadSessionInput) (*DownloadSessionOutput, error)

	// RecordDownloadEvent writes a controlled download event for an existing
	// requester-owned session. Completed events mark the session consumed and
	// later statistics refresh reads the event table instead of list queries.
	RecordDownloadEvent(ctx context.Context, in RecordDownloadEventInput) error

	// RefreshDownloadStatistics rebuilds plugin download-count snapshots from
	// completed download events. It is intended for asynchronous workers or
	// explicit maintenance calls, not ordinary list, detail, or document reads.
	RefreshDownloadStatistics(ctx context.Context, in RefreshDownloadStatisticsInput) error

	// OpenDownloadContent validates one requester-owned download session and
	// opens the bound artifact bytes for streaming. Callers must close the
	// returned reader. Completed downloads should record a completed event.
	OpenDownloadContent(ctx context.Context, in OpenDownloadContentInput) (*OpenDownloadContentOutput, error)

	// SubmitReleaseReview moves a mutable draft or rejected release to submitted
	// review state. It verifies that OwnerUserID owns the plugin before changing
	// state and returns the updated release projection. Already submitted, reviewing,
	// approved, published, delisted, or deprecated releases are rejected.
	SubmitReleaseReview(ctx context.Context, in SubmitReleaseReviewInput) (*ReleaseRecord, error)

	// ReviewRelease approves or rejects a submitted release. Approval publishes
	// the immutable release and updates the owning plugin latest-version anchor
	// in the same database transaction. Rejection records the reviewer message
	// and returns the release to draft lifecycle status with rejected review
	// state.
	ReviewRelease(ctx context.Context, in ReviewReleaseInput) (*ReleaseRecord, error)

	// UpdatePluginStatus updates marketplace lifecycle status for a plugin that
	// has a latest published release. It is used for reviewer-driven delist,
	// deprecate, or relist actions and mirrors the status to the latest release
	// when one exists. The returned record is the updated plugin projection.
	UpdatePluginStatus(ctx context.Context, in UpdatePluginStatusInput) (*PluginRecord, error)

	// RegisterGitSource registers one GitHub/Gitee repository as a marketplace
	// plugin, stores optional encrypted credentials, and immediately discovers
	// version tags as metadata without cloning full source trees.
	RegisterGitSource(ctx context.Context, in RegisterGitSourceInput) (*PluginRecord, error)

	// DiscoverGitMetadata refreshes remote tags and draft releases for one
	// Git-backed marketplace plugin.
	DiscoverGitMetadata(ctx context.Context, in DiscoverGitMetadataInput) (*DiscoverGitMetadataResult, error)

	// DiscoverAllGitSources scans every Git-backed marketplace plugin for new tags.
	DiscoverAllGitSources(ctx context.Context) (int, error)

	// GetDistribution returns CLI install metadata for one visible release.
	GetDistribution(ctx context.Context, in GetDistributionInput) (*marketv1.MarketplaceDistributionItem, error)
}

var _ Service = (*serviceImpl)(nil)

// serviceImpl is the default marketplace domain service implementation.
type serviceImpl struct {
	artifacts ArtifactStore
	gitRemote gitRemoteClient
}

// New creates a marketplace domain service. artifacts is required for package
// upload persistence, document body reads, and controlled download streaming.
// When artifacts is nil, New creates a local filesystem store under the default
// marketplace artifact root so builtin deployments remain self-contained.
func New(artifacts ArtifactStore) (Service, error) {
	if artifacts == nil {
		store, err := NewLocalArtifactStore("")
		if err != nil {
			return nil, err
		}
		artifacts = store
	}
	return &serviceImpl{artifacts: artifacts}, nil
}
