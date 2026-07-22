// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

// PluginMarketplaceDisplayI18n is the golang structure of table plugin_marketplace_display_i18n for DAO operations like Where/Data.
type PluginMarketplaceDisplayI18n struct {
	g.Meta         `orm:"table:plugin_marketplace_display_i18n, do:true"`
	Id             any        // Primary key ID
	ReleaseId      any        // Owning release ID
	PluginId       any        // Stable plugin ID
	ReleaseVersion any        // Release version
	Locale         any        // Display locale
	Name           any        // Localized display name
	Summary        any        // Localized list summary
	Source         any        // Source: package_i18n/plugin_yaml/publisher
	CreatedAt      *time.Time // Creation time
	UpdatedAt      *time.Time // Update time
	DeletedAt      *time.Time // Deletion time
}
