package container

import (
	"context"

	"github.com/atpx4869/dockpit/internal/svc"
	"github.com/atpx4869/dockpit/internal/types"
	"github.com/atpx4869/dockpit/internal/utiles"

	"github.com/zeromicro/go-zero/core/logx"
)

type LogsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewLogsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LogsLogic {
	return &LogsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *LogsLogic) GetLogs(req *types.ContainerLogsReq) (resp *types.Resp, err error) {
	resp = &types.Resp{}
	tail := req.Tail
	if tail <= 0 {
		tail = 100
	}

	logResult, err := utiles.GetContainerLogs(l.svcCtx, req.Id, tail)
	if err != nil {
		resp.Code = 400
		resp.Msg = err.Error()
		resp.Data = map[string]interface{}{}
		return resp, err
	}

	resp.Code = 200
	resp.Msg = "success"
	resp.Data = logResult
	return resp, nil
}
