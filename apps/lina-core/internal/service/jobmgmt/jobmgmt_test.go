// This file keeps shared scheduled-job management test helpers.

package jobmgmt

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gogf/gf/v2/net/ghttp"
	jobv1 "lina-core/api/job/v1"
	"lina-core/internal/dao"
	"lina-core/internal/model"
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
	"lina-core/internal/service/bizctx"
	"lina-core/internal/service/cachecoord"
	hostconfig "lina-core/internal/service/config"
	"lina-core/internal/service/datascope"
	i18nsvc "lina-core/internal/service/i18n"
	"lina-core/internal/service/jobmeta"
	"lina-core/internal/service/role"
	"lina-core/pkg/bizerr"
	_ "lina-core/pkg/dbdriver"
	"lina-core/pkg/plugin/capability/bizctxcap"
	"lina-core/pkg/plugin/capability/orgcap/orgspi"
	"lina-core/pkg/plugin/capability/tenantcap/tenantspi"
	"sync"
	"testing"
	"time"
)

// noopScheduler keeps job-management unit tests focused on validation and persistence.
type noopScheduler struct{}

// LoadAndRegister is a no-op for validation-focused unit tests.
func (noopScheduler) LoadAndRegister(ctx context.Context) error { return nil }

// Refresh is a no-op for validation-focused unit tests.
func (noopScheduler) Refresh(ctx context.Context, jobID int64) error { return nil }

// RegisterJobSnapshot is a no-op for validation-focused unit tests.
func (noopScheduler) RegisterJobSnapshot(ctx context.Context, job *entity.SysJob) error { return nil }

// Remove is a no-op for validation-focused unit tests.
func (noopScheduler) Remove(jobID int64) {}

// Trigger is unsupported in validation-focused unit tests.
func (noopScheduler) Trigger(ctx context.Context, jobID int64) (int64, error) { return 0, nil }

// CancelLog is unsupported in validation-focused unit tests.
func (noopScheduler) CancelLog(ctx context.Context, logID int64) error { return nil }

// jobmgmtStaticBizCtx returns a fixed request business context for service tests.
type jobmgmtStaticBizCtx struct {
	ctx *model.Context
}

// Init is unused by service tests because they inject context directly.
func (s jobmgmtStaticBizCtx) Init(_ *ghttp.Request, _ *model.Context) {}

// Get returns the configured business context.
func (s jobmgmtStaticBizCtx) Get(context.Context) *model.Context { return s.ctx }

// Current returns the plugin-visible business context projection.
func (s jobmgmtStaticBizCtx) Current(context.Context) bizctxcap.CurrentContext {
	if s.ctx == nil {
		return bizctxcap.CurrentContext{}
	}
	return bizctxcap.CurrentContext{
		UserID:          s.ctx.UserId,
		Username:        s.ctx.Username,
		TenantID:        s.ctx.TenantId,
		ActingUserID:    s.ctx.ActingUserId,
		ActingAsTenant:  s.ctx.ActingAsTenant,
		IsImpersonation: s.ctx.IsImpersonation,
		PlatformBypass:  s.ctx.TenantId == 0,
	}
}

// SetLocale is unused by job-management service tests.
func (s jobmgmtStaticBizCtx) SetLocale(context.Context, string) {}

// SetUser is unused by job-management service tests.
func (s jobmgmtStaticBizCtx) SetUser(context.Context, string, int, string, int, string) {}

// SetTenant is unused by job-management service tests.
func (s jobmgmtStaticBizCtx) SetTenant(context.Context, int) {}

// SetImpersonation is unused by job-management service tests.
func (s jobmgmtStaticBizCtx) SetImpersonation(context.Context, int, int, bool, bool) {}

// SetUserAccess is unused by job-management service tests.
func (s jobmgmtStaticBizCtx) SetUserAccess(context.Context, int, bool, int) {}

// trackingScheduler captures refresh and remove calls for registry-cascade tests.
type trackingScheduler struct {
	mu        sync.Mutex
	refreshed []int64
	removed   []int64
}

