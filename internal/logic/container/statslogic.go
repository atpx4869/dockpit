package container

import (
	"context"

	"github.com/atpx4869/dockpit/internal/svc"
	"github.com/atpx4869/dockpit/internal/types"
	"github.com/atpx4869/dockpit/internal/utiles"

	"github.com/zeromicro/go-zero/core/logx"
)

type StatsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewStatsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *StatsLogic {
	return &StatsLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *StatsLogic) GetStats(req *types.IdReq) (resp *types.Resp, err error) {
	resp = &types.Resp{}
	stats, err := utiles.GetContainerStats(l.svcCtx, req.Id)
	if err != nil {
		resp.Code = 400
		resp.Msg = err.Error()
		return resp, err
	}
	resp.Code = 200
	resp.Msg = "success"
	resp.Data = stats
	return resp, nil
}

func (l *StatsLogic) GetAllStats() (resp *types.Resp, err error) {
	resp = &types.Resp{}
	stats, err := utiles.GetAllContainerStats(l.svcCtx)
	if err != nil {
		resp.Code = 500
		resp.Msg = err.Error()
		return resp, err
	}
	resp.Code = 200
	resp.Msg = "success"
	resp.Data = stats
	return resp, nil
}
