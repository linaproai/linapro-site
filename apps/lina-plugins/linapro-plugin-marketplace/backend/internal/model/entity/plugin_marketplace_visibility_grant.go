// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"time"
)

// PluginMarketplaceVisibilityGrant is the golang structure for table plugin_marketplace_visibility_grant.
type PluginMarketplaceVisibilityGrant struct {
	Id             int        `json:"id"             orm:"id"               description:"Primary key ID"`
	PluginRecordId int        `json:"pluginRecordId" orm:"plugin_record_id" description:"Owning marketplace plugin record ID"`
	PluginId       string     `json:"pluginId"       orm:"plugin_id"        description:"Stable plugin ID"`
	ScopeType      string     `json:"scopeType"      orm:"scope_type"       description:"Visibility scope type: public/tenant/org/user/reserved_license"`
	ScopeId        string     `json:"scopeId"        orm:"scope_id"         description:"Scope identifier, empty for public scope"`
	Permission     string     `json:"permission"     orm:"permission"       description:"Permission covered by the grant: view/download"`
	Status         int        `json:"status"         orm:"status"           description:"Status: 0=disabled, 1=enabled"`
	ExpiresAt      *time.Time `json:"expiresAt"      orm:"expires_at"       description:"Grant expiration time"`
	CreatedAt      *time.Time `json:"createdAt"      orm:"created_at"       description:"Creation time"`
	UpdatedAt      *time.Time `json:"updatedAt"      orm:"updated_at"       description:"Update time"`
	DeletedAt      *time.Time `json:"deletedAt"      orm:"deleted_at"       description:"Deletion time"`
}
