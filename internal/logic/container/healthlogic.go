package container

import (
	"context"

	"github.com/atpx4869/dockpit/internal/svc"
	"github.com/atpx4869/dockpit/internal/types"
	"github.com/atpx4869/dockpit/internal/utiles"

	"github.com/zeromicro/go-zero/core/logx"
)

type HealthLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewHealthLogic(ctx context.Context, svcCtx *svc.ServiceContext) *HealthLogic {
	return &HealthLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *HealthLogic) GetHealth(req *types.IdReq) (resp *types.Resp, err error) {
	resp = &types.Resp{}
	health, err := utiles.GetContainerHealth(l.svcCtx, req.Id)
	if err != nil {
		resp.Code = 400
		resp.Msg = err.Error()
		return resp, err
	}
	resp.Code = 200
	resp.Msg = "success"
	resp.Data = health
	return resp, nil
}

func (l *HealthLogic) GetAllHealth() (resp *types.Resp, err error) {
	resp = &types.Resp{}
	healthList, err := utiles.GetAllContainerHealth(l.svcCtx)
	if err != nil {
		resp.Code = 500
		resp.Msg = err.Error()
		return resp, err
	}
	resp.Code = 200
	resp.Msg = "success"
	resp.Data = healthList
	return resp, nil
}
