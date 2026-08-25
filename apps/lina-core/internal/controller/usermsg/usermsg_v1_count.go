// This file implements the unread inbox count endpoint used by the workbench
// message badge.

package usermsg

import (
	"context"

	v1 "lina-core/api/usermsg/v1"
)

// Count returns unread message count
func (c *ControllerV1) Count(ctx context.Context, req *v1.CountReq) (res *v1.CountRes, err error) {
	count, err := c.notifySvc.InboxUnreadCount(ctx, c.currentUserID(ctx))
	if err != nil {
		return nil, err
	}
	return &v1.CountRes{Count: count}, nil
}
