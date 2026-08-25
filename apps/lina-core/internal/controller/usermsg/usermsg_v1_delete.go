// This file implements the current-user inbox delete endpoint.

package usermsg

import (
	"context"

	v1 "lina-core/api/usermsg/v1"
)

// Delete deletes a user message
func (c *ControllerV1) Delete(ctx context.Context, req *v1.DeleteReq) (res *v1.DeleteRes, err error) {
	if err = c.notifySvc.InboxDelete(ctx, c.currentUserID(ctx), req.Id); err != nil {
		return nil, err
	}
	return &v1.DeleteRes{}, nil
}
