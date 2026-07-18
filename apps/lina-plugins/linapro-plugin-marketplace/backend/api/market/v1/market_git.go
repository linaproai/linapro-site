// This file defines Git source registration, metadata sync, and distribution
// query API DTOs for simplified marketplace publish and CLI install flows.

package v1

import "github.com/gogf/gf/v2/frame/g"

// GitSourceRegisterReq is the request for registering one Git-backed marketplace plugin.
type GitSourceRegisterReq struct {
	g.Meta       `path:"/market/plugins/git-sources" method:"post" tags:"Plugin Marketplace" summary:"Register marketplace Git source" permission:"market:plugin:publish" dc:"Register a GitHub or Gitee repository as a marketplace plugin source owned by the current publisher. The service stores repository coordinates and optional encrypted credentials, auto-detects single-plugin or multi-plugin repository layouts with minimal remote probing, creates private owner-visible plugin identities in pending_verify, and leaves full version discovery, verification, and review submission to the async marketplace process pipeline."`
	PublisherKey string `json:"publisherKey" v:"required|length:1,64" dc:"Stable publisher key that owns the marketplace plugin identity" eg:"linapro"`
	RepoUrl      string `json:"repoUrl" v:"required|length:1,512" dc:"GitHub or Gitee HTTPS repository URL" eg:"https://github.com/linaproai/linapro-demo-source"`
	AccessToken  string `json:"accessToken" v:"length:0,512" dc:"Optional private repository access token stored encrypted by the platform; never returned by later APIs" eg:""`
	Homepage     string `json:"homepage" v:"length:0,512" dc:"Optional marketplace homepage override" eg:"https://linapro.ai/plugins/source-demo"`
	License      string `json:"license" v:"length:0,64" dc:"Optional license identifier when not discovered from plugin.yaml" eg:"Apache-2.0"`
}

// GitSourceRegisterRes is the response for registering one or more Git-backed marketplace plugins from one repository.
type GitSourceRegisterRes struct {
	Plugin  *MarketplacePluginDetailItem   `json:"plugin" dc:"Primary registered marketplace plugin identity after enqueue; equals plugins[0] when plugins is non-empty" eg:"{}"`
	Plugins []*MarketplacePluginDetailItem `json:"plugins" dc:"All marketplace plugin identities discovered from the repository, including multi-plugin monorepos, initially pending_verify" eg:"[]"`
}

// GitSourceSyncReq is the request for manually triggering Git metadata discovery.
type GitSourceSyncReq struct {
	g.Meta   `path:"/market/plugins/{pluginId}/git-sync" method:"post" tags:"Plugin Marketplace" summary:"Sync marketplace Git source metadata" permission:"market:plugin:publish" dc:"Trigger metadata discovery for one Git-backed marketplace plugin owned by the current publisher. The service lists remote tags or falls back to the main branch and reads remote plugin.yaml without cloning full source trees."`
	PluginId string `json:"pluginId" v:"required|length:1,64" dc:"Stable plugin ID whose Git source metadata is refreshed" eg:"linapro-demo-source"`
}

// GitSourceSyncRes is the response for manually triggering Git metadata discovery.
type GitSourceSyncRes struct {
	Plugin *MarketplacePluginDetailItem `json:"plugin" dc:"Marketplace plugin identity after metadata discovery" eg:"{}"`
	Synced int                          `json:"synced" dc:"Number of draft releases created or refreshed during this discovery run" eg:"2"`
}

// ReleaseDistributionReq is the request for reading CLI install distribution metadata.
type ReleaseDistributionReq struct {
	g.Meta   `path:"/market/plugins/{pluginId}/releases/{version}/distribution" method:"get" tags:"Plugin Marketplace" summary:"Get marketplace release distribution" permission:"market:plugin:download" dc:"Return CLI install distribution metadata for one visible marketplace release. Git releases return repository coordinates without platform tokens; upload releases return checksum metadata and require a download session for package bytes."`
	PluginId string `json:"pluginId" v:"required|length:1,64" dc:"Stable plugin ID whose release distribution is requested" eg:"linapro-demo-source"`
	Version  string `json:"version" v:"required|length:1,32" dc:"Release version whose distribution metadata is requested" eg:"v0.1.0"`
}

// ReleaseDistributionRes is the response for reading CLI install distribution metadata.
type ReleaseDistributionRes struct {
	Distribution *MarketplaceDistributionItem `json:"distribution" dc:"CLI-facing distribution projection for git clone or https download" eg:"{}"`
}

// MyReleaseDistributionReq is the publisher-owned distribution metadata request.
type MyReleaseDistributionReq struct {
	g.Meta   `path:"/market/my-plugins/{pluginId}/releases/{version}/distribution" method:"get" tags:"Plugin Marketplace" summary:"Get owned marketplace release distribution" permission:"market:plugin:publish" dc:"Return distribution metadata for one publisher-owned marketplace release, including draft and unpublished releases owned by the current operator."`
	PluginId string `json:"pluginId" v:"required|length:1,64" dc:"Stable plugin ID owned by the current publisher" eg:"linapro-demo-source"`
	Version  string `json:"version" v:"required|length:1,32" dc:"Release version whose distribution metadata is requested" eg:"v0.1.0"`
}

// MyReleaseDistributionRes is the publisher-owned distribution metadata response.
type MyReleaseDistributionRes struct {
	Distribution *MarketplaceDistributionItem `json:"distribution" dc:"CLI-facing distribution projection for the owned release" eg:"{}"`
}
