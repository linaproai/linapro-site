// This file implements the scheduled job handler detail endpoint.

package jobhandler

import (
	"context"

	v1 "lina-core/api/jobhandler/v1"
	jobmgmtsvc "lina-core/internal/service/jobmgmt"
	"lina-core/pkg/bizerr"
)

// Detail handles scheduled job handler detail lookup requests.
func (c *ControllerV1) Detail(ctx context.Context, req *v1.DetailReq) (res *v1.DetailRes, err error) {
	item, ok := c.registry.Lookup(req.Ref)
	if !ok {
		return nil, bizerr.NewCode(jobmgmtsvc.CodeJobHandlerNotFound)
	}
	return &v1.DetailRes{
		Ref:          item.Ref,
		DisplayName:  c.localizeHandlerName(ctx, item.Ref, item.DisplayName),
		Description:  c.localizeHandlerDescription(ctx, item.Ref, item.Description),
		Source:       v1.Source(item.Source),
		PluginId:     item.PluginID,
		ParamsSchema: item.ParamsSchema,
	}, nil
}
