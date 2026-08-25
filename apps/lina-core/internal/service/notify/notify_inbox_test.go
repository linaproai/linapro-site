// This file verifies user messages remain self-isolated regardless of role data scope.

package notify

import (
	"context"
	"fmt"
	"github.com/gogf/gf/v2/util/gconv"
	usermsgv1 "lina-core/api/usermsg/v1"
	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/pkg/bizerr"
	_ "lina-core/pkg/dbdriver"
	"lina-core/pkg/plugin/capability/tenantcap/tenantspi"
	"testing"
	"time"
)

// TestUserMessagesRemainSelfIsolatedForAllDataScope verifies all-data role
// scope never broadens inbox message reads, marks, or deletes to other users.
func TestUserMessagesRemainSelfIsolatedForAllDataScope(t *testing.T) {
	var (
		ctx           = context.Background()
		currentUserID = insertUserMsgScopeUser(t, ctx, "usermsg-current")
		otherUserID   = insertUserMsgScopeUser(t, ctx, "usermsg-other")
		roleID        = insertUserMsgScopeRole(t, ctx, "usermsg-all", 1)
	)
	t.Cleanup(func() {
		cleanupUserMsgScopeUsers(t, ctx, []int{currentUserID, otherUserID})
		cleanupUserMsgScopeRoles(t, ctx, []int{roleID})
	})
	insertUserMsgScopeUserRole(t, ctx, currentUserID, roleID)

	currentDeliveryID, currentSourceID := insertUserMsgScopeDelivery(t, ctx, currentUserID, "current-message")
	otherDeliveryID, _ := insertUserMsgScopeDelivery(t, ctx, otherUserID, "other-message")
	t.Cleanup(func() { cleanupUserMsgScopeDeliveries(t, ctx, []int64{currentDeliveryID, otherDeliveryID}) })

	svc := New(tenantspi.New(nil, nil, nil, nil))

	out, err := svc.InboxList(ctx, InboxListInput{UserID: int64(currentUserID), PageNum: 1, PageSize: 20})
	if err != nil {
		t.Fatalf("list user messages: %v", err)
	}
	if out.Total != 1 || len(out.List) != 1 || out.List[0].Id != currentDeliveryID {
		t.Fatalf("expected only current user's message, got total=%d list=%#v", out.Total, out.List)
	}
	if out.List[0].SourceID != currentSourceID {
		t.Fatalf("expected source ID %q in list projection, got %q", currentSourceID, out.List[0].SourceID)
	}

	detail, err := svc.InboxGet(ctx, int64(currentUserID), currentDeliveryID)
	if err != nil {
		t.Fatalf("get current message: %v", err)
	}
	if detail.SourceID != currentSourceID {
		t.Fatalf("expected source ID %q in detail projection, got %q", currentSourceID, detail.SourceID)
	}

	if _, err = svc.InboxGet(ctx, int64(currentUserID), otherDeliveryID); !bizerr.Is(err, CodeNotifyInboxNotFound) {
		t.Fatalf("expected other message get to be hidden as not found, got %v", err)
	}
	if err = svc.InboxMarkRead(ctx, int64(currentUserID), otherDeliveryID); err != nil {
		t.Fatalf("mark other message should be a no-op update scoped to current user: %v", err)
	}
	if read := queryUserMsgDeliveryRead(t, ctx, otherDeliveryID); read != 0 {
		t.Fatalf("expected other message to remain unread, got %d", read)
	}
	if err = svc.InboxDelete(ctx, int64(currentUserID), otherDeliveryID); err != nil {
		t.Fatalf("delete other message should be a no-op delete scoped to current user: %v", err)
	}
	if count := countUserMsgDelivery(t, ctx, otherDeliveryID); count != 1 {
		t.Fatalf("expected other message to remain after scoped delete, count=%d", count)
	}
}

