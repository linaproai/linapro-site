// This file verifies admin-account permission bypass semantics backed by the database.

package role

import (
	"context"
	"fmt"
	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/net/ghttp"
	"lina-core/internal/dao"
	"lina-core/internal/model"
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
	"lina-core/internal/service/bizctx"
	"lina-core/internal/service/cachecoord"
	hostconfig "lina-core/internal/service/config"
	"lina-core/internal/service/datascope"
	i18nsvc "lina-core/internal/service/i18n"
	"lina-core/internal/service/user/accountpolicy"
	"lina-core/pkg/bizerr"
	_ "lina-core/pkg/dbdriver"
	"lina-core/pkg/plugin/capability/bizctxcap"
	"lina-core/pkg/plugin/capability/orgcap"
	"lina-core/pkg/plugin/capability/orgcap/orgspi"
	"lina-core/pkg/plugin/capability/tenantcap"
	"lina-core/pkg/plugin/capability/tenantcap/tenantspi"
	"lina-core/pkg/statusflag"
	"testing"
	"time"
)

// TestAdminUserRetainsBypassWithoutRoleBinding verifies the built-in admin
// account still receives full access even if its admin-role association is missing.
func TestAdminUserRetainsBypassWithoutRoleBinding(t *testing.T) {
	ctx := context.Background()
	svc := newDefaultRoleTestService()

	adminUserID, adminRoleID := mustQueryAdminUserAndRoleID(t, ctx)

	_, err := dao.SysUserRole.Ctx(ctx).
		Where(do.SysUserRole{UserId: adminUserID, RoleId: adminRoleID}).
		Delete()
	if err != nil {
		t.Fatalf("delete admin user-role binding: %v", err)
	}
	t.Cleanup(func() {
		if _, cleanupErr := dao.SysUserRole.Ctx(ctx).
			Data(do.SysUserRole{UserId: adminUserID, RoleId: adminRoleID}).
			Insert(); cleanupErr != nil {
			t.Fatalf("restore admin user-role binding: %v", cleanupErr)
		}
	})

	if !svc.IsSuperAdmin(ctx, adminUserID) {
		t.Fatal("expected built-in admin user to bypass super-admin checks without role binding")
	}

	accessContext, err := svc.GetUserAccessContext(ctx, adminUserID)
	if err != nil {
		t.Fatalf("load admin access context: %v", err)
	}
	if accessContext == nil || !accessContext.IsSuperAdmin {
		t.Fatal("expected admin access context to keep built-in admin bypass flag")
	}
	if accessContext.DataScope != datascope.ScopeAll {
		t.Fatalf("expected admin access context to carry all-data scope, got %d", accessContext.DataScope)
	}

	menuIDs, err := svc.GetUserMenuIds(ctx, adminUserID)
	if err != nil {
		t.Fatalf("load admin menu ids: %v", err)
	}
	if len(menuIDs) == 0 {
		t.Fatal("expected built-in admin user to receive enabled menu ids")
	}

	permissions, err := svc.GetUserPermissions(ctx, adminUserID)
	if err != nil {
		t.Fatalf("load admin permissions: %v", err)
	}
	if len(permissions) == 0 {
		t.Fatal("expected built-in admin user to receive enabled permissions")
	}
}

// TestAdminRoleDoesNotUpgradeOtherUsers verifies assigning the admin role to a
// non-admin username does not trigger the built-in admin bypass flag.
func TestAdminRoleDoesNotUpgradeOtherUsers(t *testing.T) {
	ctx := context.Background()
	svc := newDefaultRoleTestService()

	_, adminRoleID := mustQueryAdminUserAndRoleID(t, ctx)
	username := fmt.Sprintf("role-admin-%d", time.Now().UnixNano())

	userID, err := dao.SysUser.Ctx(ctx).Data(do.SysUser{
		Username: username,
		Password: "test-password-hash",
		Nickname: "Role Admin Test",
		Status:   1,
	}).InsertAndGetId()
	if err != nil {
		t.Fatalf("insert temp user: %v", err)
	}
	t.Cleanup(func() {
		if _, cleanupErr := dao.SysUser.Ctx(ctx).
			Unscoped().
			Where(do.SysUser{Id: int(userID)}).
			Delete(); cleanupErr != nil {
			t.Fatalf("cleanup temp user: %v", cleanupErr)
		}
	})

	if _, err = dao.SysUserRole.Ctx(ctx).Data(do.SysUserRole{
		UserId: int(userID),
		RoleId: adminRoleID,
	}).Insert(); err != nil {
		t.Fatalf("bind temp user admin role: %v", err)
	}
	t.Cleanup(func() {
		if _, cleanupErr := dao.SysUserRole.Ctx(ctx).
			Where(do.SysUserRole{UserId: int(userID), RoleId: adminRoleID}).
			Delete(); cleanupErr != nil {
			t.Fatalf("cleanup temp user-role binding: %v", cleanupErr)
		}
	})

	if svc.IsSuperAdmin(ctx, int(userID)) {
		t.Fatal("expected non-admin username to stay outside built-in admin bypass")
	}

	accessContext, err := svc.GetUserAccessContext(ctx, int(userID))
	if err != nil {
		t.Fatalf("load temp user access context: %v", err)
	}
	if accessContext == nil {
		t.Fatal("expected temp user access context to exist")
	}
	if accessContext.IsSuperAdmin {
		t.Fatal("expected temp user access context to keep IsSuperAdmin=false")
	}
	if accessContext.DataScope != datascope.ScopeAll {
		t.Fatalf("expected admin-role user to carry all-data scope without super-admin bypass, got %d", accessContext.DataScope)
	}
}

