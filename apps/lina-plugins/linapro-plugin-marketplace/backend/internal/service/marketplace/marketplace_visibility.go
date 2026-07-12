// This file centralizes marketplace visibility predicates. List, detail,
// release, document, risk, and download queries all use the same
// database-side public/private/reserved grant filter so unauthorized callers do
// not learn private plugin existence from counts, versions, risks, or download
// metadata.

package marketplace

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/gogf/gf/v2/database/gdb"

	marketv1 "linapro-plugin-marketplace/backend/api/market/v1"
	"linapro-plugin-marketplace/backend/internal/dao"
	"linapro-plugin-marketplace/backend/internal/model/do"
)

// marketplaceVisibilityPermission identifies which grant permission a query needs.
type marketplaceVisibilityPermission string

const (
	// marketplaceVisibilityPermissionView allows catalog, detail, document, and risk reads.
	marketplaceVisibilityPermissionView marketplaceVisibilityPermission = "view"
	// marketplaceVisibilityPermissionDownload allows download-session creation and reads.
	marketplaceVisibilityPermissionDownload marketplaceVisibilityPermission = "download"

	// marketplaceVisibilityScopeTenant grants access to all callers in one tenant.
	marketplaceVisibilityScopeTenant = "tenant"
	// marketplaceVisibilityScopeOrg grants access to callers in one resolved organization.
	marketplaceVisibilityScopeOrg = "org"
	// marketplaceVisibilityScopeUser grants access to one specific user.
	marketplaceVisibilityScopeUser = "user"
	// marketplaceVisibilityScopeReservedLicense grants access through a future license scope.
	marketplaceVisibilityScopeReservedLicense = "reserved_license"

	// marketplaceVisibilityGrantStatusEnabled marks an active visibility grant row.
	marketplaceVisibilityGrantStatusEnabled = 1
)

// marketplaceVisibilityScope is one normalized grant lookup scope.
type marketplaceVisibilityScope struct {
	scopeType string
	scopeID   string
}

// applyMarketplaceVisibilityFilter injects public visibility and matching
// private/reserved grant predicates into an existing query model.
func (s *serviceImpl) applyMarketplaceVisibilityFilter(
	ctx context.Context,
	model *gdb.Model,
	visibilityColumn string,
	pluginRecordIDColumn string,
	subject VisibilitySubject,
	permission marketplaceVisibilityPermission,
) *gdb.Model {
	scopes := marketplaceVisibilityScopes(subject)
	if len(scopes) == 0 {
		return model.Where(visibilityColumn, marketv1.MarketplaceVisibilityPublic.String())
	}
	grantModel := marketplaceVisibilityGrantModel(ctx, scopes, permission)
	builder := model.Builder().
		Where(visibilityColumn, marketv1.MarketplaceVisibilityPublic.String()).
		WhereOrIn(pluginRecordIDColumn, grantModel)
	return model.Where(builder)
}

// marketplaceVisibilityGrantModel builds the grant subquery used by visibility filters.
func marketplaceVisibilityGrantModel(
	ctx context.Context,
	scopes []marketplaceVisibilityScope,
	permission marketplaceVisibilityPermission,
) *gdb.Model {
	cols := dao.PluginMarketplaceVisibilityGrant.Columns()
	scopeCondition, scopeArgs := marketplaceVisibilityScopeCondition(scopes, cols.ScopeType, cols.ScopeId)
	model := dao.PluginMarketplaceVisibilityGrant.Ctx(ctx).
		Fields(cols.PluginRecordId).
		Where(marketplaceVisibilityGrantCriteria(permission)).
		Where("("+cols.ExpiresAt+" IS NULL OR "+cols.ExpiresAt+" > ?)", time.Now())
	if scopeCondition == "" {
		return model.Where(cols.Id, 0)
	}
	return model.Where(scopeCondition, scopeArgs...)
}

// marketplaceVisibilityGrantCriteria restricts grants to enabled rows for one permission.
func marketplaceVisibilityGrantCriteria(permission marketplaceVisibilityPermission) do.PluginMarketplaceVisibilityGrant {
	return do.PluginMarketplaceVisibilityGrant{
		Permission: string(normalizeMarketplaceVisibilityPermission(permission)),
		Status:     marketplaceVisibilityGrantStatusEnabled,
	}
}

