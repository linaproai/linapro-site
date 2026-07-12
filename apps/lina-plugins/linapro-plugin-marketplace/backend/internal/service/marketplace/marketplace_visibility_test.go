// This file verifies marketplace visibility subject normalization and grant
// predicate construction without requiring a live database.

package marketplace

import (
	"reflect"
	"testing"
)

func TestMarketplaceVisibilityScopesNormalizeSubject(t *testing.T) {
	scopes := marketplaceVisibilityScopes(VisibilitySubject{
		UserID:             1001,
		TenantID:           2001,
		OrgIDs:             []int64{3001, 0, 3001, 3002},
		ReservedLicenseIDs: []string{" license-a ", "", "license-a", "license-b"},
	})
	expected := []marketplaceVisibilityScope{
		{scopeType: marketplaceVisibilityScopeTenant, scopeID: "2001"},
		{scopeType: marketplaceVisibilityScopeUser, scopeID: "1001"},
		{scopeType: marketplaceVisibilityScopeOrg, scopeID: "3001"},
		{scopeType: marketplaceVisibilityScopeOrg, scopeID: "3002"},
		{scopeType: marketplaceVisibilityScopeReservedLicense, scopeID: "license-a"},
		{scopeType: marketplaceVisibilityScopeReservedLicense, scopeID: "license-b"},
	}

	if !reflect.DeepEqual(scopes, expected) {
		t.Fatalf("unexpected visibility scopes: %#v", scopes)
	}
}

func TestMarketplaceVisibilityScopesZeroSubjectIsPublicOnly(t *testing.T) {
	if scopes := marketplaceVisibilityScopes(VisibilitySubject{}); len(scopes) != 0 {
		t.Fatalf("expected zero subject to produce no private grant scopes, got %#v", scopes)
	}
}

func TestMarketplaceVisibilityScopeConditionBuildsGrantPredicate(t *testing.T) {
	condition, args := marketplaceVisibilityScopeCondition([]marketplaceVisibilityScope{
		{scopeType: marketplaceVisibilityScopeTenant, scopeID: "2001"},
		{scopeType: marketplaceVisibilityScopeUser, scopeID: "1001"},
	}, "scope_type", "scope_id")

	expectedCondition := "((scope_type = ? AND scope_id = ?) OR (scope_type = ? AND scope_id = ?))"
	if condition != expectedCondition {
		t.Fatalf("unexpected visibility condition: %s", condition)
	}
	expectedArgs := []any{
		marketplaceVisibilityScopeTenant,
		"2001",
		marketplaceVisibilityScopeUser,
		"1001",
	}
	if !reflect.DeepEqual(args, expectedArgs) {
		t.Fatalf("unexpected visibility args: %#v", args)
	}
}

func TestMarketplaceVisibilityGrantCriteriaUsesDownloadPermission(t *testing.T) {
	criteria := marketplaceVisibilityGrantCriteria(marketplaceVisibilityPermissionDownload)
	if criteria.Permission != string(marketplaceVisibilityPermissionDownload) {
		t.Fatalf("expected download permission criteria, got %#v", criteria.Permission)
	}
	if criteria.Status != marketplaceVisibilityGrantStatusEnabled {
		t.Fatalf("expected enabled grant status, got %#v", criteria.Status)
	}
}

func TestNormalizeMarketplaceVisibilityPermissionDefaultsToView(t *testing.T) {
	if got := normalizeMarketplaceVisibilityPermission(""); got != marketplaceVisibilityPermissionView {
		t.Fatalf("expected blank permission to default to view, got %s", got)
	}
	if got := normalizeMarketplaceVisibilityPermission(marketplaceVisibilityPermissionDownload); got != marketplaceVisibilityPermissionDownload {
		t.Fatalf("expected download permission to be preserved, got %s", got)
	}
}

func TestMarketplaceEffectiveVisibilityUsesRestrictiveReleaseBoundary(t *testing.T) {
	cases := []struct {
		name              string
		pluginVisibility  string
		releaseVisibility string
		expected          string
	}{
		{
			name:              "public plugin and public release",
			pluginVisibility:  "public",
			releaseVisibility: "public",
			expected:          "public",
		},
		{
			name:              "public plugin and private release",
			pluginVisibility:  "public",
			releaseVisibility: "private",
			expected:          "private",
		},
		{
			name:              "private plugin and public release",
			pluginVisibility:  "private",
			releaseVisibility: "public",
			expected:          "private",
		},
		{
			name:              "private plugin and reserved release",
			pluginVisibility:  "private",
			releaseVisibility: "reserved",
			expected:          "reserved",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := marketplaceEffectiveVisibility(tc.pluginVisibility, tc.releaseVisibility); got != tc.expected {
				t.Fatalf("expected %s, got %s", tc.expected, got)
			}
		})
	}
}
