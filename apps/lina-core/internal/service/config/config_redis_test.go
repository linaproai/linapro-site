// This file verifies top-level named Redis group loading and cluster selection.

package config

import (
	"context"
	"testing"
	"time"

	"lina-core/internal/service/cluster"
)

// TestGetRedisLoadsNamedGroups verifies Redis groups are independent of cluster
// connection nesting and extra groups can coexist.
func TestGetRedisLoadsNamedGroups(t *testing.T) {
	setTestConfigContent(t, `
cluster:
  enabled: true
  coordination:
    backend: redis
    group: default
redis:
  default:
    address: "127.0.0.1:6379"
  cache:
    address: "127.0.0.1:6379"
    db: 1
`)

	svc := New(nil)
	clusterCfg := svc.GetCluster(context.Background())
	if clusterCfg.Coordination.Group != cluster.DefaultCoordinationGroup {
		t.Fatalf("expected default group, got %q", clusterCfg.Coordination.Group)
	}

	groups := svc.GetRedis(context.Background())
	if groups["default"].Address != "127.0.0.1:6379" {
		t.Fatalf("expected default redis address, got %#v", groups["default"])
	}
	if groups["cache"].DB != 1 {
		t.Fatalf("expected cache redis db 1, got %#v", groups["cache"])
	}
	if groups["default"].ConnectTimeout != 3*time.Second ||
		groups["default"].ReadTimeout != 2*time.Second ||
		groups["default"].WriteTimeout != 2*time.Second {
		t.Fatalf("expected default redis timeouts, got %#v", groups["default"])
	}
}

// TestGetRedisKeepsCommaSeparatedClusterAddress verifies Redis Cluster
// endpoints stay in the group address for client construction.
func TestGetRedisKeepsCommaSeparatedClusterAddress(t *testing.T) {
	setTestConfigContent(t, `
cluster:
  enabled: true
  coordination:
    backend: redis
    group: default
redis:
  default:
    address: "127.0.0.1:6379,127.0.0.1:6370"
`)

	group := New(nil).GetRedis(context.Background())["default"]
	if group.Address != "127.0.0.1:6379,127.0.0.1:6370" {
		t.Fatalf("expected comma-separated cluster address, got %q", group.Address)
	}
}

// TestGetClusterPanicsWhenCoordinationMissing verifies clustered deployment
// requires an explicit coordination backend.
func TestGetClusterPanicsWhenCoordinationMissing(t *testing.T) {
	setTestConfigContent(t, `
cluster:
  enabled: true
redis:
  default:
    address: "127.0.0.1:6379"
`)

	defer assertConfigPanicContains(t, "field=cluster.coordination.backend")
	New(nil).GetCluster(context.Background())
}

// TestGetClusterPanicsWhenCoordinationUnsupported verifies only Redis is
// accepted as the current coordination backend.
func TestGetClusterPanicsWhenCoordinationUnsupported(t *testing.T) {
	setTestConfigContent(t, `
cluster:
  enabled: true
  coordination:
    backend: postgres
    group: default
redis:
  default:
    address: "127.0.0.1:6379"
`)

	defer assertConfigPanicContains(t, "fix=set cluster.coordination.backend=redis")
	New(nil).GetCluster(context.Background())
}

// TestGetClusterPanicsWhenRedisGroupMissing verifies the selected Redis group
// must exist.
func TestGetClusterPanicsWhenRedisGroupMissing(t *testing.T) {
	setTestConfigContent(t, `
cluster:
  enabled: true
  coordination:
    backend: redis
    group: cache
redis:
  default:
    address: "127.0.0.1:6379"
`)

	defer assertConfigPanicContains(t, "field=redis.cache")
	New(nil).GetCluster(context.Background())
}

// TestGetClusterPanicsWhenRedisAddressMissing verifies Redis address is
// required for the selected group.
func TestGetClusterPanicsWhenRedisAddressMissing(t *testing.T) {
	setTestConfigContent(t, `
cluster:
  enabled: true
  coordination:
    backend: redis
    group: default
redis:
  default:
    db: 0
`)

	defer assertConfigPanicContains(t, "field=redis.default.address")
	New(nil).GetCluster(context.Background())
}

// TestGetClusterPanicsWhenRedisTimeoutInvalid verifies Redis timeout fields
// must be duration strings with units.
func TestGetClusterPanicsWhenRedisTimeoutInvalid(t *testing.T) {
	setTestConfigContent(t, `
cluster:
  enabled: true
  coordination:
    backend: redis
    group: default
redis:
  default:
    address: "127.0.0.1:6379"
    connectTimeout: invalid
`)

	defer assertConfigPanicContains(t, "parse config redis.default.connectTimeout failed")
	New(nil).GetCluster(context.Background())
}
