// Package redis implements the Redis-backed coordination provider used when
// cluster.enabled=true and cluster.coordination.backend=redis.
package redis

import (
	"context"
	"sync"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/redis/go-redis/v9"

	"lina-core/internal/service/coordination/internal/core"
	"lina-core/pkg/logger"
)

// Options contains normalized Redis connection settings for coordination.
type Options struct {
	Address        string           // Address is host:port or comma-separated Redis Cluster nodes.
	DB             int              // DB selects the Redis logical database for standalone instances.
	Password       string           // Password authenticates to Redis when configured.
	ConnectTimeout time.Duration    // ConnectTimeout bounds Redis connection establishment.
	ReadTimeout    time.Duration    // ReadTimeout bounds Redis read operations.
	WriteTimeout   time.Duration    // WriteTimeout bounds Redis write operations.
	KeyBuilder     *core.KeyBuilder // KeyBuilder scopes all Redis keys and channels.
}

// redisBackend implements all coordination stores through one Redis client.
type redisBackend struct {
	client  redis.UniversalClient
	keys    *core.KeyBuilder
	health  *redisHealth
	closeMu sync.Mutex
	closed  bool
}

// redisHealth stores observable Redis coordination health state.
type redisHealth struct {
	mu                  sync.RWMutex
	lastSuccessAt       time.Time
	lastError           string
	subscriberRunning   bool
	lastEventReceivedAt time.Time
}

// redisSubscription represents one active Redis pub/sub consumer.
type redisSubscription struct {
	pubsub *redis.PubSub
	cancel context.CancelFunc
	done   chan struct{}
	once   sync.Once
}

// Redis backend implementations are asserted at compile time.
var (
	_ core.LockStore     = (*redisBackend)(nil)
	_ core.KVStore       = (*redisBackend)(nil)
	_ core.RevisionStore = (*redisBackend)(nil)
	_ core.EventBus      = (*redisBackend)(nil)
	_ core.HealthChecker = (*redisBackend)(nil)
	_ core.Subscription  = (*redisSubscription)(nil)
)

// New creates a Redis coordination service and verifies connectivity.
func New(ctx context.Context, options Options) (core.Service, error) {
	keys := options.KeyBuilder
	if keys == nil {
		keys = core.DefaultKeyBuilder()
	}
	if options.ConnectTimeout <= 0 {
		options.ConnectTimeout = 3 * time.Second
	}
	if options.ReadTimeout <= 0 {
		options.ReadTimeout = 2 * time.Second
	}
	if options.WriteTimeout <= 0 {
		options.WriteTimeout = 2 * time.Second
	}
	addrs := splitRedisAddrs(options.Address)
	if len(addrs) == 0 {
		return nil, gerror.New("redis address is required")
	}
	client := redis.NewUniversalClient(&redis.UniversalOptions{
		Addrs:        addrs,
		DB:           options.DB,
		Password:     options.Password,
		DialTimeout:  options.ConnectTimeout,
		ReadTimeout:  options.ReadTimeout,
		WriteTimeout: options.WriteTimeout,
	})
	backend := &redisBackend{
		client: client,
		keys:   keys,
		health: &redisHealth{},
	}
	if err := backend.Ping(ctx); err != nil {
		closeErr := client.Close()
		if closeErr != nil {
			logger.Warningf(ctx, "[coordination] close redis client after ping failure: %v", closeErr)
		}
		return nil, err
	}
	return core.NewService(
		core.BackendRedis,
		keys,
		backend,
		backend,
		backend,
		backend,
		backend,
		backend.Close,
	), nil
}
