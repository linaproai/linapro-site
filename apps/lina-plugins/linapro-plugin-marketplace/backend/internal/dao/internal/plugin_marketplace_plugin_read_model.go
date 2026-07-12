// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// PluginMarketplacePluginReadModelDao is the data access object for the table plugin_marketplace_plugin_read_model.
type PluginMarketplacePluginReadModelDao struct {
	table    string                                  // table is the underlying table name of the DAO.
	group    string                                  // group is the database configuration group name of the current DAO.
	columns  PluginMarketplacePluginReadModelColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler                      // handlers for customized model modification.
}

// PluginMarketplacePluginReadModelColumns defines and stores column names for the table plugin_marketplace_plugin_read_model.
type PluginMarketplacePluginReadModelColumns struct {
	Id                string // Primary key ID
	PluginRecordId    string // Owning marketplace plugin record ID
	PublisherId       string // Owning publisher ID
	PublisherName     string // Publisher display name snapshot
	PublisherVerified string // Publisher verification snapshot
	PluginId          string // Stable plugin ID
	Name              string // Plugin display name
	Summary           string // Short marketplace summary
	PluginType        string // Plugin type: source/dynamic
	MarketStatus      string // Marketplace status
	Visibility        string // Visibility policy
	LatestReleaseId   string // Latest published release ID
	LatestVersion     string // Latest published version
	MinHostVersion    string // Minimum compatible LinaPro host version
	MaxHostVersion    string // Maximum compatible LinaPro host version
	PrimaryTag        string // Primary category tag code
	TagCodes          string // Tag code snapshot for display
	RiskCounts        string // Risk count snapshot grouped by severity
	DownloadCount     string // Aggregated download count snapshot
	PublishedAt       string // Latest publish time
	SearchText        string // Plain text projection used for catalog search
	CreatedAt         string // Creation time
	UpdatedAt         string // Update time
	DeletedAt         string // Deletion time
}

// pluginMarketplacePluginReadModelColumns holds the columns for the table plugin_marketplace_plugin_read_model.
var pluginMarketplacePluginReadModelColumns = PluginMarketplacePluginReadModelColumns{
	Id:                "id",
	PluginRecordId:    "plugin_record_id",
	PublisherId:       "publisher_id",
	PublisherName:     "publisher_name",
	PublisherVerified: "publisher_verified",
	PluginId:          "plugin_id",
	Name:              "name",
	Summary:           "summary",
	PluginType:        "plugin_type",
	MarketStatus:      "market_status",
	Visibility:        "visibility",
	LatestReleaseId:   "latest_release_id",
	LatestVersion:     "latest_version",
	MinHostVersion:    "min_host_version",
	MaxHostVersion:    "max_host_version",
	PrimaryTag:        "primary_tag",
	TagCodes:          "tag_codes",
	RiskCounts:        "risk_counts",
	DownloadCount:     "download_count",
	PublishedAt:       "published_at",
	SearchText:        "search_text",
	CreatedAt:         "created_at",
	UpdatedAt:         "updated_at",
	DeletedAt:         "deleted_at",
}

// NewPluginMarketplacePluginReadModelDao creates and returns a new DAO object for table data access.
func NewPluginMarketplacePluginReadModelDao(handlers ...gdb.ModelHandler) *PluginMarketplacePluginReadModelDao {
	return &PluginMarketplacePluginReadModelDao{
		group:    "default",
		table:    "plugin_marketplace_plugin_read_model",
		columns:  pluginMarketplacePluginReadModelColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *PluginMarketplacePluginReadModelDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *PluginMarketplacePluginReadModelDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *PluginMarketplacePluginReadModelDao) Columns() PluginMarketplacePluginReadModelColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *PluginMarketplacePluginReadModelDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *PluginMarketplacePluginReadModelDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *PluginMarketplacePluginReadModelDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
