// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

// PluginMarketplaceCredential is the golang structure of table plugin_marketplace_credential for DAO operations like Where/Data.
type PluginMarketplaceCredential struct {
	g.Meta        `orm:"table:plugin_marketplace_credential, do:true"`
	Id            any        // Primary key ID
	CredentialRef any        // Opaque credential reference stored on plugin records
	OwnerUserId   any        // Owning user ID of the credential
	Provider      any        // Git provider associated with the credential
	CipherText    any        // Encrypted token ciphertext; never returned by marketplace APIs
	CreatedAt     *time.Time // Creation time
	UpdatedAt     *time.Time // Update time
	DeletedAt     *time.Time // Deletion time
}
