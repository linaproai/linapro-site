// This file defines shared marketplace API value types, enum constants, and
// reusable response projections. It keeps public DTOs independent from DAO,
// DO, and entity models so controller and service layers can project only the
// fields required by each endpoint.

package v1

// MarketplacePluginType identifies the implementation family published in the marketplace.
type MarketplacePluginType string

// Marketplace plugin implementation families.
const (
	MarketplacePluginTypeSource  MarketplacePluginType = "source"
	MarketplacePluginTypeDynamic MarketplacePluginType = "dynamic"
)

// String returns the serialized marketplace plugin type value.
func (value MarketplacePluginType) String() string { return string(value) }

// MarketplaceStatus identifies the marketplace lifecycle state of a plugin or release.
type MarketplaceStatus string

// Marketplace plugin and release lifecycle states.
const (
	MarketplaceStatusDraft      MarketplaceStatus = "draft"
	MarketplaceStatusPublished  MarketplaceStatus = "published"
	MarketplaceStatusDelisted   MarketplaceStatus = "delisted"
	MarketplaceStatusDeprecated MarketplaceStatus = "deprecated"
)

// String returns the serialized marketplace status value.
func (value MarketplaceStatus) String() string { return string(value) }

// MarketplaceReviewStatus identifies the review state of one marketplace release.
type MarketplaceReviewStatus string

// Marketplace release review states.
const (
	MarketplaceReviewStatusDraft     MarketplaceReviewStatus = "draft"
	MarketplaceReviewStatusSubmitted MarketplaceReviewStatus = "submitted"
	MarketplaceReviewStatusReviewing MarketplaceReviewStatus = "reviewing"
	MarketplaceReviewStatusApproved  MarketplaceReviewStatus = "approved"
	MarketplaceReviewStatusRejected  MarketplaceReviewStatus = "rejected"
)

// String returns the serialized marketplace review status value.
func (value MarketplaceReviewStatus) String() string { return string(value) }

// MarketplaceArtifactType identifies the artifact stored for one marketplace release.
type MarketplaceArtifactType string

// Marketplace release artifact types.
const (
	MarketplaceArtifactTypeSourceZip     MarketplaceArtifactType = "source_zip"
	MarketplaceArtifactTypeSourceTarGz   MarketplaceArtifactType = "source_tar_gz"
	MarketplaceArtifactTypeDynamicZip    MarketplaceArtifactType = "dynamic_zip"
	MarketplaceArtifactTypeDynamicTarGz  MarketplaceArtifactType = "dynamic_tar_gz"
	MarketplaceArtifactTypePluginWasm    MarketplaceArtifactType = "plugin_wasm"
)

// MarketplaceSourceKind identifies how a marketplace plugin is published.
type MarketplaceSourceKind string

// Marketplace publish source kinds.
const (
	MarketplaceSourceKindGit    MarketplaceSourceKind = "git"
	MarketplaceSourceKindUpload MarketplaceSourceKind = "upload"
)

// String returns the serialized marketplace source kind value.
func (value MarketplaceSourceKind) String() string { return string(value) }

// MarketplaceDistributionMode identifies how consumers obtain one release.
type MarketplaceDistributionMode string

// Marketplace distribution modes returned to CLI installers.
const (
	MarketplaceDistributionModeGit   MarketplaceDistributionMode = "git"
	MarketplaceDistributionModeHTTPS MarketplaceDistributionMode = "https"
)

// String returns the serialized marketplace distribution mode value.
func (value MarketplaceDistributionMode) String() string { return string(value) }

// MarketplaceRepoProvider identifies a supported Git hosting provider.
type MarketplaceRepoProvider string

// Supported Git hosting providers for marketplace Git sources.
const (
	MarketplaceRepoProviderGitHub MarketplaceRepoProvider = "github"
	MarketplaceRepoProviderGitee  MarketplaceRepoProvider = "gitee"
)

// String returns the serialized repository provider value.
func (value MarketplaceRepoProvider) String() string { return string(value) }

// MarketplaceDistributionItem is the CLI-facing install projection for one release.
type MarketplaceDistributionItem struct {
	Mode                    MarketplaceDistributionMode `json:"mode" dc:"Distribution mode: git=clone by ref https=download package through controlled session" eg:"git"`
	PluginId                string                      `json:"pluginId" dc:"Stable plugin ID bound to the distribution projection" eg:"linapro-demo-source"`
	Version                 string                      `json:"version" dc:"Release version bound to the distribution projection" eg:"v1.0.0"`
	PluginType              MarketplacePluginType       `json:"pluginType" dc:"Plugin type: source or dynamic" eg:"source"`
	RepoUrl                 string                      `json:"repoUrl,omitempty" dc:"Git repository URL when mode is git; empty for https packages" eg:"https://github.com/org/plugin.git"`
	Ref                     string                      `json:"ref,omitempty" dc:"Git tag or ref when mode is git" eg:"v1.0.0"`
	Provider                MarketplaceRepoProvider     `json:"provider,omitempty" dc:"Git provider when mode is git: github or gitee" eg:"github"`
	RequiresAuth            bool                        `json:"requiresAuth,omitempty" dc:"Whether private Git access requires caller credentials; platform tokens are never returned" eg:"true"`
	ArtifactType            MarketplaceArtifactType     `json:"artifactType,omitempty" dc:"Primary package artifact type when mode is https" eg:"source_zip"`
	Sha256                  string                      `json:"sha256,omitempty" dc:"Primary package SHA-256 when mode is https" eg:"0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"`
	SizeBytes               int64                       `json:"sizeBytes,omitempty" dc:"Primary package size in bytes when mode is https" eg:"102400"`
	DownloadSessionRequired bool                        `json:"downloadSessionRequired,omitempty" dc:"Whether clients must create a download session before streaming package bytes" eg:"true"`
}