// LoadAndRegister is a no-op for registry-cascade tests.
func (s *trackingScheduler) LoadAndRegister(ctx context.Context) error { return nil }

// Refresh records refreshed job IDs.
func (s *trackingScheduler) Refresh(ctx context.Context, jobID int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.refreshed = append(s.refreshed, jobID)
	return nil
}

// RegisterJobSnapshot records refreshed job IDs for declaration-driven tests.
func (s *trackingScheduler) RegisterJobSnapshot(ctx context.Context, job *entity.SysJob) error {
	if job == nil {
		return nil
	}
	return s.Refresh(ctx, job.Id)
}

// Remove records removed job IDs.
func (s *trackingScheduler) Remove(jobID int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.removed = append(s.removed, jobID)
}

// Trigger is unsupported in registry-cascade tests.
func (s *trackingScheduler) Trigger(ctx context.Context, jobID int64) (int64, error) { return 0, nil }

// CancelLog is unsupported in registry-cascade tests.
func (s *trackingScheduler) CancelLog(ctx context.Context, logID int64) error { return nil }

// reset clears recorded scheduler calls.
func (s *trackingScheduler) reset() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.refreshed = nil
	s.removed = nil
}

// refreshedIDs returns one copy of all recorded refresh calls.
func (s *trackingScheduler) refreshedIDs() []int64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]int64(nil), s.refreshed...)
}

// removedIDs returns one copy of all recorded remove calls.
func (s *trackingScheduler) removedIDs() []int64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]int64(nil), s.removed...)
}

// containsJobID reports whether one scheduler call snapshot contains the given job ID.
func containsJobID(jobIDs []int64, target int64) bool {
	for _, jobID := range jobIDs {
		if jobID == target {
			return true
		}
	}
	return false
}

// noopCleaner keeps host-handler registration lightweight for unit tests.
type noopCleaner struct{}

// CleanupDueLogs is a no-op for host handler registration tests.
func (noopCleaner) CleanupDueLogs(ctx context.Context) (int64, error) { return 0, nil }

// newTestService constructs one DB-backed job-management service with host handlers registered.
func newTestService(t *testing.T) *serviceImpl {
	t.Helper()

	registry := NewRegistry()
	if err := RegisterHostHandlers(registry, noopCleaner{}); err != nil {
		t.Fatalf("expected host handler registration to succeed, got error: %v", err)
	}
	return newTestServiceWithExplicitDependencies(t, registry, noopScheduler{})
}

// newTestServiceWithRegistry constructs one DB-backed job-management service with
// explicit registry and scheduler dependencies for lifecycle-cascade tests.
func newTestServiceWithRegistry(
	t *testing.T,
	registry Registry,
	scheduler *trackingScheduler,
) *serviceImpl {
	t.Helper()

	if registry == nil {
		registry = NewRegistry()
	}
	if scheduler == nil {
		scheduler = &trackingScheduler{}
	}
	return newTestServiceWithExplicitDependencies(t, registry, scheduler)
}

// newTestServiceWithExplicitDependencies constructs job-management tests
// through the same explicit dependency path used by HTTP startup.
func newTestServiceWithExplicitDependencies(
	t *testing.T,
	registry Registry,
	scheduler Scheduler,
) *serviceImpl {
	t.Helper()

	var (
		bizCtxSvc = bizctx.New()
		configSvc = hostconfig.New(nil)
		i18nSvc   = i18nsvc.New(bizCtxSvc, configSvc, cachecoord.New(nil, nil))
		orgCapSvc = orgspi.New(nil, nil, nil)
		tenantSvc = tenantspi.New(nil, nil, nil, bizCtxSvc)
		roleSvc   = role.New(nil, bizCtxSvc, configSvc, i18nSvc, orgCapSvc, tenantSvc, cachecoord.New(nil, nil))
		scopeSvc  = datascope.New(bizCtxSvc, roleSvc, orgCapSvc.Scope())
	)
	roleSvc.SetDataScopeService(scopeSvc)
	return New(bizCtxSvc, configSvc, i18nSvc, registry, scheduler, scopeSvc).(*serviceImpl)
}

