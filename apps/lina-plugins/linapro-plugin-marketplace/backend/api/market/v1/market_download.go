// This file defines marketplace download-session API DTOs. Download sessions
// are short-lived authorization resources that expose checksum metadata without
// letting ordinary catalog or document reads create business write events.

package v1

import "github.com/gogf/gf/v2/frame/g"

// DownloadSessionCreateReq is the request for creating one marketplace download session.
type DownloadSessionCreateReq struct {
	g.Meta       `path:"/market/plugins/{pluginId}/releases/{version}/downloads" method:"post" tags:"Plugin Marketplace" summary:"Create marketplace download session" permission:"market:plugin:download" dc:"Create a short-lived download session for one visible marketplace release. The service verifies visibility and download permission before binding the requester, artifact, expiration, and SHA-256 checksum to the session."`
	PluginId     string                  `json:"pluginId" v:"required|length:1,64" dc:"Stable plugin ID whose release artifact is requested for download" eg:"linapro-demo-source"`
	Version      string                  `json:"version" v:"required|length:1,32" dc:"Release version requested for download" eg:"v0.1.0"`
	ArtifactType MarketplaceArtifactType `json:"artifactType" dc:"Optional artifact type to download: source_zip, dynamic_zip, or plugin_wasm; defaults to the primary artifact for the release" eg:"source_zip"`
}

// DownloadSessionCreateRes is the response for creating one marketplace download session.
type DownloadSessionCreateRes struct {
	Session *MarketplaceDownloadSessionItem `json:"session" dc:"Created short-lived marketplace download session and checksum metadata" eg:"{}"`
}

// DownloadSessionGetReq is the request for reading one marketplace download session.
type DownloadSessionGetReq struct {
	g.Meta    `path:"/market/download-sessions/{sessionId}" method:"get" tags:"Plugin Marketplace" summary:"Get marketplace download session" permission:"market:plugin:download" dc:"Return a marketplace download session that is still visible to the requester. Expired, revoked, or unauthorized sessions are rejected without returning artifact content."`
	SessionId string `json:"sessionId" v:"required|length:1,64" dc:"Opaque download session ID returned by the session creation endpoint" eg:"mpdl_0123456789abcdef"`
}

// DownloadSessionGetRes is the response for reading one marketplace download session.
type DownloadSessionGetRes struct {
	Session *MarketplaceDownloadSessionItem `json:"session" dc:"Marketplace download session metadata and current status" eg:"{}"`
}

// DownloadSessionContentReq is the request for streaming one download-session artifact body.
type DownloadSessionContentReq struct {
	g.Meta    `path:"/market/download-sessions/{sessionId}/content" method:"get" tags:"Plugin Marketplace" summary:"Download marketplace session content" permission:"market:plugin:download" dc:"Stream the artifact bytes bound to one active requester-owned marketplace download session. Expired, revoked, unauthorized, or missing sessions are rejected without returning package content."`
	SessionId string `json:"sessionId" v:"required|length:1,64" dc:"Opaque download session ID returned by the session creation endpoint" eg:"mpdl_0123456789abcdef"`
}

// DownloadSessionContentRes is intentionally empty because the handler streams binary content.
type DownloadSessionContentRes struct{}

// MarketplaceDownloadSessionItem is the short-lived download-session projection.
type MarketplaceDownloadSessionItem struct {
	SessionId    string                           `json:"sessionId" dc:"Opaque download session ID used for controlled artifact retrieval" eg:"mpdl_0123456789abcdef"`
	PluginId     string                           `json:"pluginId" dc:"Stable plugin ID bound to the download session" eg:"linapro-demo-source"`
	Version      string                           `json:"version" dc:"Release version bound to the download session" eg:"v0.1.0"`
	ArtifactType MarketplaceArtifactType          `json:"artifactType" dc:"Artifact type bound to the session: source_zip, dynamic_zip, or plugin_wasm" eg:"source_zip"`
	SizeBytes    int64                            `json:"sizeBytes" dc:"Artifact size in bytes" eg:"102400"`
	Sha256       string                           `json:"sha256" dc:"Artifact SHA-256 checksum returned to the client for download verification" eg:"0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"`
	Status       MarketplaceDownloadSessionStatus `json:"status" dc:"Download session status: active, expired, consumed, or revoked" eg:"active"`
	DownloadUrl  string                           `json:"downloadUrl" dc:"Short-lived controlled download URL or empty value when the client must stream through the session endpoint" eg:"/x/linapro-plugin-marketplace/market/download-sessions/mpdl_0123456789abcdef/content"`
	ExpiresAt    *int64                           `json:"expiresAt" dc:"Download session expiration time as Unix timestamp in milliseconds" eg:"1767243600000"`
	ConsumedAt   *int64                           `json:"consumedAt,omitempty" dc:"First successful download time as Unix timestamp in milliseconds, empty when the session has not been consumed" eg:"1767241800000"`
	CreatedAt    *int64                           `json:"createdAt" dc:"Download session creation time as Unix timestamp in milliseconds" eg:"1767240000000"`
}
