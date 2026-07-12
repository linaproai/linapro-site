// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// PluginMarketplaceTagDao is the data access object for the table plugin_marketplace_tag.
type PluginMarketplaceTagDao struct {
	table    string                      // table is the underlying table name of the DAO.
	group    string                      // group is the database configuration group name of the current DAO.
	columns  PluginMarketplaceTagColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler          // handlers for customized model modification.
}

// PluginMarketplaceTagColumns defines and stores column names for the table plugin_marketplace_tag.
type PluginMarketplaceTagColumns struct {
	Id        string // Primary key ID
	TagCode   string // Stable tag code
	Name      string // Tag display name
	TagType   string // Tag type: category/tag
	Sort      string // Display order
	Status    string // Status: 0=disabled, 1=enabled
	CreatedAt string // Creation time
	UpdatedAt string // Update time
	DeletedAt string // Deletion time
}

// pluginMarketplaceTagColumns holds the columns for the table plugin_marketplace_tag.
var pluginMarketplaceTagColumns = PluginMarketplaceTagColumns{
	Id:        "id",
	TagCode:   "tag_code",
	Name:      "name",
	TagType:   "tag_type",
	Sort:      "sort",
	Status:    "status",
	CreatedAt: "created_at",
	UpdatedAt: "updated_at",
	DeletedAt: "deleted_at",
}

// NewPluginMarketplaceTagDao creates and returns a new DAO object for table data access.
func NewPluginMarketplaceTagDao(handlers ...gdb.ModelHandler) *PluginMarketplaceTagDao {
	return &PluginMarketplaceTagDao{
		group:    "default",
		table:    "plugin_marketplace_tag",
		columns:  pluginMarketplaceTagColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *PluginMarketplaceTagDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *PluginMarketplaceTagDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *PluginMarketplaceTagDao) Columns() PluginMarketplaceTagColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *PluginMarketplaceTagDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *PluginMarketplaceTagDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *PluginMarketplaceTagDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
