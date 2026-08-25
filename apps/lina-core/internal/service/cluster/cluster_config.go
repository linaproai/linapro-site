// This file defines cluster topology configuration value objects owned by the
// cluster component. The config service aliases these types when loading
// config.yaml. Redis connection settings are not part of this value object;
// cluster.coordination only selects a backend name and a redis group.

package cluster

import (
	"strings"
	"time"
)

// DefaultCoordinationGroup is the redis group used when cluster.coordination.group
// is omitted.
const DefaultCoordinationGroup = "default"

// CoordinationBackend names the coordination implementation selected by cluster
// topology configuration.
type CoordinationBackend string

// Supported coordination backend names selected from cluster configuration.
const (
	// CoordinationBackendRedis selects the Redis coordination implementation.
	CoordinationBackendRedis CoordinationBackend = "redis"
)

// ClusterConfig holds cluster topology configuration.
type ClusterConfig struct {
	Enabled      bool                      `json:"enabled"`      // Enabled reports whether clustered deployment is enabled.
	Election     ElectionConfig            `json:"election"`     // Election contains primary-election settings for clustered mode.
	Coordination ClusterCoordinationConfig `json:"coordination"` // Coordination selects the injected backend and redis group.
}

// ClusterCoordinationConfig selects one coordination implementation and redis group.
type ClusterCoordinationConfig struct {
	Backend CoordinationBackend `json:"backend"` // Backend names the coordination implementation.
	Group   string              `json:"group"`   // Group selects a top-level redis connection group.
}

// ElectionConfig holds leader election configuration.
type ElectionConfig struct {
	Lease         time.Duration `json:"lease"`         // Lease is the lock lease duration.
	RenewInterval time.Duration `json:"renewInterval"` // RenewInterval is the lease renewal interval.
}

// normalizeClusterConfig applies default election settings and coordination
// group while preserving the caller-provided enablement flag.
func normalizeClusterConfig(cfg *ClusterConfig) *ClusterConfig {
	normalizedCfg := &ClusterConfig{
		Enabled: false,
		Election: ElectionConfig{
			Lease:         defaultElectionLease,
			RenewInterval: defaultElectionRenewInterval,
		},
		Coordination: ClusterCoordinationConfig{
			Group: DefaultCoordinationGroup,
		},
	}
	if cfg == nil {
		return normalizedCfg
	}

	normalizedCfg.Enabled = cfg.Enabled
	normalizedCfg.Coordination.Backend = CoordinationBackend(strings.TrimSpace(string(cfg.Coordination.Backend)))
	if group := strings.TrimSpace(cfg.Coordination.Group); group != "" {
		normalizedCfg.Coordination.Group = group
	}
	if cfg.Election.Lease > 0 {
		normalizedCfg.Election.Lease = cfg.Election.Lease
	}
	if cfg.Election.RenewInterval > 0 {
		normalizedCfg.Election.RenewInterval = cfg.Election.RenewInterval
	}
	return normalizedCfg
}
