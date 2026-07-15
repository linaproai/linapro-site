// This file defines marketplace publisher, plugin draft, upload, and review API
// DTOs. Upload requests describe business form fields only; controllers read the
// multipart file from the standard "file" upload field used by the host.

package v1

import "github.com/gogf/gf/v2/frame/g"

// PublisherListReq is the request for querying publishers available to the current operator.
type PublisherListReq struct {
	g.Meta   `path:"/market/publishers" method:"get" tags:"Plugin Marketplace" summary:"Query marketplace publishers" permission:"market:plugin:publish" dc:"Return publisher profiles that the current operator may use for plugin publishing. The response is bounded and does not expose publishers outside the operator ownership or review authority."`
	PageNum  int    `json:"pageNum" d:"1" v:"min:1" dc:"Page number for marketplace publisher list, starting from 1" eg:"1"`
	PageSize int    `json:"pageSize" d:"20" v:"min:1|max:100" dc:"Page size for marketplace publisher list, default 20 and maximum 100" eg:"20"`
	Keyword  string `json:"keyword" dc:"Optional fuzzy search keyword matched against publisher key, name, and summary; query all allowed publishers when empty" eg:"linapro"`
}

// PublisherListRes is the response for querying publishers available to the current operator.
type PublisherListRes struct {
	List  []*MarketplacePublisherItem `json:"list" dc:"Paginated publisher profiles available to the current operator" eg:"[]"`
	Total int                         `json:"total" dc:"Total number of allowed publishers matching the filters before pagination" eg:"1"`
}

// PublisherCreateReq is the request for creating a marketplace publisher profile.
type PublisherCreateReq struct {
	g.Meta       `path:"/market/publishers" method:"post" tags:"Plugin Marketplace" summary:"Create marketplace publisher" permission:"market:plugin:publish" dc:"Create a marketplace publisher profile for plugin publishing. Created publishers are scoped to the current operator or organization and require review authority before they are marked verified."`
	PublisherKey string `json:"publisherKey" v:"required|length:1,64" dc:"Stable publisher key used for ownership and filtering" eg:"linapro"`
	Name         string `json:"name" v:"required|length:1,128" dc:"Publisher display name shown in marketplace catalog pages" eg:"LinaPro"`
	Summary      string `json:"summary" v:"length:0,512" dc:"Short publisher summary for marketplace trust and discovery display" eg:"Official LinaPro plugin publisher"`
	Homepage     string `json:"homepage" v:"length:0,512" dc:"Publisher homepage URL, empty when the publisher does not provide one" eg:"https://linapro.ai"`
	ContactEmail string `json:"contactEmail" v:"length:0,128" dc:"Publisher contact email used by marketplace reviewers, empty when unavailable" eg:"plugins@linapro.ai"`
}

// PublisherCreateRes is the response for creating a marketplace publisher profile.
type PublisherCreateRes struct {
	Publisher *MarketplacePublisherItem `json:"publisher" dc:"Created marketplace publisher profile" eg:"{}"`
}

// PublisherUpdateReq is the request for updating the current operator's marketplace publisher profile.
type PublisherUpdateReq struct {
	g.Meta `path:"/market/publishers/{publisherKey}" method:"put" tags:"Plugin Marketplace" summary:"Update marketplace publisher" permission:"market:plugin:publish" dc:"Update the marketplace publisher profile owned by the current operator. The URL path publisher key locates the owned profile; the JSON body publisherKey is the desired key after update and may rename the profile when it differs and remains unique."`
	// PathPublisherKey is the current key from the URL path used only for ownership lookup.
	PathPublisherKey string `p:"publisherKey" json:"-" v:"required|length:1,64" dc:"Current publisher key in the URL path used to locate the owned profile" eg:"linapro"`
	// PublisherKey is the desired key after update from the JSON body.
	PublisherKey string `json:"publisherKey" v:"required|length:1,64" dc:"Desired publisher key after update; may differ from the path key to rename the profile when the new key is still unique" eg:"linapro"`
	Name         string `json:"name" v:"required|length:1,128" dc:"Publisher display name shown in marketplace catalog pages" eg:"LinaPro"`
	Summary      string `json:"summary" v:"length:0,512" dc:"Short publisher summary for marketplace trust and discovery display" eg:"Official LinaPro plugin publisher"`
	Homepage     string `json:"homepage" v:"length:0,512" dc:"Publisher homepage URL, empty when the publisher does not provide one" eg:"https://linapro.ai"`
	ContactEmail string `json:"contactEmail" v:"length:0,128" dc:"Publisher contact email used by marketplace reviewers, empty when unavailable" eg:"plugins@linapro.ai"`
}

