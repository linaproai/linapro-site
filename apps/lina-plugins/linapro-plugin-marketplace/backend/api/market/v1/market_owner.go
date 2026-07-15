// This file defines publisher-owned marketplace lifecycle API DTOs for package
// add, publish-to-market, and delist flows used by the "My Plugins" workbench.

package v1

import "github.com/gogf/gf/v2/frame/g"

// PackageAddReq is the request for adding one marketplace plugin from an uploaded package.
type PackageAddReq struct {
	g.Meta `path:"/market/my-plugins/packages" method:"post" mime:"multipart/form-data" tags:"Plugin Marketplace" summary:"Add marketplace plugin package" permission:"market:plugin:publish" dc:"Upload a source or dynamic plugin package for the current publisher. The service unpacks the archive, validates plugin directory structure, parses plugin.yaml identity fields, creates or updates a draft plugin owned by the current user, and stores a draft release without submitting review or listing the plugin in the public catalog."`
	// PublisherKey is required when the package introduces a new plugin identity.
	PublisherKey string `json:"publisherKey" v:"length:0,64" dc:"Publisher key that owns the new plugin identity; required when the package plugin ID does not already belong to the current operator" eg:"linapro"`
	// ReplaceDraft allows replacing an existing mutable draft for the same plugin ID and version.
	ReplaceDraft bool `json:"replaceDraft" d:"true" dc:"Whether an existing mutable draft for the same plugin ID and version may be replaced; defaults to true for the add-plugin workbench" eg:"true"`
}

// PackageAddRes is the response for adding one marketplace plugin from an uploaded package.
type PackageAddRes struct {
	Plugin  *MarketplacePluginDetailItem `json:"plugin" dc:"Plugin identity projection after package add, remaining draft and owner-visible only" eg:"{}"`
	Release *MarketplaceReleaseItem      `json:"release" dc:"Draft release projection parsed from the package, not yet submitted for review" eg:"{}"`
}

// PluginPublishReq is the request for submitting one owned plugin to marketplace review.
type PluginPublishReq struct {
	g.Meta   `path:"/market/my-plugins/{pluginId}/publish" method:"post" tags:"Plugin Marketplace" summary:"Publish owned marketplace plugin for review" permission:"market:plugin:publish" dc:"Submit one publisher-owned plugin for marketplace listing review. Each publish attempt requires administrator approval; the plugin stays invisible in the public catalog until review is approved. Delisted plugins may be published again and must pass review again."`
	PluginId string `json:"pluginId" v:"required|length:1,64" dc:"Stable plugin ID owned by the current publisher" eg:"linapro-demo-source"`
	Version  string `json:"version" v:"length:0,32" dc:"Optional release version to publish; when empty the service selects the latest draft, rejected, or delisted candidate" eg:"v0.1.0"`
	Message  string `json:"message" v:"length:0,1024" dc:"Optional publisher message for marketplace reviewers" eg:"Ready for marketplace listing."`
}

// PluginPublishRes is the response for submitting one owned plugin to marketplace review.
type PluginPublishRes struct {
	Release *MarketplaceReleaseItem `json:"release" dc:"Release projection after review submission or relist submission" eg:"{}"`
}

// PluginDelistReq is the request for delisting one owned published marketplace plugin.
type PluginDelistReq struct {
	g.Meta   `path:"/market/my-plugins/{pluginId}/delist" method:"post" tags:"Plugin Marketplace" summary:"Delist owned marketplace plugin" permission:"market:plugin:publish" dc:"Withdraw one publisher-owned published plugin from the public marketplace catalog. The plugin remains in the owner My Plugins list and may be published again later with a fresh review."`
	PluginId string `json:"pluginId" v:"required|length:1,64" dc:"Stable plugin ID owned by the current publisher" eg:"linapro-demo-source"`
	Message  string `json:"message" v:"length:0,1024" dc:"Optional owner message explaining the delist action" eg:"Temporarily withdrawn from marketplace."`
}

// PluginDelistRes is the response for delisting one owned published marketplace plugin.
type PluginDelistRes struct {
	Plugin *MarketplacePluginDetailItem `json:"plugin" dc:"Plugin detail projection after delist" eg:"{}"`
}
