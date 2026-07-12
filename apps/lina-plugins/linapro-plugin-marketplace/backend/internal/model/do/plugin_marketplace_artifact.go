// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

// PluginMarketplaceArtifact is the golang structure of table plugin_marketplace_artifact for DAO operations like Where/Data.
type PluginMarketplaceArtifact struct {
	g.Meta         `orm:"table:plugin_marketplace_artifact, do:true"`
	Id             any        // Primary key ID
	ReleaseId      any        // Owning release ID
	PluginId       any        // Stable plugin ID
	ReleaseVersion any        // Plugin release version
	ArtifactType   any        // Artifact type: source_zip/dynamic_zip/plugin_wasm
	StorageKey     any        // Storage object key or managed file key
	FileName       any        // Original artifact file name
	ContentType    any        // Artifact content type
	SizeBytes      any        // Artifact size in bytes
	Sha256         any        // Artifact SHA-256 checksum
	ManifestSha256 any        // Root manifest SHA-256 checksum
	WasmSha256     any        // Extracted plugin.wasm SHA-256 checksum
	CreatedAt      *time.Time // Creation time
	UpdatedAt      *time.Time // Update time
	DeletedAt      *time.Time // Deletion time
}
