// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"time"
)

// PluginMarketplacePluginTag is the golang structure for table plugin_marketplace_plugin_tag.
type PluginMarketplacePluginTag struct {
	Id             int        `json:"id"             orm:"id"               description:"Primary key ID"`
	PluginRecordId int        `json:"pluginRecordId" orm:"plugin_record_id" description:"Owning marketplace plugin record ID"`
	PluginId       string     `json:"pluginId"       orm:"plugin_id"        description:"Stable plugin ID"`
	TagCode        string     `json:"tagCode"        orm:"tag_code"         description:"Stable tag code"`
	CreatedAt      *time.Time `json:"createdAt"      orm:"created_at"       description:"Creation time"`
	UpdatedAt      *time.Time `json:"updatedAt"      orm:"updated_at"       description:"Update time"`
	DeletedAt      *time.Time `json:"deletedAt"      orm:"deleted_at"       description:"Deletion time"`
}
