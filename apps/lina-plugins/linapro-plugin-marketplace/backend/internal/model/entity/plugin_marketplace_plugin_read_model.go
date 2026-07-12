// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"time"
)

// PluginMarketplacePluginReadModel is the golang structure for table plugin_marketplace_plugin_read_model.
type PluginMarketplacePluginReadModel struct {
	Id                int        `json:"id"                orm:"id"                 description:"Primary key ID"`
	PluginRecordId    int        `json:"pluginRecordId"    orm:"plugin_record_id"   description:"Owning marketplace plugin record ID"`
	PublisherId       int        `json:"publisherId"       orm:"publisher_id"       description:"Owning publisher ID"`
	PublisherName     string     `json:"publisherName"     orm:"publisher_name"     description:"Publisher display name snapshot"`
	PublisherVerified bool       `json:"publisherVerified" orm:"publisher_verified" description:"Publisher verification snapshot"`
	PluginId          string     `json:"pluginId"          orm:"plugin_id"          description:"Stable plugin ID"`
	Name              string     `json:"name"              orm:"name"               description:"Plugin display name"`
	Summary           string     `json:"summary"           orm:"summary"            description:"Short marketplace summary"`
	PluginType        string     `json:"pluginType"        orm:"plugin_type"        description:"Plugin type: source/dynamic"`
	MarketStatus      string     `json:"marketStatus"      orm:"market_status"      description:"Marketplace status"`
	Visibility        string     `json:"visibility"        orm:"visibility"         description:"Visibility policy"`
	LatestReleaseId   int        `json:"latestReleaseId"   orm:"latest_release_id"  description:"Latest published release ID"`
	LatestVersion     string     `json:"latestVersion"     orm:"latest_version"     description:"Latest published version"`
	MinHostVersion    string     `json:"minHostVersion"    orm:"min_host_version"   description:"Minimum compatible LinaPro host version"`
	MaxHostVersion    string     `json:"maxHostVersion"    orm:"max_host_version"   description:"Maximum compatible LinaPro host version"`
	PrimaryTag        string     `json:"primaryTag"        orm:"primary_tag"        description:"Primary category tag code"`
	TagCodes          string     `json:"tagCodes"          orm:"tag_codes"          description:"Tag code snapshot for display"`
	RiskCounts        string     `json:"riskCounts"        orm:"risk_counts"        description:"Risk count snapshot grouped by severity"`
	DownloadCount     int64      `json:"downloadCount"     orm:"download_count"     description:"Aggregated download count snapshot"`
	PublishedAt       *time.Time `json:"publishedAt"       orm:"published_at"       description:"Latest publish time"`
	SearchText        string     `json:"searchText"        orm:"search_text"        description:"Plain text projection used for catalog search"`
	CreatedAt         *time.Time `json:"createdAt"         orm:"created_at"         description:"Creation time"`
	UpdatedAt         *time.Time `json:"updatedAt"         orm:"updated_at"         description:"Update time"`
	DeletedAt         *time.Time `json:"deletedAt"         orm:"deleted_at"         description:"Deletion time"`
}
