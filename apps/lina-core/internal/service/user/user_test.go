// This file verifies user deletion transaction and batch-delete protections.

package user

import (
	"context"
	"errors"
	"fmt"
	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/net/ghttp"
	"lina-core/internal/dao"
	"lina-core/internal/model"
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
	"lina-core/internal/service/auth"
	"lina-core/internal/service/bizctx"
	"lina-core/internal/service/cachecoord"
	"lina-core/internal/service/cluster"
	hostconfig "lina-core/internal/service/config"
	"lina-core/internal/service/datascope"
	i18nsvc "lina-core/internal/service/i18n"
	"lina-core/internal/service/kvcache"
	"lina-core/internal/service/role"
	"lina-core/internal/service/session"
	"lina-core/pkg/bizerr"
	_ "lina-core/pkg/dbdriver"
	"lina-core/pkg/plugin/capability/bizctxcap"
	"lina-core/pkg/plugin/capability/capmodel"
	"lina-core/pkg/plugin/capability/orgcap"
	"lina-core/pkg/plugin/capability/orgcap/orgspi"
	"lina-core/pkg/plugin/capability/tenantcap/tenantspi"
	"lina-core/pkg/plugin/pluginhost"
	"testing"
	"time"
)

// TestDeleteRollsBackWhenOrgCleanupFails verifies user soft deletion is
// rolled back when organization cleanup reports an error inside the transaction.
func TestDeleteRollsBackWhenOrgCleanupFails(t *testing.T) {
	ctx := context.Background()
	userID := insertUserDeleteTestUser(t, ctx, "delete-rollback")
	t.Cleanup(func() {
		cleanupUserDeleteTestRows(t, ctx, []int{userID})
	})

	expectedErr := errors.New("cleanup failed")
	svc := newUserTestService().(*serviceImpl)
	setUserTestOrgCap(svc, userDeleteFailingOrgCap{cleanupErr: expectedErr})
	setUserTestBizCtx(svc, userDeleteStaticBizCtx{ctx: &model.Context{UserId: mustQueryBuiltinAdminUserID(t, ctx)}})

	err := svc.Delete(ctx, userID)
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected cleanup error, got %v", err)
	}
	if count := mustCountUser(t, ctx, userID); count != 1 {
		t.Fatalf("expected user soft delete to be rolled back, visible count=%d", count)
	}
}

// TestBatchDeleteRejectsCurrentUserAtomically verifies current-user protection
// rejects the whole batch before deleting any selected users.
func TestBatchDeleteRejectsCurrentUserAtomically(t *testing.T) {
	var (
		ctx           = context.Background()
		currentUserID = insertUserDeleteTestUser(t, ctx, "current-user")
		otherUserID   = insertUserDeleteTestUser(t, ctx, "other-user")
		roleID        = insertUserDeleteTestRole(t, ctx, "current-user-role")
	)
	t.Cleanup(func() {
		cleanupUserDeleteTestRows(t, ctx, []int{currentUserID, otherUserID})
		cleanupUserDeleteTestRoles(t, ctx, []int{roleID})
	})
	insertUserDeleteTestUserRole(t, ctx, currentUserID, roleID)

	svc := newUserTestService().(*serviceImpl)
	setUserTestBizCtx(svc, userDeleteStaticBizCtx{ctx: &model.Context{UserId: currentUserID}})

	err := svc.BatchDelete(ctx, []int{otherUserID, currentUserID})
	if err == nil {
		t.Fatal("expected current user batch delete to be rejected")
	}
	messageErr, ok := bizerr.As(err)
	if !ok || !messageErr.Matches(CodeUserCurrentDeleteDenied) {
		t.Fatalf("expected CodeUserCurrentDeleteDenied, got %v", err)
	}
	if count := mustCountUser(t, ctx, otherUserID); count != 1 {
		t.Fatalf("expected other user to remain visible after rejected batch, count=%d", count)
	}
}

