// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// PluginMarketplaceDisplayI18nDao is the data access object for the table plugin_marketplace_display_i18n.
type PluginMarketplaceDisplayI18nDao struct {
	table    string                               // table is the underlying table name of the DAO.
	group    string                               // group is the database configuration group name of the current DAO.
	columns  PluginMarketplaceDisplayI18nColumns  // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler                   // handlers for customized model modification.
}

// PluginMarketplaceDisplayI18nColumns defines and stores column names for the table plugin_marketplace_display_i18n.
type PluginMarketplaceDisplayI18nColumns struct {
	Id             string // Primary key ID
	ReleaseId      string // Owning release ID
	PluginId       string // Stable plugin ID
	ReleaseVersion string // Release version
	Locale         string // Display locale
	Name           string // Localized display name
	Summary        string // Localized list summary
	Source         string // Source: package_i18n/plugin_yaml/publisher
	CreatedAt      string // Creation time
	UpdatedAt      string // Update time
	DeletedAt      string // Deletion time
}

// pluginMarketplaceDisplayI18nColumns holds the columns for the table plugin_marketplace_display_i18n.
var pluginMarketplaceDisplayI18nColumns = PluginMarketplaceDisplayI18nColumns{
	Id:             "id",
	ReleaseId:      "release_id",
	PluginId:       "plugin_id",
	ReleaseVersion: "release_version",
	Locale:         "locale",
	Name:           "name",
	Summary:        "summary",
	Source:         "source",
	CreatedAt:      "created_at",
	UpdatedAt:      "updated_at",
	DeletedAt:      "deleted_at",
}

// NewPluginMarketplaceDisplayI18nDao creates and returns a new DAO object for table data access.
func NewPluginMarketplaceDisplayI18nDao(handlers ...gdb.ModelHandler) *PluginMarketplaceDisplayI18nDao {
	return &PluginMarketplaceDisplayI18nDao{
		group:    "default",
		table:    "plugin_marketplace_display_i18n",
		columns:  pluginMarketplaceDisplayI18nColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *PluginMarketplaceDisplayI18nDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *PluginMarketplaceDisplayI18nDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *PluginMarketplaceDisplayI18nDao) Columns() PluginMarketplaceDisplayI18nColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *PluginMarketplaceDisplayI18nDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *PluginMarketplaceDisplayI18nDao) Ctx(ctx context.Context) *gdb.Model {
	model := dao.DB().Model(dao.table)
	for _, handler := range dao.handlers {
		model = handler(model)
	}
	return model.Safe().Ctx(ctx)
}

// Transaction wraps the transaction logic using function f.
func (dao *PluginMarketplaceDisplayI18nDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
