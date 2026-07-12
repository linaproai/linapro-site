// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"time"
)

// PluginMarketplaceDoc is the golang structure for table plugin_marketplace_doc.
type PluginMarketplaceDoc struct {
	Id             int        `json:"id"             orm:"id"              description:"Primary key ID"`
	ReleaseId      int        `json:"releaseId"      orm:"release_id"      description:"Owning release ID"`
	PluginId       string     `json:"pluginId"       orm:"plugin_id"       description:"Stable plugin ID"`
	ReleaseVersion string     `json:"releaseVersion" orm:"release_version" description:"Plugin release version"`
	Locale         string     `json:"locale"         orm:"locale"          description:"Document locale"`
	DocPath        string     `json:"docPath"        orm:"doc_path"        description:"Document path inside manifest/docs or README fallback"`
	SourceKind     string     `json:"sourceKind"     orm:"source_kind"     description:"Document source kind: manifest_docs/readme"`
	Title          string     `json:"title"          orm:"title"           description:"Document title"`
	Summary        string     `json:"summary"        orm:"summary"         description:"Document search summary"`
	ContentHash    string     `json:"contentHash"    orm:"content_hash"    description:"Document content hash"`
	SearchText     string     `json:"searchText"     orm:"search_text"     description:"Plain text used for search indexing"`
	CreatedAt      *time.Time `json:"createdAt"      orm:"created_at"      description:"Creation time"`
	UpdatedAt      *time.Time `json:"updatedAt"      orm:"updated_at"      description:"Update time"`
	DeletedAt      *time.Time `json:"deletedAt"      orm:"deleted_at"      description:"Deletion time"`
}
