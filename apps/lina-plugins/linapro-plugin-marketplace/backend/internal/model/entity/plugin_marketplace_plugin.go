// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"time"
)

// PluginMarketplacePlugin is the golang structure for table plugin_marketplace_plugin.
type PluginMarketplacePlugin struct {
	Id              int        `json:"id"              orm:"id"                description:"Primary key ID"`
	PublisherId     int        `json:"publisherId"     orm:"publisher_id"      description:"Owning publisher ID"`
	PluginId        string     `json:"pluginId"        orm:"plugin_id"         description:"Stable plugin ID"`
	Name            string     `json:"name"            orm:"name"              description:"Plugin display name"`
	Summary         string     `json:"summary"         orm:"summary"           description:"Short marketplace summary"`
	Description     string     `json:"description"     orm:"description"       description:"Long marketplace description"`
	PluginType      string     `json:"pluginType"      orm:"plugin_type"       description:"Plugin type: source/dynamic"`
	MarketStatus    string     `json:"marketStatus"    orm:"market_status"     description:"Marketplace status: draft/published/delisted/deprecated"`
	ProcessStatus   string     `json:"processStatus"   orm:"process_status"    description:"Async process status: pending_verify/pending_review/completed/failed"`
	Visibility      string     `json:"visibility"      orm:"visibility"        description:"Visibility policy: public/private/reserved"`
	LatestReleaseId int        `json:"latestReleaseId" orm:"latest_release_id" description:"Latest published release ID"`
	LatestVersion   string     `json:"latestVersion"   orm:"latest_version"    description:"Latest published version"`
	Icon            string     `json:"icon"            orm:"icon"              description:"Marketplace icon path or URL"`
	Homepage        string     `json:"homepage"        orm:"homepage"          description:"Plugin homepage URL"`
	Repository      string     `json:"repository"      orm:"repository"        description:"Plugin source repository URL"`
	License         string     `json:"license"         orm:"license"           description:"Plugin license identifier"`
	DownloadCount   int64      `json:"downloadCount"   orm:"download_count"    description:"Aggregated download count snapshot"`
	SourceKind      string     `json:"sourceKind"      orm:"source_kind"        description:"Publish source kind: git/upload"`
	RepoUrl         string     `json:"repoUrl"         orm:"repo_url"           description:"Git repository URL when source_kind is git"`
	RepoProvider    string     `json:"repoProvider"    orm:"repo_provider"      description:"Git provider: github/gitee, empty for upload"`
	RepoPath        string     `json:"repoPath"        orm:"repo_path"          description:"Plugin root path relative to repository root; empty when repository root is the plugin root"`
	CredentialRef   string     `json:"credentialRef"   orm:"credential_ref"     description:"Opaque credential reference for private Git access, empty when public"`
	LastSyncAt      *time.Time `json:"lastSyncAt"      orm:"last_sync_at"       description:"Last Git metadata discovery time"`
	LastSyncStatus  string     `json:"lastSyncStatus"  orm:"last_sync_status"   description:"Last Git sync status"`
	LastSyncMessage string     `json:"lastSyncMessage" orm:"last_sync_message"  description:"Last Git sync diagnostic message without secrets"`
	PublishedAt     *time.Time `json:"publishedAt"     orm:"published_at"      description:"First published time"`
	CreatedAt       *time.Time `json:"createdAt"       orm:"created_at"        description:"Creation time"`
	UpdatedAt       *time.Time `json:"updatedAt"       orm:"updated_at"        description:"Update time"`
	DeletedAt       *time.Time `json:"deletedAt"       orm:"deleted_at"        description:"Deletion time"`
}
