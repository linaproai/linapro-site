// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// PluginMarketplaceReleaseDao is the data access object for the table plugin_marketplace_release.
type PluginMarketplaceReleaseDao struct {
	table    string                          // table is the underlying table name of the DAO.
	group    string                          // group is the database configuration group name of the current DAO.
	columns  PluginMarketplaceReleaseColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler              // handlers for customized model modification.
}

// PluginMarketplaceReleaseColumns defines and stores column names for the table plugin_marketplace_release.
type PluginMarketplaceReleaseColumns struct {
	Id                 string // Primary key ID
	PluginRecordId     string // Owning marketplace plugin record ID
	PublisherId        string // Owning publisher ID
	PluginId           string // Stable plugin ID
	ReleaseVersion     string // Plugin release version
	PluginType         string // Plugin type: source/dynamic
	ReleaseStatus      string // Release status: draft/published/delisted/deprecated
	ReviewStatus       string // Review status: draft/submitted/reviewing/approved/rejected
	Visibility         string // Release visibility policy
	MinHostVersion     string // Minimum compatible LinaPro host version
	MaxHostVersion     string // Maximum compatible LinaPro host version
	ManifestSnapshot   string // Parsed plugin.yaml snapshot
	DependencySummary  string // Dependency scan summary
	HostServiceSummary string // Host service scan summary
	RouteSummary       string // Route scan summary
	SqlSummary         string // SQL resource scan summary
	I18NSummary        string // i18n resource scan summary
	DocsSummary        string // Marketplace document scan summary
	RiskSummary        string // Aggregated review risk summary
	ReviewMessage      string // Latest review message
	SubmittedAt        string // Review submission time
	ReviewedAt         string // Review completion time
	PublishedAt        string // Publish time
	CreatedAt          string // Creation time
	UpdatedAt          string // Update time
	DeletedAt          string // Deletion time
}

// pluginMarketplaceReleaseColumns holds the columns for the table plugin_marketplace_release.
var pluginMarketplaceReleaseColumns = PluginMarketplaceReleaseColumns{
	Id:                 "id",
	PluginRecordId:     "plugin_record_id",
	PublisherId:        "publisher_id",
	PluginId:           "plugin_id",
	ReleaseVersion:     "release_version",
	PluginType:         "plugin_type",
	ReleaseStatus:      "release_status",
	ReviewStatus:       "review_status",
	Visibility:         "visibility",
	MinHostVersion:     "min_host_version",
	MaxHostVersion:     "max_host_version",
	ManifestSnapshot:   "manifest_snapshot",
	DependencySummary:  "dependency_summary",
	HostServiceSummary: "host_service_summary",
	RouteSummary:       "route_summary",
	SqlSummary:         "sql_summary",
	I18NSummary:        "i18n_summary",
	DocsSummary:        "docs_summary",
	RiskSummary:        "risk_summary",
	ReviewMessage:      "review_message",
	SubmittedAt:        "submitted_at",
	ReviewedAt:         "reviewed_at",
	PublishedAt:        "published_at",
	CreatedAt:          "created_at",
	UpdatedAt:          "updated_at",
	DeletedAt:          "deleted_at",
}

// NewPluginMarketplaceReleaseDao creates and returns a new DAO object for table data access.
func NewPluginMarketplaceReleaseDao(handlers ...gdb.ModelHandler) *PluginMarketplaceReleaseDao {
	return &PluginMarketplaceReleaseDao{
		group:    "default",
		table:    "plugin_marketplace_release",
		columns:  pluginMarketplaceReleaseColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *PluginMarketplaceReleaseDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *PluginMarketplaceReleaseDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *PluginMarketplaceReleaseDao) Columns() PluginMarketplaceReleaseColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *PluginMarketplaceReleaseDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *PluginMarketplaceReleaseDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *PluginMarketplaceReleaseDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
