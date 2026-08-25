// This file implements coordination.LockStore against the standalone sys_locker
// table so clustered and single-node lockers share one backend interface.

package locker

import (
	"context"
	"database/sql"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
	"lina-core/internal/service/coordination"
	"lina-core/pkg/dialect"
	"lina-core/pkg/logger"
)

// sqlLockStore persists locks in sys_locker for standalone deployments.
type sqlLockStore struct{}

// NewSQLStore returns a LockStore backed by the sys_locker table.
func NewSQLStore() coordination.LockStore {
	return sqlLockStore{}
}

// Acquire obtains a lock when it is absent or expired.
func (sqlLockStore) Acquire(
	ctx context.Context,
	name string,
	owner string,
	reason string,
	ttl time.Duration,
) (*coordination.LockHandle, bool, error) {
	var locker *entity.SysLocker
	err := dao.SysLocker.Ctx(ctx).Where(do.SysLocker{Name: name}).Scan(&locker)
	if err != nil {
		return nil, false, err
	}

	now := time.Now()
	expireTime := now.Add(ttl)
	if locker == nil {
		result, insertErr := dao.SysLocker.Ctx(ctx).Data(do.SysLocker{
			Name:       name,
			Reason:     reason,
			Holder:     owner,
			ExpireTime: &expireTime,
		}).Insert()
		if insertErr != nil {
			if dialect.IsUniqueConstraintViolation(insertErr) {
				return nil, false, nil
			}
			return nil, false, insertErr
		}
		insertId, idErr := result.LastInsertId()
		if idErr != nil {
			return nil, false, idErr
		}
		if insertId <= 0 {
			return nil, false, nil
		}
		logger.Infof(ctx, "[locker] acquired lock '%s' (holder: %s)", name, owner)
		return sqlLockHandle(name, owner, reason, ttl, insertId, now), true, nil
	}

	if locker.Holder == owner {
		_, updateErr := dao.SysLocker.Ctx(ctx).Data(do.SysLocker{
			ExpireTime: &expireTime,
		}).Where(do.SysLocker{Id: locker.Id}).Update()
		if updateErr != nil {
			return nil, false, updateErr
		}
		return sqlLockHandle(name, owner, reason, ttl, int64(locker.Id), now), true, nil
	}

	cols := dao.SysLocker.Columns()
	affected, takeoverErr := dao.SysLocker.Ctx(ctx).Data(do.SysLocker{
		Reason:     reason,
		Holder:     owner,
		ExpireTime: &expireTime,
	}).Where(do.SysLocker{Id: locker.Id}).
		Wheref("(%s IS NULL OR %s < ?)", cols.ExpireTime, cols.ExpireTime, now).
		UpdateAndGetAffected()
	if takeoverErr != nil {
		return nil, false, takeoverErr
	}
	if affected <= 0 {
		return nil, false, nil
	}
	logger.Infof(ctx, "[locker] acquired expired lock '%s' (holder: %s)", name, owner)
	return sqlLockHandle(name, owner, reason, ttl, int64(locker.Id), now), true, nil
}

// Renew extends a lock only when the caller still owns it.
func (sqlLockStore) Renew(ctx context.Context, handle *coordination.LockHandle, ttl time.Duration) error {
	if handle == nil {
		return ErrLockNotHeld
	}
	now := time.Now()
	expireTime := now.Add(ttl)
	model := dao.SysLocker.Ctx(ctx).Where(do.SysLocker{Holder: handleOwner(handle)})
	if handle.FencingToken > 0 {
		model = model.Where(do.SysLocker{Id: handle.FencingToken})
	} else {
		model = model.Where(do.SysLocker{Name: handle.Name})
	}
	var locker struct {
		Id int64
	}
	err := model.WhereGT("expire_time", now).Scan(&locker)
	if err != nil {
		if gerror.Is(err, sql.ErrNoRows) {
			return ErrLockNotHeld
		}
		return err
	}
	if locker.Id == 0 {
		return ErrLockNotHeld
	}
	_, err = dao.SysLocker.Ctx(ctx).Data(do.SysLocker{ExpireTime: &expireTime}).Where(do.SysLocker{Id: locker.Id}).Update()
	return err
}

// Release releases a lock only when the caller still owns it.
func (sqlLockStore) Release(ctx context.Context, handle *coordination.LockHandle) error {
	if handle == nil {
		return ErrLockNotHeld
	}
	model := dao.SysLocker.Ctx(ctx).Data(do.SysLocker{
		ExpireTime: timePtr(time.Now().Add(-1 * time.Second)),
	}).Where(do.SysLocker{Holder: handleOwner(handle)})
	if handle.FencingToken > 0 {
		model = model.Where(do.SysLocker{Id: handle.FencingToken})
	} else {
		model = model.Where(do.SysLocker{Name: handle.Name})
	}
	_, err := model.Update()
	return err
}

// IsHeld reports whether the handle still owns the lock.
func (sqlLockStore) IsHeld(ctx context.Context, handle *coordination.LockHandle) (bool, error) {
	if handle == nil {
		return false, nil
	}
	model := dao.SysLocker.Ctx(ctx)
	if handle.FencingToken > 0 {
		model = model.Where(do.SysLocker{Id: handle.FencingToken})
	} else {
		model = model.Where(do.SysLocker{
			Name:   handle.Name,
			Holder: handleOwner(handle),
		})
	}
	count, err := model.WhereGT("expire_time", time.Now()).Count()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// sqlLockHandle builds a coordination lock handle from a sys_locker row.
func sqlLockHandle(
	name string,
	owner string,
	reason string,
	ttl time.Duration,
	id int64,
	acquiredAt time.Time,
) *coordination.LockHandle {
	return &coordination.LockHandle{
		Name:         name,
		Owner:        owner,
		Token:        owner,
		Reason:       reason,
		Lease:        ttl,
		FencingToken: id,
		AcquiredAt:   acquiredAt,
	}
}

// handleOwner returns the persisted locker holder, falling back to Token when
// Owner is empty so legacy rows still match.
func handleOwner(handle *coordination.LockHandle) string {
	if handle == nil {
		return ""
	}
	if handle.Owner != "" {
		return handle.Owner
	}
	return handle.Token
}
