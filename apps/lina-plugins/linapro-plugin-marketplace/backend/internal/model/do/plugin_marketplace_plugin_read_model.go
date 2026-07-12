// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

// PluginMarketplacePluginReadModel is the golang structure of table plugin_marketplace_plugin_read_model for DAO operations like Where/Data.
type PluginMarketplacePluginReadModel struct {
	g.Meta            `orm:"table:plugin_marketplace_plugin_read_model, do:true"`
	Id                any        // Primary key ID
	PluginRecordId    any        // Owning marketplace plugin record ID
	PublisherId       any        // Owning publisher ID
	PublisherName     any        // Publisher display name snapshot
	PublisherVerified any        // Publisher verification snapshot
	PluginId          any        // Stable plugin ID
	Name              any        // Plugin display name
	Summary           any        // Short marketplace summary
	PluginType        any        // Plugin type: source/dynamic
	MarketStatus      any        // Marketplace status
	Visibility        any        // Visibility policy
	LatestReleaseId   any        // Latest published release ID
	LatestVersion     any        // Latest published version
	MinHostVersion    any        // Minimum compatible LinaPro host version
	MaxHostVersion    any        // Maximum compatible LinaPro host version
	PrimaryTag        any        // Primary category tag code
	TagCodes          any        // Tag code snapshot for display
	RiskCounts        any        // Risk count snapshot grouped by severity
	DownloadCount     any        // Aggregated download count snapshot
	PublishedAt       *time.Time // Latest publish time
	SearchText        any        // Plain text projection used for catalog search
	CreatedAt         *time.Time // Creation time
	UpdatedAt         *time.Time // Update time
	DeletedAt         *time.Time // Deletion time
}
