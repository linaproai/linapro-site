package menu

import (
	"testing"

	"lina-core/internal/model/entity"
	menusvc "lina-core/internal/service/menu"
	"lina-core/pkg/menuopen"
	"lina-core/pkg/menutype"
)

// TestConvertToNavResourcesUsesIframeModeForHostedHTML verifies hosted HTML
// assets become iframe resources without compiling Vben components.
func TestConvertToNavResourcesUsesIframeModeForHostedHTML(t *testing.T) {
	resources := convertToNavResources([]*menusvc.MenuItem{
		{
			Id:      101,
			Name:    "Runtime Iframe Entry",
			Path:    "/x-assets/plugin-runtime-demo/v0.1.0/index.html",
			Type:    menutype.Menu.String(),
			IsFrame: 0,
			Visible: 1,
			Status:  1,
		},
	})

	if len(resources) != 1 {
		t.Fatalf("expected 1 resource, got %d", len(resources))
	}
	item := resources[0]
	if item.OpenMode != menuopen.IFrame {
		t.Fatalf("expected iframe open mode, got %s", item.OpenMode)
	}
	if item.Target != "/x-assets/plugin-runtime-demo/v0.1.0/index.html" {
		t.Fatalf("expected hosted target, got %s", item.Target)
	}
	if item.Resource != "" {
		t.Fatalf("expected empty resource for iframe, got %s", item.Resource)
	}
}

// TestConvertToNavResourcesUsesExternalModeForFrameFlag verifies is_frame
// stored rows become external open mode.
func TestConvertToNavResourcesUsesExternalModeForFrameFlag(t *testing.T) {
	resources := convertToNavResources([]*menusvc.MenuItem{
		{
			Id:      102,
			Name:    "Runtime New Window Entry",
			Path:    "/x-assets/plugin-runtime-demo/v0.1.0/index.html",
			Type:    menutype.Menu.String(),
			IsFrame: 1,
			Visible: 1,
			Status:  1,
		},
	})
	if len(resources) != 1 || resources[0].OpenMode != menuopen.External {
		t.Fatalf("expected external mode, got %#v", resources)
	}
}

// TestConvertToNavResourcesUsesEmbeddedModeForHostedJS verifies hosted ESM
// entries become embedded resources and drop workbench query keys.
func TestConvertToNavResourcesUsesEmbeddedModeForHostedJS(t *testing.T) {
	resources := convertToNavResources([]*menusvc.MenuItem{
		{
			Id:         103,
			Name:       "Runtime Embedded Entry",
			Path:       "/x-assets/plugin-runtime-demo/v0.1.0/mount.js",
			Component:  "system/plugin/dynamic-page",
			Type:       menutype.Menu.String(),
			IsFrame:    0,
			Visible:    1,
			Status:     1,
			QueryParam: `{"pluginAccessMode":"embedded-mount","tab":"overview"}`,
		},
	})
	if len(resources) != 1 {
		t.Fatalf("expected 1 resource, got %d", len(resources))
	}
	item := resources[0]
	if item.OpenMode != menuopen.Embedded {
		t.Fatalf("expected embedded open mode, got %s", item.OpenMode)
	}
	if item.Target != "/x-assets/plugin-runtime-demo/v0.1.0/mount.js" {
		t.Fatalf("expected hosted target, got %s", item.Target)
	}
	if item.Resource != "" {
		t.Fatalf("expected empty resource for embedded, got %s", item.Resource)
	}
	if item.Query["pluginAccessMode"] != "" || item.Query["tab"] != "overview" {
		t.Fatalf("expected workbench query keys stripped, got %#v", item.Query)
	}
}

// TestConvertToNavResourcesStripsWorkbenchPageResource verifies leftover
// workbench shell paths are not emitted as host page resources.
func TestConvertToNavResourcesStripsWorkbenchPageResource(t *testing.T) {
	resources := convertToNavResources([]*menusvc.MenuItem{
		{
			Id:        105,
			Name:      "My Plugins",
			Path:      "plugin-marketplace-mine",
			Component: "system/plugin/dynamic-page",
			Type:      menutype.Menu.String(),
			IsFrame:   0,
			Visible:   1,
			Status:    1,
		},
	})
	if len(resources) != 1 {
		t.Fatalf("expected 1 resource, got %d", len(resources))
	}
	item := resources[0]
	if item.OpenMode != menuopen.Page {
		t.Fatalf("expected page open mode, got %s", item.OpenMode)
	}
	if item.Resource != "" {
		t.Fatalf("expected empty resource for workbench shell path, got %s", item.Resource)
	}
}

