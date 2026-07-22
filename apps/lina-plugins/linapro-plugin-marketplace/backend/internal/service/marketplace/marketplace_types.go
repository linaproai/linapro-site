// This file defines marketplace service input and output records. The records
// are service-owned projections rather than DAO entities so controllers do not
// need to depend on generated persistence structs.

package marketplace

import (
	"io"
	"time"

	marketv1 "linapro-plugin-marketplace/backend/api/market/v1"
)

// PublisherStatus identifies whether a publisher may create marketplace drafts.
type PublisherStatus string

const (
	// PublisherStatusActive allows the publisher to create and submit plugin drafts.
	PublisherStatusActive PublisherStatus = "active"
	// PublisherStatusSuspended blocks publisher write actions until restored by review authority.
	PublisherStatusSuspended PublisherStatus = "suspended"
)

// String returns the serialized publisher status value.
func (value PublisherStatus) String() string { return string(value) }

// CreatePublisherInput carries publisher profile fields from the controller.
type CreatePublisherInput struct {
	PublisherKey string // PublisherKey is the stable owner key.
	Name         string // Name is the publisher display name.
	Summary      string // Summary is the short marketplace publisher summary.
	Homepage     string // Homepage is the optional publisher homepage URL.
	ContactEmail string // ContactEmail is the optional reviewer contact address.
	OwnerUserID  int64  // OwnerUserID is the user who owns this publisher profile.
	OwnerOrgID   int64  // OwnerOrgID is the owning organization, or 0 when absent.
}

// UpdatePublisherInput carries mutable publisher profile fields for an owned publisher.
type UpdatePublisherInput struct {
	CurrentPublisherKey string // CurrentPublisherKey locates the owned publisher profile before update.
	PublisherKey        string // PublisherKey is the desired key after update; may rename when unique.
	Name                string // Name is the publisher display name.
	Summary             string // Summary is the short marketplace publisher summary.
	Homepage            string // Homepage is the optional publisher homepage URL.
	ContactEmail        string // ContactEmail is the optional reviewer contact address.
	OwnerUserID         int64  // OwnerUserID is the user who must own this publisher profile.
}

// ListPublishersInput carries publisher list filters for the current operator.
type ListPublishersInput struct {
	PageNum     int    // PageNum is the 1-based page number.
	PageSize    int    // PageSize is the page size with a service-enforced upper bound.
	Keyword     string // Keyword optionally filters publisher key, name, and summary.
	OwnerUserID int64  // OwnerUserID optionally restricts results to one owning user.
}

// PublisherListOutput is the paginated publisher list projection.
type PublisherListOutput struct {
	List  []*marketv1.MarketplacePublisherItem
	Total int
}

// OpenDownloadContentInput identifies one requester-owned download session for streaming.
type OpenDownloadContentInput struct {
	SessionID       string            // SessionID is the opaque download session ID.
	RequesterUserID int64             // RequesterUserID is required and must match the session owner.
	Visibility      VisibilitySubject // Visibility is the current caller snapshot for download filtering.
}

// OpenDownloadContentOutput carries artifact metadata and an open content reader.
type OpenDownloadContentOutput struct {
	Session  *marketv1.MarketplaceDownloadSessionItem
	FileName string
	// Body is the artifact content reader. Callers must close it.
	Body io.ReadCloser
}

// SavePluginDraftInput carries marketplace plugin identity fields.
type SavePluginDraftInput struct {
	PublisherKey string                         // PublisherKey identifies the owner publisher.
	OwnerUserID  int64                          // OwnerUserID is required and must own the publisher.
	PluginID     string                         // PluginID is the stable marketplace plugin ID.
	Name         string                         // Name is the marketplace plugin display name.
	Summary      string                         // Summary is the marketplace list summary.
	Description  string                         // Description is the detail-page description.
	PluginType   marketv1.MarketplacePluginType // PluginType is source or dynamic.
	Visibility   marketv1.MarketplaceVisibility // Visibility is the initial marketplace visibility policy.
	Icon         string                         // Icon is the optional icon URL or resource path.
	Homepage     string                         // Homepage is the optional plugin homepage.
	Repository   string                         // Repository is the optional source repository URL.
	License      string                         // License is the license identifier shown before download.
}