// String returns the serialized marketplace artifact type value.
func (value MarketplaceArtifactType) String() string { return string(value) }

// MarketplaceVisibility identifies the marketplace visibility policy.
type MarketplaceVisibility string

// Marketplace visibility policies.
const (
	MarketplaceVisibilityPublic   MarketplaceVisibility = "public"
	MarketplaceVisibilityPrivate  MarketplaceVisibility = "private"
	MarketplaceVisibilityReserved MarketplaceVisibility = "reserved"
)

// String returns the serialized marketplace visibility value.
func (value MarketplaceVisibility) String() string { return string(value) }

// MarketplaceRiskType identifies one review risk finding category.
type MarketplaceRiskType string

// Marketplace review risk finding categories.
const (
	MarketplaceRiskTypeHostService     MarketplaceRiskType = "host_service"
	MarketplaceRiskTypeDynamicRoute    MarketplaceRiskType = "dynamic_route"
	MarketplaceRiskTypeMenuPermission  MarketplaceRiskType = "menu_permission"
	MarketplaceRiskTypeExternalNetwork MarketplaceRiskType = "external_network"
	MarketplaceRiskTypeDataTable       MarketplaceRiskType = "data_table"
	MarketplaceRiskTypeInstallSQL      MarketplaceRiskType = "install_sql"
	MarketplaceRiskTypeUninstallSQL    MarketplaceRiskType = "uninstall_sql"
	MarketplaceRiskTypeMockSQL         MarketplaceRiskType = "mock_sql"
	MarketplaceRiskTypeDependency      MarketplaceRiskType = "dependency"
	MarketplaceRiskTypeMultiTenant     MarketplaceRiskType = "multi_tenant"
	MarketplaceRiskTypeDocs            MarketplaceRiskType = "docs"
)

// String returns the serialized marketplace risk type value.
func (value MarketplaceRiskType) String() string { return string(value) }

// MarketplaceRiskSeverity identifies the severity of one review risk finding.
type MarketplaceRiskSeverity string

// Marketplace review risk severities.
const (
	MarketplaceRiskSeverityInfo    MarketplaceRiskSeverity = "info"
	MarketplaceRiskSeverityWarning MarketplaceRiskSeverity = "warning"
	MarketplaceRiskSeverityHigh    MarketplaceRiskSeverity = "high"
)

// String returns the serialized marketplace risk severity value.
func (value MarketplaceRiskSeverity) String() string { return string(value) }

// MarketplaceDownloadSessionStatus identifies one short-lived download session state.
type MarketplaceDownloadSessionStatus string

// Marketplace download session states.
const (
	MarketplaceDownloadSessionStatusActive   MarketplaceDownloadSessionStatus = "active"
	MarketplaceDownloadSessionStatusExpired  MarketplaceDownloadSessionStatus = "expired"
	MarketplaceDownloadSessionStatusConsumed MarketplaceDownloadSessionStatus = "consumed"
	MarketplaceDownloadSessionStatusRevoked  MarketplaceDownloadSessionStatus = "revoked"
)

// String returns the serialized marketplace download session status value.
func (value MarketplaceDownloadSessionStatus) String() string { return string(value) }

// MarketplacePublisherItem is the lightweight publisher projection used by catalog and publish APIs.
type MarketplacePublisherItem struct {
	PublisherKey string `json:"publisherKey" dc:"Stable publisher key used for ownership and filtering" eg:"linapro"`
	Name         string `json:"name" dc:"Publisher display name shown in marketplace catalog pages" eg:"LinaPro"`
	Summary      string `json:"summary" dc:"Short publisher summary for marketplace trust and discovery display" eg:"Official LinaPro plugin publisher"`
	Verified     bool   `json:"verified" dc:"Whether the marketplace publisher has been verified by the review authority" eg:"true"`
	Homepage     string `json:"homepage" dc:"Publisher homepage URL, empty when the publisher does not provide one" eg:"https://linapro.ai"`
	ContactEmail string `json:"contactEmail" dc:"Publisher contact email used by marketplace reviewers, empty when unavailable" eg:"plugins@linapro.ai"`
}

