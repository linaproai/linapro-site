// Package menuview defines the stable menu projection used across host modules
// so plugin filtering does not leak SysMenu entities.
package menuview

import "lina-core/internal/model/entity"

// FilterItem is the stable menu projection consumed by plugin visibility
// filters. Callers MUST NOT pass SysMenu entities across this boundary.
type FilterItem struct {
	Id        int    // Menu ID.
	ParentId  int    // Parent menu ID.
	Name      string // Display name.
	Path      string // Stored navigation path.
	Component string // Opaque page resource address.
	Perms     string // Permission identifier.
	MenuKey   string // Stable menu business key.
	Type      string // Menu type code.
	Visible   int    // Visibility flag.
	Status    int    // Enabled flag.
}

// FromEntities projects SysMenu rows into the stable filter contract.
func FromEntities(menus []*entity.SysMenu) []FilterItem {
	items := make([]FilterItem, 0, len(menus))
	for _, menu := range menus {
		if menu == nil {
			continue
		}
		items = append(items, FilterItem{
			Id:        menu.Id,
			ParentId:  menu.ParentId,
			Name:      menu.Name,
			Path:      menu.Path,
			Component: menu.Component,
			Perms:     menu.Perms,
			MenuKey:   menu.MenuKey,
			Type:      menu.Type,
			Visible:   menu.Visible,
			Status:    menu.Status,
		})
	}
	return items
}

// ToEntities restores SysMenu rows that survived a projection filter.
func ToEntities(items []FilterItem, menus []*entity.SysMenu) []*entity.SysMenu {
	index := make(map[int]*entity.SysMenu, len(menus))
	for _, menu := range menus {
		if menu == nil {
			continue
		}
		index[menu.Id] = menu
	}
	kept := make([]*entity.SysMenu, 0, len(items))
	for _, item := range items {
		if menu := index[item.Id]; menu != nil {
			kept = append(kept, menu)
		}
	}
	return kept
}