// PublisherUpdateRes is the response for updating the current operator's marketplace publisher profile.
type PublisherUpdateRes struct {
	Publisher *MarketplacePublisherItem `json:"publisher" dc:"Updated marketplace publisher profile" eg:"{}"`
}

// PluginCreateReq is the request for creating or updating a marketplace plugin draft.
type PluginCreateReq struct {
	g.Meta       `path:"/market/plugins" method:"post" tags:"Plugin Marketplace" summary:"Create marketplace plugin draft" permission:"market:plugin:publish" dc:"Create a marketplace plugin identity draft and bind it to a publisher before release upload. The plugin ID ownership check prevents another publisher from silently taking over an existing marketplace plugin ID."`
	PublisherKey string                `json:"publisherKey" v:"required|length:1,64" dc:"Stable publisher key that owns the marketplace plugin identity" eg:"linapro"`
	PluginId     string                `json:"pluginId" v:"required|length:1,64" dc:"Stable plugin ID to publish; the same plugin ID can only be owned by one publisher after first successful publication" eg:"linapro-demo-source"`
	Name         string                `json:"name" v:"required|length:1,128" dc:"Marketplace plugin display name" eg:"Source Plugin Demo"`
	Summary      string                `json:"summary" v:"required|length:1,512" dc:"Short marketplace summary suitable for list cards and tables" eg:"Source plugin that provides workflow automation examples"`
	Description  string                `json:"description" dc:"Long marketplace description displayed on the detail page" eg:"Provides workflow automation pages, backend APIs, and plugin resource examples."`
	PluginType   MarketplacePluginType `json:"pluginType" v:"required" dc:"Plugin type: source=source plugin dynamic=dynamic plugin" eg:"source"`
	Visibility   MarketplaceVisibility `json:"visibility" d:"public" dc:"Initial visibility policy: public, private, or reserved; defaults to public" eg:"public"`
	Icon         string                `json:"icon" v:"length:0,512" dc:"Marketplace icon URL or plugin-managed asset path, empty when no icon is available" eg:"/assets/marketplace/source-demo.png"`
	Homepage     string                `json:"homepage" v:"length:0,512" dc:"Plugin homepage URL, empty when no homepage is available" eg:"https://linapro.ai/plugins/source-demo"`
	Repository   string                `json:"repository" v:"length:0,512" dc:"Plugin source repository URL, empty when no repository is available" eg:"https://github.com/linaproai/linapro-demo-source"`
	License      string                `json:"license" v:"length:0,64" dc:"Plugin license identifier displayed before download" eg:"Apache-2.0"`
	TagCodes     []string              `json:"tagCodes" dc:"Category and tag codes associated with this marketplace plugin" eg:"[\"observability\",\"audit\"]"`
}

// PluginCreateRes is the response for creating or updating a marketplace plugin draft.
type PluginCreateRes struct {
	Plugin *MarketplacePluginDetailItem `json:"plugin" dc:"Created or updated marketplace plugin draft detail projection" eg:"{}"`
}

// ReleaseUploadReq is the request for uploading one marketplace release package.
type ReleaseUploadReq struct {
	g.Meta         `path:"/market/plugins/{pluginId}/releases" method:"post" mime:"multipart/form-data" tags:"Plugin Marketplace" summary:"Upload marketplace release package" permission:"market:plugin:publish" dc:"Upload a marketplace release ZIP package for a source plugin or dynamic runtime plugin. The controller reads the file field named file, validates package structure, parses plugin.yaml and dynamic metadata when present, stores a draft release, and returns scanner summaries without publishing the release."`
	PluginId       string                `json:"pluginId" v:"required|length:1,64" dc:"Stable plugin ID that owns the uploaded release package" eg:"linapro-demo-source"`
	Version        string                `json:"version" v:"required|length:1,32" dc:"Release version to create or replace while the release is still a draft" eg:"v0.1.0"`
	PluginType     MarketplacePluginType `json:"pluginType" v:"required" dc:"Package plugin type: source=source plugin ZIP dynamic=dynamic runtime ZIP" eg:"source"`
	Visibility     MarketplaceVisibility `json:"visibility" d:"public" dc:"Release visibility policy: public, private, or reserved; defaults to the plugin visibility when empty" eg:"public"`
	MinHostVersion string                `json:"minHostVersion" v:"length:0,32" dc:"Minimum compatible LinaPro host version declared by the uploaded release" eg:"v0.1.0"`
	MaxHostVersion string                `json:"maxHostVersion" v:"length:0,32" dc:"Maximum compatible LinaPro host version declared by the uploaded release, empty when no upper bound exists" eg:"v1.0.0"`
	ReplaceDraft   bool                  `json:"replaceDraft" dc:"Whether an existing draft artifact for the same plugin ID and version may be replaced before publication; published releases are never overwritten" eg:"true"`
}

