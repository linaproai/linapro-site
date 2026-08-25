// This file verifies cache coordination revision behavior.

package cachecoord

import (
	"context"
	"github.com/redis/go-redis/v9"
	"lina-core/internal/service/coordination"
	"os"
	"strconv"
	"sync"
	"testing"
	"time"
)

const (
	// testRuntimeConfigDomain mirrors the host runtime-config domain in cachecoord tests.
	testRuntimeConfigDomain Domain = "runtime-config"
	// testPluginRuntimeDomain mirrors the host plugin-runtime domain in cachecoord tests.
	testPluginRuntimeDomain Domain = "plugin-runtime"
)

// defaultCoordinatorTestTopology is a non-static topology used to verify that
// real cluster topology wiring is not replaced by later static placeholders.
type defaultCoordinatorTestTopology struct {
	enabled bool
}

// Start records no behavior for the test topology.
func (defaultCoordinatorTestTopology) Start(context.Context) {}

// Stop records no behavior for the test topology.
func (defaultCoordinatorTestTopology) Stop(context.Context) {}

// IsEnabled reports the configured cluster mode for the test topology.
func (t defaultCoordinatorTestTopology) IsEnabled() bool {
	return t.enabled
}

// IsPrimary reports this test node as primary.
func (defaultCoordinatorTestTopology) IsPrimary() bool {
	return true
}

// NodeID returns a stable test node identifier.
func (defaultCoordinatorTestTopology) NodeID() string {
	return "default-test-node"
}

// TestSingleNodeMarkChangedUsesProcessLocalRevision verifies local mode uses
// the instance revision store rather than a shared SQL table.
func TestSingleNodeMarkChangedUsesProcessLocalRevision(t *testing.T) {
	ctx := context.Background()
	service := New(NewStaticTopology(false), nil)

	firstRevision, err := service.MarkChanged(
		ctx,
		testRuntimeConfigDomain,
		ScopeGlobal,
		ChangeReason("unit_test_local_first"),
	)
	if err != nil {
		t.Fatalf("first local mark failed: %v", err)
	}
	secondRevision, err := service.MarkChanged(
		ctx,
		testRuntimeConfigDomain,
		ScopeGlobal,
		ChangeReason("unit_test_local_second"),
	)
	if err != nil {
		t.Fatalf("second local mark failed: %v", err)
	}
	if secondRevision != firstRevision+1 {
		t.Fatalf("expected local revision to increment from %d to %d, got %d", firstRevision, firstRevision+1, secondRevision)
	}
}

// TestNewInstancesDoNotShareLocalRevisions verifies two constructors do not
// inherit each other's standalone revision counters.
func TestNewInstancesDoNotShareLocalRevisions(t *testing.T) {
	ctx := context.Background()
	first := New(NewStaticTopology(false), nil)
	second := New(NewStaticTopology(false), nil)

	bumped, err := first.MarkChanged(ctx, testRuntimeConfigDomain, ScopeGlobal, ChangeReason("unit_test_isolation"))
	if err != nil {
		t.Fatalf("first instance mark failed: %v", err)
	}
	other, err := second.CurrentRevision(ctx, testRuntimeConfigDomain, ScopeGlobal)
	if err != nil {
		t.Fatalf("second instance current revision failed: %v", err)
	}
	if other >= bumped && bumped > 1 {
		t.Fatalf("expected isolated instance revisions, first=%d second=%d", bumped, other)
	}
	if other != 1 {
		t.Fatalf("expected a fresh instance to start at revision 1, got %d", other)
	}
}

