// This file verifies backend-neutral coordination behavior using the in-memory
// provider.

package coordination

import (
	"context"
	"errors"
	"github.com/redis/go-redis/v9"
	"lina-core/pkg/bizerr"
	"net"
	"os"
	"strconv"
	"testing"
	"time"
)

// TestKeyBuilderJoinsPlainSegments verifies coordination keys stay namespaced
// with readable segments and reject missing required segments.
func TestKeyBuilderJoinsPlainSegments(t *testing.T) {
	keys := NewKeyBuilder("app one", "dev", "node-a")

	lockKey, err := keys.LockKey("leader election")
	if err != nil {
		t.Fatalf("build lock key: %v", err)
	}
	if lockKey != "app one:dev:node-a:lock:leader election" {
		t.Fatalf("expected plain key segments, got %q", lockKey)
	}
	if _, err = keys.LockKey(" "); !bizerr.Is(err, CodeCoordinationKeyInvalid) {
		t.Fatalf("expected invalid key error, got %v", err)
	}
}

// TestKVKeyJoinsPlainSegments verifies KV keys are built by direct string
// concatenation without base64 encoding.
func TestKVKeyJoinsPlainSegments(t *testing.T) {
	keys := NewKeyBuilder("", "", "")

	key, err := keys.KVKey(0, "plugin", "media", "route-memory", "route_data:device-1:channel-2")
	if err != nil {
		t.Fatalf("build coordination KV key: %v", err)
	}
	expected := "linapro:default:default:kv:0:plugin:media:route-memory:route_data:device-1:channel-2"
	if key != expected {
		t.Fatalf("expected plain route-memory key %q, got %q", expected, key)
	}

	rawKey, err := keys.RawKVKey("session", "token", "jwt:id")
	if err != nil {
		t.Fatalf("build raw coordination KV key: %v", err)
	}
	if rawKey != "linapro:default:default:session:token:jwt:id" {
		t.Fatalf("expected plain raw key, got %q", rawKey)
	}

	revisionKey, err := keys.RevisionKey(RevisionKey{TenantID: 3, Domain: "plugin:media", Scope: "route"})
	if err != nil {
		t.Fatalf("build revision key: %v", err)
	}
	if revisionKey != "linapro:default:default:rev:3:plugin:media:route" {
		t.Fatalf("expected plain revision key, got %q", revisionKey)
	}
}

// TestMemoryLockStoreVerifiesOwnerToken covers acquire, renew, ownership, and
// release semantics for the LockStore contract.
func TestMemoryLockStoreVerifiesOwnerToken(t *testing.T) {
	ctx := context.Background()
	lockStore := NewMemory(nil).Lock()

	handle, ok, err := lockStore.Acquire(ctx, "unit-lock", "node-a", "unit", time.Second)
	if err != nil {
		t.Fatalf("acquire lock: %v", err)
	}
	if !ok || handle == nil {
		t.Fatal("expected first acquire to succeed")
	}

	competing, ok, err := lockStore.Acquire(ctx, "unit-lock", "node-b", "unit", time.Second)
	if err != nil {
		t.Fatalf("acquire competing lock: %v", err)
	}
	if ok || competing != nil {
		t.Fatal("expected competing lock acquire to miss")
	}

	if err = lockStore.Renew(ctx, handle, time.Second); err != nil {
		t.Fatalf("renew lock: %v", err)
	}
	held, err := lockStore.IsHeld(ctx, handle)
	if err != nil {
		t.Fatalf("check held: %v", err)
	}
	if !held {
		t.Fatal("expected renewed lock to be held")
	}

	stale := *handle
	stale.Token = "other-token"
	if err = lockStore.Release(ctx, &stale); !bizerr.Is(err, CodeCoordinationLockNotHeld) {
		t.Fatalf("expected stale release to fail with lock-not-held, got %v", err)
	}
	if err = lockStore.Release(ctx, handle); err != nil {
		t.Fatalf("release lock: %v", err)
	}
}

