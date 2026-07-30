// This file defines marketplace documentation and review-risk API DTOs. These
// endpoints read version-scoped document indexes and risk findings after
// visibility filtering, without loading unrelated package resources.

package v1

import "github.com/gogf/gf/v2/frame/g"

// ReleaseDocsReq is the request for reading marketplace documentation of one release.
type ReleaseDocsReq struct {
	g.Meta   `path:"/market/plugins/{pluginId}/releases/{version}/docs" method:"get" tags:"Plugin Marketplace" summary:"Get marketplace release documentation" permission:"market:plugin:view" dc:"Return the selected version-scoped marketplace document and the same-path language bundle with safe-rendered content. The endpoint reads only persisted document snapshots for a visible release and rejects paths outside the version documentation boundary."`
	PluginId string `json:"pluginId" v:"required|length:1,64" dc:"Stable plugin ID whose release documentation is read" eg:"linapro-demo-source"`
	Version  string `json:"version" v:"required|length:1,32" dc:"Release version whose documentation is read" eg:"v0.1.0"`
	Locale   string `json:"locale" dc:"Preferred document locale; when empty the server uses the request locale and then applies marketplace fallback rules" eg:"en-US"`
	Path     string `json:"path" dc:"Document path inside manifest/docs. The default empty value reads the first available marketplace document, preferring index.md." eg:"index.md"`
}

// ReleaseDocsRes is the response for reading marketplace documentation of one release.
type ReleaseDocsRes struct {
	Document  *MarketplaceDocumentItem          `json:"document" dc:"Version-scoped marketplace document with safe rendered content and fallback metadata" eg:"{}"`
	Documents []*MarketplaceDocumentItem        `json:"documents" dc:"Same-path marketplace document language bundle with safe rendered content for local switching" eg:"[]"`
	Catalog   []*MarketplaceDocumentCatalogItem `json:"catalog" dc:"Version-scoped documentation catalog listing every available document path for navigation" eg:"[]"`
}

// MyReleaseDocsReq is the request for reading documentation of one publisher-owned release.
type MyReleaseDocsReq struct {
	g.Meta   `path:"/market/my-plugins/{pluginId}/releases/{version}/docs" method:"get" tags:"Plugin Marketplace" summary:"Get my marketplace release documentation" permission:"market:plugin:publish" dc:"Return the selected version-scoped document and same-path language bundle for a release owned by the current publisher user. Ownership is enforced before unpublished documentation is loaded."`
	PluginId string `json:"pluginId" v:"required|length:1,64" dc:"Stable publisher-owned plugin ID" eg:"linapro-demo-source"`
	Version  string `json:"version" v:"required|length:1,32" dc:"Owned release version whose documentation is read" eg:"v0.1.0"`
	Locale   string `json:"locale" dc:"Preferred document locale" eg:"en-US"`
	Path     string `json:"path" dc:"Document path inside the release documentation boundary" eg:"index.md"`
}

// MyReleaseDocsRes is the response for reading publisher-owned release documentation.
type MyReleaseDocsRes struct {
	Document  *MarketplaceDocumentItem          `json:"document" dc:"Publisher-owned version-scoped marketplace document" eg:"{}"`
	Documents []*MarketplaceDocumentItem        `json:"documents" dc:"Publisher-owned same-path marketplace document language bundle" eg:"[]"`
	Catalog   []*MarketplaceDocumentCatalogItem `json:"catalog" dc:"Publisher-owned version documentation catalog for path navigation" eg:"[]"`
}

// ManagedReleaseDocsReq is the request for reading reviewer-managed release documentation.
type ManagedReleaseDocsReq struct {
	g.Meta   `path:"/market/managed-plugins/{pluginId}/releases/{version}/docs" method:"get" tags:"Plugin Marketplace" summary:"Get managed marketplace release documentation" permission:"market:plugin:review" dc:"Return the selected version-scoped document and same-path language bundle for reviewer inspection, including documentation from submitted or otherwise unpublished releases."`
	PluginId string `json:"pluginId" v:"required|length:1,64" dc:"Stable plugin ID inspected by the review operator" eg:"linapro-demo-source"`
	Version  string `json:"version" v:"required|length:1,32" dc:"Managed release version whose documentation is read" eg:"v0.1.0"`
	Locale   string `json:"locale" dc:"Preferred document locale" eg:"en-US"`
	Path     string `json:"path" dc:"Document path inside the release documentation boundary" eg:"index.md"`
}

// ManagedReleaseDocsRes is the response for reading reviewer-managed release documentation.
type ManagedReleaseDocsRes struct {
	Document  *MarketplaceDocumentItem          `json:"document" dc:"Reviewer-managed version-scoped marketplace document" eg:"{}"`
	Documents []*MarketplaceDocumentItem        `json:"documents" dc:"Reviewer-managed same-path marketplace document language bundle" eg:"[]"`
	Catalog   []*MarketplaceDocumentCatalogItem `json:"catalog" dc:"Reviewer-managed version documentation catalog for path navigation" eg:"[]"`
}