// TestTenantScopedMarkChangedIsolatesLocalRevisions verifies tenant invalidation scope uses separate revisions.
func TestTenantScopedMarkChangedIsolatesLocalRevisions(t *testing.T) {
	var (
		ctx     = context.Background()
		service = New(NewStaticTopology(false), nil)
		domain  = Domain("unit-tenant-cache")
		scope   = Scope("dict")
	)

	tenantOneRevision, err := service.MarkTenantChanged(
		ctx,
		domain,
		scope,
		InvalidationScope{TenantID: 1},
		ChangeReason("tenant_one"),
	)
	if err != nil {
		t.Fatalf("tenant one mark failed: %v", err)
	}
	tenantTwoRevision, err := service.MarkTenantChanged(
		ctx,
		domain,
		scope,
		InvalidationScope{TenantID: 2},
		ChangeReason("tenant_two"),
	)
	if err != nil {
		t.Fatalf("tenant two mark failed: %v", err)
	}
	if tenantOneRevision != 1 || tenantTwoRevision != 1 {
		t.Fatalf("expected isolated first revisions, got tenant1=%d tenant2=%d", tenantOneRevision, tenantTwoRevision)
	}
}

// TestTenantScopedMarkChangedCascadeUsesDistinctScope verifies platform
// cascade invalidation does not overwrite a tenant-only revision bucket.
func TestTenantScopedMarkChangedCascadeUsesDistinctScope(t *testing.T) {
	var (
		ctx     = context.Background()
		service = New(NewStaticTopology(false), nil)
		domain  = Domain("unit-tenant-cache-cascade")
		scope   = Scope("permission")
	)

	tenantRevision, err := service.MarkTenantChanged(
		ctx,
		domain,
		scope,
		InvalidationScope{TenantID: 9},
		ChangeReason("tenant_only"),
	)
	if err != nil {
		t.Fatalf("tenant mark failed: %v", err)
	}
	cascadeRevision, err := service.MarkTenantChanged(
		ctx,
		domain,
		scope,
		InvalidationScope{TenantID: 0, CascadeToTenants: true},
		ChangeReason("platform_cascade"),
	)
	if err != nil {
		t.Fatalf("platform cascade mark failed: %v", err)
	}
	if tenantRevision != 1 || cascadeRevision != 1 {
		t.Fatalf("expected isolated tenant/cascade first revisions, got tenant=%d cascade=%d", tenantRevision, cascadeRevision)
	}
	if scoped := ScopedScope(scope, InvalidationScope{TenantID: 9}); scoped == scope {
		t.Fatalf("expected tenant scope to include tenant discriminator, got %q", scoped)
	}
	if cascade := ScopedScope(scope, InvalidationScope{TenantID: 0, CascadeToTenants: true}); cascade == scope {
		t.Fatalf("expected cascade scope to include cascade discriminator, got %q", cascade)
	}
}

// TestNewInstancesDoNotShareProcessState verifies constructors bind topology
// at creation time and later New calls cannot rewrite an earlier instance.
func TestNewInstancesDoNotShareProcessState(t *testing.T) {
	first := New(NewStaticTopology(false), nil)
	second := New(defaultCoordinatorTestTopology{enabled: true}, nil)

	firstImpl, ok := first.(*serviceImpl)
	if !ok {
		t.Fatalf("expected cachecoord implementation, got %T", first)
	}
	secondImpl, ok := second.(*serviceImpl)
	if !ok {
		t.Fatalf("expected cachecoord implementation, got %T", second)
	}
	if first == second {
		t.Fatal("expected distinct cachecoord instances from separate New calls")
	}
	if firstImpl.clusterEnabled() {
		t.Fatal("expected the first instance to keep standalone topology")
	}
	if !secondImpl.clusterEnabled() {
		t.Fatal("expected the second instance to keep clustered topology")
	}
}

