package container

import (
	"context"

	"github.com/atpx4869/dockpit/internal/svc"
	"github.com/atpx4869/dockpit/internal/types"
	"github.com/atpx4869/dockpit/internal/utiles"

	"github.com/zeromicro/go-zero/core/logx"
)

type ExecLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewExecLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ExecLogic {
	return &ExecLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *ExecLogic) Exec(req *types.ContainerExecReq) (resp *types.Resp, err error) {
	resp = &types.Resp{}
	output, err := utiles.ExecInContainer(l.svcCtx, req.Id, req.Cmd)
	if err != nil {
		resp.Code = 400
		resp.Msg = err.Error()
		return resp, err
	}
	resp.Code = 200
	resp.Msg = "success"
	resp.Data = map[string]string{"output": output}
	return resp, nil
}