// ReleaseRisksReq is the request for querying risk findings of one release.
type ReleaseRisksReq struct {
	g.Meta   `path:"/market/plugins/{pluginId}/releases/{version}/risks" method:"get" tags:"Plugin Marketplace" summary:"Query marketplace release risks" permission:"market:plugin:view" dc:"Return paginated review risk findings for a visible marketplace release. The response lets users review hostServices, routes, SQL, network, data-table, dependency, i18n, and documentation risks before downloading."`
	PluginId string                  `json:"pluginId" v:"required|length:1,64" dc:"Stable plugin ID whose release risks are queried" eg:"linapro-demo-source"`
	Version  string                  `json:"version" v:"required|length:1,32" dc:"Release version whose risk findings are queried" eg:"v0.1.0"`
	PageNum  int                     `json:"pageNum" d:"1" v:"min:1" dc:"Page number for marketplace release risk findings, starting from 1" eg:"1"`
	PageSize int                     `json:"pageSize" d:"20" v:"min:1|max:100" dc:"Page size for marketplace release risk findings, default 20 and maximum 100" eg:"20"`
	Type     MarketplaceRiskType     `json:"type" dc:"Optional risk type filter: host_service, dynamic_route, menu_permission, external_network, data_table, install_sql, uninstall_sql, mock_sql, dependency, multi_tenant, or docs" eg:"host_service"`
	Severity MarketplaceRiskSeverity `json:"severity" dc:"Optional risk severity filter: info, warning, or high" eg:"warning"`
}

// ReleaseRisksRes is the response for querying risk findings of one release.
type ReleaseRisksRes struct {
	List  []*MarketplaceRiskItem `json:"list" dc:"Paginated review risk findings for the visible marketplace release" eg:"[]"`
	Total int                    `json:"total" dc:"Total number of risk findings matching the filters before pagination" eg:"3"`
}

// MyReleaseRisksReq is the request for querying risks of one publisher-owned release.
type MyReleaseRisksReq struct {
	g.Meta   `path:"/market/my-plugins/{pluginId}/releases/{version}/risks" method:"get" tags:"Plugin Marketplace" summary:"Query my marketplace release risks" permission:"market:plugin:publish" dc:"Return paginated scanner findings for a release owned by the current publisher user, including unpublished release findings needed before submission."`
	PluginId string                  `json:"pluginId" v:"required|length:1,64" dc:"Stable publisher-owned plugin ID" eg:"linapro-demo-source"`
	Version  string                  `json:"version" v:"required|length:1,32" dc:"Owned release version whose risks are queried" eg:"v0.1.0"`
	PageNum  int                     `json:"pageNum" d:"1" v:"min:1" dc:"Page number starting from 1" eg:"1"`
	PageSize int                     `json:"pageSize" d:"20" v:"min:1|max:100" dc:"Page size, default 20 and maximum 100" eg:"20"`
	Type     MarketplaceRiskType     `json:"type" dc:"Optional risk type filter: host_service, dynamic_route, menu_permission, external_network, data_table, install_sql, uninstall_sql, mock_sql, dependency, multi_tenant, or docs; include all risk types when empty" eg:"host_service"`
	Severity MarketplaceRiskSeverity `json:"severity" dc:"Optional risk severity filter: info=informational, warning=requires attention, or high=high impact; include all severities when empty" eg:"warning"`
}

// MyReleaseRisksRes is the response for querying publisher-owned release risks.
type MyReleaseRisksRes struct {
	List  []*MarketplaceRiskItem `json:"list" dc:"Paginated publisher-owned release risk findings" eg:"[]"`
	Total int                    `json:"total" dc:"Total owned-release risk findings matching the filters" eg:"3"`
}

// ManagedReleaseRisksReq is the request for querying reviewer-managed release risks.
type ManagedReleaseRisksReq struct {
	g.Meta   `path:"/market/managed-plugins/{pluginId}/releases/{version}/risks" method:"get" tags:"Plugin Marketplace" summary:"Query managed marketplace release risks" permission:"market:plugin:review" dc:"Return paginated scanner findings for reviewer inspection, including findings from submitted and unpublished releases."`
	PluginId string                  `json:"pluginId" v:"required|length:1,64" dc:"Stable plugin ID inspected by the review operator" eg:"linapro-demo-source"`
	Version  string                  `json:"version" v:"required|length:1,32" dc:"Managed release version whose risks are queried" eg:"v0.1.0"`
	PageNum  int                     `json:"pageNum" d:"1" v:"min:1" dc:"Page number starting from 1" eg:"1"`
	PageSize int                     `json:"pageSize" d:"20" v:"min:1|max:100" dc:"Page size, default 20 and maximum 100" eg:"20"`
	Type     MarketplaceRiskType     `json:"type" dc:"Optional risk type filter: host_service, dynamic_route, menu_permission, external_network, data_table, install_sql, uninstall_sql, mock_sql, dependency, multi_tenant, or docs; include all risk types when empty" eg:"host_service"`
	Severity MarketplaceRiskSeverity `json:"severity" dc:"Optional risk severity filter: info=informational, warning=requires attention, or high=high impact; include all severities when empty" eg:"warning"`
}

