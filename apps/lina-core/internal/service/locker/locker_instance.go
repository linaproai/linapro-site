// This file implements operations on one acquired distributed lock instance.

package locker

import (
	"context"
	"time"

	"lina-core/internal/service/coordination"
	"lina-core/pkg/bizerr"
)

// Instance represents a distributed lock instance.
type Instance struct {
	id     int64                    // Lock record ID
	name   string                   // Name is the logical lock name.
	holder string                   // Node identifier that holds this lock
	lease  time.Duration            // Lease duration used when this lock was acquired
	handle *coordination.LockHandle // Handle stores coordination owner token metadata.
	store  coordination.LockStore   // store is the backend bound when this instance was acquired.
}

// ID returns the persistent lock record ID.
func (i *Instance) ID() int64 {
	return i.id
}

// Holder returns the current lock holder token.
func (i *Instance) Holder() string {
	if i != nil && i.handle != nil {
		return i.handle.Token
	}
	return i.holder
}

// Name returns the logical lock name when it is known by the backend.
func (i *Instance) Name() string {
	return i.name
}

// Unlock releases the lock by setting its expire_time to the past.
// This effectively releases the lock for other nodes to acquire.
func (i *Instance) Unlock(ctx context.Context) error {
	if i == nil {
		return ErrLockNotHeld
	}
	if i.handle == nil || i.store == nil {
		return ErrLockNotHeld
	}
	return mapCoordinationLockError(i.store.Release(ctx, i.handle))
}

// Renew extends the lock's expiration time.
// It only succeeds if the lock is still held by the current node and hasn't expired.
// Returns ErrLockNotHeld if the lock was lost or expired.
func (i *Instance) Renew(ctx context.Context) error {
	if i == nil {
		return ErrLockNotHeld
	}
	if i.handle == nil || i.store == nil {
		return ErrLockNotHeld
	}
	return mapCoordinationLockError(i.store.Renew(ctx, i.handle, i.lease))
}

// IsHeld checks if the lock is still held by the current node.
// A lock is considered held if its expire_time is in the future.
func (i *Instance) IsHeld(ctx context.Context) (bool, error) {
	if i == nil {
		return false, nil
	}
	if i.handle == nil || i.store == nil {
		return false, nil
	}
	held, err := i.store.IsHeld(ctx, i.handle)
	return held, mapCoordinationLockError(err)
}

// lockWithCoordination acquires one distributed lock through coordination.
func lockWithCoordination(
	ctx context.Context,
	lockStore coordination.LockStore,
	name string,
	holder string,
	reason string,
	lease time.Duration,
) (*Instance, bool, error) {
	handle, ok, err := lockStore.Acquire(ctx, name, holder, reason, lease)
	if err != nil {
		return nil, false, mapCoordinationLockError(err)
	}
	if !ok || handle == nil {
		return nil, false, nil
	}
	return &Instance{
		id:     handle.FencingToken,
		name:   name,
		holder: handle.Owner,
		lease:  lease,
		handle: handle,
		store:  lockStore,
	}, true, nil
}

// mapCoordinationLockError maps coordination ownership errors to locker errors.
func mapCoordinationLockError(err error) error {
	if err == nil {
		return nil
	}
	if bizerr.Is(err, coordination.CodeCoordinationLockNotHeld) {
		return ErrLockNotHeld
	}
	return err
}
