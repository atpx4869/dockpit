package container

import (
	"context"
	"time"

	"github.com/atpx4869/dockpit/internal/svc"
	"github.com/atpx4869/dockpit/internal/types"
	"github.com/atpx4869/dockpit/internal/utiles"

	"github.com/zeromicro/go-zero/core/logx"
)

type ContainersListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

type Info struct {
	Id           string  `json:"id"`
	Status       string  `json:"status"`
	Name         string  `json:"name"`
	UsingImage   string  `json:"usingImage"`
	CreateImage  string  `json:"createImage"`
	CreateTime   string  `json:"createTime"`
	RunningTime  string  `json:"runningTime"`
	HaveUpdate   bool    `json:"haveUpdate"`
	HealthStatus string  `json:"healthStatus"`
	IsCompose    bool    `json:"isCompose"`
	ComposeName  string  `json:"composeName"`
	CPUPercent   float64 `json:"cpuPercent,omitempty"`
	MemUsage     string  `json:"memUsage,omitempty"`
	MemPercent   float64 `json:"memPercent,omitempty"`
}

func NewContainersListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ContainersListLogic {
	return &ContainersListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ContainersListLogic) ContainersList() (resp *types.Resp, err error) {
	resp = &types.Resp{}
	list, err := utiles.GetContainerList(l.svcCtx)
	if err != nil {
		resp.Code = 500
		resp.Msg = err.Error()
		resp.Data = map[string]interface{}{}
		return resp, err
	}
	resp.Msg = "success"
	var containerInfoList []Info
	list = utiles.CheckImageUpdate(l.svcCtx, list)
	for _, v := range list {
		var containerInfo Info
		containerInfo.Id = v.ID
		containerInfo.Status = v.State
		if len(v.Names) > 0 {
			containerInfo.Name = v.Names[0][1:]
		} else {
			containerInfo.Name = "unknown"
		}
		containerInfo.UsingImage = v.Image
		if containerInfo.UsingImage == "" {
			containerInfo.UsingImage = v.ImageID
		}

		containerInspect, err := utiles.GetContainerInspect(l.svcCtx, v.ID)
		if err == nil {
			containerInfo.CreateImage = containerInspect.Config.Image
			// 健康检查状态
			if containerInspect.State.Health != nil {
				containerInfo.HealthStatus = containerInspect.State.Health.Status
			} else {
				containerInfo.HealthStatus = "none"
			}
			// Compose 标签检测
			labels := containerInspect.Config.Labels
			if projectName, ok := labels["com.docker.compose.project"]; ok {
				containerInfo.IsCompose = true
				containerInfo.ComposeName = projectName
			}
		}

		t := time.Unix(v.Created, 0)
		containerInfo.CreateTime = t.Format("2006-01-02 15:04:05")
		containerInfo.RunningTime = v.Status
		containerInfo.HaveUpdate = v.Update

		// 只对 running 容器获取资源统计
		if v.State == "running" {
			stats, err := utiles.GetContainerStats(l.svcCtx, v.ID)
			if err == nil {
				containerInfo.CPUPercent = stats.CPUPercent
				containerInfo.MemUsage = stats.MemUsage
				containerInfo.MemPercent = stats.MemPercent
			}
		}

		containerInfoList = append(containerInfoList, containerInfo)
	}
	resp.Data = containerInfoList
	return resp, nil
}
