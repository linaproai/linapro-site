// This file tests locker service acquisition and lock-function behavior
// against the persistent lock table.

package locker

import (
	"context"
	"fmt"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/gogf/gf/v2/test/gtest"
	"lina-core/internal/service/coordination"
	_ "lina-core/pkg/dbdriver"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// testHolder is the holder token shared by locker integration tests.
const testHolder = "test-node"

// newTestService creates a new locker service for testing.
func newTestService() *serviceImpl {
	return New().(*serviceImpl)
}

// cleanupLock removes the lock by name after test.
func cleanupLock(name string) {
	if _, err := g.DB().Model("sys_locker").Where("name", name).Delete(); err != nil {
		panic(fmt.Sprintf("cleanup locker row failed name=%s err=%v", name, err))
	}
}

// TestService_New verifies New returns a non-nil locker service.
func TestService_New(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		svc := newTestService()
		t.AssertNE(svc, nil)
	})
}

// TestService_Lock_NewLock verifies acquiring a missing lock creates one new
// persistent lock row.
func TestService_Lock_NewLock(t *testing.T) {
	var (
		svc    = newTestService()
		name   = "test-lock-new-" + gtime.TimestampMilliStr()
		reason = "test reason"
		ctx    = context.Background()
	)

	cleanupLock(name)

	gtest.C(t, func(t *gtest.T) {
		instance, ok, err := svc.Lock(ctx, name, testHolder, reason, 30*time.Second)
		t.AssertNil(err)
		t.Assert(ok, true)
		t.AssertNE(instance, nil)

		count, err := g.DB().Model("sys_locker").Where("name", name).Count()
		t.AssertNil(err)
		t.Assert(count, 1)

		err = instance.Unlock(ctx)
		t.AssertNil(err)
	})

	cleanupLock(name)
}

// TestService_Lock_ExistingExpiredLock verifies an expired lock can be taken
// over and rewritten by the current holder.
func TestService_Lock_ExistingExpiredLock(t *testing.T) {
	var (
		svc    = newTestService()
		name   = "test-lock-expired-" + gtime.TimestampMilliStr()
		reason = "test reason"
		ctx    = context.Background()
	)

	cleanupLock(name)

	_, err := g.DB().Model("sys_locker").Data(g.Map{
		"name":        name,
		"reason":      "old reason",
		"holder":      "other-node",
		"expire_time": time.Now().Add(-10 * time.Second),
	}).Insert()
	if err != nil {
		t.Fatal(err)
	}

	gtest.C(t, func(t *gtest.T) {
		instance, ok, err := svc.Lock(ctx, name, testHolder, reason, 30*time.Second)
		t.AssertNil(err)
		t.Assert(ok, true)
		t.AssertNE(instance, nil)

		var row struct {
			Holder string
			Reason string
		}
		err = g.DB().Model("sys_locker").Where("name", name).Scan(&row)
		t.AssertNil(err)
		t.Assert(row.Holder, testHolder)
		t.Assert(row.Reason, reason)

		err = instance.Unlock(ctx)
		t.AssertNil(err)
	})

	cleanupLock(name)
}

// TestIsExpiredLockUsesExpireTime verifies lock takeover decisions are based
// on expire_time before holder data is reused.
func TestIsExpiredLockUsesExpireTime(t *testing.T) {
	now := time.Now()
	past := now.Add(-time.Second)
	if !isExpiredLock(&past, now) {
		t.Fatal("expected past expire_time to be expired")
	}
	future := now.Add(time.Second)
	if isExpiredLock(&future, now) {
		t.Fatal("expected future expire_time to remain held")
	}
}

// TestService_Lock_ExistingNonExpiredLock verifies a lock held by another node
// cannot be acquired before expiry.
func TestService_Lock_ExistingNonExpiredLock(t *testing.T) {
	var (
		svc    = newTestService()
		name   = "test-lock-active-" + gtime.TimestampMilliStr()
		reason = "test reason"
		ctx    = context.Background()
	)

	cleanupLock(name)

	_, err := g.DB().Model("sys_locker").Data(g.Map{
		"name":        name,
		"reason":      "old reason",
		"holder":      "other-node",
		"expire_time": time.Now().Add(30 * time.Second),
	}).Insert()
	if err != nil {
		t.Fatal(err)
	}

	gtest.C(t, func(t *gtest.T) {
		instance, ok, err := svc.Lock(ctx, name, testHolder, reason, 30*time.Second)
		t.AssertNil(err)
		t.Assert(ok, false)
		t.Assert(instance, nil)
	})

	cleanupLock(name)
}