// TestInboxMethodsRejectMissingUserID verifies inbox entry points fail closed
// with the notify-owned unauthenticated code before touching storage.
func TestInboxMethodsRejectMissingUserID(t *testing.T) {
	ctx := context.Background()
	svc := New(nil)

	if _, err := svc.InboxUnreadCount(ctx, 0); !bizerr.Is(err, CodeNotifyNotAuthenticated) {
		t.Fatalf("InboxUnreadCount missing user: got %v, want CodeNotifyNotAuthenticated", err)
	}
	if _, err := svc.InboxList(ctx, InboxListInput{UserID: 0, PageNum: 1, PageSize: 10}); !bizerr.Is(err, CodeNotifyNotAuthenticated) {
		t.Fatalf("InboxList missing user: got %v, want CodeNotifyNotAuthenticated", err)
	}
	if _, err := svc.InboxGet(ctx, 0, 1); !bizerr.Is(err, CodeNotifyNotAuthenticated) {
		t.Fatalf("InboxGet missing user: got %v, want CodeNotifyNotAuthenticated", err)
	}
	if err := svc.InboxMarkRead(ctx, 0, 1); !bizerr.Is(err, CodeNotifyNotAuthenticated) {
		t.Fatalf("InboxMarkRead missing user: got %v, want CodeNotifyNotAuthenticated", err)
	}
	if err := svc.InboxMarkAllRead(ctx, 0); !bizerr.Is(err, CodeNotifyNotAuthenticated) {
		t.Fatalf("InboxMarkAllRead missing user: got %v, want CodeNotifyNotAuthenticated", err)
	}
	if err := svc.InboxDelete(ctx, 0, 1); !bizerr.Is(err, CodeNotifyNotAuthenticated) {
		t.Fatalf("InboxDelete missing user: got %v, want CodeNotifyNotAuthenticated", err)
	}
	if err := svc.InboxClear(ctx, 0); !bizerr.Is(err, CodeNotifyNotAuthenticated) {
		t.Fatalf("InboxClear missing user: got %v, want CodeNotifyNotAuthenticated", err)
	}
}

// insertUserMsgScopeUser inserts one temporary user.
func insertUserMsgScopeUser(t *testing.T, ctx context.Context, prefix string) int {
	t.Helper()
	id, err := dao.SysUser.Ctx(ctx).Data(do.SysUser{
		Username: uniqueUserMsgScopeName(prefix),
		Password: "hashed",
		Nickname: prefix,
		Status:   1,
	}).InsertAndGetId()
	if err != nil {
		t.Fatalf("insert usermsg-scope user: %v", err)
	}
	return int(id)
}

// uniqueUserMsgScopeName returns one collision-resistant identifier.
func uniqueUserMsgScopeName(prefix string) string {
	return fmt.Sprintf("%s-%d", prefix, time.Now().UnixNano())
}

// insertUserMsgScopeRole inserts one temporary role.
func insertUserMsgScopeRole(t *testing.T, ctx context.Context, prefix string, scope int) int {
	t.Helper()
	id, err := dao.SysRole.Ctx(ctx).Data(do.SysRole{
		Name:      uniqueUserMsgScopeName(prefix),
		Key:       uniqueUserMsgScopeName(prefix + "-key"),
		Sort:      99,
		DataScope: scope,
		Status:    1,
	}).InsertAndGetId()
	if err != nil {
		t.Fatalf("insert usermsg-scope role: %v", err)
	}
	return int(id)
}

// insertUserMsgScopeUserRole binds one user to one role.
func insertUserMsgScopeUserRole(t *testing.T, ctx context.Context, userID int, roleID int) {
	t.Helper()
	if _, err := dao.SysUserRole.Ctx(ctx).Data(do.SysUserRole{UserId: userID, RoleId: roleID}).Insert(); err != nil {
		t.Fatalf("insert usermsg-scope user role: %v", err)
	}
}

// insertUserMsgScopeDelivery inserts one message and one inbox delivery.
func insertUserMsgScopeDelivery(t *testing.T, ctx context.Context, userID int, title string) (int64, string) {
	t.Helper()
	sourceID := uniqueUserMsgScopeName("source")
	messageID, err := dao.SysNotifyMessage.Ctx(ctx).Data(do.SysNotifyMessage{
		SourceType:   string(usermsgv1.SourceTypeSystem),
		SourceId:     sourceID,
		CategoryCode: CategoryCodeSystem.String(),
		Title:        title,
		Content:      title + " content",
		PayloadJson:  "{}",
		SenderUserId: userID,
	}).InsertAndGetId()
	if err != nil {
		t.Fatalf("insert usermsg-scope message: %v", err)
	}
	deliveryID, err := dao.SysNotifyDelivery.Ctx(ctx).Data(do.SysNotifyDelivery{
		MessageId:      messageID,
		ChannelKey:     ChannelKeyInbox,
		ChannelType:    ChannelTypeInbox.String(),
		RecipientType:  RecipientTypeUser.String(),
		RecipientKey:   gconv.String(userID),
		UserId:         userID,
		DeliveryStatus: DeliveryStatusSucceeded,
		IsRead:         0,
	}).InsertAndGetId()
	if err != nil {
		t.Fatalf("insert usermsg-scope delivery: %v", err)
	}
	return deliveryID, sourceID
}

