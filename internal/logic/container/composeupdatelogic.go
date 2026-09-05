package container

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/atpx4869/dockpit/internal/svc"
	"github.com/atpx4869/dockpit/internal/types"
	"github.com/atpx4869/dockpit/internal/utiles"

	"github.com/zeromicro/go-zero/core/logx"
)

type ComposeUpdateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewComposeUpdateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ComposeUpdateLogic {
	return &ComposeUpdateLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *ComposeUpdateLogic) ComposeUpdate(req *types.ComposeUpdateReq) (resp *types.Resp, err error) {
	resp = &types.Resp{}
	taskID := uuid.New().String()

	go func() {
		defer func() {
			if r := recover(); r != nil {
				l.Errorf("Compose 更新 panic: %v", r)
				l.svcCtx.UpdateProgress(taskID, svc.TaskProgress{
					TaskID: taskID, Percentage: 100, Name: req.Name,
					Message: "更新失败", DetailMsg: fmt.Sprintf("panic: %v", r), IsDone: true,
				})
			}
		}()

		l.svcCtx.UpdateProgress(taskID, svc.TaskProgress{
			TaskID: taskID, Percentage: 10, Name: req.Name,
			Message: "正在执行 compose 更新（失败自动回滚）", DetailMsg: "正在拉取镜像...", IsDone: false,
		})

		output, updateErr := utiles.ComposeUpdateWithRollback(l.svcCtx, req.WorkingDir, req.ConfigFile, req.Name)

		if updateErr != nil {
			l.Errorf("Compose 更新失败: %v", updateErr)
			l.svcCtx.UpdateProgress(taskID, svc.TaskProgress{
				TaskID: taskID, Percentage: 100, Name: req.Name,
				Message: "更新失败", DetailMsg: output, IsDone: true,
			})
			// 发送告警
			if l.svcCtx.Config.AlertWebhook != "" {
				_ = utiles.SendWebhookAlert(l.svcCtx.Config.AlertWebhook, utiles.UpdateFailedAlert(req.Name, output))
			}
			return
		}

		l.svcCtx.UpdateProgress(taskID, svc.TaskProgress{
			TaskID: taskID, Percentage: 100, Name: req.Name,
			Message: "Compose 更新成功", DetailMsg: output, IsDone: true,
		})
	}()

	resp.Code = 200
	resp.Msg = "success"
	resp.Data = map[string]string{"taskID": taskID}
	return resp, nil
}
