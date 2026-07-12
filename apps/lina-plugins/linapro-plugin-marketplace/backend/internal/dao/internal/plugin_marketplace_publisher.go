// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// PluginMarketplacePublisherDao is the data access object for the table plugin_marketplace_publisher.
type PluginMarketplacePublisherDao struct {
	table    string                            // table is the underlying table name of the DAO.
	group    string                            // group is the database configuration group name of the current DAO.
	columns  PluginMarketplacePublisherColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler                // handlers for customized model modification.
}

// PluginMarketplacePublisherColumns defines and stores column names for the table plugin_marketplace_publisher.
type PluginMarketplacePublisherColumns struct {
	Id           string // Primary key ID
	PublisherKey string // Stable publisher key
	Name         string // Publisher display name
	Summary      string // Publisher summary
	OwnerUserId  string // Owning user ID
	OwnerOrgId   string // Owning organization ID, 0 means none
	Verified     string // Whether the publisher has been verified
	Status       string // Publisher status: active/suspended
	Homepage     string // Publisher homepage URL
	ContactEmail string // Publisher contact email
	Remark       string // Remark
	CreatedAt    string // Creation time
	UpdatedAt    string // Update time
	DeletedAt    string // Deletion time
}

// pluginMarketplacePublisherColumns holds the columns for the table plugin_marketplace_publisher.
var pluginMarketplacePublisherColumns = PluginMarketplacePublisherColumns{
	Id:           "id",
	PublisherKey: "publisher_key",
	Name:         "name",
	Summary:      "summary",
	OwnerUserId:  "owner_user_id",
	OwnerOrgId:   "owner_org_id",
	Verified:     "verified",
	Status:       "status",
	Homepage:     "homepage",
	ContactEmail: "contact_email",
	Remark:       "remark",
	CreatedAt:    "created_at",
	UpdatedAt:    "updated_at",
	DeletedAt:    "deleted_at",
}

// NewPluginMarketplacePublisherDao creates and returns a new DAO object for table data access.
func NewPluginMarketplacePublisherDao(handlers ...gdb.ModelHandler) *PluginMarketplacePublisherDao {
	return &PluginMarketplacePublisherDao{
		group:    "default",
		table:    "plugin_marketplace_publisher",
		columns:  pluginMarketplacePublisherColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *PluginMarketplacePublisherDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *PluginMarketplacePublisherDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *PluginMarketplacePublisherDao) Columns() PluginMarketplacePublisherColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *PluginMarketplacePublisherDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *PluginMarketplacePublisherDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *PluginMarketplacePublisherDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