// TestClusterMarkChangedPersistsAtomicRevision verifies concurrent clustered
// publishers increment the same persistent row without losing revisions.
func TestClusterMarkChangedPersistsAtomicRevision(t *testing.T) {
	ctx := context.Background()
	service := New(NewStaticTopology(true), coordination.NewMemory(nil))

	const workers = 12
	revisions := make(chan int64, workers)
	errs := make(chan error, workers)
	var wg sync.WaitGroup
	for range workers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			revision, err := service.MarkChanged(
				ctx,
				testRuntimeConfigDomain,
				Scope("unit-test-atomic"),
				ChangeReason("unit_test_concurrent_publish"),
			)
			if err != nil {
				errs <- err
				return
			}
			revisions <- revision
		}()
	}
	wg.Wait()
	close(revisions)
	close(errs)

	for err := range errs {
		if err != nil {
			t.Fatalf("concurrent publish failed: %v", err)
		}
	}

	seen := make(map[int64]struct{}, workers)
	for revision := range revisions {
		seen[revision] = struct{}{}
	}
	if len(seen) != workers {
		t.Fatalf("expected %d unique revisions, got %d: %#v", workers, len(seen), seen)
	}

	latest, err := service.CurrentRevision(ctx, testRuntimeConfigDomain, Scope("unit-test-atomic"))
	if err != nil {
		t.Fatalf("read shared revision failed: %v", err)
	}
	if latest != workers {
		t.Fatalf("expected latest revision %d, got %d", workers, latest)
	}
}

// TestClusterMarkChangedAcceptsUnconfiguredDomain verifies callers can use a
// new valid domain without changing cachecoord code or configuring metadata.
func TestClusterMarkChangedAcceptsUnconfiguredDomain(t *testing.T) {
	var (
		ctx     = context.Background()
		service = New(NewStaticTopology(true), coordination.NewMemory(nil))
		domain  = Domain("plugin:unit-test:custom")
		scope   = Scope("unit-test-free-domain")
	)

	revision, err := service.MarkChanged(ctx, domain, scope, ChangeReason("free_domain"))
	if err != nil {
		t.Fatalf("publish unconfigured domain failed: %v", err)
	}
	if revision != 1 {
		t.Fatalf("expected first unconfigured domain revision 1, got %d", revision)
	}

	items, err := service.Snapshot(ctx)
	if err != nil {
		t.Fatalf("snapshot unconfigured domain failed: %v", err)
	}
	for _, item := range items {
		if item.Domain != domain || item.Scope != scope {
			continue
		}
		if item.AuthoritySource != "caller-owned cache domain" ||
			item.ConsistencyModel != ConsistencySharedRevision ||
			item.MaxStale != DefaultDomainMaxStale ||
			item.FailureStrategy != FailureStrategyReturnVisibleError {
			t.Fatalf("expected default domain spec, got %#v", item)
		}
		return
	}
	t.Fatalf("expected snapshot item for unconfigured domain %q/%q, got %#v", domain, scope, items)
}

// TestClusterCurrentRevisionHandlesMissingSharedRow verifies first reads in
// cluster mode treat a missing revision row as revision zero instead of an
// infrastructure failure.
func TestClusterCurrentRevisionHandlesMissingSharedRow(t *testing.T) {
	var (
		ctx     = context.Background()
		service = New(NewStaticTopology(true), coordination.NewMemory(nil))
		scope   = Scope("unit-test-missing-shared-row")
	)

	revision, err := service.CurrentRevision(ctx, testRuntimeConfigDomain, scope)
	if err != nil {
		t.Fatalf("expected missing shared revision row to read as zero, got error: %v", err)
	}
	if revision != 0 {
		t.Fatalf("expected missing shared revision row to return 0, got %d", revision)
	}
}

