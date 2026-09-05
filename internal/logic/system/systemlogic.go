package system

import (
	"context"

	"github.com/atpx4869/dockpit/internal/svc"
	"github.com/atpx4869/dockpit/internal/types"
	"github.com/atpx4869/dockpit/internal/utiles"

	"github.com/zeromicro/go-zero/core/logx"
)

type SystemLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSystemLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SystemLogic {
	return &SystemLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *SystemLogic) DiskUsage() (resp *types.Resp, err error) {
	resp = &types.Resp{}
	info, err := utiles.GetDiskUsage(l.svcCtx)
	if err != nil {
		resp.Code = 500
		resp.Msg = err.Error()
		return resp, err
	}
	resp.Code = 200
	resp.Msg = "success"
	resp.Data = info
	return resp, nil
}

func (l *SystemLogic) DiskCleanup() (resp *types.Resp, err error) {
	resp = &types.Resp{}
	output, err := utiles.PruneSystem(l.svcCtx)
	if err != nil {
		resp.Code = 500
		resp.Msg = err.Error()
		return resp, err
	}
	resp.Code = 200
	resp.Msg = "success"
	resp.Data = map[string]string{"output": output}
	return resp, nil
}

func (l *SystemLogic) NetworkList() (resp *types.Resp, err error) {
	resp = &types.Resp{}
	networks, err := utiles.ListNetworks(l.svcCtx)
	if err != nil {
		resp.Code = 500
		resp.Msg = err.Error()
		return resp, err
	}
	resp.Code = 200
	resp.Msg = "success"
	resp.Data = networks
	return resp, nil
}

func (l *SystemLogic) NetworkRemove(req *types.IdReq) (resp *types.Resp, err error) {
	resp = &types.Resp{}
	err = utiles.RemoveNetwork(l.svcCtx, req.Id)
	if err != nil {
		resp.Code = 400
		resp.Msg = err.Error()
		return resp, err
	}
	resp.Code = 200
	resp.Msg = "success"
	return resp, nil
}

func (l *SystemLogic) VolumeList() (resp *types.Resp, err error) {
	resp = &types.Resp{}
	volumes, err := utiles.ListVolumes(l.svcCtx)
	if err != nil {
		resp.Code = 500
		resp.Msg = err.Error()
		return resp, err
	}
	resp.Code = 200
	resp.Msg = "success"
	resp.Data = volumes
	return resp, nil
}

func (l *SystemLogic) VolumeRemove(req *types.RemoveVolumeReq) (resp *types.Resp, err error) {
	resp = &types.Resp{}
	err = utiles.RemoveVolume(l.svcCtx, req.Name, req.Force)
	if err != nil {
		resp.Code = 400
		resp.Msg = err.Error()
		return resp, err
	}
	resp.Code = 200
	resp.Msg = "success"
	return resp, nil
}
