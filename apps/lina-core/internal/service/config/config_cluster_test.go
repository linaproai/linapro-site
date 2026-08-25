// This file verifies cluster topology configuration loading and default
// election fallback behavior.

package config

import (
	"context"
	"testing"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gcfg"

	"lina-core/internal/service/cachecoord"
	"lina-core/internal/service/cluster"
	"lina-core/pkg/dialect"
)

// TestGetClusterUsesClusterElectionConfig verifies nested cluster election
// settings are loaded from config content.
func TestGetClusterUsesClusterElectionConfig(t *testing.T) {
	setTestConfigContent(t, `
cluster:
  enabled: true
  election:
    lease: 45s
    renewInterval: 15s
  coordination:
    backend: redis
    group: default
redis:
  default:
    address: "127.0.0.1:6379"
`)

	cfg := New(nil).GetCluster(context.Background())

	if !cfg.Enabled {
		t.Fatal("expected cluster mode to be enabled")
	}
	if cfg.Election.Lease != 45*time.Second {
		t.Fatalf("expected cluster lease to be 45s, got %s", cfg.Election.Lease)
	}
	if cfg.Election.RenewInterval != 15*time.Second {
		t.Fatalf("expected cluster renew interval to be 15s, got %s", cfg.Election.RenewInterval)
	}
	if cfg.Coordination.Backend != cluster.CoordinationBackendRedis {
		t.Fatalf("expected redis coordination backend, got %q", cfg.Coordination.Backend)
	}
	if cfg.Coordination.Group != cluster.DefaultCoordinationGroup {
		t.Fatalf("expected default redis group, got %q", cfg.Coordination.Group)
	}
}

// TestGetClusterUsesDefaultsWhenElectionConfigMissing verifies election timing
// falls back to defaults when the nested section is absent.
func TestGetClusterUsesDefaultsWhenElectionConfigMissing(t *testing.T) {
	setTestConfigContent(t, `
cluster:
  enabled: false
`)

	cfg := New(nil).GetCluster(context.Background())

	if cfg.Enabled {
		t.Fatal("expected cluster mode to be disabled")
	}
	if cfg.Election.Lease != 30*time.Second {
		t.Fatalf("expected default lease to be 30s, got %s", cfg.Election.Lease)
	}
	if cfg.Election.RenewInterval != 10*time.Second {
		t.Fatalf("expected default renew interval to be 10s, got %s", cfg.Election.RenewInterval)
	}
}

// TestGetClusterIgnoresRootElectionConfig verifies only cluster.election
// affects cluster election defaults.
func TestGetClusterIgnoresRootElectionConfig(t *testing.T) {
	setTestConfigContent(t, `
election:
  lease: 50s
  renewInterval: 20s
`)

	cfg := New(nil).GetCluster(context.Background())

	if cfg.Enabled {
		t.Fatal("expected cluster mode to remain disabled by default")
	}
	if cfg.Election.Lease != 30*time.Second {
		t.Fatalf("expected default lease to remain 30s, got %s", cfg.Election.Lease)
	}
	if cfg.Election.RenewInterval != 10*time.Second {
		t.Fatalf("expected default renew interval to remain 10s, got %s", cfg.Election.RenewInterval)
	}
}

// TestOverrideClusterEnabledForDialect verifies a dialect can lock cluster
// mode off in memory regardless of the configured cluster.enabled value.
func TestOverrideClusterEnabledForDialect(t *testing.T) {
	setTestConfigContent(t, `
cluster:
  enabled: true
  election:
    lease: 45s
    renewInterval: 15s
  coordination:
    backend: redis
    group: default
redis:
  default:
    address: "127.0.0.1:6379"
`)

	svc := New(nil)
	if !svc.IsClusterEnabled(context.Background()) {
		t.Fatal("expected config to enable cluster mode before dialect override")
	}

	svc.OverrideClusterEnabledForDialect(false)
	if svc.IsClusterEnabled(context.Background()) {
		t.Fatal("expected dialect override to force cluster mode off")
	}

	cfg := svc.GetCluster(context.Background())
	if cfg.Enabled {
		t.Fatal("expected GetCluster to reflect dialect cluster override")
	}
	if cfg.Election.Lease != 45*time.Second {
		t.Fatalf("expected election lease to be preserved, got %s", cfg.Election.Lease)
	}
}

// TestOverrideClusterEnabledForDialectReselectsRuntimeParamRevisionController
// verifies dialect startup overrides can force a config service constructed
// from cluster.enabled=true back to local runtime-parameter revision handling.
func TestOverrideClusterEnabledForDialectReselectsRuntimeParamRevisionController(t *testing.T) {
	svc := &serviceImpl{}
	svc.runtimeParamRevisionCtrl = newCacheCoordRuntimeParamRevisionController(true, cachecoord.New(cachecoord.NewStaticTopology(true), nil))

	if _, ok := svc.runtimeParamRevisionCtrl.(*clusterRuntimeParamRevisionController); !ok {
		t.Fatal("expected test setup to start with clustered runtime-param revision controller")
	}

	svc.OverrideClusterEnabledForDialect(false)

	if _, ok := svc.runtimeParamRevisionCtrl.(*localRuntimeParamRevisionController); !ok {
		t.Fatalf("expected dialect override to select local runtime-param revision controller, got %T", svc.runtimeParamRevisionCtrl)
	}
}

// TestPostgreSQLDialectSupportsClusterKeepsConfigServiceClusterEnabled verifies
// PostgreSQL does not trigger startup cluster compatibility overrides.
func TestPostgreSQLDialectSupportsClusterKeepsConfigServiceClusterEnabled(t *testing.T) {
	setTestConfigContent(t, `
cluster:
  enabled: true
  coordination:
    backend: redis
    group: default
redis:
  default:
    address: "127.0.0.1:6379"
`)

	svc := New(nil)

	dbDialect, err := dialect.From("pgsql:postgres:postgres@tcp(127.0.0.1:5432)/linapro?sslmode=disable")
	if err != nil {
		t.Fatalf("resolve PostgreSQL dialect failed: %v", err)
	}
	if !dbDialect.SupportsCluster() {
		svc.OverrideClusterEnabledForDialect(false)
	}

	if !svc.IsClusterEnabled(context.Background()) {
		t.Fatal("expected PostgreSQL startup hook to preserve enabled cluster mode")
	}
}

// TestSingleNodeModeDoesNotRequireRedis verifies local deployments can omit
// Redis groups.
func TestSingleNodeModeDoesNotRequireRedis(t *testing.T) {
	setTestConfigContent(t, `
cluster:
  enabled: false
`)

	svc := New(nil)
	cfg := svc.GetCluster(context.Background())
	if cfg.Enabled {
		t.Fatal("expected cluster mode disabled")
	}
	if len(svc.GetRedis(context.Background())) != 0 {
		t.Fatalf("expected no redis groups in single-node config, got %#v", svc.GetRedis(context.Background()))
	}
}

// setTestConfigContent swaps the config adapter content for one test case and
// restores it afterward.
func setTestConfigContent(t *testing.T, content string) {
	t.Helper()

	adapter, ok := g.Cfg().GetAdapter().(*gcfg.AdapterFile)
	if !ok {
		t.Fatal("expected config adapter to be *gcfg.AdapterFile")
	}

	originalContent := adapter.GetContent()
	adapter.SetContent(content)
	adapter.Clear()
	resetStaticConfigCaches()

	t.Cleanup(func() {
		if originalContent != "" {
			adapter.SetContent(originalContent)
		} else {
			adapter.RemoveContent()
		}
		adapter.Clear()
		resetStaticConfigCaches()
	})
}