// TestMemoryKVStoreHonorsTTLAndCompareDelete verifies TTL visibility and
// compare-delete semantics.
func TestMemoryKVStoreHonorsTTLAndCompareDelete(t *testing.T) {
	ctx := context.Background()
	kv := NewMemory(nil).KV()

	if err := kv.Set(ctx, "unit-key", "value", 20*time.Millisecond); err != nil {
		t.Fatalf("set kv: %v", err)
	}
	if value, ok, err := kv.Get(ctx, "unit-key"); err != nil || !ok || value != "value" {
		t.Fatalf("expected visible value, value=%q ok=%t err=%v", value, ok, err)
	}
	time.Sleep(40 * time.Millisecond)
	if _, ok, err := kv.Get(ctx, "unit-key"); err != nil || ok {
		t.Fatalf("expected expired value to be hidden, ok=%t err=%v", ok, err)
	}

	if ok, err := kv.SetNX(ctx, "unit-key", "fresh", time.Second); err != nil || !ok {
		t.Fatalf("expected setnx after expiry, ok=%t err=%v", ok, err)
	}
	if ok, err := kv.CompareAndDelete(ctx, "unit-key", "wrong"); err != nil || ok {
		t.Fatalf("expected compare delete mismatch, ok=%t err=%v", ok, err)
	}
	if ok, err := kv.CompareAndDelete(ctx, "unit-key", "fresh"); err != nil || !ok {
		t.Fatalf("expected compare delete match, ok=%t err=%v", ok, err)
	}
}

// TestMemoryRevisionStoreIsSharedByKey verifies revision increments are scoped
// by tenant, domain, and scope.
func TestMemoryRevisionStoreIsSharedByKey(t *testing.T) {
	var (
		ctx       = context.Background()
		revisions = NewMemory(nil).Revision()
		key       = RevisionKey{TenantID: 1, Domain: "role", Scope: "global"}
	)

	first, err := revisions.Bump(ctx, key, "first")
	if err != nil {
		t.Fatalf("bump first revision: %v", err)
	}
	second, err := revisions.Bump(ctx, key, "second")
	if err != nil {
		t.Fatalf("bump second revision: %v", err)
	}
	current, err := revisions.Current(ctx, key)
	if err != nil {
		t.Fatalf("read current revision: %v", err)
	}
	if first != 1 || second != 2 || current != 2 {
		t.Fatalf("unexpected revisions first=%d second=%d current=%d", first, second, current)
	}
}

// TestMemoryEventBusPublishesToSubscribers verifies published events are
// delivered and subscription close stops delivery.
func TestMemoryEventBusPublishesToSubscribers(t *testing.T) {
	var (
		ctx      = context.Background()
		events   = NewMemory(nil).Events()
		received = make(chan Event, 1)
	)

	subscription, err := events.Subscribe(ctx, func(_ context.Context, event Event) error {
		received <- event
		return nil
	})
	if err != nil {
		t.Fatalf("subscribe: %v", err)
	}
	if err = events.Publish(ctx, Event{ID: "event-1", Kind: "cache.invalidate"}); err != nil {
		t.Fatalf("publish: %v", err)
	}
	select {
	case event := <-received:
		if event.ID != "event-1" {
			t.Fatalf("expected event-1, got %#v", event)
		}
	case <-time.After(time.Second):
		t.Fatal("expected event delivery")
	}
	if err = subscription.Close(ctx); err != nil {
		t.Fatalf("close subscription: %v", err)
	}
}