// MarketplaceTagItem is the lightweight marketplace category or tag projection.
type MarketplaceTagItem struct {
	Code string `json:"code" dc:"Stable category or tag code used by filters and read-model snapshots" eg:"observability"`
	Name string `json:"name" dc:"Localized or configured category or tag display name" eg:"Observability"`
	Type string `json:"type" dc:"Tag type: category=primary catalog category tag=secondary descriptive tag" eg:"category"`
}

// MarketplaceRiskCounts summarizes risk finding counts by severity.
type MarketplaceRiskCounts struct {
	Info    int `json:"info" dc:"Number of informational risk findings in the latest release review snapshot" eg:"2"`
	Warning int `json:"warning" dc:"Number of warning risk findings in the latest release review snapshot" eg:"1"`
	High    int `json:"high" dc:"Number of high severity risk findings in the latest release review snapshot" eg:"0"`
}

// MarketplaceArtifactItem is the artifact checksum and size projection returned to clients.
type MarketplaceArtifactItem struct {
	ArtifactType   MarketplaceArtifactType `json:"artifactType" dc:"Artifact type: source_zip=source plugin package dynamic_zip=dynamic runtime package plugin_wasm=extracted dynamic wasm artifact" eg:"source_zip"`
	FileName       string                  `json:"fileName" dc:"Original artifact file name uploaded by the publisher" eg:"linapro-demo-source-v0.1.0.zip"`
	ContentType    string                  `json:"contentType" dc:"Artifact MIME content type captured during upload" eg:"application/zip"`
	SizeBytes      int64                   `json:"sizeBytes" dc:"Artifact size in bytes" eg:"102400"`
	Sha256         string                  `json:"sha256" dc:"Artifact SHA-256 checksum used by clients to verify downloads" eg:"0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"`
	ManifestSha256 string                  `json:"manifestSha256,omitempty" dc:"Root plugin.yaml SHA-256 checksum when the uploaded package contains a root manifest" eg:"abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789"`
	WasmSha256     string                  `json:"wasmSha256,omitempty" dc:"Extracted plugin.wasm SHA-256 checksum for dynamic runtime packages" eg:"fedcba9876543210fedcba9876543210fedcba9876543210fedcba9876543210"`
}

// MarketplaceReleaseItem is the release projection shared by catalog, review, and publish APIs.
type MarketplaceReleaseItem struct {
	PluginId       string                   `json:"pluginId" dc:"Stable plugin ID owned by the marketplace record" eg:"linapro-demo-source"`
	Version        string                   `json:"version" dc:"Immutable release version; a published pluginId and version pair cannot be overwritten" eg:"v0.1.0"`
	PluginType     MarketplacePluginType    `json:"pluginType" dc:"Plugin type: source=source plugin dynamic=dynamic plugin" eg:"source"`
	ReleaseStatus  MarketplaceStatus        `json:"releaseStatus" dc:"Release lifecycle status: draft, published, delisted, or deprecated" eg:"published"`
	ReviewStatus   MarketplaceReviewStatus  `json:"reviewStatus" dc:"Release review status: draft, submitted, reviewing, approved, or rejected" eg:"approved"`
	Visibility     MarketplaceVisibility    `json:"visibility" dc:"Release visibility policy: public, private, or reserved" eg:"public"`
	MinHostVersion string                   `json:"minHostVersion" dc:"Minimum compatible LinaPro host version declared or inferred for this release" eg:"v0.1.0"`
	MaxHostVersion string                   `json:"maxHostVersion" dc:"Maximum compatible LinaPro host version, empty when no upper bound is declared" eg:"v1.0.0"`
	ReviewMessage  string                       `json:"reviewMessage" dc:"Latest reviewer message or scanner diagnostic summary, empty when no message exists" eg:"Approved for marketplace listing"`
	SourceRef      string                       `json:"sourceRef,omitempty" dc:"Git tag or ref for git-sourced releases" eg:"v0.1.0"`
	Distribution   *MarketplaceDistributionItem `json:"distribution,omitempty" dc:"CLI install distribution projection when available for this release" eg:"{}"`
	Artifact       *MarketplaceArtifactItem     `json:"artifact,omitempty" dc:"Primary artifact checksum and size summary for this release" eg:"{}"`
	SubmittedAt    *int64                       `json:"submittedAt,omitempty" dc:"Review submission time as Unix timestamp in milliseconds" eg:"1767240000000"`
	ReviewedAt     *int64                       `json:"reviewedAt,omitempty" dc:"Review completion time as Unix timestamp in milliseconds" eg:"1767243600000"`
	PublishedAt    *int64                       `json:"publishedAt,omitempty" dc:"Publish time as Unix timestamp in milliseconds" eg:"1767247200000"`
	UpdatedAt      *int64                       `json:"updatedAt,omitempty" dc:"Release last updated time as Unix timestamp in milliseconds" eg:"1767247200000"`
}
