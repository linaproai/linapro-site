// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

// PluginMarketplacePlugin is the golang structure of table plugin_marketplace_plugin for DAO operations like Where/Data.
type PluginMarketplacePlugin struct {
	g.Meta          `orm:"table:plugin_marketplace_plugin, do:true"`
	Id              any        // Primary key ID
	PublisherId     any        // Owning publisher ID
	PluginId        any        // Stable plugin ID
	Name            any        // Plugin display name
	Summary         any        // Short marketplace summary
	Description     any        // Long marketplace description
	PluginType      any        // Plugin type: source/dynamic
	MarketStatus    any        // Marketplace status: draft/published/delisted/deprecated
	ProcessStatus   any        // Async process status: pending_verify/pending_review/completed/failed
	Visibility      any        // Visibility policy: public/private/reserved
	LatestReleaseId any        // Latest published release ID
	LatestVersion   any        // Latest published version
	Icon            any        // Marketplace icon path or URL
	Homepage        any        // Plugin homepage URL
	Repository      any        // Plugin source repository URL
	License         any        // Plugin license identifier
	DownloadCount   any        // Aggregated download count snapshot
	SourceKind      any        // Publish source kind: git/upload
	RepoUrl         any        // Git repository URL when source_kind is git
	RepoProvider    any        // Git provider: github/gitee, empty for upload
	RepoPath        any        // Plugin root path relative to repository root; empty when repository root is the plugin root
	CredentialRef   any        // Opaque credential reference for private Git access, empty when public
	LastSyncAt      *time.Time // Last Git metadata discovery time
	LastSyncStatus  any        // Last Git sync status
	LastSyncMessage any        // Last Git sync diagnostic message without secrets
	PublishedAt     *time.Time // First published time
	CreatedAt       *time.Time // Creation time
	UpdatedAt       *time.Time // Update time
	DeletedAt       *time.Time // Deletion time
}