// setJobMgmtTestBizCtx replaces the context dependency and refreshes the
// derived data-scope service used by scheduled-job tests.
func setJobMgmtTestBizCtx(svc *serviceImpl, bizCtxSvc bizctx.Service) {
	svc.bizCtxSvc = bizCtxSvc
	configSvc := svc.configSvc
	if configSvc == nil {
		configSvc = hostconfig.New(nil)
	}
	var (
		i18nSvc   = i18nsvc.New(bizCtxSvc, configSvc, cachecoord.New(nil, nil))
		orgCapSvc = orgspi.New(nil, nil, nil)
		tenantSvc = tenantspi.New(nil, nil, nil, bizCtxSvc)
		roleSvc   = role.New(nil, bizCtxSvc, configSvc, i18nSvc, orgCapSvc, tenantSvc, cachecoord.New(nil, nil))
		scopeSvc  = datascope.New(bizCtxSvc, roleSvc, orgCapSvc.Scope())
	)
	roleSvc.SetDataScopeService(scopeSvc)
	svc.scopeSvc = scopeSvc
}

// defaultGroupID resolves the current tenant's default job group ID for tests.
func defaultGroupID(t *testing.T, ctx context.Context) int64 {
	t.Helper()

	var group *entity.SysJobGroup
	if err := dao.SysJobGroup.Ctx(ctx).
		Where(do.SysJobGroup{TenantId: datascope.CurrentTenantID(ctx), IsDefault: 1}).
		Scan(&group); err != nil {
		t.Fatalf("expected default job group query to succeed, got error: %v", err)
	}
	if group == nil {
		t.Fatal("expected default scheduled job group to exist")
	}
	return group.Id
}

// cleanupJobHard removes one job and its logs using hard-delete semantics.
func cleanupJobHard(t *testing.T, ctx context.Context, jobID int64) {
	t.Helper()
	if jobID == 0 {
		return
	}
	if _, err := dao.SysJobLog.Ctx(ctx).Where(do.SysJobLog{JobId: jobID}).Delete(); err != nil {
		t.Fatalf("expected job-log cleanup to succeed, got error: %v", err)
	}
	if _, err := dao.SysJob.Ctx(ctx).Unscoped().Where(do.SysJob{Id: jobID}).Delete(); err != nil {
		t.Fatalf("expected job cleanup to succeed, got error: %v", err)
	}
}

// cleanupGroupHard removes one group using hard-delete semantics.
func cleanupGroupHard(t *testing.T, ctx context.Context, groupID int64) {
	t.Helper()
	if groupID == 0 {
		return
	}
	if _, err := dao.SysJobGroup.Ctx(ctx).Unscoped().Where(do.SysJobGroup{Id: groupID}).Delete(); err != nil {
		t.Fatalf("expected group cleanup to succeed, got error: %v", err)
	}
}

// decodeJobParams converts one persisted params JSON string back to a map for tests.
func decodeJobParams(raw string) map[string]any {
	if raw == "" {
		return map[string]any{}
	}
	var result map[string]any
	if err := json.Unmarshal([]byte(raw), &result); err != nil {
		panic(fmt.Sprintf("invalid persisted job params JSON: %v", err))
	}
	return result
}

// retentionOverrideFromJob converts one persisted override JSON string to a typed option for tests.
func retentionOverrideFromJob(raw string) *jobmeta.RetentionOption {
	option, err := jobmeta.ParseRetentionOption(raw)
	if err != nil {
		panic(fmt.Sprintf("invalid persisted job retention override JSON: %v", err))
	}
	return option
}

