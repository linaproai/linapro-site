// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

// PluginMarketplaceDoc is the golang structure of table plugin_marketplace_doc for DAO operations like Where/Data.
type PluginMarketplaceDoc struct {
	g.Meta         `orm:"table:plugin_marketplace_doc, do:true"`
	Id             any        // Primary key ID
	ReleaseId      any        // Owning release ID
	PluginId       any        // Stable plugin ID
	ReleaseVersion any        // Plugin release version
	Locale         any        // Document locale
	DocPath        any        // Document path inside manifest/docs or README fallback
	SourceKind     any        // Document source kind: manifest_docs/readme
	Title          any        // Document title
	Summary        any        // Document search summary
	ContentHash    any        // Document content hash
	SearchText     any        // Plain text used for search indexing
	CreatedAt      *time.Time // Creation time
	UpdatedAt      *time.Time // Update time
	DeletedAt      *time.Time // Deletion time
}
