// This file returns generic navigation resources for the current user.
// Workbenches compile the payload into shell-specific routes.

package menu

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	v1 "lina-core/api/menu/v1"
	"lina-core/internal/model/entity"
	menusvc "lina-core/internal/service/menu"
	"lina-core/pkg/apitime"
	"lina-core/pkg/menuopen"
	"lina-core/pkg/menutype"
	"lina-core/pkg/statusflag"
)

// GetAll returns generic navigation resources for the current user.
func (c *ControllerV1) GetAll(ctx context.Context, req *v1.GetAllReq) (res *v1.GetAllRes, err error) {
	// Get user ID from business context (set by auth middleware)
	bizCtx := c.bizCtxSvc.Get(ctx)
	if bizCtx == nil {
		return &v1.GetAllRes{List: []*v1.NavResourceItem{}}, nil
	}
	userId := bizCtx.UserId

	// Check if super admin
	isSuperAdmin := c.roleSvc.IsSuperAdmin(ctx, userId)

	var menuTree []*menusvc.MenuItem

	statusNormal := 1
	if isSuperAdmin {
		// Super admin gets all enabled menus
		allMenus, err := c.menuSvc.List(ctx, menusvc.ListInput{
			Status:    &statusNormal,
			Localized: true,
		})
		if err != nil {
			return nil, err
		}
		menuTree = c.menuSvc.BuildTree(allMenus.List)
	} else {
		// Regular user gets menus based on roles
		menuIds, err := c.roleSvc.GetUserMenuIds(ctx, userId)
		if err != nil {
			return nil, err
		}
		if len(menuIds) > 0 {
			allMenus, err := c.menuSvc.List(ctx, menusvc.ListInput{
				Status:    &statusNormal,
				Localized: true,
			})
			if err != nil {
				return nil, err
			}
			menuTree = buildFilteredTree(allMenus.List, menuIds)
		}
	}

	return &v1.GetAllRes{List: convertToNavResources(menuTree)}, nil
}

// buildFilteredTree builds a tree from one flat menu list and automatically
// keeps ancestor directories required by the selected menu IDs.
func buildFilteredTree(allMenus []*entity.SysMenu, selectedIDs []int) []*menusvc.MenuItem {
	if len(allMenus) == 0 || len(selectedIDs) == 0 {
		return []*menusvc.MenuItem{}
	}

	entityMap := make(map[int]*entity.SysMenu, len(allMenus))
	for _, item := range allMenus {
		if item == nil {
			continue
		}
		entityMap[item.Id] = item
	}

	selectedMap := make(map[int]struct{}, len(selectedIDs))
	for _, id := range selectedIDs {
		currentID := id
		for currentID > 0 {
			if _, ok := selectedMap[currentID]; ok {
				break
			}
			selectedMap[currentID] = struct{}{}
			parent, ok := entityMap[currentID]
			if !ok {
				break
			}
			currentID = parent.ParentId
		}
	}

	items := make([]*menusvc.MenuItem, 0, len(selectedMap))
	for _, item := range allMenus {
		if item == nil {
			continue
		}
		if _, ok := selectedMap[item.Id]; !ok {
			continue
		}
		items = append(items, cloneMenuItem(item))
	}

	nodeMap := make(map[int]*menusvc.MenuItem, len(items))
	for _, m := range items {
		nodeMap[m.Id] = m
	}

	var roots []*menusvc.MenuItem
	for _, m := range items {
		if m.ParentId == 0 {
			roots = append(roots, m)
		} else {
			if parent, ok := nodeMap[m.ParentId]; ok {
				parent.Children = append(parent.Children, m)
			}
		}
	}
	return roots
}

// cloneMenuItem detaches one menu item from the service tree so controller-side
// filtering can rebuild children without mutating shared slices.
func cloneMenuItem(item *entity.SysMenu) *menusvc.MenuItem {
	if item == nil {
		return nil
	}
	return &menusvc.MenuItem{
		Id:         item.Id,
		ParentId:   item.ParentId,
		Name:       item.Name,
		MenuKey:    item.MenuKey,
		Path:       item.Path,
		Component:  item.Component,
		Perms:      item.Perms,
		Icon:       item.Icon,
		Type:       item.Type,
		Sort:       item.Sort,
		Visible:    item.Visible,
		Status:     item.Status,
		IsFrame:    item.IsFrame,
		IsCache:    item.IsCache,
		QueryParam: item.QueryParam,
		Remark:     item.Remark,
		CreatedAt:  apitime.Milli(item.CreatedAt),
		UpdatedAt:  apitime.Milli(item.UpdatedAt),
		Children:   []*menusvc.MenuItem{},
	}
}

// convertToNavResources projects menu items into generic navigation resources.
func convertToNavResources(items []*menusvc.MenuItem) []*v1.NavResourceItem {
	result := make([]*v1.NavResourceItem, 0, len(items))
	for _, item := range items {
		if item.Type == menutype.Button.String() {
			continue
		}

		rawQuery := parseMenuQueryParams(item.QueryParam)
		openMode, target := menuopen.Resolve(item.Path, item.IsFrame)
		node := &v1.NavResourceItem{
			Id:       item.Id,
			ParentId: item.ParentId,
			MenuKey:  item.MenuKey,
			Title:    item.Name,
			I18nKey:  buildRouteTitleI18nKey(item.MenuKey, item.Name),
			Path:     item.Path,
			Resource: menuopen.StripWorkbenchResource(item.Component),
			Type:     menutype.Code(item.Type),
			Icon:     item.Icon,
			Perms:    item.Perms,
			Sort:     item.Sort,
			Visible:  statusflag.Visibility(item.Visible),
			Status:   statusflag.Enabled(item.Status),
			Cache:    statusflag.YesNo(item.IsCache),
			OpenMode: openMode,
			Target:   target,
			Query:    menuopen.SanitizeQuery(rawQuery),
		}
		if openMode != menuopen.Page {
			node.Resource = ""
		}

		if len(item.Children) > 0 {
			node.Children = convertToNavResources(item.Children)
		}
		result = append(result, node)
	}
	return result
}

// buildRouteTitleI18nKey derives the runtime i18n key that lets the frontend
// relocalize a route title without refetching the menu tree.
func buildRouteTitleI18nKey(menuKey string, name string) string {
	trimmedMenuKey := strings.TrimSpace(menuKey)
	if trimmedMenuKey != "" {
		return "menu." + trimmedMenuKey + ".title"
	}

	trimmedName := strings.TrimSpace(name)
	if strings.Contains(trimmedName, ".") {
		return trimmedName
	}
	return ""
}

// parseMenuQueryParams decodes the persisted JSON query payload into trimmed
// string pairs used by navigation resources.
func parseMenuQueryParams(queryParam string) map[string]string {
	trimmedQuery := strings.TrimSpace(queryParam)
	if trimmedQuery == "" {
		return nil
	}

	rawQuery := make(map[string]interface{})
	if err := json.Unmarshal([]byte(trimmedQuery), &rawQuery); err != nil {
		return nil
	}

	query := make(map[string]string, len(rawQuery))
	for key, value := range rawQuery {
		trimmedKey := strings.TrimSpace(key)
		if trimmedKey == "" || value == nil {
			continue
		}
		query[trimmedKey] = strings.TrimSpace(fmt.Sprint(value))
	}
	if len(query) == 0 {
		return nil
	}
	return query
}
