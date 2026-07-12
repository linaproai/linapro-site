// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// PluginMarketplaceRiskDao is the data access object for the table plugin_marketplace_risk.
type PluginMarketplaceRiskDao struct {
	table    string                       // table is the underlying table name of the DAO.
	group    string                       // group is the database configuration group name of the current DAO.
	columns  PluginMarketplaceRiskColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler           // handlers for customized model modification.
}

// PluginMarketplaceRiskColumns defines and stores column names for the table plugin_marketplace_risk.
type PluginMarketplaceRiskColumns struct {
	Id             string // Primary key ID
	ReleaseId      string // Owning release ID
	PluginId       string // Stable plugin ID
	ReleaseVersion string // Plugin release version
	RiskType       string // Risk type: host_service/dynamic_route/menu_permission/external_network/data_table/install_sql/uninstall_sql/mock_sql/dependency/multi_tenant/docs
	Severity       string // Risk severity: info/warning/high
	Source         string // Scanner or resource source
	Summary        string // Human-readable risk summary
	Payload        string // Structured scanner payload
	CreatedAt      string // Creation time
	UpdatedAt      string // Update time
	DeletedAt      string // Deletion time
}

// pluginMarketplaceRiskColumns holds the columns for the table plugin_marketplace_risk.
var pluginMarketplaceRiskColumns = PluginMarketplaceRiskColumns{
	Id:             "id",
	ReleaseId:      "release_id",
	PluginId:       "plugin_id",
	ReleaseVersion: "release_version",
	RiskType:       "risk_type",
	Severity:       "severity",
	Source:         "source",
	Summary:        "summary",
	Payload:        "payload",
	CreatedAt:      "created_at",
	UpdatedAt:      "updated_at",
	DeletedAt:      "deleted_at",
}

// NewPluginMarketplaceRiskDao creates and returns a new DAO object for table data access.
func NewPluginMarketplaceRiskDao(handlers ...gdb.ModelHandler) *PluginMarketplaceRiskDao {
	return &PluginMarketplaceRiskDao{
		group:    "default",
		table:    "plugin_marketplace_risk",
		columns:  pluginMarketplaceRiskColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *PluginMarketplaceRiskDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *PluginMarketplaceRiskDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *PluginMarketplaceRiskDao) Columns() PluginMarketplaceRiskColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *PluginMarketplaceRiskDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *PluginMarketplaceRiskDao) Ctx(ctx context.Context) *gdb.Model {
	model := dao.DB().Model(dao.table)
	for _, handler := range dao.handlers {
		model = handler(model)
	}
	return model.Safe().Ctx(ctx)
}

// Transaction wraps the transaction logic using function f.
// It rolls back the transaction and returns the error if function f returns a non-nil error.
// It commits the transaction and returns nil if function f returns nil.
//
// Note: Do not commit or roll back the transaction in function f,
// as it is automatically handled by this function.
func (dao *PluginMarketplaceRiskDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
