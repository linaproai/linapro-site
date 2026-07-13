// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// PluginMarketplaceCredentialDao is the data access object for the table plugin_marketplace_credential.
type PluginMarketplaceCredentialDao struct {
	table    string                             // table is the underlying table name of the DAO.
	group    string                             // group is the database configuration group name of the current DAO.
	columns  PluginMarketplaceCredentialColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler                 // handlers for customized model modification.
}

// PluginMarketplaceCredentialColumns defines and stores column names for the table plugin_marketplace_credential.
type PluginMarketplaceCredentialColumns struct {
	Id            string // Primary key ID
	CredentialRef string // Opaque credential reference stored on plugin records
	OwnerUserId   string // Owning user ID of the credential
	Provider      string // Git provider associated with the credential
	CipherText    string // Encrypted token ciphertext; never returned by marketplace APIs
	CreatedAt     string // Creation time
	UpdatedAt     string // Update time
	DeletedAt     string // Deletion time
}

// pluginMarketplaceCredentialColumns holds the columns for the table plugin_marketplace_credential.
var pluginMarketplaceCredentialColumns = PluginMarketplaceCredentialColumns{
	Id:            "id",
	CredentialRef: "credential_ref",
	OwnerUserId:   "owner_user_id",
	Provider:      "provider",
	CipherText:    "cipher_text",
	CreatedAt:     "created_at",
	UpdatedAt:     "updated_at",
	DeletedAt:     "deleted_at",
}

// NewPluginMarketplaceCredentialDao creates and returns a new DAO object for table data access.
func NewPluginMarketplaceCredentialDao(handlers ...gdb.ModelHandler) *PluginMarketplaceCredentialDao {
	return &PluginMarketplaceCredentialDao{
		group:    "default",
		table:    "plugin_marketplace_credential",
		columns:  pluginMarketplaceCredentialColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *PluginMarketplaceCredentialDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *PluginMarketplaceCredentialDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *PluginMarketplaceCredentialDao) Columns() PluginMarketplaceCredentialColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *PluginMarketplaceCredentialDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *PluginMarketplaceCredentialDao) Ctx(ctx context.Context) *gdb.Model {
	model := dao.DB().Model(dao.table)
	for _, handler := range dao.handlers {
		model = handler(model)
	}
	return model.Safe().Ctx(ctx)
}

// Transaction wraps the transaction logic using function f.
func (dao *PluginMarketplaceCredentialDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) error {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