// TestRedisProviderReportsUnavailableWhenEndpointCloses verifies startup
// probing returns a structured Redis-unavailable error for broken endpoints.
func TestRedisProviderReportsUnavailableWhenEndpointCloses(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen on local tcp endpoint: %v", err)
	}
	t.Cleanup(func() {
		if err := listener.Close(); err != nil {
			t.Fatalf("close local tcp listener: %v", err)
		}
	})

	accepted := make(chan struct{})
	go func() {
		defer close(accepted)
		conn, acceptErr := listener.Accept()
		if acceptErr != nil {
			return
		}
		if closeErr := conn.Close(); closeErr != nil {
			t.Errorf("close accepted tcp connection: %v", closeErr)
		}
	}()

	_, err = NewRedis(context.Background(), RedisOptions{
		Address:        listener.Addr().String(),
		ConnectTimeout: 100 * time.Millisecond,
		ReadTimeout:    100 * time.Millisecond,
		WriteTimeout:   100 * time.Millisecond,
		KeyBuilder:     NewKeyBuilder("linapro-test", "redis-failure", "unit"),
	})
	if !bizerr.Is(err, CodeCoordinationRedisUnavailable) {
		t.Fatalf("expected Redis unavailable error, got %v", err)
	}

	select {
	case <-accepted:
	case <-time.After(time.Second):
		t.Fatal("expected local endpoint to accept Redis probe")
	}
}

// TestRedisProviderReportsUnavailableWhenEndpointTimesOut verifies startup
// probing respects configured timeouts for endpoints that accept but do not
// speak the Redis protocol.
func TestRedisProviderReportsUnavailableWhenEndpointTimesOut(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen on local tcp endpoint: %v", err)
	}
	t.Cleanup(func() {
		if err := listener.Close(); err != nil {
			t.Fatalf("close local tcp listener: %v", err)
		}
	})

	accepted := make(chan struct{})
	release := make(chan struct{})
	go func() {
		conn, acceptErr := listener.Accept()
		if acceptErr != nil {
			close(accepted)
			return
		}
		close(accepted)
		<-release
		if closeErr := conn.Close(); closeErr != nil {
			t.Errorf("close accepted tcp connection: %v", closeErr)
		}
	}()
	t.Cleanup(func() {
		close(release)
	})

	_, err = NewRedis(context.Background(), RedisOptions{
		Address:        listener.Addr().String(),
		ConnectTimeout: 100 * time.Millisecond,
		ReadTimeout:    20 * time.Millisecond,
		WriteTimeout:   20 * time.Millisecond,
		KeyBuilder:     NewKeyBuilder("linapro-test", "redis-timeout", "unit"),
	})
	if !bizerr.Is(err, CodeCoordinationRedisUnavailable) {
		t.Fatalf("expected Redis unavailable timeout error, got %v", err)
	}

	select {
	case <-accepted:
	case <-time.After(time.Second):
		t.Fatal("expected local endpoint to accept Redis probe")
	}
}