// mustQueryAdminUserAndRoleID resolves the built-in admin user ID and role ID for tests.
func mustQueryAdminUserAndRoleID(t *testing.T, ctx context.Context) (int, int) {
	t.Helper()

	var (
		adminUser *entity.SysUser
		adminRole *entity.SysRole
	)

	err := dao.SysUser.Ctx(ctx).
		Where(do.SysUser{Username: accountpolicy.DefaultAdminUsername}).
		Scan(&adminUser)
	if err != nil {
		t.Fatalf("query built-in admin user: %v", err)
	}
	if adminUser == nil {
		t.Fatal("expected built-in admin user to exist")
	}

	err = dao.SysRole.Ctx(ctx).
		Where(do.SysRole{Key: "admin"}).
		Scan(&adminRole)
	if err != nil {
		t.Fatalf("query built-in admin role: %v", err)
	}
	if adminRole == nil {
		t.Fatal("expected built-in admin role to exist")
	}

	return adminUser.Id, adminRole.Id
}

// TestBatchDeleteRemovesRolesAndAssociations verifies batch role deletion
// clears role-menu and user-role associations in one service call.
func TestBatchDeleteRemovesRolesAndAssociations(t *testing.T) {
	var (
		ctx    = context.Background()
		svc    = newDefaultRoleTestService()
		roleID = insertTestRole(t, ctx, "batch-delete-role")
		userID = insertRoleTestUser(t, ctx, "batch-delete-user")
		menuID = insertRoleTestMenu(t, ctx, "batch-delete-menu")
	)
	t.Cleanup(func() {
		cleanupRoleTestRows(t, ctx, []int{roleID}, []int{userID}, []int{menuID})
	})

	if _, err := dao.SysRoleMenu.Ctx(ctx).Data(do.SysRoleMenu{
		RoleId: roleID,
		MenuId: menuID,
	}).Insert(); err != nil {
		t.Fatalf("insert role-menu relation: %v", err)
	}
	if _, err := dao.SysUserRole.Ctx(ctx).Data(do.SysUserRole{
		UserId: userID,
		RoleId: roleID,
	}).Insert(); err != nil {
		t.Fatalf("insert user-role relation: %v", err)
	}

	if err := svc.BatchDelete(ctx, []int{roleID}); err != nil {
		t.Fatalf("batch delete role: %v", err)
	}

	if count := mustCountRole(t, ctx, roleID); count != 0 {
		t.Fatalf("expected role to be soft-deleted, visible count=%d", count)
	}
	if count := mustCountRoleMenu(t, ctx, roleID); count != 0 {
		t.Fatalf("expected role-menu relations to be deleted, got %d", count)
	}
	if count := mustCountUserRole(t, ctx, roleID); count != 0 {
		t.Fatalf("expected user-role relations to be deleted, got %d", count)
	}
}

// TestBatchDeleteRejectsBuiltinAdminRoleAtomically verifies a mixed batch with
// the built-in admin role is rejected before any custom role is deleted.
func TestBatchDeleteRejectsBuiltinAdminRoleAtomically(t *testing.T) {
	var (
		ctx    = context.Background()
		svc    = newDefaultRoleTestService()
		roleID = insertTestRole(t, ctx, "batch-delete-admin-guard")
	)
	_, adminRoleID := mustQueryAdminUserAndRoleID(t, ctx)
	t.Cleanup(func() {
		cleanupRoleTestRows(t, ctx, []int{roleID}, nil, nil)
	})

	err := svc.BatchDelete(ctx, []int{roleID, adminRoleID})
	if err == nil {
		t.Fatal("expected builtin admin role batch delete to be rejected")
	}
	messageErr, ok := bizerr.As(err)
	if !ok || !messageErr.Matches(CodeRoleBuiltinDeleteDenied) {
		t.Fatalf("expected CodeRoleBuiltinDeleteDenied, got %v", err)
	}
	if count := mustCountRole(t, ctx, roleID); count != 1 {
		t.Fatalf("expected custom role to remain visible after rejected batch, count=%d", count)
	}
}

// TestBatchDeleteRejectsEmptyList verifies empty role batches return a stable
// bizerr code before touching the database.
func TestBatchDeleteRejectsEmptyList(t *testing.T) {
	err := newDefaultRoleTestService().BatchDelete(context.Background(), nil)
	if err == nil {
		t.Fatal("expected empty batch delete to be rejected")
	}
	messageErr, ok := bizerr.As(err)
	if !ok || !messageErr.Matches(CodeRoleDeleteIdsRequired) {
		t.Fatalf("expected CodeRoleDeleteIdsRequired, got %v", err)
	}
}

// insertTestRole inserts one temporary role row.
func insertTestRole(t *testing.T, ctx context.Context, label string) int {
	t.Helper()

	suffix := time.Now().UnixNano()
	id, err := dao.SysRole.Ctx(ctx).Data(do.SysRole{
		Name:      fmt.Sprintf("%s-%d", label, suffix),
		Key:       fmt.Sprintf("%s-%d", label, suffix),
		Sort:      99,
		DataScope: roleDataScopeAll,
		Status:    1,
		TenantId:  0,
	}).InsertAndGetId()
	if err != nil {
		t.Fatalf("insert test role: %v", err)
	}
	return int(id)
}

