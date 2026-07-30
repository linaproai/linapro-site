// This file defines marketplace catalog query API DTOs. The catalog endpoints
// expose paginated read-model projections and detail projections without
// requiring the frontend to issue one detail request per list row.

package v1

import "github.com/gogf/gf/v2/frame/g"

// PluginListReq is the request for querying the marketplace plugin catalog.
type PluginListReq struct {
	g.Meta      `path:"/market/plugins" method:"get" tags:"Plugin Marketplace" summary:"Query marketplace plugins" permission:"market:plugin:view" dc:"Return a visibility-filtered paginated marketplace plugin catalog using the read model projection. The response contains the fields needed by the list page and does not include full Markdown documents or artifact content."`
	PageNum     int                   `json:"pageNum" d:"1" v:"min:1" dc:"Page number for the marketplace plugin catalog, starting from 1" eg:"1"`
	PageSize    int                   `json:"pageSize" d:"20" v:"min:1|max:100" dc:"Page size for the marketplace plugin catalog, default 20 and maximum 100" eg:"20"`
	Keyword     string                `json:"keyword" dc:"Optional fuzzy search keyword matched against plugin name, plugin ID, summary, publisher, and search projection; query all visible plugins when empty" eg:"workflow"`
	PluginType  MarketplacePluginType `json:"pluginType" dc:"Optional plugin type filter: source=source plugin dynamic=dynamic plugin, query all plugin types when empty" eg:"source"`
	TagCode     string                `json:"tagCode" dc:"Optional category or tag code filter, query all tags when empty" eg:"observability"`
	Publisher   string                `json:"publisher" dc:"Optional publisher key filter, query all publishers when empty" eg:"linapro"`
	HostVersion string                `json:"hostVersion" dc:"Optional LinaPro host version used for compatibility filtering, skip compatibility filtering when empty" eg:"v0.5.0"`
}

// PluginListRes is the response for querying the marketplace plugin catalog.
type PluginListRes struct {
	List  []*MarketplacePluginListItem `json:"list" dc:"Paginated marketplace plugin list items with minimal display projection" eg:"[]"`
	Total int                          `json:"total" dc:"Total number of visible marketplace plugins matching the filters before pagination" eg:"1"`
}

// MyPluginListReq is the request for listing marketplace plugins owned by the current publisher.
type MyPluginListReq struct {
	g.Meta         `path:"/market/my-plugins" method:"get" tags:"Plugin Marketplace" summary:"Query my marketplace plugins" permission:"market:plugin:publish" dc:"Return paginated marketplace plugins owned by publishers bound to the current user. Draft and unpublished plugins are included so publishers can manage their own catalog entries."`
	PageNum        int                   `json:"pageNum" d:"1" v:"min:1" dc:"Page number starting from 1" eg:"1"`
	PageSize       int                   `json:"pageSize" d:"20" v:"min:1|max:100" dc:"Page size, default 20 and maximum 100" eg:"20"`
	Keyword        string                `json:"keyword" dc:"Optional fuzzy search against plugin ID, name, and summary; match all owned plugins when empty" eg:"workflow"`
	PluginType     MarketplacePluginType `json:"pluginType" dc:"Optional plugin type filter: source=source-code plugin or dynamic=WASM runtime plugin; include all plugin types when empty" eg:"source"`
	Status         string                `json:"status" dc:"Optional list status filter. Accepts marketplace lifecycle values (draft, published, delisted, deprecated) or process pipeline values shown in publisher UI (pending_verify, pending_review, completed, failed); include all statuses when empty" eg:"pending_verify"`
	OrderBy        string                `json:"orderBy" dc:"Optional sort field: pluginId, marketStatus, downloadCount, or updatedAt; defaults to pluginId ascending when empty or unsupported" eg:"pluginId"`
	OrderDirection string                `json:"orderDirection" d:"asc" dc:"Sort direction: asc=ascending or desc=descending; defaults to asc" eg:"asc"`
}

// MyPluginListRes is the response for listing marketplace plugins owned by the current publisher.
type MyPluginListRes struct {
	List  []*MarketplacePluginListItem `json:"list" dc:"Paginated owned marketplace plugins" eg:"[]"`
	Total int                          `json:"total" dc:"Total owned plugins matching filters" eg:"1"`
}