// TestRedisProviderIntegration verifies Redis lock, KV, revision, event, and
// health behavior through the public coordination service contract.
func TestRedisProviderIntegration(t *testing.T) {
	ctx := context.Background()
	service, keys := newRedisIntegrationService(t)

	lockName := "redis-integration-lock"
	handle, ok, err := service.Lock().Acquire(ctx, lockName, "node-a", "integration", time.Second)
	if err != nil {
		t.Fatalf("acquire redis lock: %v", err)
	}
	if !ok || handle == nil {
		t.Fatal("expected redis lock acquisition to succeed")
	}
	t.Cleanup(func() {
		cleanupRedisIntegrationLock(t, keys, lockName)
	})

	competing, ok, err := service.Lock().Acquire(ctx, lockName, "node-b", "integration", time.Second)
	if err != nil {
		t.Fatalf("acquire competing redis lock: %v", err)
	}
	if ok || competing != nil {
		t.Fatal("expected competing redis lock acquisition to miss")
	}
	if err = service.Lock().Renew(ctx, handle, time.Second); err != nil {
		t.Fatalf("renew redis lock: %v", err)
	}
	stale := *handle
	stale.Token = "other-token"
	if err = service.Lock().Release(ctx, &stale); !bizerr.Is(err, CodeCoordinationLockNotHeld) {
		t.Fatalf("expected non-owner release to fail, got %v", err)
	}
	if err = service.Lock().Release(ctx, handle); err != nil {
		t.Fatalf("release redis lock: %v", err)
	}

	kvKey, err := keys.RawKVKey("test-kv", "ttl")
	if err != nil {
		t.Fatalf("build redis kv key: %v", err)
	}
	t.Cleanup(func() {
		cleanupRedisIntegrationKeys(t, kvKey)
	})
	if err = service.KV().Set(ctx, kvKey, "value", 30*time.Millisecond); err != nil {
		t.Fatalf("set redis kv: %v", err)
	}
	if value, found, readErr := service.KV().Get(ctx, kvKey); readErr != nil || !found || value != "value" {
		t.Fatalf("expected redis kv value, value=%q found=%t err=%v", value, found, readErr)
	}
	time.Sleep(60 * time.Millisecond)
	if _, found, readErr := service.KV().Get(ctx, kvKey); readErr != nil || found {
		t.Fatalf("expected expired redis kv miss, found=%t err=%v", found, readErr)
	}

	revisionKey := RevisionKey{TenantID: 7, Domain: "redis-integration", Scope: "global"}
	revisionRedisKey, err := keys.RevisionKey(revisionKey)
	if err != nil {
		t.Fatalf("build redis revision key: %v", err)
	}
	t.Cleanup(func() {
		cleanupRedisIntegrationKeys(t, revisionRedisKey)
	})
	first, err := service.Revision().Bump(ctx, revisionKey, "first")
	if err != nil {
		t.Fatalf("bump first redis revision: %v", err)
	}
	second, err := service.Revision().Bump(ctx, revisionKey, "second")
	if err != nil {
		t.Fatalf("bump second redis revision: %v", err)
	}
	current, err := service.Revision().Current(ctx, revisionKey)
	if err != nil {
		t.Fatalf("read redis revision: %v", err)
	}
	if first != 1 || second != 2 || current != 2 {
		t.Fatalf("unexpected redis revisions first=%d second=%d current=%d", first, second, current)
	}

	received := make(chan Event, 1)
	subscription, err := service.Events().Subscribe(ctx, func(_ context.Context, event Event) error {
		received <- event
		return nil
	})
	if err != nil {
		t.Fatalf("subscribe redis events: %v", err)
	}
	t.Cleanup(func() {
		if closeErr := subscription.Close(ctx); closeErr != nil {
			t.Fatalf("close redis subscription: %v", closeErr)
		}
	})
	if err = service.Events().Publish(ctx, Event{ID: "redis-integration-event", Kind: "cache.invalidate"}); err != nil {
		t.Fatalf("publish redis event: %v", err)
	}
	select {
	case event := <-received:
		if event.ID != "redis-integration-event" {
			t.Fatalf("expected redis integration event, got %#v", event)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("expected redis event delivery")
	}

	snapshot := service.Health().Snapshot(ctx)
	if snapshot.Backend != BackendRedis || !snapshot.Healthy {
		t.Fatalf("expected healthy redis snapshot, got %#v", snapshot)
	}
}

// TestRedisProviderFailureIntegration verifies common Redis failure semantics
// through a closed client without requiring a second broken Redis endpoint.
func TestRedisProviderFailureIntegration(t *testing.T) {
	ctx := context.Background()
	service, _ := newRedisIntegrationService(t)

	if err := service.Close(ctx); err != nil {
		t.Fatalf("close redis coordination service: %v", err)
	}
	if _, _, err := service.Lock().Acquire(ctx, "closed-lock", "node-a", "failure", time.Second); !bizerr.Is(err, CodeCoordinationRedisUnavailable) {
		t.Fatalf("expected closed redis lock acquire to report unavailable, got %v", err)
	}
	if err := service.KV().Set(ctx, "closed-kv", "value", 0); !bizerr.Is(err, CodeCoordinationKVOperationFailed) {
		t.Fatalf("expected closed redis kv set to report operation failure, got %v", err)
	}
	if _, err := service.Revision().Current(ctx, RevisionKey{TenantID: 1, Domain: "closed", Scope: "global"}); !bizerr.Is(err, CodeCoordinationRevisionUnavailable) {
		t.Fatalf("expected closed redis revision read to report unavailable, got %v", err)
	}
	if err := service.Events().Publish(ctx, Event{ID: "closed-event", Kind: "cache.invalidate"}); !bizerr.Is(err, CodeCoordinationEventPublishFailed) {
		t.Fatalf("expected closed redis event publish to report failure, got %v", err)
	}

	canceledCtx, cancel := context.WithCancel(ctx)
	cancel()
	if _, _, err := service.KV().Get(canceledCtx, "closed-kv"); err == nil {
		t.Fatal("expected canceled context redis kv get to fail")
	}
}

// newRedisIntegrationService creates one Redis-backed coordination service for
// integration tests and skips unless LINA_TEST_REDIS_ADDR is set.
func newRedisIntegrationService(t *testing.T) (Service, *KeyBuilder) {
	t.Helper()

	address := os.Getenv("LINA_TEST_REDIS_ADDR")
	if address == "" {
		t.Skip("set LINA_TEST_REDIS_ADDR to enable Redis coordination integration tests")
	}
	db := 0
	if rawDB := os.Getenv("LINA_TEST_REDIS_DB"); rawDB != "" {
		parsedDB, err := strconv.Atoi(rawDB)
		if err != nil {
			t.Fatalf("parse LINA_TEST_REDIS_DB: %v", err)
		}
		db = parsedDB
	}

	ctx := context.Background()
	keys := NewKeyBuilder("linapro-test", "redis-integration", strconv.FormatInt(time.Now().UnixNano(), 10))
	service, err := NewRedis(ctx, RedisOptions{
		Address:        address,
		DB:             db,
		Password:       os.Getenv("LINA_TEST_REDIS_PASSWORD"),
		ConnectTimeout: time.Second,
		ReadTimeout:    time.Second,
		WriteTimeout:   time.Second,
		KeyBuilder:     keys,
	})
	if err != nil {
		t.Fatalf("create redis coordination service: %v", err)
	}
	t.Cleanup(func() {
		if closeErr := service.Close(ctx); closeErr != nil && !errors.Is(closeErr, redis.ErrClosed) {
			t.Fatalf("close redis coordination service: %v", closeErr)
		}
	})
	return service, keys
}

// cleanupRedisIntegrationLock removes exact Redis lock keys created by tests.
func cleanupRedisIntegrationLock(t *testing.T, keys *KeyBuilder, lockName string) {
	t.Helper()

	lockKey, err := keys.LockKey(lockName)
	if err != nil {
		t.Fatalf("build redis lock cleanup key: %v", err)
	}
	fenceKey, err := keys.LockFenceKey(lockName)
	if err != nil {
		t.Fatalf("build redis fence cleanup key: %v", err)
	}
	cleanupRedisIntegrationKeys(t, lockKey, fenceKey)
}

// cleanupRedisIntegrationKeys deletes exact Redis keys created by integration
// tests without scanning the database or using FLUSHDB.
func cleanupRedisIntegrationKeys(t *testing.T, keys ...string) {
	t.Helper()

	address := os.Getenv("LINA_TEST_REDIS_ADDR")
	if address == "" || len(keys) == 0 {
		return
	}
	db := 0
	if rawDB := os.Getenv("LINA_TEST_REDIS_DB"); rawDB != "" {
		parsedDB, err := strconv.Atoi(rawDB)
		if err != nil {
			t.Fatalf("parse LINA_TEST_REDIS_DB for cleanup: %v", err)
		}
		db = parsedDB
	}
	client := redis.NewClient(&redis.Options{
		Addr:     address,
		DB:       db,
		Password: os.Getenv("LINA_TEST_REDIS_PASSWORD"),
	})
	defer func() {
		if err := client.Close(); err != nil {
			t.Fatalf("close redis cleanup client: %v", err)
		}
	}()
	if err := client.Del(context.Background(), keys...).Err(); err != nil {
		t.Fatalf("cleanup redis integration keys: %v", err)
	}
}
