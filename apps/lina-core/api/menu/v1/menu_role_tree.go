// This file defines the role-assigned menu ID API DTOs.

package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// RoleMenuTreeReq defines the request for querying a role's menu tree.
type RoleMenuTreeReq struct {
	g.Meta `path:"/menu/role/{roleId}" method:"get" tags:"Menu Management" summary:"Get role menu IDs" dc:"Return the menu IDs assigned to this role. Management workbenches compile the assignable menu list from GET /menu/treeselect into a tree and apply these IDs as the checked set." permission:"system:menu:query"`
	RoleId int `json:"roleId" v:"required|min:1" dc:"Role ID" eg:"1"`
}

// RoleMenuTreeRes defines the response for querying a role's assigned menu IDs.
type RoleMenuTreeRes struct {
	MenuIds []int `json:"menuIds" dc:"Assigned menu ID list" eg:"[1,2,3]"`
}
