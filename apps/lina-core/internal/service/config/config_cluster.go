// This file defines cluster topology configuration loading and default election
// settings for single-node and multi-node deployments. Redis connections are
// loaded from top-level named groups and selected by cluster.coordination.

package config

import (
	"context"
	"strings"
	"time"

	"lina-core/internal/service/cluster"
)

// ClusterConfig aliases the cluster-owned topology configuration value.
type ClusterConfig = cluster.ClusterConfig

// ElectionConfig aliases the cluster-owned leader election configuration value.
type ElectionConfig = cluster.ElectionConfig

// defaultElectionConfig returns the host defaults for leader-election timing.
func defaultElectionConfig() *ElectionConfig {
	return &ElectionConfig{
		Lease:         30 * time.Second,
		RenewInterval: 10 * time.Second,
	}
}

// getStaticClusterConfig lazily loads the cluster deployment mode from
// config.yaml so callers can branch on single-node vs multi-node behavior
// without reparsing the config section on hot paths.
func (s *serviceImpl) getStaticClusterConfig(ctx context.Context) *ClusterConfig {
	return processStaticConfigCaches.cluster.load(func() *ClusterConfig {
		cfg := &ClusterConfig{
			Enabled:  false,
			Election: *defaultElectionConfig(),
		}
		mustScanConfig(ctx, "cluster", cfg)
		cfg.Election.Lease = mustLoadDurationConfig(ctx, "cluster.election.lease", cfg.Election.Lease)
		cfg.Election.RenewInterval = mustLoadDurationConfig(ctx, "cluster.election.renewInterval", cfg.Election.RenewInterval)
		cfg.Coordination.Backend = cluster.CoordinationBackend(strings.TrimSpace(string(cfg.Coordination.Backend)))
		if strings.TrimSpace(cfg.Coordination.Group) == "" {
			cfg.Coordination.Group = cluster.DefaultCoordinationGroup
		}
		if cfg.Enabled {
			mustValidateClusterCoordination(cfg, s.getStaticRedisConfig(ctx))
		}
		return cfg
	})
}

// GetCluster reads cluster topology config from configuration file.
func (s *serviceImpl) GetCluster(ctx context.Context) *ClusterConfig {
	cfg := cloneClusterConfig(s.getStaticClusterConfig(ctx))
	if s != nil && s.clusterOverride != nil {
		cfg.Enabled = *s.clusterOverride
	}
	return cfg
}

// IsClusterEnabled reports whether multi-node cluster mode is enabled.
func (s *serviceImpl) IsClusterEnabled(ctx context.Context) bool {
	if s != nil && s.clusterOverride != nil {
		return *s.clusterOverride
	}
	cfg := s.getStaticClusterConfig(ctx)
	return cfg != nil && cfg.Enabled
}

// OverrideClusterEnabledForDialect locks cluster.enabled in memory for dialects
// that cannot safely back multi-node coordination.
func (s *serviceImpl) OverrideClusterEnabledForDialect(value bool) {
	if s == nil {
		return
	}
	s.clusterOverride = &value
	s.runtimeParamRevisionCtrl = newCacheCoordRuntimeParamRevisionController(
		s.IsClusterEnabled(context.Background()),
		s.cacheCoordSvc,
	)
}
