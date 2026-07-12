// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

// PluginMarketplaceTag is the golang structure of table plugin_marketplace_tag for DAO operations like Where/Data.
type PluginMarketplaceTag struct {
	g.Meta    `orm:"table:plugin_marketplace_tag, do:true"`
	Id        any        // Primary key ID
	TagCode   any        // Stable tag code
	Name      any        // Tag display name
	TagType   any        // Tag type: category/tag
	Sort      any        // Display order
	Status    any        // Status: 0=disabled, 1=enabled
	CreatedAt *time.Time // Creation time
	UpdatedAt *time.Time // Update time
	DeletedAt *time.Time // Deletion time
}
