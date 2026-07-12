// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// PluginMarketplaceArtifactDao is the data access object for the table plugin_marketplace_artifact.
type PluginMarketplaceArtifactDao struct {
	table    string                           // table is the underlying table name of the DAO.
	group    string                           // group is the database configuration group name of the current DAO.
	columns  PluginMarketplaceArtifactColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler               // handlers for customized model modification.
}

// PluginMarketplaceArtifactColumns defines and stores column names for the table plugin_marketplace_artifact.
type PluginMarketplaceArtifactColumns struct {
	Id             string // Primary key ID
	ReleaseId      string // Owning release ID
	PluginId       string // Stable plugin ID
	ReleaseVersion string // Plugin release version
	ArtifactType   string // Artifact type: source_zip/dynamic_zip/plugin_wasm
	StorageKey     string // Storage object key or managed file key
	FileName       string // Original artifact file name
	ContentType    string // Artifact content type
	SizeBytes      string // Artifact size in bytes
	Sha256         string // Artifact SHA-256 checksum
	ManifestSha256 string // Root manifest SHA-256 checksum
	WasmSha256     string // Extracted plugin.wasm SHA-256 checksum
	CreatedAt      string // Creation time
	UpdatedAt      string // Update time
	DeletedAt      string // Deletion time
}

// pluginMarketplaceArtifactColumns holds the columns for the table plugin_marketplace_artifact.
var pluginMarketplaceArtifactColumns = PluginMarketplaceArtifactColumns{
	Id:             "id",
	ReleaseId:      "release_id",
	PluginId:       "plugin_id",
	ReleaseVersion: "release_version",
	ArtifactType:   "artifact_type",
	StorageKey:     "storage_key",
	FileName:       "file_name",
	ContentType:    "content_type",
	SizeBytes:      "size_bytes",
	Sha256:         "sha256",
	ManifestSha256: "manifest_sha256",
	WasmSha256:     "wasm_sha256",
	CreatedAt:      "created_at",
	UpdatedAt:      "updated_at",
	DeletedAt:      "deleted_at",
}

// NewPluginMarketplaceArtifactDao creates and returns a new DAO object for table data access.
func NewPluginMarketplaceArtifactDao(handlers ...gdb.ModelHandler) *PluginMarketplaceArtifactDao {
	return &PluginMarketplaceArtifactDao{
		group:    "default",
		table:    "plugin_marketplace_artifact",
		columns:  pluginMarketplaceArtifactColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *PluginMarketplaceArtifactDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *PluginMarketplaceArtifactDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *PluginMarketplaceArtifactDao) Columns() PluginMarketplaceArtifactColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *PluginMarketplaceArtifactDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *PluginMarketplaceArtifactDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *PluginMarketplaceArtifactDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