// cleanupUserMsgScopeUsers removes temporary users.
func cleanupUserMsgScopeUsers(t *testing.T, ctx context.Context, ids []int) {
	t.Helper()
	if len(ids) == 0 {
		return
	}
	if _, err := dao.SysUserRole.Ctx(ctx).WhereIn(dao.SysUserRole.Columns().UserId, ids).Delete(); err != nil {
		t.Fatalf("cleanup usermsg user roles: %v", err)
	}
	if _, err := dao.SysUser.Ctx(ctx).Unscoped().WhereIn(dao.SysUser.Columns().Id, ids).Delete(); err != nil {
		t.Fatalf("cleanup usermsg users: %v", err)
	}
}

// cleanupUserMsgScopeRoles removes temporary roles.
func cleanupUserMsgScopeRoles(t *testing.T, ctx context.Context, ids []int) {
	t.Helper()
	if len(ids) == 0 {
		return
	}
	if _, err := dao.SysRole.Ctx(ctx).Unscoped().WhereIn(dao.SysRole.Columns().Id, ids).Delete(); err != nil {
		t.Fatalf("cleanup usermsg roles: %v", err)
	}
}

// cleanupUserMsgScopeDeliveries removes temporary deliveries and messages.
func cleanupUserMsgScopeDeliveries(t *testing.T, ctx context.Context, ids []int64) {
	t.Helper()
	if len(ids) == 0 {
		return
	}
	var rows []struct {
		MessageId int64 `json:"messageId"`
	}
	if err := dao.SysNotifyDelivery.Ctx(ctx).Unscoped().Fields(dao.SysNotifyDelivery.Columns().MessageId).WhereIn(dao.SysNotifyDelivery.Columns().Id, ids).Scan(&rows); err != nil {
		t.Fatalf("query usermsg deliveries for cleanup: %v", err)
	}
	if _, err := dao.SysNotifyDelivery.Ctx(ctx).Unscoped().WhereIn(dao.SysNotifyDelivery.Columns().Id, ids).Delete(); err != nil {
		t.Fatalf("cleanup usermsg deliveries: %v", err)
	}
	messageIDs := make([]int64, 0, len(rows))
	for _, row := range rows {
		messageIDs = append(messageIDs, row.MessageId)
	}
	if len(messageIDs) > 0 {
		if _, err := dao.SysNotifyMessage.Ctx(ctx).WhereIn(dao.SysNotifyMessage.Columns().Id, messageIDs).Delete(); err != nil {
			t.Fatalf("cleanup usermsg messages: %v", err)
		}
	}
}

// queryUserMsgDeliveryRead returns one delivery read flag.
func queryUserMsgDeliveryRead(t *testing.T, ctx context.Context, id int64) int {
	t.Helper()
	var row *struct {
		IsRead int `json:"isRead"`
	}
	if err := dao.SysNotifyDelivery.Ctx(ctx).Unscoped().Fields(dao.SysNotifyDelivery.Columns().IsRead).Where(do.SysNotifyDelivery{Id: id}).Scan(&row); err != nil {
		t.Fatalf("query usermsg read flag: %v", err)
	}
	if row == nil {
		t.Fatalf("expected usermsg delivery %d", id)
	}
	return row.IsRead
}

// countUserMsgDelivery counts visible delivery rows by ID.
func countUserMsgDelivery(t *testing.T, ctx context.Context, id int64) int {
	t.Helper()
	count, err := dao.SysNotifyDelivery.Ctx(ctx).Where(do.SysNotifyDelivery{Id: id}).Count()
	if err != nil {
		t.Fatalf("count usermsg delivery: %v", err)
	}
	return count
}
