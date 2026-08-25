// This file loads top-level named Redis connection groups so cluster topology
// and other consumers can select a group without owning connection settings.

package config

import (
	"context"
	"strings"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/internal/service/cluster"
	"lina-core/internal/service/coordination"
)

// RedisConfig maps group names to Redis connection settings.
type RedisConfig map[string]RedisGroupConfig

// RedisGroupConfig holds one named Redis connection group.
type RedisGroupConfig struct {
	Address        string        `json:"address"`        // Address is host:port or comma-separated Redis Cluster nodes.
	DB             int           `json:"db"`             // DB selects the Redis logical database for standalone instances.
	Password       string        `json:"password"`       // Password authenticates to Redis when configured.
	ConnectTimeout time.Duration `json:"connectTimeout"` // ConnectTimeout bounds Redis connection establishment.
	ReadTimeout    time.Duration `json:"readTimeout"`    // ReadTimeout bounds Redis read operations.
	WriteTimeout   time.Duration `json:"writeTimeout"`   // WriteTimeout bounds Redis write operations.
}

const (
	defaultRedisConnectTimeout = 3 * time.Second
	defaultRedisReadTimeout    = 2 * time.Second
	defaultRedisWriteTimeout   = 2 * time.Second
)

// Options converts one Redis group into coordination Redis client options.
func (c RedisGroupConfig) Options(keys *coordination.KeyBuilder) coordination.RedisOptions {
	return coordination.RedisOptions{
		Address:        c.Address,
		DB:             c.DB,
		Password:       c.Password,
		ConnectTimeout: c.ConnectTimeout,
		ReadTimeout:    c.ReadTimeout,
		WriteTimeout:   c.WriteTimeout,
		KeyBuilder:     keys,
	}
}

// getStaticRedisConfig lazily loads named Redis groups from config.yaml.
func (s *serviceImpl) getStaticRedisConfig(ctx context.Context) RedisConfig {
	return *processStaticConfigCaches.redis.load(func() *RedisConfig {
		groups := RedisConfig{}
		mustScanConfig(ctx, "redis", &groups)
		normalized := make(RedisConfig, len(groups))
		for name, group := range groups {
			groupName := strings.TrimSpace(name)
			if groupName == "" {
				continue
			}
			group.ConnectTimeout = mustLoadDurationConfig(
				ctx,
				"redis."+groupName+".connectTimeout",
				defaultIfZero(group.ConnectTimeout, defaultRedisConnectTimeout),
			)
			group.ReadTimeout = mustLoadDurationConfig(
				ctx,
				"redis."+groupName+".readTimeout",
				defaultIfZero(group.ReadTimeout, defaultRedisReadTimeout),
			)
			group.WriteTimeout = mustLoadDurationConfig(
				ctx,
				"redis."+groupName+".writeTimeout",
				defaultIfZero(group.WriteTimeout, defaultRedisWriteTimeout),
			)
			normalized[groupName] = group
		}
		return &normalized
	})
}

// GetRedis reads named Redis connection groups from configuration file.
func (s *serviceImpl) GetRedis(ctx context.Context) RedisConfig {
	return cloneRedisConfig(s.getStaticRedisConfig(ctx))
}

// mustValidateClusterCoordination validates backend selection and the selected
// Redis group when clustered deployment is enabled.
func mustValidateClusterCoordination(clusterCfg *cluster.ClusterConfig, redisCfg RedisConfig) {
	if clusterCfg == nil || !clusterCfg.Enabled {
		return
	}
	backend := clusterCfg.Coordination.Backend
	if backend == "" {
		panic(clusterRedisDiagnosticError(
			"cluster.coordination.backend",
			"required when cluster.enabled=true",
			"set cluster.coordination.backend=redis",
		))
	}
	if backend != cluster.CoordinationBackendRedis {
		panic(clusterRedisDiagnosticError(
			"cluster.coordination.backend",
			"unsupported value "+string(backend),
			"set cluster.coordination.backend=redis",
		))
	}
	group := strings.TrimSpace(clusterCfg.Coordination.Group)
	if group == "" {
		group = cluster.DefaultCoordinationGroup
	}
	redisGroup, ok := redisCfg[group]
	if !ok {
		panic(clusterRedisDiagnosticError(
			"redis."+group,
			"required when cluster.coordination.group="+group,
			"define redis."+group+" with a non-empty address",
		))
	}
	if strings.TrimSpace(redisGroup.Address) == "" {
		panic(clusterRedisDiagnosticError(
			"redis."+group+".address",
			"required when cluster.coordination.backend=redis",
			"set redis."+group+".address to one host:port or comma-separated cluster nodes",
		))
	}
}

// clusterRedisDiagnosticError formats static Redis grouping failures.
func clusterRedisDiagnosticError(field string, reason string, fix string) error {
	return gerror.Newf("cluster startup diagnostic field=%s reason=%s fix=%s", field, reason, fix)
}

// defaultIfZero returns fallback when duration is unset.
func defaultIfZero(value time.Duration, fallback time.Duration) time.Duration {
	if value > 0 {
		return value
	}
	return fallback
}