// SaveReleaseDraftInput carries a scanned marketplace package draft.
type SaveReleaseDraftInput struct {
	PublisherKey       string                         // PublisherKey identifies the owner publisher.
	OwnerUserID        int64                          // OwnerUserID is required and must own the publisher.
	PluginID           string                         // PluginID identifies the marketplace plugin.
	Version            string                         // Version is the immutable release version.
	PluginType         marketv1.MarketplacePluginType // PluginType is source or dynamic.
	Visibility         marketv1.MarketplaceVisibility // Visibility is the release visibility policy.
	MinHostVersion     string                         // MinHostVersion is the lower host compatibility bound.
	MaxHostVersion     string                         // MaxHostVersion is the optional upper host compatibility bound.
	ManifestSnapshot   string                         // ManifestSnapshot is JSON produced by package scanning.
	DependencySummary  string                         // DependencySummary is JSON produced by package scanning.
	HostServiceSummary string                         // HostServiceSummary is JSON produced by package scanning.
	RouteSummary       string                         // RouteSummary is JSON produced by package scanning.
	SQLSummary         string                         // SQLSummary is JSON produced by package scanning.
	I18NSummary        string                         // I18NSummary is JSON produced by package scanning.
	DocsSummary        string                         // DocsSummary is JSON produced by package scanning.
	RiskSummary        string                         // RiskSummary is JSON produced by package scanning.
	ReviewMessage      string                         // ReviewMessage stores a scanner or publisher note.
	ReplaceDraft       bool                           // ReplaceDraft allows replacing an existing mutable draft.
	SourceRef          string                         // SourceRef is the Git logical tag/branch name for git-sourced drafts.
	SourceCommit       string                         // SourceCommit is the pinned full commit SHA resolved during Git discovery.
}

// UploadSourcePackageInput carries one uploaded source plugin marketplace package.
type UploadSourcePackageInput struct {
	PublisherKey   string                         // PublisherKey identifies the owner publisher.
	OwnerUserID    int64                          // OwnerUserID is required and must own the plugin publisher.
	PluginID       string                         // PluginID identifies the marketplace plugin.
	Version        string                         // Version identifies the release draft.
	PackagePath    string                         // PackagePath is the server-local uploaded ZIP file path.
	FileName       string                         // FileName is the original client file name.
	ContentType    string                         // ContentType is the uploaded file MIME type.
	StorageKey     string                         // StorageKey is the durable object key, or empty to derive from checksum.
	Visibility     marketv1.MarketplaceVisibility // Visibility is the release visibility policy.
	MinHostVersion string                         // MinHostVersion overrides the lower host compatibility bound.
	MaxHostVersion string                         // MaxHostVersion overrides the upper host compatibility bound.
	ReplaceDraft   bool                           // ReplaceDraft allows replacing an existing mutable draft.
	AutoCreate     bool                           // AutoCreate creates a private draft plugin from plugin.yaml when missing.
}

// UploadDynamicPackageInput carries one uploaded dynamic runtime marketplace package.
type UploadDynamicPackageInput struct {
	PublisherKey   string                         // PublisherKey identifies the owner publisher.
	OwnerUserID    int64                          // OwnerUserID is required and must own the plugin publisher.
	PluginID       string                         // PluginID identifies the marketplace plugin.
	Version        string                         // Version identifies the release draft.
	PackagePath    string                         // PackagePath is the server-local uploaded ZIP file path.
	FileName       string                         // FileName is the original client file name.
	ContentType    string                         // ContentType is the uploaded file MIME type.
	StorageKey     string                         // StorageKey is the durable ZIP object key, or empty to derive from checksum.
	WasmStorageKey string                         // WasmStorageKey is the extracted plugin.wasm object key, or empty to derive.
	Visibility     marketv1.MarketplaceVisibility // Visibility is the release visibility policy.
	MinHostVersion string                         // MinHostVersion overrides the lower host compatibility bound.
	MaxHostVersion string                         // MaxHostVersion overrides the upper host compatibility bound.
	ReplaceDraft   bool                           // ReplaceDraft allows replacing an existing mutable draft.
	AutoCreate     bool                           // AutoCreate creates a private draft plugin from plugin.yaml when missing.
}

// VisibilitySubject carries the current caller identity used to inject
// marketplace visibility filters into database queries. Zero values deliberately
// match only public marketplace records until controllers provide request
// identity, tenant, organization, or reserved-license scope snapshots.
type VisibilitySubject struct {
	UserID             int64    // UserID is the current authenticated user ID, or 0 when anonymous.
	TenantID           int64    // TenantID is the current tenant ID, or 0 when no tenant scope exists.
	OrgIDs             []int64  // OrgIDs are organization IDs already resolved as visible to the caller.
	ReservedLicenseIDs []string // ReservedLicenseIDs are future paid or reserved authorization scope IDs.
	CanPublish         bool     // CanPublish reports whether an owner may inspect unpublished publishing data.
	CanReview          bool     // CanReview reports whether the caller may inspect unpublished review data.
}

