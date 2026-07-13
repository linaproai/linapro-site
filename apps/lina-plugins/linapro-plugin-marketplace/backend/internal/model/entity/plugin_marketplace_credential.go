// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"time"
)

// PluginMarketplaceCredential is the golang structure for table plugin_marketplace_credential.
type PluginMarketplaceCredential struct {
	Id            int        `json:"id"            orm:"id"             description:"Primary key ID"`
	CredentialRef string     `json:"credentialRef" orm:"credential_ref" description:"Opaque credential reference stored on plugin records"`
	OwnerUserId   int64      `json:"ownerUserId"   orm:"owner_user_id"  description:"Owning user ID of the credential"`
	Provider      string     `json:"provider"      orm:"provider"       description:"Git provider associated with the credential"`
	CipherText    string     `json:"cipherText"    orm:"cipher_text"    description:"Encrypted token ciphertext; never returned by marketplace APIs"`
	CreatedAt     *time.Time `json:"createdAt"     orm:"created_at"     description:"Creation time"`
	UpdatedAt     *time.Time `json:"updatedAt"     orm:"updated_at"     description:"Update time"`
	DeletedAt     *time.Time `json:"deletedAt"     orm:"deleted_at"     description:"Deletion time"`
}
