// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"time"
)

// PluginMarketplaceRisk is the golang structure for table plugin_marketplace_risk.
type PluginMarketplaceRisk struct {
	Id             int        `json:"id"             orm:"id"              description:"Primary key ID"`
	ReleaseId      int        `json:"releaseId"      orm:"release_id"      description:"Owning release ID"`
	PluginId       string     `json:"pluginId"       orm:"plugin_id"       description:"Stable plugin ID"`
	ReleaseVersion string     `json:"releaseVersion" orm:"release_version" description:"Plugin release version"`
	RiskType       string     `json:"riskType"       orm:"risk_type"       description:"Risk type: host_service/dynamic_route/menu_permission/external_network/data_table/install_sql/uninstall_sql/mock_sql/dependency/multi_tenant/docs"`
	Severity       string     `json:"severity"       orm:"severity"        description:"Risk severity: info/warning/high"`
	Source         string     `json:"source"         orm:"source"          description:"Scanner or resource source"`
	Summary        string     `json:"summary"        orm:"summary"         description:"Human-readable risk summary"`
	Payload        string     `json:"payload"        orm:"payload"         description:"Structured scanner payload"`
	CreatedAt      *time.Time `json:"createdAt"      orm:"created_at"      description:"Creation time"`
	UpdatedAt      *time.Time `json:"updatedAt"      orm:"updated_at"      description:"Update time"`
	DeletedAt      *time.Time `json:"deletedAt"      orm:"deleted_at"      description:"Deletion time"`
}
