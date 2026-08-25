// This file covers declared hook execution helpers owned by the integration service.

package integration_test

import (
	"context"
	"testing"

	"lina-core/internal/service/plugin/internal/catalog"
	"lina-core/internal/service/plugin/internal/testutil"
	"lina-core/pkg/plugin/pluginhost"
)

// TestRunPluginDeclaredHookRejectsDemoActions verifies production dispatch
// does not interpret insert/sleep/error demo hooks.
func TestRunPluginDeclaredHookRejectsDemoActions(t *testing.T) {
	services := testutil.NewServices()

	sleepHook := &catalog.HookSpec{
		Event:     pluginhost.ExtensionPointAuthLoginSucceeded,
		Action:    pluginhost.HookActionSleep,
		TimeoutMs: 10,
		SleepMs:   80,
	}
	err := services.Integration.RunPluginDeclaredHook(context.Background(), "plugin-dev-dynamic-timeout", sleepHook, nil)
	if err == nil {
		t.Fatal("expected production dispatch to reject demo sleep hook")
	}

	errorHook := &catalog.HookSpec{
		Event:        pluginhost.ExtensionPointAuthLoginSucceeded,
		Action:       pluginhost.HookActionError,
		ErrorMessage: "runtime hook failed on purpose",
	}
	err = services.Integration.RunPluginDeclaredHook(context.Background(), "plugin-dev-dynamic-error", errorHook, nil)
	if err == nil {
		t.Fatal("expected production dispatch to reject demo error hook")
	}
}