// ManagedPluginListReq is the request for listing all marketplace plugins for reviewers.
type ManagedPluginListReq struct {
	g.Meta     `path:"/market/managed-plugins" method:"get" tags:"Plugin Marketplace" summary:"Query managed marketplace plugins" permission:"market:plugin:review" dc:"Return paginated marketplace plugins across all publishers for marketplace operators. Includes draft, published, delisted, and deprecated statuses with latest review state."`
	PageNum    int                   `json:"pageNum" d:"1" v:"min:1" dc:"Page number starting from 1" eg:"1"`
	PageSize   int                   `json:"pageSize" d:"20" v:"min:1|max:100" dc:"Page size, default 20 and maximum 100" eg:"20"`
	Keyword    string                `json:"keyword" dc:"Optional fuzzy search against plugin ID, name, summary, and publisher; match all managed plugins when empty" eg:"workflow"`
	PluginType MarketplacePluginType `json:"pluginType" dc:"Optional plugin type filter: source=source-code plugin or dynamic=WASM runtime plugin; include all plugin types when empty" eg:"source"`
	Status     string                `json:"status" dc:"Optional list status filter. Accepts marketplace lifecycle values (draft, published, delisted, deprecated) or process pipeline values (pending_verify, pending_review, completed, failed); include all statuses when empty" eg:"published"`
	Publisher  string                `json:"publisher" dc:"Optional publisher key filter; include all publishers when empty" eg:"linapro"`
}

// ManagedPluginListRes is the response for listing all marketplace plugins for reviewers.
type ManagedPluginListRes struct {
	List  []*MarketplacePluginListItem `json:"list" dc:"Paginated managed marketplace plugins" eg:"[]"`
	Total int                          `json:"total" dc:"Total managed plugins matching filters" eg:"1"`
}

// ReviewQueueListReq is the request for listing marketplace releases pending review.
type ReviewQueueListReq struct {
	g.Meta       `path:"/market/review-queue" method:"get" tags:"Plugin Marketplace" summary:"Query marketplace review queue" permission:"market:plugin:review" dc:"Return paginated marketplace releases awaiting review. Defaults to submitted and reviewing states and supports optional plugin ID and review status filters."`
	PageNum      int                     `json:"pageNum" d:"1" v:"min:1" dc:"Page number starting from 1" eg:"1"`
	PageSize     int                     `json:"pageSize" d:"20" v:"min:1|max:100" dc:"Page size, default 20 and maximum 100" eg:"20"`
	PluginId     string                  `json:"pluginId" dc:"Optional stable plugin ID filter; include releases from all plugins when empty" eg:"linapro-demo-source"`
	ReviewStatus MarketplaceReviewStatus `json:"reviewStatus" dc:"Optional review status filter: draft=mutable draft, submitted=awaiting review, reviewing=in review, approved=approved, or rejected=rejected; defaults to submitted and reviewing when empty" eg:"submitted"`
	Keyword      string                  `json:"keyword" dc:"Optional fuzzy search against plugin ID and version; do not apply keyword filtering when empty" eg:"v0.1.0"`
}

// ReviewQueueListRes is the response for listing marketplace releases pending review.
type ReviewQueueListRes struct {
	List  []*MarketplaceReviewQueueItem `json:"list" dc:"Paginated review queue items" eg:"[]"`
	Total int                           `json:"total" dc:"Total review queue items matching filters" eg:"1"`
}

// MarketplaceReviewQueueItem is one release row shown in the marketplace review queue.
type MarketplaceReviewQueueItem struct {
	PluginId      string                    `json:"pluginId" dc:"Stable plugin ID" eg:"linapro-demo-source"`
	PluginName    string                    `json:"pluginName" dc:"Marketplace plugin display name" eg:"Source Plugin Demo"`
	Version       string                    `json:"version" dc:"Release version" eg:"v0.1.0"`
	PluginType    MarketplacePluginType     `json:"pluginType" dc:"Plugin type: source or dynamic" eg:"source"`
	ReleaseStatus MarketplaceStatus         `json:"releaseStatus" dc:"Release lifecycle status" eg:"draft"`
	ReviewStatus  MarketplaceReviewStatus   `json:"reviewStatus" dc:"Release review status" eg:"submitted"`
	Visibility    MarketplaceVisibility     `json:"visibility" dc:"Release visibility policy" eg:"public"`
	Publisher     *MarketplacePublisherItem `json:"publisher" dc:"Publisher snapshot" eg:"{}"`
	Artifact      *MarketplaceArtifactItem  `json:"artifact" dc:"Primary artifact summary" eg:"{}"`
	ReviewMessage string                    `json:"reviewMessage" dc:"Latest review or scanner message" eg:""`
	SubmittedAt   *int64                    `json:"submittedAt,omitempty" dc:"Review submission time as Unix milliseconds" eg:"1767247200000"`
	UpdatedAt     *int64                    `json:"updatedAt,omitempty" dc:"Release last updated time as Unix milliseconds" eg:"1767247200000"`
}