// syncBuiltinHandlerJob projects one handler-based code-owned job into sys_job
// and returns the persisted row ID for assertions.
func syncBuiltinHandlerJob(
	t *testing.T,
	ctx context.Context,
	svc *serviceImpl,
	def BuiltinJobDef,
) int64 {
	t.Helper()

	if svc == nil {
		t.Fatal("expected service to be initialized")
	}
	if _, err := svc.SyncBuiltinJobs(ctx, []BuiltinJobDef{def}); err != nil {
		t.Fatalf("expected builtin job sync to succeed, got error: %v", err)
	}

	var job *entity.SysJob
	if err := dao.SysJob.Ctx(ctx).
		Where(do.SysJob{IsBuiltin: 1, HandlerRef: def.HandlerRef}).
		Scan(&job); err != nil {
		t.Fatalf("expected builtin job query to succeed, got error: %v", err)
	}
	if job == nil || job.Id == 0 {
		t.Fatalf("expected builtin job %s to exist after sync", def.HandlerRef)
	}
	return job.Id
}

// uniqueTestName returns one collision-resistant identifier for DB-backed tests.
func uniqueTestName(prefix string) string {
	return fmt.Sprintf("%s-%d", prefix, time.Now().UnixNano())
}

// assertBusinessCode verifies that err carries the expected structured
// business error code.
func assertBusinessCode(t *testing.T, err error, code *bizerr.Code) {
	t.Helper()
	if err == nil {
		t.Fatal("expected structured business error")
	}
	actual, ok := bizerr.As(err)
	if !ok {
		t.Fatalf("expected structured business error, got %v", err)
	}
	if !actual.Matches(code) {
		t.Fatalf("expected business code %s, got %s", code.RuntimeCode(), actual.RuntimeCode())
	}
}

// TestDeleteGroupsMigratesJobsToDefault verifies non-default group deletion migrates jobs to the default group.
func TestDeleteGroupsMigratesJobsToDefault(t *testing.T) {
	var (
		ctx       = context.Background()
		svc       = newTestService(t)
		defaultID = defaultGroupID(t, ctx)
		groupID   int64
		jobID     int64
		groupCode = uniqueTestName("test-job-group")
		groupName = uniqueTestName("测试任务分组")
		jobName   = uniqueTestName("测试任务")
	)

	groupID, err := svc.CreateGroup(ctx, SaveGroupInput{
		Code: groupCode,
		Name: groupName,
	})
	if err != nil {
		t.Fatalf("expected group create to succeed, got error: %v", err)
	}
	t.Cleanup(func() { cleanupGroupHard(t, ctx, groupID) })

	insertedJobID, err := dao.SysJob.Ctx(ctx).Data(do.SysJob{
		GroupId:        groupID,
		Name:           jobName,
		Description:    "Temporary job used to verify group deletion migration.",
		TaskType:       jobv1.TaskTypeShell,
		TimeoutSeconds: int64((5 * time.Minute).Seconds()),
		ShellCmd:       "printf 'group-migration'",
		CronExpr:       "*/5 * * * *",
		Timezone:       "Asia/Shanghai",
		Scope:          jobv1.ScopeMasterOnly,
		Concurrency:    jobv1.ConcurrencySingleton,
		MaxConcurrency: 1,
		MaxExecutions:  0,
		ExecutedCount:  0,
		StopReason:     "",
		Status:         jobv1.StatusDisabled,
		IsBuiltin:      0,
		SeedVersion:    0,
		CreatedBy:      0,
		UpdatedBy:      0,
	}).InsertAndGetId()
	if err != nil {
		t.Fatalf("expected job fixture insert to succeed, got error: %v", err)
	}
	jobID = int64(insertedJobID)
	t.Cleanup(func() { cleanupJobHard(t, ctx, jobID) })

	if err = svc.DeleteGroups(ctx, []int64{groupID}); err != nil {
		t.Fatalf("expected group delete to succeed, got error: %v", err)
	}

	var jobRow *entity.SysJob
	if err = dao.SysJob.Ctx(ctx).Where(do.SysJob{Id: jobID}).Scan(&jobRow); err != nil {
		t.Fatalf("expected migrated job query to succeed, got error: %v", err)
	}
	if jobRow == nil {
		t.Fatal("expected migrated job to remain present after group deletion")
	}
	if jobRow.GroupId != defaultID {
		t.Fatalf("expected migrated job group_id=%d, got %d", defaultID, jobRow.GroupId)
	}
}