// GetReleaseDocumentInput carries one marketplace document index lookup request.
type GetReleaseDocumentInput struct {
	PluginID      string            // PluginID identifies the marketplace plugin.
	Version       string            // Version identifies the release.
	Locale        string            // Locale is the preferred document locale.
	DefaultLocale string            // DefaultLocale is the plugin-declared fallback locale.
	Path          string            // Path is the requested document path, empty for index.md.
	Visibility    VisibilitySubject // Visibility is the current caller snapshot for view filtering.
}

// ListPluginsInput carries marketplace catalog filters and pagination.
type ListPluginsInput struct {
	PageNum     int                            // PageNum is the one-based catalog page number.
	PageSize    int                            // PageSize is bounded by maxMarketplacePageSize.
	Keyword     string                         // Keyword matches the read-model search projection.
	PluginType  marketv1.MarketplacePluginType // PluginType optionally narrows source or dynamic plugins.
	TagCode     string                         // TagCode optionally narrows plugins through the tag relation table.
	Publisher   string                         // Publisher optionally narrows by publisher key.
	HostVersion string                         // HostVersion optionally applies a database-side compatibility range filter.
	Visibility  VisibilitySubject              // Visibility is the current caller snapshot for catalog filtering.
	Locale      string                         // Locale is the preferred display language for name/summary projection.
}

// ListOwnedPluginsInput carries publisher-owned plugin list filters.
type ListOwnedPluginsInput struct {
	PageNum    int                            // PageNum is the one-based page number.
	PageSize   int                            // PageSize is bounded by maxMarketplacePageSize.
	Keyword    string                         // Keyword optionally matches plugin ID, name, and summary.
	PluginType marketv1.MarketplacePluginType // PluginType optionally narrows source or dynamic plugins.
	// Status optionally narrows by marketplace lifecycle status
	// (draft/published/delisted/deprecated) or process pipeline status
	// (pending_verify/pending_review/completed/failed).
	Status      string
	OwnerUserID int64  // OwnerUserID is required and filters publishers owned by the user.
	Locale      string // Locale is the preferred display language for name/summary projection.
}

// ListManagedPluginsInput carries reviewer-managed plugin list filters.
type ListManagedPluginsInput struct {
	PageNum    int                            // PageNum is the one-based page number.
	PageSize   int                            // PageSize is bounded by maxMarketplacePageSize.
	Keyword    string                         // Keyword optionally matches plugin ID, name, summary, and publisher.
	PluginType marketv1.MarketplacePluginType // PluginType optionally narrows source or dynamic plugins.
	// Status optionally narrows by marketplace lifecycle status
	// (draft/published/delisted/deprecated) or process pipeline status
	// (pending_verify/pending_review/completed/failed).
	Status    string
	Publisher string // Publisher optionally narrows by publisher key.
	Locale    string // Locale is the preferred display language for name/summary projection.
}

// ListReviewQueueInput carries cross-plugin review queue filters.
type ListReviewQueueInput struct {
	PageNum      int                              // PageNum is the one-based page number.
	PageSize     int                              // PageSize is bounded by maxMarketplacePageSize.
	PluginID     string                           // PluginID optionally narrows one plugin.
	ReviewStatus marketv1.MarketplaceReviewStatus // ReviewStatus optionally narrows one review state.
	Keyword      string                           // Keyword optionally matches plugin ID and version.
}

// ReviewQueueOutput contains one paginated marketplace review queue page.
type ReviewQueueOutput struct {
	List  []*marketv1.MarketplaceReviewQueueItem
	Total int
}

// GetPluginDetailInput identifies one marketplace plugin detail projection.
type GetPluginDetailInput struct {
	PluginID   string            // PluginID is the stable marketplace plugin ID.
	Visibility VisibilitySubject // Visibility is the current caller snapshot for detail filtering.
	Locale     string            // Locale is the preferred display language for name/summary projection.
}

// ListReleasesInput carries release list filters and pagination.
type ListReleasesInput struct {
	PluginID     string                           // PluginID identifies the owning marketplace plugin.
	PageNum      int                              // PageNum is the one-based release page number.
	PageSize     int                              // PageSize is bounded by maxMarketplacePageSize.
	Status       marketv1.MarketplaceStatus       // Status optionally narrows release lifecycle status.
	ReviewStatus marketv1.MarketplaceReviewStatus // ReviewStatus optionally narrows release review status.
	Visibility   VisibilitySubject                // Visibility is the current caller snapshot for release filtering.
}

