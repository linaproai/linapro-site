// Package session implements online-session storage and activity validation.
package session

import (
	"context"
	"time"

	"lina-core/internal/service/datascope"
	"lina-core/pkg/plugin/capability/tenantcap/tenantspi"
)

// sessionLastActiveUpdateWindow is the minimum interval between two
// last_active_time writes for one valid session.
const sessionLastActiveUpdateWindow time.Duration = time.Minute

const (
	sessionHotStateComponent = "session-hot-state"
	sessionHotStateSchema    = 1
	sessionUserIndexSchema   = 1
)

// Session represents an online user session.
type Session struct {
	TokenId        string     // Unique token identifier
	TenantId       int        // Tenant ID, where 0 means platform
	UserId         int        // User ID
	Username       string     // Username
	ClientType     string     // User-session client type
	DeptName       string     // Department name
	Ip             string     // Login IP address
	Browser        string     // Browser information
	Os             string     // Operating system
	LoginTime      *time.Time // Login time
	LastActiveTime *time.Time // Last active time
}

// ListFilter defines filter options for listing sessions.
type ListFilter struct {
	Username string // Username, supports fuzzy search
	Ip       string // Login IP, supports fuzzy search
}

// ListResult defines the result for paginated session list.
type ListResult struct {
	Items []*Session // Session items
	Total int        // Total count
}

// UserOnlineStatus reports projected online-session counts for one visible user.
type UserOnlineStatus struct {
	UserId       int // User ID
	SessionCount int // Number of visible online sessions
}

// Store defines the session storage interface for persistent online-session
// records used by authentication and timeout paths.
type Store interface {
	// Set persists one online session record.
	Set(ctx context.Context, session *Session) error
	// Get returns one online session by its globally unique token ID.
	Get(ctx context.Context, tokenId string) (*Session, error)
	// Delete removes one online session by its globally unique token ID.
	Delete(ctx context.Context, tokenId string) error
	// DeleteByUserId removes all online sessions that belong to one user in one tenant.
	DeleteByUserId(ctx context.Context, tenantId int, userId int) error
	// TouchOrValidate validates tenant ownership and session timeout, then
	// refreshes last_active_time outside the short write-throttle window for the
	// given tokenId. It returns true when the session remains valid.
	TouchOrValidate(ctx context.Context, tenantId int, tokenId string, timeout time.Duration) (bool, error)
}

// Directory is the online-user management projection over stored sessions.
type Directory interface {
	Store
	// Count returns the total number of active online sessions.
	Count(ctx context.Context) (int, error)
	// CleanupInactive deletes sessions whose last_active_time exceeds the given timeout duration.
	CleanupInactive(ctx context.Context, timeout time.Duration) (int64, error)
	// BatchGetScoped returns online sessions for the requested token IDs after
	// applying tenant ownership and data-scope constraints.
	BatchGetScoped(
		ctx context.Context,
		tokenIds []string,
		scopeSvc datascope.Service,
		tenantSvc tenantspi.ScopeService,
	) ([]*Session, error)
	// BatchGetUserOnlineStatusScoped returns online-session counts for the
	// requested users after applying tenant ownership and data-scope constraints.
	BatchGetUserOnlineStatusScoped(
		ctx context.Context,
		userIds []int,
		scopeSvc datascope.Service,
		tenantSvc tenantspi.ScopeService,
	) ([]*UserOnlineStatus, error)
	// List returns all online sessions that match the optional filter.
	List(ctx context.Context, filter *ListFilter) ([]*Session, error)
	// ListPage returns one paginated online-session list for the optional filter.
	ListPage(ctx context.Context, filter *ListFilter, pageNum, pageSize int) (*ListResult, error)
	// ListPageScoped returns one paginated online-session list constrained by
	// tenant ownership and the supplied data-scope service.
	ListPageScoped(
		ctx context.Context,
		filter *ListFilter,
		pageNum, pageSize int,
		scopeSvc datascope.Service,
		tenantSvc tenantspi.ScopeService,
	) (*ListResult, error)
}

var (
	_ Store     = (*DBStore)(nil)
	_ Directory = (*DBStore)(nil)
)

// DBStore implements Store using the persistent online-session table.
type DBStore struct{}

// sessionConfigurableStore extends Store with runtime session-timeout
// propagation for hot-state implementations.
type sessionConfigurableStore interface {
	Store
	// SetDefaultTTL updates the hot-state TTL used for login-time writes.
	SetDefaultTTL(ttl time.Duration)
}

// NewDBStore creates a PostgreSQL-backed session store. Clustered deployments
// must wrap this projection with NewCoordinationStore at construction time;
// this constructor never reads process-global coordination state.
func NewDBStore() Directory {
	return &DBStore{}
}
