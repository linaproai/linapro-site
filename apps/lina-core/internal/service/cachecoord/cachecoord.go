// Package cachecoord provides topology-aware revision coordination for
// process-local cache domains.
package cachecoord

import (
	"context"
	"sync"
	"time"

	"lina-core/internal/service/cluster"
	"lina-core/internal/service/coordination"
)

// Domain identifies one cache domain coordinated by the host.
type Domain string

// Scope identifies one explicit invalidation scope inside a cache domain.
type Scope string

// ConsistencyModel names the coordination model used by one cache domain.
type ConsistencyModel string

// FailureStrategy names how callers should degrade when freshness cannot be
// confirmed inside the declared stale window.
type FailureStrategy string

// ChangeReason describes why a cache domain revision was published.
type ChangeReason string

// TenantID identifies the tenant scope of one cache invalidation message.
type TenantID int

// Cache scope constants centralize stable invalidation scopes.
const (
	// ScopeGlobal invalidates the whole cache domain.
	ScopeGlobal Scope = "global"
	// ScopeReconciler invalidates dynamic-plugin runtime reconciler wake-up state.
	ScopeReconciler Scope = "reconciler"
)

// Cache consistency model constants describe runtime review metadata.
const (
	// ConsistencyLocalOnly keeps coordination in the current process.
	ConsistencyLocalOnly ConsistencyModel = "local-only"
	// ConsistencySharedRevision uses a shared revision in the injected
	// coordination backend (Redis in clustered deployments).
	ConsistencySharedRevision ConsistencyModel = "shared-revision"
)

// Cache failure strategy constants describe caller-visible degradation behavior.
const (
	// FailureStrategyFailClosed rejects uncertain access after the stale window.
	FailureStrategyFailClosed FailureStrategy = "fail-closed"
	// FailureStrategyReturnVisibleError returns a visible runtime error to callers.
	FailureStrategyReturnVisibleError FailureStrategy = "return-visible-error"
	// FailureStrategyConservativeHide hides uncertain plugin capabilities.
	FailureStrategyConservativeHide FailureStrategy = "conservative-hide"
)

// DefaultDomainMaxStale is the freshness budget used by domains that do not
// configure a domain-specific stale window.
const (
	DefaultDomainMaxStale = 5 * time.Second
)

// Refresher rebuilds or invalidates one process-local cache domain after a
// newer revision is observed.
type Refresher func(ctx context.Context, revision int64) error

// DomainSpec optionally declares the reviewable consistency contract for one
// cache domain.
type DomainSpec struct {
	Domain           Domain           // Domain is the stable cache domain identifier.
	AuthoritySource  string           // AuthoritySource describes the canonical data source.
	ConsistencyModel ConsistencyModel // ConsistencyModel describes local or shared revision coordination.
	MaxStale         time.Duration    // MaxStale is the maximum acceptable local stale window.
	SyncMechanism    string           // SyncMechanism describes cross-node synchronization.
	FailureStrategy  FailureStrategy  // FailureStrategy describes degradation after MaxStale.
}

// SnapshotItem exposes one cache domain and scope coordination status.
type SnapshotItem struct {
	Domain                 Domain                   // Domain is the cache domain identifier.
	Scope                  Scope                    // Scope is the explicit invalidation scope.
	AuthoritySource        string                   // AuthoritySource is the canonical data source.
	ConsistencyModel       ConsistencyModel         // ConsistencyModel is the declared consistency model.
	MaxStale               time.Duration            // MaxStale is the configured stale window.
	FailureStrategy        FailureStrategy          // FailureStrategy is the configured degradation behavior.
	LocalRevision          int64                    // LocalRevision is the latest revision consumed locally.
	SharedRevision         int64                    // SharedRevision is the latest shared revision when cluster mode is enabled.
	LastSyncedAt           time.Time                // LastSyncedAt records the latest successful local sync.
	Backend                coordination.BackendName // Backend is the active coordination backend for this snapshot.
	CoordinationHealthy    bool                     // CoordinationHealthy reports the backend health snapshot when clustered coordination is active.
	EventSubscriberRunning bool                     // EventSubscriberRunning reports whether the backend event consumer is active.
	LastEventReceivedAt    time.Time                // LastEventReceivedAt records the latest consumed backend event time.
	RecentError            string                   // RecentError records the latest coordination failure.
	StaleSeconds           int64                    // StaleSeconds reports seconds elapsed since LastSyncedAt.
}

