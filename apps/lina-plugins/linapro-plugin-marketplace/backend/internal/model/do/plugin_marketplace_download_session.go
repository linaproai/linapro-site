// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

// PluginMarketplaceDownloadSession is the golang structure of table plugin_marketplace_download_session for DAO operations like Where/Data.
type PluginMarketplaceDownloadSession struct {
	g.Meta                `orm:"table:plugin_marketplace_download_session, do:true"`
	Id                    any        // Primary key ID
	SessionId             any        // Opaque download session ID
	ReleaseId             any        // Owning release ID
	ArtifactId            any        // Owning artifact ID
	PluginId              any        // Stable plugin ID
	ReleaseVersion        any        // Plugin release version
	RequesterUserId       any        // Requester user ID
	Status                any        // Session status: active/expired/consumed/revoked
	ArtifactType          any        // Artifact type bound to the session
	ArtifactSizeBytes     any        // Artifact size in bytes
	Sha256                any        // Artifact SHA-256 checksum returned to the client
	AuthorizationSnapshot any        // Authorization decision snapshot captured at session creation
	ExpiresAt             *time.Time // Session expiration time
	ConsumedAt            *time.Time // First successful download time
	CreatedAt             *time.Time // Creation time
	UpdatedAt             *time.Time // Update time
	DeletedAt             *time.Time // Deletion time
}