// TestEnsureFreshRefreshesOncePerRevision verifies the refresher only runs when
// the observed revision advances.
func TestEnsureFreshRefreshesOncePerRevision(t *testing.T) {
	var (
		ctx       = context.Background()
		coordSvc  = coordination.NewMemory(nil)
		publisher = New(NewStaticTopology(true), coordSvc)
		consumer  = New(NewStaticTopology(true), coordSvc)
	)

	if _, err := publisher.MarkChanged(ctx, testPluginRuntimeDomain, Scope("unit-test-refresh"), ChangeReason("first")); err != nil {
		t.Fatalf("publish first revision failed: %v", err)
	}

	refreshCalls := 0
	refresher := func(_ context.Context, revision int64) error {
		refreshCalls++
		if revision != 1 && revision != 2 {
			t.Fatalf("unexpected refresh revision %d", revision)
		}
		return nil
	}
	if _, err := consumer.EnsureFresh(ctx, testPluginRuntimeDomain, Scope("unit-test-refresh"), refresher); err != nil {
		t.Fatalf("first ensure fresh failed: %v", err)
	}
	if _, err := consumer.EnsureFresh(ctx, testPluginRuntimeDomain, Scope("unit-test-refresh"), refresher); err != nil {
		t.Fatalf("second ensure fresh failed: %v", err)
	}
	if refreshCalls != 1 {
		t.Fatalf("expected one refresh for first revision, got %d", refreshCalls)
	}

	if _, err := publisher.MarkChanged(ctx, testPluginRuntimeDomain, Scope("unit-test-refresh"), ChangeReason("second")); err != nil {
		t.Fatalf("publish second revision failed: %v", err)
	}
	if _, err := consumer.EnsureFresh(ctx, testPluginRuntimeDomain, Scope("unit-test-refresh"), refresher); err != nil {
		t.Fatalf("third ensure fresh failed: %v", err)
	}
	if refreshCalls != 2 {
		t.Fatalf("expected second refresh after revision change, got %d", refreshCalls)
	}
}

// TestSnapshotDoesNotLeakStatusAcrossInstances verifies diagnostics stay on
// the constructing cachecoord instance.
func TestSnapshotDoesNotLeakStatusAcrossInstances(t *testing.T) {
	var (
		ctx              = context.Background()
		scope            = Scope("unit-test-process-snapshot")
		coordSvc         = coordination.NewMemory(nil)
		publisher        = New(NewStaticTopology(true), coordSvc)
		diagnosticReader = New(NewStaticTopology(true), coordSvc)
	)

	revision, err := publisher.MarkChanged(ctx, testRuntimeConfigDomain, scope, ChangeReason("diagnostic_snapshot"))
	if err != nil {
		t.Fatalf("publish revision failed: %v", err)
	}
	items, err := publisher.Snapshot(ctx)
	if err != nil {
		t.Fatalf("publisher snapshot failed: %v", err)
	}
	found := false
	for _, item := range items {
		if item.Domain == testRuntimeConfigDomain && item.Scope == scope {
			found = true
			if item.SharedRevision != revision {
				t.Fatalf("expected publisher snapshot revision %d, got %#v", revision, item)
			}
		}
	}
	if !found {
		t.Fatalf("expected publisher snapshot item for scope %q, got %#v", scope, items)
	}

	otherItems, err := diagnosticReader.Snapshot(ctx)
	if err != nil {
		t.Fatalf("reader snapshot failed: %v", err)
	}
	for _, item := range otherItems {
		if item.Domain == testRuntimeConfigDomain && item.Scope == scope {
			t.Fatalf("expected isolated snapshot, reader leaked %#v", item)
		}
	}
}

// TestSnapshotIncludesCoordinationHealth verifies clustered cache diagnostics
// expose the active coordination backend and event subscription status.
func TestSnapshotIncludesCoordinationHealth(t *testing.T) {
	var (
		ctx      = context.Background()
		scope    = Scope("unit-test-coordination-health")
		coordSvc = coordination.NewMemory(nil)
		service  = New(NewStaticTopology(true), coordSvc)
	)

	subscription, err := coordSvc.Events().Subscribe(ctx, func(context.Context, coordination.Event) error {
		return nil
	})
	if err != nil {
		t.Fatalf("subscribe coordination events: %v", err)
	}
	t.Cleanup(func() {
		if err := subscription.Close(ctx); err != nil {
			t.Fatalf("close coordination subscription: %v", err)
		}
	})

	if _, err := service.MarkChanged(ctx, testRuntimeConfigDomain, scope, ChangeReason("diagnostic_health")); err != nil {
		t.Fatalf("publish revision failed: %v", err)
	}

	items, err := service.Snapshot(ctx)
	if err != nil {
		t.Fatalf("snapshot failed: %v", err)
	}
	for _, item := range items {
		if item.Domain != testRuntimeConfigDomain || item.Scope != scope {
			continue
		}
		if item.Backend != coordination.BackendMemory {
			t.Fatalf("expected memory backend in snapshot, got %#v", item)
		}
		if !item.CoordinationHealthy || !item.EventSubscriberRunning {
			t.Fatalf("expected healthy event subscriber diagnostics, got %#v", item)
		}
		return
	}
	t.Fatalf("expected snapshot item for scope %q, got %#v", scope, items)
}