// PluginDetailReq is the request for reading one marketplace plugin detail.
type PluginDetailReq struct {
	g.Meta   `path:"/market/plugins/{pluginId}" method:"get" tags:"Plugin Marketplace" summary:"Get marketplace plugin detail" permission:"market:plugin:view" dc:"Return one visibility-filtered marketplace plugin detail with publisher, tags, latest release, compatibility, risk summary, and source-delivery guidance. The endpoint returns not found for invisible private plugins to avoid existence disclosure."`
	PluginId string `json:"pluginId" v:"required|length:1,64" dc:"Stable plugin ID to read from the marketplace catalog" eg:"linapro-demo-source"`
}

// PluginDetailRes is the response for reading one marketplace plugin detail.
type PluginDetailRes struct {
	Plugin *MarketplacePluginDetailItem `json:"plugin" dc:"Marketplace plugin detail projection for the requested visible plugin" eg:"{}"`
}

// MyPluginDetailReq is the request for reading one publisher-owned plugin detail.
type MyPluginDetailReq struct {
	g.Meta   `path:"/market/my-plugins/{pluginId}" method:"get" tags:"Plugin Marketplace" summary:"Get my marketplace plugin detail" permission:"market:plugin:publish" dc:"Return one marketplace plugin detail owned by the current publisher user. The service enforces publisher ownership before exposing draft or unpublished state."`
	PluginId string `json:"pluginId" v:"required|length:1,64" dc:"Stable plugin ID owned by the current publisher user" eg:"linapro-demo-source"`
}

// MyPluginDetailRes is the response for reading one publisher-owned plugin detail.
type MyPluginDetailRes struct {
	Plugin *MarketplacePluginDetailItem `json:"plugin" dc:"Publisher-owned marketplace plugin detail projection" eg:"{}"`
}

// ManagedPluginDetailReq is the request for reading one reviewer-managed plugin detail.
type ManagedPluginDetailReq struct {
	g.Meta   `path:"/market/managed-plugins/{pluginId}" method:"get" tags:"Plugin Marketplace" summary:"Get managed marketplace plugin detail" permission:"market:plugin:review" dc:"Return one marketplace plugin detail for a review operator, including unpublished lifecycle state needed by the management workbench."`
	PluginId string `json:"pluginId" v:"required|length:1,64" dc:"Stable plugin ID inspected by the review operator" eg:"linapro-demo-source"`
}

// ManagedPluginDetailRes is the response for reading one reviewer-managed plugin detail.
type ManagedPluginDetailRes struct {
	Plugin *MarketplacePluginDetailItem `json:"plugin" dc:"Reviewer-managed marketplace plugin detail projection" eg:"{}"`
}

// ReleaseListReq is the request for querying visible releases of one marketplace plugin.
type ReleaseListReq struct {
	g.Meta       `path:"/market/plugins/{pluginId}/releases" method:"get" tags:"Plugin Marketplace" summary:"Query marketplace plugin releases" permission:"market:plugin:view" dc:"Return paginated visible releases for one marketplace plugin. The response includes release review state, artifact checksum summary, and publish timestamps without loading package content."`
	PluginId     string                  `json:"pluginId" v:"required|length:1,64" dc:"Stable plugin ID whose releases are queried" eg:"linapro-demo-source"`
	PageNum      int                     `json:"pageNum" d:"1" v:"min:1" dc:"Page number for marketplace release list, starting from 1" eg:"1"`
	PageSize     int                     `json:"pageSize" d:"20" v:"min:1|max:100" dc:"Page size for marketplace release list, default 20 and maximum 100" eg:"20"`
	Status       MarketplaceStatus       `json:"status" dc:"Optional release lifecycle status filter: draft, published, delisted, or deprecated; query all visible statuses when empty" eg:"published"`
	ReviewStatus MarketplaceReviewStatus `json:"reviewStatus" dc:"Optional review status filter: draft, submitted, reviewing, approved, or rejected; query all review states when empty" eg:"approved"`
}

