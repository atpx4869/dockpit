package utiles

import (
	"context"
	"github.com/docker/docker/api/types/container"
	"github.com/atpx4869/dockpit/internal/svc"
	MyType "github.com/atpx4869/dockpit/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
)

func GetContainerList(ctx *svc.ServiceContext) ([]MyType.Container, error) {
	dockerContainerList, err := ctx.DockerClient.ContainerList(context.Background(), container.ListOptions{
		All: true,
	})
	if err != nil {
		logx.Errorf("get container list error: %v", err)
		return nil, err
	}
	var containerList []MyType.Container
	for _, dockerContainerInfo := range dockerContainerList {
		containerInfo := MyType.Container{
			Container: dockerContainerInfo,
		}
		containerList = append(containerList, containerInfo)
	}
	return containerList, nil
}

func CheckImageUpdate(ctx *svc.ServiceContext, containerListData []MyType.Container) []MyType.Container {
	for i, v := range containerListData {
		if _, ok := ctx.HubImageInfo.Data[v.ImageID]; ok {
			if ctx.HubImageInfo.Data[v.ImageID].NeedUpdate {
				containerListData[i].Update = true
			}
		}
	}
	return containerListData
}

// ContainerHealthInfo 容器健康检查信息
type ContainerHealthInfo struct {
	ContainerID   string `json:"containerId"`
	ContainerName string `json:"containerName"`
	HealthStatus  string `json:"healthStatus"` // healthy / unhealthy / starting / none
	HealthLog     string `json:"healthLog"`
}

// GetContainerHealth 获取容器健康检查状态
func GetContainerHealth(ctx *svc.ServiceContext, id string) (*ContainerHealthInfo, error) {
	info, err := ctx.DockerClient.ContainerInspect(context.Background(), id)
	if err != nil {
		return nil, err
	}

	healthInfo := &ContainerHealthInfo{
		ContainerID: id,
	}

	if len(info.Name) > 0 {
		healthInfo.ContainerName = info.Name[1:]
	}

	if info.State.Health == nil {
		healthInfo.HealthStatus = "none"
		return healthInfo, nil
	}

	healthInfo.HealthStatus = info.State.Health.Status

	// 获取最后一条健康检查日志
	if len(info.State.Health.Log) > 0 {
		lastLog := info.State.Health.Log[len(info.State.Health.Log)-1]
		healthInfo.HealthLog = lastLog.Output
	}

	return healthInfo, nil
}

// GetAllContainerHealth 获取所有容器的健康检查状态
func GetAllContainerHealth(ctx *svc.ServiceContext) ([]ContainerHealthInfo, error) {
	containers, err := ctx.DockerClient.ContainerList(context.Background(), container.ListOptions{All: true})
	if err != nil {
		return nil, err
	}

	var results []ContainerHealthInfo
	for _, c := range containers {
		name := ""
		if len(c.Names) > 0 {
			name = c.Names[0][1:]
		}
		info, err := ctx.DockerClient.ContainerInspect(context.Background(), c.ID)
		if err != nil {
			results = append(results, ContainerHealthInfo{
				ContainerID:   c.ID,
				ContainerName: name,
				HealthStatus:  "unknown",
			})
			continue
		}

		healthStatus := "none"
		var healthLog string
		if info.State.Health != nil {
			healthStatus = info.State.Health.Status
			if len(info.State.Health.Log) > 0 {
				healthLog = info.State.Health.Log[len(info.State.Health.Log)-1].Output
			}
		}

		results = append(results, ContainerHealthInfo{
			ContainerID:   c.ID,
			ContainerName: name,
			HealthStatus:  healthStatus,
			HealthLog:     healthLog,
		})
	}
	return results, nil
}
