// This file implements the current-user mark-all-read inbox endpoint.

package usermsg

import (
	"context"

	v1 "lina-core/api/usermsg/v1"
)

// ReadAll marks all messages as read
func (c *ControllerV1) ReadAll(ctx context.Context, req *v1.ReadAllReq) (res *v1.ReadAllRes, err error) {
	if err = c.notifySvc.InboxMarkAllRead(ctx, c.currentUserID(ctx)); err != nil {
		return nil, err
	}
	return &v1.ReadAllRes{}, nil
}