// ListReleaseRisksInput carries release risk filters and pagination.
type ListReleaseRisksInput struct {
	PluginID   string                           // PluginID identifies the owning marketplace plugin.
	Version    string                           // Version identifies the release whose risks are queried.
	PageNum    int                              // PageNum is the one-based risk page number.
	PageSize   int                              // PageSize is bounded by maxMarketplacePageSize.
	Type       marketv1.MarketplaceRiskType     // Type optionally narrows risk findings.
	Severity   marketv1.MarketplaceRiskSeverity // Severity optionally narrows risk findings.
	Visibility VisibilitySubject                // Visibility is the current caller snapshot for risk filtering.
}

// PluginListOutput contains one paginated marketplace catalog page.
type PluginListOutput struct {
	List  []*marketv1.MarketplacePluginListItem
	Total int
}

// PluginDetailOutput contains one marketplace plugin detail projection.
type PluginDetailOutput struct {
	Plugin *marketv1.MarketplacePluginDetailItem
}

// ReleaseListOutput contains one paginated marketplace release page.
type ReleaseListOutput struct {
	List  []*marketv1.MarketplaceReleaseItem
	Total int
}

// RiskListOutput contains one paginated marketplace risk page.
type RiskListOutput struct {
	List  []*marketv1.MarketplaceRiskItem
	Total int
}

// DocumentOutput contains the selected marketplace release document plus the
// same-path language bundle for local switching.
type DocumentOutput struct {
	Document  *marketv1.MarketplaceDocumentItem
	Documents []*marketv1.MarketplaceDocumentItem
}

// DownloadEventType identifies one controlled marketplace download event.
type DownloadEventType string

const (
	// DownloadEventTypeCreated records successful download-session creation.
	DownloadEventTypeCreated DownloadEventType = "created"
	// DownloadEventTypeStarted records that controlled artifact streaming started.
	DownloadEventTypeStarted DownloadEventType = "started"
	// DownloadEventTypeCompleted records that controlled artifact streaming completed.
	DownloadEventTypeCompleted DownloadEventType = "completed"
	// DownloadEventTypeFailed records that controlled artifact streaming failed.
	DownloadEventTypeFailed DownloadEventType = "failed"
)

// String returns the serialized download event type.
func (value DownloadEventType) String() string { return string(value) }

// CreateDownloadSessionInput carries one download-session creation request.
type CreateDownloadSessionInput struct {
	PluginID        string                           // PluginID identifies the marketplace plugin.
	Version         string                           // Version identifies the release requested for download.
	ArtifactType    marketv1.MarketplaceArtifactType // ArtifactType optionally selects a specific release artifact.
	RequesterUserID int64                            // RequesterUserID binds the session to the current user.
	Visibility      VisibilitySubject                // Visibility is the current caller snapshot for download filtering.
	TTL             time.Duration                    // TTL overrides the default short-lived session duration when positive.
}

// GetDownloadSessionInput identifies one requester-owned download session.
type GetDownloadSessionInput struct {
	SessionID       string            // SessionID is the opaque download session ID.
	RequesterUserID int64             // RequesterUserID is required and must match the session owner.
	Visibility      VisibilitySubject // Visibility is the current caller snapshot for download filtering.
}

// RecordDownloadEventInput carries one controlled download event.
type RecordDownloadEventInput struct {
	SessionID       string            // SessionID identifies the download session.
	EventType       DownloadEventType // EventType is started, completed, or failed; blank defaults to started.
	RequesterUserID int64             // RequesterUserID is required and must match the session owner.
	ClientIPHash    string            // ClientIPHash is an optional hashed client IP for statistics.
	UserAgentHash   string            // UserAgentHash is an optional hashed user agent for statistics.
	Visibility      VisibilitySubject // Visibility is the current caller snapshot for download filtering.
}

// RefreshDownloadStatisticsInput identifies the plugin whose snapshot is rebuilt.
type RefreshDownloadStatisticsInput struct {
	PluginID string // PluginID identifies the marketplace plugin to refresh.
}

// DownloadSessionOutput contains one download-session projection.
type DownloadSessionOutput struct {
	Session *marketv1.MarketplaceDownloadSessionItem
}

// SubmitReleaseReviewInput identifies a mutable release submitted by a publisher.
type SubmitReleaseReviewInput struct {
	PublisherKey string // PublisherKey identifies the owner publisher.
	OwnerUserID  int64  // OwnerUserID is required and must own the plugin publisher.
	PluginID     string // PluginID identifies the marketplace plugin.
	Version      string // Version identifies the draft release.
	Message      string // Message is the optional publisher note for reviewers.
}