// marketplaceVisibilityScopeCondition builds one grouped OR condition for grant scope matching.
func marketplaceVisibilityScopeCondition(
	scopes []marketplaceVisibilityScope,
	scopeTypeColumn string,
	scopeIDColumn string,
) (string, []any) {
	if len(scopes) == 0 {
		return "", nil
	}
	clauses := make([]string, 0, len(scopes))
	args := make([]any, 0, len(scopes)*2)
	for _, scope := range scopes {
		if scope.scopeType == "" || scope.scopeID == "" {
			continue
		}
		clauses = append(clauses, "("+scopeTypeColumn+" = ? AND "+scopeIDColumn+" = ?)")
		args = append(args, scope.scopeType, scope.scopeID)
	}
	if len(clauses) == 0 {
		return "", nil
	}
	return "(" + strings.Join(clauses, " OR ") + ")", args
}

// marketplaceVisibilityScopes converts the caller snapshot to grant lookup scopes.
func marketplaceVisibilityScopes(subject VisibilitySubject) []marketplaceVisibilityScope {
	scopes := make([]marketplaceVisibilityScope, 0, 2+len(subject.OrgIDs)+len(subject.ReservedLicenseIDs))
	if subject.TenantID > 0 {
		scopes = append(scopes, marketplaceVisibilityScope{
			scopeType: marketplaceVisibilityScopeTenant,
			scopeID:   strconv.FormatInt(subject.TenantID, 10),
		})
	}
	if subject.UserID > 0 {
		scopes = append(scopes, marketplaceVisibilityScope{
			scopeType: marketplaceVisibilityScopeUser,
			scopeID:   strconv.FormatInt(subject.UserID, 10),
		})
	}
	for _, orgID := range uniqueInt64s(subject.OrgIDs) {
		if orgID <= 0 {
			continue
		}
		scopes = append(scopes, marketplaceVisibilityScope{
			scopeType: marketplaceVisibilityScopeOrg,
			scopeID:   strconv.FormatInt(orgID, 10),
		})
	}
	for _, licenseID := range uniqueNormalizedStrings(subject.ReservedLicenseIDs) {
		scopes = append(scopes, marketplaceVisibilityScope{
			scopeType: marketplaceVisibilityScopeReservedLicense,
			scopeID:   licenseID,
		})
	}
	return scopes
}

// normalizeMarketplaceVisibilityPermission defaults blank permission checks to view.
func normalizeMarketplaceVisibilityPermission(permission marketplaceVisibilityPermission) marketplaceVisibilityPermission {
	if strings.TrimSpace(string(permission)) == "" {
		return marketplaceVisibilityPermissionView
	}
	return permission
}

// marketplaceEffectiveVisibility returns the restrictive list/detail visibility for a plugin release.
func marketplaceEffectiveVisibility(pluginVisibility string, releaseVisibility string) string {
	pluginValue := normalizeVisibility(marketv1.MarketplaceVisibility(pluginVisibility))
	releaseValue := normalizeVisibility(marketv1.MarketplaceVisibility(releaseVisibility))
	if pluginValue == marketv1.MarketplaceVisibilityPublic {
		return releaseValue.String()
	}
	if releaseValue == marketv1.MarketplaceVisibilityPublic {
		return pluginValue.String()
	}
	if pluginValue == marketv1.MarketplaceVisibilityReserved ||
		releaseValue == marketv1.MarketplaceVisibilityReserved {
		return marketv1.MarketplaceVisibilityReserved.String()
	}
	return marketv1.MarketplaceVisibilityPrivate.String()
}

// uniqueInt64s removes invalid duplicate numeric scope identifiers.
func uniqueInt64s(values []int64) []int64 {
	out := make([]int64, 0, len(values))
	seen := make(map[int64]struct{}, len(values))
	for _, value := range values {
		if value <= 0 {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	return out
}

// uniqueNormalizedStrings removes blank duplicate string scope identifiers.
func uniqueNormalizedStrings(values []string) []string {
	out := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		normalized := normalizeKey(value)
		if normalized == "" {
			continue
		}
		if _, ok := seen[normalized]; ok {
			continue
		}
		seen[normalized] = struct{}{}
		out = append(out, normalized)
	}
	return out
}
