package container

import (
	"context"

	"github.com/atpx4869/dockpit/internal/svc"
	"github.com/atpx4869/dockpit/internal/types"
	"github.com/atpx4869/dockpit/internal/utiles"

	"github.com/zeromicro/go-zero/core/logx"
)

type BatchLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewBatchLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BatchLogic {
	return &BatchLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *BatchLogic) Batch(req *types.BatchContainerReq) (resp *types.Resp, err error) {
	resp = &types.Resp{}
	results, err := utiles.BatchContainerOperation(l.svcCtx, req.IDs, req.Action)
	if err != nil {
		resp.Code = 400
		resp.Msg = err.Error()
		return resp, err
	}
	resp.Code = 200
	resp.Msg = "success"
	resp.Data = results
	return resp, nil
}