// TestBatchDeleteRemovesUsersAndAssociations verifies batch deletion soft
// deletes users and clears user-role associations in one service call.
func TestBatchDeleteRemovesUsersAndAssociations(t *testing.T) {
	ctx := context.Background()
	userIDs := []int{
		insertUserDeleteTestUser(t, ctx, "batch-delete-a"),
		insertUserDeleteTestUser(t, ctx, "batch-delete-b"),
	}
	roleID := insertUserDeleteTestRole(t, ctx, "batch-delete-role")
	t.Cleanup(func() {
		cleanupUserDeleteTestRows(t, ctx, userIDs)
		cleanupUserDeleteTestRoles(t, ctx, []int{roleID})
	})

	for _, userID := range userIDs {
		if _, err := dao.SysUserRole.Ctx(ctx).Data(do.SysUserRole{
			UserId: userID,
			RoleId: roleID,
		}).Insert(); err != nil {
			t.Fatalf("insert user-role relation: %v", err)
		}
	}

	svc := newUserTestService().(*serviceImpl)
	setUserTestBizCtx(svc, userDeleteStaticBizCtx{ctx: &model.Context{UserId: mustQueryBuiltinAdminUserID(t, ctx)}})
	if err := svc.BatchDelete(ctx, userIDs); err != nil {
		t.Fatalf("batch delete users: %v", err)
	}
	for _, userID := range userIDs {
		if count := mustCountUser(t, ctx, userID); count != 0 {
			t.Fatalf("expected user %d to be soft-deleted, visible count=%d", userID, count)
		}
		if count := mustCountUserRoles(t, ctx, userID); count != 0 {
			t.Fatalf("expected user-role rows for user %d to be deleted, count=%d", userID, count)
		}
	}
}

// TestBatchDeleteRejectsBuiltinAdminAtomically verifies built-in administrator
// protection rejects the whole batch before deleting any selected users.
func TestBatchDeleteRejectsBuiltinAdminAtomically(t *testing.T) {
	var (
		ctx         = context.Background()
		otherUserID = insertUserDeleteTestUser(t, ctx, "other-admin-guard")
		adminUserID = mustQueryBuiltinAdminUserID(t, ctx)
	)
	t.Cleanup(func() {
		cleanupUserDeleteTestRows(t, ctx, []int{otherUserID})
	})

	svc := newUserTestService().(*serviceImpl)
	setUserTestBizCtx(svc, userDeleteStaticBizCtx{ctx: &model.Context{UserId: adminUserID}})
	err := svc.BatchDelete(ctx, []int{otherUserID, adminUserID})
	if err == nil {
		t.Fatal("expected builtin admin batch delete to be rejected")
	}
	messageErr, ok := bizerr.As(err)
	if !ok || !messageErr.Matches(CodeUserBuiltinAdminDeleteDenied) {
		t.Fatalf("expected CodeUserBuiltinAdminDeleteDenied, got %v", err)
	}
	if count := mustCountUser(t, ctx, otherUserID); count != 1 {
		t.Fatalf("expected other user to remain visible after rejected batch, count=%d", count)
	}
}

// TestBatchDeleteRejectsEmptyList verifies empty batch deletes return a stable
// bizerr code before touching the database.
func TestBatchDeleteRejectsEmptyList(t *testing.T) {
	err := newUserTestService().BatchDelete(context.Background(), nil)
	if err == nil {
		t.Fatal("expected empty batch delete to be rejected")
	}
	messageErr, ok := bizerr.As(err)
	if !ok || !messageErr.Matches(CodeUserDeleteIdsRequired) {
		t.Fatalf("expected CodeUserDeleteIdsRequired, got %v", err)
	}
}

// insertUserDeleteTestUser inserts a temporary user row for delete tests.
func insertUserDeleteTestUser(t *testing.T, ctx context.Context, label string) int {
	t.Helper()

	username := fmt.Sprintf("%s-%d", label, time.Now().UnixNano())
	id, err := dao.SysUser.Ctx(ctx).Data(do.SysUser{
		Username: username,
		Password: "test-password-hash",
		Nickname: username,
		Status:   1,
	}).InsertAndGetId()
	if err != nil {
		t.Fatalf("insert test user: %v", err)
	}
	return int(id)
}

// insertUserDeleteTestRole inserts one temporary role row for user deletion tests.
func insertUserDeleteTestRole(t *testing.T, ctx context.Context, label string) int {
	t.Helper()

	suffix := time.Now().UnixNano()
	id, err := dao.SysRole.Ctx(ctx).Data(do.SysRole{
		Name:      fmt.Sprintf("%s-%d", label, suffix),
		Key:       fmt.Sprintf("%s-%d", label, suffix),
		Sort:      99,
		DataScope: int(userDataScopeAll),
		Status:    1,
	}).InsertAndGetId()
	if err != nil {
		t.Fatalf("insert test role: %v", err)
	}
	return int(id)
}

