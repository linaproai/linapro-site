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
	Visibility      string     `json:"visibility"      orm:"visibility"        description:"Visibility policy: public/private/reserved"`
	LatestReleaseId int        `json:"latestReleaseId" orm:"latest_release_id" description:"Latest published release ID"`
	LatestVersion   string     `json:"latestVersion"   orm:"latest_version"    description:"Latest published version"`
	Icon            string     `json:"icon"            orm:"icon"              description:"Marketplace icon path or URL"`
	Homepage        string     `json:"homepage"        orm:"homepage"          description:"Plugin homepage URL"`
	Repository      string     `json:"repository"      orm:"repository"        description:"Plugin source repository URL"`
	License         string     `json:"license"         orm:"license"           description:"Plugin license identifier"`
	DownloadCount   int64      `json:"downloadCount"   orm:"download_count"    description:"Aggregated download count snapshot"`
	PublishedAt     *time.Time `json:"publishedAt"     orm:"published_at"      description:"First published time"`
	CreatedAt       *time.Time `json:"createdAt"       orm:"created_at"        description:"Creation time"`
	UpdatedAt       *time.Time `json:"updatedAt"       orm:"updated_at"        description:"Update time"`
	DeletedAt       *time.Time `json:"deletedAt"       orm:"deleted_at"        description:"Deletion time"`
}
