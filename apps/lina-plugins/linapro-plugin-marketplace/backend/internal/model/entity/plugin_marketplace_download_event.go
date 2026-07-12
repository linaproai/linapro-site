// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"time"
)

// PluginMarketplaceDownloadEvent is the golang structure for table plugin_marketplace_download_event.
type PluginMarketplaceDownloadEvent struct {
	Id              int        `json:"id"              orm:"id"                description:"Primary key ID"`
	SessionId       string     `json:"sessionId"       orm:"session_id"        description:"Opaque download session ID"`
	ReleaseId       int        `json:"releaseId"       orm:"release_id"        description:"Owning release ID"`
	ArtifactId      int        `json:"artifactId"      orm:"artifact_id"       description:"Owning artifact ID"`
	PluginId        string     `json:"pluginId"        orm:"plugin_id"         description:"Stable plugin ID"`
	ReleaseVersion  string     `json:"releaseVersion"  orm:"release_version"   description:"Plugin release version"`
	RequesterUserId int64      `json:"requesterUserId" orm:"requester_user_id" description:"Requester user ID"`
	EventType       string     `json:"eventType"       orm:"event_type"        description:"Download event type: created/started/completed/failed"`
	ClientIpHash    string     `json:"clientIpHash"    orm:"client_ip_hash"    description:"Hashed client IP for statistics"`
	UserAgentHash   string     `json:"userAgentHash"   orm:"user_agent_hash"   description:"Hashed user agent for statistics"`
	CreatedAt       *time.Time `json:"createdAt"       orm:"created_at"        description:"Creation time"`
}