// TestUpdateBuiltInJobRejectsLockedFields verifies built-in job immutable fields stay protected.
func TestUpdateBuiltInJobRejectsLockedFields(t *testing.T) {
	var (
		ctx = context.Background()
		svc = newTestService(t)
		job *entity.SysJob
	)

	jobID := syncBuiltinHandlerJob(t, ctx, svc, BuiltinJobDef{
		GroupCode:      "default",
		Name:           uniqueTestName("builtin-locked"),
		Description:    "Temporary built-in used to verify immutable field protection.",
		TaskType:       jobv1.TaskTypeHandler,
		HandlerRef:     uniqueTestName("host:builtin-locked"),
		Params:         map[string]any{},
		Timeout:        5 * time.Minute,
		Pattern:        "# 17 3 * * *",
		Timezone:       "Asia/Shanghai",
		Scope:          jobv1.ScopeMasterOnly,
		Concurrency:    jobv1.ConcurrencySingleton,
		MaxConcurrency: 1,
		MaxExecutions:  0,
		Status:         jobv1.StatusEnabled,
	})
	defer cleanupJobHard(t, ctx, jobID)

	if err := dao.SysJob.Ctx(ctx).
		Where(do.SysJob{Id: jobID}).
		Scan(&job); err != nil {
		t.Fatalf("expected built-in job query to succeed, got error: %v", err)
	}
	if job == nil {
		t.Fatal("expected synced built-in job to exist")
	}

	err := svc.UpdateJob(ctx, UpdateJobInput{
		ID: job.Id,
		SaveJobInput: SaveJobInput{
			GroupID:              job.GroupId,
			Name:                 job.Name,
			Description:          job.Description,
			TaskType:             jobmeta.NormalizeTaskType(job.TaskType),
			HandlerRef:           "host:another-handler",
			Params:               decodeJobParams(job.Params),
			Timeout:              time.Duration(job.TimeoutSeconds) * time.Second,
			CronExpr:             job.CronExpr,
			Timezone:             job.Timezone,
			Scope:                jobmeta.NormalizeJobScope(job.Scope),
			Concurrency:          jobmeta.NormalizeJobConcurrency(job.Concurrency),
			MaxConcurrency:       job.MaxConcurrency,
			MaxExecutions:        job.MaxExecutions,
			Status:               jobmeta.NormalizeJobStatus(job.Status),
			LogRetentionOverride: retentionOverrideFromJob(job.LogRetentionOverride),
		},
	})
	if err == nil {
		t.Fatal("expected built-in job update to be rejected")
	}
	assertBusinessCode(t, err, CodeJobBuiltinUpdateDenied)
}

// TestCreateJobValidatesTimeoutAndConcurrency verifies core runtime validation rejects invalid settings.
func TestCreateJobValidatesTimeoutAndConcurrency(t *testing.T) {
	var (
		ctx       = context.Background()
		svc       = newTestService(t)
		defaultID = defaultGroupID(t, ctx)
	)

	_, err := svc.CreateJob(ctx, SaveJobInput{
		GroupID:        defaultID,
		Name:           uniqueTestName("invalid-timeout"),
		TaskType:       jobv1.TaskTypeShell,
		Timeout:        0,
		ShellCmd:       "printf 'timeout'",
		CronExpr:       "*/5 * * * *",
		Timezone:       "Asia/Shanghai",
		Scope:          jobv1.ScopeMasterOnly,
		Concurrency:    jobv1.ConcurrencySingleton,
		MaxConcurrency: 1,
		Status:         jobv1.StatusDisabled,
	})
	if err == nil {
		t.Fatal("expected zero timeout to fail validation")
	}

	_, err = svc.CreateJob(ctx, SaveJobInput{
		GroupID:        defaultID,
		Name:           uniqueTestName("invalid-concurrency"),
		TaskType:       jobv1.TaskTypeShell,
		Timeout:        5 * time.Minute,
		ShellCmd:       "printf 'concurrency'",
		CronExpr:       "*/5 * * * *",
		Timezone:       "Asia/Shanghai",
		Scope:          jobv1.ScopeMasterOnly,
		Concurrency:    jobv1.ConcurrencyParallel,
		MaxConcurrency: 0,
		Status:         jobv1.StatusDisabled,
	})
	if err == nil {
		t.Fatal("expected zero maxConcurrency to fail validation")
	}
}