// ReleaseListRes is the response for querying visible releases of one marketplace plugin.
type ReleaseListRes struct {
	List  []*MarketplaceReleaseItem `json:"list" dc:"Paginated visible marketplace release list" eg:"[]"`
	Total int                       `json:"total" dc:"Total number of visible releases matching the filters before pagination" eg:"1"`
}

// MyReleaseListReq is the request for querying releases of one publisher-owned plugin.
type MyReleaseListReq struct {
	g.Meta       `path:"/market/my-plugins/{pluginId}/releases" method:"get" tags:"Plugin Marketplace" summary:"Query my marketplace plugin releases" permission:"market:plugin:publish" dc:"Return paginated releases for a plugin owned by the current publisher user, including mutable and unpublished states required by the publishing workbench."`
	PluginId     string                  `json:"pluginId" v:"required|length:1,64" dc:"Stable publisher-owned plugin ID whose releases are queried" eg:"linapro-demo-source"`
	PageNum      int                     `json:"pageNum" d:"1" v:"min:1" dc:"Page number starting from 1" eg:"1"`
	PageSize     int                     `json:"pageSize" d:"20" v:"min:1|max:100" dc:"Page size, default 20 and maximum 100" eg:"20"`
	Status       MarketplaceStatus       `json:"status" dc:"Optional release lifecycle status filter: draft=unpublished draft, published=market-visible, delisted=removed from listings, or deprecated=discouraged; include all lifecycle statuses when empty" eg:"draft"`
	ReviewStatus MarketplaceReviewStatus `json:"reviewStatus" dc:"Optional release review status filter: draft=mutable draft, submitted=awaiting review, reviewing=in review, approved=approved, or rejected=rejected; include all review statuses when empty" eg:"submitted"`
}

// MyReleaseListRes is the response for querying publisher-owned plugin releases.
type MyReleaseListRes struct {
	List  []*MarketplaceReleaseItem `json:"list" dc:"Paginated publisher-owned marketplace release list" eg:"[]"`
	Total int                       `json:"total" dc:"Total owned-plugin releases matching the filters" eg:"1"`
}

// ManagedReleaseListReq is the request for querying reviewer-managed plugin releases.
type ManagedReleaseListReq struct {
	g.Meta       `path:"/market/managed-plugins/{pluginId}/releases" method:"get" tags:"Plugin Marketplace" summary:"Query managed marketplace plugin releases" permission:"market:plugin:review" dc:"Return paginated releases for one reviewer-managed plugin, including submitted, reviewing, approved, rejected, and draft lifecycle projections."`
	PluginId     string                  `json:"pluginId" v:"required|length:1,64" dc:"Stable plugin ID whose releases are inspected by the review operator" eg:"linapro-demo-source"`
	PageNum      int                     `json:"pageNum" d:"1" v:"min:1" dc:"Page number starting from 1" eg:"1"`
	PageSize     int                     `json:"pageSize" d:"20" v:"min:1|max:100" dc:"Page size, default 20 and maximum 100" eg:"20"`
	Status       MarketplaceStatus       `json:"status" dc:"Optional release lifecycle status filter: draft=unpublished draft, published=market-visible, delisted=removed from listings, or deprecated=discouraged; include all lifecycle statuses when empty" eg:"draft"`
	ReviewStatus MarketplaceReviewStatus `json:"reviewStatus" dc:"Optional release review status filter: draft=mutable draft, submitted=awaiting review, reviewing=in review, approved=approved, or rejected=rejected; include all review statuses when empty" eg:"submitted"`
}

// ManagedReleaseListRes is the response for querying reviewer-managed plugin releases.
type ManagedReleaseListRes struct {
	List  []*MarketplaceReleaseItem `json:"list" dc:"Paginated reviewer-managed marketplace release list" eg:"[]"`
	Total int                       `json:"total" dc:"Total managed releases matching the filters" eg:"1"`
}

