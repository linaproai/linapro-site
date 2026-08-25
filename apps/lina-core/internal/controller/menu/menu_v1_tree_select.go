// This file returns the flat assignable menu resource list used by role
// authorization. Workbenches compile parent/child links into a tree.

package menu

import (
	"context"

	v1 "lina-core/api/menu/v1"
	menusvc "lina-core/internal/service/menu"
	"lina-core/pkg/menutype"
)

// TreeSelect returns the flat assignable menu resource list.
func (c *ControllerV1) TreeSelect(ctx context.Context, req *v1.TreeSelectReq) (res *v1.TreeSelectRes, err error) {
	nodes, err := c.menuSvc.GetTreeSelect(ctx)
	if err != nil {
		return nil, err
	}
	return &v1.TreeSelectRes{List: flattenMenuTreeNodes(nodes)}, nil
}

// flattenMenuTreeNodes projects nested service nodes into a flat resource list.
func flattenMenuTreeNodes(nodes []*menusvc.MenuTreeNode) []*v1.MenuTreeNode {
	items := make([]*v1.MenuTreeNode, 0)
	var walk func([]*menusvc.MenuTreeNode)
	walk = func(list []*menusvc.MenuTreeNode) {
		for _, node := range list {
			if node == nil {
				continue
			}
			items = append(items, &v1.MenuTreeNode{
				Id:       node.Id,
				ParentId: node.ParentId,
				Label:    node.Label,
				Type:     menutype.Code(node.Type),
				Icon:     node.Icon,
			})
			walk(node.Children)
		}
	}
	walk(nodes)
	return items
}
