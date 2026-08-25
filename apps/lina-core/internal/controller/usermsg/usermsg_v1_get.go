// usermsg_v1_get.go implements the current-user message-detail endpoint used by
// the inbox preview dialog.

package usermsg

import (
	"context"

	v1 "lina-core/api/usermsg/v1"
	"lina-core/pkg/apitime"
)

// Get returns one current-user message detail for inbox preview.
func (c *ControllerV1) Get(ctx context.Context, req *v1.GetReq) (res *v1.GetRes, err error) {
	detail, err := c.notifySvc.InboxGet(ctx, c.currentUserID(ctx), req.Id)
	if err != nil {
		return nil, err
	}
	categoryCode := resolveCategoryCode(detail.CategoryCode)
	return &v1.GetRes{
		Id:            detail.Id,
		Title:         detail.Title,
		CategoryCode:  categoryCode,
		TypeLabel:     c.localizeCategoryLabel(ctx, categoryCode),
		TypeColor:     c.localizeCategoryColor(ctx, categoryCode),
		SourceType:    v1.SourceType(detail.SourceType),
		SourceId:      detail.SourceID,
		Content:       detail.Content,
		CreatedByName: detail.CreatedByName,
		CreatedAt:     apitime.Milli(detail.CreatedAt),
	}, nil
}
