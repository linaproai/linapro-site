// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// PluginMarketplaceDownloadEventDao is the data access object for the table plugin_marketplace_download_event.
type PluginMarketplaceDownloadEventDao struct {
	table    string                                // table is the underlying table name of the DAO.
	group    string                                // group is the database configuration group name of the current DAO.
	columns  PluginMarketplaceDownloadEventColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler                    // handlers for customized model modification.
}

// PluginMarketplaceDownloadEventColumns defines and stores column names for the table plugin_marketplace_download_event.
type PluginMarketplaceDownloadEventColumns struct {
	Id              string // Primary key ID
	SessionId       string // Opaque download session ID
	ReleaseId       string // Owning release ID
	ArtifactId      string // Owning artifact ID
	PluginId        string // Stable plugin ID
	ReleaseVersion  string // Plugin release version
	RequesterUserId string // Requester user ID
	EventType       string // Download event type: created/started/completed/failed
	ClientIpHash    string // Hashed client IP for statistics
	UserAgentHash   string // Hashed user agent for statistics
	CreatedAt       string // Creation time
}

// pluginMarketplaceDownloadEventColumns holds the columns for the table plugin_marketplace_download_event.
var pluginMarketplaceDownloadEventColumns = PluginMarketplaceDownloadEventColumns{
	Id:              "id",
	SessionId:       "session_id",
	ReleaseId:       "release_id",
	ArtifactId:      "artifact_id",
	PluginId:        "plugin_id",
	ReleaseVersion:  "release_version",
	RequesterUserId: "requester_user_id",
	EventType:       "event_type",
	ClientIpHash:    "client_ip_hash",
	UserAgentHash:   "user_agent_hash",
	CreatedAt:       "created_at",
}

// NewPluginMarketplaceDownloadEventDao creates and returns a new DAO object for table data access.
func NewPluginMarketplaceDownloadEventDao(handlers ...gdb.ModelHandler) *PluginMarketplaceDownloadEventDao {
	return &PluginMarketplaceDownloadEventDao{
		group:    "default",
		table:    "plugin_marketplace_download_event",
		columns:  pluginMarketplaceDownloadEventColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *PluginMarketplaceDownloadEventDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *PluginMarketplaceDownloadEventDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *PluginMarketplaceDownloadEventDao) Columns() PluginMarketplaceDownloadEventColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *PluginMarketplaceDownloadEventDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *PluginMarketplaceDownloadEventDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *PluginMarketplaceDownloadEventDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
