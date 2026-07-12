// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// PluginMarketplaceVisibilityGrantDao is the data access object for the table plugin_marketplace_visibility_grant.
type PluginMarketplaceVisibilityGrantDao struct {
	table    string                                  // table is the underlying table name of the DAO.
	group    string                                  // group is the database configuration group name of the current DAO.
	columns  PluginMarketplaceVisibilityGrantColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler                      // handlers for customized model modification.
}

// PluginMarketplaceVisibilityGrantColumns defines and stores column names for the table plugin_marketplace_visibility_grant.
type PluginMarketplaceVisibilityGrantColumns struct {
	Id             string // Primary key ID
	PluginRecordId string // Owning marketplace plugin record ID
	PluginId       string // Stable plugin ID
	ScopeType      string // Visibility scope type: public/tenant/org/user/reserved_license
	ScopeId        string // Scope identifier, empty for public scope
	Permission     string // Permission covered by the grant: view/download
	Status         string // Status: 0=disabled, 1=enabled
	ExpiresAt      string // Grant expiration time
	CreatedAt      string // Creation time
	UpdatedAt      string // Update time
	DeletedAt      string // Deletion time
}

// pluginMarketplaceVisibilityGrantColumns holds the columns for the table plugin_marketplace_visibility_grant.
var pluginMarketplaceVisibilityGrantColumns = PluginMarketplaceVisibilityGrantColumns{
	Id:             "id",
	PluginRecordId: "plugin_record_id",
	PluginId:       "plugin_id",
	ScopeType:      "scope_type",
	ScopeId:        "scope_id",
	Permission:     "permission",
	Status:         "status",
	ExpiresAt:      "expires_at",
	CreatedAt:      "created_at",
	UpdatedAt:      "updated_at",
	DeletedAt:      "deleted_at",
}

// NewPluginMarketplaceVisibilityGrantDao creates and returns a new DAO object for table data access.
func NewPluginMarketplaceVisibilityGrantDao(handlers ...gdb.ModelHandler) *PluginMarketplaceVisibilityGrantDao {
	return &PluginMarketplaceVisibilityGrantDao{
		group:    "default",
		table:    "plugin_marketplace_visibility_grant",
		columns:  pluginMarketplaceVisibilityGrantColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *PluginMarketplaceVisibilityGrantDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *PluginMarketplaceVisibilityGrantDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *PluginMarketplaceVisibilityGrantDao) Columns() PluginMarketplaceVisibilityGrantColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *PluginMarketplaceVisibilityGrantDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *PluginMarketplaceVisibilityGrantDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *PluginMarketplaceVisibilityGrantDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
