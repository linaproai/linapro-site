// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"time"
)

// PluginMarketplaceDisplayI18n is the golang structure for table plugin_marketplace_display_i18n.
type PluginMarketplaceDisplayI18n struct {
	Id             int        `json:"id"             orm:"id"              description:"Primary key ID"`
	ReleaseId      int        `json:"releaseId"      orm:"release_id"      description:"Owning release ID"`
	PluginId       string     `json:"pluginId"       orm:"plugin_id"       description:"Stable plugin ID"`
	ReleaseVersion string     `json:"releaseVersion" orm:"release_version" description:"Release version"`
	Locale         string     `json:"locale"         orm:"locale"          description:"Display locale"`
	Name           string     `json:"name"           orm:"name"            description:"Localized display name"`
	Summary        string     `json:"summary"        orm:"summary"         description:"Localized list summary"`
	Source         string     `json:"source"         orm:"source"          description:"Source: package_i18n/plugin_yaml/publisher"`
	CreatedAt      *time.Time `json:"createdAt"      orm:"created_at"      description:"Creation time"`
	UpdatedAt      *time.Time `json:"updatedAt"      orm:"updated_at"      description:"Update time"`
	DeletedAt      *time.Time `json:"deletedAt"      orm:"deleted_at"      description:"Deletion time"`
}
