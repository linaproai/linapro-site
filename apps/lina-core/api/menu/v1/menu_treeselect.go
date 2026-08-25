// This file defines assignable menu resource DTOs used by role authorization.

package v1

import (
	"lina-core/pkg/menutype"

	"github.com/gogf/gf/v2/frame/g"
)

// TreeSelectReq defines the request for querying the menu tree select data.
type TreeSelectReq struct {
	g.Meta `path:"/menu/treeselect" method:"get" tags:"Menu Management" summary:"Get assignable menu resources" dc:"Return the flat assignable menu resource list used when authorizing a role. Management workbenches compile parent/child relationships into a tree. Button-type menus remain included." permission:"system:menu:query"`
}

// MenuTreeNode represents one assignable menu resource.
type MenuTreeNode struct {
	Id       int           `json:"id" dc:"Menu ID" eg:"1"`
	ParentId int           `json:"parentId" dc:"Parent menu ID" eg:"0"`
	Label    string        `json:"label" dc:"Menu name" eg:"System management"`
	Type     menutype.Code `json:"type" dc:"Menu type: D=Directory M=Menu B=Button" eg:"D"`
	Icon     string        `json:"icon" dc:"menu icon" eg:"ant-design:dashboard-outlined"`
}

// TreeSelectRes defines the response for querying assignable menu resources.
type TreeSelectRes struct {
	List []*MenuTreeNode `json:"list" dc:"Flat assignable menu resource list" eg:"[]"`
}
