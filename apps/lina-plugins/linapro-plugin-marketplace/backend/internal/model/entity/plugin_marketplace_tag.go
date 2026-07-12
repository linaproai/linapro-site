// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"time"
)

// PluginMarketplaceTag is the golang structure for table plugin_marketplace_tag.
type PluginMarketplaceTag struct {
	Id        int        `json:"id"        orm:"id"         description:"Primary key ID"`
	TagCode   string     `json:"tagCode"   orm:"tag_code"   description:"Stable tag code"`
	Name      string     `json:"name"      orm:"name"       description:"Tag display name"`
	TagType   string     `json:"tagType"   orm:"tag_type"   description:"Tag type: category/tag"`
	Sort      int        `json:"sort"      orm:"sort"       description:"Display order"`
	Status    int        `json:"status"    orm:"status"     description:"Status: 0=disabled, 1=enabled"`
	CreatedAt *time.Time `json:"createdAt" orm:"created_at" description:"Creation time"`
	UpdatedAt *time.Time `json:"updatedAt" orm:"updated_at" description:"Update time"`
	DeletedAt *time.Time `json:"deletedAt" orm:"deleted_at" description:"Deletion time"`
}
