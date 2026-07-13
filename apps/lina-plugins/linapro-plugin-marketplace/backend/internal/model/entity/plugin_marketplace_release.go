// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"time"
)

// PluginMarketplaceRelease is the golang structure for table plugin_marketplace_release.
type PluginMarketplaceRelease struct {
	Id                 int        `json:"id"                 orm:"id"                   description:"Primary key ID"`
	PluginRecordId     int        `json:"pluginRecordId"     orm:"plugin_record_id"     description:"Owning marketplace plugin record ID"`
	PublisherId        int        `json:"publisherId"        orm:"publisher_id"         description:"Owning publisher ID"`
	PluginId           string     `json:"pluginId"           orm:"plugin_id"            description:"Stable plugin ID"`
	ReleaseVersion     string     `json:"releaseVersion"     orm:"release_version"      description:"Plugin release version"`
	SourceRef          string     `json:"sourceRef"        orm:"source_ref"         description:"Git tag or ref for git-sourced releases, empty for upload packages"`
	PluginType         string     `json:"pluginType"         orm:"plugin_type"          description:"Plugin type: source/dynamic"`
	ReleaseStatus      string     `json:"releaseStatus"      orm:"release_status"       description:"Release status: draft/published/delisted/deprecated"`
	ReviewStatus       string     `json:"reviewStatus"       orm:"review_status"        description:"Review status: draft/submitted/reviewing/approved/rejected"`
	Visibility         string     `json:"visibility"         orm:"visibility"           description:"Release visibility policy"`
	MinHostVersion     string     `json:"minHostVersion"     orm:"min_host_version"     description:"Minimum compatible LinaPro host version"`
	MaxHostVersion     string     `json:"maxHostVersion"     orm:"max_host_version"     description:"Maximum compatible LinaPro host version"`
	ManifestSnapshot   string     `json:"manifestSnapshot"   orm:"manifest_snapshot"    description:"Parsed plugin.yaml snapshot"`
	DependencySummary  string     `json:"dependencySummary"  orm:"dependency_summary"   description:"Dependency scan summary"`
	HostServiceSummary string     `json:"hostServiceSummary" orm:"host_service_summary" description:"Host service scan summary"`
	RouteSummary       string     `json:"routeSummary"       orm:"route_summary"        description:"Route scan summary"`
	SqlSummary         string     `json:"sqlSummary"         orm:"sql_summary"          description:"SQL resource scan summary"`
	I18NSummary        string     `json:"i18NSummary"        orm:"i18n_summary"         description:"i18n resource scan summary"`
	DocsSummary        string     `json:"docsSummary"        orm:"docs_summary"         description:"Marketplace document scan summary"`
	RiskSummary        string     `json:"riskSummary"        orm:"risk_summary"         description:"Aggregated review risk summary"`
	ReviewMessage      string     `json:"reviewMessage"      orm:"review_message"       description:"Latest review message"`
	SubmittedAt        *time.Time `json:"submittedAt"        orm:"submitted_at"         description:"Review submission time"`
	ReviewedAt         *time.Time `json:"reviewedAt"         orm:"reviewed_at"          description:"Review completion time"`
	PublishedAt        *time.Time `json:"publishedAt"        orm:"published_at"         description:"Publish time"`
	CreatedAt          *time.Time `json:"createdAt"          orm:"created_at"           description:"Creation time"`
	UpdatedAt          *time.Time `json:"updatedAt"          orm:"updated_at"           description:"Update time"`
	DeletedAt          *time.Time `json:"deletedAt"          orm:"deleted_at"           description:"Deletion time"`
}
