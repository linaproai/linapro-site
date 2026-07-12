// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// PluginMarketplaceDownloadSessionDao is the data access object for the table plugin_marketplace_download_session.
type PluginMarketplaceDownloadSessionDao struct {
	table    string                                  // table is the underlying table name of the DAO.
	group    string                                  // group is the database configuration group name of the current DAO.
	columns  PluginMarketplaceDownloadSessionColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler                      // handlers for customized model modification.
}

// PluginMarketplaceDownloadSessionColumns defines and stores column names for the table plugin_marketplace_download_session.
type PluginMarketplaceDownloadSessionColumns struct {
	Id                    string // Primary key ID
	SessionId             string // Opaque download session ID
	ReleaseId             string // Owning release ID
	ArtifactId            string // Owning artifact ID
	PluginId              string // Stable plugin ID
	ReleaseVersion        string // Plugin release version
	RequesterUserId       string // Requester user ID
	Status                string // Session status: active/expired/consumed/revoked
	ArtifactType          string // Artifact type bound to the session
	ArtifactSizeBytes     string // Artifact size in bytes
	Sha256                string // Artifact SHA-256 checksum returned to the client
	AuthorizationSnapshot string // Authorization decision snapshot captured at session creation
	ExpiresAt             string // Session expiration time
	ConsumedAt            string // First successful download time
	CreatedAt             string // Creation time
	UpdatedAt             string // Update time
	DeletedAt             string // Deletion time
}

// pluginMarketplaceDownloadSessionColumns holds the columns for the table plugin_marketplace_download_session.
var pluginMarketplaceDownloadSessionColumns = PluginMarketplaceDownloadSessionColumns{
	Id:                    "id",
	SessionId:             "session_id",
	ReleaseId:             "release_id",
	ArtifactId:            "artifact_id",
	PluginId:              "plugin_id",
	ReleaseVersion:        "release_version",
	RequesterUserId:       "requester_user_id",
	Status:                "status",
	ArtifactType:          "artifact_type",
	ArtifactSizeBytes:     "artifact_size_bytes",
	Sha256:                "sha256",
	AuthorizationSnapshot: "authorization_snapshot",
	ExpiresAt:             "expires_at",
	ConsumedAt:            "consumed_at",
	CreatedAt:             "created_at",
	UpdatedAt:             "updated_at",
	DeletedAt:             "deleted_at",
}

// NewPluginMarketplaceDownloadSessionDao creates and returns a new DAO object for table data access.
func NewPluginMarketplaceDownloadSessionDao(handlers ...gdb.ModelHandler) *PluginMarketplaceDownloadSessionDao {
	return &PluginMarketplaceDownloadSessionDao{
		group:    "default",
		table:    "plugin_marketplace_download_session",
		columns:  pluginMarketplaceDownloadSessionColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *PluginMarketplaceDownloadSessionDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *PluginMarketplaceDownloadSessionDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *PluginMarketplaceDownloadSessionDao) Columns() PluginMarketplaceDownloadSessionColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *PluginMarketplaceDownloadSessionDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *PluginMarketplaceDownloadSessionDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *PluginMarketplaceDownloadSessionDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