// TestService_Lock_RecordSurvivesServiceRecreation verifies a valid lock row
// remains effective when a new service instance is constructed after restart.
func TestService_Lock_RecordSurvivesServiceRecreation(t *testing.T) {
	var (
		firstService  = newTestService()
		secondService = newTestService()
		name          = "test-lock-restart-" + gtime.TimestampMilliStr()
		reason        = "test reason"
		ctx           = context.Background()
	)

	cleanupLock(name)

	instance, ok, err := firstService.Lock(ctx, name, testHolder, reason, 30*time.Second)
	if err != nil {
		t.Fatalf("acquire lock before service recreation: %v", err)
	}
	if !ok || instance == nil {
		t.Fatal("expected first service to acquire lock")
	}

	restartedInstance, ok, err := secondService.Lock(ctx, name, "other-node", "after restart", 30*time.Second)
	if err != nil {
		t.Fatalf("acquire lock after service recreation: %v", err)
	}
	if ok || restartedInstance != nil {
		t.Fatal("expected valid lock to remain held after service recreation")
	}

	if err = instance.Unlock(ctx); err != nil {
		t.Fatalf("unlock retained lock: %v", err)
	}
	cleanupLock(name)
}

// TestService_Lock_ConcurrentFreshLockRace verifies duplicate insert races are
// reported as clean acquisition misses instead of database errors.
func TestService_Lock_ConcurrentFreshLockRace(t *testing.T) {
	var (
		name   = "test-lock-concurrent-" + gtime.TimestampMilliStr()
		reason = "test concurrent reason"
		ctx    = context.Background()
	)

	cleanupLock(name)
	t.Cleanup(func() {
		cleanupLock(name)
	})

	const contenders = 16
	var (
		start      = make(chan struct{})
		ready      sync.WaitGroup
		done       sync.WaitGroup
		successes  int32
		failures   int32
		firstError = make(chan error, 1)
	)

	for i := 0; i < contenders; i++ {
		holder := fmt.Sprintf("test-holder-%02d", i)
		ready.Add(1)
		done.Add(1)
		go func(holder string) {
			defer done.Done()
			svc := newTestService()
			ready.Done()
			<-start
			instance, ok, err := svc.Lock(ctx, name, holder, reason, 30*time.Second)
			if err != nil {
				select {
				case firstError <- err:
				default:
				}
				atomic.AddInt32(&failures, 1)
				return
			}
			if ok {
				atomic.AddInt32(&successes, 1)
				if instance == nil {
					select {
					case firstError <- fmt.Errorf("nil instance for holder %s", holder):
					default:
					}
				}
				return
			}
			if instance != nil {
				select {
				case firstError <- fmt.Errorf("unexpected instance for holder %s", holder):
				default:
				}
			}
		}(holder)
	}
	ready.Wait()
	close(start)
	done.Wait()

	select {
	case errValue := <-firstError:
		t.Fatalf("concurrent lock acquisition surfaced error: %v", errValue)
	default:
	}
	if failures != 0 {
		t.Fatalf("expected no failed acquisitions, got %d", failures)
	}
	if successes != 1 {
		t.Fatalf("expected exactly one successful acquisition, got %d", successes)
	}
	count, err := g.DB().Model("sys_locker").Where("name", name).Count()
	if err != nil {
		t.Fatalf("count lock row failed: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected exactly one lock row, got %d", count)
	}
}

// TestService_Lock_SameHolder verifies the current holder can reacquire its
// own lock and extend ownership.
func TestService_Lock_SameHolder(t *testing.T) {
	var (
		svc    = newTestService()
		name   = "test-lock-same-holder-" + gtime.TimestampMilliStr()
		reason = "test reason"
		ctx    = context.Background()
	)

	cleanupLock(name)

	gtest.C(t, func(t *gtest.T) {
		instance1, ok1, err1 := svc.Lock(ctx, name, testHolder, reason, 30*time.Second)
		t.AssertNil(err1)
		t.Assert(ok1, true)

		instance2, ok2, err2 := svc.Lock(ctx, name, testHolder, "new reason", 30*time.Second)
		t.AssertNil(err2)
		t.Assert(ok2, true)

		unlockErr := instance1.Unlock(ctx)
		t.AssertNil(unlockErr)
		unlockErr = instance2.Unlock(ctx)
		t.AssertNil(unlockErr)
	})

	cleanupLock(name)
}

// TestService_LockFunc verifies LockFunc executes the callback and releases the
// lock afterward.
func TestService_LockFunc(t *testing.T) {
	var (
		svc      = newTestService()
		name     = "test-lock-func-" + gtime.TimestampMilliStr()
		executed = false
		reason   = "test reason"
		ctx      = context.Background()
	)

	cleanupLock(name)

	gtest.C(t, func(t *gtest.T) {
		ok, err := svc.LockFunc(ctx, name, testHolder, reason, 30*time.Second, func() error {
			executed = true
			return nil
		})
		t.AssertNil(err)
		t.Assert(ok, true)
		t.Assert(executed, true)

		count, err := g.DB().Model("sys_locker").Where("name", name).
			WhereGTE("expire_time", time.Now()).Count()
		t.AssertNil(err)
		t.Assert(count, 0)
	})

	cleanupLock(name)

	gtest.C(t, func(t *gtest.T) {
		executed = false
		ok, err := svc.LockFunc(ctx, name, testHolder, reason, 30*time.Second, func() error {
			executed = true
			return ErrLockNotHeld
		})
		t.AssertNE(err, nil)
		t.Assert(ok, true)
		t.Assert(executed, true)
	})

	cleanupLock(name)
}

