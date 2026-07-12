// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

// PluginMarketplaceDownloadEvent is the golang structure of table plugin_marketplace_download_event for DAO operations like Where/Data.
type PluginMarketplaceDownloadEvent struct {
	g.Meta          `orm:"table:plugin_marketplace_download_event, do:true"`
	Id              any        // Primary key ID
	SessionId       any        // Opaque download session ID
	ReleaseId       any        // Owning release ID
	ArtifactId      any        // Owning artifact ID
	PluginId        any        // Stable plugin ID
	ReleaseVersion  any        // Plugin release version
	RequesterUserId any        // Requester user ID
	EventType       any        // Download event type: created/started/completed/failed
	ClientIpHash    any        // Hashed client IP for statistics
	UserAgentHash   any        // Hashed user agent for statistics
	CreatedAt       *time.Time // Creation time
}