// ReleaseUploadRes is the response for uploading one marketplace release package.
type ReleaseUploadRes struct {
	Release *MarketplaceReleaseItem `json:"release" dc:"Draft release projection with scanner and artifact checksum summary" eg:"{}"`
}

// ReleaseSubmitReviewReq is the request for submitting one draft release for review.
type ReleaseSubmitReviewReq struct {
	g.Meta   `path:"/market/plugins/{pluginId}/releases/{version}/submit-review" method:"post" tags:"Plugin Marketplace" summary:"Submit marketplace release review" permission:"market:plugin:publish" dc:"Submit one draft marketplace release for reviewer processing. The service verifies plugin ID ownership, draft mutability, package scan completeness, documentation entry, and version uniqueness before changing the review state."`
	PluginId string `json:"pluginId" v:"required|length:1,64" dc:"Stable plugin ID whose draft release is submitted for review" eg:"linapro-demo-source"`
	Version  string `json:"version" v:"required|length:1,32" dc:"Draft release version submitted for review" eg:"v0.1.0"`
	Message  string `json:"message" v:"length:0,1024" dc:"Optional publisher message for marketplace reviewers" eg:"Initial public release with documentation and SQL review notes."`
}

// ReleaseSubmitReviewRes is the response for submitting one draft release for review.
type ReleaseSubmitReviewRes struct {
	Release *MarketplaceReleaseItem `json:"release" dc:"Release projection after review submission" eg:"{}"`
}

// ReleaseReviewReq is the request for approving or rejecting one marketplace release review.
type ReleaseReviewReq struct {
	g.Meta       `path:"/market/plugins/{pluginId}/releases/{version}/review" method:"put" tags:"Plugin Marketplace" summary:"Review marketplace release" permission:"market:plugin:review" dc:"Approve or reject one submitted marketplace release. Approval publishes an immutable visible release after permission and visibility checks; rejection records a reviewer message without exposing private release details to unauthorized users."`
	PluginId     string                  `json:"pluginId" v:"required|length:1,64" dc:"Stable plugin ID whose release is reviewed" eg:"linapro-demo-source"`
	Version      string                  `json:"version" v:"required|length:1,32" dc:"Release version being reviewed" eg:"v0.1.0"`
	ReviewStatus MarketplaceReviewStatus `json:"reviewStatus" v:"required" dc:"Review decision to apply: approved or rejected" eg:"approved"`
	Message      string                  `json:"message" v:"length:0,1024" dc:"Reviewer message stored with the release review state" eg:"Approved after SQL and hostServices review."`
}

// ReleaseReviewRes is the response for approving or rejecting one marketplace release review.
type ReleaseReviewRes struct {
	Release *MarketplaceReleaseItem `json:"release" dc:"Release projection after reviewer decision" eg:"{}"`
}

// PluginStatusUpdateReq is the request for updating one marketplace plugin lifecycle status.
type PluginStatusUpdateReq struct {
	g.Meta   `path:"/market/plugins/{pluginId}/status" method:"put" tags:"Plugin Marketplace" summary:"Update marketplace plugin status" permission:"market:plugin:review" dc:"Update the marketplace lifecycle state of one plugin, such as delisting or deprecating a published plugin. The service applies visibility and reviewer authority checks before changing the status and refreshing the read model."`
	PluginId string            `json:"pluginId" v:"required|length:1,64" dc:"Stable plugin ID whose marketplace status is updated" eg:"linapro-demo-source"`
	Status   MarketplaceStatus `json:"status" v:"required" dc:"Target marketplace status: published, delisted, or deprecated; draft is only used before first publication" eg:"delisted"`
	Message  string            `json:"message" v:"length:0,1024" dc:"Optional reviewer or operator message explaining the status change" eg:"Delisted because the plugin is superseded by a newer package."`
}

// PluginStatusUpdateRes is the response for updating one marketplace plugin lifecycle status.
type PluginStatusUpdateRes struct {
	Plugin *MarketplacePluginDetailItem `json:"plugin" dc:"Marketplace plugin detail projection after status update" eg:"{}"`
}
