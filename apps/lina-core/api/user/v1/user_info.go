// This file defines the current-user information API DTOs.

package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// GetInfoReq defines the request for querying current frontend user info.
type GetInfoReq struct {
	g.Meta `path:"/user/info" method:"get" tags:"User Management" summary:"Get host user context" dc:"Obtain the identity, roles, permissions, landing path, and optional capability flags of the currently logged-in user. Navigation resources are returned by GET /menus/all."`
}

// GetInfoRes defines the response for querying current frontend user info.
type GetInfoRes struct {
	UserId              int      `json:"userId" dc:"User ID" eg:"1"`
	Username            string   `json:"username" dc:"Username" eg:"admin"`
	RealName            string   `json:"realName" dc:"Real name (nickname)" eg:"Administrator"`
	Email               string   `json:"email" dc:"Email address" eg:"admin@example.com"`
	Avatar              string   `json:"avatar" dc:"Avatar address" eg:"/upload/avatar/default.png"`
	Roles               []string `json:"roles" dc:"List of user role identifiers" eg:"['admin','user']"`
	HomePath            string   `json:"homePath" dc:"Home path resolved from accessible navigation resources" eg:"/dashboard"`
	Permissions         []string `json:"permissions" dc:"List of user effective permission identifiers, including menu permissions and button permissions, used for interface declaration permission verification and button-level permission control" eg:"['system:user:list','system:user:add','system:user:edit']"`
	OrganizationEnabled bool     `json:"organizationEnabled" dc:"Whether organization management is available for the current session" eg:"true"`
	TenantEnabled       bool     `json:"tenantEnabled" dc:"Whether tenant management is available for the current session" eg:"false"`
}