// redisCacheCoordTestTopology gives Redis integration tests distinct node IDs.
type redisCacheCoordTestTopology struct {
	nodeID string
}

// Start records no behavior for Redis cachecoord integration test topology.
func (redisCacheCoordTestTopology) Start(context.Context) {}

// Stop records no behavior for Redis cachecoord integration test topology.
func (redisCacheCoordTestTopology) Stop(context.Context) {}

// IsEnabled reports clustered mode for Redis cachecoord integration tests.
func (t redisCacheCoordTestTopology) IsEnabled() bool {
	return true
}

// IsPrimary reports the test node as primary.
func (t redisCacheCoordTestTopology) IsPrimary() bool {
	return true
}

// NodeID returns the test node identifier.
func (t redisCacheCoordTestTopology) NodeID() string {
	if t.nodeID == "" {
		return "redis-cachecoord-test-node"
	}
	return t.nodeID
}

// TestRedisCacheCoordIntegrationConcurrentRevisionAndEvent verifies Redis
// backed cachecoord revisions are atomic across concurrent publishers and that
// another node receives the Redis pub/sub notification needed to refresh local
// state.
func TestRedisCacheCoordIntegrationConcurrentRevisionAndEvent(t *testing.T) {
	var (
		ctx         = context.Background()
		keys        = newRedisCacheCoordIntegrationKeyBuilder(t)
		writerCoord = newRedisCacheCoordIntegrationService(t, keys)
		readerCoord = newRedisCacheCoordIntegrationService(t, keys)
		publisher   = New(redisCacheCoordTestTopology{nodeID: "redis-cachecoord-writer"}, writerCoord)
		consumer    = New(redisCacheCoordTestTopology{nodeID: "redis-cachecoord-reader"}, readerCoord)
		domain      = Domain("redis-cachecoord")
		scope       = Scope("concurrent-event")
	)

	revisionRedisKey, err := keys.RevisionKey(coordination.RevisionKey{
		TenantID: 0,
		Domain:   string(domain),
		Scope:    string(scope),
	})
	if err != nil {
		t.Fatalf("build cachecoord redis revision key: %v", err)
	}
	t.Cleanup(func() {
		cleanupRedisCacheCoordIntegrationKeys(t, revisionRedisKey)
	})

	const workers = 8
	var (
		events      = make(chan coordination.Event, workers)
		refreshed   = make(chan int64, workers)
		handlerErrs = make(chan error, workers)
	)
	subscription, err := readerCoord.Events().Subscribe(ctx, func(handlerCtx context.Context, event coordination.Event) error {
		if event.Kind != cacheInvalidateEventKind ||
			event.Domain != string(domain) ||
			event.Scope != string(scope) {
			return nil
		}
		events <- event
		_, ensureErr := consumer.EnsureFresh(handlerCtx, domain, scope, func(_ context.Context, revision int64) error {
			refreshed <- revision
			return nil
		})
		if ensureErr != nil {
			handlerErrs <- ensureErr
			return ensureErr
		}
		return nil
	})
	if err != nil {
		t.Fatalf("subscribe cachecoord redis events: %v", err)
	}
	t.Cleanup(func() {
		if err := subscription.Close(ctx); err != nil {
			t.Fatalf("close cachecoord redis subscription: %v", err)
		}
	})

	revisions := publishRedisCacheCoordChanges(t, ctx, publisher, domain, scope, workers)
	assertRedisCacheCoordRevisions(t, revisions, workers)

	current, err := consumer.CurrentRevision(ctx, domain, scope)
	if err != nil {
		t.Fatalf("read cachecoord redis revision: %v", err)
	}
	if current != workers {
		t.Fatalf("expected redis cachecoord revision %d, got %d", workers, current)
	}

	event := waitForRedisCacheCoordEvent(t, events, handlerErrs, 3*time.Second)
	if event.SourceNode != "redis-cachecoord-writer" {
		t.Fatalf("expected event source node redis-cachecoord-writer, got %#v", event)
	}
	if revision := waitForRedisCacheCoordRefresh(t, refreshed, handlerErrs, workers, 3*time.Second); revision != workers {
		t.Fatalf("expected redis event refresh revision %d, got %d", workers, revision)
	}

	items, err := consumer.Snapshot(ctx)
	if err != nil {
		t.Fatalf("read cachecoord redis snapshot: %v", err)
	}
	assertRedisCacheCoordSnapshot(t, items, domain, scope, workers)
}

