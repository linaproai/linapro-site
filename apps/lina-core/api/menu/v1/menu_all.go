// This file defines DTOs for the authenticated generic navigation resource API.

package v1

import (
	"lina-core/pkg/menuopen"
	"lina-core/pkg/menutype"
	"lina-core/pkg/statusflag"

	"github.com/gogf/gf/v2/frame/g"
)

// GetAllReq defines the request for querying current-user navigation resources.
type GetAllReq struct {
	g.Meta `path:"/menus/all" method:"get" tags:"Menu Management" summary:"Get host navigation resources" dc:"Return the generic navigation resources accessible to the currently logged-in user. Management workbenches compile these resources into shell routes. The payload is not a Vben route record."`
}

// NavResourceItem represents one generic navigation node.
type NavResourceItem struct {
	Id       int                   `json:"id" dc:"Menu ID" eg:"1"`
	ParentId int                   `json:"parentId" dc:"Parent menu ID" eg:"0"`
	MenuKey  string                `json:"menuKey,omitempty" dc:"Stable menu business key" eg:"system:user:list"`
	Title    string                `json:"title" dc:"Localized menu title" eg:"User management"`
	I18nKey  string                `json:"i18nKey,omitempty" dc:"Runtime i18n key used to relocalize the title without refetching navigation" eg:"menu.system:user:list.title"`
	Path     string                `json:"path" dc:"Stored navigation path, not a compiled workbench route slug" eg:"/system/user"`
	Resource string                `json:"resource,omitempty" dc:"Opaque page resource address interpreted by the workbench" eg:"system/user/index"`
	Type     menutype.Code         `json:"type" dc:"Menu type: D=Directory M=Menu B=Button" eg:"M"`
	Icon     string                `json:"icon,omitempty" dc:"Menu icon identifier" eg:"ant-design:user-outlined"`
	Perms    string                `json:"perms,omitempty" dc:"Permission identifier" eg:"system:user:list"`
	Sort     int                   `json:"sort" dc:"Display order, smaller values appear first" eg:"1"`
	Visible  statusflag.Visibility `json:"visible" dc:"Whether to show in navigation: 1=show 0=hide" eg:"1"`
	Status   statusflag.Enabled    `json:"status" dc:"Enabled status: 1=normal 0=disabled" eg:"1"`
	Cache    statusflag.YesNo      `json:"cache" dc:"Whether the workbench may cache the page: 1=yes 0=no" eg:"1"`
	OpenMode menuopen.Mode         `json:"openMode" dc:"Open mode: page=in-app page embedded=hosted asset iframe=iframe external=new window" eg:"page"`
	Target   string                `json:"target,omitempty" dc:"Target URL for embedded, iframe, or external modes" eg:"/x-assets/plugin-runtime-demo/v0.1.0/index.html"`
	Query    map[string]string     `json:"query,omitempty" dc:"Generic query parameters attached when the menu is opened" eg:"{\"tab\":\"overview\"}"`
	Children []*NavResourceItem    `json:"children,omitempty" dc:"Child navigation nodes" eg:"[]"`
}

// GetAllRes defines the wrapped response for user navigation resources.
type GetAllRes struct {
	List []*NavResourceItem `json:"list" dc:"Current-user generic navigation resource list" eg:"[]"`
}
