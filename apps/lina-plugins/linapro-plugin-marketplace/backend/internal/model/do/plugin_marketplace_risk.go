// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

// PluginMarketplaceRisk is the golang structure of table plugin_marketplace_risk for DAO operations like Where/Data.
type PluginMarketplaceRisk struct {
	g.Meta         `orm:"table:plugin_marketplace_risk, do:true"`
	Id             any        // Primary key ID
	ReleaseId      any        // Owning release ID
	PluginId       any        // Stable plugin ID
	ReleaseVersion any        // Plugin release version
	RiskType       any        // Risk type: host_service/dynamic_route/menu_permission/external_network/data_table/install_sql/uninstall_sql/mock_sql/dependency/multi_tenant/docs
	Severity       any        // Risk severity: info/warning/high
	Source         any        // Scanner or resource source
	Summary        any        // Human-readable risk summary
	Payload        any        // Structured scanner payload
	CreatedAt      *time.Time // Creation time
	UpdatedAt      *time.Time // Update time
	DeletedAt      *time.Time // Deletion time
}
