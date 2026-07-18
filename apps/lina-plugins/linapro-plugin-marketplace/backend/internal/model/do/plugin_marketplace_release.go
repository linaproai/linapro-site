// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

// PluginMarketplaceRelease is the golang structure of table plugin_marketplace_release for DAO operations like Where/Data.
type PluginMarketplaceRelease struct {
	g.Meta             `orm:"table:plugin_marketplace_release, do:true"`
	Id                 any        // Primary key ID
	PluginRecordId     any        // Owning marketplace plugin record ID
	PublisherId        any        // Owning publisher ID
	PluginId           any        // Stable plugin ID
	ReleaseVersion     any        // Plugin release version
	SourceRef          any        // Git logical tag or branch name for git-sourced releases, empty for upload packages
	SourceCommit       any        // Pinned full commit SHA resolved during Git discovery, empty for upload packages
	PluginType         any        // Plugin type: source/dynamic
	ReleaseStatus      any        // Release status: draft/published/delisted/deprecated
	ReviewStatus       any        // Review status: draft/submitted/reviewing/approved/rejected
	ProcessStatus      any        // Async process status: pending_verify/pending_review/completed/failed
	Visibility         any        // Release visibility policy
	MinHostVersion     any        // Minimum compatible LinaPro host version
	MaxHostVersion     any        // Maximum compatible LinaPro host version
	ManifestSnapshot   any        // Parsed plugin.yaml snapshot
	DependencySummary  any        // Dependency scan summary
	HostServiceSummary any        // Host service scan summary
	RouteSummary       any        // Route scan summary
	SqlSummary         any        // SQL resource scan summary
	I18NSummary        any        // i18n resource scan summary
	DocsSummary        any        // Marketplace document scan summary
	RiskSummary        any        // Aggregated review risk summary
	ReviewMessage      any        // Latest review message
	SubmittedAt        *time.Time // Review submission time
	ReviewedAt         *time.Time // Review completion time
	PublishedAt        *time.Time // Publish time
	CreatedAt          *time.Time // Creation time
	UpdatedAt          *time.Time // Update time
	DeletedAt          *time.Time // Deletion time
}