// ReviewReleaseInput carries reviewer decision for one submitted release.
type ReviewReleaseInput struct {
	PluginID     string                           // PluginID identifies the marketplace plugin.
	Version      string                           // Version identifies the submitted release.
	ReviewStatus marketv1.MarketplaceReviewStatus // ReviewStatus must be approved or rejected.
	Message      string                           // Message is the reviewer note stored with the release.
}

// UpdatePluginStatusInput carries reviewer lifecycle status update fields.
type UpdatePluginStatusInput struct {
	PluginID string                     // PluginID identifies the marketplace plugin.
	Status   marketv1.MarketplaceStatus // Status is the target marketplace lifecycle state.
	Message  string                     // Message is the optional status-change reason.
}

// PublisherRecord is the service-owned publisher projection.
type PublisherRecord struct {
	ID           int
	PublisherKey string
	Name         string
	Summary      string
	OwnerUserID  int64
	OwnerOrgID   int64
	Verified     bool
	Status       PublisherStatus
	Homepage     string
	ContactEmail string
}

// PluginRecord is the service-owned marketplace plugin identity projection.
type PluginRecord struct {
	ID              int
	PublisherID     int
	PluginID        string
	Name            string
	Summary         string
	Description     string
	PluginType      marketv1.MarketplacePluginType
	MarketStatus    marketv1.MarketplaceStatus
	ProcessStatus   marketv1.MarketplaceProcessStatus
	Visibility      marketv1.MarketplaceVisibility
	LatestReleaseID int
	LatestVersion   string
	Icon            string
	Homepage        string
	Repository      string
	License         string
	DownloadCount   int64
	SourceKind      marketv1.MarketplaceSourceKind
	RepoURL         string
	RepoProvider    marketv1.MarketplaceRepoProvider
	RepoPath        string // relative plugin root inside the repository; empty for repository-root plugins
	CredentialRef   string // never expose token; presence implies requiresAuth
	LastSyncAt      *time.Time
	LastSyncStatus  string
	LastSyncMessage string
	PublishedAt     *time.Time
	UpdatedAt       *time.Time
}

// RegisterGitSourceResult carries one or more plugins discovered from a repository.
type RegisterGitSourceResult struct {
	Plugins []*PluginRecord
}

// ReleaseRecord is the service-owned marketplace release projection.
type ReleaseRecord struct {
	ID             int
	PluginRecordID int
	PublisherID    int
	PluginID       string
	Version        string
	SourceRef      string
	SourceCommit   string
	PluginType     marketv1.MarketplacePluginType
	ReleaseStatus  marketv1.MarketplaceStatus
	ReviewStatus   marketv1.MarketplaceReviewStatus
	ProcessStatus  marketv1.MarketplaceProcessStatus
	Visibility     marketv1.MarketplaceVisibility
	MinHostVersion string
	MaxHostVersion string
	ReviewMessage  string
	SubmittedAt    *time.Time
	ReviewedAt     *time.Time
	PublishedAt    *time.Time
	UpdatedAt      *time.Time
}

// ArtifactRecord is the service-owned marketplace artifact projection.
type ArtifactRecord struct {
	ID             int
	ReleaseID      int
	PluginID       string
	Version        string
	ArtifactType   marketv1.MarketplaceArtifactType
	StorageKey     string
	FileName       string
	ContentType    string
	SizeBytes      int64
	Sha256         string
	ManifestSha256 string
	WasmSha256     string
	UpdatedAt      *time.Time
}

// DocumentRecord is the service-owned marketplace document index projection.
type DocumentRecord struct {
	ID              int
	ReleaseID       int
	PluginID        string
	Version         string
	Locale          string
	RequestedLocale string
	ResolvedLocale  string
	Path            string
	SourceKind      string
	Title           string
	Summary         string
	ContentHash     string
	SearchText      string
	RenderedContent string
	FallbackUsed    bool
	UpdatedAt       *time.Time
}

// PackageDiagnostic describes one package scanner finding returned to callers.
type PackageDiagnostic struct {
	Code     string
	Severity marketv1.MarketplaceRiskSeverity
	Source   string
	Message  string
}

// SourcePackageUploadResult returns the draft release and source artifact scan result.
type SourcePackageUploadResult struct {
	Release     *ReleaseRecord
	Artifact    *ArtifactRecord
	Diagnostics []*PackageDiagnostic
}

// DynamicPackageUploadResult returns the draft release and dynamic artifact scan result.
type DynamicPackageUploadResult struct {
	Release         *ReleaseRecord
	PackageArtifact *ArtifactRecord
	WasmArtifact    *ArtifactRecord
	Diagnostics     []*PackageDiagnostic
}
