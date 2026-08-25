// Package locker provides distributed lock acquisition, renewal, and lease
// management services for clustered host components.
package locker

import (
	"context"
	"time"

	"lina-core/internal/service/coordination"
)

// Service is the host lock helper over one LockStore implementation.
// Acquire/Renew/Release/IsHeld come from the bound store; LockFunc is the
// only extra convenience method.
type Service interface {
	coordination.LockStore
	// Lock acquires a distributed lock when it is absent or expired and returns
	// a held instance for later release. Prefer Acquire when the caller already
	// holds a LockHandle.
	Lock(ctx context.Context, name, holder, reason string, lease time.Duration) (*Instance, bool, error)
	// LockFunc acquires a lock and executes the given function.
	// The lock is automatically released after the function completes; function
	// errors are returned after successful acquisition. Failure to acquire
	// returns ok=false with nil error, while backend failures are propagated.
	LockFunc(ctx context.Context, name, holder, reason string, lease time.Duration, f func() error) (bool, error)
}

// Interface compliance assertion for the default locker service implementation.
var (
	_ Service                = (*serviceImpl)(nil)
	_ coordination.LockStore = (*serviceImpl)(nil)
)

// serviceImpl implements Service using the LockStore bound at construction.
// A nil store selects the SQL table implementation used by standalone nodes.
type serviceImpl struct {
	store coordination.LockStore
}

// New creates a locker Service bound to store. Pass a coordination LockStore
// for clustered deployments, or NewSQLStore() for standalone nodes. Callers
// must pass the construction-bound store; HTTP startup does not rewrite it.
func New(store coordination.LockStore) Service {
	return &serviceImpl{store: store}
}