// TestCreateJobRejectsInvalidCronAndManagedStatus verifies save-time validation
// rejects unsupported cron formats, managed status values, and invalid runtime fields.
func TestCreateJobRejectsInvalidCronAndManagedStatus(t *testing.T) {
	var (
		ctx       = context.Background()
		svc       = newTestService(t)
		defaultID = defaultGroupID(t, ctx)
	)

	testCases := []struct {
		name     string
		input    SaveJobInput
		wantCode *bizerr.Code
	}{
		{
			name: "unsupported cron field count",
			input: SaveJobInput{
				GroupID:        defaultID,
				Name:           uniqueTestName("invalid-cron-count"),
				TaskType:       jobv1.TaskTypeShell,
				Timeout:        5 * time.Minute,
				ShellCmd:       "printf 'cron-count'",
				CronExpr:       "* * * *",
				Timezone:       "Asia/Shanghai",
				Scope:          jobv1.ScopeMasterOnly,
				Concurrency:    jobv1.ConcurrencySingleton,
				MaxConcurrency: 1,
				Status:         jobv1.StatusDisabled,
			},
			wantCode: CodeJobCronFieldCountInvalid,
		},
		{
			name: "manual hash seconds placeholder",
			input: SaveJobInput{
				GroupID:        defaultID,
				Name:           uniqueTestName("invalid-cron-hash"),
				TaskType:       jobv1.TaskTypeShell,
				Timeout:        5 * time.Minute,
				ShellCmd:       "printf 'cron-hash'",
				CronExpr:       "# 17 3 * * *",
				Timezone:       "Asia/Shanghai",
				Scope:          jobv1.ScopeMasterOnly,
				Concurrency:    jobv1.ConcurrencySingleton,
				MaxConcurrency: 1,
				Status:         jobv1.StatusDisabled,
			},
			wantCode: CodeJobCronSecondsRequired,
		},
		{
			name: "timezone must be valid",
			input: SaveJobInput{
				GroupID:        defaultID,
				Name:           uniqueTestName("invalid-timezone"),
				TaskType:       jobv1.TaskTypeShell,
				Timeout:        5 * time.Minute,
				ShellCmd:       "printf 'timezone'",
				CronExpr:       "*/5 * * * *",
				Timezone:       "Mars/Phobos",
				Scope:          jobv1.ScopeMasterOnly,
				Concurrency:    jobv1.ConcurrencySingleton,
				MaxConcurrency: 1,
				Status:         jobv1.StatusDisabled,
			},
			wantCode: CodeJobTimezoneInvalid,
		},
		{
			name: "status is system managed",
			input: SaveJobInput{
				GroupID:        defaultID,
				Name:           uniqueTestName("invalid-status"),
				TaskType:       jobv1.TaskTypeShell,
				Timeout:        5 * time.Minute,
				ShellCmd:       "printf 'status'",
				CronExpr:       "*/5 * * * *",
				Timezone:       "Asia/Shanghai",
				Scope:          jobv1.ScopeMasterOnly,
				Concurrency:    jobv1.ConcurrencySingleton,
				MaxConcurrency: 1,
				Status:         jobv1.StatusPausedByPlugin,
			},
			wantCode: CodeJobStatusInvalid,
		},
		{
			name: "timeout must use whole seconds",
			input: SaveJobInput{
				GroupID:        defaultID,
				Name:           uniqueTestName("invalid-timeout-seconds"),
				TaskType:       jobv1.TaskTypeShell,
				Timeout:        1500 * time.Millisecond,
				ShellCmd:       "printf 'timeout-seconds'",
				CronExpr:       "*/5 * * * *",
				Timezone:       "Asia/Shanghai",
				Scope:          jobv1.ScopeMasterOnly,
				Concurrency:    jobv1.ConcurrencySingleton,
				MaxConcurrency: 1,
				Status:         jobv1.StatusDisabled,
			},
			wantCode: CodeJobTimeoutSecondAlignedRequired,
		},
		{
			name: "parallel max concurrency upper bound",
			input: SaveJobInput{
				GroupID:        defaultID,
				Name:           uniqueTestName("invalid-max-concurrency"),
				TaskType:       jobv1.TaskTypeShell,
				Timeout:        5 * time.Minute,
				ShellCmd:       "printf 'max-concurrency'",
				CronExpr:       "*/5 * * * *",
				Timezone:       "Asia/Shanghai",
				Scope:          jobv1.ScopeMasterOnly,
				Concurrency:    jobv1.ConcurrencyParallel,
				MaxConcurrency: 101,
				Status:         jobv1.StatusDisabled,
			},
			wantCode: CodeJobMaxConcurrencyInvalid,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := svc.CreateJob(ctx, tc.input)
			if err == nil {
				t.Fatalf("expected CreateJob to reject %s", tc.name)
			}
			assertBusinessCode(t, err, tc.wantCode)
		})
	}
}