// newRedisCacheCoordIntegrationKeyBuilder creates a unique Redis key namespace
// and skips unless Redis integration tests are explicitly enabled.
func newRedisCacheCoordIntegrationKeyBuilder(t *testing.T) *coordination.KeyBuilder {
	t.Helper()

	if os.Getenv("LINA_TEST_REDIS_ADDR") == "" {
		t.Skip("set LINA_TEST_REDIS_ADDR to enable Redis cachecoord integration tests")
	}
	return coordination.NewKeyBuilder(
		"linapro-test",
		"cachecoord-redis",
		strconv.FormatInt(time.Now().UnixNano(), 10),
	)
}

// newRedisCacheCoordIntegrationService creates one Redis-backed coordination
// service sharing the provided test key namespace.
func newRedisCacheCoordIntegrationService(t *testing.T, keys *coordination.KeyBuilder) coordination.Service {
	t.Helper()

	db := 0
	if rawDB := os.Getenv("LINA_TEST_REDIS_DB"); rawDB != "" {
		parsedDB, err := strconv.Atoi(rawDB)
		if err != nil {
			t.Fatalf("parse LINA_TEST_REDIS_DB: %v", err)
		}
		db = parsedDB
	}

	ctx := context.Background()
	service, err := coordination.NewRedis(ctx, coordination.RedisOptions{
		Address:        os.Getenv("LINA_TEST_REDIS_ADDR"),
		DB:             db,
		Password:       os.Getenv("LINA_TEST_REDIS_PASSWORD"),
		ConnectTimeout: time.Second,
		ReadTimeout:    time.Second,
		WriteTimeout:   time.Second,
		KeyBuilder:     keys,
	})
	if err != nil {
		t.Fatalf("create redis cachecoord coordination service: %v", err)
	}
	t.Cleanup(func() {
		if err := service.Close(ctx); err != nil {
			t.Fatalf("close redis cachecoord coordination service: %v", err)
		}
	})
	return service
}

// publishRedisCacheCoordChanges publishes concurrent cachecoord revisions and
// returns every revision observed by the publishers.
func publishRedisCacheCoordChanges(
	t *testing.T,
	ctx context.Context,
	publisher Service,
	domain Domain,
	scope Scope,
	workers int,
) []int64 {
	t.Helper()

	revisions := make(chan int64, workers)
	errs := make(chan error, workers)
	var wg sync.WaitGroup
	for range workers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			revision, err := publisher.MarkChanged(
				ctx,
				domain,
				scope,
				ChangeReason("redis_cachecoord_concurrent_publish"),
			)
			if err != nil {
				errs <- err
				return
			}
			revisions <- revision
		}()
	}
	wg.Wait()
	close(revisions)
	close(errs)

	for err := range errs {
		if err != nil {
			t.Fatalf("publish redis cachecoord revision: %v", err)
		}
	}
	values := make([]int64, 0, workers)
	for revision := range revisions {
		values = append(values, revision)
	}
	return values
}

