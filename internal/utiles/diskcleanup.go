package utiles

import (
	"context"
	"fmt"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/volume"
	"github.com/atpx4869/dockpit/internal/svc"
)

type DiskUsageInfo struct {
	Images     DiskItem   `json:"images"`
	Containers DiskItem   `json:"containers"`
	Volumes    DiskItem   `json:"volumes"`
	BuildCache DiskItem   `json:"buildCache"`
	TotalSize  string     `json:"totalSize"`
}

type DiskItem struct {
	Count int    `json:"count"`
	Size  string `json:"size"`
	Reclaimable string `json:"reclaimable"`
}

// GetDiskUsage 获取 Docker 磁盘使用情况
func GetDiskUsage(ctx *svc.ServiceContext) (*DiskUsageInfo, error) {
	du, err := ctx.DockerClient.DiskUsage(context.Background(), types.DiskUsageOptions{})
	if err != nil {
		return nil, err
	}

	var imageSize, imageReclaim uint64
	for _, img := range du.Images {
		if img.Size > 0 {
			imageSize += uint64(img.Size)
		}
		if img.SharedSize == 0 {
			imageReclaim += uint64(img.Size)
		}
	}

	var volSize uint64
	for _, v := range du.Volumes {
		if v.UsageData != nil {
			volSize += uint64(v.UsageData.Size)
		}
	}

	var buildCacheSize uint64
	for _, bc := range du.BuildCache {
		if !bc.InUse {
			buildCacheSize += uint64(bc.Size)
		}
	}

	totalSize := imageSize + volSize + buildCacheSize

	return &DiskUsageInfo{
		Images: DiskItem{
			Count: len(du.Images),
			Size:  formatBytes(imageSize),
		},
		Containers: DiskItem{
			Count: len(du.Containers),
			Size:  func() string { var s uint64; for _, c := range du.Containers { s += uint64(c.SizeRw) }; return formatBytes(s) }(),
		},
		Volumes: DiskItem{
			Count: len(du.Volumes),
			Size:  formatBytes(volSize),
		},
		BuildCache: DiskItem{
			Count: len(du.BuildCache),
			Size:  formatBytes(buildCacheSize),
		},
		TotalSize: formatBytes(totalSize),
	}, nil
}

// PruneSystem 清理未使用的 Docker 资源
func PruneSystem(ctx *svc.ServiceContext) (string, error) {
	var output string

	// 清理未使用的镜像
	imgPrune, err := ctx.DockerClient.ImagesPrune(context.Background(), filters.NewArgs())
	if err == nil {
		output += fmt.Sprintf("镜像清理: 释放 %d 个镜像\n", len(imgPrune.ImagesDeleted))
	}

	// 清理停止的容器
	containerPrune, err := ctx.DockerClient.ContainersPrune(context.Background(), filters.NewArgs())
	if err == nil {
		output += fmt.Sprintf("容器清理: 释放 %d 个容器\n", len(containerPrune.ContainersDeleted))
	}

	// 清理未使用的卷
	volumePrune, err := ctx.DockerClient.VolumesPrune(context.Background(), filters.NewArgs())
	if err == nil {
		output += fmt.Sprintf("卷清理: 释放 %d 个卷\n", len(volumePrune.VolumesDeleted))
	}

	// 清理未使用的网络
	networkPrune, err := ctx.DockerClient.NetworksPrune(context.Background(), filters.NewArgs())
	if err == nil {
		output += fmt.Sprintf("网络清理: 释放 %d 个网络\n", len(networkPrune.NetworksDeleted))
	}

	// 清理构建缓存
	buildCachePrune, err := ctx.DockerClient.BuildCachePrune(context.Background(), types.BuildCachePruneOptions{All: true})
	if err == nil {
		output += fmt.Sprintf("构建缓存清理: 释放 %s\n", formatBytes(buildCachePrune.SpaceReclaimed))
	}

	return output, nil
}

// --- 网络管理 ---

type NetworkInfo struct {
	ID      string            `json:"id"`
	Name    string            `json:"name"`
	Driver  string            `json:"driver"`
	Scope   string            `json:"scope"`
	Subnet  string            `json:"subnet"`
	Gateway string            `json:"gateway"`
	IPv6    bool              `json:"ipv6"`
	Labels  map[string]string `json:"labels"`
}

func ListNetworks(ctx *svc.ServiceContext) ([]NetworkInfo, error) {
	networks, err := ctx.DockerClient.NetworkList(context.Background(), network.ListOptions{})
	if err != nil {
		return nil, err
	}

	var result []NetworkInfo
	for _, n := range networks {
		info := NetworkInfo{
			ID:     n.ID,
			Name:   n.Name,
			Driver: n.Driver,
			Scope:  n.Scope,
			IPv6:   n.EnableIPv6,
			Labels: n.Labels,
		}
		if len(n.IPAM.Config) > 0 {
			info.Subnet = n.IPAM.Config[0].Subnet
			info.Gateway = n.IPAM.Config[0].Gateway
		}
		result = append(result, info)
	}
	return result, nil
}

func RemoveNetwork(ctx *svc.ServiceContext, id string) error {
	return ctx.DockerClient.NetworkRemove(context.Background(), id)
}

// --- Volume 管理 ---

type VolumeInfo struct {
	Name       string            `json:"name"`
	Driver     string            `json:"driver"`
	Mountpoint string            `json:"mountpoint"`
	Size       string            `json:"size"`
	Labels     map[string]string `json:"labels"`
}

func ListVolumes(ctx *svc.ServiceContext) ([]VolumeInfo, error) {
	vols, err := ctx.DockerClient.VolumeList(context.Background(), volume.ListOptions{})
	if err != nil {
		return nil, err
	}

	var result []VolumeInfo
	for _, v := range vols.Volumes {
		info := VolumeInfo{
			Name:       v.Name,
			Driver:     v.Driver,
			Mountpoint: v.Mountpoint,
			Labels:     v.Labels,
		}
		if v.UsageData != nil {
			info.Size = formatBytes(uint64(v.UsageData.Size))
		}
		result = append(result, info)
	}
	return result, nil
}

func RemoveVolume(ctx *svc.ServiceContext, name string, force bool) error {
	return ctx.DockerClient.VolumeRemove(context.Background(), name, force)
}

// --- 镜像管理扩展 ---

type ImageDetail struct {
	ID      string   `json:"id"`
	Tags    []string `json:"tags"`
	Size    string   `json:"size"`
	Created string   `json:"created"`
}

func GetImagesDetailed(ctx *svc.ServiceContext) ([]ImageDetail, error) {
	images, err := ctx.DockerClient.ImageList(context.Background(), image.ListOptions{})
	if err != nil {
		return nil, err
	}

	var result []ImageDetail
	for _, img := range images {
		result = append(result, ImageDetail{
			ID:      img.ID,
			Tags:    img.RepoTags,
			Size:    formatBytes(uint64(img.Size)),
		})
	}
	return result, nil
}