// insertUserDeleteTestUserRole inserts one temporary user-role relation.
func insertUserDeleteTestUserRole(t *testing.T, ctx context.Context, userID int, roleID int) {
	t.Helper()

	if _, err := dao.SysUserRole.Ctx(ctx).Data(do.SysUserRole{
		UserId: userID,
		RoleId: roleID,
	}).Insert(); err != nil {
		t.Fatalf("insert test user-role relation: %v", err)
	}
}

// cleanupUserDeleteTestRows removes temporary user rows and their role bindings.
func cleanupUserDeleteTestRows(t *testing.T, ctx context.Context, userIDs []int) {
	t.Helper()

	if _, err := dao.SysUserRole.Ctx(ctx).WhereIn(dao.SysUserRole.Columns().UserId, userIDs).Delete(); err != nil {
		t.Fatalf("cleanup user-role rows: %v", err)
	}
	if _, err := dao.SysUser.Ctx(ctx).Unscoped().WhereIn(dao.SysUser.Columns().Id, userIDs).Delete(); err != nil {
		t.Fatalf("cleanup user rows: %v", err)
	}
}

// cleanupUserDeleteTestRoles removes temporary role rows and their bindings.
func cleanupUserDeleteTestRoles(t *testing.T, ctx context.Context, roleIDs []int) {
	t.Helper()

	if _, err := dao.SysUserRole.Ctx(ctx).WhereIn(dao.SysUserRole.Columns().RoleId, roleIDs).Delete(); err != nil {
		t.Fatalf("cleanup user-role role rows: %v", err)
	}
	if _, err := dao.SysRole.Ctx(ctx).Unscoped().WhereIn(dao.SysRole.Columns().Id, roleIDs).Delete(); err != nil {
		t.Fatalf("cleanup role rows: %v", err)
	}
}

// mustCountUser counts visible user rows for one user ID.
func mustCountUser(t *testing.T, ctx context.Context, userID int) int {
	t.Helper()

	count, err := dao.SysUser.Ctx(ctx).Where(do.SysUser{Id: userID}).Count()
	if err != nil {
		t.Fatalf("count user: %v", err)
	}
	return count
}

// mustCountUserRoles counts user-role rows for one user ID.
func mustCountUserRoles(t *testing.T, ctx context.Context, userID int) int {
	t.Helper()

	count, err := dao.SysUserRole.Ctx(ctx).Where(do.SysUserRole{UserId: userID}).Count()
	if err != nil {
		t.Fatalf("count user-role rows: %v", err)
	}
	return count
}

// mustQueryBuiltinAdminUserID resolves the built-in admin user ID for tests.
func mustQueryBuiltinAdminUserID(t *testing.T, ctx context.Context) int {
	t.Helper()

	var adminUser *entity.SysUser
	if err := dao.SysUser.Ctx(ctx).Where(do.SysUser{Username: "admin"}).Scan(&adminUser); err != nil {
		t.Fatalf("query built-in admin user: %v", err)
	}
	if adminUser == nil {
		t.Fatal("expected built-in admin user to exist")
	}
	return adminUser.Id
}

// userDeleteStaticBizCtx returns a fixed business context for current-user tests.
type userDeleteStaticBizCtx struct {
	ctx *model.Context
}

// Init is unused by service tests because they inject context directly.
func (s userDeleteStaticBizCtx) Init(_ *ghttp.Request, _ *model.Context) {}

// Get returns the configured business context.
func (s userDeleteStaticBizCtx) Get(context.Context) *model.Context {
	return s.ctx
}

