package container

import (
	"context"

	"github.com/atpx4869/dockpit/internal/svc"
	"github.com/atpx4869/dockpit/internal/types"
	"github.com/atpx4869/dockpit/internal/utiles"

	"github.com/zeromicro/go-zero/core/logx"
)

type ComposeLogsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewComposeLogsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ComposeLogsLogic {
	return &ComposeLogsLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *ComposeLogsLogic) GetLogs(req *types.ComposeLogsReq) (resp *types.Resp, err error) {
	resp = &types.Resp{}
	tail := req.Tail
	if tail == "" {
		tail = "100"
	}

	output, err := utiles.ComposeLogsStream(req.WorkingDir, req.ConfigFile, req.Name, tail)
	if err != nil {
		resp.Code = 400
		resp.Msg = err.Error()
		return resp, err
	}

	resp.Code = 200
	resp.Msg = "success"
	resp.Data = map[string]string{"content": output}
	return resp, nil
}
