// This file returns the menu IDs assigned to one role. Workbenches apply those
// IDs onto the assignable menu list from GET /menu/treeselect.

package menu

import (
	"context"

	v1 "lina-core/api/menu/v1"
)

// RoleMenuTree returns the menu IDs assigned to a role.
func (c *ControllerV1) RoleMenuTree(ctx context.Context, req *v1.RoleMenuTreeReq) (res *v1.RoleMenuTreeRes, err error) {
	out, err := c.menuSvc.GetRoleMenuTree(ctx, req.RoleId)
	if err != nil {
		return nil, err
	}
	return &v1.RoleMenuTreeRes{MenuIds: out.CheckedKeys}, nil
}