// TestConvertToNavResourcesKeepsOpaquePageResource verifies in-app pages keep
// stored resource addresses without #/views prefixes.
func TestConvertToNavResourcesKeepsOpaquePageResource(t *testing.T) {
	resources := convertToNavResources([]*menusvc.MenuItem{
		{
			Id:        104,
			Name:      "User",
			Path:      "/system/user",
			Component: "system/user/index",
			Type:      menutype.Menu.String(),
			IsFrame:   0,
			Visible:   1,
			Status:    1,
		},
	})
	if len(resources) != 1 {
		t.Fatalf("expected 1 resource, got %d", len(resources))
	}
	item := resources[0]
	if item.OpenMode != menuopen.Page || item.Resource != "system/user/index" {
		t.Fatalf("expected page resource, got %#v", item)
	}
	if item.Path != "/system/user" {
		t.Fatalf("expected stored path, got %s", item.Path)
	}
}

// TestConvertToNavResourcesKeepsAbsoluteChildPath verifies grouped directory
// menus keep child stored paths.
func TestConvertToNavResourcesKeepsAbsoluteChildPath(t *testing.T) {
	resources := convertToNavResources([]*menusvc.MenuItem{
		{
			Id:      201,
			Name:    "定时任务",
			Path:    "scheduled-job",
			Type:    menutype.Directory.String(),
			Visible: 1,
			Status:  1,
			Children: []*menusvc.MenuItem{
				{
					Id:        202,
					ParentId:  201,
					Name:      "任务管理",
					Path:      "/system/job",
					Component: "system/job/index",
					Type:      menutype.Menu.String(),
					Visible:   1,
					Status:    1,
				},
			},
		},
	})
	if len(resources) != 1 || len(resources[0].Children) != 1 {
		t.Fatalf("expected directory with child, got %#v", resources)
	}
	if resources[0].Children[0].Path != "/system/job" {
		t.Fatalf("expected absolute child path, got %q", resources[0].Children[0].Path)
	}
}

// TestConvertToNavResourcesDropsButtonsButKeepsEmptyDirectory verifies buttons
// are omitted while empty directories stay for the workbench compiler.
func TestConvertToNavResourcesDropsButtonsButKeepsEmptyDirectory(t *testing.T) {
	resources := convertToNavResources([]*menusvc.MenuItem{
		{
			Id:      301,
			Name:    "系统监控",
			Path:    "monitor",
			Type:    menutype.Directory.String(),
			Visible: 1,
			Status:  1,
			Children: []*menusvc.MenuItem{
				{
					Id:       302,
					ParentId: 301,
					Name:     "操作日志查看",
					Path:     "linapro-monitor-operlog-view",
					Type:     menutype.Button.String(),
					Visible:  1,
					Status:   1,
				},
			},
		},
	})
	if len(resources) != 1 {
		t.Fatalf("expected directory resource, got %#v", resources)
	}
	if len(resources[0].Children) != 0 {
		t.Fatalf("expected buttons omitted, got %#v", resources[0].Children)
	}
}

// TestBuildFilteredTreeKeepsAncestors verifies selected leaf menus project the
// full ancestor chain required by the stable host catalog tree.
func TestBuildFilteredTreeKeepsAncestors(t *testing.T) {
	menuTree := buildFilteredTree([]*entity.SysMenu{
		{Id: 1, Name: "权限管理", Path: "iam", Type: menutype.Directory.String(), Visible: 1, Status: 1},
		{Id: 2, ParentId: 1, Name: "用户治理", Path: "iam-user", Type: menutype.Directory.String(), Visible: 1, Status: 1},
		{Id: 3, ParentId: 2, Name: "用户管理", Path: "/system/user", Component: "system/user/index", Type: menutype.Menu.String(), Visible: 1, Status: 1},
	}, []int{3})

	if len(menuTree) != 1 {
		t.Fatalf("expected one root ancestor, got %#v", menuTree)
	}
	if len(menuTree[0].Children) != 1 {
		t.Fatalf("expected one middle ancestor, got %#v", menuTree[0].Children)
	}
	if len(menuTree[0].Children[0].Children) != 1 {
		t.Fatalf("expected selected leaf to remain attached, got %#v", menuTree[0].Children[0].Children)
	}
	if menuTree[0].Children[0].Children[0].Id != 3 {
		t.Fatalf("expected selected leaf id=3, got %#v", menuTree[0].Children[0].Children[0])
	}
}
