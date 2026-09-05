package container

import (
	"context"

	"github.com/atpx4869/dockpit/internal/svc"
	"github.com/atpx4869/dockpit/internal/types"
	"github.com/atpx4869/dockpit/internal/utiles"

	"github.com/zeromicro/go-zero/core/logx"
)

type ComposeListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewComposeListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ComposeListLogic {
	return &ComposeListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ComposeListLogic) ComposeList() (resp *types.Resp, err error) {
	resp = &types.Resp{}

	projects, err := utiles.ListComposeProjects(l.svcCtx)
	if err != nil {
		resp.Code = 500
		resp.Msg = err.Error()
		resp.Data = map[string]interface{}{}
		return resp, err
	}

	var projectInfoList []types.ComposeProjectInfo
	for _, p := range projects {
		projectInfoList = append(projectInfoList, types.ComposeProjectInfo{
			Name:       p.Name,
			WorkingDir: p.WorkingDir,
			ConfigFile: p.ConfigFile,
			Containers: p.Containers,
		})
	}

	resp.Code = 200
	resp.Msg = "success"
	resp.Data = projectInfoList
	return resp, nil
}
