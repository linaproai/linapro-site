// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"time"
)

// PluginMarketplacePublisher is the golang structure for table plugin_marketplace_publisher.
type PluginMarketplacePublisher struct {
	Id           int        `json:"id"           orm:"id"            description:"Primary key ID"`
	PublisherKey string     `json:"publisherKey" orm:"publisher_key" description:"Stable publisher key"`
	Name         string     `json:"name"         orm:"name"          description:"Publisher display name"`
	Summary      string     `json:"summary"      orm:"summary"       description:"Publisher summary"`
	OwnerUserId  int64      `json:"ownerUserId"  orm:"owner_user_id" description:"Owning user ID"`
	OwnerOrgId   int64      `json:"ownerOrgId"   orm:"owner_org_id"  description:"Owning organization ID, 0 means none"`
	Verified     bool       `json:"verified"     orm:"verified"      description:"Whether the publisher has been verified"`
	Status       string     `json:"status"       orm:"status"        description:"Publisher status: active/suspended"`
	Homepage     string     `json:"homepage"     orm:"homepage"      description:"Publisher homepage URL"`
	ContactEmail string     `json:"contactEmail" orm:"contact_email" description:"Publisher contact email"`
	Remark       string     `json:"remark"       orm:"remark"        description:"Remark"`
	CreatedAt    *time.Time `json:"createdAt"    orm:"created_at"    description:"Creation time"`
	UpdatedAt    *time.Time `json:"updatedAt"    orm:"updated_at"    description:"Update time"`
	DeletedAt    *time.Time `json:"deletedAt"    orm:"deleted_at"    description:"Deletion time"`
}