// Current returns the plugin-visible business context projection.
func (s userDeleteStaticBizCtx) Current(context.Context) bizctxcap.CurrentContext {
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

// SetLocale is unused by delete tests.
func (s userDeleteStaticBizCtx) SetLocale(context.Context, string) {}

// SetUser is unused by delete tests.
func (s userDeleteStaticBizCtx) SetUser(context.Context, string, int, string, int, string) {}

// SetTenant is unused by delete tests.
func (s userDeleteStaticBizCtx) SetTenant(context.Context, int) {}

// SetImpersonation is unused by delete tests.
func (s userDeleteStaticBizCtx) SetImpersonation(context.Context, int, int, bool, bool) {}

// SetUserAccess is unused by delete tests.
func (s userDeleteStaticBizCtx) SetUserAccess(context.Context, int, bool, int) {}

// userDeleteFailingOrgCap fails cleanup while otherwise behaving as disabled.
type userDeleteFailingOrgCap struct {
	cleanupErr error
}

// Available reports the optional organization capability provider as unavailable.
func (f userDeleteFailingOrgCap) Available(context.Context) bool { return false }

// Status returns an unavailable organization capability status.
func (f userDeleteFailingOrgCap) Status(context.Context) capmodel.CapabilityStatus {
	return capmodel.CapabilityStatus{}
}

// ListUserDeptAssignments returns no organization assignments.
func (f userDeleteFailingOrgCap) ListUserDeptAssignments(context.Context, []int) (map[int]*orgcap.UserDeptAssignment, error) {
	return map[int]*orgcap.UserDeptAssignment{}, nil
}

// BatchGetUserOrgProfiles returns no organization profiles.
func (f userDeleteFailingOrgCap) BatchGetUserOrgProfiles(_ context.Context, userIDs []int) (*capmodel.BatchResult[*orgcap.UserOrgProfile, int], error) {
	return &capmodel.BatchResult[*orgcap.UserOrgProfile, int]{
		Items:      map[int]*orgcap.UserOrgProfile{},
		MissingIDs: append([]int(nil), userIDs...),
	}, nil
}

// GetUserDeptInfo returns an empty department projection.
func (f userDeleteFailingOrgCap) GetUserDeptInfo(context.Context, int) (int, string, error) {
	return 0, "", nil
}

// GetUserDeptName returns an empty department name.
func (f userDeleteFailingOrgCap) GetUserDeptName(context.Context, int) (string, error) {
	return "", nil
}

// GetUserDeptIDs returns no department IDs.
func (f userDeleteFailingOrgCap) GetUserDeptIDs(context.Context, int) ([]int, error) {
	return []int{}, nil
}

// ApplyUserDeptScope reports an empty department scope for delete tests.
func (f userDeleteFailingOrgCap) ApplyUserDeptScope(_ context.Context, model *gdb.Model, _ string, _ int) (*gdb.Model, bool, error) {
	return model, true, nil
}

// BuildUserDeptScopeExists reports an empty department scope for delete tests.
func (f userDeleteFailingOrgCap) BuildUserDeptScopeExists(context.Context, string, int) (*gdb.Model, bool, error) {
	return nil, true, nil
}

// ApplyUserDeptFilter reports an empty department filter for delete tests.
func (f userDeleteFailingOrgCap) ApplyUserDeptFilter(_ context.Context, model *gdb.Model, _ string, _ int) (*gdb.Model, bool, error) {
	return model, true, nil
}

// ApplyUserDeptUnassignedFilter leaves delete-test models unchanged.
func (f userDeleteFailingOrgCap) ApplyUserDeptUnassignedFilter(_ context.Context, model *gdb.Model, _ string) (*gdb.Model, bool, error) {
	return model, false, nil
}

// GetUserPostIDs returns no post IDs.
func (f userDeleteFailingOrgCap) GetUserPostIDs(context.Context, int) ([]int, error) {
	return []int{}, nil
}

// ListDeptTree returns no department tree nodes.
func (f userDeleteFailingOrgCap) ListDeptTree(context.Context, orgcap.DeptTreeInput) (*orgcap.DeptTreeResult, error) {
	return &orgcap.DeptTreeResult{Items: []*orgcap.DeptTreeNode{}}, nil
}

// SearchDepartments returns no department candidates.
func (f userDeleteFailingOrgCap) SearchDepartments(context.Context, orgcap.DeptListInput) (*capmodel.PageResult[*orgcap.DeptInfo], error) {
	return &capmodel.PageResult[*orgcap.DeptInfo]{Items: []*orgcap.DeptInfo{}}, nil
}

// ListPostOptionsPage returns no post candidates.
func (f userDeleteFailingOrgCap) ListPostOptionsPage(context.Context, orgcap.PostOptionsInput) (*capmodel.PageResult[*orgcap.PostOption], error) {
	return &capmodel.PageResult[*orgcap.PostOption]{Items: []*orgcap.PostOption{}}, nil
}

// EnsureDepartmentsVisible accepts visibility checks in delete tests.
func (f userDeleteFailingOrgCap) EnsureDepartmentsVisible(context.Context, []int) error {
	return nil
}

// EnsurePostsVisible accepts visibility checks in delete tests.
func (f userDeleteFailingOrgCap) EnsurePostsVisible(context.Context, []int) error {
	return nil
}

// ReplaceUserAssignments accepts assignment replacement without doing work.
func (f userDeleteFailingOrgCap) ReplaceUserAssignments(context.Context, int, *int, []int) error {
	return nil
}

// CleanupUserAssignments returns the configured cleanup error.
func (f userDeleteFailingOrgCap) CleanupUserAssignments(context.Context, int) error {
	return f.cleanupErr
}

// newUserTestService constructs user service tests through explicit dependencies.
func newUserTestService(tenantManagersAndRuntimes ...any) Service {
	var (
		bizCtxSvc     = bizctx.New()
		configSvc     = hostconfig.New()
		clusterSvc    = cluster.New(configSvc.GetCluster(context.Background()))
		cacheCoordSvc = cachecoord.Default(clusterSvc)
		i18nSvc       = i18nsvc.New(bizCtxSvc, configSvc, cacheCoordSvc)
		sessionStore  = session.NewDBStore()
		pluginRuntime = userTestPluginRuntime{}
		orgCapSvc     = orgspi.New(nil, pluginRuntime, pluginRuntime.OrgProviderEnv)
		tenantSvc     = tenantspi.New(nil, nil, nil, nil)
	)
	if len(tenantManagersAndRuntimes) > 0 {
		var (
			manager    *tenantspi.Manager
			enablement interface {
				IsProviderEnabled(context.Context, string) bool
			}
		)
		if value, ok := tenantManagersAndRuntimes[0].(*tenantspi.Manager); ok {
			manager = value
			if len(tenantManagersAndRuntimes) > 1 {
				enablement, _ = tenantManagersAndRuntimes[1].(interface {
					IsProviderEnabled(context.Context, string) bool
				})
			}
		} else {
			enablement, _ = tenantManagersAndRuntimes[0].(interface {
				IsProviderEnabled(context.Context, string) bool
			})
		}
		tenantSvc = tenantspi.New(manager, enablement, nil, nil)
	}
	roleSvc := role.New(pluginRuntime, bizCtxSvc, configSvc, i18nSvc, orgCapSvc, tenantSvc)
	scopeSvc := datascope.New(bizCtxSvc, roleSvc, orgCapSvc.Scope())
	roleSvc.SetDataScopeService(scopeSvc)
	var (
		kvCacheSvc = kvcache.New()
		authSvc    = auth.New(configSvc, pluginRuntime, orgCapSvc, roleSvc, tenantSvc, sessionStore, kvCacheSvc)
		userSvc    = New(authSvc, bizCtxSvc, i18nSvc, orgCapSvc, roleSvc, scopeSvc, tenantSvc)
	)
	return userSvc.(*serviceImpl)
}

// userTestPluginRuntime supplies the narrow plugin-facing seams needed by user
// service tests without importing the plugin service package back into user.
type userTestPluginRuntime struct{}

// DispatchHookEvent ignores plugin hooks in user tests.
func (userTestPluginRuntime) DispatchHookEvent(_ context.Context, _ pluginhost.ExtensionPoint, _ map[string]interface{}) error {
	return nil
}

// FilterPermissionMenus leaves permission menus unchanged in user tests.
func (userTestPluginRuntime) FilterPermissionMenus(_ context.Context, menus []*entity.SysMenu) []*entity.SysMenu {
	return menus
}

// IsProviderEnabled reports that no plugin provider is active in user tests.
func (userTestPluginRuntime) IsProviderEnabled(_ context.Context, _ string) bool {
	return false
}

// OrgProviderEnv returns a neutral organization provider environment.
func (userTestPluginRuntime) OrgProviderEnv(_ context.Context, pluginID string) orgspi.ProviderEnv {
	return orgspi.ProviderEnv{PluginID: pluginID}
}

// setUserTestBizCtx replaces the business context dependency and refreshes
// the derived data-scope service used by user-management tests.
func setUserTestBizCtx(svc *serviceImpl, bizCtxSvc bizctx.Service) {
	svc.bizCtxSvc = bizCtxSvc
	refreshUserTestScope(svc)
}

// setUserTestOrgCap replaces the organization capability dependency and
// refreshes the derived data-scope service used by user-management tests.
func setUserTestOrgCap(svc *serviceImpl, orgCapSvc any) {
	if service, ok := orgCapSvc.(orgspi.Service); ok {
		svc.orgCapSvc = service
	} else {
		var orgScope orgspi.ScopeService
		if scope, ok := orgCapSvc.(orgspi.ScopeService); ok {
			orgScope = scope
		}
		svc.orgCapSvc = userTestOrgService{
			userTestOrgCapDirectory: userTestOrgCapDirectory{legacy: orgCapSvc},
			scope:                   orgScope,
		}
	}
	refreshUserTestScope(svc)
}

// userTestOrgService adapts legacy flat organization fakes to the host-owned
// orgspi.Service contract used by the user service.
type userTestOrgService struct {
	userTestOrgCapDirectory
	scope orgspi.ScopeService
}

// Scope returns the optional organization data-scope fake.
func (s userTestOrgService) Scope() orgspi.ScopeService {
	return s.scope
}

// Workspace returns optional user-management workspace projections.
func (s userTestOrgService) Workspace() orgspi.WorkspaceViewService {
	return userTestOrgWorkspace{legacy: s.legacy}
}

// userTestOrgCapDirectory adapts older flat organization fakes to the current
// orgcap.Service subresource directory used by user service internals.
type userTestOrgCapDirectory struct {
	legacy any
}

// userTestOrgWorkspace adapts optional legacy workspace projection fakes.
type userTestOrgWorkspace struct {
	legacy any
}

// UserDeptTree returns a fake department tree for user-management tests.
func (w userTestOrgWorkspace) UserDeptTree(ctx context.Context) ([]*orgcap.DeptTreeNode, error) {
	if provider, ok := w.legacy.(interface {
		UserDeptTree(context.Context) ([]*orgcap.DeptTreeNode, error)
	}); ok {
		return provider.UserDeptTree(ctx)
	}
	return []*orgcap.DeptTreeNode{}, nil
}

// ListPostOptions returns fake post options for user-management tests.
func (w userTestOrgWorkspace) ListPostOptions(ctx context.Context, deptID *int) ([]*orgcap.PostOption, error) {
	if provider, ok := w.legacy.(interface {
		ListPostOptions(context.Context, *int) ([]*orgcap.PostOption, error)
	}); ok {
		return provider.ListPostOptions(ctx, deptID)
	}
	return []*orgcap.PostOption{}, nil
}

// Available reports whether the legacy fake organization capability is active.
func (d userTestOrgCapDirectory) Available(ctx context.Context) bool {
	if provider, ok := d.legacy.(interface {
		Available(context.Context) bool
	}); ok {
		return provider.Available(ctx)
	}
	return false
}

// Status returns the legacy fake organization status when available.
func (d userTestOrgCapDirectory) Status(ctx context.Context) capmodel.CapabilityStatus {
	if provider, ok := d.legacy.(interface {
		Status(context.Context) capmodel.CapabilityStatus
	}); ok {
		return provider.Status(ctx)
	}
	return capmodel.CapabilityStatus{}
}

// Department returns the department subresource adapter.
func (d userTestOrgCapDirectory) Department() orgcap.DepartmentService {
	return userTestOrgDepartment{legacy: d.legacy}
}

// Post returns the post subresource adapter.
func (d userTestOrgCapDirectory) Post() orgcap.PostService {
	return userTestOrgPost{legacy: d.legacy}
}

// Assignment returns the user assignment subresource adapter.
func (d userTestOrgCapDirectory) Assignment() orgcap.AssignmentService {
	return userTestOrgAssignment{legacy: d.legacy}
}

// userTestOrgDepartment adapts flat department fake methods.
type userTestOrgDepartment struct {
	legacy any
}

// Get returns no department projection in user tests.
func (d userTestOrgDepartment) Get(context.Context, int) (*orgcap.DeptInfo, error) {
	return nil, nil
}

// BatchGet returns all requested departments as missing in user tests.
func (d userTestOrgDepartment) BatchGet(_ context.Context, deptIDs []int) (*capmodel.BatchResult[*orgcap.DeptInfo, int], error) {
	return &capmodel.BatchResult[*orgcap.DeptInfo, int]{
		Items:      map[int]*orgcap.DeptInfo{},
		MissingIDs: append([]int(nil), deptIDs...),
	}, nil
}

// List returns bounded fake departments.
func (d userTestOrgDepartment) List(ctx context.Context, input orgcap.DeptListInput) (*capmodel.PageResult[*orgcap.DeptInfo], error) {
	if provider, ok := d.legacy.(interface {
		SearchDepartments(context.Context, orgcap.DeptListInput) (*capmodel.PageResult[*orgcap.DeptInfo], error)
	}); ok {
		return provider.SearchDepartments(ctx, input)
	}
	return &capmodel.PageResult[*orgcap.DeptInfo]{Items: []*orgcap.DeptInfo{}}, nil
}

// ListTree returns a fake department tree.
func (d userTestOrgDepartment) ListTree(ctx context.Context, input orgcap.DeptTreeInput) (*orgcap.DeptTreeResult, error) {
	if provider, ok := d.legacy.(interface {
		ListDeptTree(context.Context, orgcap.DeptTreeInput) (*orgcap.DeptTreeResult, error)
	}); ok {
		return provider.ListDeptTree(ctx, input)
	}
	return &orgcap.DeptTreeResult{Items: []*orgcap.DeptTreeNode{}}, nil
}

// ListOptions reuses the list adapter for user tests.
func (d userTestOrgDepartment) ListOptions(ctx context.Context, input orgcap.DeptOptionsInput) (*capmodel.PageResult[*orgcap.DeptInfo], error) {
	return d.List(ctx, orgcap.DeptListInput{
		Keyword: input.Keyword,
		Page:    input.Page,
	})
}

// EnsureVisible verifies department identifiers through the legacy fake.
func (d userTestOrgDepartment) EnsureVisible(ctx context.Context, deptIDs []int) error {
	if provider, ok := d.legacy.(interface {
		EnsureDepartmentsVisible(context.Context, []int) error
	}); ok {
		return provider.EnsureDepartmentsVisible(ctx, deptIDs)
	}
	return nil
}

// Create accepts department creation in user tests.
func (d userTestOrgDepartment) Create(context.Context, orgcap.DeptCreateInput) (int, error) {
	return 0, nil
}

// Update accepts department updates in user tests.
func (d userTestOrgDepartment) Update(context.Context, orgcap.DeptUpdateInput) error {
	return nil
}

// Delete accepts department deletion in user tests.
func (d userTestOrgDepartment) Delete(context.Context, int) error {
	return nil
}

// userTestOrgPost adapts flat post fake methods.
type userTestOrgPost struct {
	legacy any
}

// Get returns no post projection in user tests.
func (p userTestOrgPost) Get(context.Context, int) (*orgcap.PostInfo, error) {
	return nil, nil
}

// BatchGet returns all requested posts as missing in user tests.
func (p userTestOrgPost) BatchGet(_ context.Context, postIDs []int) (*capmodel.BatchResult[*orgcap.PostInfo, int], error) {
	return &capmodel.BatchResult[*orgcap.PostInfo, int]{
		Items:      map[int]*orgcap.PostInfo{},
		MissingIDs: append([]int(nil), postIDs...),
	}, nil
}

// List returns empty post projections in user tests.
func (p userTestOrgPost) List(context.Context, orgcap.PostListInput) (*capmodel.PageResult[*orgcap.PostInfo], error) {
	return &capmodel.PageResult[*orgcap.PostInfo]{Items: []*orgcap.PostInfo{}}, nil
}

// ListOptions returns fake post options.
func (p userTestOrgPost) ListOptions(ctx context.Context, input orgcap.PostOptionsInput) (*capmodel.PageResult[*orgcap.PostOption], error) {
	if provider, ok := p.legacy.(interface {
		ListPostOptionsPage(context.Context, orgcap.PostOptionsInput) (*capmodel.PageResult[*orgcap.PostOption], error)
	}); ok {
		return provider.ListPostOptionsPage(ctx, input)
	}
	return &capmodel.PageResult[*orgcap.PostOption]{Items: []*orgcap.PostOption{}}, nil
}

// EnsureVisible verifies post identifiers through the legacy fake.
func (p userTestOrgPost) EnsureVisible(ctx context.Context, postIDs []int) error {
	if provider, ok := p.legacy.(interface {
		EnsurePostsVisible(context.Context, []int) error
	}); ok {
		return provider.EnsurePostsVisible(ctx, postIDs)
	}
	return nil
}

// Create accepts post creation in user tests.
func (p userTestOrgPost) Create(context.Context, orgcap.PostCreateInput) (int, error) {
	return 0, nil
}

// Update accepts post updates in user tests.
func (p userTestOrgPost) Update(context.Context, orgcap.PostUpdateInput) error {
	return nil
}

// Delete accepts post deletion in user tests.
func (p userTestOrgPost) Delete(context.Context, int) error {
	return nil
}

// userTestOrgAssignment adapts flat assignment fake methods.
type userTestOrgAssignment struct {
	legacy any
}

// BatchGetUserProfiles returns fake user organization profiles.
func (a userTestOrgAssignment) BatchGetUserProfiles(ctx context.Context, userIDs []int) (*capmodel.BatchResult[*orgcap.UserOrgProfile, int], error) {
	if provider, ok := a.legacy.(interface {
		BatchGetUserOrgProfiles(context.Context, []int) (*capmodel.BatchResult[*orgcap.UserOrgProfile, int], error)
	}); ok {
		return provider.BatchGetUserOrgProfiles(ctx, userIDs)
	}
	return &capmodel.BatchResult[*orgcap.UserOrgProfile, int]{
		Items:      map[int]*orgcap.UserOrgProfile{},
		MissingIDs: append([]int(nil), userIDs...),
	}, nil
}

// ListByUser returns one fake user organization profile.
func (a userTestOrgAssignment) ListByUser(ctx context.Context, userID int) (*orgcap.UserOrgProfile, error) {
	result, err := a.BatchGetUserProfiles(ctx, []int{userID})
	if err != nil || result == nil {
		return nil, err
	}
	return result.Items[userID], nil
}

// BatchListByUsers returns fake department assignments.
func (a userTestOrgAssignment) BatchListByUsers(ctx context.Context, userIDs []int) (map[int]*orgcap.UserDeptAssignment, error) {
	if provider, ok := a.legacy.(interface {
		ListUserDeptAssignments(context.Context, []int) (map[int]*orgcap.UserDeptAssignment, error)
	}); ok {
		return provider.ListUserDeptAssignments(ctx, userIDs)
	}
	return map[int]*orgcap.UserDeptAssignment{}, nil
}

// GetUserDeptInfo returns one fake user department.
func (a userTestOrgAssignment) GetUserDeptInfo(ctx context.Context, userID int) (int, string, error) {
	if provider, ok := a.legacy.(interface {
		GetUserDeptInfo(context.Context, int) (int, string, error)
	}); ok {
		return provider.GetUserDeptInfo(ctx, userID)
	}
	return 0, "", nil
}

// GetUserDeptIDs returns fake user department IDs.
func (a userTestOrgAssignment) GetUserDeptIDs(ctx context.Context, userID int) ([]int, error) {
	if provider, ok := a.legacy.(interface {
		GetUserDeptIDs(context.Context, int) ([]int, error)
	}); ok {
		return provider.GetUserDeptIDs(ctx, userID)
	}
	return []int{}, nil
}

// GetUserPostIDs returns fake user post IDs.
func (a userTestOrgAssignment) GetUserPostIDs(ctx context.Context, userID int) ([]int, error) {
	if provider, ok := a.legacy.(interface {
		GetUserPostIDs(context.Context, int) ([]int, error)
	}); ok {
		return provider.GetUserPostIDs(ctx, userID)
	}
	return []int{}, nil
}

// ReplaceByUser delegates fake assignment replacement.
func (a userTestOrgAssignment) ReplaceByUser(ctx context.Context, userID int, deptID *int, postIDs []int) error {
	if provider, ok := a.legacy.(interface {
		ReplaceUserAssignments(context.Context, int, *int, []int) error
	}); ok {
		return provider.ReplaceUserAssignments(ctx, userID, deptID, postIDs)
	}
	return nil
}

// CleanupByUser delegates fake assignment cleanup.
func (a userTestOrgAssignment) CleanupByUser(ctx context.Context, userID int) error {
	if provider, ok := a.legacy.(interface {
		CleanupUserAssignments(context.Context, int) error
	}); ok {
		return provider.CleanupUserAssignments(ctx, userID)
	}
	return nil
}

// refreshUserTestScope rebuilds the stateless data-scope helper from the
// current explicit fake dependencies.
func refreshUserTestScope(svc *serviceImpl) {
	var orgScope orgspi.ScopeService
	if svc.orgCapSvc != nil {
		orgScope = svc.orgCapSvc.Scope()
	}
	svc.scopeSvc = datascope.New(svc.bizCtxSvc, svc.roleSvc, orgScope)
}
