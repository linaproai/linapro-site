// This file implements the current-user mark-one-read inbox endpoint.

package usermsg

import (
	"context"

	v1 "lina-core/api/usermsg/v1"
)

// Read marks a message as read
func (c *ControllerV1) Read(ctx context.Context, req *v1.ReadReq) (res *v1.ReadRes, err error) {
	if err = c.notifySvc.InboxMarkRead(ctx, c.currentUserID(ctx), req.Id); err != nil {
		return nil, err
	}
	return &v1.ReadRes{}, nil
}