// InvalidationScope declares the tenant range for one cache invalidation.
type InvalidationScope struct {
	TenantID         TenantID // TenantID is the target tenant, 0 platform, or -1 all tenants.
	CascadeToTenants bool     // CascadeToTenants invalidates tenant buckets after platform default changes.
}

// Service defines the cache coordination contract.
type Service interface {
	// ConfigureDomain configures or replaces one cache domain consistency contract.
	// The spec declares the authority source, stale window, and degradation
	// policy used by Snapshot and review tooling; it does not invalidate cached
	// data by itself. Invalid domain identifiers return a business error.
	ConfigureDomain(spec DomainSpec) error
	// MarkChanged publishes one explicit cache domain/scope revision change.
	// In standalone mode the revision is process-local; in clustered mode it is
	// written through the coordination backend so other instances can refresh.
	// Invalid domains/scopes or backend failures are returned to the caller.
	MarkChanged(ctx context.Context, domain Domain, scope Scope, reason ChangeReason) (int64, error)
	// MarkTenantChanged publishes one tenant-scoped cache domain/scope revision change.
	// The tenant scope is folded into the revision key so tenant-local
	// invalidations stay explicit; cascade flags are preserved in the keying
	// contract and backend failures are returned.
	MarkTenantChanged(ctx context.Context, domain Domain, scope Scope, tenantScope InvalidationScope, reason ChangeReason) (int64, error)
	// EnsureFresh refreshes local state if the shared or local revision advanced.
	// The refresher runs after a newer revision is observed and must rebuild the
	// caller's cache idempotently. Refresh or coordination errors are returned;
	// successful calls record local freshness for Snapshot diagnostics.
	EnsureFresh(ctx context.Context, domain Domain, scope Scope, refresher Refresher) (int64, error)
	// CurrentRevision returns the latest visible revision for one domain/scope.
	// Clustered mode reads the shared revision when available; standalone mode
	// returns the local process revision. Backend errors are propagated.
	CurrentRevision(ctx context.Context, domain Domain, scope Scope) (int64, error)
	// Snapshot returns observable status for configured cache domains and touched scopes.
	// The result is diagnostic-only and includes declared consistency metadata,
	// local freshness, shared revision, backend health, and recent errors.
	Snapshot(ctx context.Context) ([]SnapshotItem, error)
}

// Interface compliance assertion for the default cachecoord service.
var _ Service = (*serviceImpl)(nil)

// serviceImpl implements Service.
type serviceImpl struct {
	topologyMu     sync.RWMutex
	topology       cluster.Service
	coordMu        sync.RWMutex
	coord          coordination.Service
	mu             sync.RWMutex
	domains        map[Domain]DomainSpec
	observed       map[revisionKey]int64
	status         map[revisionKey]*coordinationStatus
	localRevisions map[revisionKey]int64
}

// coordinationStatus stores local observable state for one domain/scope.
type coordinationStatus struct {
	localRevision  int64
	sharedRevision int64
	lastSyncedAt   time.Time
	recentError    string
	recentErrorAt  time.Time
}

// New creates one cache coordination service bound to topology and the
// optional coordination backend. A nil topology uses the standalone static
// view. A nil coordination backend keeps revisions process-local. Callers
// must pass the startup-owned instance; this constructor does not share
// process-global state with later New calls.
func New(topology cluster.Service, coordinationSvc coordination.Service) Service {
	service := newServiceImpl(topology)
	service.setCoordination(coordinationSvc)
	return service
}

// newServiceImpl allocates one cache coordination implementation.
func newServiceImpl(topology cluster.Service) *serviceImpl {
	if topology == nil {
		topology = NewStaticTopology(false)
	}
	service := &serviceImpl{
		topology:       topology,
		domains:        make(map[Domain]DomainSpec),
		observed:       make(map[revisionKey]int64),
		status:         make(map[revisionKey]*coordinationStatus),
		localRevisions: make(map[revisionKey]int64),
	}
	return service
}
