// This file implements locker.Service methods against the LockStore bound at
// construction. SQL and Redis backends share that interface.

package locker

import (
	"context"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/internal/service/coordination"
	"lina-core/pkg/logger"
)

// Acquire obtains a lock when it is absent or expired.
func (s *serviceImpl) Acquire(
	ctx context.Context,
	name string,
	owner string,
	reason string,
	ttl time.Duration,
) (*coordination.LockHandle, bool, error) {
	if s == nil || s.store == nil {
		return nil, false, gerror.New("locker store is not bound at construction")
	}
	return s.store.Acquire(ctx, name, owner, reason, ttl)
}

// Renew extends a lock only when the caller still owns it.
func (s *serviceImpl) Renew(ctx context.Context, handle *coordination.LockHandle, ttl time.Duration) error {
	if s == nil || s.store == nil {
		return gerror.New("locker store is not bound at construction")
	}
	return mapCoordinationLockError(s.store.Renew(ctx, handle, ttl))
}

// Release releases a lock only when the caller still owns it.
func (s *serviceImpl) Release(ctx context.Context, handle *coordination.LockHandle) error {
	if s == nil || s.store == nil {
		return gerror.New("locker store is not bound at construction")
	}
	return mapCoordinationLockError(s.store.Release(ctx, handle))
}

// IsHeld reports whether the handle still owns the lock.
func (s *serviceImpl) IsHeld(ctx context.Context, handle *coordination.LockHandle) (bool, error) {
	if s == nil || s.store == nil {
		return false, gerror.New("locker store is not bound at construction")
	}
	return s.store.IsHeld(ctx, handle)
}

// Lock acquires a distributed lock when it is absent or expired.
func (s *serviceImpl) Lock(ctx context.Context, name, holder, reason string, lease time.Duration) (*Instance, bool, error) {
	if s == nil || s.store == nil {
		return nil, false, gerror.New("locker store is not bound at construction")
	}
	return lockWithCoordination(ctx, s.store, name, holder, reason, lease)
}

// timePtr returns a pointer to value for generated DO time fields that preserve
// database NULL semantics with *time.Time.
func timePtr(value time.Time) *time.Time {
	return &value
}

// LockFunc acquires a lock and executes the given function.
// The lock is automatically released after the function completes.
func (s *serviceImpl) LockFunc(ctx context.Context, name, holder, reason string, lease time.Duration, f func() error) (bool, error) {
	instance, ok, err := s.Lock(ctx, name, holder, reason, lease)
	if err != nil {
		return false, err
	}
	if !ok {
		return false, nil
	}
	defer func() {
		if unlockErr := instance.Unlock(ctx); unlockErr != nil {
			logger.Warningf(ctx, "[locker] failed to unlock '%s': %v", name, unlockErr)
		}
	}()
	if err = f(); err != nil {
		return true, err
	}
	return true, nil
}