// ManagedReleaseRisksRes is the response for querying reviewer-managed release risks.
type ManagedReleaseRisksRes struct {
	List  []*MarketplaceRiskItem `json:"list" dc:"Paginated reviewer-managed release risk findings" eg:"[]"`
	Total int                    `json:"total" dc:"Total managed-release risk findings matching the filters" eg:"3"`
}

// MarketplaceDocumentItem is the version-scoped documentation projection returned by the marketplace.
type MarketplaceDocumentItem struct {
	PluginId       string `json:"pluginId" dc:"Stable plugin ID that owns this marketplace document" eg:"linapro-demo-source"`
	Version        string `json:"version" dc:"Release version that owns this marketplace document" eg:"v0.1.0"`
	Locale         string `json:"locale" dc:"Requested or preferred locale used for documentation lookup" eg:"en-US"`
	ResolvedLocale string `json:"resolvedLocale" dc:"Actual locale selected after applying marketplace documentation fallback rules" eg:"zh-CN"`
	Path           string `json:"path" dc:"Document path inside manifest/docs that was returned" eg:"index.md"`
	SourceKind     string `json:"sourceKind" dc:"Document source kind; marketplace documentation responses return manifest_docs entries, or readme when a release has no manifest docs" eg:"manifest_docs"`
	Title          string `json:"title" dc:"Document title extracted during indexing or fallback parsing" eg:"Source Plugin Demo"`
	Summary        string `json:"summary" dc:"Document search summary extracted during indexing" eg:"Learn how to install and use the source plugin demo."`
	Content        string `json:"content" dc:"Safe rendered HTML after script blocking and resource path checks; clients should prefer markdown for full formatting" eg:"<h1>Source Plugin Demo</h1>"`
	Markdown       string `json:"markdown" dc:"Raw Markdown source for the selected document; preferred by workbench clients for client-side Markdown rendering" eg:"# Source Plugin Demo\n"`
	ContentHash    string `json:"contentHash" dc:"Document content hash used by cache keys and stale-content checks" eg:"0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"`
	FallbackUsed   bool   `json:"fallbackUsed" dc:"Whether the response used a fallback locale or another catalog document instead of the requested document" eg:"true"`
	UpdatedAt      *int64 `json:"updatedAt,omitempty" dc:"Document index last updated time as Unix timestamp in milliseconds" eg:"1767247200000"`
}

// MarketplaceDocumentCatalogItem is one navigable document path available on a release.
type MarketplaceDocumentCatalogItem struct {
	Path       string   `json:"path" dc:"Document path inside manifest/docs" eg:"configuration.md"`
	Title      string   `json:"title" dc:"Preferred-locale document title for catalog navigation" eg:"Configuration"`
	SourceKind string   `json:"sourceKind" dc:"Document source kind; catalog entries are manifest_docs documents, or readme when a release has no manifest docs" eg:"manifest_docs"`
	Locales    []string `json:"locales" dc:"Locales that provide this document path" eg:"[\"zh-CN\",\"en-US\"]"`
}

// MarketplaceRiskItem is one review risk finding produced by marketplace scanning.
type MarketplaceRiskItem struct {
	Type      MarketplaceRiskType     `json:"type" dc:"Risk type: host_service, dynamic_route, menu_permission, external_network, data_table, install_sql, uninstall_sql, mock_sql, dependency, multi_tenant, or docs" eg:"host_service"`
	Severity  MarketplaceRiskSeverity `json:"severity" dc:"Risk severity: info, warning, or high" eg:"warning"`
	Source    string                  `json:"source" dc:"Scanner, manifest section, or package resource that produced this risk finding" eg:"plugin.yaml hostServices"`
	Summary   string                  `json:"summary" dc:"Human-readable risk summary produced by marketplace review scanning" eg:"The plugin requests storage read access under reports/."`
	Payload   map[string]any          `json:"payload" dc:"Structured scanner payload for review UI details; shape depends on the risk type" eg:"{}"`
	CreatedAt *int64                  `json:"createdAt,omitempty" dc:"Risk finding creation time as Unix timestamp in milliseconds" eg:"1767240000000"`
}