// insertRoleTestUser inserts one temporary user row for role tests.
func insertRoleTestUser(t *testing.T, ctx context.Context, label string) int {
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

// insertRoleTestMenu inserts one temporary menu row for role tests.
func insertRoleTestMenu(t *testing.T, ctx context.Context, label string) int {
	t.Helper()

	key := fmt.Sprintf("%s-%d", label, time.Now().UnixNano())
	id, err := dao.SysMenu.Ctx(ctx).Data(do.SysMenu{
		MenuKey: key,
		Name:    key,
		Type:    "M",
		Sort:    1,
		Visible: 1,
		Status:  1,
	}).InsertAndGetId()
	if err != nil {
		t.Fatalf("insert test menu: %v", err)
	}
	return int(id)
}

// cleanupRoleTestRows removes temporary role, user, and menu rows unscoped.
func cleanupRoleTestRows(t *testing.T, ctx context.Context, roleIDs []int, userIDs []int, menuIDs []int) {
	t.Helper()

	if len(roleIDs) > 0 {
		if _, err := dao.SysRoleMenu.Ctx(ctx).WhereIn(dao.SysRoleMenu.Columns().RoleId, roleIDs).Delete(); err != nil {
			t.Fatalf("cleanup role-menu rows: %v", err)
		}
		if _, err := dao.SysUserRole.Ctx(ctx).WhereIn(dao.SysUserRole.Columns().RoleId, roleIDs).Delete(); err != nil {
			t.Fatalf("cleanup user-role rows: %v", err)
		}
		if _, err := dao.SysRole.Ctx(ctx).Unscoped().WhereIn(dao.SysRole.Columns().Id, roleIDs).Delete(); err != nil {
			t.Fatalf("cleanup role rows: %v", err)
		}
	}
	if len(userIDs) > 0 {
		if _, err := dao.SysUserRole.Ctx(ctx).WhereIn(dao.SysUserRole.Columns().UserId, userIDs).Delete(); err != nil {
			t.Fatalf("cleanup user-role user rows: %v", err)
		}
		if _, err := dao.SysUser.Ctx(ctx).Unscoped().WhereIn(dao.SysUser.Columns().Id, userIDs).Delete(); err != nil {
			t.Fatalf("cleanup user rows: %v", err)
		}
	}
	if len(menuIDs) > 0 {
		if _, err := dao.SysRoleMenu.Ctx(ctx).WhereIn(dao.SysRoleMenu.Columns().MenuId, menuIDs).Delete(); err != nil {
			t.Fatalf("cleanup role-menu menu rows: %v", err)
		}
		if _, err := dao.SysMenu.Ctx(ctx).Unscoped().WhereIn(dao.SysMenu.Columns().Id, menuIDs).Delete(); err != nil {
			t.Fatalf("cleanup menu rows: %v", err)
		}
	}
}

// mustCountRole counts visible role rows for one role ID.
func mustCountRole(t *testing.T, ctx context.Context, roleID int) int {
	t.Helper()

	count, err := dao.SysRole.Ctx(ctx).Where(do.SysRole{Id: roleID}).Count()
	if err != nil {
		t.Fatalf("count role: %v", err)
	}
	return count
}

// mustCountRoleMenu counts role-menu rows for one role ID.
func mustCountRoleMenu(t *testing.T, ctx context.Context, roleID int) int {
	t.Helper()

	count, err := dao.SysRoleMenu.Ctx(ctx).Where(do.SysRoleMenu{RoleId: roleID}).Count()
	if err != nil {
		t.Fatalf("count role-menu relations: %v", err)
	}
	return count
}

// mustCountUserRole counts user-role rows for one role ID.
func mustCountUserRole(t *testing.T, ctx context.Context, roleID int) int {
	t.Helper()

	count, err := dao.SysUserRole.Ctx(ctx).Where(do.SysUserRole{RoleId: roleID}).Count()
	if err != nil {
		t.Fatalf("count user-role relations: %v", err)
	}
	return count
}

// roleTestTranslator stubs the role translation dependency.
type roleTestTranslator struct {
	i18nsvc.Service

	values map[string]string
}

// Translate returns a configured translation or the caller fallback.
func (t roleTestTranslator) Translate(_ context.Context, key string, fallback string) string {
	if value, ok := t.values[key]; ok {
		return value
	}
	return fallback
}

// TestDisplayNameTranslatesBuiltinAdmin verifies the built-in admin role is projected.
func TestDisplayNameTranslatesBuiltinAdmin(t *testing.T) {
	svc := &serviceImpl{
		i18nSvc: roleTestTranslator{
			values: map[string]string{
				"role.builtin.admin.name": "Administrator",
			},
		},
	}

	name := svc.DisplayName(context.Background(), &entity.SysRole{
		Key:  "admin",
		Name: "超级管理员",
	})

	if name != "Administrator" {
		t.Fatalf("expected built-in admin role name to be localized, got %q", name)
	}
}

// TestDisplayNameTranslatesBuiltinUser verifies the built-in standard user role is projected.
func TestDisplayNameTranslatesBuiltinUser(t *testing.T) {
	svc := &serviceImpl{
		i18nSvc: roleTestTranslator{
			values: map[string]string{
				"role.builtin.user.name": "User",
			},
		},
	}

	name := svc.DisplayName(context.Background(), &entity.SysRole{
		Key:  "user",
		Name: "普通用户",
	})

	if name != "User" {
		t.Fatalf("expected built-in user role name to be localized, got %q", name)
	}
}

// TestDisplayNameKeepsCustomRole verifies custom role names remain stored values.
func TestDisplayNameKeepsCustomRole(t *testing.T) {
	svc := &serviceImpl{
		i18nSvc: roleTestTranslator{
			values: map[string]string{
				"role.builtin.admin.name": "Administrator",
			},
		},
	}

	name := svc.DisplayName(context.Background(), &entity.SysRole{
		Key:  "operator",
		Name: "Operator",
	})

	if name != "Operator" {
		t.Fatalf("expected custom role name to remain unchanged, got %q", name)
	}
}

// TestBuildRoleMenuRelationsNormalizesInput verifies invalid and duplicate menu
// IDs do not reach the batch insert payload.
func TestBuildRoleMenuRelationsNormalizesInput(t *testing.T) {
	relations := buildRoleMenuRelations(7, []int{3, 0, 3, -1, 9}, 42)
	if len(relations) != 2 {
		t.Fatalf("expected 2 normalized relations, got %d", len(relations))
	}
	if relations[0].RoleId != 7 || relations[0].MenuId != 3 || relations[0].TenantId != 42 {
		t.Fatalf("unexpected first relation: %#v", relations[0])
	}
	if relations[1].RoleId != 7 || relations[1].MenuId != 9 || relations[1].TenantId != 42 {
		t.Fatalf("unexpected second relation: %#v", relations[1])
	}
}

// TestBuildRoleMenuRelationsRejectsInvalidRole verifies an invalid role ID
// produces no insert payload.
func TestBuildRoleMenuRelationsRejectsInvalidRole(t *testing.T) {
	relations := buildRoleMenuRelations(0, []int{1, 2}, 42)
	if len(relations) != 0 {
		t.Fatalf("expected no relations for invalid role, got %#v", relations)
	}
}

// TestCreateWritesTenantOwnershipAndRoleMenuTenant verifies role creation
// persists the current tenant on both role and role-menu rows.
func TestCreateWritesTenantOwnershipAndRoleMenuTenant(t *testing.T) {
	var (
		ctx    = datascope.WithTenantScope(context.Background(), 62001)
		svc    = newDefaultRoleTestService()
		menuID = insertRoleTenantBoundaryMenu(t, ctx, "tenant-create", "system:tenant:plugin:list", 62001)
	)
	t.Cleanup(func() {
		cleanupRoleTestRows(t, ctx, nil, nil, []int{menuID})
	})

	roleID, err := svc.Create(ctx, CreateInput{
		Name:      uniqueRoleTenantBoundaryName("tenant-role"),
		Key:       uniqueRoleTenantBoundaryName("tenant-role-key"),
		Sort:      1,
		DataScope: roleDataScopeSelf,
		Status:    statusflag.EnabledValue.Int(),
		MenuIds:   []int{menuID},
	})
	if err != nil {
		t.Fatalf("create tenant role: %v", err)
	}
	t.Cleanup(func() {
		cleanupRoleTestRows(t, ctx, []int{roleID}, nil, nil)
	})

	roleRow := mustQueryRoleTenantBoundaryRole(t, ctx, roleID)
	if roleRow.TenantId != 62001 {
		t.Fatalf("expected tenant role ownership 62001, got tenant=%d", roleRow.TenantId)
	}
	if count := mustCountRoleTenantBoundaryRoleMenu(t, ctx, roleID, 62001); count != 1 {
		t.Fatalf("expected one tenant role-menu row, got %d", count)
	}
}

// TestAssignUsersWritesCurrentTenantRelation verifies role assignments persist
// the role tenant boundary.
func TestAssignUsersWritesCurrentTenantRelation(t *testing.T) {
	var (
		ctx            = datascope.WithTenantScope(context.Background(), 62011)
		svc            = newDefaultRoleTestService()
		roleID         = insertRoleTenantBoundaryRole(t, ctx, "tenant-assign", 62011)
		operatorRoleID = insertRoleTenantBoundaryRoleWithScope(t, ctx, "tenant-assign-operator", 62011, roleDataScopeTenant)
		userID         = insertRoleTenantBoundaryUser(t, ctx, "tenant-assign-user", 62011)
	)
	ensureRoleTenantBoundaryMembershipTable(t, ctx)
	svc.tenantSvc = activateRoleTenantBoundaryProvider(t)
	t.Cleanup(func() {
		cleanupRoleTestRows(t, ctx, []int{roleID, operatorRoleID}, []int{userID}, nil)
		cleanupRoleTenantBoundaryMembershipRows(t, ctx, []int{userID})
	})
	insertRoleTenantBoundaryUserRole(t, ctx, userID, operatorRoleID, 62011)
	insertRoleTenantBoundaryMembership(t, ctx, userID, 62011, 1)
	setRoleTestBizCtx(svc, roleScopeStaticBizCtx{ctx: &model.Context{UserId: userID, TenantId: 62011}})

	if err := svc.AssignUsers(ctx, roleID, []int{userID}); err != nil {
		t.Fatalf("assign tenant role user: %v", err)
	}
	if count := mustCountRoleTenantBoundaryUserRole(t, ctx, roleID, userID, 62011); count != 1 {
		t.Fatalf("expected one tenant user-role row, got %d", count)
	}
}

// TestTenantRoleRejectsAllDataScope verifies tenant-local roles cannot receive
// cross-tenant all-data scope.
func TestTenantRoleRejectsAllDataScope(t *testing.T) {
	ctx := datascope.WithTenantScope(context.Background(), 62021)
	svc := newDefaultRoleTestService()

	_, err := svc.Create(ctx, CreateInput{
		Name:      uniqueRoleTenantBoundaryName("tenant-all-data-deny"),
		Key:       uniqueRoleTenantBoundaryName("tenant-all-data-deny-key"),
		Sort:      1,
		DataScope: roleDataScopeAll,
		Status:    statusflag.EnabledValue.Int(),
	})
	if !bizerr.Is(err, CodeTenantRoleAllDataScopeForbidden) {
		t.Fatalf("expected tenant all-data scope denial, got %v", err)
	}
}

// TestPlatformContextRoleRejectsTenantPrimaryUser verifies platform-context
// roles cannot be assigned to tenant-primary users.
func TestPlatformContextRoleRejectsTenantPrimaryUser(t *testing.T) {
	ctx := datascope.WithTenantScope(context.Background(), datascope.PlatformTenantID)
	svc := newDefaultRoleTestService()
	adminUserID, _ := mustQueryAdminUserAndRoleID(t, ctx)
	roleID := insertRoleTenantBoundaryRoleWithScope(t, ctx, "platform-role", 0, roleDataScopeAll)
	userID := insertRoleTenantBoundaryUser(t, ctx, "tenant-primary-user", 62022)
	t.Cleanup(func() {
		cleanupRoleTestRows(t, ctx, []int{roleID}, []int{userID}, nil)
	})
	setRoleTestBizCtx(svc, roleScopeStaticBizCtx{ctx: &model.Context{UserId: adminUserID, TenantId: datascope.PlatformTenantID}})

	err := svc.AssignUsers(ctx, roleID, []int{userID})
	if !bizerr.Is(err, CodePlatformRoleAssignmentForbidden) {
		t.Fatalf("expected platform role assignment denial, got %v", err)
	}
}

// TestTenantRoleRejectsPlatformPrimaryUser verifies tenant roles cannot be
// assigned to platform users, which would make platform authority tenant-local.
func TestTenantRoleRejectsPlatformPrimaryUser(t *testing.T) {
	var (
		ctx            = datascope.WithTenantScope(context.Background(), 62023)
		svc            = newDefaultRoleTestService()
		roleID         = insertRoleTenantBoundaryRoleWithScope(t, ctx, "tenant-role-platform-user-deny", 62023, roleDataScopeTenant)
		operatorRoleID = insertRoleTenantBoundaryRoleWithScope(t, ctx, "tenant-role-platform-user-operator", 62023, roleDataScopeTenant)
		operatorUserID = insertRoleTenantBoundaryUser(t, ctx, "tenant-role-platform-user-operator", 62023)
		platformUserID = insertRoleTenantBoundaryUser(t, ctx, "tenant-role-platform-user", 0)
	)
	t.Cleanup(func() {
		cleanupRoleTestRows(t, ctx, []int{roleID, operatorRoleID}, []int{operatorUserID, platformUserID}, nil)
	})
	insertRoleTenantBoundaryUserRole(t, ctx, operatorUserID, operatorRoleID, 62023)
	setRoleTestBizCtx(svc, roleScopeStaticBizCtx{ctx: &model.Context{UserId: operatorUserID, TenantId: 62023}})

	err := svc.AssignUsers(ctx, roleID, []int{platformUserID})
	if !bizerr.Is(err, CodeTenantRoleAssignmentForbidden) {
		t.Fatalf("expected tenant role assignment to platform user to be denied, got %v", err)
	}
	if count := mustCountRoleTenantBoundaryUserRole(t, ctx, roleID, platformUserID, 62023); count != 0 {
		t.Fatalf("expected no tenant role relation for platform user, got %d", count)
	}
}

// TestTenantRoleRequiresActiveMembershipWhenTableExists verifies tenant role
// assignment checks the linapro-tenant-core membership table when it is installed.
func TestTenantRoleRequiresActiveMembershipWhenTableExists(t *testing.T) {
	ctx := datascope.WithTenantScope(context.Background(), 62024)
	svc := newDefaultRoleTestService()
	ensureRoleTenantBoundaryMembershipTable(t, ctx)
	var (
		roleID         = insertRoleTenantBoundaryRoleWithScope(t, ctx, "tenant-role-membership-deny", 62024, roleDataScopeTenant)
		operatorRoleID = insertRoleTenantBoundaryRoleWithScope(t, ctx, "tenant-role-membership-operator", 62024, roleDataScopeTenant)
		operatorUserID = insertRoleTenantBoundaryUser(t, ctx, "tenant-role-membership-operator", 62024)
		targetUserID   = insertRoleTenantBoundaryUser(t, ctx, "tenant-role-membership-target", 62024)
	)
	t.Cleanup(func() {
		cleanupRoleTestRows(t, ctx, []int{roleID, operatorRoleID}, []int{operatorUserID, targetUserID}, nil)
		cleanupRoleTenantBoundaryMembershipRows(t, ctx, []int{operatorUserID, targetUserID})
	})
	insertRoleTenantBoundaryUserRole(t, ctx, operatorUserID, operatorRoleID, 62024)
	insertRoleTenantBoundaryMembership(t, ctx, operatorUserID, 62024, 1)
	svc.tenantSvc = activateRoleTenantBoundaryProvider(t)
	setRoleTestBizCtx(svc, roleScopeStaticBizCtx{ctx: &model.Context{UserId: operatorUserID, TenantId: 62024}})

	err := svc.AssignUsers(ctx, roleID, []int{targetUserID})
	if !bizerr.Is(err, CodeTenantRoleAssignmentForbidden) {
		t.Fatalf("expected tenant role assignment without membership to be denied, got %v", err)
	}
	if count := mustCountRoleTenantBoundaryUserRole(t, ctx, roleID, targetUserID, 62024); count != 0 {
		t.Fatalf("expected no tenant role relation without membership, got %d", count)
	}

	insertRoleTenantBoundaryMembership(t, ctx, targetUserID, 62024, 1)
	if err = svc.AssignUsers(ctx, roleID, []int{targetUserID}); err != nil {
		t.Fatalf("expected tenant role assignment with active membership to succeed, got %v", err)
	}
	if count := mustCountRoleTenantBoundaryUserRole(t, ctx, roleID, targetUserID, 62024); count != 1 {
		t.Fatalf("expected one tenant role relation with membership, got %d", count)
	}
}

// TestTenantRoleAccessFiltersRoleMenuByTenant verifies permission resolution
// does not reuse role-menu rows from another tenant for the same role ID.
func TestTenantRoleAccessFiltersRoleMenuByTenant(t *testing.T) {
	var (
		ctx          = datascope.WithTenantScope(context.Background(), 62031)
		svc          = newDefaultRoleTestService()
		roleID       = insertRoleTenantBoundaryRole(t, ctx, "tenant-access", 62031)
		userID       = insertRoleTenantBoundaryUser(t, ctx, "tenant-access-user", 62031)
		tenantMenuID = insertRoleTenantBoundaryMenu(t, ctx, "tenant-access-menu", "system:tenant:visible", 62031)
		otherMenuID  = insertRoleTenantBoundaryMenu(t, ctx, "tenant-access-other-menu", "system:tenant:hidden", 62032)
	)
	t.Cleanup(func() {
		cleanupRoleTestRows(t, ctx, []int{roleID}, []int{userID}, []int{tenantMenuID, otherMenuID})
	})
	insertRoleTenantBoundaryRoleMenu(t, ctx, roleID, tenantMenuID, 62031)
	insertRoleTenantBoundaryRoleMenu(t, ctx, roleID, otherMenuID, 62032)
	insertRoleTenantBoundaryUserRole(t, ctx, userID, roleID, 62031)

	access, err := svc.GetUserAccessContext(ctx, userID)
	if err != nil {
		t.Fatalf("load tenant role access: %v", err)
	}
	if access == nil || !containsRoleTenantBoundaryString(access.Permissions, "system:tenant:visible") {
		t.Fatalf("expected tenant permission in access snapshot, got %#v", access)
	}
	if containsRoleTenantBoundaryString(access.Permissions, "system:tenant:hidden") {
		t.Fatalf("did not expect cross-tenant permission in access snapshot, got %#v", access.Permissions)
	}
}

// TestImpersonationAccessUsesPlatformRoles verifies tenant impersonation keeps
// target-tenant data context while permission grants come from platform roles.
func TestImpersonationAccessUsesPlatformRoles(t *testing.T) {
	var (
		ctx    = datascope.WithTenantScope(context.Background(), 62041)
		svc    = newDefaultRoleTestService()
		roleID = insertRoleTenantBoundaryRoleWithScope(t, ctx, "impersonation-platform-role", datascope.PlatformTenantID, roleDataScopeAll)
		userID = insertRoleTenantBoundaryUser(t, ctx, "impersonation-platform-user", datascope.PlatformTenantID)
		menuID = insertRoleTenantBoundaryMenu(t, ctx, "impersonation-platform-menu", "system:tenant:impersonate:test", datascope.PlatformTenantID)
	)
	t.Cleanup(func() {
		cleanupRoleTestRows(t, ctx, []int{roleID}, []int{userID}, []int{menuID})
	})
	insertRoleTenantBoundaryRoleMenu(t, ctx, roleID, menuID, datascope.PlatformTenantID)
	insertRoleTenantBoundaryUserRole(t, ctx, userID, roleID, datascope.PlatformTenantID)
	setRoleTestBizCtx(svc, roleScopeStaticBizCtx{ctx: &model.Context{
		UserId:         userID,
		TenantId:       62041,
		ActingAsTenant: true,
		ActingUserId:   userID,
	}})

	access, err := svc.GetUserAccessContext(ctx, userID)
	if err != nil {
		t.Fatalf("load impersonation access: %v", err)
	}
	if access == nil || !containsRoleTenantBoundaryString(access.Permissions, "system:tenant:impersonate:test") {
		t.Fatalf("expected platform role permission during impersonation, got %#v", access)
	}
	if access.DataScope != datascope.ScopeAll {
		t.Fatalf("expected platform role data scope during impersonation, got %d", access.DataScope)
	}
	if datascope.CurrentTenantID(ctx) != 62041 {
		t.Fatalf("expected request tenant to remain target tenant, got %d", datascope.CurrentTenantID(ctx))
	}
}

// activateRoleTenantBoundaryProvider returns a real tenant service with one
// enabled membership provider for role assignment tests.
func activateRoleTenantBoundaryProvider(t *testing.T) tenantspi.Service {
	t.Helper()
	providerPluginID := fmt.Sprintf("plugin-test-role-tenant-provider-%d", time.Now().UnixNano())
	manager := tenantspi.NewManager()
	if err := manager.RegisterFactory(providerPluginID, func(context.Context, tenantspi.ProviderEnv) (tenantspi.Provider, error) {
		return roleTenantBoundaryProvider{}, nil
	}); err != nil {
		t.Fatalf("register role tenant provider: %v", err)
	}
	return tenantspi.New(manager, roleTenantBoundaryRuntime{pluginID: providerPluginID}, nil, roleTenantBoundaryBizCtx{})
}

// roleTenantBoundaryRuntime marks exactly one role test tenant provider enabled.
type roleTenantBoundaryRuntime struct {
	pluginID string
}

// IsProviderEnabled reports whether the given test provider plugin is enabled.
func (r roleTenantBoundaryRuntime) IsProviderEnabled(_ context.Context, pluginID string) bool {
	return pluginID == r.pluginID
}

// roleTenantBoundaryBizCtx adapts role data-scope test context into tenant context.
type roleTenantBoundaryBizCtx struct{}

// Current returns the plugin-visible tenant context derived from datascope test helpers.
func (roleTenantBoundaryBizCtx) Current(ctx context.Context) bizctxcap.CurrentContext {
	tenantID := datascope.CurrentTenantID(ctx)
	return bizctxcap.CurrentContext{
		TenantID:        tenantID,
		PlatformBypass:  tenantID == datascope.PlatformTenantID,
		ActingAsTenant:  tenantID != datascope.PlatformTenantID,
		IsImpersonation: false,
	}
}

// roleTenantBoundaryProvider simulates the plugin-owned membership governance provider.
type roleTenantBoundaryProvider struct{}

// ResolveTenant is unused by role assignment tests.
func (roleTenantBoundaryProvider) ResolveTenant(
	context.Context,
	*ghttp.Request,
) (*tenantcap.ResolverResult, error) {
	return &tenantcap.ResolverResult{TenantID: tenantcap.PLATFORM, Matched: true}, nil
}

// ValidateUserInTenant is unused by role assignment tests.
func (roleTenantBoundaryProvider) ValidateUserInTenant(context.Context, int, tenantcap.TenantID) error {
	return nil
}

// ListUserTenants is unused by role assignment tests.
func (roleTenantBoundaryProvider) ListUserTenants(context.Context, int) ([]tenantcap.TenantInfo, error) {
	return nil, nil
}

// SwitchTenant is unused by role assignment tests.
func (roleTenantBoundaryProvider) SwitchTenant(context.Context, int, tenantcap.TenantID) error {
	return nil
}

// ApplyUserTenantScope is unused by role assignment tests.
func (roleTenantBoundaryProvider) ApplyUserTenantScope(
	_ context.Context,
	model *gdb.Model,
	_ string,
) (*gdb.Model, bool, error) {
	return model, false, nil
}

// ApplyUserTenantFilter is unused by role assignment tests.
func (roleTenantBoundaryProvider) ApplyUserTenantFilter(
	_ context.Context,
	model *gdb.Model,
	_ string,
	_ tenantcap.TenantID,
) (*gdb.Model, bool, error) {
	return model, false, nil
}

// ListUserTenantMemberships is unused by role assignment tests.
func (roleTenantBoundaryProvider) ListUserTenantMemberships(
	context.Context,
	[]int,
) (map[int]*tenantcap.TenantMembershipInfo, error) {
	return nil, nil
}

// ResolveUserTenantAssignment is unused by role assignment tests.
func (roleTenantBoundaryProvider) ResolveUserTenantAssignment(
	context.Context,
	[]tenantcap.TenantID,
	tenantcap.UserTenantAssignmentMode,
) (*tenantcap.UserTenantAssignmentPlan, error) {
	return &tenantcap.UserTenantAssignmentPlan{}, nil
}

// ReplaceUserTenantAssignments is unused by role assignment tests.
func (roleTenantBoundaryProvider) ReplaceUserTenantAssignments(
	context.Context,
	int,
	*tenantcap.UserTenantAssignmentPlan,
) error {
	return nil
}

// EnsureUsersInTenant verifies role assignment targets are active tenant members.
func (roleTenantBoundaryProvider) EnsureUsersInTenant(
	ctx context.Context,
	userIDs []int,
	tenantID tenantcap.TenantID,
) error {
	if len(userIDs) == 0 || tenantID <= tenantcap.PLATFORM {
		return nil
	}
	count, err := dao.SysUser.DB().Model("plugin_linapro_tenant_core_user_membership").Safe().Ctx(ctx).
		WhereIn("user_id", userIDs).
		Where("tenant_id", int(tenantID)).
		Where("status", 1).
		Count()
	if err != nil {
		return err
	}
	if count != len(userIDs) {
		return bizerr.NewCode(tenantcap.CodeTenantForbidden, bizerr.P("tenantId", int(tenantID)))
	}
	return nil
}

// ValidateStartupConsistency is unused by role assignment tests.
func (roleTenantBoundaryProvider) ValidateStartupConsistency(context.Context) ([]string, error) {
	return nil, nil
}

// uniqueRoleTenantBoundaryName builds a stable unique test label.
func uniqueRoleTenantBoundaryName(prefix string) string {
	return fmt.Sprintf("%s-%d", prefix, time.Now().UnixNano())
}

// insertRoleTenantBoundaryRole inserts one role with explicit tenant ownership.
func insertRoleTenantBoundaryRole(t *testing.T, ctx context.Context, label string, tenantID int) int {
	t.Helper()
	return insertRoleTenantBoundaryRoleWithScope(t, ctx, label, tenantID, roleDataScopeSelf)
}

// insertRoleTenantBoundaryRoleWithScope inserts one role with explicit tenant
// ownership and data scope.
func insertRoleTenantBoundaryRoleWithScope(
	t *testing.T,
	ctx context.Context,
	label string,
	tenantID int,
	dataScope int,
) int {
	t.Helper()

	name := uniqueRoleTenantBoundaryName(label)
	id, err := dao.SysRole.Ctx(ctx).Data(do.SysRole{
		Name:      name,
		Key:       name,
		Sort:      99,
		DataScope: dataScope,
		Status:    statusflag.EnabledValue.Int(),
		TenantId:  tenantID,
	}).InsertAndGetId()
	if err != nil {
		t.Fatalf("insert tenant boundary role: %v", err)
	}
	return int(id)
}

// insertRoleTenantBoundaryUser inserts one temporary user.
func insertRoleTenantBoundaryUser(t *testing.T, ctx context.Context, label string, tenantID int) int {
	t.Helper()

	username := uniqueRoleTenantBoundaryName(label)
	id, err := dao.SysUser.Ctx(ctx).Data(do.SysUser{
		Username: username,
		Password: "test-password-hash",
		Nickname: username,
		Status:   statusflag.EnabledValue.Int(),
		TenantId: tenantID,
	}).InsertAndGetId()
	if err != nil {
		t.Fatalf("insert tenant boundary user: %v", err)
	}
	return int(id)
}

// insertRoleTenantBoundaryMenu inserts one temporary global menu permission row.
func insertRoleTenantBoundaryMenu(t *testing.T, ctx context.Context, label string, permission string, _ int) int {
	t.Helper()

	key := uniqueRoleTenantBoundaryName(label)
	id, err := dao.SysMenu.Ctx(ctx).Data(do.SysMenu{
		MenuKey: key,
		Name:    key,
		Perms:   permission,
		Type:    "F",
		Sort:    99,
		Visible: statusflag.Visible.Int(),
		Status:  statusflag.EnabledValue.Int(),
	}).InsertAndGetId()
	if err != nil {
		t.Fatalf("insert tenant boundary menu: %v", err)
	}
	return int(id)
}

// insertRoleTenantBoundaryRoleMenu inserts one role-menu relation.
func insertRoleTenantBoundaryRoleMenu(t *testing.T, ctx context.Context, roleID int, menuID int, tenantID int) {
	t.Helper()

	if _, err := dao.SysRoleMenu.Ctx(ctx).Data(do.SysRoleMenu{
		RoleId:   roleID,
		MenuId:   menuID,
		TenantId: tenantID,
	}).Insert(); err != nil {
		t.Fatalf("insert tenant boundary role-menu: %v", err)
	}
}

// insertRoleTenantBoundaryUserRole inserts one user-role relation.
func insertRoleTenantBoundaryUserRole(t *testing.T, ctx context.Context, userID int, roleID int, tenantID int) {
	t.Helper()

	if _, err := dao.SysUserRole.Ctx(ctx).Data(do.SysUserRole{
		UserId:   userID,
		RoleId:   roleID,
		TenantId: tenantID,
	}).Insert(); err != nil {
		t.Fatalf("insert tenant boundary user-role: %v", err)
	}
}

// ensureRoleTenantBoundaryMembershipTable creates the minimal linapro-tenant-core
// membership table needed by role assignment tests when the plugin schema is not installed.
func ensureRoleTenantBoundaryMembershipTable(t *testing.T, ctx context.Context) {
	t.Helper()
	_, err := dao.SysUser.DB().Exec(ctx, `
CREATE TABLE IF NOT EXISTS plugin_linapro_tenant_core_user_membership (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    tenant_id BIGINT NOT NULL,
    status INT NOT NULL DEFAULT 1,
    deleted_at TIMESTAMP NULL
)`)
	if err != nil {
		t.Fatalf("ensure role tenant boundary membership table failed: %v", err)
	}
}

// insertRoleTenantBoundaryMembership inserts one active or inactive membership row.
func insertRoleTenantBoundaryMembership(t *testing.T, ctx context.Context, userID int, tenantID int, status int) {
	t.Helper()
	if _, err := dao.SysUser.DB().Model("plugin_linapro_tenant_core_user_membership").Safe().Ctx(ctx).Data(struct {
		UserID   int `orm:"user_id"`
		TenantID int `orm:"tenant_id"`
		Status   int `orm:"status"`
	}{
		UserID:   userID,
		TenantID: tenantID,
		Status:   status,
	}).Insert(); err != nil {
		t.Fatalf("insert role tenant boundary membership: %v", err)
	}
}

// cleanupRoleTenantBoundaryMembershipRows removes temporary membership rows.
func cleanupRoleTenantBoundaryMembershipRows(t *testing.T, ctx context.Context, userIDs []int) {
	t.Helper()
	if len(userIDs) == 0 {
		return
	}
	if _, err := dao.SysUser.DB().Model("plugin_linapro_tenant_core_user_membership").Safe().Ctx(ctx).Unscoped().WhereIn("user_id", userIDs).Delete(); err != nil {
		t.Errorf("cleanup role tenant boundary membership rows failed: %v", err)
	}
}

// mustQueryRoleTenantBoundaryRole loads a role row for assertions.
func mustQueryRoleTenantBoundaryRole(t *testing.T, ctx context.Context, roleID int) *roleTenantBoundaryRoleProjection {
	t.Helper()

	var roleRow *roleTenantBoundaryRoleProjection
	if err := dao.SysRole.Ctx(ctx).Where(do.SysRole{Id: roleID}).Scan(&roleRow); err != nil {
		t.Fatalf("query tenant boundary role: %v", err)
	}
	if roleRow == nil {
		t.Fatalf("expected role %d to exist", roleID)
	}
	return roleRow
}

// mustCountRoleTenantBoundaryRoleMenu counts role-menu rows by tenant.
func mustCountRoleTenantBoundaryRoleMenu(t *testing.T, ctx context.Context, roleID int, tenantID int) int {
	t.Helper()

	count, err := dao.SysRoleMenu.Ctx(ctx).Where(do.SysRoleMenu{RoleId: roleID, TenantId: tenantID}).Count()
	if err != nil {
		t.Fatalf("count tenant boundary role-menu: %v", err)
	}
	return count
}

// mustCountRoleTenantBoundaryUserRole counts user-role rows by tenant.
func mustCountRoleTenantBoundaryUserRole(t *testing.T, ctx context.Context, roleID int, userID int, tenantID int) int {
	t.Helper()

	count, err := dao.SysUserRole.Ctx(ctx).
		Where(do.SysUserRole{RoleId: roleID, UserId: userID, TenantId: tenantID}).
		Count()
	if err != nil {
		t.Fatalf("count tenant boundary user-role: %v", err)
	}
	return count
}

// roleTenantBoundaryRoleProjection is a compact role ownership projection.
type roleTenantBoundaryRoleProjection struct {
	TenantId int `json:"tenantId" orm:"tenant_id"`
}

// containsRoleTenantBoundaryString reports whether a slice contains a value.
func containsRoleTenantBoundaryString(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

// newRoleTestService constructs a role service with explicit test dependencies,
// including the shared data-scope service required by role user operations.
func newRoleTestService(permissionFilter permissionMenuFilter, orgCapSvc orgcap.Service) *serviceImpl {
	var (
		bizCtxSvc        = bizctx.New()
		configSvc        = hostconfig.New(nil)
		cacheCoordSvc    = cachecoord.New(nil, nil)
		i18nSvc          = i18nsvc.New(bizCtxSvc, configSvc, cacheCoordSvc)
		defaultOrgCapSvc = orgspi.New(nil, nil, nil)
		tenantSvc        = tenantspi.New(nil, nil, nil, nil)
	)
	if orgCapSvc == nil {
		orgCapSvc = defaultOrgCapSvc
	}
	svc := New(permissionFilter, bizCtxSvc, configSvc, i18nSvc, orgCapSvc, tenantSvc, cacheCoordSvc).(*serviceImpl)
	refreshRoleTestScope(svc, orgCapSvc)
	return svc
}

// newDefaultRoleTestService constructs the default role service used by most tests.
func newDefaultRoleTestService() *serviceImpl {
	return newRoleTestService(nil, nil)
}

// setRoleTestBizCtx replaces the business context dependency and refreshes
// the derived data-scope service used by role-management tests.
func setRoleTestBizCtx(svc *serviceImpl, bizCtxSvc bizctx.Service) {
	svc.bizCtxSvc = bizCtxSvc
	refreshRoleTestScope(svc, nil)
}

// refreshRoleTestScope rebuilds the stateless data-scope helper from the
// current explicit fake dependencies.
func refreshRoleTestScope(svc *serviceImpl, orgCapSvc orgcap.Service) {
	var orgScope orgspi.ScopeService
	if scope, ok := orgCapSvc.(orgspi.ScopeService); ok {
		orgScope = scope
	}
	svc.SetDataScopeService(datascope.New(svc.bizCtxSvc, svc, orgScope))
}
