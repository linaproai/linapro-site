// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

// PluginMarketplacePublisher is the golang structure of table plugin_marketplace_publisher for DAO operations like Where/Data.
type PluginMarketplacePublisher struct {
	g.Meta       `orm:"table:plugin_marketplace_publisher, do:true"`
	Id           any        // Primary key ID
	PublisherKey any        // Stable publisher key
	Name         any        // Publisher display name
	Summary      any        // Publisher summary
	OwnerUserId  any        // Owning user ID
	OwnerOrgId   any        // Owning organization ID, 0 means none
	Verified     any        // Whether the publisher has been verified
	Status       any        // Publisher status: active/suspended
	Homepage     any        // Publisher homepage URL
	ContactEmail any        // Publisher contact email
	Remark       any        // Remark
	CreatedAt    *time.Time // Creation time
	UpdatedAt    *time.Time // Update time
	DeletedAt    *time.Time // Deletion time
}
