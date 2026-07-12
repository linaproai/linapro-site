// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// PluginMarketplaceDocDao is the data access object for the table plugin_marketplace_doc.
type PluginMarketplaceDocDao struct {
	table    string                      // table is the underlying table name of the DAO.
	group    string                      // group is the database configuration group name of the current DAO.
	columns  PluginMarketplaceDocColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler          // handlers for customized model modification.
}

// PluginMarketplaceDocColumns defines and stores column names for the table plugin_marketplace_doc.
type PluginMarketplaceDocColumns struct {
	Id             string // Primary key ID
	ReleaseId      string // Owning release ID
	PluginId       string // Stable plugin ID
	ReleaseVersion string // Plugin release version
	Locale         string // Document locale
	DocPath        string // Document path inside manifest/docs or README fallback
	SourceKind     string // Document source kind: manifest_docs/readme
	Title          string // Document title
	Summary        string // Document search summary
	ContentHash    string // Document content hash
	SearchText     string // Plain text used for search indexing
	CreatedAt      string // Creation time
	UpdatedAt      string // Update time
	DeletedAt      string // Deletion time
}

// pluginMarketplaceDocColumns holds the columns for the table plugin_marketplace_doc.
var pluginMarketplaceDocColumns = PluginMarketplaceDocColumns{
	Id:             "id",
	ReleaseId:      "release_id",
	PluginId:       "plugin_id",
	ReleaseVersion: "release_version",
	Locale:         "locale",
	DocPath:        "doc_path",
	SourceKind:     "source_kind",
	Title:          "title",
	Summary:        "summary",
	ContentHash:    "content_hash",
	SearchText:     "search_text",
	CreatedAt:      "created_at",
	UpdatedAt:      "updated_at",
	DeletedAt:      "deleted_at",
}

// NewPluginMarketplaceDocDao creates and returns a new DAO object for table data access.
func NewPluginMarketplaceDocDao(handlers ...gdb.ModelHandler) *PluginMarketplaceDocDao {
	return &PluginMarketplaceDocDao{
		group:    "default",
		table:    "plugin_marketplace_doc",
		columns:  pluginMarketplaceDocColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *PluginMarketplaceDocDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *PluginMarketplaceDocDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *PluginMarketplaceDocDao) Columns() PluginMarketplaceDocColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *PluginMarketplaceDocDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *PluginMarketplaceDocDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *PluginMarketplaceDocDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