// MarketplacePluginListItem is the minimal read-model projection used by marketplace catalog pages.
type MarketplacePluginListItem struct {
	PluginId           string                    `json:"pluginId" dc:"Stable plugin ID shown in marketplace catalog and used for detail routing" eg:"linapro-demo-source"`
	Name               string                    `json:"name" dc:"Marketplace plugin display name" eg:"Source Plugin Demo"`
	Summary            string                    `json:"summary" dc:"Short marketplace summary suitable for list cards and tables" eg:"Source plugin that provides workflow automation examples"`
	Publisher          *MarketplacePublisherItem `json:"publisher" dc:"Publisher snapshot displayed with the catalog item" eg:"{}"`
	PluginType         MarketplacePluginType     `json:"pluginType" dc:"Plugin type: source=source plugin dynamic=dynamic plugin" eg:"source"`
	MarketStatus       MarketplaceStatus         `json:"marketStatus" dc:"Marketplace status: draft, published, delisted, or deprecated" eg:"published"`
	ProcessStatus      MarketplaceProcessStatus  `json:"processStatus,omitempty" dc:"Async process status: pending_verify, pending_review, completed, or failed" eg:"pending_verify"`
	Visibility         MarketplaceVisibility     `json:"visibility" dc:"Visibility policy: public, private, or reserved" eg:"public"`
	LatestVersion      string                    `json:"latestVersion" dc:"Latest version associated with the plugin, including drafts for managed lists" eg:"v0.1.0"`
	LatestReviewStatus MarketplaceReviewStatus   `json:"latestReviewStatus,omitempty" dc:"Review status of the latest release when available" eg:"submitted"`
	MinHostVersion     string                    `json:"minHostVersion" dc:"Minimum compatible LinaPro host version in the latest release" eg:"v0.1.0"`
	MaxHostVersion     string                    `json:"maxHostVersion" dc:"Maximum compatible LinaPro host version in the latest release, empty when no upper bound exists" eg:"v1.0.0"`
	PrimaryTag         string                    `json:"primaryTag" dc:"Primary category tag code used for catalog grouping" eg:"observability"`
	TagCodes           []string                  `json:"tagCodes" dc:"Category and tag code snapshot for display and client-side badges" eg:"[\"observability\",\"audit\"]"`
	RiskCounts         MarketplaceRiskCounts     `json:"riskCounts" dc:"Risk finding count snapshot grouped by severity for the latest visible release" eg:"{}"`
	DownloadCount      int64                     `json:"downloadCount" dc:"Aggregated marketplace download count snapshot" eg:"1200"`
	SourceKind         MarketplaceSourceKind     `json:"sourceKind,omitempty" dc:"Publish source kind: git or upload" eg:"upload"`
	RepoUrl            string                    `json:"repoUrl,omitempty" dc:"Git repository URL when sourceKind is git" eg:"https://github.com/org/plugin.git"`
	RepoProvider       MarketplaceRepoProvider   `json:"repoProvider,omitempty" dc:"Git provider when sourceKind is git" eg:"github"`
	LastSyncStatus     string                    `json:"lastSyncStatus,omitempty" dc:"Last Git metadata sync status for git sources" eg:"success"`
	LastSyncMessage    string                    `json:"lastSyncMessage,omitempty" dc:"Last Git metadata sync diagnostic without secrets" eg:"discovered 2 draft releases"`
	LastSyncAt         *int64                    `json:"lastSyncAt,omitempty" dc:"Last Git metadata sync time as Unix timestamp in milliseconds" eg:"1767247200000"`
	PublishedAt        *int64                    `json:"publishedAt,omitempty" dc:"Latest publish time as Unix timestamp in milliseconds" eg:"1767247200000"`
	UpdatedAt          *int64                    `json:"updatedAt,omitempty" dc:"Read model last updated time as Unix timestamp in milliseconds" eg:"1767247200000"`
}

