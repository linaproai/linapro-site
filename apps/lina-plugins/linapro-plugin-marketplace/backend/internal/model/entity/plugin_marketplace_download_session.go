// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"time"
)

// PluginMarketplaceDownloadSession is the golang structure for table plugin_marketplace_download_session.
type PluginMarketplaceDownloadSession struct {
	Id                    int        `json:"id"                    orm:"id"                     description:"Primary key ID"`
	SessionId             string     `json:"sessionId"             orm:"session_id"             description:"Opaque download session ID"`
	ReleaseId             int        `json:"releaseId"             orm:"release_id"             description:"Owning release ID"`
	ArtifactId            int        `json:"artifactId"            orm:"artifact_id"            description:"Owning artifact ID"`
	PluginId              string     `json:"pluginId"              orm:"plugin_id"              description:"Stable plugin ID"`
	ReleaseVersion        string     `json:"releaseVersion"        orm:"release_version"        description:"Plugin release version"`
	RequesterUserId       int64      `json:"requesterUserId"       orm:"requester_user_id"      description:"Requester user ID"`
	Status                string     `json:"status"                orm:"status"                 description:"Session status: active/expired/consumed/revoked"`
	ArtifactType          string     `json:"artifactType"          orm:"artifact_type"          description:"Artifact type bound to the session"`
	ArtifactSizeBytes     int64      `json:"artifactSizeBytes"     orm:"artifact_size_bytes"    description:"Artifact size in bytes"`
	Sha256                string     `json:"sha256"                orm:"sha256"                 description:"Artifact SHA-256 checksum returned to the client"`
	AuthorizationSnapshot string     `json:"authorizationSnapshot" orm:"authorization_snapshot" description:"Authorization decision snapshot captured at session creation"`
	ExpiresAt             *time.Time `json:"expiresAt"             orm:"expires_at"             description:"Session expiration time"`
	ConsumedAt            *time.Time `json:"consumedAt"            orm:"consumed_at"            description:"First successful download time"`
	CreatedAt             *time.Time `json:"createdAt"             orm:"created_at"             description:"Creation time"`
	UpdatedAt             *time.Time `json:"updatedAt"             orm:"updated_at"             description:"Update time"`
	DeletedAt             *time.Time `json:"deletedAt"             orm:"deleted_at"             description:"Deletion time"`
}