// assertRedisCacheCoordRevisions verifies concurrent Redis revision bumps are
// atomic and no increment is lost.
func assertRedisCacheCoordRevisions(t *testing.T, revisions []int64, workers int) {
	t.Helper()

	if len(revisions) != workers {
		t.Fatalf("expected %d redis cachecoord revisions, got %d: %#v", workers, len(revisions), revisions)
	}
	seen := make(map[int64]struct{}, workers)
	for _, revision := range revisions {
		seen[revision] = struct{}{}
	}
	if len(seen) != workers {
		t.Fatalf("expected %d unique redis cachecoord revisions, got %d: %#v", workers, len(seen), seen)
	}
	for revision := int64(1); revision <= int64(workers); revision++ {
		if _, ok := seen[revision]; !ok {
			t.Fatalf("expected redis cachecoord revision %d in %#v", revision, seen)
		}
	}
}

// waitForRedisCacheCoordEvent waits for the peer-node Redis invalidation event.
func waitForRedisCacheCoordEvent(
	t *testing.T,
	events <-chan coordination.Event,
	errs <-chan error,
	timeout time.Duration,
) coordination.Event {
	t.Helper()

	timer := time.NewTimer(timeout)
	defer timer.Stop()
	select {
	case event := <-events:
		return event
	case err := <-errs:
		t.Fatalf("handle cachecoord redis event: %v", err)
	case <-timer.C:
		t.Fatal("expected redis cachecoord invalidation event")
	}
	return coordination.Event{}
}

// waitForRedisCacheCoordRefresh waits until Redis pub/sub drives the consumer
// instance to the requested revision.
func waitForRedisCacheCoordRefresh(
	t *testing.T,
	revisions <-chan int64,
	errs <-chan error,
	target int64,
	timeout time.Duration,
) int64 {
	t.Helper()

	timer := time.NewTimer(timeout)
	defer timer.Stop()
	latest := int64(0)
	for latest < target {
		select {
		case revision := <-revisions:
			if revision > latest {
				latest = revision
			}
		case err := <-errs:
			t.Fatalf("handle cachecoord redis refresh: %v", err)
		case <-timer.C:
			t.Fatalf("expected redis cachecoord refresh revision %d, latest=%d", target, latest)
		}
	}
	return latest
}

// assertRedisCacheCoordSnapshot verifies cachecoord diagnostics reflect Redis
// backend state after the peer-node event refresh.
func assertRedisCacheCoordSnapshot(
	t *testing.T,
	items []SnapshotItem,
	domain Domain,
	scope Scope,
	revision int64,
) {
	t.Helper()

	for _, item := range items {
		if item.Domain != domain || item.Scope != scope {
			continue
		}
		if item.Backend != coordination.BackendRedis ||
			!item.CoordinationHealthy ||
			!item.EventSubscriberRunning ||
			item.LastEventReceivedAt.IsZero() {
			t.Fatalf("expected healthy redis cachecoord diagnostics, got %#v", item)
		}
		if item.LocalRevision != revision || item.SharedRevision != revision {
			t.Fatalf("expected redis cachecoord local/shared revision %d, got %#v", revision, item)
		}
		return
	}
	t.Fatalf("expected redis cachecoord snapshot for %q/%q, got %#v", domain, scope, items)
}

// cleanupRedisCacheCoordIntegrationKeys deletes exact Redis keys created by
// cachecoord integration tests without scanning or flushing the database.
func cleanupRedisCacheCoordIntegrationKeys(t *testing.T, keys ...string) {
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
			t.Fatalf("close redis cachecoord cleanup client: %v", err)
		}
	}()
	if err := client.Del(context.Background(), keys...).Err(); err != nil {
		t.Fatalf("cleanup redis cachecoord integration keys: %v", err)
	}
}
