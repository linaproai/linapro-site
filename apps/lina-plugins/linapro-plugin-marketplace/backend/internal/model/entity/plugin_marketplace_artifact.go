// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"time"
)

// PluginMarketplaceArtifact is the golang structure for table plugin_marketplace_artifact.
type PluginMarketplaceArtifact struct {
	Id             int        `json:"id"             orm:"id"              description:"Primary key ID"`
	ReleaseId      int        `json:"releaseId"      orm:"release_id"      description:"Owning release ID"`
	PluginId       string     `json:"pluginId"       orm:"plugin_id"       description:"Stable plugin ID"`
	ReleaseVersion string     `json:"releaseVersion" orm:"release_version" description:"Plugin release version"`
	ArtifactType   string     `json:"artifactType"   orm:"artifact_type"   description:"Artifact type: source_zip/dynamic_zip/plugin_wasm"`
	StorageKey     string     `json:"storageKey"     orm:"storage_key"     description:"Storage object key or managed file key"`
	FileName       string     `json:"fileName"       orm:"file_name"       description:"Original artifact file name"`
	ContentType    string     `json:"contentType"    orm:"content_type"    description:"Artifact content type"`
	SizeBytes      int64      `json:"sizeBytes"      orm:"size_bytes"      description:"Artifact size in bytes"`
	Sha256         string     `json:"sha256"         orm:"sha256"          description:"Artifact SHA-256 checksum"`
	ManifestSha256 string     `json:"manifestSha256" orm:"manifest_sha256" description:"Root manifest SHA-256 checksum"`
	WasmSha256     string     `json:"wasmSha256"     orm:"wasm_sha256"     description:"Extracted plugin.wasm SHA-256 checksum"`
	CreatedAt      *time.Time `json:"createdAt"      orm:"created_at"      description:"Creation time"`
	UpdatedAt      *time.Time `json:"updatedAt"      orm:"updated_at"      description:"Update time"`
	DeletedAt      *time.Time `json:"deletedAt"      orm:"deleted_at"      description:"Deletion time"`
}