// MarketplacePluginDetailItem is the detail projection used by marketplace detail pages.
type MarketplacePluginDetailItem struct {
	PluginId        string                       `json:"pluginId" dc:"Stable plugin ID shown in marketplace detail and used for release routing" eg:"linapro-demo-source"`
	Name            string                       `json:"name" dc:"Marketplace plugin display name" eg:"Source Plugin Demo"`
	Summary         string                       `json:"summary" dc:"Short marketplace summary" eg:"Source plugin that provides workflow automation examples"`
	Description     string                       `json:"description" dc:"Long marketplace description displayed on the detail page" eg:"Provides workflow automation pages, backend APIs, and plugin resource examples."`
	Publisher       *MarketplacePublisherItem    `json:"publisher" dc:"Publisher profile snapshot for this marketplace plugin" eg:"{}"`
	PluginType      MarketplacePluginType        `json:"pluginType" dc:"Plugin type: source=source plugin dynamic=dynamic plugin" eg:"source"`
	MarketStatus    MarketplaceStatus            `json:"marketStatus" dc:"Marketplace status: draft, published, delisted, or deprecated" eg:"published"`
	ProcessStatus   MarketplaceProcessStatus     `json:"processStatus,omitempty" dc:"Async process status: pending_verify, pending_review, completed, or failed" eg:"pending_review"`
	Visibility      MarketplaceVisibility        `json:"visibility" dc:"Visibility policy: public, private, or reserved" eg:"public"`
	LatestVersion   string                       `json:"latestVersion" dc:"Latest visible published version in the marketplace" eg:"v0.1.0"`
	Icon            string                       `json:"icon" dc:"Marketplace icon URL or plugin-managed asset path, empty when no icon is available" eg:"/assets/marketplace/source-demo.png"`
	Homepage        string                       `json:"homepage" dc:"Plugin homepage URL, empty when no homepage is available" eg:"https://linapro.ai/plugins/source-demo"`
	Repository      string                       `json:"repository" dc:"Plugin source repository URL, empty when no repository is available" eg:"https://github.com/linaproai/linapro-demo-source"`
	License         string                       `json:"license" dc:"Plugin license identifier displayed before download" eg:"Apache-2.0"`
	Tags            []*MarketplaceTagItem        `json:"tags" dc:"Marketplace categories and tags associated with this plugin" eg:"[]"`
	LatestRelease   *MarketplaceReleaseItem      `json:"latestRelease" dc:"Latest visible release summary for compatibility, risk, and download decisions" eg:"{}"`
	RiskCounts      MarketplaceRiskCounts        `json:"riskCounts" dc:"Risk finding count snapshot grouped by severity for the latest visible release" eg:"{}"`
	DownloadCount   int64                        `json:"downloadCount" dc:"Aggregated marketplace download count snapshot" eg:"1200"`
	SourceDelivery  string                       `json:"sourceDelivery" dc:"Source plugin delivery guidance; source plugins require placement under apps/lina-plugins and host rebuild, while dynamic plugins continue through local dynamic upload governance" eg:"source_rebuild_required"`
	SourceKind      MarketplaceSourceKind        `json:"sourceKind,omitempty" dc:"Publish source kind: git or upload" eg:"git"`
	RepoUrl         string                       `json:"repoUrl,omitempty" dc:"Git repository URL when sourceKind is git" eg:"https://github.com/org/plugin.git"`
	RepoProvider    MarketplaceRepoProvider      `json:"repoProvider,omitempty" dc:"Git provider when sourceKind is git" eg:"github"`
	RepoPath        string                       `json:"repoPath,omitempty" dc:"Plugin root path relative to the repository root when sourceKind is git; empty when the repository root is the plugin root" eg:"apps/lina-plugins/linapro-demo-source"`
	RequiresAuth    bool                         `json:"requiresAuth,omitempty" dc:"Whether the Git source is private; platform tokens are never returned" eg:"true"`
	LastSyncStatus  string                       `json:"lastSyncStatus,omitempty" dc:"Last Git metadata sync status for git sources" eg:"success"`
	LastSyncMessage string                       `json:"lastSyncMessage,omitempty" dc:"Last Git metadata sync diagnostic without secrets" eg:"discovered 2 draft releases"`
	LastSyncAt      *int64                       `json:"lastSyncAt,omitempty" dc:"Last Git metadata sync time as Unix timestamp in milliseconds" eg:"1767247200000"`
	Distribution    *MarketplaceDistributionItem `json:"distribution,omitempty" dc:"Latest release distribution projection when available" eg:"{}"`
	PublishedAt     *int64                       `json:"publishedAt,omitempty" dc:"First marketplace publish time as Unix timestamp in milliseconds" eg:"1767247200000"`
	UpdatedAt       *int64                       `json:"updatedAt,omitempty" dc:"Marketplace plugin last updated time as Unix timestamp in milliseconds" eg:"1767247200000"`
}
