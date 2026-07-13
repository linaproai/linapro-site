// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// PluginMarketplacePluginDao is the data access object for the table plugin_marketplace_plugin.
type PluginMarketplacePluginDao struct {
	table    string                         // table is the underlying table name of the DAO.
	group    string                         // group is the database configuration group name of the current DAO.
	columns  PluginMarketplacePluginColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler             // handlers for customized model modification.
}

// PluginMarketplacePluginColumns defines and stores column names for the table plugin_marketplace_plugin.
type PluginMarketplacePluginColumns struct {
	Id              string // Primary key ID
	PublisherId     string // Owning publisher ID
	PluginId        string // Stable plugin ID
	Name            string // Plugin display name
	Summary         string // Short marketplace summary
	Description     string // Long marketplace description
	PluginType      string // Plugin type: source/dynamic
	MarketStatus    string // Marketplace status: draft/published/delisted/deprecated
	Visibility      string // Visibility policy: public/private/reserved
	LatestReleaseId string // Latest published release ID
	LatestVersion   string // Latest published version
	Icon            string // Marketplace icon path or URL
	Homepage        string // Plugin homepage URL
	Repository      string // Plugin source repository URL
	License         string // Plugin license identifier
	DownloadCount   string // Aggregated download count snapshot
	SourceKind      string // Publish source kind: git/upload
	RepoUrl         string // Git repository URL when source_kind is git
	RepoProvider    string // Git provider: github/gitee, empty for upload
	CredentialRef   string // Opaque credential reference for private Git access, empty when public
	LastSyncAt      string // Last Git metadata discovery time
	LastSyncStatus  string // Last Git sync status
	LastSyncMessage string // Last Git sync diagnostic message without secrets
	PublishedAt     string // First published time
	CreatedAt       string // Creation time
	UpdatedAt       string // Update time
	DeletedAt       string // Deletion time
}

// pluginMarketplacePluginColumns holds the columns for the table plugin_marketplace_plugin.
var pluginMarketplacePluginColumns = PluginMarketplacePluginColumns{
	Id:              "id",
	PublisherId:     "publisher_id",
	PluginId:        "plugin_id",
	Name:            "name",
	Summary:         "summary",
	Description:     "description",
	PluginType:      "plugin_type",
	MarketStatus:    "market_status",
	Visibility:      "visibility",
	LatestReleaseId: "latest_release_id",
	LatestVersion:   "latest_version",
	Icon:            "icon",
	Homepage:        "homepage",
	Repository:      "repository",
	License:         "license",
	DownloadCount:   "download_count",
	SourceKind:      "source_kind",
	RepoUrl:         "repo_url",
	RepoProvider:    "repo_provider",
	CredentialRef:   "credential_ref",
	LastSyncAt:      "last_sync_at",
	LastSyncStatus:  "last_sync_status",
	LastSyncMessage: "last_sync_message",
	PublishedAt:     "published_at",
	CreatedAt:       "created_at",
	UpdatedAt:       "updated_at",
	DeletedAt:       "deleted_at",
}

// NewPluginMarketplacePluginDao creates and returns a new DAO object for table data access.
func NewPluginMarketplacePluginDao(handlers ...gdb.ModelHandler) *PluginMarketplacePluginDao {
	return &PluginMarketplacePluginDao{
		group:    "default",
		table:    "plugin_marketplace_plugin",
		columns:  pluginMarketplacePluginColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *PluginMarketplacePluginDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *PluginMarketplacePluginDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *PluginMarketplacePluginDao) Columns() PluginMarketplacePluginColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *PluginMarketplacePluginDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *PluginMarketplacePluginDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *PluginMarketplacePluginDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
