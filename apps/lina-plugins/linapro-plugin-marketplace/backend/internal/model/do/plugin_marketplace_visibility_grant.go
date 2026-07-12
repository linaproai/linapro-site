// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

// PluginMarketplaceVisibilityGrant is the golang structure of table plugin_marketplace_visibility_grant for DAO operations like Where/Data.
type PluginMarketplaceVisibilityGrant struct {
	g.Meta         `orm:"table:plugin_marketplace_visibility_grant, do:true"`
	Id             any        // Primary key ID
	PluginRecordId any        // Owning marketplace plugin record ID
	PluginId       any        // Stable plugin ID
	ScopeType      any        // Visibility scope type: public/tenant/org/user/reserved_license
	ScopeId        any        // Scope identifier, empty for public scope
	Permission     any        // Permission covered by the grant: view/download
	Status         any        // Status: 0=disabled, 1=enabled
	ExpiresAt      *time.Time // Grant expiration time
	CreatedAt      *time.Time // Creation time
	UpdatedAt      *time.Time // Update time
	DeletedAt      *time.Time // Deletion time
}
