// This file covers pluginhost enum helpers used by public contracts.

package pluginhost

import "testing"

func TestAuthHookReasonIsValid(t *testing.T) {
	t.Parallel()

	if !AuthHookReasonLoginSuccessful.IsValid() {
		t.Fatal("login successful should be valid")
	}
	if !AuthHookReasonTenantUnavailable.IsValid() {
		t.Fatal("tenant unavailable should be valid")
	}
	if !AuthHookReasonExternalNotProvisioned.IsValid() {
		t.Fatal("external not provisioned should be valid")
	}
	if AuthHookReason("").IsValid() {
		t.Fatal("empty reason should be invalid")
	}
	if AuthHookReason("unknown").IsValid() {
		t.Fatal("unknown reason should be invalid")
	}
}

func TestDynamicAccessModeIsValid(t *testing.T) {
	t.Parallel()

	if !DynamicAccessModeEmbeddedMount.IsValid() {
		t.Fatal("embedded-mount should be valid")
	}
	if DynamicAccessMode("").IsValid() || DynamicAccessMode("iframe").IsValid() {
		t.Fatal("empty or unknown access mode should be invalid")
	}
}

func TestPluginInstallModeIsValid(t *testing.T) {
	t.Parallel()

	if !PluginInstallModeGlobal.IsValid() || !PluginInstallModeTenantScoped.IsValid() {
		t.Fatal("known install modes should be valid")
	}
	if PluginInstallMode("").IsValid() || PluginInstallMode("mixed").IsValid() {
		t.Fatal("empty or unknown install mode should be invalid")
	}
}

func TestJobScopeIsValid(t *testing.T) {
	t.Parallel()

	cases := []struct {
		scope JobScope
		want  bool
	}{
		{JobScopeMasterOnly, true},
		{JobScopeAllNode, true},
		{"", false},
		{"Master_Only", false},
		{"all", false},
	}
	for _, tc := range cases {
		if got := tc.scope.IsValid(); got != tc.want {
			t.Fatalf("JobScope(%q).IsValid()=%v want %v", tc.scope, got, tc.want)
		}
	}
}

func TestJobConcurrencyIsValid(t *testing.T) {
	t.Parallel()

	cases := []struct {
		mode JobConcurrency
		want bool
	}{
		{JobConcurrencySingleton, true},
		{JobConcurrencyParallel, true},
		{"", false},
		{"single", false},
	}
	for _, tc := range cases {
		if got := tc.mode.IsValid(); got != tc.want {
			t.Fatalf("JobConcurrency(%q).IsValid()=%v want %v", tc.mode, got, tc.want)
		}
	}
}