// TestPreviewCronSupportsFiveFieldAndTimezone verifies cron preview accepts 5-field expressions and applies the requested timezone.
func TestPreviewCronSupportsFiveFieldAndTimezone(t *testing.T) {
	var (
		ctx = context.Background()
		svc = newTestService(t)
	)

	times, err := svc.PreviewCron(ctx, "17 3 * * *", "UTC")
	if err != nil {
		t.Fatalf("expected cron preview to succeed, got error: %v", err)
	}
	if len(times) != 5 {
		t.Fatalf("expected 5 preview times, got %d", len(times))
	}
	for i, item := range times {
		if got := item.Location().String(); got != "UTC" {
			t.Fatalf("expected preview time %d to use UTC, got %s", i, got)
		}
		if item.Minute() != 17 || item.Hour() != 3 || item.Second() != 0 {
			t.Fatalf("expected preview time %d to be 03:17:00 UTC, got %s", i, item.Format(time.RFC3339))
		}
		if i > 0 && !item.After(times[i-1]) {
			t.Fatalf("expected preview times to be strictly increasing, got %s then %s", times[i-1], item)
		}
	}
}

// TestPreviewCronRejectsInvalidFormats verifies preview shares the strict cron validation rules.
func TestPreviewCronRejectsInvalidFormats(t *testing.T) {
	var (
		ctx = context.Background()
		svc = newTestService(t)
	)

	testCases := []struct {
		expr     string
		timezone string
		wantCode *bizerr.Code
	}{
		{
			expr:     "* * * *",
			timezone: "UTC",
			wantCode: CodeJobCronFieldCountInvalid,
		},
		{
			expr:     "# 17 3 * * *",
			timezone: "UTC",
			wantCode: CodeJobCronSecondsRequired,
		},
		{
			expr:     "17 3 * * *",
			timezone: "Invalid/Timezone",
			wantCode: CodeJobTimezoneInvalid,
		},
	}

	for _, tc := range testCases {
		_, err := svc.PreviewCron(ctx, tc.expr, tc.timezone)
		if err == nil {
			t.Fatalf("expected PreviewCron(%q, %q) to fail", tc.expr, tc.timezone)
		}
		assertBusinessCode(t, err, tc.wantCode)
	}
}
