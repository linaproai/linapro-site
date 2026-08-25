// This file contains supporting menu service rules: runtime localization
// and platform-context checks for global menu writes.

package menu

import (
	"context"
	"strings"

	"lina-core/internal/model/entity"
	"lina-core/pkg/bizerr"
	"lina-core/pkg/plugin/capability/tenantcap"
	"lina-core/pkg/plugin/capability/tenantcap/tenantspi"
)

// localizeMenuEntities localizes one menu-entity list in place.
func (s *serviceImpl) localizeMenuEntities(ctx context.Context, menus []*entity.SysMenu) {
	for _, menu := range menus {
		s.localizeMenuEntity(ctx, menu)
	}
}

// localizeMenuEntity localizes one menu entity in place.
func (s *serviceImpl) localizeMenuEntity(ctx context.Context, menu *entity.SysMenu) {
	if s == nil || s.i18nSvc == nil || menu == nil {
		return
	}
	translationKey := buildMenuTitleKey(menu.MenuKey, menu.Name)
	if translationKey == "" {
		return
	}
	menu.Name = s.i18nSvc.Translate(ctx, translationKey, menu.Name)
}

// buildMenuTitleKey derives the runtime translation key for one menu title.
func buildMenuTitleKey(menuKey string, name string) string {
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

// normalizeMenuIcon trims menu icon input before validation or persistence.
func normalizeMenuIcon(icon string) string {
	return strings.TrimSpace(icon)
}

// ensurePlatformMenuGovernance verifies the current request can mutate the
// global menu topology.
func (s *serviceImpl) ensurePlatformMenuGovernance(ctx context.Context) error {
	if s == nil {
		return nil
	}
	return ensurePlatformMenuGovernanceContext(ctx, s.tenantSvc)
}

// ensurePlatformMenuGovernanceContext applies platform-menu checks to one
// tenant service.
func ensurePlatformMenuGovernanceContext(ctx context.Context, tenantSvc tenantspi.Service) error {
	if tenantSvc == nil || !tenantSvc.Available(ctx) || tenantSvc.PlatformBypass(ctx) {
		return nil
	}
	return bizerr.NewCode(tenantcap.CodePlatformPermissionRequired)
}
