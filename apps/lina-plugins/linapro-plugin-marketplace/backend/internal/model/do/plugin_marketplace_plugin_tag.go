// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

// PluginMarketplacePluginTag is the golang structure of table plugin_marketplace_plugin_tag for DAO operations like Where/Data.
type PluginMarketplacePluginTag struct {
	g.Meta         `orm:"table:plugin_marketplace_plugin_tag, do:true"`
	Id             any        // Primary key ID
	PluginRecordId any        // Owning marketplace plugin record ID
	PluginId       any        // Stable plugin ID
	TagCode        any        // Stable tag code
	CreatedAt      *time.Time // Creation time
	UpdatedAt      *time.Time // Update time
	DeletedAt      *time.Time // Deletion time
}