// TestService_LockFunc_AlreadyLocked verifies LockFunc does not execute the
// callback when another holder already owns the lock.
func TestService_LockFunc_AlreadyLocked(t *testing.T) {
	var (
		svc    = newTestService()
		name   = "test-lock-func-locked-" + gtime.TimestampMilliStr()
		reason = "test reason"
		ctx    = context.Background()
	)

	cleanupLock(name)

	_, err := g.DB().Model("sys_locker").Data(g.Map{
		"name":        name,
		"reason":      "other reason",
		"holder":      "other-node",
		"expire_time": time.Now().Add(30 * time.Second),
	}).Insert()
	if err != nil {
		t.Fatal(err)
	}

	gtest.C(t, func(t *gtest.T) {
		executed := false
		ok, err := svc.LockFunc(ctx, name, testHolder, reason, 30*time.Second, func() error {
			executed = true
			return nil
		})
		t.AssertNil(err)
		t.Assert(ok, false)
		t.Assert(executed, false)
	})

	cleanupLock(name)
}

// TestCoordinationLockerLifecycle verifies locks use coordination ownership
// tokens for acquire, renew, held-state checks, and release.
func TestCoordinationLockerLifecycle(t *testing.T) {
	ctx := context.Background()
	ConfigureCoordination(coordination.NewMemory(nil))
	t.Cleanup(func() {
		ConfigureCoordination(nil)
	})

	svc := New()
	instance, ok, err := svc.Lock(ctx, "unit-coord-lock", "node-a", "unit", time.Second)
	if err != nil {
		t.Fatalf("acquire coordination lock: %v", err)
	}
	if !ok || instance == nil {
		t.Fatal("expected coordination lock acquisition")
	}
	if instance.ID() <= 0 || instance.Holder() == "node-a" || instance.Name() != "unit-coord-lock" {
		t.Fatalf("expected coordination instance metadata, got id=%d holder=%q name=%q", instance.ID(), instance.Holder(), instance.Name())
	}
	if held, err := instance.IsHeld(ctx); err != nil || !held {
		t.Fatalf("expected coordination lock held, held=%t err=%v", held, err)
	}
	if err = instance.Renew(ctx); err != nil {
		t.Fatalf("renew coordination lock through instance: %v", err)
	}
	if err = svc.RenewByName(ctx, instance.Name(), instance.Holder(), time.Second); err != nil {
		t.Fatalf("renew coordination lock by name: %v", err)
	}
	if err = svc.UnlockByName(ctx, instance.Name(), "wrong-token"); err != ErrLockNotHeld {
		t.Fatalf("expected wrong-token release to fail, got %v", err)
	}
	if err = svc.UnlockByName(ctx, instance.Name(), instance.Holder()); err != nil {
		t.Fatalf("release coordination lock by name: %v", err)
	}
	if held, err := instance.IsHeld(ctx); err != nil || held {
		t.Fatalf("expected coordination lock released, held=%t err=%v", held, err)
	}
}

// TestCoordinationLockerIsolatesNames verifies logical lock names remain
// independent even when held by the same owner.
func TestCoordinationLockerIsolatesNames(t *testing.T) {
	ctx := context.Background()
	ConfigureCoordination(coordination.NewMemory(nil))
	t.Cleanup(func() {
		ConfigureCoordination(nil)
	})

	svc := New()
	first, ok, err := svc.Lock(ctx, "plugin:a:sync", "node-a", "first", time.Second)
	if err != nil || !ok || first == nil {
		t.Fatalf("acquire first coordination lock, ok=%t err=%v", ok, err)
	}
	second, ok, err := svc.Lock(ctx, "plugin:b:sync", "node-a", "second", time.Second)
	if err != nil || !ok || second == nil {
		t.Fatalf("acquire second coordination lock, ok=%t err=%v", ok, err)
	}
	if first.Holder() == second.Holder() {
		t.Fatal("expected distinct coordination owner tokens for isolated locks")
	}
	if err = first.Unlock(ctx); err != nil {
		t.Fatalf("release first coordination lock: %v", err)
	}
	if held, err := second.IsHeld(ctx); err != nil || !held {
		t.Fatalf("expected second coordination lock to remain held, held=%t err=%v", held, err)
	}
}

// TestCoordinationLockerFailureReturnsError verifies coordination backend
// failures are surfaced instead of being treated as acquisition misses.
func TestCoordinationLockerFailureReturnsError(t *testing.T) {
	ctx := context.Background()
	coordSvc := coordination.NewMemory(nil)
	ConfigureCoordination(coordSvc)
	t.Cleanup(func() {
		ConfigureCoordination(nil)
	})
	if err := coordSvc.Close(ctx); err != nil {
		t.Fatalf("close coordination backend: %v", err)
	}

	if instance, ok, err := New().Lock(ctx, "unit-closed-lock", "node-a", "unit", time.Second); err == nil || ok || instance != nil {
		t.Fatalf("expected coordination lock failure, instance=%#v ok=%t err=%v", instance, ok, err)
	}
}

// isExpiredLock reports whether one lock row is available for takeover.
func isExpiredLock(expireTime *time.Time, now time.Time) bool {
	if expireTime == nil {
		return true
	}
	return now.After(*expireTime)
}
